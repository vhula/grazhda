package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newCheckoutCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool

	cmd := &cobra.Command{
		Use:   "checkout <branch>",
		Short: "Check out a branch across all repositories in a workspace",
		Long: `Run "git checkout <branch>" in every targeted repository.

The **branch name** is a required positional argument. Repositories that
are not yet cloned on disk are automatically **skipped**. Use **--parallel**
to switch branches concurrently and **--dry-run** to preview the operations.

**Tip:** pair with ws stash and ws pull for a safe branch-switch workflow:

  zgard ws stash -n myworkspace && zgard ws checkout -n myworkspace main && zgard ws pull -n myworkspace`,
		Example: `  # Check out 'main' across all repos in the default workspace
  zgard ws checkout main

  # Check out a feature branch in a named workspace
  zgard ws checkout -n myworkspace feature/my-feature

  # Check out in parallel across all workspaces
  zgard ws checkout --all main --parallel

  # Only check out a specific project's repos
  zgard ws checkout -n myworkspace -p backend develop

  # Preview the checkout without making changes
  zgard ws checkout -n myworkspace main --dry-run

  # Safe branch-switch: stash changes then checkout
  zgard ws stash -n myworkspace && zgard ws checkout -n myworkspace main`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			workspaces, err := workspace.Resolve(cfg, wsName, wsAll)
			if err != nil {
				return err
			}

			if wsName == "" && !wsAll {
				warnDefaultTarget(os.Stderr, workspaces[0])
			}

			exec := executor.OsExecutor{}
			rep := reporter.NewReporter(os.Stdout, os.Stderr)
			rep.ShowElapsed = verbose
			opts := workspace.RunOptions{
				DryRun:      dryRun,
				Verbose:     verbose,
				Parallel:    parallel,
				ProjectName: projectName,
				RepoName:    repoName,
				Tags:        tagFilter,
			}

			for _, ws := range workspaces {
				if err := workspace.Checkout(ws, branch, exec, rep, opts); err != nil {
					return err
				}
			}

			label := "checked out"
			if dryRun {
				label = "would checkout"
			}
			rep.Summary(label, dryRun)
			os.Exit(rep.ExitCode())
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Checkout all repositories concurrently")

	return cmd
}
