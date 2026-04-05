---
stepsCompleted: [step-01-init, step-02-discovery, step-02b-vision, step-02c-executive-summary, step-03-success, step-04-journeys, step-05-domain-skipped, step-06-innovation-skipped, step-07-project-type, step-08-scoping, step-09-functional, step-10-nonfunctional, step-11-polish, step-12-complete]
workflowStatus: complete
completedAt: '2026-04-05T11:26:38Z'
inputDocuments: [README.md, config.template.yaml, BUILD.md]
workflowType: 'prd'
classification:
  projectType: cli_tool
  domain: developer_tools_workspace_automation
  complexity: medium
  projectContext: greenfield
---

# Product Requirements Document - zgard CLI

**Author:** jake
**Date:** 2026-04-05

## Executive Summary

`zgard` is a local-first CLI for declarative multi-repo workspace lifecycle management. Developers define their entire workspace topology — workspaces, projects, repositories, and branches — in a single YAML config file. `zgard` materializes that topology on disk: creating directory structures, cloning repositories, and executing templated clone commands exactly as specified. No server, no agent, no network dependency beyond the git remote itself.

**Target users:** Individual developers and teams who maintain structured multi-repo environments and need onboarding, environment reset, or workspace provisioning to be repeatable and zero-effort.

**Problem being solved:** Multi-repo workspace setup is manual, fragile, and undocumented. Each machine bootstrap or branch context-switch requires remembering which repos, which branches, and which directory structure — knowledge that lives in someone's head, not in code. `zgard` makes that implicit knowledge explicit and executable.

### What Makes This Special

`zgard` is the single point of control for a developer's entire multi-repo environment. The config file is the source of truth — it encodes not just what to clone, but *how* (via Go-templated clone commands), *where* (workspace + project hierarchy), and *which state* (per-repo branch). Phase 1 delivers workspace initialization and teardown. The architecture is designed to extend naturally to cross-workspace operations: pulling, searching, running LLM agents, and fanning out custom commands — all scoped to the same declared topology.

## Project Classification

| Field | Value |
|---|---|
| **Project Type** | CLI Tool (Go + Cobra) |
| **Domain** | Developer Tools / Workspace Automation |
| **Complexity** | Medium — hierarchical config model, Go template rendering, multi-workspace targeting semantics |
| **Project Context** | Greenfield — full reset, no existing code to preserve |
| **Phase 1 Scope** | `ws init`, `ws purge`, `ws pull` |
| **Future Scope** | Cross-repo search, LLM agent execution, custom command fanout, `dukh`/`molfar` integration |

## Success Criteria

### User Success

- A developer can run `zgard ws init` on a freshly provisioned machine and have a fully structured workspace — correct directory hierarchy, all repos cloned, correct branches checked out — without any manual steps.
- A developer can re-run `zgard ws init` after a partial failure and have it clone only the missing repositories, leaving already-cloned repos untouched.
- A developer can run `zgard ws pull` to bring all repositories in a workspace up to date with their configured branches via `git pull --rebase` in a single command.
- A developer can run `zgard ws purge` with confidence, knowing an interactive confirmation prompt prevents accidental destruction and the `--dry-run` flag shows exactly what will be removed.
- All output is clear, actionable, and scannable — success states, skip states, and failures are visually distinct.

### Business Success

- `zgard` is the tool developers reach for first when setting up or resetting a workspace — not a shell script or manual steps.
- The config file (`config.yaml`) becomes the authoritative record of a team's workspace topology, committed to source control and shared.
- The Phase 1 command surface (`ws init`, `ws purge`, `ws pull`) establishes the config model and targeting conventions that Phase 2+ features build on without breaking changes.

### Technical Success

- Config validation catches all errors (missing required fields, malformed Go templates, duplicate workspace/project names) before any filesystem changes occur.
- `zgard ws init` is fully idempotent — running it N times produces the same end state as running it once.
- All commands exit with correct exit codes (0 = success, non-zero = partial or full failure) and are usable in non-interactive scripts and CI pipelines.
- A `--no-confirm` flag (or equivalent) bypasses interactive prompts for automation contexts.
- `--dry-run` is supported on both `ws init` and `ws purge`, producing accurate previews without side effects.

### Measurable Outcomes

- `ws init` completes directory setup and dispatches clone commands with negligible overhead (sub-second for config parse + dir creation).
- Zero filesystem changes occur when config validation fails.
- A workspace with N repos, re-initialized after M failures, has exactly N repos cloned after a successful second run.

## Product Scope

### MVP Strategy

**Approach:** Problem-solving MVP — replace every manual step in the init/purge/pull workflow. The bar is not "feature complete" but "no developer needs a shell script or manual steps to set up a workspace."

**Resource profile:** Single developer. Every MVP feature must justify its place.

**Platforms:** Linux (amd64) and macOS (amd64, arm64). Windows out of scope for MVP.

### MVP — Minimum Viable Product (Phase 1)

- `zgard ws init` — initializes the default workspace (or `--ws <name>` / `--all`); creates project directories; clones repos via Go-templated clone commands; skips already-cloned repos; continues on clone failure; reports failures at end; validates config up-front.
- `zgard ws purge` — removes the targeted workspace directory; requires explicit `--ws <name>` or `--all`; interactive Y/N confirmation; `--dry-run` support.
- `zgard ws pull` — runs `git pull --rebase` on each repo's configured branch; same targeting flags as `ws init`.
- Shared flags across all commands: `--dry-run`, `--verbose` / `-v`, `--no-confirm`, `--parallel`.
- Config model: workspaces → projects → repositories (branch + optional `local_dir_name` per repo); Go template variables `{{.Branch}}`, `{{.RepoName}}`, `{{.DestDir}}`.

**MVP must-have justifications:**

| Capability | Justification |
|---|---|
| `ws init` with targeting | Core value — without this, the product doesn't exist |
| `ws purge` with confirmation | Lifecycle completion — init without purge is incomplete |
| `ws pull` with rebase | Closes the "keep in sync" gap; first candidate to defer if schedule slips |
| Up-front config validation | Filesystem ops must not start on a bad config |
| Skip-if-exists idempotency | Without this, re-running init is destructive or fragile |
| Continue-on-failure + summary | One network drop otherwise leaves workspace in unknown state |
| `--dry-run`, `--verbose`, `--no-confirm` | Dry-run makes purge safe; `--no-confirm` unblocks CI use |
| Non-zero exit codes + stderr/stdout split | Required for all scripting use cases |

### Growth Features (Post-MVP)

- `zgard ws search` — grep/glob across all repos in a workspace.
- `zgard ws run <command>` — fan out a custom shell command across all repos/projects in a workspace.
- Shell completion (bash, zsh, fish).
- `zgard config init` — scaffold a starter `config.yaml`.

### Vision (Future)

- `zgard ws agent` — LLM agent execution scoped to a workspace for cross-repo analysis.
- `dukh` integration — process lifecycle management within workspaces.
- `molfar` orchestration consuming `zgard` as a workspace primitive.
- Windows platform support.

### Risk Mitigation

- *Go template rendering edge cases* — mitigated by up-front template validation before any filesystem work.
- *`os/exec` command injection* — `config.yaml` is a trusted file; rendered commands execute as full shell strings.
- *Path handling macOS vs Linux* — `filepath` package throughout; no string path concatenation.
- *Single developer schedule risk* — `ws pull` is the first feature to defer to Phase 2 if Phase 1 runs long.

## User Journeys

### Journey 1: The Developer — New Machine Setup (Happy Path)

**Meet Alex.** Alex just received a new laptop. Their team works across six repositories split between two projects: a `main` branch for the stable service layer and a `dev` branch for active feature work. In the past, setting up a new machine meant a full afternoon of cloning repos one by one, making sure each was on the right branch, and re-creating the exact directory structure they had before. Three times this year, someone got a repo on the wrong branch and didn't notice for days.

Today Alex pulls down the team's `config.yaml` from the shared repo, drops it in `$GRAZHDA_DIR`, and runs:

```
zgard ws init
```

`zgard` validates the config instantly — catching a missing `branch` on one repo and reporting it clearly before touching the filesystem. Alex fixes it, re-runs. This time: directories created, clone commands dispatched, progress streamed to the terminal. Two minutes later, the workspace is ready — correct hierarchy, correct branches, exactly as declared. Alex runs `zgard ws init` again just to be sure. Nothing happens. All repos already exist; they're skipped silently. ✓

**What this journey reveals:** config up-front validation; clear error messages with location; directory creation; clone command execution via template; idempotency; progress output; `--verbose` flag for detail.

---

### Journey 2: The Developer — Partial Failure Recovery (Edge Case)

**Meet Sam.** Sam runs `zgard ws init` on a flaky hotel WiFi. Three of eight repos clone successfully before the connection drops. The terminal shows a clear failure summary: five repos failed, three succeeded, workspace partially initialized.

The next morning Sam re-runs `zgard ws init`. `zgard` checks each repo directory — three exist and are skipped, five are missing and cloned fresh. The workspace is now complete. Sam also wants to check what *would* happen before running `ws purge` later:

```
zgard ws purge --ws default --dry-run
```

The terminal lists every directory that would be deleted. Sam runs it without `--dry-run`, confirms the Y/N prompt, and the workspace is cleanly removed.

**What this journey reveals:** skip-if-exists logic; per-repo failure tracking; end-of-run failure summary; non-zero exit code on partial failure; `--dry-run` for purge; interactive confirmation prompt.

---

### Journey 3: The Config Author — Team Setup Distribution

**Meet Jordan.** Jordan is the senior engineer who owns the team's `config.yaml`. They need to add a new project with three repos — two on `main`, one on a feature branch — to the workspace definition. Jordan edits the config and runs:

```
zgard ws init --dry-run --verbose
```

The terminal shows exactly what *would* happen: which directories would be created, which clone commands would be run, which repos would be skipped. No changes made. Jordan is satisfied, commits the updated config, and shares it with the team. Anyone who pulls it and runs `zgard ws init` gets the new project added to their workspace — existing repos are untouched, only the new ones are cloned.

Jordan also wants to keep repos current across the team:

```
zgard ws pull --all
```

All repos across all workspaces are pulled and rebased to their configured branches. Jordan gets a clean summary of which succeeded, which had conflicts, and which were skipped (repos not yet initialized).

**What this journey reveals:** `--dry-run` + `--verbose` as a config-authoring workflow; additive init behavior; `ws pull` with `--all`; pull summary output; graceful handling of repos that don't exist yet during pull (skip, not error).

---

### Journey 4: The Automation Script — CI / Non-Interactive Use

A team's onboarding script provisions a dev container and needs to initialize a workspace without any interactive prompts:

```bash
zgard ws init --ws default --no-confirm
if [ $? -ne 0 ]; then
  echo "Workspace init failed" && exit 1
fi
```

`zgard` runs silently — no Y/N prompts, no spinner, clean output — and exits `0` on full success or non-zero if any repo failed. The script acts on the exit code. Logs are captured for audit.

**What this journey reveals:** `--no-confirm` flag; scripting-safe exit codes; clean, non-interactive stdout output; no TTY-dependent behavior.

---

### Journey Requirements Summary

| Capability | Journeys |
|---|---|
| Up-front config validation with precise error locations | 1, 3 |
| Directory creation from workspace/project hierarchy | 1 |
| Clone command execution via Go template rendering | 1 |
| Skip-if-exists idempotency | 1, 2, 3 |
| Continue-on-failure with end-of-run failure summary | 2 |
| Non-zero exit code on partial or full failure | 2, 4 |
| `--dry-run` flag (init + purge) | 2, 3 |
| `--verbose` flag for detailed output | 3 |
| `--no-confirm` flag for non-interactive use | 4 |
| Interactive Y/N confirmation for purge | 2 |
| `ws pull --rebase` on configured branches | 3 |
| `--ws <name>` and `--all` targeting flags | 3, 4 |
| Pull summary (success / conflict / skipped) | 3 |

## CLI Tool Specific Requirements

### Project-Type Overview

`zgard` is a local developer CLI built with Go + Cobra. It is both interactive (human-operated, coloured terminal output) and scriptable (clean exit codes, `--no-confirm` flag for automation). The config file at `$GRAZHDA_DIR/config.yaml` is the single source of truth. Shell completion is a post-MVP growth feature.

### Command Structure

```
zgard
└── ws
    ├── init    [--ws <name> | --all] [--dry-run] [--verbose | -v] [--no-confirm]
    ├── purge   [--ws <name> | --all] [--dry-run] [--verbose | -v] [--no-confirm]
    └── pull    [--ws <name> | --all] [--dry-run] [--verbose | -v] [--no-confirm]
```

- All `ws` subcommands share the same targeting and behavior flags.
- `ws init` and `ws pull` default to the workspace named `default` when no targeting flag is given; `ws purge` requires explicit `--ws` or `--all`.
- The `--all` flag targets every workspace defined in the config.
- Global flags (e.g. `--help`, version) follow Cobra conventions.

### Output Formats

- **Terminal output only** — human-readable, coloured via `charmbracelet/log`.
- Output levels: INFO (progress, skips), WARN (non-fatal issues), ERROR (failures).
- `--verbose` / `-v` enables detailed per-operation output (e.g. exact clone command being run).
- `--dry-run` prefixes all output lines with `[DRY RUN]` to make preview mode unambiguous.
- End-of-run summary for `ws init` and `ws pull`: counts of succeeded, skipped, and failed operations.
- No JSON or machine-readable output format. Exit codes carry the scripting signal.

### Config Schema

Config file location: `$GRAZHDA_DIR/config.yaml` (no alternatives supported).

```yaml
workspaces:
  - name: <string>                        # required; unique identifier; "default" = implicit default workspace
    path: <string>                        # required; absolute path for workspace root
    clone_command_template: <string>      # required; Go template with {{.Branch}}, {{.RepoName}}, {{.DestDir}}
    projects:
      - name: <string>                    # required; project directory name
        repositories:
          - name: <string>               # required; repo name (used as {{.RepoName}})
            branch: <string>             # required; branch to clone/pull
            local_dir_name: <string>     # optional; overrides local directory name ({{.DestDir}})
```

**Validation rules** (enforced before any filesystem changes):

- Workspace `name` must be unique across all workspaces.
- The workspace named `default` is the implicit default; no flag required or supported.
- `ws init` and `ws pull` with no targeting flags operate on the `default` workspace; if none exists, error with a clear message.
- `path` and `clone_command_template` are required on every workspace.
- Project `name` must be unique within a workspace.
- Repository `name` and `branch` are required on every repository entry.
- `clone_command_template` must be a valid Go template containing at minimum `{{.RepoName}}` and `{{.DestDir}}`.

### Implementation Considerations

- Built with Go + Cobra; each `ws` subcommand is a separate Cobra command registered under a `ws` parent.
- Clone commands are executed via `os/exec` using the rendered Go template output as the shell command string.
- Config loaded once at startup via `gopkg.in/yaml.v3`; validation runs immediately after parsing.
- Logging via `github.com/charmbracelet/log` with level control tied to `--verbose` flag.
- Workspace targeting logic (default / `--ws` / `--all`) is shared across all three commands via a common resolver function.

## Functional Requirements

### Configuration Management

- **FR1:** The system can load the YAML config file from `$GRAZHDA_DIR/config.yaml` at startup.
- **FR2:** The system can validate the config file for all required fields before performing any filesystem operation.
- **FR3:** The system can validate that workspace names are unique across all workspaces in the config.
- **FR4:** The system can validate that project names are unique within each workspace.
- **FR5:** The system can validate that each `clone_command_template` is a syntactically valid Go template.
- **FR6:** The system can report all config validation errors with precise field location context before any filesystem changes occur.
- **FR7:** The system can resolve the implicit default workspace as the workspace with `name: default`.
- **FR8:** The system can report a clear, actionable error when no `default` workspace exists and no targeting flag is provided.

### Workspace Targeting

- **FR9:** Users can target the `default` workspace implicitly (no flag) when running `ws init` or `ws pull`.
- **FR10:** Users can target a specific named workspace with `--ws <name>` / `-w <name>` on any `ws` command.
- **FR11:** Users can target all configured workspaces simultaneously with `--all` on any `ws` command.
- **FR12:** The system requires an explicit `--ws <name>` or `--all` flag to execute `ws purge` — no implicit default.

### Workspace Initialization

- **FR13:** Users can create a workspace's full directory structure (workspace root and all project subdirectories) from config with `ws init`.
- **FR14:** Users can clone all configured repositories into their project directories using the workspace's `clone_command_template`.
- **FR15:** The system can render clone command templates substituting `{{.Branch}}`, `{{.RepoName}}`, and `{{.DestDir}}` per repository.
- **FR16:** Users can override a repository's local directory name via the optional `local_dir_name` field, which becomes `{{.DestDir}}`.
- **FR17:** The system can skip cloning a repository when a local directory already exists at the expected path (idempotent re-init).
- **FR18:** The system can continue processing remaining repositories when a clone operation fails, rather than aborting.
- **FR19:** Users can preview all directory creation and clone operations that `ws init` would perform, without executing them (`--dry-run`).

### Workspace Teardown

- **FR20:** Users can remove a targeted workspace's directory structure with `ws purge`.
- **FR21:** The system prompts users for interactive Y/N confirmation before executing any destructive purge operation.
- **FR22:** Users can preview all directories that `ws purge` would remove without executing the operation (`--dry-run`).

### Repository Synchronization

- **FR23:** Users can update all repositories in a targeted workspace to their configured branches with `ws pull`.
- **FR24:** The system can execute `git pull --rebase` against each repository's configured branch.
- **FR25:** The system can skip repositories that do not yet have a local directory during `ws pull` (not treated as an error).
- **FR26:** Users can preview which repositories `ws pull` would update without executing the operation (`--dry-run`).

### Operation Feedback & Reporting

- **FR27:** The system provides real-time progress output for each individual operation (directory creation, clone, pull).
- **FR28:** Users can enable verbose output to see the exact commands being executed (`--verbose` / `-v`).
- **FR29:** The system reports an end-of-run summary for `ws init` and `ws pull` showing counts of succeeded, skipped, and failed operations.
- **FR30:** The system reports each failed operation with enough context (repo name, project, error message) to diagnose the cause.
- **FR31:** The system clearly marks all output as preview mode when `--dry-run` is active (e.g. `[DRY RUN]` prefix).

### Automation & Scripting Support

- **FR32:** The system exits with code `0` when all operations in a command complete successfully.
- **FR33:** The system exits with a non-zero code when any operation fails or when config validation fails.
- **FR34:** Users can suppress all interactive prompts for non-interactive scripts and CI pipelines (`--no-confirm`).
- **FR35:** The system writes error and failure messages to stderr and progress/info messages to stdout.
- **FR36:** Users can enable parallel repository operations for `ws init` and `ws pull` via `--parallel` to reduce total execution time.

## Non-Functional Requirements

### Performance

- **NFR1:** Config file loading and validation must complete in under 500ms for configs with up to 10 workspaces and 100 repositories total.
- **NFR2:** Directory creation and pre-clone setup (all filesystem work excluding actual git clone/pull network operations) must complete in under 1 second.
- **NFR3:** By default, repository operations (clone, pull) execute sequentially to avoid overwhelming the network or git host with concurrent connections.
- **NFR4:** Users can enable parallel repository operations via `--parallel` to reduce total wall-clock time when network bandwidth allows.

### Reliability

- **NFR5:** A failed clone or pull operation must leave no partial artifacts — if a clone fails mid-way, any partially created directory must be cleaned up.
- **NFR6:** Config validation failure must guarantee zero filesystem changes have occurred before the error is reported.
- **NFR7:** `--dry-run` output must accurately reflect what the real run would do — no divergence between preview and execution behaviour.
- **NFR8:** The tool must never silently ignore errors; every failure must produce a visible, actionable message.

### Portability

- **NFR9:** `zgard` must build and run correctly on Linux (amd64) and macOS (amd64 and arm64) without platform-specific code paths.
- **NFR10:** All filesystem path construction must use `filepath` package functions — no string concatenation for paths.
- **NFR11:** The binary must have no runtime dependencies beyond the OS and `git` being present in `$PATH`.

### Testability

- **NFR12:** Workspace domain logic (`ws init`, `ws purge`, `ws pull`) must be implemented in a package separate from the CLI entry point to enable unit testing without invoking Cobra.
- **NFR13:** Clone and pull command execution must be injectable (via an interface or function parameter) to allow tests to substitute a mock executor without spawning real git processes.
- **NFR14:** Config loading and validation must be independently testable with fixture YAML files, without requiring a real filesystem or git remote.
