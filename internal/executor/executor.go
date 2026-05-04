package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Executor runs shell commands in a given working directory.
type Executor interface {
	// Run executes command in dir using a background context.
	Run(dir string, command string) error

	// RunContext executes command in dir, honouring ctx cancellation.
	// When ctx is cancelled the child process is killed and an error is returned.
	RunContext(ctx context.Context, dir string, command string) error

	// RunCapture runs command in dir and returns its stdout. On failure the
	// error message contains the last meaningful line of stderr, identical to Run.
	RunCapture(dir string, command string) (string, error)

	// RunCaptureContext is like RunCapture but honours ctx cancellation.
	RunCaptureContext(ctx context.Context, dir string, command string) (string, error)

	// RunInteractive runs command in dir with stdin, stdout, and stderr
	// connected directly to the terminal. Use this for interactive programs
	// such as editors and pagers.
	RunInteractive(ctx context.Context, dir string, command string) error
}

// OsExecutor runs commands using os/exec.
// On Unix systems commands are executed via sh -c; on Windows via cmd /C.
// When a command fails, the error message includes the command's stderr output
// so callers see the actual failure reason (e.g. "fatal: repository not found")
// rather than a bare exit code.
type OsExecutor struct{}

// Run executes command in dir using a background context.
func (e OsExecutor) Run(dir string, command string) error {
	return e.RunContext(context.Background(), dir, command)
}

// RunContext executes command in dir, honouring ctx cancellation.
func (e OsExecutor) RunContext(ctx context.Context, dir string, command string) error {
	var stderr bytes.Buffer
	cmd := shellCommand(ctx, command)
	cmd.Dir = dir
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("interrupted: %w", ctx.Err())
		}
		msg := lastMeaningfulLine(stderr.String())
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

// RunCapture runs command in dir and returns its combined stdout.
func (e OsExecutor) RunCapture(dir, command string) (string, error) {
	return e.RunCaptureContext(context.Background(), dir, command)
}

// RunCaptureContext is like RunCapture but honours ctx cancellation.
func (e OsExecutor) RunCaptureContext(ctx context.Context, dir, command string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := shellCommand(ctx, command)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return stdout.String(), fmt.Errorf("interrupted: %w", ctx.Err())
		}
		msg := lastMeaningfulLine(stderr.String())
		if msg != "" {
			return stdout.String(), fmt.Errorf("%s", msg)
		}
		return stdout.String(), err
	}
	return stdout.String(), nil
}

// RunInteractive runs command in dir with stdin/stdout/stderr attached to the terminal.
func (e OsExecutor) RunInteractive(ctx context.Context, dir string, command string) error {
	cmd := shellCommand(ctx, command)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// shellCommand returns an exec.Cmd for the given command string.
// On Windows it uses cmd /C; on all other platforms it uses sh -c.
func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}

// lastMeaningfulLine returns the last non-empty line from s.
// Git (and most CLI tools) write the most relevant failure reason last,
// preceded by progress lines like "Cloning into '...'...".
func lastMeaningfulLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}
