package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/ui"
	"github.com/vhula/grazhda/zgard/cfg"
	"github.com/vhula/grazhda/zgard/pkg"
	"github.com/vhula/grazhda/zgard/ws"
)

// version is overridden at build time via:
//
//	-ldflags "-X main.version=<tag>"
var version = "dev"

var noColor bool

var rootCmd = &cobra.Command{
	Use:           "zgard",
	Short:         "Workspace lifecycle manager",
	Version:       version,
	Long:          rootLong,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// rootLong is the markdown long description for the root command.
var rootLong = `# zgard — Workspace Lifecycle Manager

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

## Package management

Use **zgard pkg** to install and manage developer tools (SDKs, CLIs, runtimes)
inside ` + "`$GRAZHDA_DIR/pkgs/`" + ` — isolated from the host OS.

Registries:
- Global (managed by install/upgrade): ` + "`$GRAZHDA_DIR/.grazhda.pkgs.yaml`" + `
- Local (user-managed): ` + "`$GRAZHDA_DIR/registry.pkgs.local.yaml`" + `

During install and purge, both registries are merged. Local entries override
global entries when **name+version** match exactly. Shell env blocks are written
to ` + "`$GRAZHDA_DIR/.grazhda.env`" + `.

| Command          | Description                                               |
|------------------|-----------------------------------------------------------|
| ` + "`pkg install`" + `   | Install one or all packages (deps resolved automatically) |
| ` + "`pkg purge`" + `     | Remove packages and excise their env blocks               |
| ` + "`pkg register`" + `  | Interactively add/update a package in local registry      |
| ` + "`pkg unregister`" + `| Remove one or all packages from local registry            |

Run ` + "`zgard pkg <command> --help`" + ` for full documentation.

## Configuration commands

Use **zgard config** to inspect, validate, and query the configuration file:

| Command                  | Description                                          |
|--------------------------|------------------------------------------------------|
| ` + "`config path`" + `     | Print the resolved configuration file path           |
| ` + "`config validate`" + ` | Validate the configuration and report errors         |
| ` + "`config list`" + `     | List all workspaces and projects from the config     |
| ` + "`config get <key>`" + `| Get a specific value by dotted-path (e.g. dukh.port) |`

// Execute wires a signal-aware context, runs the root Cobra command, and maps
// reporter.ExitError values to the appropriate process exit code.
func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		var exitErr reporter.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, clr.Red(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable all colored output (overrides NO_COLOR env)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress all output except errors")
	cobra.OnInitialize(func() {
		if noColor {
			clr.Disable()
		}
	})

	// Render Long descriptions as styled Markdown via glamour.
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
	rootCmd.AddCommand(cfg.NewCmd())
	rootCmd.AddCommand(pkg.NewCmd())
}
