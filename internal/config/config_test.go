package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()

	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent, err := os.ReadFile("../../config.template.yaml")
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(configPath, yamlContent, 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Dukh.Host != "localhost" {
		t.Errorf("Expected Dukh.Host to be 'localhost', got '%s'", cfg.Dukh.Host)
	}
	if cfg.Dukh.Port != 50501 {
		t.Errorf("Expected Dukh.Port to be 50501, got %d", cfg.Dukh.Port)
	}
	if cfg.Zgard.Config == nil {
		t.Error("Expected Zgard.Config to be initialized")
	}
	if cfg.General.InstallDir != "${GRAZHDA_DIR}" {
		t.Errorf("Expected General.InstallDir to be '${GRAZHDA_DIR}', got '%s'", cfg.General.InstallDir)
	}
	if cfg.General.SourcesDir != "${GRAZHDA_DIR}/sources" {
		t.Errorf("Expected General.SourcesDir to be '${GRAZHDA_DIR}/sources', got '%s'", cfg.General.SourcesDir)
	}
	if cfg.General.BinDir != "${GRAZHDA_DIR}/bin" {
		t.Errorf("Expected General.BinDir to be '${GRAZHDA_DIR}/bin', got '%s'", cfg.General.BinDir)
	}
	if len(cfg.Workspaces) != 2 {
		t.Errorf("Expected 2 workspaces, got %d", len(cfg.Workspaces))
	}

	ws1 := cfg.Workspaces[0]
	if ws1.Name != "default" {
		t.Errorf("Expected first workspace name to be 'default', got '%s'", ws1.Name)
	}
	if !ws1.Default {
		t.Error("Expected first workspace to be default")
	}
	if ws1.Path != "/home/jake/ws" {
		t.Errorf("Expected first workspace path to be '/home/jake/ws', got '%s'", ws1.Path)
	}
	if ws1.CloneCommandTemplate != "git clone --branch {{.Branch}} https://github.com/grazhda/{{.RepoName}} {{.DestDir}}" {
		t.Errorf("Unexpected CloneCommandTemplate for first workspace")
	}
	if len(ws1.Projects) != 2 {
		t.Errorf("Expected 2 projects in first workspace, got %d", len(ws1.Projects))
	}

	ws2 := cfg.Workspaces[1]
	if ws2.Name != "secondary" {
		t.Errorf("Expected second workspace name to be 'secondary', got '%s'", ws2.Name)
	}
	if ws2.Default {
		t.Error("Expected second workspace not to be default")
	}
	if ws2.Path != "/home/jake/secondary_ws" {
		t.Errorf("Expected second workspace path to be '/home/jake/secondary_ws', got '%s'", ws2.Path)
	}
	if ws2.CloneCommandTemplate != "git clone --branch {{.Branch}} https://github.com/grazhda/{{.RepoName}} {{.DestDir}}" {
		t.Errorf("Unexpected CloneCommandTemplate for second workspace")
	}
	if len(ws2.Projects) != 1 {
		t.Errorf("Expected 1 project in second workspace, got %d", len(ws2.Projects))
	}
}

func TestLoadConfig_NoEnv(t *testing.T) {
	os.Unsetenv("GRAZHDA_DIR")

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error when GRAZHDA_DIR is not set")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error when config.yaml does not exist")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()

	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	configPath := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = LoadConfig()
	if err == nil {
		t.Error("Expected error when YAML is invalid")
	}
}
