package pkgman

import (
	"context"
	"fmt"
	"io"
	"os"

	clr "github.com/vhula/grazhda/internal/color"
)

// Purger orchestrates the removal of packages from the local environment.
type Purger struct {
	grazhdaDir string
	reg        *Registry
	out        io.Writer
	errOut     io.Writer
	verbose    bool
}

// NewPurger returns a new Purger for the given GRAZHDA_DIR.
// When verbose is true, script stdout/stderr is streamed to out/errOut.
// When false, script output is suppressed and a spinner is shown instead.
func NewPurger(grazhdaDir string, reg *Registry, out, errOut io.Writer, verbose bool) *Purger {
	return &Purger{grazhdaDir: grazhdaDir, reg: reg, out: out, errOut: errOut, verbose: verbose}
}

// Purge removes the named packages in reverse topological order so dependents
// are removed before their dependencies. Pass an empty names slice to purge all.
func (p *Purger) Purge(ctx context.Context, names []string) error {
	ordered, err := ResolveReverse(p.reg, names)
	if err != nil {
		return fmt.Errorf("resolve purge order: %w", err)
	}

	for _, pkg := range ordered {
		if err := p.purgeOne(ctx, pkg); err != nil {
			return err
		}
	}
	return nil
}

func (p *Purger) purgeOne(ctx context.Context, pkg Package) error {
	label := PkgLabel(pkg)

	fmt.Fprintf(p.out, "\nPurging %s\n", clr.Yellow(label))

	runner := newRunner(p.grazhdaDir, pkg, p.runnerOut(), p.runnerErrOut())

	// Run the optional purge script first.
	if pkg.Purge != "" {
		var spin *Spinner
		if !p.verbose {
			spin = NewSpinner(p.errOut, "[purge] running…")
		}
		err := runner.RunPhase(ctx, "purge", pkg.Purge)
		if !p.verbose {
			if err != nil {
				spin.Stop(clr.Red("✗"), "[purge] failed")
			} else {
				spin.Stop(clr.Green("✓"), "[purge] done")
			}
		}
		if err != nil {
			return fmt.Errorf("package %q purge script: %w", pkg.Name, err)
		}
	}

	// Remove pkg directory if it was pre-created.
	if pkg.PreCreateDir {
		dir := PkgDir(p.grazhdaDir, pkg.Name)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove package dir %q: %w", dir, err)
		}
		fmt.Fprintf(p.out, "    %s removed %s\n", dimArrow(), dir)
	}

	// Excise both env blocks from .grazhda.env.
	envPath := EnvPath(p.grazhdaDir)
	for _, blockName := range []string{pkg.Name + ":pre", pkg.Name + ":post"} {
		if removed, err := removeBlockIfPresent(envPath, blockName); err != nil {
			return fmt.Errorf("remove env block %q for package %q: %w", blockName, pkg.Name, err)
		} else if removed {
			fmt.Fprintf(p.out, "    %s removed env block %q from %s\n", dimArrow(), blockName, envPath)
		}
	}

	fmt.Fprintf(p.out, "  %s %s purged\n",
		clr.Green("✓"), clr.Green(label))
	return nil
}

// removeBlockIfPresent removes the named env block and reports whether it existed.
func removeBlockIfPresent(path, name string) (bool, error) {
	present, err := HasBlock(path, name)
	if err != nil || !present {
		return false, err
	}
	return true, RemoveBlock(path, name)
}

func (p *Purger) runnerOut() io.Writer {
	if p.verbose {
		return p.out
	}
	return io.Discard
}

func (p *Purger) runnerErrOut() io.Writer {
	if p.verbose {
		return p.errOut
	}
	return io.Discard
}
