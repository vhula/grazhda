package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommandFactories(t *testing.T) {
	for _, c := range []struct {
		name string
		use  string
	}{
		{"start", startCmd().Use},
		{"stop", stopCmd().Use},
		{"status", statusCmd().Use},
		{"scan", scanCmd().Use},
	} {
		if c.use == "" {
			t.Fatalf("expected non-empty Use for %s", c.name)
		}
	}
}

func TestReadPIDFile(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "run")
	if err := os.MkdirAll(pidPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidPath, "dukh.pid"), []byte("1234\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pid, err := readPIDFile(dir)
	if err != nil {
		t.Fatalf("readPIDFile failed: %v", err)
	}
	if pid != 1234 {
		t.Fatalf("pid=%d, want 1234", pid)
	}
}

func TestRunStartRequiresEnvInLauncherMode(t *testing.T) {
	t.Setenv("DUKH_DAEMON", "")
	t.Setenv("GRAZHDA_DIR", "")
	err := runStart(nil, nil)
	if err == nil {
		t.Fatal("expected error when GRAZHDA_DIR is unset")
	}
}

func TestTryGetUptime_NoServer(t *testing.T) {
	if got := tryGetUptime(); got != "" {
		t.Fatalf("expected empty uptime without server, got %q", got)
	}
}
