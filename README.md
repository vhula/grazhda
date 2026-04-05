# Grazhda

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Phase](https://img.shields.io/badge/Phase-1%20%E2%80%94%20zgard-brightgreen)](https://github.com/vhula/grazhda)

**One command to clone your entire dev environment. Exactly how you left it.**

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

The installer builds `zgard` from source and places it in `$GRAZHDA_DIR/bin/`.

### Configure

```bash
cp config.template.yaml "$GRAZHDA_DIR/config.yaml"
$EDITOR "$GRAZHDA_DIR/config.yaml"
```

### Run

```bash
zgard ws init            # clone everything in the default workspace
zgard ws pull            # pull latest on all repos
zgard ws purge --name old-ws  # remove a workspace
```

> **Prerequisites:** `bash`, `curl`, `git`, Go `1.26+`

---

## 🗂️ Commands

### `zgard ws init`
Clone all repositories for a workspace. Skips repos that already exist. Continues on failure and reports all errors at the end.

```bash
zgard ws init                    # default workspace
zgard ws init --name myws        # named workspace
zgard ws init --all --parallel   # all workspaces, concurrently
zgard ws init --dry-run          # preview without executing
```

### `zgard ws pull`
Run `git pull --rebase` for every repo in a workspace. Skips repos that haven't been cloned yet.

```bash
zgard ws pull                    # default workspace
zgard ws pull --all --parallel   # all workspaces, concurrently
zgard ws pull --dry-run          # preview without executing
```

### `zgard ws purge`
Remove a workspace directory tree. Always asks for confirmation. Always requires an explicit target.

```bash
zgard ws purge --name myws            # remove one workspace (prompts)
zgard ws purge --all --no-confirm     # remove all, no prompt (for CI)
zgard ws purge --name myws --dry-run  # preview what would be removed
```

### Common Flags

| Flag | Commands | Description |
| :--- | :--- | :--- |
| `-n, --name <name>` | init, pull, purge | Target a named workspace |
| `--all` | init, pull, purge | Target all workspaces |
| `--dry-run` | init, pull, purge | Print actions without executing |
| `--parallel` | init, pull | Run repo operations concurrently |
| `-v, --verbose` | init, pull, purge | Print the rendered git command before each operation |
| `--no-confirm` | purge | Skip the confirmation prompt |

---

## ⚙️ Configuration

`zgard` resolves `config.yaml` from:
1. `$GRAZHDA_DIR/config.yaml` — when `$GRAZHDA_DIR` is set
2. `~/.grazhda/config.yaml` — default fallback

### Example

```yaml
workspaces:
  - name: default
    default: true
    path: ~/ws
    clone_command_template: "git clone --branch {{.Branch}} git@github.com:myorg/{{.RepoName}} {{.DestDir}}"
    projects:
      - name: backend
        branch: main
        repositories:
          - name: api
          - name: auth
            branch: dev            # overrides project branch
          - name: api
            local_dir_name: api-v2 # cloned into <project>/api-v2

  - name: personal
    path: ~/personal
    clone_command_template: "git clone git@github.com:me/{{.RepoName}} {{.DestDir}}"
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
| `{{.RepoName}}` | `repository.name` |
| `{{.DestDir}}` | `<project_path>/<local_dir_name>` or `<project_path>/<name>` |

### Field Reference

<details>
<summary><strong>Workspace fields</strong></summary>

| Field | Required | Description |
| :--- | :---: | :--- |
| `name` | ✅ | Unique identifier; used with `--name` |
| `default` | — | Marks this workspace as the default target |
| `path` | ✅ | Root directory for the workspace (`~` is expanded) |
| `clone_command_template` | ✅ | Go template string for the clone command |
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
| `name` | ✅ | Repository name; used as `{{.RepoName}}` |
| `branch` | — | Overrides the project-level branch |
| `local_dir_name` | — | Clone destination name; overrides `name` |

</details>

---

## 🗺️ Roadmap

| Tool | Role | Status |
| :--- | :--- | :---: |
| **zgard** | Workspace lifecycle CLI | ✅ |
| **Grazhda installer** | Source-build installer script | 🚧 |
| **dukh** | Background gRPC worker | 📅 Phase 2 |
| **Molfar** | Orchestration server | 📅 Phase 3 |
| **Molf** | Orchestrator CLI | 📅 Phase 3 |

---

## 🛠️ Development

```bash
just build-zgard   # build → bin/zgard
just test          # run all tests
just fmt           # format all Go source
just tidy          # tidy all modules
```

### Module Layout

```
grazhda/
├── go.work
├── Justfile
├── config.template.yaml
├── internal/               # shared module
│   ├── color/              # terminal colour helpers
│   ├── config/             # load · validate · render templates
│   ├── executor/           # shell command interface + mock
│   ├── reporter/           # per-repo progress + run summary
│   └── workspace/          # init · purge · pull · targeting
└── zgard/                  # CLI module
    ├── main.go
    ├── root.go
    └── ws/                 # ws init · ws purge · ws pull
```

---

## 📄 License

GNU GPL v3 — see [`LICENSE`](LICENSE).


