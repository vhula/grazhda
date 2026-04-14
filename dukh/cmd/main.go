package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/ui"
)

// version is overridden at build time via:
//
//	-ldflags "-X main.version=<tag>"
var version = "dev"

var noColor bool

var rootCmd = &cobra.Command{
	Use:           "dukh",
	Short:         "Dukh — Grazhda workspace health monitor",
	Version:       version,
	Long:          rootLong,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var rootLong = `# dukh — Workspace Health Monitor

**dukh** is a lightweight background daemon that continuously watches your
Grazhda workspaces and reports repository health in real time over gRPC.

It pairs with ` + "`zgard ws status`" + ` to give you an instant snapshot of
branch state, uncommitted changes, and upstream drift — without manually
running ` + "`git status`" + ` in every repository.

## How it works

1. ` + "`dukh start`" + ` launches a background daemon (detached process).
2. The daemon reads ` + "`$GRAZHDA_DIR/config.yaml`" + ` and scans every repository
   on a configurable interval.
3. Results are served via gRPC — ` + "`zgard ws status`" + ` and ` + "`dukh status`" + `
   query this endpoint.
4. ` + "`dukh stop`" + ` gracefully shuts the daemon down.

## Commands

| Command   | Description                                        |
|-----------|----------------------------------------------------|
| ` + "`start`" + `   | Launch the health-monitor daemon in the background |
| ` + "`stop`" + `    | Gracefully stop the running daemon                 |
| ` + "`status`" + `  | Show whether dukh is running (PID, uptime)         |
| ` + "`scan`" + `    | Trigger an immediate workspace rescan              |

Run ` + "`dukh <command> --help`" + ` for full documentation of any subcommand.

## Configuration

dukh reads its settings from the ` + "`dukh:`" + ` section in
` + "`$GRAZHDA_DIR/config.yaml`" + `:

` + "```yaml" + `
dukh:
  port: 50051          # gRPC listen port (default 50051)
  scan_interval: 30s   # rescan period     (default 30s)
` + "```" + `
`

func main() {
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

	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(stopCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(scanCmd())
}
