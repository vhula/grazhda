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

// effectiveTags returns the merged tag set for a repo: project-level tags
// first, then any additional repo-level tags (deduped, preserving order).
func effectiveTags(proj config.Project, repo config.Repository) []string {
	seen := map[string]bool{}
	var tags []string
	for _, t := range proj.Tags {
		if !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	}
	for _, t := range repo.Tags {
		if !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	}
	return tags
}

// repoTagsMatch reports whether a repo's effective tag set intersects with
// filter (OR logic). An empty filter always matches.
func repoTagsMatch(proj config.Project, repo config.Repository, filter []string) bool {
	if len(filter) == 0 {
		return true
	}
	effective := effectiveTags(proj, repo)
	for _, ft := range filter {
		for _, et := range effective {
			if et == ft {
				return true
			}
		}
	}
	return false
}

// repoMatchesFilters returns true when the repo satisfies ALL active filters
// (RepoName substring and Tags). An absent filter always matches.
func repoMatchesFilters(proj config.Project, repo config.Repository, opts RunOptions) bool {
	if opts.RepoName != "" && !repoNameMatches(repo.Name, opts.RepoName) {
		return false
	}
	if len(opts.Tags) > 0 && !repoTagsMatch(proj, repo, opts.Tags) {
		return false
	}
	return true
}

// CountMatchingRepos returns the number of repositories in ws that satisfy
// opts.RepoName (substring match) and opts.Tags (OR match).
// If opts.ProjectName is set only that project's repositories are considered.
// Returns 0 when both opts.RepoName and opts.Tags are empty (no filter active).
func CountMatchingRepos(ws config.Workspace, opts RunOptions) int {
	if opts.RepoName == "" && len(opts.Tags) == 0 {
		return 0
	}
	total := 0
	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		for _, repo := range proj.Repositories {
			if repoMatchesFilters(proj, repo, opts) {
				total++
			}
		}
	}
	return total
}

// ValidateFilters returns an error if opts.ProjectName, opts.RepoName, or
// opts.Tags do not match any entry in the workspace configuration.
// Validates config structure only — filesystem presence is not checked here.
func ValidateFilters(ws config.Workspace, opts RunOptions) error {
	projectFound := false
	repoFound := false
	tagFound := false

	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		projectFound = true
		for _, repo := range proj.Repositories {
			nameOK := opts.RepoName == "" || repoNameMatches(repo.Name, opts.RepoName)
			tagOK := len(opts.Tags) == 0 || repoTagsMatch(proj, repo, opts.Tags)
			if nameOK {
				repoFound = true
			}
			if nameOK && tagOK {
				tagFound = true
			}
		}
	}

	if opts.ProjectName != "" && !projectFound {
		return fmt.Errorf("project %q not found in workspace %q", opts.ProjectName, ws.Name)
	}
	if opts.RepoName != "" && opts.ProjectName != "" && !repoFound {
		return fmt.Errorf("repository filter %q matched no repositories in project %q", opts.RepoName, opts.ProjectName)
	}
	if opts.RepoName != "" && opts.ProjectName != "" && repoFound && len(opts.Tags) > 0 && !tagFound {
		return fmt.Errorf("tag filter %v matched no repositories matching %q in project %q", opts.Tags, opts.RepoName, opts.ProjectName)
	}
	if len(opts.Tags) > 0 && !tagFound {
		return fmt.Errorf("tag filter %v matched no repositories in workspace %q", opts.Tags, ws.Name)
	}
	return nil
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
