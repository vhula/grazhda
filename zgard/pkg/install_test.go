package pkg

import (
	"bytes"
	"strings"
	"testing"
)

func TestInstallRequiresFlag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	cmd := newInstallCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no --name or --all provided")
	}
	if !strings.Contains(err.Error(), "--name") && !strings.Contains(err.Error(), "--all") {
		t.Fatalf("expected flag hint in error, got: %v", err)
	}
}

func TestInstallAcceptsVerboseFlag(t *testing.T) {
	cmd := newInstallCmd()
	f := cmd.Flags().Lookup("verbose")
	if f == nil {
		t.Fatal("expected --verbose flag to be defined")
	}
	if f.Shorthand != "v" {
		t.Fatalf("expected --verbose shorthand 'v', got %q", f.Shorthand)
	}
}

func TestInstallAcceptsAllFlag(t *testing.T) {
	cmd := newInstallCmd()
	f := cmd.Flags().Lookup("all")
	if f == nil {
		t.Fatal("expected --all flag to be defined")
	}
}

func TestInstallFailsWithoutRegistry(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	cmd := newInstallCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--all"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when registry files are missing")
	}
	if !strings.Contains(err.Error(), "registry") {
		t.Fatalf("expected 'registry' in error, got: %v", err)
	}
}

func TestInstallFailsWithoutGrazhdaDir(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "")
	t.Setenv("HOME", t.TempDir())

	cmd := newInstallCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--all"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when GRAZHDA_DIR is unset and fallback dir doesn't exist")
	}
}
