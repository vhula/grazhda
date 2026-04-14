package executor

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestLastMeaningfulLine(t *testing.T) {
	got := lastMeaningfulLine("line1\n\nline2\n")
	if got != "line2" {
		t.Fatalf("expected last non-empty line, got %q", got)
	}
}

func TestRunCapture(t *testing.T) {
	exec := OsExecutor{}
	out, err := exec.RunCapture(".", "echo hello")
	if err != nil {
		t.Fatalf("RunCapture failed: %v", err)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected output to contain hello, got %q", out)
	}
}

func TestRunContextCancelled(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sleep command differs on windows")
	}
	exec := OsExecutor{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := exec.RunContext(ctx, ".", "sleep 5")
	if err == nil {
		t.Fatal("expected interruption error, got nil")
	}
	if !strings.Contains(err.Error(), "interrupted") {
		t.Fatalf("expected interrupted error, got %v", err)
	}
}
