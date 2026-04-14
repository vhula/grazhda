package ws

import "github.com/vhula/grazhda/internal/config"

// resolveConfigPath delegates to the shared config.ConfigPath helper.
func resolveConfigPath() string { return config.ConfigPath() }
