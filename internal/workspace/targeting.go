package workspace

import (
	"fmt"

	"github.com/vhula/grazhda/internal/config"
)

// Resolve returns the workspaces to operate on based on flag inputs.
// wsName selects a specific workspace by name; all selects every workspace.
// When both are empty/false, the default workspace is returned.
func Resolve(cfg *config.Config, wsName string, all bool) ([]config.Workspace, error) {
	if wsName != "" && all {
		return nil, fmt.Errorf("--ws and --all are mutually exclusive")
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
