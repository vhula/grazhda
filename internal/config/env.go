package config

import (
	"os"
	"strconv"
)

// ApplyEnvOverrides overlays environment variable values onto cfg after loading
// from the config file. Environment variables take precedence over file settings.
//
// Supported variables:
//
//	GRAZHDA_EDITOR   — overrides cfg.Editor
//	DUKH_HOST        — overrides cfg.Dukh.Host
//	DUKH_PORT        — overrides cfg.Dukh.Port (must be a valid integer)
func ApplyEnvOverrides(cfg *Config) {
	if v := os.Getenv("GRAZHDA_EDITOR"); v != "" {
		cfg.Editor = v
	}
	if v := os.Getenv("DUKH_HOST"); v != "" {
		cfg.Dukh.Host = v
	}
	if v := os.Getenv("DUKH_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Dukh.Port = p
		}
	}
}
