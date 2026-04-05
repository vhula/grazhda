package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level Grazhda configuration.
type Config struct {
	Workspaces []Workspace `yaml:"workspaces"`
}

// Workspace represents a named collection of projects.
type Workspace struct {
	Name                 string    `yaml:"name"`
	Default              bool      `yaml:"default"`
	Path                 string    `yaml:"path"`
	CloneCommandTemplate string    `yaml:"clone_command_template"`
	Projects             []Project `yaml:"projects"`
}

// Project groups repositories under a common branch.
type Project struct {
	Name         string       `yaml:"name"`
	Branch       string       `yaml:"branch"`
	Repositories []Repository `yaml:"repositories"`
}

// Repository is a single repository entry within a project.
type Repository struct {
	Name         string `yaml:"name"`
	Branch       string `yaml:"branch,omitempty"`
	LocalDirName string `yaml:"local_dir_name,omitempty"`
}

// Load reads and parses a Grazhda config file from path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
