package ws

import (
	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newPullCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull latest changes for all repositories in a workspace",
		Long: `Run "git pull --rebase" on every repository in the targeted workspace.

Repositories that are not yet cloned are automatically **skipped**.
Use **--parallel** to pull concurrently across all repos, and **--dry-run**
to preview which repositories would be updated.`,
		Example: `  # Pull the default workspace
  zgard ws pull

  # Pull a named workspace
  zgard ws pull -n myworkspace

  # Pull all workspaces concurrently
  zgard ws pull --all --parallel

  # Preview which repos would be pulled
  zgard ws pull -n myworkspace --dry-run

  # Pull only a specific project's repositories
  zgard ws pull -n myworkspace -p backend

  # Pull repositories tagged 'api'
  zgard ws pull -t api --parallel`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspaceOp(cmd, workspace.RunOptions{
				DryRun:      dryRun,
				Verbose:     verbose,
				Parallel:    parallel,
				ProjectName: projectName,
				RepoName:    repoName,
				Tags:        tagFilter,
			}, "pulled", "would pull", func(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts workspace.RunOptions) error {
				return workspace.Pull(ws, exec, rep, opts)
			})
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Pull all repositories concurrently")

	return cmd
}
