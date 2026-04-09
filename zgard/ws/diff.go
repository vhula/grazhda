package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/workspace"
)

func newDiffCmd() *cobra.Command {
	var parallel bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show Git state (uncommitted, ahead, behind) across repositories",
		Long: `Display a per-repository Git state summary in aligned, project-grouped tables.

Columns reported for each repository:

- **REPO** — repository name
- **UNCOMMITTED** — count of modified / untracked files
- **AHEAD** — local commits not yet pushed to the upstream branch
- **BEHIND** — upstream commits available to pull

Repos without a tracking branch show "--" for AHEAD/BEHIND.
Repos not yet cloned are listed as "(not cloned)".`,
		Example: `  # Show diff summary for the default workspace
  zgard ws diff

  # Show diff for a named workspace
  zgard ws diff -n myworkspace

  # Collect diff data in parallel across all workspaces
  zgard ws diff --all --parallel

  # Filter to a specific project
  zgard ws diff -n myworkspace -p backend`,
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
				if err := workspace.Diff(ws, exec, opts, os.Stdout); err != nil {
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
