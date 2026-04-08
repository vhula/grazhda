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
	var parallelAll bool

	cmd := &cobra.Command{
		Use:   "search <pattern>",
		Short: "Search for a pattern across repositories in a workspace",
		Long: `Search files across all resolved repositories.

By default, performs a case-sensitive substring grep of file contents.
Use --glob to match filenames instead, or --regex to treat the pattern as a Go regular expression.
Binary files and .git directories are automatically skipped.`,
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
					ParallelAll: parallelAll,
					ProjectName: projectName,
					RepoName:    repoName,
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
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Search within each project concurrently")
	cmd.Flags().BoolVar(&parallelAll, "parallel-all", false, "Search across all projects concurrently")

	return cmd
}
