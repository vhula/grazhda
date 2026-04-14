package pkg

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/pkgman"
)

func newUnregisterCmd() *cobra.Command {
	var name string
	var version string
	var all bool

	cmd := &cobra.Command{
		Use:   "unregister",
		Short: "Remove packages from the local registry",
		Long: `Remove one or more packages from ` + "`$GRAZHDA_DIR/registry.pkgs.local.yaml`" + `.

- ` + "`--name`" + ` removes all versions for a package name.
- ` + "`--name`" + ` + ` + "`--version`" + ` removes only an exact name+version.
- ` + "`--all`" + ` removes all local registry entries.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				if name != "" || version != "" {
					return fmt.Errorf("--all is mutually exclusive with --name/--version")
				}
			} else {
				if name == "" {
					return fmt.Errorf("provide --name <name> (or --all)")
				}
			}
			if name == "" && version != "" {
				return fmt.Errorf("--version requires --name")
			}

			dir, err := grazhdaDir()
			if err != nil {
				return err
			}

			localPath := pkgman.LocalRegistryPath(dir)
			local, err := pkgman.LoadLocalRegistry(localPath)
			if err != nil {
				return err
			}

			if all {
				local.Packages = nil
				if err := pkgman.SaveRegistry(localPath, local); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "unregistered all packages from %s\n", localPath)
				return nil
			}

			updated, removed := pkgman.RemovePackage(local, name, version)
			if !removed {
				if version == "" {
					return fmt.Errorf("package %q not found in local registry", name)
				}
				return fmt.Errorf("package %q@%s not found in local registry", name, version)
			}
			if err := pkgman.SaveRegistry(localPath, updated); err != nil {
				return err
			}

			if version == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "unregistered %q from %s\n", name, localPath)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "unregistered %q@%s from %s\n", name, version, localPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Package name to remove from local registry")
	cmd.Flags().StringVar(&version, "version", "", "Optional package version for exact-match removal")
	cmd.Flags().BoolVar(&all, "all", false, "Unregister all local packages")
	return cmd
}
