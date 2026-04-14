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
	verbose    bool
}

// NewInstaller returns a new Installer for the given GRAZHDA_DIR.
// When verbose is true, script stdout/stderr is streamed to out/errOut.
// When false, script output is suppressed and a spinner is shown instead.
func NewInstaller(grazhdaDir string, reg *Registry, out, errOut io.Writer, verbose bool) *Installer {
	return &Installer{grazhdaDir: grazhdaDir, reg: reg, out: out, errOut: errOut, verbose: verbose}
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
	label := PkgLabel(pkg)

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

	runner := newRunner(inst.grazhdaDir, pkg, inst.runnerOut(), inst.runnerErrOut())
	envPath := EnvPath(inst.grazhdaDir)

	// Write pre-install env block and source the env file so the install
	// script sees the exported variables.
	if pkg.PreInstallEnv != "" {
		if err := UpsertBlock(envPath, pkg.Name+":pre", pkg.PreInstallEnv); err != nil {
			return fmt.Errorf("write pre-install-env block for %q: %w", pkg.Name, err)
		}
		fmt.Fprintf(inst.out, "    %s wrote pre-install-env block to %s\n", dimArrow(), envPath)
		if err := runner.RunPhase(ctx, "source env", sourceEnvScript()); err != nil {
			return fmt.Errorf("package %q source env after pre-install-env: %w", pkg.Name, err)
		}
		fmt.Fprintf(inst.out, "    %s sourced %s\n", dimArrow(), envPath)
	}

	// Run the install script. Always prepend a source of .grazhda.env so any
	// variables written by pre_install_env are available to the script.
	if pkg.Install != "" {
		script := sourceEnvScript() + "\n" + pkg.Install
		var spin *Spinner
		if !inst.verbose {
			spin = NewSpinner(inst.errOut, "[install] running…")
		}
		err := runner.RunPhase(ctx, "install", script)
		if !inst.verbose {
			if err != nil {
				spin.Stop(clr.Red("✗"), "[install] failed")
			} else {
				spin.Stop(clr.Green("✓"), "[install] done")
			}
		}
		if err != nil {
			return fmt.Errorf("package %q install: %w", pkg.Name, err)
		}
	}

	// Write post-install env block and source the env file so subsequent
	// packages see the exported variables.
	if pkg.PostInstallEnv != "" {
		if err := UpsertBlock(envPath, pkg.Name+":post", pkg.PostInstallEnv); err != nil {
			return fmt.Errorf("write post-install-env block for %q: %w", pkg.Name, err)
		}
		fmt.Fprintf(inst.out, "    %s wrote post-install-env block to %s\n", dimArrow(), envPath)
		if err := runner.RunPhase(ctx, "source env", sourceEnvScript()); err != nil {
			return fmt.Errorf("package %q source env after post-install-env: %w", pkg.Name, err)
		}
		fmt.Fprintf(inst.out, "    %s sourced %s\n", dimArrow(), envPath)
	}

	fmt.Fprintf(inst.out, "  %s %s installed\n",
		clr.Green("✓"), clr.Green(label))
	return nil
}

// sourceEnvScript returns a bash snippet that sources .grazhda.env when present.
// It is prepended to the install script so pre_install_env variables are
// available, and also run as a standalone phase after each env block write to
// provide visible confirmation that the file was sourced.
func sourceEnvScript() string {
	return `[ -f "$GRAZHDA_DIR/.grazhda.env" ] && source "$GRAZHDA_DIR/.grazhda.env" || true`
}

// runnerOut returns the writer the runner should use for script stdout.
// When not verbose, output is discarded and the spinner provides feedback.
func (inst *Installer) runnerOut() io.Writer {
	if inst.verbose {
		return inst.out
	}
	return io.Discard
}

// runnerErrOut returns the writer the runner should use for script stderr.
func (inst *Installer) runnerErrOut() io.Writer {
	if inst.verbose {
		return inst.errOut
	}
	return io.Discard
}

func dimArrow() string { return "→" }
