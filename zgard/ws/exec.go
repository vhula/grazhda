package ws

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newExecCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool

	cmd := &cobra.Command{
		Use:   "exec <command> [args...]",
		Short: "Fan out a shell command to all repositories in a workspace",
		Long: `Execute an arbitrary shell command inside every targeted repository directory.

The command and its arguments are passed as positional arguments after the
subcommand name. Each repository's output is prefixed with its name for easy
scanning. Use **--parallel** to run concurrently, **--dry-run** to preview,
and targeting flags to narrow the scope.

**Tip:** combine stash + checkout + pull + exec for a full workspace refresh:

  zgard ws stash && zgard ws checkout main && zgard ws pull --parallel && zgard ws exec --parallel make build`,
		Example: `  # Run 'git status' in all default workspace repos
  zgard ws exec git status

  # Build all repos in parallel
  zgard ws exec --parallel make build

  # Run a command only in repos tagged 'backend'
  zgard ws exec -t backend npm install

  # Run in a specific project
  zgard ws exec -p frontend npm ci

  # Preview what would run without executing
  zgard ws exec --dry-run make test

  # Run in a specific repository
  zgard ws exec -p backend -r api-service go test ./...

  # Cross-workspace fan-out
  zgard ws exec --all --parallel docker compose pull

  # Full workspace refresh composition
  zgard ws stash && zgard ws checkout main && zgard ws pull --parallel && zgard ws exec --parallel make build`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := strings.Join(args, " ")
			return runWorkspaceOp(cmd, workspace.RunOptions{
				DryRun:      dryRun,
				Verbose:     verbose,
				Parallel:    parallel,
				ProjectName: projectName,
				RepoName:    repoName,
				Tags:        tagFilter,
			}, "executed", "would exec", func(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts workspace.RunOptions) error {
				return workspace.Exec(ws, command, exec, rep, opts)
			})
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Execute across all repositories concurrently")

	return cmd
}
