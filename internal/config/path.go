package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GrazhdaDir returns the Grazhda root directory from $GRAZHDA_DIR,
// defaulting to $HOME/.grazhda.
func GrazhdaDir() (string, error) {
	if dir := os.Getenv("GRAZHDA_DIR"); dir != "" {
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
