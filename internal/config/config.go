package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

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

// CloneTemplateData holds variables available in clone_command_template.
type CloneTemplateData struct {
	Branch   string
	RepoName string
	DestDir  string
}

// Load reads and parses a Grazhda config file from path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config file %q: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}
	return &cfg, nil
}

// DefaultWorkspace returns the workspace marked as default (default:true or name:"default").
func DefaultWorkspace(cfg *Config) (*Workspace, error) {
	for i := range cfg.Workspaces {
		ws := &cfg.Workspaces[i]
		if ws.Default || ws.Name == "default" {
			return ws, nil
		}
	}
	return nil, fmt.Errorf("no default workspace found: add a workspace with name: default or use --ws")
}

// Validate returns all validation errors in cfg; an empty slice means valid.
func Validate(cfg *Config) []string {
	var errs []string
	seenWS := make(map[string]bool)

	for i, ws := range cfg.Workspaces {
		if ws.Name == "" {
			errs = append(errs, fmt.Sprintf("workspace[%d].name: required field missing", i))
		} else if seenWS[ws.Name] {
			errs = append(errs, fmt.Sprintf("workspace names must be unique: duplicate name %q", ws.Name))
		} else {
			seenWS[ws.Name] = true
		}

		if ws.Path == "" {
			errs = append(errs, fmt.Sprintf("workspace[%d].path: required field missing", i))
		}

		if ws.CloneCommandTemplate == "" {
			errs = append(errs, fmt.Sprintf("workspace[%d].clone_command_template: required field missing", i))
		} else if _, err := template.New("").Parse(ws.CloneCommandTemplate); err != nil {
			errs = append(errs, fmt.Sprintf("workspace[%d].clone_command_template: invalid template: %s", i, err))
		}

		seenProj := make(map[string]bool)
		for j, proj := range ws.Projects {
			if proj.Name == "" {
				errs = append(errs, fmt.Sprintf("workspace[%d].projects[%d].name: required field missing", i, j))
			} else if seenProj[proj.Name] {
				errs = append(errs, fmt.Sprintf("workspace[%d].projects: duplicate project name %q", i, proj.Name))
			} else {
				seenProj[proj.Name] = true
			}

			if proj.Branch == "" {
				errs = append(errs, fmt.Sprintf("workspace[%d].projects[%d].branch: required field missing", i, j))
			}

			for k, repo := range proj.Repositories {
				if repo.Name == "" {
					errs = append(errs, fmt.Sprintf("workspace[%d].projects[%d].repositories[%d].name: required field missing", i, j, k))
				}
			}
		}
	}

	return errs
}

// RenderCloneCmd renders the workspace clone command template for the given project and repo.
// projectPath is the full filesystem path to the project directory.
func RenderCloneCmd(tmplStr, projectPath string, proj Project, repo Repository) (string, error) {
	branch := repo.Branch
	if branch == "" {
		branch = proj.Branch
	}
	destName := repo.LocalDirName
	if destName == "" {
		destName = repo.Name
	}
	data := CloneTemplateData{
		Branch:   branch,
		RepoName: repo.Name,
		DestDir:  filepath.Join(projectPath, destName),
	}
	t, err := template.New("clone").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parsing clone template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering clone template: %w", err)
	}
	return buf.String(), nil
}
