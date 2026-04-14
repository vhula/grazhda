package pkgman

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunPhase_CapturesStdout(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	r := newRunner(dir, Package{Name: "test"}, &out, &errOut)

	err := r.RunPhase(context.Background(), "greet", `echo hello world`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "hello world") {
		t.Errorf("stdout should contain script output, got %q", out.String())
	}
}

func TestRunPhase_PrefixesOutputLines(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	r := newRunner(dir, Package{Name: "test"}, &out, &errOut)

	err := r.RunPhase(context.Background(), "multi", "echo line1\necho line2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, line := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		if !strings.HasPrefix(line, "      ") {
			t.Errorf("line should be prefixed with 6-space indent, got %q", line)
		}
	}
}

func TestRunPhase_StreamsStderr(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	r := newRunner(dir, Package{Name: "test"}, &out, &errOut)

	err := r.RunPhase(context.Background(), "warn", `echo warning >&2`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errOut.String(), "warning") {
		t.Errorf("stderr should contain warning, got %q", errOut.String())
	}
}

func TestRunPhase_ErrorOnNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	r := newRunner(dir, Package{Name: "fail-pkg"}, &out, &errOut)

	err := r.RunPhase(context.Background(), "bad", `exit 42`)
	if err == nil {
		t.Fatal("expected error for non-zero exit code")
	}
	if !strings.Contains(err.Error(), "fail-pkg") {
		t.Errorf("error should mention package name, got %q", err.Error())
	}
}

func TestRunPhase_EnvOverlay(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	p := Package{Name: "mypkg", Version: "2.5"}
	r := newRunner(dir, p, &out, &errOut)

	err := r.RunPhase(context.Background(), "env", `echo "DIR=$GRAZHDA_DIR VER=$VERSION NAME=$PKG_NAME"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "DIR="+dir) {
		t.Errorf("GRAZHDA_DIR not set correctly in output: %q", got)
	}
	if !strings.Contains(got, "VER=2.5") {
		t.Errorf("VERSION not set correctly in output: %q", got)
	}
	if !strings.Contains(got, "NAME=mypkg") {
		t.Errorf("PKG_NAME not set correctly in output: %q", got)
	}
}

func TestRunPhase_PKG_DIR(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	r := newRunner(dir, Package{Name: "tool"}, &out, &errOut)

	err := r.RunPhase(context.Background(), "dir", `echo "PKG=$PKG_DIR"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := PkgDir(dir, "tool")
	if !strings.Contains(out.String(), "PKG="+expected) {
		t.Errorf("PKG_DIR not set correctly, got %q, want substring PKG=%s", out.String(), expected)
	}
}

func TestRunPhase_CancelledContext(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	r := newRunner(dir, Package{Name: "slow"}, &out, &errOut)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := r.RunPhase(ctx, "cancel", `sleep 10`)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
