package config_test

import (
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/config"
)

func newTestConfig() *config.Config {
	return &config.Config{
		Editor: "vim",
		Dukh:   config.DukhConfig{Host: "localhost", Port: 9090},
		Workspaces: []config.Workspace{
			{Name: "test-ws", Path: "~/ws"},
		},
	}
}

func TestGetByPath(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name: "scalar string",
			key:  "editor",
			want: "vim",
		},
		{
			name: "scalar int",
			key:  "dukh.port",
			want: "9090",
		},
		{
			name: "nested map",
			key:  "dukh.host",
			want: "localhost",
		},
		{
			name: "array index",
			key:  "workspaces.0.name",
			want: "test-ws",
		},
		{
			name:    "not found",
			key:     "nonexistent",
			wantErr: true,
		},
	}

	cfg := newTestConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.GetByPath(cfg, tt.key)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("GetByPath(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestSerialize(t *testing.T) {
	cfg := newTestConfig()

	data, err := config.Serialize(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "editor: vim") {
		t.Errorf("expected 'editor: vim' in output, got:\n%s", s)
	}
	if !strings.Contains(s, "host: localhost") {
		t.Errorf("expected 'host: localhost' in output, got:\n%s", s)
	}
	if !strings.Contains(s, "port: 9090") {
		t.Errorf("expected 'port: 9090' in output, got:\n%s", s)
	}
}
