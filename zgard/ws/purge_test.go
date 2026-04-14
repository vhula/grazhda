package ws

import (
	"strings"
	"testing"
)

func TestPurgeCmd_UseField(t *testing.T) {
	cmd := newPurgeCmd()
	if cmd.Use != "purge" {
		t.Fatalf("purge Use = %q, want %q", cmd.Use, "purge")
	}
}

func TestPurgeCmd_Flags(t *testing.T) {
	cmd := newPurgeCmd()
	for _, name := range []string{"dry-run", "verbose", "no-confirm"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s on purge command", name)
		}
	}
}

func TestPurgeCmd_RequiresExplicitTarget(t *testing.T) {
	// Reset package-level targeting vars.
	wsName = ""
	wsAll = false

	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	parent := NewCmd()
	parent.SetArgs([]string{"purge"})
	err := parent.Execute()
	if err == nil {
		t.Fatal("expected error when no --name or --all provided")
	}
	if !strings.Contains(err.Error(), "--name") && !strings.Contains(err.Error(), "--all") {
		t.Fatalf("expected error mentioning --name or --all, got: %v", err)
	}
}

func TestPurgeCmd_AllFlagAccepted(t *testing.T) {
	// Reset package-level targeting vars.
	wsName = ""
	wsAll = false

	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	parent := NewCmd()
	parent.SetArgs([]string{"purge", "--all"})
	err := parent.Execute()
	// Should fail with a config error (not a flag error) when --all is provided.
	if err == nil {
		t.Fatal("expected error (config missing), but got nil")
	}
	if strings.Contains(err.Error(), "requires --name") {
		t.Fatalf("should not get targeting error with --all, got: %v", err)
	}
}
