package search

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestWithOpenJobsPreservesFilterGroups(t *testing.T) {
	cases := []struct {
		name        string
		input, want any
	}{
		{"unfiltered", nil, "NOT is_closed = true AND NOT is_duplicate = true"},
		{"expression", `source = "a" OR source = "b"`, `(source = "a" OR source = "b") AND NOT is_closed = true AND NOT is_duplicate = true`},
		{"groups", [][]string{{`source = "a"`, `source = "b"`}, {`country = "DE"`}}, [][]string{{`source = "a"`, `source = "b"`}, {`country = "DE"`}, {"NOT is_closed = true AND NOT is_duplicate = true"}}},
		{"flat", []string{`country = "DE"`}, []string{`country = "DE"`, "NOT is_closed = true AND NOT is_duplicate = true"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := withOpenJobs(c.input); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("filter=%v, want %v", got, c.want)
			}
		})
	}
}

// TestEnableDocumentEditsTurnsOnTheFeature pins the one request CloseJobs and
// MarkDuplicateJobs depend on: without it a fresh engine refuses every marker write.
func TestEnableDocumentEditsTurnsOnTheFeature(t *testing.T) {
	var method, path, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		method, path, body = r.Method, r.URL.Path, string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"editDocumentsByFunction":true}`)
	}))
	defer srv.Close()

	if err := NewClient(srv.URL, "test-key").EnableDocumentEdits(t.Context()); err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPatch || path != "/experimental-features" {
		t.Errorf("request %s %s, want PATCH /experimental-features", method, path)
	}
	if !strings.Contains(body, `"editDocumentsByFunction":true`) {
		t.Errorf("body %s does not enable editDocumentsByFunction", body)
	}
}
