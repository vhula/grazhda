package ws

import (
	"github.com/spf13/cobra"
)

// NewCmd returns the `ws` parent command with all subcommands registered.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ws",
		Short: "Workspace operations",
		Long:  "Manage workspace lifecycle: initialize, purge, or pull repositories.",
	}
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newPurgeCmd())
	cmd.AddCommand(newPullCmd())
	cmd.AddCommand(newStatusCmd())
	return cmd
}
