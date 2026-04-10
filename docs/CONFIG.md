# Configuration Reference

> [← Back to README](../README.md) · [CLI Reference](CLI.md)

`zgard` and `dukh` both resolve `config.yaml` from:

1. `$GRAZHDA_DIR/config.yaml` — when `$GRAZHDA_DIR` is set
2. `~/.grazhda/config.yaml` — default fallback

---

## Example

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
        tags: [backend, critical]        # project-level tags (inherited by repos)
        repositories:
          - name: api
          - name: auth
            branch: dev                  # overrides project branch
            tags: [auth]                 # repo-level tags (merged with project tags)
          - name: api
            local_dir_name: api-v2       # cloned into <project>/api-v2
          - name: org/pack/repo          # with structure:list → cloned as <project>/repo

  - name: personal
    path: ~/personal
    clone_command_template: "git clone git@github.com:me/{{.RepoName}} {{.DestDir}}"
    structure: list                      # last URL segment used as dest dir
    projects:
      - name: tools
        branch: main
        repositories:
          - name: dotfiles
          - name: scripts
```

---

## Clone Template Variables

| Variable | Resolves to |
| :--- | :--- |
| `{{.Branch}}` | `repository.branch` if set, otherwise `project.branch` |
| `{{.RepoName}}` | `repository.name` (full value, including any slashes) |
| `{{.DestDir}}` | Full filesystem path to the clone destination (see `structure`) |

---

## Workspace Structure Modes

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

---

## Field Reference

### Workspace fields

| Field | Required | Description |
| :--- | :---: | :--- |
| `name` | ✅ | Unique identifier; used with `--name` |
| `default` | — | Marks this workspace as the default target |
| `path` | ✅ | Root directory for the workspace (`~` is expanded) |
| `clone_command_template` | ✅ | Go template string for the clone command |
| `structure` | — | `tree` (default) or `list` — controls dest dir for repo names with `/` |
| `projects` | — | List of project subdirectories |

### Project fields

| Field | Required | Description |
| :--- | :---: | :--- |
| `name` | ✅ | Directory name under the workspace path |
| `branch` | ✅ | Default branch for all repos in this project |
| `tags` | — | List of string tags (inherited by child repositories) |
| `repositories` | — | List of repositories to clone |

### Repository fields

| Field | Required | Description |
| :--- | :---: | :--- |
| `name` | ✅ | Repository name; used as `{{.RepoName}}` (may contain `/`) |
| `branch` | — | Overrides the project-level branch |
| `local_dir_name` | — | Explicit clone destination name; overrides both `name` and `structure` |
| `tags` | — | List of string tags (merged with project-level tags) |

---

## Environment Variables

| Variable           | Description                                              | Default              |
|--------------------|----------------------------------------------------------|----------------------|
| `GRAZHDA_DIR`      | Installation root for grazhda (required for `dukh`)     | `~/.grazhda`         |
| `GRAZHDA_EDITOR`   | Override the editor for `grazhda config --edit`         | config `editor` field |
| `DUKH_HOST`        | Override the dukh gRPC server host                       | `localhost`          |
| `DUKH_PORT`        | Override the dukh gRPC server port                       | `50501`              |
| `NO_COLOR`         | Disable all ANSI colour output (any non-empty value)     | unset                |

> You can also pass `--no-color` as a flag to `zgard` for the same effect.
