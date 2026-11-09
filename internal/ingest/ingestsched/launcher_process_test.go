package ingestsched

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"
)

// fakeChild stands in for a started process: the test decides when it exits and how.
type fakeChild struct {
	args, env []string
	exit      chan int
	killed    chan struct{}
}

func newFakeLauncher() (*ProcessLauncher, *[]*fakeChild) {
	l := NewProcessLauncher("/app/ingest")
	var children []*fakeChild
	l.start = func(_ Run, args, env []string) (func() (int, string), func(), error) {
		c := &fakeChild{args: args, env: env, exit: make(chan int, 1), killed: make(chan struct{})}
		children = append(children, c)
		wait := func() (int, string) {
			select {
			case code := <-c.exit:
				return code, "exit-code"
			case <-c.killed:
				return 1, "signal"
			}
		}
		return wait, func() { close(c.killed) }, nil
	}
	return l, &children
}

func waitDone(t *testing.T, l *ProcessLauncher, run Run) Outcome {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		out, err := l.Finished(context.Background(), run)
		if err != nil {
			t.Fatal(err)
		}
		if out.Done {
			return out
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("run never finished")
	return Outcome{}
}

func TestProcessLauncher_RunsTheProviderAndReportsItsExit(t *testing.T) {
	l, children := newFakeLauncher()
	run := Run{Provider: "greenhouse", Shard: 1, Shards: 1, RunTimeout: time.Minute}

	if err := l.Launch(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	c := (*children)[0]
	if !slices.Equal(c.args, []string{"greenhouse"}) {
		t.Errorf("args = %v, want just the provider for an unsharded run", c.args)
	}
	if !slices.Contains(c.env, "INGEST_RUN_TIMEOUT_SECONDS=60") {
		t.Error("the run is not told its budget")
	}
	if out, _ := l.Finished(context.Background(), run); out.Done {
		t.Fatal("reported finished while still running")
	}
	if err := l.Launch(context.Background(), run); err == nil {
		t.Error("launched a second copy of a running shard")
	}

	c.exit <- 3
	if out := waitDone(t, l, run); out.ExitCode != 3 {
		t.Errorf("exit code = %d, want 3", out.ExitCode)
	}
	// Once read, the shard is free to launch again.
	if err := l.Launch(context.Background(), run); err != nil {
		t.Errorf("relaunch after finish: %v", err)
	}
}

func TestProcessLauncher_PassesTheShardSelector(t *testing.T) {
	l, children := newFakeLauncher()
	if err := l.Launch(context.Background(), Run{Provider: "workday", Shard: 2, Shards: 6, RunTimeout: time.Minute}); err != nil {
		t.Fatal(err)
	}
	if got := (*children)[0].args; !slices.Equal(got, []string{"workday", "--shard=2/6"}) {
		t.Errorf("args = %v", got)
	}
}

func TestProcessLauncher_KillsARunPastItsBudget(t *testing.T) {
	l, _ := newFakeLauncher()
	run := Run{Provider: "greenhouse", Shard: 1, Shards: 1, RunTimeout: 20 * time.Millisecond}
	if err := l.Launch(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	if out := waitDone(t, l, run); out.Detail != "timeout" || out.ExitCode == 0 {
		t.Errorf("outcome = %+v, want a failed timeout", out)
	}
}

func TestProcessLauncher_RefusesAnUnsafeProviderKey(t *testing.T) {
	l, children := newFakeLauncher()
	err := l.Launch(context.Background(), Run{Provider: "x; rm -rf /", Shard: 1, Shards: 1, RunTimeout: time.Minute})
	if err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if len(*children) != 0 {
		t.Error("started a process for an unsafe key")
	}
}

// A restart loses every child: a run this launcher never saw must release its claim now.
func TestProcessLauncher_AnUnknownRunIsReportedAsFailed(t *testing.T) {
	l, _ := newFakeLauncher()
	out, err := l.Finished(context.Background(), Run{Provider: "lever", Shard: 1, Shards: 1})
	if err != nil || !out.Done || out.ExitCode == 0 {
		t.Fatalf("outcome = %+v, %v; want done and failed", out, err)
	}
}
