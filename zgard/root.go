package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/zgard/ws"
)

var rootCmd = &cobra.Command{
	Use:   "zgard",
	Short: "Workspace lifecycle manager",
	Long:  "zgard manages local workspace lifecycle — init, purge, and pull repositories.",
}

// Execute runs the root Cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(ws.NewCmd())
}
