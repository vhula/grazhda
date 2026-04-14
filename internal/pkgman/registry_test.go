package pkgman_test

import (
	"path/filepath"
	"testing"

	"github.com/vhula/grazhda/internal/pkgman"
)

func TestMerge_LocalOverridesGlobal(t *testing.T) {
	global := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "jdk", Version: "17", Install: "global"},
	}}
	local := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "jdk", Version: "17", Install: "local"},
	}}

	merged := pkgman.MergeRegistries(global, local)
	if len(merged.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(merged.Packages))
	}
	if merged.Packages[0].Install != "local" {
		t.Fatalf("expected local override, got install=%q", merged.Packages[0].Install)
	}
}

func TestMerge_DifferentVersionsCoexist(t *testing.T) {
	global := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "jdk", Version: "17"},
	}}
	local := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "jdk", Version: "21"},
	}}

	merged := pkgman.MergeRegistries(global, local)
	if len(merged.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(merged.Packages))
	}
}

func TestMerge_UnversionedOverride(t *testing.T) {
	global := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "sdkman", Install: "global"},
	}}
	local := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "sdkman", Install: "local"},
	}}

	merged := pkgman.MergeRegistries(global, local)
	if len(merged.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(merged.Packages))
	}
	if merged.Packages[0].Install != "local" {
		t.Fatalf("expected local override, got install=%q", merged.Packages[0].Install)
	}
}

func TestRemovePackage_ByName(t *testing.T) {
	reg := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "jdk", Version: "17"},
		{Name: "jdk", Version: "21"},
		{Name: "sdkman"},
	}}

	updated, removed := pkgman.RemovePackage(reg, "jdk", "")
	if !removed {
		t.Fatal("expected removal=true")
	}
	if len(updated.Packages) != 1 || updated.Packages[0].Name != "sdkman" {
		t.Fatalf("expected only sdkman left, got %#v", updated.Packages)
	}
}

func TestRemovePackage_ByNameVersion(t *testing.T) {
	reg := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "jdk", Version: "17"},
		{Name: "jdk", Version: "21"},
	}}

	updated, removed := pkgman.RemovePackage(reg, "jdk", "21")
	if !removed {
		t.Fatal("expected removal=true")
	}
	if len(updated.Packages) != 1 {
		t.Fatalf("expected 1 package left, got %d", len(updated.Packages))
	}
	if updated.Packages[0].Version != "17" {
		t.Fatalf("expected version 17 left, got %q", updated.Packages[0].Version)
	}
}

func TestLoadLocalRegistry_MissingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.pkgs.local.yaml")

	reg, err := pkgman.LoadLocalRegistry(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Packages) != 0 {
		t.Fatalf("expected empty registry, got %d packages", len(reg.Packages))
	}
}
