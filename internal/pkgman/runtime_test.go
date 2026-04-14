package pkgman

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestSourceEnvScript(t *testing.T) {
	s := sourceEnvScript()
	if !strings.Contains(s, ".grazhda.env") {
		t.Fatalf("sourceEnvScript missing env file reference: %q", s)
	}
}

func TestRemoveBlockIfPresent_MissingFile(t *testing.T) {
	removed, err := removeBlockIfPresent("/definitely/missing/.grazhda.env", "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed {
		t.Fatal("expected removed=false for missing file")
	}
}

func TestStreamLines(t *testing.T) {
	var out bytes.Buffer
	streamLines(strings.NewReader("a\nb\n"), ">> ", &out)
	got := out.String()
	if !strings.Contains(got, ">> a") || !strings.Contains(got, ">> b") {
		t.Fatalf("unexpected stream output: %q", got)
	}
}

func TestSpinnerStop(t *testing.T) {
	var out bytes.Buffer
	s := NewSpinner(&out, "working")
	time.Sleep(30 * time.Millisecond)
	s.Stop("✓", "done")
	if !strings.Contains(out.String(), "done") {
		t.Fatalf("expected final spinner message in output, got %q", out.String())
	}
}

func TestRunnerRunPhaseEmptyScript(t *testing.T) {
	r := newRunner("", Package{Name: "x"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err := r.RunPhase(context.Background(), "empty", "   "); err != nil {
		t.Fatalf("empty script should be skipped, got %v", err)
	}
}
