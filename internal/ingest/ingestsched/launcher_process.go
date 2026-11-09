package ingestsched

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// ProcessLauncher runs each claimed run as a CHILD PROCESS of the scheduler — the launcher
// for a deployment with no systemd, such as the local Docker edition, where the scheduler
// is one long-lived container.
//
// It keeps the contract SystemdLauncher keeps, for the same reasons: the provider key is
// validated at launch, the run is told its budget (INGEST_RUN_TIMEOUT_SECONDS) and is killed
// past it, and Finished reports how it ended so the reaper can release the claim. Children
// outlive the tick that started them because the scheduler process does: cmd/ingest-scheduler
// loops when this launcher is selected instead of exiting after one tick.
//
// What a child cannot do is outlive the scheduler. A restart loses every child, so Finished
// reports a run it never launched as ended in failure — its claim is released now rather
// than held for the whole reclaim window by a process that no longer exists.
type ProcessLauncher struct {
	// IngestBinary is the path to cmd/ingest inside the container.
	IngestBinary string

	mu   sync.Mutex
	runs map[string]*processRun

	// start is the seam the tests use. nil means really start the process.
	start func(run Run, args []string, env []string) (wait func() (int, string), kill func(), err error)
}

type processRun struct {
	done   bool
	code   int
	detail string
}

// NewProcessLauncher builds the launcher for a deployment without systemd.
func NewProcessLauncher(ingestBinary string) *ProcessLauncher {
	return &ProcessLauncher{IngestBinary: ingestBinary, runs: map[string]*processRun{}}
}

func processKey(run Run) string { return run.Provider + "#" + strconv.Itoa(run.Shard) }

// Launch starts run as a child process and returns once it has started.
func (l *ProcessLauncher) Launch(_ context.Context, run Run) error {
	if err := ValidateProviderKey(run.Provider); err != nil {
		return fmt.Errorf("refusing to launch: %w", err)
	}
	key := processKey(run)
	l.mu.Lock()
	if r, ok := l.runs[key]; ok && !r.done {
		l.mu.Unlock()
		return fmt.Errorf("%s is still running", key)
	}
	l.mu.Unlock()

	timeoutSec := strconv.Itoa(int(run.RunTimeout.Seconds()))
	args := []string{run.Provider}
	if run.Shards > 1 {
		// No selector for an unsharded provider, as SystemdLauncher: one spelling per instruction.
		args = append(args, fmt.Sprintf("--shard=%d/%d", run.Shard, run.Shards))
	}
	env := append(os.Environ(), "INGEST_RUN_TIMEOUT_SECONDS="+timeoutSec)

	start := l.start
	if start == nil {
		start = l.startProcess
	}
	wait, kill, err := start(run, args, env)
	if err != nil {
		return err
	}

	r := &processRun{}
	l.mu.Lock()
	l.runs[key] = r
	l.mu.Unlock()

	// The budget is enforced here because nothing else will: systemd's TimeoutStartSec has
	// no counterpart for a bare child. cmd/ingest stops its own crawl a reserve short of it,
	// so this kill is the backstop for a run that does not.
	killed := make(chan struct{})
	timer := time.AfterFunc(run.RunTimeout, func() { close(killed); kill() })
	go func() {
		code, detail := wait()
		timer.Stop()
		select {
		case <-killed:
			detail = "timeout"
		default:
		}
		l.mu.Lock()
		r.done, r.code, r.detail = true, code, detail
		l.mu.Unlock()
	}()
	return nil
}

func (l *ProcessLauncher) startProcess(run Run, args []string, env []string) (func() (int, string), func(), error) {
	// exec.Command, not CommandContext: the run must outlive the tick's context.
	cmd := exec.Command(l.IngestBinary, args...)
	cmd.Env = env
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start %s: %w", processKey(run), err)
	}
	log.Printf("scheduler: started %s (pid %d)", processKey(run), cmd.Process.Pid)
	wait := func() (int, string) {
		err := cmd.Wait()
		var exit *exec.ExitError
		switch {
		case err == nil:
			return 0, "success"
		case errors.As(err, &exit):
			if exit.ExitCode() < 0 {
				return 1, "signal"
			}
			return exit.ExitCode(), "exit-code"
		default:
			return 1, err.Error()
		}
	}
	kill := func() { _ = cmd.Process.Kill() }
	return wait, kill, nil
}

// Finished reports how a run ended, and forgets it once read — the counterpart of
// SystemdLauncher's reset-failed, freeing the shard for its next launch.
func (l *ProcessLauncher) Finished(_ context.Context, run Run) (Outcome, error) {
	key := processKey(run)
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.runs[key]
	if !ok {
		return Outcome{Done: true, ExitCode: 1, Detail: "lost: the scheduler restarted while it ran"}, nil
	}
	if !r.done {
		return Outcome{Done: false}, nil
	}
	delete(l.runs, key)
	return Outcome{Done: true, ExitCode: r.code, Detail: r.detail}, nil
}
