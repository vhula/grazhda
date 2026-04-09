package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/workspace"
)

func newStatsCmd() *cobra.Command {
	var parallel bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show repository metadata (last commit, 30-day commits, contributors)",
		Long: `Display aggregated repository metadata in aligned, project-grouped tables.

Columns reported for each repository:

- **REPO** — repository name
- **LAST COMMIT** — timestamp of the most recent commit (YYYY-MM-DD HH:MM)
- **30D COMMITS** — commit count over the past 30 days
- **CONTRIBUTORS** — unique author count across all commits

Repos not yet cloned are listed as "(not cloned)" with "-" for all values.`,
		Example: `  # Show commit stats for the default workspace
  zgard ws stats

  # Show stats for a named workspace
  zgard ws stats -n myworkspace

  # Collect stats in parallel across all workspaces
  zgard ws stats --all --parallel

  # Filter stats to a specific project
  zgard ws stats -n myworkspace -p backend`,
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

			opts := workspace.InspectOptions{
				Parallel:    parallel,
				ProjectName: projectName,
				RepoName:    repoName,
				Verbose:     verbose,
				Tags:        tagFilter,
			}

			exec := executor.OsExecutor{}
			for _, ws := range workspaces {
				if err := workspace.Stats(ws, exec, opts, os.Stdout); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&parallel, "parallel", false, "Query all repositories concurrently")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
