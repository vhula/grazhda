package pkgman_test

import (
	"path/filepath"
	"testing"

	"github.com/vhula/grazhda/internal/pkgman"
)

func TestRegistryPath(t *testing.T) {
	got := pkgman.RegistryPath("/home/user/.grazhda")
	want := filepath.Join("/home/user/.grazhda", pkgman.RegistryFile)
	if got != want {
		t.Fatalf("RegistryPath = %q, want %q", got, want)
	}
}

func TestLocalRegistryPath(t *testing.T) {
	got := pkgman.LocalRegistryPath("/home/user/.grazhda")
	want := filepath.Join("/home/user/.grazhda", pkgman.LocalRegistryFile)
	if got != want {
		t.Fatalf("LocalRegistryPath = %q, want %q", got, want)
	}
}

func TestEnvPath(t *testing.T) {
	got := pkgman.EnvPath("/home/user/.grazhda")
	want := filepath.Join("/home/user/.grazhda", pkgman.EnvFile)
	if got != want {
		t.Fatalf("EnvPath = %q, want %q", got, want)
	}
}

func TestPkgDir(t *testing.T) {
	got := pkgman.PkgDir("/home/user/.grazhda", "jdk")
	want := filepath.Join("/home/user/.grazhda", "pkgs", "jdk")
	if got != want {
		t.Fatalf("PkgDir = %q, want %q", got, want)
	}
}

func TestSaveAndLoadRegistry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, pkgman.RegistryFile)

	original := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "sdkman", Install: "curl -s https://sdkman.io | bash"},
		{Name: "jdk", Version: "17", DependsOn: []string{"sdkman"}, PreCreateDir: true},
	}}

	if err := pkgman.SaveRegistry(path, original); err != nil {
		t.Fatalf("SaveRegistry: %v", err)
	}

	loaded, err := pkgman.LoadRegistry(path)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	if len(loaded.Packages) != len(original.Packages) {
		t.Fatalf("expected %d packages, got %d", len(original.Packages), len(loaded.Packages))
	}
	for i, want := range original.Packages {
		got := loaded.Packages[i]
		if got.Name != want.Name || got.Version != want.Version || got.Install != want.Install || got.PreCreateDir != want.PreCreateDir {
			t.Errorf("package[%d] mismatch: got %+v, want %+v", i, got, want)
		}
	}
}

func TestLoadRegistry_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.yaml")

	_, err := pkgman.LoadRegistry(path)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestByName_Found(t *testing.T) {
	reg := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "sdkman", Install: "install-sdkman"},
		{Name: "jdk", Version: "17", Install: "install-jdk"},
	}}

	pkg, err := reg.ByName("jdk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Name != "jdk" || pkg.Version != "17" {
		t.Fatalf("expected jdk@17, got %s@%s", pkg.Name, pkg.Version)
	}
}

func TestByName_NotFound(t *testing.T) {
	reg := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "sdkman"},
	}}

	_, err := reg.ByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing package, got nil")
	}
}

func TestAddPackage_New(t *testing.T) {
	reg := &pkgman.Registry{}
	pkg := pkgman.Package{Name: "jdk", Version: "17", Install: "install-jdk"}

	result := pkgman.AddPackage(reg, pkg)
	if len(result.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(result.Packages))
	}
	if result.Packages[0].Name != "jdk" {
		t.Fatalf("expected jdk, got %s", result.Packages[0].Name)
	}
}

func TestAddPackage_Update(t *testing.T) {
	reg := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "jdk", Version: "17", Install: "old-install"},
	}}
	updated := pkgman.Package{Name: "jdk", Version: "17", Install: "new-install"}

	result := pkgman.AddPackage(reg, updated)
	if len(result.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(result.Packages))
	}
	if result.Packages[0].Install != "new-install" {
		t.Fatalf("expected updated install, got %q", result.Packages[0].Install)
	}
}

func TestRemovePackage_NotFound(t *testing.T) {
	reg := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "sdkman"},
	}}

	_, removed := pkgman.RemovePackage(reg, "nonexistent", "")
	if removed {
		t.Fatal("expected removed=false for nonexistent package")
	}
}

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
