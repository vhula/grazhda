package workspace

import (
	"os"
	"path/filepath"

	"github.com/vhula/grazhda/internal/config"
)

// RepoInfo holds the resolved state of a single repository for display purposes.
type RepoInfo struct {
	config.Repository
	ProjectName   string
	ProjectBranch string
	ProjectTags   []string
	LocalPath     string // absolute path on disk
	Cloned        bool   // true if the directory exists on disk
}

// FilteredRepos returns a copy of ws.Projects filtered by InspectOptions.
// Projects and repositories that do not match the active filters are omitted.
// An empty result is returned (rather than an error) if filters match nothing.
func FilteredRepos(ws config.Workspace, opts InspectOptions) []config.Project {
	var result []config.Project
	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		var repos []config.Repository
		for _, repo := range proj.Repositories {
			ro := RunOptions{RepoName: opts.RepoName, Tags: opts.Tags}
			if repoMatchesFilters(proj, repo, ro) {
				repos = append(repos, repo)
			}
		}
		if len(repos) > 0 {
			p := proj
			p.Repositories = repos
			result = append(result, p)
		}
	}
	return result
}

// ResolveRepoInfos returns RepoInfo entries for every repository in ws that
// matches opts. Each entry includes the computed local path and clone status.
func ResolveRepoInfos(ws config.Workspace, opts InspectOptions) []RepoInfo {
	wsPath := ExpandHome(ws.Path)
	projects := FilteredRepos(ws, opts)

	var infos []RepoInfo
	for _, proj := range projects {
		projPath := filepath.Join(wsPath, proj.Name)
		for _, repo := range proj.Repositories {
			destName := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ResolveStructure(ws, proj))
			localPath := filepath.Join(projPath, destName)
			_, err := os.Stat(localPath)
			infos = append(infos, RepoInfo{
				Repository:    repo,
				ProjectName:   proj.Name,
				ProjectBranch: proj.Branch,
				ProjectTags:   proj.Tags,
				LocalPath:     localPath,
				Cloned:        err == nil,
			})
		}
	}
	return infos
}
