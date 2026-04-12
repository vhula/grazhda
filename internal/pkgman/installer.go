package pkgman

import (
	"context"
	"fmt"
	"io"
	"os"

	clr "github.com/vhula/grazhda/internal/color"
)

// Installer orchestrates the installation of packages from the registry.
type Installer struct {
	grazhdaDir string
	reg        *Registry
	out        io.Writer
	errOut     io.Writer
}

// NewInstaller returns a new Installer for the given GRAZHDA_DIR.
func NewInstaller(grazhdaDir string, reg *Registry, out, errOut io.Writer) *Installer {
	return &Installer{grazhdaDir: grazhdaDir, reg: reg, out: out, errOut: errOut}
}

// Install installs the named packages and their transitive dependencies in
// topological order. Pass an empty names slice to install all packages.
func (inst *Installer) Install(ctx context.Context, names []string) error {
	ordered, err := Resolve(inst.reg, names)
	if err != nil {
		return fmt.Errorf("resolve dependencies: %w", err)
	}

	for _, pkg := range ordered {
		if err := inst.installOne(ctx, pkg); err != nil {
			return err
		}
	}
	return nil
}

func (inst *Installer) installOne(ctx context.Context, pkg Package) error {
	label := pkg.Name
	if pkg.Version != "" {
		label = pkg.Name + "@" + pkg.Version
	}

	fmt.Fprintf(inst.out, "\n%s Installing %s\n",
		clr.Blue("▶"), clr.Blue(label))

	// Create package directory before any script runs if requested.
	if pkg.PreCreateDir {
		dir := PkgDir(inst.grazhdaDir, pkg.Name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create package dir for %q: %w", pkg.Name, err)
		}
		fmt.Fprintf(inst.out, "    %s created %s\n", dimArrow(), dir)
	}

	runner := newRunner(inst.grazhdaDir, pkg, inst.out, inst.errOut)

	phases := []struct {
		label  string
		script string
	}{
		{"pre-install", pkg.PreInstall},
		{"install", pkg.Install},
		{"post-install", pkg.PostInstall},
	}

	for _, phase := range phases {
		if phase.script == "" {
			continue
		}
		spin := NewSpinner(inst.errOut, fmt.Sprintf("[%s] running…", phase.label))
		err := runner.RunPhase(ctx, phase.label, phase.script)
		if err != nil {
			spin.Stop(clr.Red("✗"), fmt.Sprintf("[%s] failed", phase.label))
			return fmt.Errorf("package %q %s: %w", pkg.Name, phase.label, err)
		}
		spin.Stop(clr.Green("✓"), fmt.Sprintf("[%s] done", phase.label))
	}

	// Write env block if defined.
	if pkg.Env != "" {
		envPath := EnvPath(inst.grazhdaDir)
		if err := UpsertBlock(envPath, pkg.Name, pkg.Env); err != nil {
			return fmt.Errorf("write env block for %q: %w", pkg.Name, err)
		}
		fmt.Fprintf(inst.out, "    %s wrote env block to %s\n", dimArrow(), envPath)
	}

	fmt.Fprintf(inst.out, "  %s %s installed\n",
		clr.Green("✓"), clr.Green(label))
	return nil
}

func dimArrow() string { return "→" }
