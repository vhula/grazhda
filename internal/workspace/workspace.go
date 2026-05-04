package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
)

// ExpandHome expands a leading ~ to the user's home directory.
func ExpandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}

// lastSegment returns the portion of name that follows the last "/", with any
// trailing ".git" suffix stripped.  Used by the "list" structure mode.
//
// Examples:
//
//	"org/pack/repo"     → "repo"
//	"org/repo.git"      → "repo"
//	"repo"              → "repo"
func lastSegment(name string) string {
	name = strings.TrimSuffix(name, ".git")
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

// ResolveStructure returns the effective structure for a project, implementing
// the precedence: project.Structure > workspace.Structure > default (tree).
func ResolveStructure(ws config.Workspace, proj config.Project) string {
	if proj.Structure != "" {
		return proj.Structure
	}
	if ws.Structure != "" {
		return ws.Structure
	}
	return config.StructureTree
}

// ResolveDestName returns the local directory name for a repository.
//
// If localDirName is non-empty it is always used unchanged.
//
// For structure == "tree" (default, or any unrecognised value): the full
// repoName is returned as-is, so slashes in the name become nested
// sub-directories (e.g. "org/pack/repo" → <project>/org/pack/repo).
//
// For structure == "list": only the last "/"-delimited segment of repoName
// is used, after stripping any ".git" suffix
// (e.g. "org/pack/repo.git" → <project>/repo).
// If two repositories in the same project share the same last segment the
// second one will be skipped as "already exists" during ws init/pull.
// Use local_dir_name to resolve such naming conflicts explicitly.
//
// projPath is accepted for API compatibility but is not consulted.
func ResolveDestName(_ /*projPath*/ string, repoName, localDirName, structure string) string {
	if localDirName != "" {
		return localDirName
	}
	if structure == config.StructureList {
		return lastSegment(repoName)
	}
	return repoName
}

// ResolveDestNamesForProject returns the destination directory name for each
// repository in repos, using the workspace structure setting.
//
// For structure == "list" every entry is the last "/"-delimited segment of
// its repo name (with ".git" stripped).  Duplicates are possible; the caller
// (cloneRepo / pullRepo) already handles the "skip if directory exists" case.
//
// This function is used by dukh's monitoring loop to determine where repos
// were placed on disk without touching the filesystem.
func ResolveDestNamesForProject(repos []config.Repository, structure string) []string {
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = ResolveDestName("", repo.Name, repo.LocalDirName, structure)
	}
	return names
}

// Init initializes the workspace by creating directory structure and cloning all repositories.
//
// When opts.Parallel is true, all repositories across every project are
// cloned concurrently in a single goroutine pool (directories are created
// before the pool starts).
// opts.CloneDelaySeconds introduces a per-repo sleep after each clone
// command; it is applied even in parallel mode (each goroutine sleeps after
// its own clone).
//
// When opts.ProjectName or opts.RepoName are set, only matching projects/repos
// are created and cloned. ValidateFilters is called first so unmatched filters
// return an error before any filesystem changes occur.
func Init(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) error {
	if err := ValidateFilters(ws, opts); err != nil {
		return err
	}
	if n := CountMatchingRepos(ws, opts); n > 1 {
		rep.PrintWarn(fmt.Sprintf(
			"Warning: --repo-name %q matches %d repositories",
			opts.RepoName, n,
		))
	}
	rep.PrintLine("Workspace: " + ws.Name)
	wsPath := ExpandHome(ws.Path)

	// Pre-create project directories for matching projects before any
	// cloning starts so parallel goroutines never race on MkdirAll.
	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		projPath := filepath.Join(wsPath, proj.Name)
		if opts.DryRun {
			rep.PrintLine(fmt.Sprintf("    [DRY RUN] would create directory: %s", projPath))
		} else {
			if err := os.MkdirAll(projPath, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", projPath, err)
			}
		}
	}

	if opts.Parallel {
		// Count total matching repos so the reporter can show progress (N/total).
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
		if total > 0 {
			rep.SetTotal(total)
			defer rep.SetTotal(0)
		}

		var wg sync.WaitGroup
		for _, proj := range ws.Projects {
			if opts.ProjectName != "" && proj.Name != opts.ProjectName {
				continue
			}
			proj := proj
			rep.PrintLine("  Project: " + proj.Name)
			projPath := filepath.Join(wsPath, proj.Name)
			for _, repo := range proj.Repositories {
				if !repoMatchesFilters(proj, repo, opts) {
					continue
				}
				repo := repo
				wg.Add(1)
				go func() {
					defer wg.Done()
					cloneRepo(ws, proj, projPath, repo, exec, rep, opts)
				}()
			}
		}
		wg.Wait()
		return nil
	}

	// Sequential path.
	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		rep.PrintLine("  Project: " + proj.Name)
		projPath := filepath.Join(wsPath, proj.Name)

		for _, repo := range proj.Repositories {
			if !repoMatchesFilters(proj, repo, opts) {
				continue
			}
			// Check cancellation between repos in sequential mode.
			if err := opts.ctx().Err(); err != nil {
				return fmt.Errorf("%w: %w", ErrCancelled, err)
			}
			cloneRepo(ws, proj, projPath, repo, exec, rep, opts)
		}
	}
	return nil
}

func cloneRepo(ws config.Workspace, proj config.Project, projPath string, repo config.Repository, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) {
	destName := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ResolveStructure(ws, proj))
	repoPath := filepath.Join(projPath, destName)

	branch := repo.Branch
	if branch == "" {
		branch = proj.Branch
	}

	if opts.DryRun {
		if _, err := os.Stat(repoPath); err == nil {
			rep.Record(reporter.OpResult{
				Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
				Skipped: true, Msg: "[DRY RUN] already exists, would skip",
			})
			return
		}
		cmd, err := config.RenderCloneCmd(ws.CloneCommandTemplate, proj, repo, repoPath)
		if err != nil {
			rep.Record(reporter.OpResult{
				Workspace: ws.Name, Project: proj.Name, Repo: repo.Name, Err: err,
			})
			return
		}
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Msg: fmt.Sprintf("[DRY RUN] would clone (%s)  → %s", branch, cmd),
		})
		return
	}

	// Skip if already cloned
	if _, err := os.Stat(repoPath); err == nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Skipped: true, Msg: "already exists, skipped",
		})
		return
	}

	cmd, err := config.RenderCloneCmd(ws.CloneCommandTemplate, proj, repo, repoPath)
	if err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name, Err: err,
		})
		return
	}

	if opts.Verbose {
		rep.PrintLine(fmt.Sprintf("  → %s", cmd))
	}

	var success bool
	start := time.Now()
	defer func() {
		if !success {
			os.RemoveAll(repoPath) //nolint:errcheck
		}
	}()

	if err := exec.RunContext(opts.ctx(), projPath, cmd); err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Err: err, Elapsed: time.Since(start),
		})
		return
	}

	success = true
	rep.Record(reporter.OpResult{
		Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
		Msg: fmt.Sprintf("cloned (%s)", branch), Elapsed: time.Since(start),
	})

	if opts.CloneDelaySeconds > 0 {
		time.Sleep(time.Duration(opts.CloneDelaySeconds) * time.Second)
	}
}

// Purge removes the workspace directory tree.
func Purge(ws config.Workspace, rep *reporter.Reporter, opts RunOptions) error {
	wsPath := ExpandHome(ws.Path)

	if opts.DryRun {
		rep.PrintLine(fmt.Sprintf("[DRY RUN] would remove: %s", wsPath))
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Repo: ws.Name,
			Msg: fmt.Sprintf("[DRY RUN] would remove %s", wsPath),
		})
		return nil
	}

	if _, err := os.Stat(wsPath); os.IsNotExist(err) {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Repo: ws.Name,
			Skipped: true, Msg: "directory not found, skipped",
		})
		return nil
	}

	if err := os.RemoveAll(wsPath); err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Repo: ws.Name, Err: err,
		})
		return nil
	}

	rep.Record(reporter.OpResult{
		Workspace: ws.Name, Repo: ws.Name,
		Msg: fmt.Sprintf("removed %s", wsPath),
	})
	return nil
}

// Pull runs git pull --rebase for each repository in the workspace.
//
// When opts.Parallel is true, all repositories across every project are
// pulled concurrently in a single goroutine pool.
//
// When opts.ProjectName or opts.RepoName are set, only matching projects/repos
// are pulled. ValidateFilters is called first so unmatched filters return an
// error before any git operations run.
func Pull(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) error {
	if err := ValidateFilters(ws, opts); err != nil {
		return err
	}
	if n := CountMatchingRepos(ws, opts); n > 1 {
		rep.PrintWarn(fmt.Sprintf(
			"Warning: --repo-name %q matches %d repositories",
			opts.RepoName, n,
		))
	}
	rep.PrintLine("Workspace: " + ws.Name)
	wsPath := ExpandHome(ws.Path)

	if opts.Parallel {
		// Count total matching repos so the reporter can show progress (N/total).
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
		if total > 0 {
			rep.SetTotal(total)
			defer rep.SetTotal(0)
		}

		var wg sync.WaitGroup
		for _, proj := range ws.Projects {
			if opts.ProjectName != "" && proj.Name != opts.ProjectName {
				continue
			}
			proj := proj
			rep.PrintLine("  Project: " + proj.Name)
			projPath := filepath.Join(wsPath, proj.Name)
			for _, repo := range proj.Repositories {
				if !repoMatchesFilters(proj, repo, opts) {
					continue
				}
				repo := repo
				wg.Add(1)
				go func() {
					defer wg.Done()
					pullRepo(ws, proj, projPath, repo, exec, rep, opts)
				}()
			}
		}
		wg.Wait()
		return nil
	}

	for _, proj := range ws.Projects {
		if opts.ProjectName != "" && proj.Name != opts.ProjectName {
			continue
		}
		rep.PrintLine("  Project: " + proj.Name)
		projPath := filepath.Join(wsPath, proj.Name)

		for _, repo := range proj.Repositories {
			if !repoMatchesFilters(proj, repo, opts) {
				continue
			}
			if err := opts.ctx().Err(); err != nil {
				return fmt.Errorf("%w: %w", ErrCancelled, err)
			}
			pullRepo(ws, proj, projPath, repo, exec, rep, opts)
		}
	}
	return nil
}

func pullRepo(ws config.Workspace, proj config.Project, projPath string, repo config.Repository, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) {
	destName := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ResolveStructure(ws, proj))
	repoPath := filepath.Join(projPath, destName)

	branch := repo.Branch
	if branch == "" {
		branch = proj.Branch
	}

	if opts.DryRun {
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			rep.Record(reporter.OpResult{
				Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
				Skipped: true, Msg: "[DRY RUN] not present, would skip",
			})
			return
		}
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Msg: fmt.Sprintf("[DRY RUN] would pull (%s)", branch),
		})
		return
	}

	// Skip repos not yet cloned
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Skipped: true, Msg: "not present, skipped",
		})
		return
	}

	cmd := fmt.Sprintf("git pull --rebase origin %s", branch)

	if opts.Verbose {
		rep.PrintLine(fmt.Sprintf("  → %s", cmd))
	}

	start := time.Now()
	if err := exec.RunContext(opts.ctx(), repoPath, cmd); err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Err: err, Elapsed: time.Since(start),
		})
		return
	}

	rep.Record(reporter.OpResult{
		Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
		Msg: fmt.Sprintf("pulled (%s)", branch), Elapsed: time.Since(start),
	})
}
