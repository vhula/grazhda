package cfg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
