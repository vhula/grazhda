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

Grazhda currently uses a split design:

- `zgard` is a local CLI that performs workspace operations directly.
- Workspace logic lives in `internal/ws` and is driven by `config.yaml`.
- `dukh` is a separate gRPC server for Dukh process control endpoints only.

### Workspace Operations

`zgard ws init` and `zgard ws purge` are local workspace operations.

`zgard` loads `${GRAZHDA_DIR}/config.yaml`, creates/purges workspace directories, creates `dukh.log` files in workspace roots, creates project directories, and runs repository clone commands based on `clone_command_template`.

## Components

| Component | Role |
| :--- | :--- |
| `zgard` | User CLI for local workspace operations (`ws init`, `ws purge`) |
| `internal/ws` | Workspace domain logic used by `zgard` |
| `internal/config` | Config loading from `${GRAZHDA_DIR}/config.yaml` |
| `dukh` | gRPC server process with Dukh service methods |

## Technology Stack

| Area | Technology |
| :--- | :--- |
| CLI | Go + Cobra |
| Logging | `github.com/charmbracelet/log` |
| RPC | gRPC + Protocol Buffers |
| Config | YAML (`gopkg.in/yaml.v3`) |
| Build orchestration | `just` (`Justfile`) |

## Current Status

| Tool | Language | Role | Status |
| :--- | :------- | :--- | :----- |
| **Grazhda** | Bash | Installer | 🚧 In Development |
| **Dukh** | Go | Worker (gRPC server) | 🚧 In Development |
| **Zgard** | Go | Command CLI | 🚧 In Development |
| **Molfar** | Java | Brain | 📅 Planned |
| **Molf** | Java | Interface CLI | 📅 Planned |

## Quick Start

### Prerequisites

- `bash`
- `curl`
- `git`
- Go `1.26+` (for local development/build)

### Install

```bash
curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda.sh | bash
```

### Initialize and Purge Workspaces

```bash
zgard ws init
zgard ws purge
```

### Dukh Commands

```bash
dukh start
dukh stop
dukh status
```

## Configuration

### Location

Grazhda reads config from:

```text
${GRAZHDA_DIR}/config.yaml
```

Use `config.template.yaml` as the base template.

### Top-Level Sections

```yaml
dukh:
  host: localhost
  port: 50501

general:
  install_dir: ${GRAZHDA_DIR}
  sources_dir: ${GRAZHDA_DIR}/sources
  bin_dir: ${GRAZHDA_DIR}/bin

workspaces:
  - name: default
    default: true
    path: /home/jake/ws
    clone_command_template: git clone --branch {{.Branch}} https://github.com/grazhda/{{.RepoName}} {{.DestDir}}
```

- `dukh`: network settings for the Dukh server.
- `general`: installation and binary/source paths.
- `workspaces`: workspace layout and clone definitions used by `zgard`.

### Workspace Model

Each workspace defines:

- `name`: workspace identifier.
- `path`: absolute path where the workspace directory is created.
- `clone_command_template`: Go template used to build repository clone commands.
- `projects`: project directories and repositories/subprojects.

Repository options:

- `name`: repository name (`{{.RepoName}}`).
- `local_dir_name` (optional): local destination directory (`{{.DestDir}}`).

Template variables supported by clone command rendering:

- `{{.Branch}}`
- `{{.RepoName}}`
- `{{.DestDir}}`

## Development

Common tasks:

```bash
just generate
just build
just test
just fmt
just tidy
```

`just generate` generates protobuf code from `proto/dukh.proto`.

## License

This project is licensed under GNU GPL v3. See `LICENSE`.
