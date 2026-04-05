package ws

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPurgeCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var noConfirm bool
	var workspace string

	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge a workspace by removing all cloned repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("ws purge: not yet implemented")
			_ = dryRun
			_ = verbose
			_ = noConfirm
			_ = workspace
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompts")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Target workspace name (default: default workspace)")

	return cmd
}
