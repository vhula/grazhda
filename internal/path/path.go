package path

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const grazhdaDirVarName = "GRAZHDA_DIR"

// GrazhdaDir returns the Grazhda root directory from $GRAZHDA_DIR,
// defaulting to $HOME/.grazhda.
func GrazhdaDir() (string, error) {
	if dir := os.Getenv(grazhdaDirVarName); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(home, ".grazhda"), nil
}

// ConfigPath returns the resolved path to config.yaml using $GRAZHDA_DIR
// or defaulting to $HOME/.grazhda/config.yaml.
func ConfigPath() string {
	dir, err := GrazhdaDir()
	if err != nil {
		return filepath.Join(".", "config.yaml")
	}
	return filepath.Join(dir, "config.yaml")
}

// DukhBin returns the path to the dukh binary. It checks
// $GRAZHDA_DIR/bin/dukh first, then falls back to PATH lookup.
func DukhBin() string {
	grazhdaDir := os.Getenv("GRAZHDA_DIR")
	if grazhdaDir != "" {
		candidate := grazhdaDir + "/bin/dukh"
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if p, err := exec.LookPath("dukh"); err == nil {
		return p
	}
	return "dukh"
}
