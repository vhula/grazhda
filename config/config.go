package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Dukh struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"dukh"`
	Zgard struct {
		Config map[string]interface{} `yaml:"config"`
	} `yaml:"zgard"`
	General struct {
		InstallDir string `yaml:"install_dir"`
		SourcesDir string `yaml:"sources_dir"`
		BinDir     string `yaml:"bin_dir"`
	} `yaml:"general"`
	Workspaces []WorkspaceConfig `yaml:"workspaces"`
}

type WorkspaceConfig struct {
	Name                 string          `yaml:"name"`
	Default              bool            `yaml:"default,omitempty"`
	Path                 string          `yaml:"path"`
	CloneCommandTemplate string          `yaml:"clone_command_template,omitempty"`
	Projects             []ProjectConfig `yaml:"projects"`
}

type ProjectConfig struct {
	Name         string             `yaml:"name"`
	Subprojects  []SubprojectConfig `yaml:"subprojects,omitempty"`
	Branch       string             `yaml:"branch,omitempty"`
	Repositories []RepositoryConfig `yaml:"repositories,omitempty"`
}

type RepositoryConfig struct {
	Name         string `yaml:"name"`
	LocalDirName string `yaml:"local_dir_name,omitempty"`
}

type SubprojectConfig struct {
	Branch       string             `yaml:"branch"`
	Repositories []RepositoryConfig `yaml:"repositories"`
}

func LoadConfig() (*Config, error) {
	grazhdaDir := os.Getenv("GRAZHDA_DIR")
	if grazhdaDir == "" {
		return nil, fmt.Errorf("GRAZHDA_DIR environment variable not set")
	}
	configPath := filepath.Join(grazhdaDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
