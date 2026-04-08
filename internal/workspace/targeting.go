package workspace

import (
	"fmt"

	"github.com/vhula/grazhda/internal/config"
)

// ValidateFilters returns an error if opts.ProjectName or opts.RepoName
// does not match any entry in the given workspace's configuration.
// Validates config structure only — filesystem presence is not checked here.
func ValidateFilters(ws config.Workspace, opts RunOptions) error {
	if opts.ProjectName == "" {
		return nil
	}
	for _, proj := range ws.Projects {
		if proj.Name == opts.ProjectName {
			if opts.RepoName == "" {
				return nil
			}
			for _, repo := range proj.Repositories {
				if repo.Name == opts.RepoName {
					return nil
				}
			}
			return fmt.Errorf("repository %q not found in project %q", opts.RepoName, opts.ProjectName)
		}
	}
	return fmt.Errorf("project %q not found in workspace %q", opts.ProjectName, ws.Name)
}

// Resolve returns the workspaces to operate on based on flag inputs.
// wsName selects a specific workspace by name; all selects every workspace.
// When both are empty/false, the default workspace is returned.
func Resolve(cfg *config.Config, wsName string, all bool) ([]config.Workspace, error) {
	if wsName != "" && all {
		return nil, fmt.Errorf("--name and --all are mutually exclusive")
	}

	if all {
		return cfg.Workspaces, nil
	}

	if wsName != "" {
		for _, ws := range cfg.Workspaces {
			if ws.Name == wsName {
				return []config.Workspace{ws}, nil
			}
		}
		return nil, fmt.Errorf("workspace %q not found in config", wsName)
	}

	ws, err := config.DefaultWorkspace(cfg)
	if err != nil {
		return nil, err
	}
	return []config.Workspace{*ws}, nil
}
