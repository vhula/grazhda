# CLI Reference

> [← Back to README](../README.md)

Grazhda has three CLI tools: `grazhda` (installer), `zgard` (workspace operator), and `dukh` (health monitor).

---

## `grazhda` — Installation Manager

```
Grazhda Management Script

Usage: grazhda <command> [options]

Commands:
  upgrade          Pull latest sources and rebuild all binaries
  config --edit    Open config.yaml in the configured editor
  help             Show this help message

Environment:
  GRAZHDA_DIR      Installation directory (default: $HOME/.grazhda)
```

### `grazhda upgrade`
Pull the latest sources and rebuild all binaries in one command.

```bash
grazhda upgrade
```

What it does:
1. Stops `dukh` gracefully if it is currently running
2. `git pull` in `$GRAZHDA_DIR/sources`
3. `just build` (regenerates proto code, recompiles `zgard` and `dukh`)
4. Copies all binaries to `$GRAZHDA_DIR/bin/`

> **Note:** `dukh` is not restarted automatically after the upgrade. Run `dukh start` (or `zgard ws status`) to bring it back up.

### `grazhda config --edit`
Open `$GRAZHDA_DIR/config.yaml` in your preferred editor.

```bash
grazhda config --edit
```

Editor resolution order:
1. `editor:` field in `config.yaml`
2. `$VISUAL` environment variable
3. `$EDITOR` environment variable
4. `vi` (fallback)

---

## `zgard` — Workspace CLI

```
zgard manages local workspace lifecycle — init, purge, and pull repositories.

Usage:
  zgard [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  ws          Workspace operations

Flags:
  -h, --help   help for zgard

Use "zgard [command] --help" for more information about a command.
```

### Targeting Logic

Every `zgard ws` subcommand shares the same targeting flags, inherited from the `ws` parent:

| Flag | Short | Description |
|---|---|---|
| `--name <name>` | `-n` | Target a single named workspace |
| `--all` | | Operate on all configured workspaces |
| `--project-name <name>` | `-p` | Filter to a specific project within the targeted workspace |
| `--repo-name <name>` | `-r` | Substring filter on repository names (requires `--project-name`) |
| `--tag <tag>` | `-t` | Filter by tag (OR logic; repeat for multiple tags) |

**Flag constraints:**
- `--name` and `--all` are mutually exclusive.
- `--repo-name` (`-r`) requires `--project-name` (`-p`).
- `--all` cannot be combined with `--project-name` or `--repo-name`.
- Violating any constraint prints a **red error** and exits immediately.

**Default workspace info:** If you omit all targeting flags, `zgard` falls back to the `default` workspace and prints a blue info message to stderr so you always know what is being targeted:

```
Info: Targeting default workspace: /home/alice/workspaces/default
```

Suppress the message by being explicit: `zgard ws init --name default`.

**Safety contract for `ws purge`:** Purge never falls back to a default. If no targeting flag is provided it exits immediately with a red error, requiring `--name <name>` or `--all`.

**`--repo-name` substring matching:** The filter value is matched as a case-sensitive substring against the full repository config name, regardless of workspace structure. For example, `--repo-name cool` matches `ORG/PACK/my-cool-backend-lol`. Matching more than one repository is valid — a **yellow warning** is printed before the operation begins:

```
Warning: --repo-name "cool" matches 3 repositories
```

If the filter matches nothing, `zgard` exits with a **red error**.

**`--tag` filtering (OR logic):** Repositories and projects can declare a `tags` list in config. The `--tag` flag filters to repos whose effective tag set (project tags merged with repo tags) contains the specified tag. Supplying multiple `--tag` flags uses **OR** logic — a repo matching any one of the provided tags is included.

```bash
# Run pull only on repositories tagged "backend" or "api"
zgard ws pull -t backend -t api
```

Project-level tags are **inherited** by all repositories in that project (merged with any repo-specific tags). If `--tag` matches no repositories, `zgard` exits with a **red error**.

---

### Subcommands

#### `zgard ws init`
Clone all repositories for a workspace. Skips repos that already exist. Continues on failure and reports all errors at the end.

```bash
zgard ws init                                       # default workspace (shows info)
zgard ws init -n myws                               # named workspace, no info message
zgard ws init --all --parallel                      # all workspaces, concurrently
zgard ws init -p backend                            # only clone backend project
zgard ws init -p backend -r api                     # repos whose name contains "api"
zgard ws init --clone-delay-seconds=5               # sleep 5s after each clone
zgard ws init --dry-run                             # preview without executing
```

#### `zgard ws pull`
Run `git pull --rebase` for every repo in a workspace. Skips repos that haven't been cloned yet.

```bash
zgard ws pull                                       # default workspace (shows info)
zgard ws pull --all --parallel                      # all workspaces, concurrently
zgard ws pull -p backend                            # only pull backend project
zgard ws pull -p backend -r api                     # repos whose name contains "api"
zgard ws pull --dry-run                             # preview without executing
```

#### `zgard ws purge`
Remove a workspace directory tree. Always asks for confirmation. Always requires an explicit target.

```bash
zgard ws purge --name myws            # remove one workspace (prompts)
zgard ws purge --all --no-confirm     # remove all, no prompt (for CI)
zgard ws purge --name myws --dry-run  # preview what would be removed
```

#### `zgard ws status`
Show workspace health as tracked by `dukh`. If `dukh` is not running, it is
automatically started before querying status.

```bash
zgard ws status              # all workspaces (cached); starts dukh if needed
zgard ws status --name myws  # one workspace
zgard ws status --rescan     # trigger a fresh scan, wait, then report
```

Use `--rescan` when you want up-to-the-moment results instead of the last cached snapshot.

```
⟳ rescanning workspaces…

Dukh  running  •  uptime: 2h 34m

Workspace: default
  Project: backend
    ✓ api             main → main
    ✗ auth            main → feat/login  (branch mismatch)
    ✗ gateway         (missing)
  Project: infra
    ✓ terraform       main → main

✓ 2 aligned  ⚠ 1 drifted  ✗ 1 missing
```

#### `zgard ws exec`
Run any shell command in every repository of a workspace. Output from each repo is printed below its status line.

```bash
zgard ws exec "make test"                                  # default workspace, all repos
zgard ws exec --name myws "make lint"                      # named workspace
zgard ws exec --all "echo hi"                              # every workspace
zgard ws exec -p backend "make test"                       # one project only
zgard ws exec -p backend -r api "go build ./..."           # repos whose name contains "api"
zgard ws exec -p backend -r service "go build ./..."       # all repos containing "service"
zgard ws exec --parallel "make test"                       # run concurrently
zgard ws exec --dry-run "make test"                        # preview without executing
```

Sample output:

```
Workspace: default
  Project: backend
    ✓ api          — done
      running tests...
      ok  github.com/acme/api
    ✗ auth         — exit 1: make: No rule to make target 'test'
      make: No rule to make target 'test'. Stop.
    ⏭ gateway      — not present, skipped

✓ 1 executed  ⏭ 1 skipped  ✗ 1 failed
```

#### `zgard ws stash`
Run `git stash push` in every repository. Useful before coordinated branch switches.

```bash
zgard ws stash                                             # default workspace
zgard ws stash --all --parallel                        # all workspaces, concurrently
zgard ws stash -p backend                                  # one project only
zgard ws stash -p backend -r service                       # repos containing "service"
zgard ws stash --dry-run                                   # preview without executing
```

Sample output:

```
Workspace: default
  Project: backend
    ✓ api          — stashed
    ✓ auth         — stashed
    ⏭ gateway      — not present, skipped

✓ 2 stashed  ⏭ 1 skipped  ✗ 0 failed
```

#### `zgard ws checkout`
Run `git checkout <branch>` in every repository. Combine with `ws stash` to safely switch the whole workspace to a feature branch.

```bash
zgard ws checkout main                                     # default workspace → main
zgard ws checkout --name myws feature-x                    # named workspace
zgard ws checkout -p backend feature-x                     # one project only
zgard ws checkout -p backend -r api feature-x             # repos containing "api"
zgard ws checkout --all --parallel main                # all workspaces, concurrently
zgard ws checkout --dry-run feature-x                      # preview without executing
```

Sample output:

```
Workspace: default
  Project: backend
    ✓ api          — checked out feature-x
    ✗ auth         — pathspec 'feature-x' did not match any file(s) known to git
    ⏭ gateway      — not present, skipped

✓ 1 checked out  ⏭ 1 skipped  ✗ 1 failed
```

#### `zgard ws search`

Search file contents (default), filenames (`--glob`), or regular expressions (`--regex`) across all resolved repositories. Binary files and `.git` directories are automatically skipped.

```sh
zgard ws search "TODO"                                     # content grep, default workspace
zgard ws search --glob "*.go"                              # filename glob
zgard ws search --regex "^func\s"                          # Go regex on file contents
zgard ws search -p backend "TODO"                          # one project only
zgard ws search -p backend -r api "TODO"                   # repos containing "api"
zgard ws search --parallel "TODO"                          # search concurrently
```

Sample output:

```
[backend/api] src/handler.go:42: // TODO: fix this
[backend/api] src/handler.go:87: // TODO: add tests
[backend/auth] pkg/auth.go:12:  // TODO: rotate key

3 match(es) across 2 repo(s)
```

#### `zgard ws diff`

Show per-repo Git state in aligned, project-grouped tables. Rows are colour-coded: red for uncommitted changes, yellow for ahead/behind or not cloned, green for fully clean.

```sh
zgard ws diff                                              # default workspace
zgard ws diff --name myws                                  # named workspace
zgard ws diff -p backend                                   # one project only
zgard ws diff --all --parallel                         # all workspaces, concurrently
```

Sample output:

```
Workspace: myws
  Project: backend

    REPO              UNCOMMITTED  AHEAD  BEHIND
    ──────────────────────────────────────────────
    api               3            2      0
    auth-service      0            0      1
    gateway           (not cloned)  --     --

✓ 0 clean  ✗ 2 dirty  ⏭ 1 not cloned
```

#### `zgard ws stats`

Show aggregated repository metadata in aligned, project-grouped tables: last commit date, 30-day commit count, and unique contributor count.

```sh
zgard ws stats                                             # default workspace
zgard ws stats --name myws                                 # named workspace
zgard ws stats -p backend                                  # one project only
zgard ws stats --all --parallel                        # all workspaces, concurrently
```

Sample output:

```
Workspace: myws
  Project: backend

    REPO              LAST COMMIT        30D COMMITS  CONTRIBUTORS
    ──────────────────────────────────────────────────────────────
    api               2024-06-15 10:30   47           8
    auth-service      2024-06-14 09:00   12           3
    gateway           (not cloned)       -            -
```

---

#### `zgard ws list`

Display a tree-formatted view of all repositories in a workspace with real-time clone status.

```bash
zgard ws list                              # list default workspace
zgard ws list -n myws                      # list a named workspace
zgard ws list --all                        # list all workspaces
zgard ws list -p backend                   # list only backend project
zgard ws list -t api                       # list repos tagged "api"
```

Sample output:

```
Workspace: myws  (~/.grazhda/ws/myws)
  Project: backend
    ✓  api                     ~/workspaces/myws/api
    ✓  auth-service             ~/workspaces/myws/auth-service
    ✗  gateway                  ~/workspaces/myws/gateway

  Project: frontend
    ✓  web-app                  ~/workspaces/myws/web-app
    ✗  design-system            ~/workspaces/myws/design-system

Summary: 3 cloned, 2 not cloned (5 total)
```

✓ = directory exists on disk; ✗ = not yet cloned.

---

### `zgard config` — Configuration Inspection

Inspect, validate, and query the loaded configuration without opening the file.

#### `zgard config path`

Print the resolved path of the active configuration file.

```bash
zgard config path
# Output: /home/alice/.grazhda/config.yaml
```

#### `zgard config validate`

Load the configuration and report any validation errors. Exits 0 on success.

```bash
zgard config validate
# Output (success): ✓ Configuration is valid.
# Output (error):   ✗ Configuration error: ...
```

#### `zgard config list`

List all workspaces and their projects from the configuration file (no filesystem access).

```bash
zgard config list
```

Sample output:

```
Workspaces in /home/alice/.grazhda/config.yaml:

  myws
    backend  (3 repos)
    frontend (2 repos)

  devws
    services (5 repos)
```

#### `zgard config get <key>`

Get a specific configuration value using a dotted-path key (based on YAML field names).

```bash
zgard config get dukh.port         # e.g. "50501"
zgard config get install_dir       # e.g. "/home/alice/.grazhda"
zgard config get workspaces.0.name # first workspace name
```

---

### Common Flags

| Flag | Commands | Description |
| :--- | :--- | :--- |
| `-n, --name <name>` | all | Target a named workspace (persistent, inherited) |
| `--all` | all | Operate on all workspaces (persistent, inherited) |
| `-p, --project-name <name>` | all | Filter to a specific project (persistent, inherited) |
| `-r, --repo-name <name>` | all | Substring filter on repo names — requires `-p`; may match multiple repos (persistent, inherited) |
| `-t, --tag <tag>` | all | Filter by tag (OR logic; repeat for multiple: `-t backend -t api`) (persistent, inherited) |
| `--dry-run` | init, pull, exec, stash, checkout, purge | Print actions without executing; shows `[DRY RUN]` banner |
| `--parallel` | init, pull, exec, stash, checkout, search, diff, stats | Run all repositories concurrently |
| `--clone-delay-seconds=N` | init | Sleep N seconds after each clone command |
| `-v, --verbose` | init, pull, exec, stash, checkout, diff, stats, purge | Print the rendered command before each operation |
| `--no-confirm` | init, purge | Skip the confirmation prompt |
| `--rescan` | status | Trigger a fresh scan before reporting |
| `--glob` | search | Match filenames instead of file contents |
| `--regex` | search | Treat pattern as a Go regular expression |

---

## `dukh` — Workspace Health Monitor

`dukh` is a long-running background daemon. It continuously polls every workspace defined in `config.yaml`, checks branch alignment and repository existence, and exposes the live health snapshot over gRPC so `zgard ws status` can query it instantly.

```
Dukh — Grazhda workspace health monitor

Usage:
  dukh [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  scan        Trigger an immediate workspace rescan in dukh
  start       Start the dukh workspace monitor in the background
  status      Show dukh process status (running, PID, uptime)
  stop        Stop the running dukh workspace monitor

Flags:
  -h, --help   help for dukh

Use "dukh [command] --help" for more information about a command.
```

### `dukh start`
Start dukh as a detached background process. Self-daemonizes — no `&` needed. Logs go to `$GRAZHDA_DIR/logs/dukh.log`.

```bash
dukh start
# ✓ dukh started (pid 12345)
```

### `dukh stop`
Stop the running dukh monitor server gracefully via gRPC.

```bash
dukh stop
```

### `dukh status`
Show **process** health — whether dukh is running, its PID, and uptime. Distinct from `zgard ws status` which shows workspace health.

```bash
dukh status
# ●  dukh: running  (pid 12345, uptime: 2h 34m)
# ○  dukh: not running
```

### `dukh scan`
Trigger an immediate out-of-cycle workspace rescan. Fire-and-forget — use `zgard ws status --rescan` if you need to wait for fresh results.

```bash
dukh scan
# ✓ rescan initiated
```
