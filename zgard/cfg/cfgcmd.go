// Package cfg implements the `zgard config` subcommand suite for
// inspecting and validating the Grazhda configuration file.
package cfg

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/path"
	"github.com/vhula/grazhda/internal/reporter"
)

// NewCmd returns the `config` parent command with all subcommands registered.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and validate the Grazhda configuration",
		Long: `# zgard config — Configuration Files Manager

Inspect, validate, and query the Grazhda configuration file.

The configuration is loaded from **$GRAZHDA_DIR/config.yaml** when
` + "`GRAZHDA_DIR`" + ` is set, otherwise from **~/.grazhda/config.yaml**.

## Subcommands

| Command    | Description                                               |
|------------|-----------------------------------------------------------|
| ` + "`path`" + `     | Print the resolved path of the configuration file         |
| ` + "`validate`" + ` | Validate the configuration and report any errors          |
| ` + "`list`" + `     | List all workspaces and their projects from the config    |
| ` + "`get <key>`" + `| Get a specific value by dotted-path key                   |
| ` + "`edit`" + `      | Open config.yaml in the configured editor                 |`,
	}

	cmd.AddCommand(newPathCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newEditCmd(executor.OsExecutor{}))
	return cmd
}

// resolveConfigPath delegates to the shared config.ConfigPath helper.
func resolveConfigPath() string { return path.ConfigPath() }

func newPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the resolved configuration file path",
		Long: `Print the absolute path of the Grazhda configuration file.

The path is resolved from the **GRAZHDA_DIR** environment variable
(appending **config.yaml**), or falls back to **~/.grazhda/config.yaml**
when the variable is unset.

Useful for debugging environment setup or scripting config file edits.`,
		Example: `  # Print the default config path
  zgard config path

  # Use in a script to edit the config
  $EDITOR $(zgard config path)

  # Override the config location with GRAZHDA_DIR
  GRAZHDA_DIR=/opt/grazhda zgard config path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), resolveConfigPath())
			return nil
		},
	}
}

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration file and report errors",
		Long: `Load and validate the Grazhda configuration file.

Checks all required fields (workspace names, paths, clone command templates,
project branches) and reports every validation error with its field path.

Exits with status **0** on success and status **1** if any errors are found.`,
		Example: `  # Validate the default config
  zgard config validate

  # Validate an alternate config directory
  GRAZHDA_DIR=/opt/grazhda zgard config validate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := resolveConfigPath()
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("%s", clr.Red("✗ "+err.Error()))
			}

			errs := config.Validate(cfg)
			if len(errs) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%s Config is valid  (%s)\n",
					clr.Green("✓"), cfgPath)
				return nil
			}

			for _, e := range errs {
				fmt.Fprintln(os.Stderr, clr.Red("  ✗ "+e))
			}
			fmt.Fprintf(os.Stderr, "\n%s %d error(s) found in %s\n",
				clr.Red("✗"), len(errs), cfgPath)
			return reporter.ExitError{Code: 1}
		},
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all workspaces and their projects from the config",
		Long: `Display the full workspace hierarchy from the configuration file.

For each workspace, shows its name, path, and a project summary including
the default branch and repository count. No filesystem access is performed —
the output reflects the **config** only, not the current clone status.

Use **zgard ws list** to see real-time clone status for each repository.`,
		Example: `  # List all workspaces from the config
  zgard config list

  # List from an alternate config directory
  GRAZHDA_DIR=/opt/grazhda zgard config list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(resolveConfigPath())
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s %d workspace(s)\n\n",
				clr.Blue("Workspaces:"), len(cfg.Workspaces))

			for _, ws := range cfg.Workspaces {
				label := clr.Blue(ws.Name)
				if ws.Default || ws.Name == "default" {
					label += clr.Yellow("  [default]")
				}
				fmt.Fprintf(out, "workspace: %s\n", label)
				fmt.Fprintf(out, "  path: %s\n", ws.Path)

				totalRepos := 0
				for _, p := range ws.Projects {
					totalRepos += len(p.Repositories)
				}
				fmt.Fprintf(out, "  projects: %d  repositories: %d\n",
					len(ws.Projects), totalRepos)

				for _, proj := range ws.Projects {
					repoCount := len(proj.Repositories)
					tagStr := ""
					if len(proj.Tags) > 0 {
						tagStr = fmt.Sprintf("  [%s]", strings.Join(proj.Tags, ", "))
					}
					fmt.Fprintf(out, "    %s %-20s branch: %-14s repos: %d%s\n",
						clr.Blue("→"),
						proj.Name,
						proj.Branch,
						repoCount,
						clr.Yellow(tagStr),
					)
				}
				fmt.Fprintln(out)
			}
			return nil
		},
	}
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value by dotted-path key",
		Long: `Retrieve a single value from the configuration using a dotted-path key.

Key segments correspond to YAML field names as they appear in config.yaml.
Array elements are addressed by zero-based integer index.

- Scalar values are printed as plain text.
- Complex values (maps, lists) are printed as compact YAML.

Exits with status **1** if the key does not exist.`,
		Example: `  # Get the default editor
  zgard config get editor

  # Get the dukh port number
  zgard config get dukh.port

  # Get the install directory
  zgard config get general.install_dir

  # Get the name of the first workspace
  zgard config get workspaces.0.name

  # Get the full dukh configuration block
  zgard config get dukh

  # Get the clone command template for the first workspace
  zgard config get workspaces.0.clone_command_template`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(resolveConfigPath())
			if err != nil {
				return err
			}

			val, err := config.GetByPath(cfg, args[0])
			if err != nil {
				return fmt.Errorf("%s", clr.Red("✗ "+err.Error()))
			}
			fmt.Fprintln(cmd.OutOrStdout(), val)
			return nil
		},
	}
}

func newEditCmd(exec executor.Executor) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open config.yaml in the configured editor",
		Long: `Open the Grazhda configuration file in an editor.

Editor resolution order:
1. **editor** field in config.yaml (or **GRAZHDA_EDITOR** env override)
2. **$VISUAL** environment variable
3. **$EDITOR** environment variable
4. **vi** as a fallback`,
		Example: `  # Open config in the configured editor
  zgard config edit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := resolveConfigPath()
			if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
				return fmt.Errorf("%s\n%s",
					clr.Red("✗ config file not found: "+cfgPath),
					clr.Yellow("Run the Grazhda installer to create it."))
			}

			cfg, _ := config.Load(cfgPath)
			config.ApplyEnvOverrides(cfg)

			editorBin := cfg.Editor
			if editorBin == "" {
				editorBin = os.Getenv("VISUAL")
			}
			if editorBin == "" {
				editorBin = os.Getenv("EDITOR")
			}
			if editorBin == "" {
				editorBin = "vi"
			}

			// Shell-quote the path so spaces and special characters are safe under sh -c.
			quotedPath := "'" + strings.ReplaceAll(cfgPath, "'", "'\\''") + "'"
			return exec.RunInteractive(cmd.Context(), ".", editorBin+" "+quotedPath)
		},
	}
}
