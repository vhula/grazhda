// Package pkg provides the `zgard pkg` command suite for declarative package
// management within the grazhda ecosystem.
package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// grazhdaDir returns $GRAZHDA_DIR, defaulting to $HOME/.grazhda.
func grazhdaDir() (string, error) {
	if dir := os.Getenv("GRAZHDA_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(home, ".grazhda"), nil
}

// NewCmd returns the `pkg` parent command with install and purge subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: "Declarative package management inside $GRAZHDA_DIR",
		Long: `# zgard pkg — Declarative Package Manager

**zgard pkg** installs and purges developer tools (SDKs, CLIs, runtimes) inside
` + "`$GRAZHDA_DIR/pkgs/`" + ` so they never contaminate the host OS.

Packages are declared in **` + "`$GRAZHDA_DIR/.grazhda.pkgs.yaml`" + `** (the _registry_).
Dependencies are resolved automatically in topological order via a DAG engine,
guaranteeing that every dependency is installed before its dependents.

After installation, shell environment variables are written into
**` + "`$GRAZHDA_DIR/.grazhda.env`" + `** inside idempotent named blocks so they are
available in every new shell session.

Each package supports two env blocks:

- **` + "`pre_install_env`" + `** — written before the install script runs, then
  ` + "`$GRAZHDA_DIR/.grazhda.env`" + ` is sourced so the install script sees the
  exported variables (e.g. ` + "`SDKMAN_DIR`" + ` before installing via sdkman).
- **` + "`post_install_env`" + `** — written after the install script succeeds, then
  ` + "`$GRAZHDA_DIR/.grazhda.env`" + ` is sourced so subsequent packages see the
  exported variables.

## Subcommands

| Command                    | Description                                       |
|----------------------------|---------------------------------------------------|
| ` + "`pkg install`" + `           | Install packages from the registry                |
| ` + "`pkg purge`" + `             | Remove packages and excise their env blocks       |

Run ` + "`zgard pkg <command> --help`" + ` for full documentation.`,
	}

	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newPurgeCmd())
	return cmd
}
