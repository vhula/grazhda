package reporter

import (
	"fmt"
	"io"
	"sync"
)

// OpResult records the outcome of a single repository operation.
type OpResult struct {
	Workspace string
	Project   string
	Repo      string
	Skipped   bool
	Err       error
	Msg       string // human-readable description, e.g. "cloned (main)", "already exists, skipped"
}

// Reporter accumulates operation results and produces structured progress output.
type Reporter struct {
	out     io.Writer
	errOut  io.Writer
	mu      sync.Mutex
	results []OpResult
}

// NewReporter creates a Reporter that writes progress to out and errors to errOut.
func NewReporter(out, errOut io.Writer) *Reporter {
	return &Reporter{out: out, errOut: errOut}
}

// PrintLine writes a plain line to stdout (for section headers and verbose output).
func (r *Reporter) PrintLine(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprintln(r.out, msg)
}

// Record appends an operation result and prints the per-repo status line.
func (r *Reporter) Record(res OpResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results = append(r.results, res)

	symbol := "✓"
	if res.Err != nil {
		symbol = "✗"
	} else if res.Skipped {
		symbol = "⏭"
	}

	displayMsg := res.Msg
	if res.Err != nil && displayMsg == "" {
		displayMsg = res.Err.Error()
	}

	fmt.Fprintf(r.out, "    %s %-14s — %s\n", symbol, res.Repo, displayMsg)
}

// Summary prints the run summary to stdout and failure details to stderr.
// successLabel is the verb for successful operations (e.g. "cloned", "pulled", "removed").
// When dryRun is true, the summary line is prefixed with "[DRY RUN]".
func (r *Reporter) Summary(successLabel string, dryRun bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var success, skipped, failed int
	for _, res := range r.results {
		if res.Err != nil {
			failed++
		} else if res.Skipped {
			skipped++
		} else {
			success++
		}
	}

	prefix := ""
	if dryRun {
		prefix = "[DRY RUN] "
	}
	fmt.Fprintf(r.out, "\n%s✓ %d %s  ⏭ %d skipped  ✗ %d failed\n",
		prefix, success, successLabel, skipped, failed)

	for _, res := range r.results {
		if res.Err != nil {
			fmt.Fprintf(r.errOut, "      %s: %s\n", res.Repo, res.Err.Error())
		}
	}
}

// ExitCode returns 0 if all operations succeeded, 1 if any failed.
func (r *Reporter) ExitCode() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, res := range r.results {
		if res.Err != nil {
			return 1
		}
	}
	return 0
}
