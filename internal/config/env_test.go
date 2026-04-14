package config_test

import (
	"testing"

	"github.com/vhula/grazhda/internal/config"
)

func TestApplyEnvOverrides(t *testing.T) {
	tests := []struct {
		name       string
		envVars    map[string]string
		wantEditor string
		wantHost   string
		wantPort   int
	}{
		{
			name:       "editor override",
			envVars:    map[string]string{"GRAZHDA_EDITOR": "nano"},
			wantEditor: "nano",
			wantHost:   "original-host",
			wantPort:   8080,
		},
		{
			name:       "dukh host override",
			envVars:    map[string]string{"DUKH_HOST": "remotehost"},
			wantEditor: "vim",
			wantHost:   "remotehost",
			wantPort:   8080,
		},
		{
			name:       "dukh port override valid int",
			envVars:    map[string]string{"DUKH_PORT": "9090"},
			wantEditor: "vim",
			wantHost:   "original-host",
			wantPort:   9090,
		},
		{
			name:       "dukh port override invalid int",
			envVars:    map[string]string{"DUKH_PORT": "notanumber"},
			wantEditor: "vim",
			wantHost:   "original-host",
			wantPort:   8080,
		},
		{
			name:       "no env vars set",
			envVars:    map[string]string{},
			wantEditor: "vim",
			wantHost:   "original-host",
			wantPort:   8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars first.
			t.Setenv("GRAZHDA_EDITOR", "")
			t.Setenv("DUKH_HOST", "")
			t.Setenv("DUKH_PORT", "")

			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg := &config.Config{
				Editor: "vim",
				Dukh:   config.DukhConfig{Host: "original-host", Port: 8080},
			}

			config.ApplyEnvOverrides(cfg)

			if cfg.Editor != tt.wantEditor {
				t.Errorf("Editor = %q, want %q", cfg.Editor, tt.wantEditor)
			}
			if cfg.Dukh.Host != tt.wantHost {
				t.Errorf("Dukh.Host = %q, want %q", cfg.Dukh.Host, tt.wantHost)
			}
			if cfg.Dukh.Port != tt.wantPort {
				t.Errorf("Dukh.Port = %d, want %d", cfg.Dukh.Port, tt.wantPort)
			}
		})
	}
}
