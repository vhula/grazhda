# 🏔️ Grazhda (Grazhda)

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

Grazhda is a local automation ecosystem focused on workspace lifecycle, filesystem operations, and repeatable developer setup.

## Table of Contents

- [Zgard + Dukh Design](#zgard--dukh-design)
  - [Responsibilities](#responsibilities)
  - [Flow](#flow)
- [Toolkit Status](#toolkit-status)
- [Quick Start](#quick-start)
  - [Prerequisites](#prerequisites)
  - [Install](#install)
  - [Start Dukh](#start-dukh)
  - [Run Zgard Commands](#run-zgard-commands)
- [Configuration](#configuration)
  - [Location](#location)
  - [Sections](#sections)
  - [Workspaces](#workspaces)
  - [Clone Command Template](#clone-command-template)
- [License](#license)

## Zgard + Dukh Design

The current foundation of Grazhda is built around two Go services:

- **`dukh` (server):** a gRPC worker that manages workspaces from `config.yaml`.
- **`zgard` (client):** a CLI that reads config and sends commands to `dukh`.

### Responsibilities

- **`dukh`**
  - Initializes all configured workspaces.
  - Creates workspace directories and project subdirectories.
  - Builds and executes clone commands from each workspace `clone_command_template`.
  - Purges all configured workspaces.
  - Exposes gRPC operations for init, purge, and get.

- **`zgard`**
  - Reads `dukh.host` and `dukh.port` from config.
  - Calls `dukh` over gRPC.
  - Provides user-facing commands such as:
	- `zgard ws init`
	- `zgard ws purge`
    - `zgard dukh status`
    - `zgard dukh stop`

### Flow

1. `zgard` loads `${GRAZHDA_DIR}/config.yaml`.
2. `zgard` connects to `dukh` via gRPC.
3. `dukh` executes filesystem operations using the configured workspace layout.

## Toolkit Status

| Tool | Language | Role | Status |
| :--- | :------- | :--- | :----- |
| **Grazhda** | Bash | Installer | 🚧 In Development |
| **Dukh** | Go | Worker (gRPC server) | 🚧 In Development |
| **Zgard** | Go | Command CLI | 🚧 In Development |
| **Molfar** | Java | Brain | 📅 Planned |
| **Molf** | Java | Interface CLI | 📅 Planned |

## Quick Start

### Prerequisites

- Bash
- Go 1.26+
- Administrative privileges (if required by your environment)

### Install

```bash
curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda.sh | bash
```

### Start Dukh

```bash
dukh start
```

### Run Zgard Commands

```bash
zgard ws init
zgard ws purge
zgard dukh status
zgard dukh stop
```

## Configuration

### Location

Grazhda expects a YAML config file at:

```
${GRAZHDA_DIR}/config.yaml
```

`GRAZHDA_DIR` must be set in your environment before running `dukh` or `zgard`. The installer sets this automatically.

### Sections

#### `dukh`

Connection settings for the Dukh gRPC server. Used by both `dukh` (to bind) and `zgard` (to connect).

```yaml
dukh:
  host: localhost
  port: 50501
```

| Field | Description |
| :---- | :---------- |
| `host` | Hostname the Dukh server listens on |
| `port` | Port the Dukh server listens on |

#### `general`

Paths used by the installer and tooling.

```yaml
general:
  install_dir: ${GRAZHDA_DIR}
  sources_dir: ${GRAZHDA_DIR}/sources
  bin_dir: ${GRAZHDA_DIR}/bin
```

| Field | Description |
| :---- | :---------- |
| `install_dir` | Root installation directory |
| `sources_dir` | Where source repositories are kept |
| `bin_dir` | Where compiled binaries are placed |

### Workspaces

Workspaces are the top-level unit of organization. Each workspace has its own directory, log file, and set of projects.

```yaml
workspaces:
  - name: default
    default: true
    path: /home/jake/ws
    clone_command_template: git clone --branch {{.Branch}} https://github.com/grazhda/{{.RepoName}} {{.DestDir}}
    projects:
      - name: ws-project1
        subprojects:
          - branch: main
            repositories:
              - name: test-repo-1
              - name: test-repo-2
              - name: test-repo-1
                local_dir_name: test-repo-1-copy
          - branch: dev
            repositories:
              - name: test-repo-3
      - name: ws-project2
        branch: main
        repositories:
          - name: test-repo-1
```

| Field | Description |
| :---- | :---------- |
| `name` | Unique workspace identifier |
| `default` | If `true`, this workspace is the default one |
| `path` | Absolute path where the workspace directory will be created |
| `clone_command_template` | Go template used to build the git clone command for repositories |
| `projects` | List of projects inside this workspace |

Each **project** maps to a subdirectory inside the workspace:

| Field | Description |
| :---- | :---------- |
| `name` | Project directory name |
| `branch` | Branch to clone (used when no subprojects are defined) |
| `repositories` | List of repositories to clone directly under this project |
| `subprojects` | List of subproject groups, each with their own branch and repositories |

Each **repository** entry:

| Field | Description |
| :---- | :---------- |
| `name` | Repository name (used in the clone template as `{{.RepoName}}`) |
| `local_dir_name` | Optional. Overrides the local directory name (`{{.DestDir}}`). Defaults to `name` |

### Clone Command Template

The `clone_command_template` is a Go template evaluated per repository. The following variables are available:

| Variable | Description |
| :------- | :---------- |
| `{{.RepoName}}` | The repository name from config |
| `{{.Branch}}` | The branch from the project or subproject |
| `{{.DestDir}}` | The local target directory (`local_dir_name` if set, otherwise `RepoName`) |

Example:

```yaml
clone_command_template: git clone --branch {{.Branch}} https://github.com/grazhda/{{.RepoName}} {{.DestDir}}
```

This produces commands such as:

```bash
git clone --branch main https://github.com/grazhda/test-repo-1 test-repo-1
git clone --branch main https://github.com/grazhda/test-repo-1 test-repo-1-copy  # when local_dir_name is set
```

## License

This project is licensed under the GNU General Public License v3.0. See `LICENSE`.
