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

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install one or all packages from the registry",
		Long: `Install packages declared in ` + "`$GRAZHDA_DIR/.grazhda.pkgs.yaml`" + `.

Dependencies are resolved automatically via topological sort (Kahn's algorithm).
Each package runs through three lifecycle phases in order:

  1. **pre-install** — environment setup and assertions
  2. **install**     — primary download / compilation
  3. **post-install** — init sourcing and PATH fixups

After a successful install the package's ` + "`env`" + ` block (if declared) is
written into ` + "`$GRAZHDA_DIR/.grazhda.env`" + ` between named markers so your shell
picks it up on next login.`,

		Example: `  # Install a single package (deps resolved automatically)
  zgard pkg install --name jdk

  # Install all packages in dependency order
  zgard pkg install --all`,

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

			installer := pkgman.NewInstaller(dir, reg, os.Stdout, os.Stderr)

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
	return cmd
}
