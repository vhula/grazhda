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

// runOverRepos iterates all matching repos in a workspace and calls fn for each,
// honouring opts.Parallel, opts.ParallelAll, opts.ProjectName, and opts.RepoName.
// "Workspace:" and per-project "Project:" headers are printed to rep.
// Returns an error if opts.ProjectName or opts.RepoName matches nothing in config.
func runOverRepos(
	ws config.Workspace,
	opts RunOptions,
	rep *reporter.Reporter,
	fn func(proj config.Project, projPath string, repo config.Repository),
) error {
	if err := ValidateFilters(ws, opts); err != nil {
		return err
	}
	if n := CountMatchingRepos(ws, opts); n > 1 {
		rep.PrintWarn(fmt.Sprintf(
			"Warning: --repo-name %q matches %d repositories",
			opts.RepoName, n,
		))
	}

	wsPath := ExpandHome(ws.Path)
	rep.PrintLine("Workspace: " + ws.Name)

	if opts.ParallelAll {
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
					fn(proj, projPath, repo)
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

		if opts.Parallel {
			var wg sync.WaitGroup
			for _, repo := range proj.Repositories {
				if !repoMatchesFilters(proj, repo, opts) {
					continue
				}
				repo := repo
				wg.Add(1)
				go func() {
					defer wg.Done()
					fn(proj, projPath, repo)
				}()
			}
			wg.Wait()
		} else {
			for _, repo := range proj.Repositories {
				if !repoMatchesFilters(proj, repo, opts) {
					continue
				}
				fn(proj, projPath, repo)
			}
		}
	}
	return nil
}

// Exec fans out a shell command to all resolved repository directories.
// Output captured from each repo is printed indented below its status line.
// All repos are attempted; failures are recorded and the function returns nil
// (the caller checks rep.ExitCode()).
func Exec(ws config.Workspace, command string, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) error {
	return runOverRepos(ws, opts, rep, func(proj config.Project, projPath string, repo config.Repository) {
		execRepo(ws, proj, projPath, repo, command, exec, rep, opts)
	})
}

// Stash runs "git stash push" in all resolved repository directories.
func Stash(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) error {
	return runOverRepos(ws, opts, rep, func(proj config.Project, projPath string, repo config.Repository) {
		stashRepo(ws, proj, projPath, repo, exec, rep, opts)
	})
}

// Checkout runs "git checkout <branch>" in all resolved repository directories.
func Checkout(ws config.Workspace, branch string, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) error {
	return runOverRepos(ws, opts, rep, func(proj config.Project, projPath string, repo config.Repository) {
		checkoutRepo(ws, proj, projPath, repo, branch, exec, rep, opts)
	})
}

func execRepo(
	ws config.Workspace,
	proj config.Project,
	projPath string,
	repo config.Repository,
	command string,
	exec executor.Executor,
	rep *reporter.Reporter,
	opts RunOptions,
) {
	destName := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ws.Structure)
	repoPath := filepath.Join(projPath, destName)

	if opts.DryRun {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Msg: fmt.Sprintf("[DRY RUN] would exec: %s", command),
		})
		return
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Skipped: true, Msg: "not present, skipped",
		})
		return
	}

	if opts.Verbose {
		rep.PrintLine(fmt.Sprintf("  → %s: %s", repo.Name, command))
	}

	output, err := exec.RunCapture(repoPath, command)

	var outputLines []string
	if output != "" {
		for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
			outputLines = append(outputLines, line)
		}
	}

	if err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Err: err, OutputLines: outputLines,
		})
		return
	}

	rep.Record(reporter.OpResult{
		Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
		Msg: "done", OutputLines: outputLines,
	})
}

func stashRepo(
	ws config.Workspace,
	proj config.Project,
	projPath string,
	repo config.Repository,
	exec executor.Executor,
	rep *reporter.Reporter,
	opts RunOptions,
) {
	destName := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ws.Structure)
	repoPath := filepath.Join(projPath, destName)

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
			Msg: "[DRY RUN] would stash",
		})
		return
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Skipped: true, Msg: "not present, skipped",
		})
		return
	}

	if opts.Verbose {
		rep.PrintLine(fmt.Sprintf("  → %s: git stash push", repo.Name))
	}

	if err := exec.Run(repoPath, "git stash push"); err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name, Err: err,
		})
		return
	}

	rep.Record(reporter.OpResult{
		Workspace: ws.Name, Project: proj.Name, Repo: repo.Name, Msg: "stashed",
	})
}

func checkoutRepo(
	ws config.Workspace,
	proj config.Project,
	projPath string,
	repo config.Repository,
	branch string,
	exec executor.Executor,
	rep *reporter.Reporter,
	opts RunOptions,
) {
	destName := ResolveDestName(projPath, repo.Name, repo.LocalDirName, ws.Structure)
	repoPath := filepath.Join(projPath, destName)

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
			Msg: fmt.Sprintf("[DRY RUN] would checkout: %s", branch),
		})
		return
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
			Skipped: true, Msg: "not present, skipped",
		})
		return
	}

	cmd := fmt.Sprintf("git checkout %s", branch)

	if opts.Verbose {
		rep.PrintLine(fmt.Sprintf("  → %s: %s", repo.Name, cmd))
	}

	if err := exec.Run(repoPath, cmd); err != nil {
		rep.Record(reporter.OpResult{
			Workspace: ws.Name, Project: proj.Name, Repo: repo.Name, Err: err,
		})
		return
	}

	rep.Record(reporter.OpResult{
		Workspace: ws.Name, Project: proj.Name, Repo: repo.Name,
		Msg: fmt.Sprintf("checked out %s", branch),
	})
}
