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
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Pull all repositories concurrently")

	return cmd
}
