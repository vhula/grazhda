package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/workspace"
)

func newStatsCmd() *cobra.Command {
	var parallel bool
	var parallelAll bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show repository metadata (last commit, 30-day commits, contributors)",
		Long: `Display aggregated repository metadata in aligned project-grouped tables.

Columns: REPO, LAST COMMIT (YYYY-MM-DD HH:MM), 30D COMMITS, CONTRIBUTORS.
Repos not yet cloned are shown as '(not cloned)' with '-' values.`,
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
				ParallelAll: parallelAll,
				ProjectName: projectName,
				RepoName:    repoName,
				Verbose:     verbose,
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

	cmd.Flags().BoolVar(&parallel, "parallel", false, "Query repos within each project concurrently")
	cmd.Flags().BoolVar(&parallelAll, "parallel-all", false, "Query repos across all projects concurrently")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
