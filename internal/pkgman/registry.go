// Package pkgman implements the declarative package management domain for grazhda.
//
// It handles loading the package registry from .grazhda.pkgs.yaml, resolving
// dependency order via a topological sort, orchestrating install/purge phases
// with injected context, and surgically modifying .grazhda.env using named
// lexical block markers.
package pkgman

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RegistryFile is the canonical registry filename within GRAZHDA_DIR.
const RegistryFile = ".grazhda.pkgs.yaml"

// LocalRegistryFile is the user-managed package registry filename.
const LocalRegistryFile = "registry.pkgs.local.yaml"

// EnvFile is the canonical shell-env filename within GRAZHDA_DIR.
const EnvFile = ".grazhda.env"

// Registry holds the full set of packages declared in .grazhda.pkgs.yaml.
type Registry struct {
	Packages []Package `yaml:"registry"`
}

// Package describes a single installable tool with all of its lifecycle phases.
type Package struct {
	// Name is the unique identifier for this package (e.g. "sdkman", "jdk").
	Name string `yaml:"name"`

	// Version is an optional version string injected as $VERSION into every phase script.
	Version string `yaml:"version,omitempty"`

	// PreCreateDir instructs the installer to create $GRAZHDA_DIR/pkgs/<name>
	// before any phase script runs.
	PreCreateDir bool `yaml:"pre_create_dir"`

	// DependsOn lists package names that must be installed before this one.
	DependsOn []string `yaml:"depends_on,omitempty"`

	// PreInstallEnv holds shell statements written into .grazhda.env (inside a
	// named ":pre" block) before the install script runs. After writing, the env
	// file is sourced so the install script sees the exported variables.
	PreInstallEnv string `yaml:"pre_install_env,omitempty"`

	// Install is the primary installation script.
	Install string `yaml:"install,omitempty"`

	// PostInstallEnv holds shell statements written into .grazhda.env (inside a
	// named ":post" block) after a successful install. After writing, the env
	// file is sourced so subsequent packages see the exported variables.
	PostInstallEnv string `yaml:"post_install_env,omitempty"`

	// Purge is an optional script executed during `zgard pkg purge` before the
	// package directory is removed from disk.
	Purge string `yaml:"purge,omitempty"`
}

// ByName returns the package with the given name, or an error if not found.
func (r *Registry) ByName(name string) (Package, error) {
	for _, p := range r.Packages {
		if p.Name == name {
			return p, nil
		}
	}
	return Package{}, fmt.Errorf("package %q not found in registry", name)
}

// LoadRegistry reads the package registry YAML from path.
func LoadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read registry %q: %w", path, err)
	}
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registry %q: %w", path, err)
	}
	return &reg, nil
}

// LoadLocalRegistry reads the local package registry YAML from path.
// A missing file is treated as an empty registry.
func LoadLocalRegistry(path string) (*Registry, error) {
	reg, err := LoadRegistry(path)
	if err == nil {
		return reg, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return &Registry{}, nil
	}
	return nil, err
}

// SaveRegistry writes the package registry YAML to path.
func SaveRegistry(path string, reg *Registry) error {
	data, err := yaml.Marshal(reg)
	if err != nil {
		return fmt.Errorf("serialize registry %q: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write registry %q: %w", path, err)
	}
	return nil
}

// RegistryPath returns the canonical registry file path for the given grazhdaDir.
func RegistryPath(grazhdaDir string) string {
	return filepath.Join(grazhdaDir, RegistryFile)
}

// LocalRegistryPath returns the canonical local registry path for the given grazhdaDir.
func LocalRegistryPath(grazhdaDir string) string {
	return filepath.Join(grazhdaDir, LocalRegistryFile)
}

// EnvPath returns the canonical .grazhda.env path for the given grazhdaDir.
func EnvPath(grazhdaDir string) string {
	return filepath.Join(grazhdaDir, EnvFile)
}

// PkgDir returns the installation directory for a specific package.
func PkgDir(grazhdaDir, pkgName string) string {
	return filepath.Join(grazhdaDir, "pkgs", pkgName)
}

// MergeRegistries merges global and local registries.
// Local packages override global packages only when name and version both match.
func MergeRegistries(global, local *Registry) *Registry {
	if global == nil {
		global = &Registry{}
	}
	if local == nil {
		local = &Registry{}
	}
	merged := &Registry{
		Packages: make([]Package, 0, len(global.Packages)+len(local.Packages)),
	}
	merged.Packages = append(merged.Packages, global.Packages...)
	for _, lp := range local.Packages {
		found := false
		for i := range merged.Packages {
			if samePkgIdentity(merged.Packages[i], lp) {
				merged.Packages[i] = lp
				found = true
				break
			}
		}
		if !found {
			merged.Packages = append(merged.Packages, lp)
		}
	}
	return merged
}

// AddPackage inserts or updates a package in the registry by name+version identity.
func AddPackage(reg *Registry, pkg Package) *Registry {
	for i := range reg.Packages {
		if samePkgIdentity(reg.Packages[i], pkg) {
			reg.Packages[i] = pkg
			return reg
		}
	}
	reg.Packages = append(reg.Packages, pkg)
	return reg
}

// RemovePackage removes packages from the registry.
// If version is empty, all packages matching name are removed.
// If version is set, only the exact name+version match is removed.
func RemovePackage(reg *Registry, name, version string) (*Registry, bool) {
	if reg == nil {
		return &Registry{}, false
	}
	out := make([]Package, 0, len(reg.Packages))
	removed := false
	for _, p := range reg.Packages {
		if p.Name != name {
			out = append(out, p)
			continue
		}
		if version != "" && p.Version != version {
			out = append(out, p)
			continue
		}
		removed = true
	}
	reg.Packages = out
	return reg, removed
}

func samePkgIdentity(a, b Package) bool {
	return a.Name == b.Name && a.Version == b.Version
}
