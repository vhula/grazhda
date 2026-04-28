package reporter

import (
"fmt"
"io"
"sync"
"time"

clr "github.com/vhula/grazhda/internal/color"
)

// Color helpers — disabled automatically when output is not a TTY or NO_COLOR is set.
var (
green  = clr.Green
red    = clr.Red
yellow = clr.Yellow
blue   = clr.Blue
)

// ExitError is returned by RunE handlers when operations completed but some
// repositories failed (i.e. non-zero ExitCode). Using a typed error lets
// Execute() distinguish "partial failure" from "command error" and apply the
// correct exit code without calling os.Exit directly inside RunE.
type ExitError struct {
Code int
}

// Error implements the error interface, returning the exit code as a string.
func (e ExitError) Error() string { return fmt.Sprintf("exit code %d", e.Code) }

// OpResult records the outcome of a single repository operation.
type OpResult struct {
Workspace   string
Project     string
Repo        string
Skipped     bool
Err         error
Msg         string        // human-readable description, e.g. "cloned (main)", "already exists, skipped"
OutputLines []string      // optional per-repo command output printed after the status line
Elapsed     time.Duration // duration of the operation; 0 means not measured
}

// Reporter accumulates operation results and produces structured progress output.
type Reporter struct {
out         io.Writer
errOut      io.Writer
mu          sync.Mutex
results     []OpResult
total       int  // expected total for parallel progress (0 = disabled)
done        int  // number of completed operations
ShowElapsed bool // when true, print elapsed time per repo after the status message
Quiet       bool // when true, suppress all stdout progress (errors still go to stderr)
}

// NewReporter creates a Reporter that writes progress to out and errors to errOut.
func NewReporter(out, errOut io.Writer) *Reporter {
return &Reporter{out: out, errOut: errOut}
}

// SetTotal configures the parallel progress counter. Call before launching
// goroutines with the total number of repositories that will be processed.
// Pass 0 to disable progress tracking (the default).
func (r *Reporter) SetTotal(n int) {
r.mu.Lock()
defer r.mu.Unlock()
r.total = n
r.done = 0
}

// PrintDryRunBanner writes a prominent "[DRY RUN] Preview" header to stdout.
// Call this at the start of any mutating command when --dry-run is active.
// Suppressed in Quiet mode.
func (r *Reporter) PrintDryRunBanner() {
r.mu.Lock()
defer r.mu.Unlock()
if r.Quiet {
return
}
fmt.Fprintln(r.out, yellow("[DRY RUN] Preview — no changes will be made."))
fmt.Fprintln(r.out)
}

// PrintLine writes an informational line (blue) to stdout, e.g. section headers.
// Suppressed in Quiet mode.
func (r *Reporter) PrintLine(msg string) {
r.mu.Lock()
defer r.mu.Unlock()
if r.Quiet {
return
}
fmt.Fprintln(r.out, blue(msg))
}

// PrintWarn writes a yellow warning line to stdout.
// Suppressed in Quiet mode.
func (r *Reporter) PrintWarn(msg string) {
r.mu.Lock()
defer r.mu.Unlock()
if r.Quiet {
return
}
fmt.Fprintln(r.out, yellow(msg))
}

// Record appends an operation result and prints the per-repo status line.
func (r *Reporter) Record(res OpResult) {
r.mu.Lock()
defer r.mu.Unlock()
r.results = append(r.results, res)

if r.Quiet {
// In quiet mode only increment the done counter; no stdout output.
if r.total > 0 {
r.done++
}
return
}

symbol := green("✓")
if res.Err != nil {
symbol = red("✗")
} else if res.Skipped {
symbol = yellow("⏭")
}

displayMsg := res.Msg
if res.Err != nil {
if displayMsg == "" {
displayMsg = res.Err.Error()
}
displayMsg = red(displayMsg)
} else if res.Skipped {
displayMsg = yellow(displayMsg)
} else {
displayMsg = green(displayMsg)
}

line := fmt.Sprintf("    %s %-14s — %s", symbol, res.Repo, displayMsg)

if r.ShowElapsed && res.Elapsed > 0 {
line += fmt.Sprintf("  [%s]", fmtDuration(res.Elapsed))
}

if r.total > 0 {
r.done++
line += fmt.Sprintf("  (%d/%d)", r.done, r.total)
}

fmt.Fprintln(r.out, line)

for _, line := range res.OutputLines {
fmt.Fprintf(r.out, "      %s\n", line)
}
}

// Summary prints the run summary to stdout and failure details to stderr.
// successLabel is the verb for successful operations (e.g. "cloned", "pulled", "removed").
// When dryRun is true, the summary line is prefixed with "[DRY RUN]".
// In Quiet mode, only failures are written to stderr (no stdout summary line).
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

// Always write failure details to stderr regardless of mode.
for _, res := range r.results {
if res.Err != nil {
fmt.Fprintf(r.errOut, "      %s\n", red(res.Repo+": "+res.Err.Error()))
}
}

if r.Quiet {
return
}

prefix := ""
if dryRun {
prefix = yellow("[DRY RUN] ")
}
fmt.Fprintf(r.out, "\n%s%s %d %s  %s %d skipped  %s %d failed\n",
prefix,
green("✓"), success, successLabel,
yellow("⏭"), skipped,
red("✗"), failed)
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

// fmtDuration formats an elapsed duration as "1.2s" or "345ms".
func fmtDuration(d time.Duration) string {
if d >= time.Second {
return fmt.Sprintf("%.1fs", d.Seconds())
}
return fmt.Sprintf("%dms", d.Milliseconds())
}
