package cfg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveConfigPath_UsesGrazhdaDir(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "/tmp/grazhda-test")
	got := resolveConfigPath()
	want := filepath.Join("/tmp/grazhda-test", "config.yaml")
	if got != want {
		t.Fatalf("resolveConfigPath = %q, want %q", got, want)
	}
}

func TestNewCmd_HasSubcommands(t *testing.T) {
	cmd := NewCmd()
	for _, name := range []string{"path", "validate", "list", "get"} {
		if _, _, err := cmd.Find([]string{name}); err != nil {
			t.Fatalf("expected subcommand %q: %v", name, err)
		}
	}
}

func TestPathCommand_PrintsResolvedPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	cmd := newPathCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute path cmd: %v", err)
	}

	got := strings.TrimSpace(out.String())
	want := filepath.Join(dir, "config.yaml")
	if got != want {
		t.Fatalf("path output = %q, want %q", got, want)
	}
}

func TestResolveConfigPath_FallbackUsesHome(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtimeHome := os.Getenv("HOME"); runtimeHome != home {
		t.Fatalf("expected HOME=%q, got %q", home, runtimeHome)
	}
	got := resolveConfigPath()
	want := filepath.Join(home, ".grazhda", "config.yaml")
	if got != want {
		t.Fatalf("resolveConfigPath fallback = %q, want %q", got, want)
	}
}

func findSubcommand(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	t.Fatalf("subcommand %q not found", name)
	return nil
}

func TestValidateCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "validate")
	if cmd.Use != "validate" {
		t.Fatalf("validate Use = %q, want %q", cmd.Use, "validate")
	}
}

func TestListCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "list")
	if cmd.Use != "list" {
		t.Fatalf("list Use = %q, want %q", cmd.Use, "list")
	}
}

func TestGetCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "get")
	if cmd.Use != "get <key>" {
		t.Fatalf("get Use = %q, want %q", cmd.Use, "get <key>")
	}
}

func TestGetCmd_RequiresArg(t *testing.T) {
	parent := NewCmd()
	parent.SetArgs([]string{"get"})
	var out bytes.Buffer
	parent.SetOut(&out)
	parent.SetErr(&out)
	err := parent.Execute()
	if err == nil {
		t.Fatal("expected error when get is called with no args")
	}
}
