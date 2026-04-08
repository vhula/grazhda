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
	Long: `zgard manages the full workspace lifecycle — from cloning repositories
and pulling updates to cross-repo orchestration, inspection, and IDE integration.

Subcommands (under "zgard ws"):
  init        Clone all repositories defined in a workspace
  pull        Pull latest changes for every repository
  purge       Remove workspace directories from disk
  exec        Run an arbitrary shell command across repositories
  stash       Stash uncommitted changes across repositories
  checkout    Switch branches across repositories
  status      Show workspace health as monitored by dukh
  search      Search for files or content across repositories
  diff        Show uncommitted changes and upstream sync status
  stats       Aggregate commit metadata across repositories
  open        Launch an IDE for targeted repository paths`,
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
