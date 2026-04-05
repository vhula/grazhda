# Grazhda

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

Grazhda is a local automation toolkit for workspace lifecycle and repeatable developer setup.

## Table of Contents

- [Design](#design)
- [Components](#components)
- [Technology Stack](#technology-stack)
- [Current Status](#current-status)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Development](#development)
- [License](#license)

## Design

Grazhda uses a split design where `zgard` is the primary user-facing CLI for direct workspace operations and `dukh` is a future gRPC server for background process control.

**Phase 1 focuses exclusively on `zgard`.**

### How it works

`zgard` reads `$GRAZHDA_DIR/config.yaml` (or `~/.grazhda/config.yaml` by default), validates the configuration up-front, then creates workspace directory structures and runs `git` commands on your behalf.

The full pipeline for any command is:

1. Load and validate `config.yaml` – all errors are reported before any filesystem change
2. Resolve target workspace(s) from flags (`--name`, `--all`, or default)
3. Execute the operation (`init` / `purge` / `pull`) with live per-repo progress output
4. Print a run summary with counts and exit non-zero if any operation failed

## Components

| Component | Module | Role |
| :--- | :--- | :--- |
| `zgard` | `github.com/vhula/grazhda/zgard` | Entry point — root command + `ws` subcommands |
| `zgard/ws` | `github.com/vhula/grazhda/zgard` | Cobra command definitions for `ws init`, `ws purge`, `ws pull` |
| `internal/config` | `github.com/vhula/grazhda/internal` | Config loading, validation, and clone template rendering |
| `internal/workspace` | `github.com/vhula/grazhda/internal` | `Init`, `Purge`, `Pull` orchestration + workspace targeting (`Resolve`) |
| `internal/executor` | `github.com/vhula/grazhda/internal` | `Executor` interface; `OsExecutor` (real) and `MockExecutor` (tests) |
| `internal/reporter` | `github.com/vhula/grazhda/internal` | Per-operation progress output and run summary |
| `dukh` | `github.com/vhula/grazhda/dukh` | gRPC server — placeholder (Phase 2) |

## Technology Stack

| Area | Technology |
| :--- | :--- |
| CLI | Go + [Cobra](https://github.com/spf13/cobra) |
| Terminal Output | [fatih/color](https://github.com/fatih/color) |
| Config | YAML (`gopkg.in/yaml.v3`) |
| Build | `just` (`Justfile`) |
| Module layout | Go workspace (`go.work`) — two modules: `internal`, `zgard` |

## Current Status

| Tool | Language | Role | Status |
| :--- | :------- | :--- | :----- |
| **zgard** | Go | Workspace lifecycle CLI | ✅ Implemented |
| **Grazhda installer** | Bash | Source-build installer | 🚧 In Progress |
| **dukh** | Go | Worker (gRPC server) | 📅 Planned (Phase 2) |
| **Molfar** | Java | Brain orchestrator | 📅 Planned (Phase 3) |
| **Molf** | Java | Orchestrator CLI | 📅 Planned (Phase 3) |

## Quick Start

### Prerequisites

- `bash`, `curl`, `git`
- Go `1.26+` (required to build from source)

### Install

```bash
curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda.sh | bash
```

The installer clones the repo, builds `zgard`, and places the binary in `$GRAZHDA_DIR/bin/`.

### Configure

Copy the template and edit it:

```bash
cp config.template.yaml "$GRAZHDA_DIR/config.yaml"
$EDITOR "$GRAZHDA_DIR/config.yaml"
```

### Workspace Commands

```bash
# Initialize the default workspace (clone all repos)
zgard ws init

# Initialize a specific workspace
zgard ws init --name myws

# Initialize all workspaces in parallel
zgard ws init --all --parallel

# Preview what init would do without making changes
zgard ws init --dry-run

# Pull latest changes for all repos in the default workspace
zgard ws pull

# Pull in parallel
zgard ws pull --parallel

# Preview pull
zgard ws pull --dry-run

# Remove a workspace (prompts for confirmation)
zgard ws purge --name myws

# Remove all workspaces without prompting (for CI)
zgard ws purge --all --no-confirm

# Preview purge
zgard ws purge --name myws --dry-run
```

### Common Flags

| Flag | Commands | Description |
| :--- | :--- | :--- |
| `--name <name>` | init, pull, purge | Target a named workspace |
| `--all` | init, pull, purge | Target all workspaces |
| `--dry-run` | init, pull, purge | Print what would happen without executing |
| `--parallel` | init, pull | Run repository operations concurrently |
| `--verbose` / `-v` | init, pull, purge | Print rendered commands before execution |
| `--no-confirm` | purge | Skip the confirmation prompt (for CI/scripts) |

> **Note:** `ws purge` always requires `--name <name>` or `--all` – it will not run against the default workspace without an explicit flag.

## Configuration

### Location

`zgard` resolves the config path in this order:

1. `$GRAZHDA_DIR/config.yaml` (when `$GRAZHDA_DIR` is set)
2. `~/.grazhda/config.yaml` (default fallback)

### Full Example

```yaml
workspaces:
  - name: default
    default: true
    path: /home/jake/ws
    clone_command_template: "git clone --branch {{.Branch}} https://github.com/myorg/{{.RepoName}} {{.DestDir}}"
    projects:
      - name: backend
        branch: main
        repositories:
          - name: api
          - name: auth
            branch: dev            # overrides project branch
          - name: api
            local_dir_name: api-v2 # cloned into <project_path>/api-v2

  - name: personal
    path: /home/jake/personal
    clone_command_template: "git clone git@github.com:jake/{{.RepoName}} {{.DestDir}}"
    projects:
      - name: tools
        branch: main
        repositories:
          - name: dotfiles
          - name: scripts
```

### Workspace Fields

| Field | Required | Description |
| :--- | :--- | :--- |
| `name` | ✅ | Unique workspace identifier; used with `--name` |
| `default` | – | Mark one workspace as the default (or name it `"default"`) |
| `path` | ✅ | Absolute filesystem path for the workspace root |
| `clone_command_template` | ✅ | Go template for clone commands (see below) |
| `projects` | – | List of project subdirectories |

### Project Fields

| Field | Required | Description |
| :--- | :--- | :--- |
| `name` | ✅ | Project directory name under the workspace path |
| `branch` | ✅ | Default branch for all repos in the project |
| `repositories` | – | List of repositories to clone into this project |

### Repository Fields

| Field | Required | Description |
| :--- | :--- | :--- |
| `name` | ✅ | Repository name; used as `{{.RepoName}}` in the template |
| `branch` | – | Overrides the project-level branch for this repo |
| `local_dir_name` | – | Local directory name; overrides `name` as the clone destination |

### Clone Template Variables

| Variable | Value |
| :--- | :--- |
| `{{.Branch}}` | `repository.branch` if set, otherwise `project.branch` |
| `{{.RepoName}}` | `repository.name` |
| `{{.DestDir}}` | `<project_path>/<local_dir_name>` if set, otherwise `<project_path>/<name>` |

### Output Format

```
Workspace: default
  Project: backend
    ✓ api          — cloned (main)
    ✗ auth         — fatal: Remote branch dev not found in upstream origin
    ⏭ api-v2       — already exists, skipped

✓ 1 cloned  ⏭ 1 skipped  ✗ 1 failed
```

Progress lines go to **stdout**; failure details go to **stderr**.

## Development

### Build

```bash
just build-zgard   # produces bin/zgard
```

### Test

```bash
just test          # go test ./... across all modules
just test          # with race detector: cd internal && go test -race ./...
```

### Format & Tidy

```bash
just fmt           # gofmt across all modules
just tidy          # go mod tidy per module
```

### Module Layout

```
grazhda/
├── go.work                  # workspace: internal, zgard, dukh
├── Justfile                 # build/test/fmt/tidy targets
├── config.template.yaml     # workspace config template
├── internal/                # module: github.com/vhula/grazhda/internal
│   ├── color/               # Terminal color helpers (fatih/color wrapper)
│   ├── config/              # Load, Validate, DefaultWorkspace, RenderCloneCmd
│   ├── executor/            # Executor interface, OsExecutor, MockExecutor
│   ├── reporter/            # Reporter — Record, Summary, ExitCode
│   ├── workspace/           # Init, Purge, Pull, Resolve + RunOptions
│   └── testdata/            # YAML fixtures for unit tests
├── zgard/                   # module: github.com/vhula/grazhda/zgard
│   ├── main.go              # entry point → Execute()
│   ├── root.go              # Cobra root command + Execute()
│   └── ws/                  # ws parent + init/purge/pull subcommands
└── dukh/                    # module: github.com/vhula/grazhda/dukh (placeholder)
```

## License

This project is licensed under GNU GPL v3. See `LICENSE`.

