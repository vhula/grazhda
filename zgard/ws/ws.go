package ws

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Targeting flags shared by all ws subcommands via PersistentFlags.
// Each subcommand reads these package-level variables directly.
var (
	wsName      string   // --name / -n
	wsAll       bool     // --all
	projectName string   // --project-name / -p
	repoName    string   // --repo-name / -r
	tagFilter   []string // --tag / -t
)

// NewCmd returns the `ws` parent command with all subcommands registered.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ws",
		Short: "Workspace operations",
		Long: `Manage workspace lifecycle: initialize, pull, purge, and run coordinated
operations across all repositories defined in your grazhda configuration.

## Targeting flags (available for every ws subcommand)

| Flag                    | Description                                       |
|-------------------------|---------------------------------------------------|
| **-n / --name**         | Target a named workspace                          |
| **--all**               | Operate on all workspaces                         |
| **-p / --project-name** | Filter to a specific project                      |
| **-r / --repo-name**    | Narrow to a repository (requires -p)              |
| **-t / --tag**          | Filter by tag — repeatable for OR logic           |

With no flag, zgard uses the **default** workspace and prints a notice.

> **Note:** ws purge is an exception — it always requires an explicit target.</p>

## Discovery commands

Use **ws list** to see real-time clone status for all repositories.
Use **zgard config list** to inspect the raw workspace hierarchy from config.`,
		// Validate cross-flag constraints before any subcommand runs.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if repoName != "" && projectName == "" {
				return fmt.Errorf("--repo-name (-r) requires --project-name (-p)")
			}
			if wsAll && projectName != "" {
				return fmt.Errorf("--all and --project-name (-p) are mutually exclusive")
			}
			if wsAll && repoName != "" {
				return fmt.Errorf("--all and --repo-name (-r) are mutually exclusive")
			}
			return nil
		},
	}

	// Targeting flags — inherited by every ws subcommand.
	cmd.PersistentFlags().StringVarP(&wsName, "name", "n", "", "Target workspace name (default: default workspace)")
	cmd.PersistentFlags().BoolVar(&wsAll, "all", false, "Operate on all workspaces")
	cmd.PersistentFlags().StringVarP(&projectName, "project-name", "p", "", "Filter to a specific project")
	cmd.PersistentFlags().StringVarP(&repoName, "repo-name", "r", "", "Filter to a specific repository (requires --project-name / -p)")
	cmd.PersistentFlags().StringArrayVarP(&tagFilter, "tag", "t", nil, "Filter by tag (OR logic; repeat for multiple: -t backend -t api)")

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newPurgeCmd())
	cmd.AddCommand(newPullCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newExecCmd())
	cmd.AddCommand(newStashCmd())
	cmd.AddCommand(newCheckoutCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newDiffCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newListCmd())
	return cmd
}
