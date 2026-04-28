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

func TestSubcommandsHaveLongDescriptions(t *testing.T) {
	cmds := []struct {
		name string
		long string
	}{
		{"start", startCmd().Long},
		{"stop", stopCmd().Long},
		{"status", statusCmd().Long},
		{"scan", scanCmd().Long},
	}
	for _, c := range cmds {
		if c.long == "" {
			t.Errorf("%s: expected non-empty Long description", c.name)
		}
	}
}

func TestRootHasLongAndVersion(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("root command should have a Long description")
	}
	if rootCmd.Version == "" {
		t.Error("root command should have a Version")
	}
}

func TestNoColorFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("no-color")
	if f == nil {
		t.Fatal("--no-color flag not registered")
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

func TestRunStartDaemonModeRequiresConfig(t *testing.T) {
	t.Setenv("DUKH_DAEMON", "1")
	t.Setenv("GRAZHDA_DIR", t.TempDir())
	err := runStart(nil, nil)
	if err == nil {
		t.Fatal("expected error when config.yaml is missing")
	}
}

func TestTryGetUptime_NoServer(t *testing.T) {
	if got := tryGetUptime(); got != "" {
		t.Fatalf("expected empty uptime without server, got %q", got)
	}
}

func TestStopCmd_HasCorrectUse(t *testing.T) {
	cmd := stopCmd()
	if cmd.Use != "stop" {
		t.Fatalf("stopCmd().Use = %q, want %q", cmd.Use, "stop")
	}
}

func TestScanCmd_HasCorrectUse(t *testing.T) {
	cmd := scanCmd()
	if cmd.Use != "scan" {
		t.Fatalf("scanCmd().Use = %q, want %q", cmd.Use, "scan")
	}
}

func TestStatusCmd_HasCorrectUse(t *testing.T) {
	cmd := statusCmd()
	if cmd.Use != "status" {
		t.Fatalf("statusCmd().Use = %q, want %q", cmd.Use, "status")
	}
}

func TestErrsHelper(t *testing.T) {
	// Verify printErr doesn't panic when called with a test error.
	printErr("test error message")
}
