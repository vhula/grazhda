package ws

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/workspace"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories in a workspace with their clone status",
		Long: `Display a tree-formatted view of workspaces, projects, and repositories.

For each repository, the **clone status** is shown by checking whether the
expected directory exists on disk. Targeting flags narrow the output to
a specific workspace, project, or repository.

A summary line at the end counts total cloned and not-yet-cloned repositories.`,
		Example: `  # List all repos in the default workspace
  zgard ws list

  # List repos in a named workspace
  zgard ws list -n myworkspace

  # List repos across all workspaces
  zgard ws list --all

  # List repos in a specific project
  zgard ws list -n myworkspace -p backend

  # List repos matching a name substring
  zgard ws list -n myworkspace -p backend -r api

  # List repos with a specific tag
  zgard ws list -t backend`,
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

			opts := workspace.InspectOptions{
				ProjectName: projectName,
				RepoName:    repoName,
				Tags:        tagFilter,
			}

			out := cmd.OutOrStdout()
			totalCloned, totalMissing := 0, 0

			for _, ws := range workspaces {
				label := clr.Blue(ws.Name)
				if ws.Default || ws.Name == "default" {
					label += clr.Yellow("  [default]")
				}
				fmt.Fprintf(out, "workspace: %s\n", label)
				fmt.Fprintf(out, "  path: %s\n\n", clr.Blue(workspace.ExpandHome(ws.Path)))

				// Validate filters — report error but continue to next workspace.
				if err := workspace.ValidateFilters(ws, workspace.RunOptions{
					ProjectName: opts.ProjectName,
					RepoName:    opts.RepoName,
					Tags:        opts.Tags,
				}); err != nil {
					fmt.Fprintf(os.Stderr, "  %s %s\n\n", clr.Red("✗"), err.Error())
					continue
				}

				// Resolve all matching repos (includes clone status).
				infos := workspace.ResolveRepoInfos(ws, opts)

				// Group results by project, preserving order.
				type projGroup struct {
					Branch string
					Tags   []string
					Repos  []workspace.RepoInfo
				}
				var projectOrder []string
				byProject := map[string]*projGroup{}
				for _, info := range infos {
					if _, exists := byProject[info.ProjectName]; !exists {
						projectOrder = append(projectOrder, info.ProjectName)
						byProject[info.ProjectName] = &projGroup{
							Branch: info.ProjectBranch,
							Tags:   info.ProjectTags,
						}
					}
					byProject[info.ProjectName].Repos = append(
						byProject[info.ProjectName].Repos, info)
				}

				wsCloned, wsMissing := 0, 0
				for _, projName := range projectOrder {
					g := byProject[projName]

					tagStr := ""
					if len(g.Tags) > 0 {
						tagStr = fmt.Sprintf("  [%s]", strings.Join(g.Tags, ", "))
					}
					fmt.Fprintf(out, "  %s %s  (branch: %s)%s\n",
						clr.Blue("project:"),
						projName,
						g.Branch,
						clr.Yellow(tagStr),
					)

					for _, info := range g.Repos {
						repoTags := ""
						if len(info.Tags) > 0 {
							repoTags = fmt.Sprintf("  [%s]",
								strings.Join(info.Tags, ", "))
						}
						if info.Cloned {
							wsCloned++
							fmt.Fprintf(out, "    %s  %-22s %s%s\n",
								clr.Green("✓"),
								info.Name,
								clr.Blue(info.LocalPath),
								clr.Yellow(repoTags),
							)
						} else {
							wsMissing++
							fmt.Fprintf(out, "    %s  %-22s %s%s\n",
								clr.Red("✗"),
								info.Name,
								clr.Yellow("(not cloned)"),
								clr.Yellow(repoTags),
							)
						}
					}
					fmt.Fprintln(out)
				}

				totalCloned += wsCloned
				totalMissing += wsMissing
			}

			fmt.Fprintf(out, "%s %d cloned  %s %d not cloned\n",
				clr.Green("✓"), totalCloned,
				clr.Red("✗"), totalMissing,
			)
			return nil
		},
	}

	return cmd
}
