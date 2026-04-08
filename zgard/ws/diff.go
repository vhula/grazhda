package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/workspace"
)

func newDiffCmd() *cobra.Command {
	var parallel bool
	var parallelAll bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show Git state (uncommitted, ahead, behind) across repositories",
		Long: `Display a per-repository Git state summary in aligned project-grouped tables.

Columns: REPO, UNCOMMITTED (file count), AHEAD (commits), BEHIND (commits).
Repos without an upstream tracking branch show '--' for AHEAD/BEHIND.
Repos not yet cloned are shown as '(not cloned)'.`,
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
				if err := workspace.Diff(ws, exec, opts, os.Stdout); err != nil {
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
