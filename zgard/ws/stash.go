package ws

import (
	"os"

	"github.com/spf13/cobra"
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
			if dryRun {
				rep.PrintDryRunBanner()
			}
			opts := workspace.RunOptions{
				DryRun:      dryRun,
				Verbose:     verbose,
				Parallel:    parallel,
				ProjectName: projectName,
				RepoName:    repoName,
				Tags:        tagFilter,
			}

			for _, ws := range workspaces {
				if err := workspace.Stash(ws, exec, rep, opts); err != nil {
					return err
				}
			}

			label := "stashed"
			if dryRun {
				label = "would stash"
			}
			rep.Summary(label, dryRun)
			os.Exit(rep.ExitCode())
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Stash all repositories concurrently")

	return cmd
}
