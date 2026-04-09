package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/ui"
	"github.com/vhula/grazhda/zgard/cfgcmd"
	"github.com/vhula/grazhda/zgard/ws"
)

var noColor bool

var rootCmd = &cobra.Command{
	Use:   "zgard",
	Short: "Workspace lifecycle manager",
	Long: `# zgard — Workspace Lifecycle Manager

**zgard** manages the full workspace lifecycle: from cloning repositories and
pulling updates to cross-repo orchestration, inspection, and health monitoring.

## Subcommands

All commands live under ` + "`zgard ws`" + `:

| Command     | Description                                            |
|-------------|--------------------------------------------------------|
| ` + "`init`" + `       | Clone all repositories defined in a workspace          |
| ` + "`pull`" + `       | Pull latest changes for every repository               |
| ` + "`purge`" + `      | Remove workspace directories from disk                 |
| ` + "`exec`" + `       | Run an arbitrary shell command across repositories     |
| ` + "`stash`" + `      | Stash uncommitted changes across repositories          |
| ` + "`checkout`" + `   | Switch branches across repositories                    |
| ` + "`status`" + `     | Show workspace health as monitored by **dukh**         |
| ` + "`search`" + `     | Search for files or content across repositories        |
| ` + "`diff`" + `       | Show uncommitted changes and upstream sync status      |
| ` + "`stats`" + `      | Aggregate commit metadata across repositories          |

Run ` + "`zgard ws <command> --help`" + ` for full documentation of any subcommand.

## Configuration commands

Use **zgard config** to inspect, validate, and query the configuration file:

| Command                  | Description                                          |
|--------------------------|------------------------------------------------------|
| ` + "`config path`" + `     | Print the resolved configuration file path           |
| ` + "`config validate`" + ` | Validate the configuration and report errors         |
| ` + "`config list`" + `     | List all workspaces and projects from the config     |
| ` + "`config get <key>`" + `| Get a specific value by dotted-path (e.g. dukh.port) |`,
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
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable all colored output (overrides NO_COLOR env)")
	cobra.OnInitialize(func() {
		if noColor {
			clr.Disable()
		}
	})

	// Render Long descriptions as styled Markdown via glamour.
	// This is set on the root command so every subcommand inherits it.
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		if cmd.Long != "" {
			fmt.Fprint(out, ui.Render(cmd.Long))
		} else {
			fmt.Fprintln(out, cmd.Short)
			fmt.Fprintln(out)
		}
		if cmd.Runnable() || cmd.HasSubCommands() {
			fmt.Fprint(out, cmd.UsageString())
		}
	})

	rootCmd.AddCommand(ws.NewCmd())
	rootCmd.AddCommand(cfgcmd.NewCmd())
}
