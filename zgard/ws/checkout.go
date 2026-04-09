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

The branch name is a required positional argument. Repositories not
present on disk are skipped. Use --dry-run to preview the operations.`,
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
