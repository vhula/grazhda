package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newPullCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool
	var parallelAll bool

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull latest changes for all repositories in a workspace",
		Long: `Run "git pull --rebase" on every repository in the targeted workspace.

Repositories not yet cloned are skipped. Use --parallel or --parallel-all
for concurrent pulls, and --dry-run to preview the operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			opts := workspace.RunOptions{
				DryRun:      dryRun,
				Verbose:     verbose,
				Parallel:    parallel,
				ParallelAll: parallelAll,
				ProjectName: projectName,
				RepoName:    repoName,
				Tags:        tagFilter,
			}

			for _, ws := range workspaces {
				if err := workspace.Pull(ws, exec, rep, opts); err != nil {
					return err
				}
			}

			label := "pulled"
			if dryRun {
				label = "would pull"
			}
			rep.Summary(label, dryRun)
			os.Exit(rep.ExitCode())
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Pull repositories within each project concurrently")
	cmd.Flags().BoolVar(&parallelAll, "parallel-all", false, "Pull all repositories across all projects concurrently")

	return cmd
}
