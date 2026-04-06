package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
)

// expandHome expands a leading ~ to the user's home directory.
func expandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}

// ResolveDestName returns the local directory name for a repository inside
// projPath, taking the workspace Structure setting into account.
//
// If localDirName is non-empty it is always used unchanged.
//
// For structure == "tree" (or empty / any unrecognised value): the full
// repoName is returned (slashes become nested sub-directories).
//
// For structure == "list": the shortest trailing suffix of repoName (split on
// "/") that does not already exist as a directory inside projPath is returned.
// If all suffixes are taken the full repoName is returned as a fallback.
//
// Example with repoName = "org/pack/repo1" and structure = "list":
//   - tries "repo1"       — returns if <projPath>/repo1 does not exist
//   - tries "pack/repo1"  — returns if <projPath>/pack/repo1 does not exist
//   - tries "org/pack/repo1" — always the final fallback
func ResolveDestName(projPath, repoName, localDirName, structure string) string {
	if localDirName != "" {
		return localDirName
	}
	if structure != config.StructureList {
		return repoName
	}
	parts := strings.Split(repoName, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.Join(parts[i:], string(filepath.Separator))
		if _, err := os.Stat(filepath.Join(projPath, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
	return repoName
}

// Init initializes the workspace by creating directory structure and cloning all repositories.
func Init(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) error {
	rep.PrintLine("Workspace: " + ws.Name)
	wsPath := expandHome(ws.Path)

	for _, proj := range ws.Projects {
		rep.PrintLine("  Project: " + proj.Name)
		projPath := filepath.Join(wsPath, proj.Name)

		if opts.DryRun {
			rep.PrintLine(fmt.Sprintf("    [DRY RUN] would create directory: %s", wsPath))
			rep.PrintLine(fmt.Sprintf("    [DRY RUN] would create directory: %s", projPath))
		} else {
			if err := os.MkdirAll(projPath, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", projPath, err)
			}
		}

		if opts.Parallel {
			var wg sync.WaitGroup
			for _, repo := range proj.Repositories {
				wg.Add(1)
				repo := repo
				go func() {
					defer wg.Done()
					cloneRepo(ws, proj, projPath, repo, exec, rep, opts)
				}()
			}
			wg.Wait()
		} else {
			for _, repo := range proj.Repositories {
				cloneRepo(ws, proj, projPath, repo, exec, rep, opts)
			}
		}
	}
	return nil
}

func cloneRepo(ws config.Workspace, proj config.Project, projPath string, repo config.Repository, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) {
	destName := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ws.Structure)
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
	defer func() {
		if !success {
			os.RemoveAll(repoPath) //nolint:errcheck
		}
	}()

	if err := exec.Run(projPath, cmd); err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name, Err: err,
		})
		return
	}

	success = true
	rep.Record(reporter.OpResult{
		Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
		Msg: fmt.Sprintf("cloned (%s)", branch),
	})
}

// Purge removes the workspace directory tree.
func Purge(ws config.Workspace, rep *reporter.Reporter, opts RunOptions) error {
	wsPath := expandHome(ws.Path)

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
func Pull(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) error {
	rep.PrintLine("Workspace: " + ws.Name)
	wsPath := expandHome(ws.Path)

	for _, proj := range ws.Projects {
		rep.PrintLine("  Project: " + proj.Name)
		projPath := filepath.Join(wsPath, proj.Name)

		if opts.Parallel {
			var wg sync.WaitGroup
			for _, repo := range proj.Repositories {
				wg.Add(1)
				repo := repo
				go func() {
					defer wg.Done()
					pullRepo(ws, proj, projPath, repo, exec, rep, opts)
				}()
			}
			wg.Wait()
		} else {
			for _, repo := range proj.Repositories {
				pullRepo(ws, proj, projPath, repo, exec, rep, opts)
			}
		}
	}
	return nil
}

func pullRepo(ws config.Workspace, proj config.Project, projPath string, repo config.Repository, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) {
	destName := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ws.Structure)
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

	if err := exec.Run(repoPath, cmd); err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name, Err: err,
		})
		return
	}

	rep.Record(reporter.OpResult{
		Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
		Msg: fmt.Sprintf("pulled (%s)", branch),
	})
}
