package ws

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/workspace"
)

// findIDEBinary resolves the CLI binary for the requested IDE.
// It tries exec.LookPath first, then a small set of hardcoded fallback paths.
func findIDEBinary(ide string) (string, error) {
	switch ide {
	case "vscode", "code":
		for _, name := range []string{"code", "code-insiders"} {
			if p, err := exec.LookPath(name); err == nil {
				return p, nil
			}
		}
		return "", fmt.Errorf(
			"VS Code CLI 'code' not found on PATH\n  Install VS Code and run: Shell Command → Install 'code' command in PATH",
		)
	case "idea":
		for _, name := range []string{"idea", "idea.sh"} {
			if p, err := exec.LookPath(name); err == nil {
				return p, nil
			}
		}
		for _, p := range []string{"/usr/local/bin/idea", "/opt/idea/bin/idea.sh"} {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
		return "", fmt.Errorf(
			"IntelliJ IDEA CLI 'idea' not found on PATH\n  Add the JetBrains Toolbox shell scripts directory to your PATH",
		)
	default:
		return "", fmt.Errorf("unsupported IDE %q; supported values: vscode, idea", ide)
	}
}

func newOpenCmd() *cobra.Command {
	var ide string

	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open targeted repositories in an IDE",
		Long: `Resolve the targeted repository directories and launch an IDE on each one.

Supported IDEs:
  vscode   — launches 'code' (or 'code-insiders') for each repo directory
  idea     — launches 'idea' (or 'idea.sh') for each repo directory

Use targeting flags (-p, -r, -t) to limit which repositories are opened.
When more than 5 windows would open, a warning is printed first.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if ide == "" {
				return fmt.Errorf("--ide is required; supported values: vscode, idea")
			}

			binary, err := findIDEBinary(ide)
			if err != nil {
				return err
			}

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			workspaces, err := workspace.Resolve(cfg, wsName, wsAll)
			if err != nil {
				return err
			}

			if wsName == "" && !wsAll {
				warnDefaultTarget(os.Stderr, workspaces[0])
			}

			opts := workspace.RunOptions{
				ProjectName: projectName,
				RepoName:    repoName,
				Tags:        tagFilter,
			}

			var allPaths []string
			for _, ws := range workspaces {
				paths, err := workspace.CollectRepoPaths(ws, opts)
				if err != nil {
					return err
				}
				allPaths = append(allPaths, paths...)
			}

			if len(allPaths) == 0 {
				fmt.Fprintln(os.Stderr, clr.Yellow("Warning: no repositories matched the targeting filters"))
				return nil
			}

			ideLabel := "VS Code"
			if ide == "idea" {
				ideLabel = "IntelliJ IDEA"
			}

			if len(allPaths) > 5 {
				fmt.Fprintf(os.Stderr, "%s\n",
					clr.Yellow(fmt.Sprintf("Warning: %d IDE windows will open", len(allPaths))))
			}
			fmt.Fprintf(os.Stdout, "Opening %d repo(s) in %s...\n", len(allPaths), ideLabel)

			for _, path := range allPaths {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					fmt.Fprintf(os.Stdout, "  %s %s — not cloned, skipped\n", clr.Yellow("⏭"), path)
					continue
				}
				c := exec.Command(binary, path)
				if err := c.Start(); err != nil {
					fmt.Fprintf(os.Stderr, "  %s %s — %s\n", clr.Red("✗"), path, clr.Red(err.Error()))
					continue
				}
				fmt.Fprintf(os.Stdout, "  %s %s\n", clr.Green("✓"), path)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ide, "ide", "", "IDE to open: vscode or idea (required)")

	return cmd
}
