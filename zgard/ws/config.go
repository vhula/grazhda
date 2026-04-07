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
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		dir = filepath.Join(home, ".grazhda")
	}
	return filepath.Join(dir, "config.yaml")
}
