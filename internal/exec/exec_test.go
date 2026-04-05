package exec

import (
	"strings"
	"testing"
)

func TestRunShell_SuccessReturnsOutput(t *testing.T) {
	output, err := RunShell("printf 'hello'", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "hello" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunShell_FailureReturnsOutputAndError(t *testing.T) {
	output, err := RunShell("echo boom >&2; exit 12", t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(output, "boom") {
		t.Fatalf("expected output to include stderr message, got: %q", output)
	}
}
