# Grazhda — CLI Improvement Suggestions

> Grounded in the 14 tips from [_14 Great Tips to Make Amazing CLI Applications_](https://dev.to/wesen/14-great-tips-to-make-amazing-cli-applications-3gp3).  
> Each section maps the tip to grazhda's current state and proposes concrete, actionable improvements.

---

## 1. Use a Command-Line Parsing Framework

**Current state:** ✅ Cobra is used throughout `zgard` and `dukh`. Persistent flags are properly inherited via `PersistentFlags` on the `ws` parent command.

**Improvement:**  
Cobra's `ValidArgsFunction` API enables **dynamic shell completion** — tab-completing workspace names, project names, and tag values straight from the live config. Currently `zgard ws init --name [TAB]` returns nothing.

```go
// Example: complete --name from config
cmd.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    cfg, err := loadConfig()
    if err != nil {
        return nil, cobra.ShellCompDirectiveError
    }
    var names []string
    for _, ws := range cfg.Workspaces {
        names = append(names, ws.Name)
    }
    return names, cobra.ShellCompDirectiveNoFileComp
})
```

Apply this to `--name`, `--project-name`, `--repo-name`, and `--tag` across all subcommands. It turns the targeting system into a discoverable, fast-to-type interface.

---

## 2. Distribute as a Single Binary with No Dependencies

**Current state:** ✅ Go compiles static binaries (`bin/zgard`, `bin/dukh`). `grazhda-install.sh` builds from source. The install script correctly handles `GRAZHDA_DIR`.

**Improvement:**  
The install script requires users to pre-install **five tools** (`bash`, `curl`, `git`, `just`, `protoc`, Go 1.26+). This is a high barrier for the "just try it out" audience.

Consider **publishing pre-compiled GitHub Releases** for common platforms (Linux/amd64, macOS/arm64, macOS/amd64) via a GitHub Actions release workflow. The installer could detect whether Go is available and fall back to downloading a release binary:

```bash
# Simplified installer fallback
if ! command -v go &>/dev/null; then
    curl -sL "https://github.com/vhula/grazhda/releases/latest/download/zgard-linux-amd64.tar.gz" | tar xz -C "$GRAZHDA_DIR/bin/"
fi
```

This separates the _developer_ experience (build from source) from the _user_ experience (run a pre-built binary).

---

## 3. Make Documentation Part of the Tool Itself

**Current state:** ⚠️ Commands have `Short` and `Long` descriptions but no `Example` fields.

**Improvement:**  
Cobra has a dedicated `Example` field that is rendered in a distinct section of `--help` output. Every subcommand should populate it. This is where tools like `kubectl`, `gh`, and `docker` are most helpful.

```go
cmd := &cobra.Command{
    Use:   "exec <command> [args...]",
    Short: "Fan out a shell command to all repositories in a workspace",
    Long:  `...`,
    Example: `  # Run tests in every repository of the default workspace
  zgard ws exec "make test"

  # Run only in the backend project, on repos containing "auth"
  zgard ws exec -p backend -r auth "go build ./..."

  # Preview what would run, without executing
  zgard ws exec --dry-run "make test"

  # Collect code coverage from all repos tagged "core"
  zgard ws exec -t core "go test -coverprofile=coverage.out ./..."`,
}
```

Adding `Example` fields to all 10 subcommands (`init`, `pull`, `purge`, `exec`, `stash`, `checkout`, `search`, `diff`, `stats`, `status`) would make `--help` self-contained as a learning resource.

Additionally, consider using [**glamour**](https://github.com/charmbracelet/glamour) to render the `Long` description as Markdown in the terminal — the project already uses charmbracelet's ecosystem (`charmbracelet/log`), so this would be a natural fit.

---

## 4. Add Examples to Documentation

**Current state:** ✅ `docs/CLI.md` has rich sample output for every subcommand.  
⚠️ Those examples don't appear when you run `zgard ws exec --help` from the terminal.

**Improvement:**  
Examples in `docs/CLI.md` should be mirrored in the Cobra `Example` field (see tip 3 above) so they are available offline without opening a browser. Specifically, add advanced composition examples showing **flag combinations**:

```bash
# Orchestrate a full "safe branch switch" across a workspace:
zgard ws stash -p backend          # stash dirty repos first
zgard ws checkout -p backend dev   # switch all backend repos to "dev"
zgard ws pull -p backend           # pull latest on the new branch

# Weekly workspace health check:
zgard ws diff --all --parallel     # spot dirty repos
zgard ws stats --all               # review commit velocity
```

These patterns are not obvious from individual command docs — they demonstrate the _system_ working together.

---

## 5. Make It Pretty

**Current state:** ✅ Color is applied via `github.com/fatih/color`, respects `NO_COLOR` env and non-TTY output. The reporter outputs `✓ / ✗ / ⏭` symbols. Tables in `diff` and `stats` are column-aligned.

**Improvements:**

**a) Progress indicator for parallel mode**  
When `--parallel` is used, output lines arrive in non-deterministic order and there is no indication of _how many_ repos are remaining. A live progress counter would reduce perceived latency:

```
Cloning 12 repositories in parallel... (7/12 done)
```

[bubbletea](https://github.com/charmbracelet/bubbletea) + [bubbles/progress](https://github.com/charmbracelet/bubbles) would handle this. For a simpler approach, an atomic counter printed to stderr would also help.

**b) Richer `--verbose` output**  
Today `--verbose` prints the rendered shell command. Consider also printing timestamps (total elapsed per repo) so users can identify slow clone sources.

**c) Add a `--no-color` flag**  
`NO_COLOR` is honoured via `fatih/color`, but there is no explicit `--no-color` CLI flag. Some CI systems set neither `NO_COLOR` nor pipe output, so an explicit flag provides a reliable escape hatch:

```go
rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
```

**d) Consider `lipgloss` for table formatting**  
`diff` and `stats` build tables with manual `fmt.Sprintf` padding. [lipgloss](https://github.com/charmbracelet/lipgloss) is already in the charmbracelet family and provides border, padding, and alignment primitives that adapt to the terminal width automatically.

---

## 6. Make It Interactive

**Current state:** ❌ No interactive mode exists.

**Improvement:**  
Many of grazhda's power users manage 20+ repositories. An interactive TUI would let them visually navigate the workspace tree and trigger operations without remembering exact flag syntax.

The highest-ROI addition would be an **interactive `ws status` view** using [bubbletea](https://github.com/charmbracelet/bubbletea):

- Navigate workspaces/projects/repos with arrow keys
- `r` to trigger a rescan on the selected workspace
- `p` to pull the selected item
- Live auto-refresh every N seconds (configurable)
- Filter view by typing (fuzzy match on repo name)

Even a basic `--watch` flag on `zgard ws status` that refreshes the screen every 30 seconds would be a significant UX improvement:

```bash
zgard ws status --watch          # auto-refresh every 30s
zgard ws status --watch --rescan # refresh + force git poll
```

---

## 7. Use asciinema to Record Short Tutorials

**Current state:** ✅ `demo.cast` and `grazhda-demo.gif` exist. The GIF is embedded in the README.

**Improvements:**

**a) Update the demo to show recent features**  
The demo likely predates the targeting system (`-p`, `-r`, `-t`), `ws exec`, `ws diff`, and `ws stats`. A refresh showing a real-world "board of directors" workflow (stash → checkout → pull across 8 repos) would be compelling.

**b) Publish the `.cast` to asciinema.org**  
A hosted asciinema recording lets users pause, copy commands, and replay at their own speed — a GIF cannot. Embed an asciinema player badge in the README alongside the GIF.

**c) Add per-feature short demos**  
Short (15–30 second) casts for individual features (`ws exec`, `ws diff --parallel`) placed in `docs/` would serve as living documentation that stays in sync with the tool.

---

## 8. Carefully Craft a CLI Grammar

**Current state:** ✅ The `zgard ws <verb>` grammar is generally consistent. Destructive operations (`purge`) require explicit targeting.

**Improvements:**

**a) Add `zgard config` subcommand**  
Currently there is no way to inspect the resolved configuration from the CLI. Users must open `~/.grazhda/config.yaml` manually to check values or debug targeting. A `zgard config` command would mirror `git config`:

```bash
zgard config list               # print all resolved fields
zgard config get workspaces     # print the workspaces array as YAML/JSON
zgard config validate           # run validation and report errors
zgard config path               # print the resolved config file path
```

This is especially valuable when troubleshooting why `--name myws` fails (typo in config? wrong GRAZHDA_DIR?).

**b) Introduce `zgard ws list`**  
There is no way to list available workspaces, projects, or repositories without opening the config file. A read-only listing command closes this gap:

```bash
zgard ws list                   # list all workspaces
zgard ws list -n myws           # list projects in "myws"
zgard ws list -n myws -p backend # list repos in "backend"
```

**c) Normalize `--dry-run` as universal**  
`--dry-run` is missing from `pull`, `stash`, and `checkout` (available only on `init`, `exec`, `purge`). It should be a universal flag on all mutating commands for safety and discoverability.

---

## 9. Add Structured Input and Output

**Current state:** ❌ All output is human-readable text. There is no machine-readable output format. `ws stats` and `ws diff` produce aligned text tables that are not parseable by downstream tools.

**Improvement:**  
Add a global `--output` flag (short: `-o`) with three formats:

| Value | Description |
|---|---|
| `text` (default) | Current human-readable, colorized output |
| `json` | JSON array of operation results |
| `tsv` | Tab-separated values (easy for `awk`, spreadsheets) |

```bash
# Human use
zgard ws stats -p backend

# Feed into jq for custom dashboards
zgard ws stats -p backend --output json | jq '.[] | select(.behind > 0)'

# Import into a spreadsheet
zgard ws stats --all --output tsv > report.tsv
```

The `Reporter` and `OpResult` structs are already good candidates for JSON serialization — `OpResult` maps directly to a JSON object with `workspace`, `project`, `repo`, `skipped`, `error`, `message` fields.

For `diff` and `stats`, the table rows are also natural JSON objects. The addition would be entirely in the output layer with no changes to domain logic.

---

## 10. Add Filtering and Aggregation Options

**Current state:** ✅ Rich targeting (`--name`, `--project-name`, `--repo-name`, `--tag`) plus `--glob`/`--regex` on `search`.

**Improvements:**

**a) Add `--since` to `ws stats`**  
The 30-day commit window is hardcoded. Allow users to specify a date range:

```bash
zgard ws stats --since 2024-01-01         # commits since a date
zgard ws stats --since 7d                 # relative: last 7 days
```

**b) Add `--sort` to `ws stats` and `ws diff`**  
Make tables actionable rather than just readable:

```bash
zgard ws stats --sort commits:desc        # most active repos first
zgard ws diff --sort uncommitted:desc     # dirtiest repos first
```

**c) Add `--filter` for output rows**  
Allow post-processing filters on the output without requiring `jq`:

```bash
zgard ws diff --filter ahead>0            # only repos with unpushed commits
zgard ws diff --filter uncommitted>0      # only dirty repos
```

**d) Document `jq` patterns in CLI.md**  
Until structured output is implemented, document canonical `jq` and `awk` patterns for processing current output in `docs/CLI.md`. This gives power users a path today.

---

## 11. Add Templating Options

**Current state:** ✅ `clone_command_template` uses Go's `text/template` for clone command construction. This is the strongest templating story in the project.  
❌ No output templating for `stats`, `diff`, `exec` results.

**Improvement:**  
Add a `--template` flag that accepts a Go template string for customising output. This matches the `kubectl -o go-template='...'` pattern:

```bash
# Custom one-liner per repo
zgard ws stats --template '{{.Repo}}: {{.Commits}} commits by {{.Contributors}} people'

# Generate a Markdown table
zgard ws diff --template '| {{.Repo}} | {{.Uncommitted}} | {{.Ahead}} | {{.Behind}} |'
```

Expose a `TemplateData` struct per command (mirroring the existing `CloneTemplateData` pattern) so the template variables are documented and type-safe. Add examples to `docs/CLI.md` showing the available fields for each command.

---

## 12. Make It Easy to Extend

**Current state:** ❌ No plugin system. There is no mechanism for users or teams to add new `zgard ws` subcommands without modifying the source.

**Improvement:**  
Adopt the **`git`/`gh` plugin pattern**: when `zgard ws <verb>` is invoked with an unknown verb, look for an executable named `zgard-ws-<verb>` on `$PATH` and invoke it, passing all remaining arguments.

```bash
# A user creates ~/bin/zgard-ws-deploy (a shell script):
#!/usr/bin/env bash
# Custom deployment orchestration
zgard ws exec "make deploy" -p "$@"

# Now this works:
zgard ws deploy backend
```

This gives teams a powerful escape hatch for project-specific automation without forking grazhda. The implementation is small (~20 lines in `ws.go`'s `PersistentPreRunE` or as a custom `RunE` fallback on the parent `ws` command).

---

## 13. Make It Easy to Configure

**Current state:** ✅ `GRAZHDA_DIR` env var controls config location. `grazhda config --edit` opens the file in the configured editor.  
⚠️ The `ZgardConfig.Config` field is `map[string]interface{}` — untyped and unparseable.  
❌ No per-project or per-directory config override. No way to query config from the CLI.

**Improvements:**

**a) Type the `zgard.config` YAML section**  
Replace `map[string]interface{}` with a real struct containing useful options:

```go
type ZgardConfig struct {
    DefaultOutput   string `yaml:"default_output"`    // "text"|"json"|"tsv"
    DefaultParallel bool   `yaml:"default_parallel"`  // auto-enable --parallel
    ColorEnabled    *bool  `yaml:"color_enabled"`     // tri-state: nil = auto
}
```

**b) Add `zgard config` inspection commands**  
_(See also tip 8a above.)_

```bash
zgard config path      # $HOME/.grazhda/config.yaml
zgard config validate  # run Validate() and report all errors
zgard config list      # dump resolved config as YAML or JSON
```

**c) Support per-directory config (`.grazhda.yaml`)**  
Allow a `.grazhda.yaml` in the current working directory or any parent to override specific fields (e.g., select a named workspace by default for a particular project tree). This follows the `git` model of local config overriding global config and is especially useful in monorepos where different subdirectories correspond to different workspaces.

**d) Surface `GRAZHDA_DIR` more prominently**  
Add `GRAZHDA_DIR` to `zgard --help` output (currently it is only mentioned in the Quick Start docs). Users debugging config resolution shouldn't need to read documentation.

---

## 14. Provide Autocompletion Integration

**Current state:** ✅ `zgard completion` and `dukh completion` generate static shell completion scripts for bash, zsh, fish, and PowerShell.  
❌ Completion is entirely static — no workspace names, project names, or tags are completed dynamically.

**Improvement:**  
Register `ValidArgsFunction` on the three most important flags so tab-completion reads from the live config:

| Flag | Completes |
|---|---|
| `--name` / `-n` | Workspace names from `config.yaml` |
| `--project-name` / `-p` | Project names (filtered by `--name` if set) |
| `--tag` / `-t` | All unique tag values across the workspace |

Also register positional argument completion where applicable:

```go
// zgard ws checkout <branch> — complete from git branches in targeted repos
cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    if len(args) == 0 {
        return gitBranchesFromTargetedRepos(), cobra.ShellCompDirectiveNoFileComp
    }
    return nil, cobra.ShellCompDirectiveNoFileComp
}
```

Dynamic completion for `checkout` (listing git branches) and `exec` (listing executables/Makefile targets) would make the cross-repo workflow substantially faster for daily use.

---

## Summary Table

| # | Tip | Current State | Priority |
|---|---|---|---|
| 1 | Use a parsing framework | ✅ Cobra | 🟡 Add dynamic completion |
| 2 | Single binary, no deps | ✅ Go static binaries | 🟡 Publish release binaries |
| 3 | Docs in the tool | ⚠️ No `Example:` fields | 🔴 High impact, low effort |
| 4 | Examples in docs | ⚠️ Only in markdown files | 🟡 Mirror to `Example:` fields |
| 5 | Make it pretty | ✅ Colors + symbols | 🟡 Progress indicator, `--no-color` |
| 6 | Make it interactive | ❌ No TUI/watch mode | 🟢 `--watch` flag or bubbletea TUI |
| 7 | asciinema demos | ✅ demo.cast exists | 🟡 Refresh + publish to asciinema.org |
| 8 | Craft a CLI grammar | ✅ Consistent verbs | 🔴 `zgard config`, `zgard ws list` |
| 9 | Structured I/O | ❌ Text only | 🔴 `--output json/tsv` |
| 10 | Filtering & aggregation | ✅ Targeting flags | 🟡 `--since`, `--sort`, `--filter` |
| 11 | Templating options | ⚠️ Clone templates only | 🟢 `--template` for output |
| 12 | Easy to extend | ❌ No plugin system | 🟢 `zgard-ws-*` plugin pattern |
| 13 | Easy to configure | ⚠️ Untyped config section | 🔴 `zgard config` commands, typed struct |
| 14 | Autocompletion | ⚠️ Static only | 🔴 Dynamic completion for targeting flags |

**Legend:** 🔴 High value / low effort — do first · 🟡 Medium effort · 🟢 Large feature / do later
