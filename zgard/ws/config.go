package ws

import (
	"os"
	"path/filepath"
)

// resolveConfigPath returns the path to config.yaml, using GRAZHDA_DIR env var
// or defaulting to $HOME/.grazhda/config.yaml.
func resolveConfigPath() string {
	dir := os.Getenv("GRAZHDA_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".grazhda")
	}
	return filepath.Join(dir, "config.yaml")
}
