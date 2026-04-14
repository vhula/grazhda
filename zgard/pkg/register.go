package pkg

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/pkgman"
)

func newRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Interactively register a package in the local registry",
		Long: `Create or update a package entry in ` + "`$GRAZHDA_DIR/registry.pkgs.local.yaml`" + `.

The prompt asks for all package fields used by install/purge flows, including
env hooks and scripts. Existing packages are listed for depends_on selection.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := grazhdaDir()
			if err != nil {
				return err
			}

			merged, err := loadMergedRegistry(dir)
			if err != nil {
				return err
			}

			localPath := pkgman.LocalRegistryPath(dir)
			local, err := pkgman.LoadLocalRegistry(localPath)
			if err != nil {
				return err
			}

			in := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			name, err := promptRequired(in, out, "Package name")
			if err != nil {
				return err
			}
			version, err := promptOptional(in, out, "Version (optional)")
			if err != nil {
				return err
			}
			preCreateDir, err := promptBool(in, out, "Pre-create package directory? [y/N]")
			if err != nil {
				return err
			}
			dependsOn, err := promptDependsOn(in, out, merged)
			if err != nil {
				return err
			}
			preInstallEnv, err := promptMultiline(in, out, "pre_install_env")
			if err != nil {
				return err
			}
			install, err := promptMultiline(in, out, "install script")
			if err != nil {
				return err
			}
			postInstallEnv, err := promptMultiline(in, out, "post_install_env")
			if err != nil {
				return err
			}
			purge, err := promptMultiline(in, out, "purge script")
			if err != nil {
				return err
			}

			pkg := pkgman.Package{
				Name:           name,
				Version:        version,
				PreCreateDir:   preCreateDir,
				DependsOn:      dependsOn,
				PreInstallEnv:  preInstallEnv,
				Install:        install,
				PostInstallEnv: postInstallEnv,
				Purge:          purge,
			}
			pkgman.AddPackage(local, pkg)
			if err := pkgman.SaveRegistry(localPath, local); err != nil {
				return err
			}
			fmt.Fprintf(out, "%s registered %s in %s\n", clr.Green("✓"), pkgman.PkgLabel(pkg), localPath)
			return nil
		},
	}
	return cmd
}

func promptRequired(in *bufio.Reader, out io.Writer, label string) (string, error) {
	for {
		fmt.Fprintf(out, "%s: ", label)
		line, err := in.ReadString('\n')
		if err != nil {
			return "", err
		}
		v := strings.TrimSpace(line)
		if v != "" {
			return v, nil
		}
		fmt.Fprintln(out, "value is required")
	}
}

func promptOptional(in *bufio.Reader, out io.Writer, label string) (string, error) {
	fmt.Fprintf(out, "%s: ", label)
	line, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptBool(in *bufio.Reader, out io.Writer, label string) (bool, error) {
	fmt.Fprintf(out, "%s ", label)
	line, err := in.ReadString('\n')
	if err != nil {
		return false, err
	}
	v := strings.TrimSpace(strings.ToLower(line))
	return v == "y" || v == "yes", nil
}

func promptDependsOn(in *bufio.Reader, out io.Writer, reg *pkgman.Registry) ([]string, error) {
	refs := make([]string, 0, len(reg.Packages))
	for _, p := range reg.Packages {
		refs = append(refs, pkgman.PkgLabel(p))
	}
	slices.Sort(refs)
	if len(refs) == 0 {
		fmt.Fprintln(out, "depends_on: no existing packages to select")
		return nil, nil
	}

	fmt.Fprintln(out, "depends_on selection (space-separated numbers, Enter for none):")
	for i, ref := range refs {
		fmt.Fprintf(out, "  %d) %s\n", i+1, ref)
	}
	fmt.Fprint(out, "select: ")
	line, err := in.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	parts := strings.Fields(strings.ReplaceAll(line, ",", " "))
	outDeps := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		var idx int
		_, scanErr := fmt.Sscanf(part, "%d", &idx)
		if scanErr != nil || idx < 1 || idx > len(refs) {
			return nil, fmt.Errorf("invalid depends_on selection %q", part)
		}
		ref := refs[idx-1]
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		outDeps = append(outDeps, ref)
	}
	return outDeps, nil
}

func promptMultiline(in *bufio.Reader, out io.Writer, label string) (string, error) {
	fmt.Fprintf(out, "%s (finish with empty line):\n", label)
	var lines []string
	for {
		line, err := in.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), nil
}
