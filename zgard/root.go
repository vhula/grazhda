package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/zgard/ws"
)

var rootCmd = &cobra.Command{
	Use:           "zgard",
	Short:         "Workspace lifecycle manager",
	Long:          "zgard manages local workspace lifecycle — init, purge, and pull repositories.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root Cobra command, printing any error in red to stderr.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, clr.Red(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(ws.NewCmd())
}
