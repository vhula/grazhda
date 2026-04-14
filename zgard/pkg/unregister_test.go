package pkg

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/pkgman"
)

func TestUnregisterAll_EmptiesRegistry(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	localPath := pkgman.LocalRegistryPath(dir)
	seed := &pkgman.Registry{
		Packages: []pkgman.Package{
			{Name: "a", Version: "1"},
			{Name: "b"},
		},
	}
	if err := pkgman.SaveRegistry(localPath, seed); err != nil {
		t.Fatalf("seed local registry: %v", err)
	}

	cmd := newUnregisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unregister --all: %v", err)
	}

	reg, err := pkgman.LoadLocalRegistry(localPath)
	if err != nil {
		t.Fatalf("load local: %v", err)
	}
	if len(reg.Packages) != 0 {
		t.Fatalf("expected empty registry, got %d packages", len(reg.Packages))
	}
}

func TestUnregisterByName_RemovesPackage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	localPath := pkgman.LocalRegistryPath(dir)
	seed := &pkgman.Registry{
		Packages: []pkgman.Package{
			{Name: "keep"},
			{Name: "remove"},
		},
	}
	if err := pkgman.SaveRegistry(localPath, seed); err != nil {
		t.Fatalf("seed local registry: %v", err)
	}

	cmd := newUnregisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--name", "remove"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unregister --name remove: %v", err)
	}

	reg, err := pkgman.LoadLocalRegistry(localPath)
	if err != nil {
		t.Fatalf("load local: %v", err)
	}
	if len(reg.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(reg.Packages))
	}
	if reg.Packages[0].Name != "keep" {
		t.Fatalf("expected 'keep' to remain, got %q", reg.Packages[0].Name)
	}
}

func TestUnregisterByNameAndVersion_RemovesExactMatch(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	localPath := pkgman.LocalRegistryPath(dir)
	seed := &pkgman.Registry{
		Packages: []pkgman.Package{
			{Name: "jdk", Version: "17"},
			{Name: "jdk", Version: "21"},
		},
	}
	if err := pkgman.SaveRegistry(localPath, seed); err != nil {
		t.Fatalf("seed local registry: %v", err)
	}

	cmd := newUnregisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--name", "jdk", "--version", "17"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unregister --name jdk --version 17: %v", err)
	}

	reg, err := pkgman.LoadLocalRegistry(localPath)
	if err != nil {
		t.Fatalf("load local: %v", err)
	}
	if len(reg.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(reg.Packages))
	}
	if reg.Packages[0].Version != "21" {
		t.Fatalf("expected version 21 to remain, got %q", reg.Packages[0].Version)
	}
}

func TestUnregisterNotFound_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	localPath := pkgman.LocalRegistryPath(dir)
	if err := pkgman.SaveRegistry(localPath, &pkgman.Registry{}); err != nil {
		t.Fatalf("seed empty registry: %v", err)
	}

	cmd := newUnregisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--name", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing package")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' in error, got: %v", err)
	}
}

func TestUnregisterNoFlags_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	cmd := newUnregisterCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no flags provided")
	}
	if !strings.Contains(err.Error(), "--name") && !strings.Contains(err.Error(), "--all") {
		t.Fatalf("expected error mentioning --name or --all, got: %v", err)
	}
}
