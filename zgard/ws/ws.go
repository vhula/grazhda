package ws

import (
"fmt"

"github.com/spf13/cobra"
)

// Targeting flags shared by all ws subcommands via PersistentFlags.
// Each subcommand reads these package-level variables directly.
var (
wsName      string // --name / -n
wsAll       bool   // --all
projectName string // --project-name / -p
repoName    string // --repo-name / -r
)

// NewCmd returns the `ws` parent command with all subcommands registered.
func NewCmd() *cobra.Command {
cmd := &cobra.Command{
Use:   "ws",
Short: "Workspace operations",
Long:  "Manage workspace lifecycle: initialize, purge, pull, or run coordinated operations across repositories.",
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

cmd.AddCommand(newInitCmd())
cmd.AddCommand(newPurgeCmd())
cmd.AddCommand(newPullCmd())
cmd.AddCommand(newStatusCmd())
cmd.AddCommand(newExecCmd())
cmd.AddCommand(newStashCmd())
cmd.AddCommand(newCheckoutCmd())
return cmd
}
