package ws

import (
	"github.com/vhula/grazhda/internal/path"
)

// resolveConfigPath delegates to the shared config.ConfigPath helper.
func resolveConfigPath() string { return path.ConfigPath() }
