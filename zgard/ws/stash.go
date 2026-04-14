package ws

import (
	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newStashCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool

	cmd := &cobra.Command{
		Use:   "stash",
		Short: "Stash local changes in all repositories in a workspace",
		Long: `Run "git stash" in every repository that has uncommitted changes.

Repositories with a clean working tree are automatically **skipped** and
reported as such. Use **--parallel** to stash concurrently and **--dry-run**
to preview which repositories have changes that would be stashed.`,
		Example: `  # Stash all repos in the default workspace
  zgard ws stash

  # Stash a named workspace
  zgard ws stash -n myworkspace

  # Stash all workspaces concurrently
  zgard ws stash --all --parallel

  # Preview which repos have changes to stash
  zgard ws stash -n myworkspace --dry-run

  # Stash only a specific project's repositories
  zgard ws stash -n myworkspace -p backend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspaceOp(cmd, workspace.RunOptions{
				DryRun:      dryRun,
				Verbose:     verbose,
				Parallel:    parallel,
				ProjectName: projectName,
				RepoName:    repoName,
				Tags:        tagFilter,
			}, "stashed", "would stash", func(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts workspace.RunOptions) error {
				return workspace.Stash(ws, exec, rep, opts)
			})
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Stash all repositories concurrently")

	return cmd
}
