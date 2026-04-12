package pkgman

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Runner executes phase scripts for a specific package with a fully-populated
// environment and streams every output line to the terminal with a styled prefix.
type Runner struct {
	grazhdaDir string
	pkg        Package
	out        io.Writer
	errOut     io.Writer
}

func newRunner(grazhdaDir string, pkg Package, out, errOut io.Writer) *Runner {
	return &Runner{grazhdaDir: grazhdaDir, pkg: pkg, out: out, errOut: errOut}
}

// RunPhase executes a shell script under the given phase label.
// Output lines are forwarded to out/errOut with a visual indent.
// An empty script is silently skipped.
func (r *Runner) RunPhase(ctx context.Context, phase, script string) error {
	if strings.TrimSpace(script) == "" {
		return nil
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", script) //nolint:gosec
	cmd.Env = r.buildEnv()
	cmd.Dir = r.grazhdaDir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe for %s/%s: %w", r.pkg.Name, phase, err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe for %s/%s: %w", r.pkg.Name, phase, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s/%s: %w", r.pkg.Name, phase, err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		streamLines(stdoutPipe, "    │ ", r.out)
	}()
	streamLines(stderrPipe, "    │ ", r.errOut)
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("phase %s for package %q failed: %w", phase, r.pkg.Name, err)
	}
	return nil
}

// buildEnv returns the environment for phase script execution.
// It inherits the current process env and overlays grazhda-specific variables.
func (r *Runner) buildEnv() []string {
	pkgDir := PkgDir(r.grazhdaDir, r.pkg.Name)
	overlay := []string{
		"GRAZHDA_DIR=" + r.grazhdaDir,
		"PKG_DIR=" + pkgDir,
		"PKG_PREFIX=" + pkgDir,
		"PKG_NAME=" + r.pkg.Name,
		"VERSION=" + r.pkg.Version,
		// Source .grazhda.env so dependencies' exports are already active.
		// (Downstream scripts that call `source $GRAZHDA_DIR/.grazhda.env` will
		// pick up e.g. SDKMAN_DIR set by a previously installed package.)
	}
	// Replace any existing keys in os.Environ() so overlays win.
	base := os.Environ()
	overrideSet := make(map[string]string, len(overlay))
	for _, kv := range overlay {
		k, _, _ := strings.Cut(kv, "=")
		overrideSet[k] = kv
	}
	merged := make([]string, 0, len(base)+len(overlay))
	for _, kv := range base {
		k, _, _ := strings.Cut(kv, "=")
		if replacement, ok := overrideSet[k]; ok {
			merged = append(merged, replacement)
			delete(overrideSet, k)
		} else {
			merged = append(merged, kv)
		}
	}
	for _, kv := range overrideSet {
		merged = append(merged, kv)
	}
	return merged
}

// streamLines reads from r line-by-line and writes each line prefixed to w.
func streamLines(r io.Reader, prefix string, w io.Writer) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fmt.Fprintf(w, "%s%s\n", prefix, sc.Text())
	}
}
