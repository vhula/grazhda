package pkg

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/pkgman"
)

func TestGrazhdaDir_FromEnv(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "/tmp/grazhda-x")
	got, err := grazhdaDir()
	if err != nil {
		t.Fatalf("grazhdaDir error: %v", err)
	}
	if got != "/tmp/grazhda-x" {
		t.Fatalf("grazhdaDir = %q, want /tmp/grazhda-x", got)
	}
}

func TestNewCmd_HasSubcommands(t *testing.T) {
	cmd := NewCmd()
	for _, name := range []string{"install", "purge", "register", "unregister"} {
		if _, _, err := cmd.Find([]string{name}); err != nil {
			t.Fatalf("expected subcommand %q: %v", name, err)
		}
	}
}

func TestPromptDependsOn_SelectsItems(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("1 2\n"))
	var out bytes.Buffer
	reg := &pkgman.Registry{
		Packages: []pkgman.Package{{Name: "a"}, {Name: "b", Version: "1"}},
	}
	deps, err := promptDependsOn(in, &out, reg)
	if err != nil {
		t.Fatalf("promptDependsOn error: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %v", deps)
	}
}

func TestPromptMultiline(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("line1\nline2\n\n"))
	var out bytes.Buffer
	got, err := promptMultiline(in, &out, "install")
	if err != nil {
		t.Fatalf("promptMultiline error: %v", err)
	}
	if got != "line1\nline2" {
		t.Fatalf("promptMultiline = %q", got)
	}
}

func TestLoadMergedRegistry(t *testing.T) {
	dir := t.TempDir()
	global := &pkgman.Registry{
		Packages: []pkgman.Package{{Name: "jdk", Version: "17", Install: "global"}},
	}
	local := &pkgman.Registry{
		Packages: []pkgman.Package{{Name: "jdk", Version: "17", Install: "local"}},
	}
	if err := pkgman.SaveRegistry(pkgman.RegistryPath(dir), global); err != nil {
		t.Fatalf("save global: %v", err)
	}
	if err := pkgman.SaveRegistry(pkgman.LocalRegistryPath(dir), local); err != nil {
		t.Fatalf("save local: %v", err)
	}

	merged, err := loadMergedRegistry(dir)
	if err != nil {
		t.Fatalf("loadMergedRegistry error: %v", err)
	}
	if len(merged.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(merged.Packages))
	}
	if merged.Packages[0].Install != "local" {
		t.Fatalf("expected local override, got %q", merged.Packages[0].Install)
	}
}

func TestUnregisterAll(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	localPath := pkgman.LocalRegistryPath(dir)
	if err := pkgman.SaveRegistry(localPath, &pkgman.Registry{Packages: []pkgman.Package{{Name: "x"}}}); err != nil {
		t.Fatalf("save local: %v", err)
	}

	cmd := newUnregisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unregister --all failed: %v", err)
	}

	reg, err := pkgman.LoadLocalRegistry(localPath)
	if err != nil {
		t.Fatalf("load local: %v", err)
	}
	if len(reg.Packages) != 0 {
		t.Fatalf("expected empty local registry, got %d", len(reg.Packages))
	}
}

func TestRegisterCommand_MissingGlobalRegistryErrors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	cmd := newRegisterCmd()
	cmd.SetIn(strings.NewReader("x\n\nn\n\n\n\n\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when global registry is missing")
	}
	if !strings.Contains(err.Error(), filepath.Base(pkgman.RegistryPath(dir))) {
		t.Fatalf("unexpected error: %v", err)
	}
}
