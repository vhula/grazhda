package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newInitCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool
	var cloneDelaySeconds int
	var noConfirm bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a workspace by cloning all repositories",
		Long: `Clone every repository listed in the workspace configuration.

Directories are created before cloning begins. Repositories that already
exist on disk are skipped. Use --parallel to clone
concurrently, and --dry-run to preview without making changes.
Use --no-confirm to skip the confirmation prompt.`,
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
			rep.ShowElapsed = verbose
			opts := workspace.RunOptions{
				DryRun:            dryRun,
				Verbose:           verbose,
				Parallel:          parallel,
				NoConfirm:         noConfirm,
				CloneDelaySeconds: cloneDelaySeconds,
				ProjectName:       projectName,
				RepoName:          repoName,
				Tags:              tagFilter,
			}

			for _, ws := range workspaces {
				if err := workspace.Init(ws, exec, rep, opts); err != nil {
					return err
				}
			}

			label := "cloned"
			if dryRun {
				label = "would clone"
			}
			rep.Summary(label, dryRun)
			os.Exit(rep.ExitCode())
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Clone all repositories concurrently")
	cmd.Flags().IntVar(&cloneDelaySeconds, "clone-delay-seconds", 0, "Seconds to sleep after each clone command (0 = disabled)")
	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompts")

	return cmd
}
