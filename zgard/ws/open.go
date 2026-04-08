package ws

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// commonAncestor returns the longest common directory prefix for a set of
// absolute paths. For a single path it returns that path unchanged.
// All paths must be absolute and clean (filepath.Clean applied).
func commonAncestor(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return paths[0]
	}

	// Split each path into segments and find the common prefix.
	split := func(p string) []string { return strings.Split(p, string(filepath.Separator)) }
	segments := split(paths[0])
	for _, p := range paths[1:] {
		parts := split(p)
		min := len(segments)
		if len(parts) < min {
			min = len(parts)
		}
		end := 0
		for i := 0; i < min; i++ {
			if segments[i] != parts[i] {
				break
			}
			end = i + 1
		}
		segments = segments[:end]
	}
	if len(segments) == 0 {
		return string(filepath.Separator)
	}
	return strings.Join(segments, string(filepath.Separator))
}

// launchIDE executes the IDE binary exactly once in single-window mode.
//
//   - VS Code: passes all paths as separate arguments — `code path1 path2 …`
//     which opens a multi-root workspace window.
//   - IntelliJ IDEA: passes only the common ancestor of all paths — `idea <dir>`
//     because the IDEA CLI accepts a single project root.
func launchIDE(binary, ide string, paths []string) error {
	var cmdArgs []string
	switch ide {
	case "idea":
		// IntelliJ supports a single directory; use the common ancestor so the
		// user can navigate all cloned repos from one project window.
		cmdArgs = []string{commonAncestor(paths)}
	default: // vscode / code
		// VS Code opens all paths in one multi-root workspace window when they
		// are passed as separate positional arguments.
		cmdArgs = paths
	}

	c := exec.Command(binary, cmdArgs...) //nolint:gosec // binary is resolved via LookPath
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Start()
}

func newOpenCmd() *cobra.Command {
	var ide string

	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open targeted repositories in a single IDE window",
		Long: `Aggregate all targeted repository directories and launch an IDE exactly once.

Single-window policy:
  vscode — passes all repo paths as arguments to 'code', opening a
           VS Code multi-root workspace in a single window.
  idea   — passes the common ancestor directory to 'idea', opening
           all repos under one IntelliJ project root.

Repositories that have not been cloned yet are listed as skipped and
excluded from the IDE invocation.

Use targeting flags (-p, -r, -t) to control which repositories are included.`,
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

			// Collect all candidate paths from every targeted workspace.
			var candidatePaths []string
			for _, ws := range workspaces {
				paths, err := workspace.CollectRepoPaths(ws, opts)
				if err != nil {
					return err
				}
				candidatePaths = append(candidatePaths, paths...)
			}

			if len(candidatePaths) == 0 {
				fmt.Fprintln(os.Stderr, clr.Red("Error: no repositories matched the targeting filters"))
				return fmt.Errorf("no repositories matched")
			}

			// Separate cloned repos from missing ones.
			var validPaths []string
			for _, path := range candidatePaths {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					fmt.Fprintf(os.Stderr, "  %s %s — not cloned, skipped\n", clr.Yellow("⏭"), path)
					continue
				}
				validPaths = append(validPaths, path)
			}

			if len(validPaths) == 0 {
				fmt.Fprintln(os.Stderr, clr.Yellow("Warning: all matched repositories are not yet cloned; nothing to open"))
				return nil
			}

			ideLabel := "VS Code"
			if ide == "idea" {
				ideLabel = "IntelliJ IDEA"
			}

			// Print the single opening message before launching.
			fmt.Fprintf(os.Stdout, "Opening %d %s in one %s window...\n",
				len(validPaths),
				pluralRepos(len(validPaths)),
				ideLabel,
			)
			for _, p := range validPaths {
				fmt.Fprintf(os.Stdout, "  %s %s\n", clr.Green("✓"), p)
			}

			if err := launchIDE(binary, ide, validPaths); err != nil {
				return fmt.Errorf("launching %s: %w", ideLabel, err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ide, "ide", "", "IDE to open: vscode or idea (required)")

	return cmd
}

func pluralRepos(n int) string {
	if n == 1 {
		return "repository"
	}
	return "repositories"
}
