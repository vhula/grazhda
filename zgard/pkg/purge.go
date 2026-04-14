package pkg

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/pkgman"
)

func newPurgeCmd() *cobra.Command {
	var pkgName string
	var all bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Remove one or all installed packages",
		Long: `Remove packages installed by zgard and clean up their shell environment.

Packages are purged in **reverse** topological order so dependents are always
removed before their dependencies. For each package:

  1. The optional **purge** script runs (unregistering tool versions, etc.)
  2. The package directory ` + "`$GRAZHDA_DIR/pkgs/<name>`" + ` is deleted (if ` + "`pre_create_dir: true`" + `)
  3. The named env block is excised from ` + "`$GRAZHDA_DIR/.grazhda.env`" + ``,

		Example: `  # Remove a single package (and its env block)
  zgard pkg purge --name sdkman

  # Remove every installed package
  zgard pkg purge --all`,

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

			reg, err := loadMergedRegistry(dir)
			if err != nil {
				return fmt.Errorf("load registry: %w", err)
			}

			purger := pkgman.NewPurger(dir, reg, os.Stdout, os.Stderr, verbose)

			var names []string
			if pkgName != "" {
				names = []string{pkgName}
			}

			if err := purger.Purge(cmd.Context(), names); err != nil {
				return fmt.Errorf("purge: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&pkgName, "name", "n", "", "Package ref to purge: <name> or <name>@<version>")
	cmd.Flags().BoolVar(&all, "all", false, "Purge all packages listed in the registry")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Stream script output to the terminal")
	return cmd
}
