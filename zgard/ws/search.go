package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/workspace"
)

func newSearchCmd() *cobra.Command {
	var glob bool
	var regex bool
	var parallel bool

	cmd := &cobra.Command{
		Use:   "search <pattern>",
		Short: "Search for a pattern across repositories in a workspace",
		Long: `Search file contents or filenames across all resolved repositories.

Three search modes are available:

- **Default (grep):** case-sensitive substring match of file contents
- **--glob:** match filenames using glob syntax (e.g. "*.go", "Makefile")
- **--regex:** treat the pattern as a Go regular expression

Binary files and **.git** directories are automatically skipped.
Use targeting flags to narrow the search scope to a project or repository.`,
		Example: `  # Search for a string in all file contents (default grep mode)
  zgard ws search "TODO: remove"

  # Find all Go source files by name
  zgard ws search --glob "*.go"

  # Search using a regular expression
  zgard ws search --regex "func.*Handler"

  # Narrow search to a specific project
  zgard ws search -p backend "connectionString"

  # Search only a specific repository
  zgard ws search -p backend -r api-service "db.Connect"

  # Find all YAML config files across all workspaces
  zgard ws search --all --glob "*.yaml"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			opts := workspace.SearchOptions{
				InspectOptions: workspace.InspectOptions{
					Parallel:    parallel,
					ProjectName: projectName,
					RepoName:    repoName,
					Tags:        tagFilter,
				},
				Pattern: args[0],
				Glob:    glob,
				Regex:   regex,
			}

			for _, ws := range workspaces {
				if err := workspace.Search(ws, opts, os.Stdout); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&glob, "glob", false, "Match filenames instead of file contents")
	cmd.Flags().BoolVar(&regex, "regex", false, "Treat pattern as a Go regular expression")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Search all repositories concurrently")

	return cmd
}
