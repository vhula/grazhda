package targeting

import "github.com/vhula/grazhda/internal/config"

// Target identifies a resolved set of workspaces and projects to operate on.
type Target struct {
	Workspaces []config.Workspace
}

// Resolve selects workspaces and projects from cfg based on the provided
// workspace name (empty = default workspace).
func Resolve(cfg *config.Config, workspaceName string) (*Target, error) {
	if workspaceName != "" {
		for _, ws := range cfg.Workspaces {
			if ws.Name == workspaceName {
				return &Target{Workspaces: []config.Workspace{ws}}, nil
			}
		}
	}
	for _, ws := range cfg.Workspaces {
		if ws.Default {
			return &Target{Workspaces: []config.Workspace{ws}}, nil
		}
	}
	if len(cfg.Workspaces) > 0 {
		return &Target{Workspaces: []config.Workspace{cfg.Workspaces[0]}}, nil
	}
	return &Target{}, nil
}
