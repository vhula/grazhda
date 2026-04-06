package dukh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	icolor "github.com/vhula/grazhda/internal/color"
)

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the dukh workspace monitor in the background",
		RunE:  runDukhStart,
	}
}

func runDukhStart(_ *cobra.Command, _ []string) error {
	dukhBin, err := resolveDukhBinary()
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ "+err.Error()))
		return err
	}

	cmd := exec.Command(dukhBin, "start")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	// SysProcAttr sets the process-group ID so dukh is fully detached
	// from the zgard process and survives zgard exiting.
	setDetach(cmd)

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ failed to start dukh: "+err.Error()))
		return err
	}

	fmt.Printf("%s dukh started (pid %d)\n", icolor.Green("✓"), cmd.Process.Pid)
	return nil
}

// resolveDukhBinary finds the dukh binary. It prefers $GRAZHDA_DIR/bin/dukh
// and falls back to PATH lookup.
func resolveDukhBinary() (string, error) {
	if dir := os.Getenv("GRAZHDA_DIR"); dir != "" {
		candidate := filepath.Join(dir, "bin", "dukh")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	path, err := exec.LookPath("dukh")
	if err != nil {
		return "", fmt.Errorf("dukh binary not found in $GRAZHDA_DIR/bin or $PATH")
	}
	return path, nil
}
