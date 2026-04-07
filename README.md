# Grazhda

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Phase](https://img.shields.io/badge/Phase-2%20%E2%80%94%20dukh-brightgreen)](https://github.com/vhula/grazhda)

**One command to clone your entire dev environment. Exactly how you left it.**

---

## 🎬 Demo

![Grazhda demo](grazhda-demo.gif)

---

## 🔥 The Problem

You change laptops. You onboard a new teammate. You re-provision a machine after a crash. Then comes the ritual: remember which repos go where, which branches are right for each project, which SSH key format each remote uses. Repeat forty times. Miss one. Break something.

Developer environments are configuration. Configuration should be code.

---

## ⚡ The Solution

`zgard` reads a single YAML file that describes your workspaces — where they live on disk, which repos belong to each project, which branch each repo tracks. Then it does the work.

```
zgard ws init
```

```
Workspace: default
  Project: backend
    ✓ api          — cloned (main)
    ✓ auth         — cloned (dev)
    ✓ gateway      — cloned (main)
  Project: infra
    ✓ terraform    — cloned (main)
    ⏭ k8s-configs  — already exists, skipped

✓ 4 cloned  ⏭ 1 skipped  ✗ 0 failed
```

Every failure is reported with the actual git error — no more hunting for `exit status 128`.

---

## 🚀 Quick Start

### Install

```bash
curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda.sh | bash
```

The installer builds `zgard` and `dukh` from source and places them in `$GRAZHDA_DIR/bin/`.

### Configure

```bash
cp config.template.yaml "$GRAZHDA_DIR/config.yaml"
$EDITOR "$GRAZHDA_DIR/config.yaml"
```

### Run

```bash
zgard ws init            # clone everything in the default workspace
zgard ws pull            # pull latest on all repos
dukh start               # start background workspace health monitor
zgard ws status          # check workspace health
```

> **Prerequisites:** `bash`, `curl`, `git`, `just`, `protoc`, Go `1.26+`

---

## 🗂️ Workspace Concept

A **workspace** is the central organizing unit in Grazhda. It describes a group of related projects that live together on your local machine — their disk locations, their repositories, and the exact branch each repository should track.

```
workspace: default
  └── project: backend
  │     ├── repo: api          (branch: main)
  │     ├── repo: auth         (branch: dev)
  │     └── repo: gateway      (branch: main)
  └── project: infra
        └── repo: terraform    (branch: main)
```

The three Grazhda tools each have a distinct role in the workspace lifecycle:

| Tool | Role |
| :--- | :--- |
| **`grazhda`** | Manages your Grazhda installation — upgrade, configure |
| **`zgard`** | Manages workspace operations on demand — clone, pull, purge, inspect health |
| **`dukh`** | Monitors workspace health continuously in the background — detects branch drift and missing repos |

`dukh` is the *overseer*: it runs silently in the background and keeps a live health snapshot of every workspace. `zgard` is the *operator*: you run it when you want to act — clone repositories, pull changes, or query the health snapshot that `dukh` maintains. Both tools read the same `config.yaml` and share the same workspace model.

---

## 🧰 CLI Tools

### `grazhda` — Installation Manager

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

#### `grazhda upgrade`
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

#### `grazhda config --edit`
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

### `zgard` — Workspace CLI

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

#### `zgard ws init`
Clone all repositories for a workspace. Skips repos that already exist. Continues on failure and reports all errors at the end.

```bash
zgard ws init                            # default workspace
zgard ws init --name myws                # named workspace
zgard ws init --all --parallel-all       # all workspaces, all repos concurrently
zgard ws init --all --parallel           # all workspaces, per-project concurrency
zgard ws init --clone-delay-seconds=5   # sleep 5s after each clone
zgard ws init --dry-run                  # preview without executing
```

| Flag | Description |
|---|---|
| `--parallel` | Clone repos within each project concurrently |
| `--parallel-all` | Clone **all** repos across all projects concurrently |
| `--clone-delay-seconds=N` | Sleep N seconds after each clone command |

#### `zgard ws pull`
Run `git pull --rebase` for every repo in a workspace. Skips repos that haven't been cloned yet.

```bash
zgard ws pull                            # default workspace
zgard ws pull --all --parallel-all       # all workspaces, all repos concurrently
zgard ws pull --all --parallel           # all workspaces, per-project concurrency
zgard ws pull --dry-run                  # preview without executing
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

Common flags for `zgard ws` commands:

| Flag | Commands | Description |
| :--- | :--- | :--- |
| `-n, --name <name>` | init, pull, purge, status | Target a named workspace |
| `--all` | init, pull, purge | Target all workspaces |
| `--dry-run` | init, pull, purge | Print actions without executing |
| `--parallel` | init, pull | Clone/pull repos within each project concurrently |
| `--parallel-all` | init, pull | Clone/pull all repos across all projects concurrently |
| `--clone-delay-seconds=N` | init | Sleep N seconds after each clone command |
| `-v, --verbose` | init, pull, purge | Print the rendered git command before each operation |
| `--no-confirm` | purge | Skip the confirmation prompt |
| `--rescan` | status | Trigger a fresh scan before reporting |

---

### `dukh` — Workspace Health Monitor

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

#### `dukh start`
Start dukh as a detached background process. Self-daemonizes — no `&` needed. Logs go to `$GRAZHDA_DIR/logs/dukh.log`.

```bash
dukh start
# ✓ dukh started (pid 12345)
```

#### `dukh stop`
Stop the running dukh monitor server gracefully via gRPC.

```bash
dukh stop
```

#### `dukh status`
Show **process** health — whether dukh is running, its PID, and uptime. Distinct from `zgard ws status` which shows workspace health.

```bash
dukh status
# ●  dukh: running  (pid 12345, uptime: 2h 34m)
# ○  dukh: not running
```

#### `dukh scan`
Trigger an immediate out-of-cycle workspace rescan. Fire-and-forget — use `zgard ws status --rescan` if you need to wait for fresh results.

```bash
dukh scan
# ✓ rescan initiated
```

---

## ⚙️ Configuration

`zgard` and `dukh` both resolve `config.yaml` from:
1. `$GRAZHDA_DIR/config.yaml` — when `$GRAZHDA_DIR` is set
2. `~/.grazhda/config.yaml` — default fallback

### Example

```yaml
editor: vim      # used by `grazhda config --edit`; fallback: $VISUAL → $EDITOR → vi

dukh:
  host: localhost
  port: 50501
  monitoring:
    period_mins: 5       # how often dukh polls workspace health (default: 5)

workspaces:
  - name: default
    default: true
    path: ~/ws
    clone_command_template: "git clone --branch {{.Branch}} git@github.com:myorg/{{.RepoName}} {{.DestDir}}"
    # structure: tree  # "tree" (default) or "list" — controls how "/" in repo names maps to dirs
    projects:
      - name: backend
        branch: main
        repositories:
          - name: api
          - name: auth
            branch: dev            # overrides project branch
          - name: api
            local_dir_name: api-v2 # cloned into <project>/api-v2
          - name: org/pack/repo    # with structure:list → cloned as <project>/repo

  - name: personal
    path: ~/personal
    clone_command_template: "git clone git@github.com:me/{{.RepoName}} {{.DestDir}}"
    structure: list                # last URL segment used as dest dir
    projects:
      - name: tools
        branch: main
        repositories:
          - name: dotfiles
          - name: scripts
```

### Clone Template Variables

| Variable | Resolves to |
| :--- | :--- |
| `{{.Branch}}` | `repository.branch` if set, otherwise `project.branch` |
| `{{.RepoName}}` | `repository.name` (full value, including any slashes) |
| `{{.DestDir}}` | Full filesystem path to the clone destination (see `structure`) |

### Workspace Structure Modes

The optional `structure` field controls how repository names that contain **`/`** (common with namespaced registries like `org/team/repo`) are mapped to local directories inside each project folder.

| Mode | Behaviour | Example `org/pack/repo` |
| :--- | :--- | :--- |
| `tree` *(default)* | Preserves the full name as nested subdirectories | `<project>/org/pack/repo` |
| `list` | Uses the **last `/`-delimited segment** of the name (`.git` stripped) | `<project>/repo` |

**Conflict handling in `list` mode** — if two repos share the same last segment (e.g. `org/api` and `other/api`), the second clone will be skipped as "already exists". Use `local_dir_name` to resolve such conflicts:

```yaml
repositories:
  - name: org/api
  - name: other/api
    local_dir_name: other-api   # cloned as <project>/other-api
```

### Field Reference

<details>
<summary><strong>Workspace fields</strong></summary>

| Field | Required | Description |
| :--- | :---: | :--- |
| `name` | ✅ | Unique identifier; used with `--name` |
| `default` | — | Marks this workspace as the default target |
| `path` | ✅ | Root directory for the workspace (`~` is expanded) |
| `clone_command_template` | ✅ | Go template string for the clone command |
| `structure` | — | `tree` (default) or `list` — controls dest dir for repo names with `/` |
| `projects` | — | List of project subdirectories |

</details>

<details>
<summary><strong>Project fields</strong></summary>

| Field | Required | Description |
| :--- | :---: | :--- |
| `name` | ✅ | Directory name under the workspace path |
| `branch` | ✅ | Default branch for all repos in this project |
| `repositories` | — | List of repositories to clone |

</details>

<details>
<summary><strong>Repository fields</strong></summary>

| Field | Required | Description |
| :--- | :---: | :--- |
| `name` | ✅ | Repository name; used as `{{.RepoName}}` (may contain `/`) |
| `branch` | — | Overrides the project-level branch |
| `local_dir_name` | — | Explicit clone destination name; overrides both `name` and `structure` |

</details>

---

## 🗺️ Roadmap

| Tool | Role | Status |
| :--- | :--- | :---: |
| **zgard** | Workspace lifecycle CLI | ✅ Phase 1 |
| **dukh** | Background gRPC workspace monitor | ✅ Phase 2 |
| **grazhda** | Installer + management script | ✅ |
| **Molfar** | Orchestration server | 📅 Phase 3 |
| **Molf** | Orchestrator CLI | 📅 Phase 3 |

---

## 📚 Documentation

Detailed design and planning documents live in the [`docs/`](docs/) directory:

| Document | Description |
| :--- | :--- |
| [Product Requirements](docs/prd.md) | Functional requirements and acceptance criteria |
| [Architecture](docs/architecture.md) | System design, gRPC contracts, and data flow |
| [UX Design Specification](docs/ux-design-specification.md) | CLI output formatting, colors, and layout rules |
| [Epics — Workspace](docs/epics.md) | User stories for zgard workspace features |
| [Epics — Dukh](docs/epics-dukh.md) | User stories for dukh server features |

---

## 🛠️ Development

```bash
just generate      # regenerate protobuf Go code from proto/dukh.proto
just build-zgard   # build → bin/zgard
just build-dukh    # build → bin/dukh
just test          # run all tests
just fmt           # format all Go source
just tidy          # tidy all modules
```

### Module Layout

```
grazhda/
├── go.work
├── Justfile
├── proto/                  # protobuf sources
│   └── dukh.proto
├── config.template.yaml
├── internal/               # shared module
│   ├── color/              # terminal colour helpers
│   ├── config/             # load · validate · render templates
│   ├── executor/           # shell command interface + mock
│   ├── reporter/           # per-repo progress + run summary
│   └── workspace/          # init · purge · pull · targeting
├── zgard/                  # CLI module
│   ├── main.go
│   ├── root.go
│   └── ws/                 # ws init · ws purge · ws pull · ws status
└── dukh/                   # gRPC server module
    ├── cmd/                # dukh start · stop · status · scan
    ├── proto/              # generated protobuf (do not edit)
    └── server/             # gRPC server · monitor · logging
```

---

## 📄 License

GNU GPL v3 — see [`LICENSE`](LICENSE).
