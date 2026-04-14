package ws

import (
	"testing"
)

func TestInitCmd_UseField(t *testing.T) {
	cmd := newInitCmd()
	if cmd.Use != "init" {
		t.Fatalf("init Use = %q, want %q", cmd.Use, "init")
	}
}

func TestInitCmd_Flags(t *testing.T) {
	cmd := newInitCmd()
	for _, name := range []string{"dry-run", "verbose", "parallel", "no-confirm", "clone-delay-seconds"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s on init command", name)
		}
	}
}

func TestInitCmd_InheritedTargetingFlags(t *testing.T) {
	parent := NewCmd()
	initCmd, _, err := parent.Find([]string{"init"})
	if err != nil {
		t.Fatalf("find init: %v", err)
	}
	for _, name := range []string{"name", "all", "project-name", "repo-name", "tag"} {
		if initCmd.InheritedFlags().Lookup(name) == nil {
			t.Errorf("expected inherited flag --%s on init command", name)
		}
	}
}

func TestInitCmd_RequiresConfig(t *testing.T) {
	// Reset package-level vars that persist across tests.
	wsName = ""
	wsAll = false

	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	parent := NewCmd()
	parent.SetArgs([]string{"init", "--all"})
	err := parent.Execute()
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
}
