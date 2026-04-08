package executor

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Executor runs shell commands in a given working directory.
type Executor interface {
	Run(dir string, command string) error
	// RunCapture runs command in dir and returns its stdout. On failure the
	// error message contains the last meaningful line of stderr, identical to Run.
	RunCapture(dir string, command string) (string, error)
}

// OsExecutor runs commands via sh -c using os/exec.
// When a command fails, the error message includes the command's stderr output
// so callers see the actual failure reason (e.g. "fatal: repository not found")
// rather than a bare exit code.
type OsExecutor struct{}

func (e OsExecutor) Run(dir string, command string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := lastMeaningfulLine(stderr.String())
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

func (e OsExecutor) RunCapture(dir, command string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := lastMeaningfulLine(stderr.String())
		if msg != "" {
			return stdout.String(), fmt.Errorf("%s", msg)
		}
		return stdout.String(), err
	}
	return stdout.String(), nil
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
