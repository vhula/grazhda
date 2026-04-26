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

---

## Package Registry

grazhda ships a declarative package manager that installs developer tools into `$GRAZHDA_DIR/pkgs/`. Packages are defined in YAML registry files and installed in dependency order.

### Dual Registry System

Two registry files are loaded and merged at runtime:

| Registry | Path | Managed by |
| :--- | :--- | :--- |
| **Global** | `$GRAZHDA_DIR/.grazhda.pkgs.yaml` | `grazhda install` / `grazhda upgrade` — **replaced on each upgrade** |
| **Local** | `$GRAZHDA_DIR/registry.pkgs.local.yaml` | You — **survives upgrades** |

Use the local registry for personal additions or overrides. The global registry is the distribution default and should not be edited by hand.

### Merge Semantics

When both registries are loaded, local entries override global entries that share the **same identity**. Identity is the combination of `name` + `version`:

- Two packages with the same `name` and the same `version` (including both empty) are considered identical — the local entry wins.
- A local package with a **different** version from a global package of the same name is treated as a separate entry and both are kept.

### YAML Schema

Each registry file has a single top-level key `registry` containing a list of package objects:

```yaml
registry:
  - name: mytool
    version: "2.1.0"
    depends_on:
      - sdkman
      - jdk@17.0.8-tem
    pre_create_dir: true
    pre_install_env: |
      export MYTOOL_HOME="$GRAZHDA_DIR/pkgs/mytool"
    install: |
      curl -fsSL https://example.com/mytool-$VERSION.tar.gz | tar xz -C "$MYTOOL_HOME"
    post_install_env: |
      export PATH="$MYTOOL_HOME/bin:$PATH"
    purge: |
      rm -rf "$MYTOOL_HOME"
```

### Package Fields

| Field | Type | Required | Description |
| :--- | :--- | :---: | :--- |
| `name` | string | ✅ | Unique package identifier (e.g. `sdkman`, `jdk`) |
| `version` | string | — | Optional version string; injected as `$VERSION` into every phase script |
| `depends_on` | list | — | Packages that must be installed first; entries are `name` or `name@version` |
| `pre_create_dir` | bool | — | When `true`, create `$GRAZHDA_DIR/pkgs/<name>` before any phase runs |
| `pre_install_env` | string | — | Shell statements written to `.grazhda.env` **before** the install script runs |
| `install` | string | — | Bash script that performs the actual installation |
| `post_install_env` | string | — | Shell statements written to `.grazhda.env` **after** a successful install |
| `purge` | string | — | Bash script executed during `zgard pkg purge` before the package directory is removed |

### Environment Block Format

The `pre_install_env` and `post_install_env` fields are not executed directly. Their content is written into `$GRAZHDA_DIR/.grazhda.env` inside named marker blocks, then the env file is sourced so subsequent scripts see the exported variables.

Marker format:

```bash
# === BEGIN GRAZHDA: <key> ===
export SDKMAN_DIR="$GRAZHDA_DIR/pkgs/sdkman"
# === END GRAZHDA: <key> ===
```

- `<key>` is `<name>:pre` for `pre_install_env` and `<name>:post` for `post_install_env`.
- Blocks are idempotent — writing the same content again replaces the existing block in-place.
- `zgard pkg purge` removes the corresponding blocks from `.grazhda.env`.

### Dependency Resolution

`depends_on` entries reference other packages by name or by `name@version`:

```yaml
depends_on:
  - sdkman            # matches the sole "sdkman" package (any version)
  - jdk@17.0.8-tem    # matches only jdk with version "17.0.8-tem"
```

- A bare name resolves to the single package with that name. If multiple versions exist, use the `name@version` form.
- Transitive dependencies are expanded automatically — if A depends on B and B depends on C, requesting A pulls in both B and C.
- Packages are installed in **topological order** (Kahn's algorithm). A dependency cycle is detected and reported as an error.
- For purge operations, the order is **reversed** (dependents are purged before their dependencies).
