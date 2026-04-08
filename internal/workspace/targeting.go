package workspace

import (
	"fmt"
	"strings"

	"github.com/vhula/grazhda/internal/config"
)

// repoNameMatches reports whether a repository config name satisfies a filter
// string using case-sensitive substring matching on the full config name.
//
// The workspace structure setting is ignored — "cool" matches
// "ORG/PACK/my-cool-backend-lol" regardless of whether the workspace uses
// "tree" or "list" structure.  An empty filter always matches.
func repoNameMatches(repoConfigName, filter string) bool {
	if filter == "" {
		return true
	}
	return strings.Contains(repoConfigName, filter)
}

// CountMatchingRepos returns the number of repositories in ws that satisfy
// opts.RepoName (substring match).  If opts.ProjectName is set only that
// project's repositories are considered.  Returns 0 when opts.RepoName is
// empty (no filter active).
func CountMatchingRepos(ws config.Workspace, opts RunOptions) int {
	if opts.RepoName == "" {
		return 0
	}
	total := 0
	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		for _, repo := range proj.Repositories {
			if repoNameMatches(repo.Name, opts.RepoName) {
				total++
			}
		}
	}
	return total
}

// ValidateFilters returns an error if opts.ProjectName or opts.RepoName
// does not match any entry in the given workspace's configuration.
// Validates config structure only — filesystem presence is not checked here.
//
// opts.RepoName is matched via case-sensitive substring against the full
// repository config name (e.g. "cool" matches "ORG/PACK/my-cool-backend-lol").
// Matching more than one repository is valid; the caller is responsible for
// warning the user when multiple matches are found.
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
				if repoNameMatches(repo.Name, opts.RepoName) {
					return nil // at least one match → valid
				}
			}
			return fmt.Errorf("repository filter %q matched no repositories in project %q", opts.RepoName, opts.ProjectName)
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
