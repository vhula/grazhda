package pkg

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/pkgman"
)

func newInstallCmd() *cobra.Command {
	var pkgName string
	var all bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install one or all packages from the registry",
		Long: `Install packages declared in ` + "`$GRAZHDA_DIR/.grazhda.pkgs.yaml`" + `.

Dependencies are resolved automatically via topological sort (Kahn's algorithm).
Each package runs through the following lifecycle:

  1. pre_install_env block is written to ` + "`$GRAZHDA_DIR/.grazhda.env`" + ` (if declared)
     and the env file is sourced so the install script sees the exported vars.
  2. **install** script runs (with env file pre-sourced).
  3. post_install_env block is written to ` + "`$GRAZHDA_DIR/.grazhda.env`" + ` (if declared)
     and the env file is sourced again for subsequent packages.

By default, script output is suppressed and a spinner indicates progress.
Pass --verbose to stream raw script stdout/stderr to the terminal.`,

		Example: `  # Install a single package (deps resolved automatically)
  zgard pkg install --name jdk

  # Install all packages in dependency order
  zgard pkg install --all

  # Install all packages with full script output
  zgard pkg install --all --verbose`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if !all && pkgName == "" {
				return fmt.Errorf("provide --name <pkg> or --all")
			}
			if all && pkgName != "" {
				return fmt.Errorf("--all and --name are mutually exclusive")
			}

			dir, err := grazhdaDir()
			if err != nil {
				return err
			}

			reg, err := pkgman.LoadRegistry(pkgman.RegistryPath(dir))
			if err != nil {
				return fmt.Errorf("load registry: %w", err)
			}

			installer := pkgman.NewInstaller(dir, reg, os.Stdout, os.Stderr, verbose)

			var names []string
			if pkgName != "" {
				names = []string{pkgName}
			}

			if err := installer.Install(cmd.Context(), names); err != nil {
				return fmt.Errorf("install: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&pkgName, "name", "n", "", "Name of the package to install")
	cmd.Flags().BoolVar(&all, "all", false, "Install all packages in the registry")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Stream script output to the terminal")
	return cmd
}
