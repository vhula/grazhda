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
}

// NewPurger returns a new Purger for the given GRAZHDA_DIR.
func NewPurger(grazhdaDir string, reg *Registry, out, errOut io.Writer) *Purger {
	return &Purger{grazhdaDir: grazhdaDir, reg: reg, out: out, errOut: errOut}
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
	label := pkg.Name
	if pkg.Version != "" {
		label = pkg.Name + "@" + pkg.Version
	}

	fmt.Fprintf(p.out, "\n%s Purging %s\n",
		clr.Yellow("▶"), clr.Yellow(label))

	runner := newRunner(p.grazhdaDir, pkg, p.out, p.errOut)

	// Run the optional purge script first.
	if pkg.Purge != "" {
		spin := NewSpinner(p.errOut, "[purge] running…")
		if err := runner.RunPhase(ctx, "purge", pkg.Purge); err != nil {
			spin.Stop(clr.Red("✗"), "[purge] failed")
			return fmt.Errorf("package %q purge script: %w", pkg.Name, err)
		}
		spin.Stop(clr.Green("✓"), "[purge] done")
	}

	// Remove pkg directory if it was pre-created.
	if pkg.PreCreateDir {
		dir := PkgDir(p.grazhdaDir, pkg.Name)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove package dir %q: %w", dir, err)
		}
		fmt.Fprintf(p.out, "    %s removed %s\n", dimArrow(), dir)
	}

	// Excise env block from .grazhda.env.
	envPath := EnvPath(p.grazhdaDir)
	if removed, err := removeBlockIfPresent(envPath, pkg.Name); err != nil {
		return fmt.Errorf("remove env block for %q: %w", pkg.Name, err)
	} else if removed {
		fmt.Fprintf(p.out, "    %s removed env block from %s\n", dimArrow(), envPath)
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
