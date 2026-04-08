---
stepsCompleted: [step-01-validate-prerequisites, step-02-design-epics, step-03-create-stories, step-04-final-validation]
workflowStatus: complete
completedAt: '2026-04-05T12:12:00Z'
inputDocuments:
  - docs/prd.md
  - docs/architecture.md
  - docs/ux-design-specification.md
---

# Grazhda — zgard CLI Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for **Grazhda — zgard CLI**, decomposing all requirements from the PRD, Architecture, and UX Design Specification into development-ready stories.

---

## Requirements Inventory

### Functional Requirements

- FR1: The system can load the YAML config file from `$GRAZHDA_DIR/config.yaml` at startup.
- FR2: The system can validate the config file for all required fields before performing any filesystem operation.
- FR3: The system can validate that workspace names are unique across all workspaces in the config.
- FR4: The system can validate that project names are unique within each workspace.
- FR5: The system can validate that each `clone_command_template` is a syntactically valid Go template.
- FR6: The system can report all config validation errors with precise field location context before any filesystem changes occur.
- FR7: The system can resolve the implicit default workspace as the workspace with `name: default`.
- FR8: The system can report a clear, actionable error when no `default` workspace exists and no targeting flag is provided.
- FR9: Users can target the `default` workspace implicitly (no flag) when running `ws init` or `ws pull`.
- FR10: Users can target a specific named workspace with `--name <name>` / `-n <name>` on any `ws` command.
- FR11: Users can target all configured workspaces simultaneously with `--all` on any `ws` command.
- FR12: The system requires an explicit `--name <name>` or `--all` flag to execute `ws purge` — no implicit default.
- FR13: Users can create a workspace's full directory structure (workspace root and all project subdirectories) from config with `ws init`.
- FR14: Users can clone all configured repositories into their project directories using the workspace's `clone_command_template`.
- FR15: The system can render clone command templates substituting `{{.Branch}}`, `{{.RepoName}}`, and `{{.DestDir}}` per repository.
- FR16: Users can override a repository's local directory name via the optional `local_dir_name` field, which becomes `{{.DestDir}}`.
- FR17: The system can skip cloning a repository when a local directory already exists at the expected path (idempotent re-init).
- FR18: The system can continue processing remaining repositories when a clone operation fails, rather than aborting.
- FR19: Users can preview all directory creation and clone operations that `ws init` would perform, without executing them (`--dry-run`).
- FR20: Users can remove a targeted workspace's directory structure with `ws purge`.
- FR21: The system prompts users for interactive Y/N confirmation before executing any destructive purge operation.
- FR22: Users can preview all directories that `ws purge` would remove without executing the operation (`--dry-run`).
- FR23: Users can update all repositories in a targeted workspace to their configured branches with `ws pull`.
- FR24: The system can execute `git pull --rebase` against each repository's configured branch.
- FR25: The system can skip repositories that do not yet have a local directory during `ws pull` (not treated as an error).
- FR26: Users can preview which repositories `ws pull` would update without executing the operation (`--dry-run`).
- FR27: The system provides real-time progress output for each individual operation (directory creation, clone, pull).
- FR28: Users can enable verbose output to see the exact commands being executed (`--verbose` / `-v`).
- FR29: The system reports an end-of-run summary for `ws init` and `ws pull` showing counts of succeeded, skipped, and failed operations.
- FR30: The system reports each failed operation with enough context (repo name, project, error message) to diagnose the cause.
- FR31: The system clearly marks all output as preview mode when `--dry-run` is active (e.g. `[DRY RUN]` prefix).
- FR32: The system exits with code `0` when all operations in a command complete successfully.
- FR33: The system exits with a non-zero code when any operation fails or when config validation fails.
- FR34: Users can suppress all interactive prompts for non-interactive scripts and CI pipelines (`--no-confirm`).
- FR35: The system writes error and failure messages to stderr and progress/info messages to stdout.
- FR36: Users can enable parallel repository operations for `ws init` and `ws pull` via `--parallel` to reduce total execution time.

### Non-Functional Requirements

- NFR1: Config file loading and validation must complete in under 500ms for configs with up to 10 workspaces and 100 repositories total.
- NFR2: Directory creation and pre-clone setup must complete in under 1 second.
- NFR3: By default, repository operations (clone, pull) execute sequentially.
- NFR4: Users can enable parallel repository operations via `--parallel`.
- NFR5: A failed clone or pull operation must leave no partial artifacts — if a clone fails mid-way, any partially created directory must be cleaned up.
- NFR6: Config validation failure must guarantee zero filesystem changes have occurred before the error is reported.
- NFR7: `--dry-run` output must accurately reflect what the real run would do — no divergence between preview and execution behaviour.
- NFR8: The tool must never silently ignore errors; every failure must produce a visible, actionable message.
- NFR9: `zgard` must build and run correctly on Linux (amd64) and macOS (amd64 and arm64) without platform-specific code paths.
- NFR10: All filesystem path construction must use `filepath` package functions — no string concatenation for paths.
- NFR11: The binary must have no runtime dependencies beyond the OS and `git` being present in `$PATH`.
- NFR12: Workspace domain logic must be implemented in a package separate from the CLI entry point to enable unit testing without invoking Cobra.
- NFR13: Clone and pull command execution must be injectable (via an interface or function parameter) to allow tests to substitute a mock executor without spawning real git processes.
- NFR14: Config loading and validation must be independently testable with fixture YAML files, without requiring a real filesystem or git remote.

### Additional Requirements (Architecture)

- **Three-module workspace:** `cmd/` (`github.com/vhula/grazhda/cmd`), `internal/` (`github.com/vhula/grazhda/internal`), and `zgard/` (`github.com/vhula/grazhda/zgard`) are separate Go modules registered in `go.work`.
- **Executor interface:** `Run(dir string, command string) error` — `OsExecutor` wraps `sh -c`; `MockExecutor` records calls for tests.
- **RunOptions struct:** `{DryRun, Verbose, Parallel, NoConfirm bool}` — passed to all workspace functions; never read Cobra flags inside `internal/`.
- **Reporter:** writes to stdout (progress) and stderr (errors); `Record(OpResult)`, `Summary()`, `ExitCode()`; coloured output via `fatih/color`.
- **OpResult:** `{Workspace, Project, Repo string; Skipped bool; Err error}`.
- **Defer cleanup (NFR5):** `success := false; defer func() { if !success { os.RemoveAll(destPath) } }()`.
- **Parallel:** `sync.WaitGroup` + goroutines; mutex-protected Reporter; uncapped Phase 1.
- **`$GRAZHDA_DIR`:** read in `cmd/ws/*.go` only; defaults to `$HOME/.grazhda`; passed as string to `config.Load(path)`.
- **Confirmation prompt:** inline `confirm(prompt string, r io.Reader) bool` in `internal/workspace/purge.go`.
- **Dependencies:** `github.com/spf13/cobra v1.9.1`, `gopkg.in/yaml.v3 v3.0.1`, `github.com/fatih/color v1.19.0`.
- **Build:** `just build-zgard` → `bin/zgard`; `just test` → `go test ./...` across all modules.

### UX Design Requirements

- UX-DR1: Every status line uses the symbol vocabulary: `✓` (success), `✗` (failure), `⏭` (skipped), `[DRY RUN]` (preview prefix).
- UX-DR2: Output follows workspace → project → repo hierarchy with defined indentation: 0 spaces (workspace), 2 spaces (project), 4 spaces (repo line).
- UX-DR3: Dry-run output is structurally identical to real-run output — only the `[DRY RUN]` prefix distinguishes them; no different format or detail level.
- UX-DR4: Summary block is always the last thing printed: `✓ N cloned  ⏭ N skipped  ✗ N failed` (verb varies by command), preceded by a blank line.
- UX-DR5: stdout receives progress/INFO lines; stderr receives error/failure lines — safe for independent redirection.
- UX-DR6: Colour is auto-detected via TTY — full colour in terminal, plain text when piped/redirected; all status conveyed by symbol first, colour second.
- UX-DR7: `ws purge` confirmation prompt lists every path to be removed before asking `Confirm? [y/N]:` — default answer is `N`.
- UX-DR8: Error detail lines format: `      <repo>: <error context>` — printed in summary block on stderr after the count line.
- UX-DR9: Verbose mode (`--verbose`) prints the rendered command on the line before execution: `  → <rendered command>`.
- UX-DR10: Section headers use bold labels — `Workspace: <name>` at root, `  Project: <name>` at 2-space indent.

---

### FR Coverage Map

| FR | Epic | Story |
|---|---|---|
| FR1 | Epic 2 | Story 2.1 |
| FR2 | Epic 2 | Story 2.2 |
| FR3 | Epic 2 | Story 2.2 |
| FR4 | Epic 2 | Story 2.2 |
| FR5 | Epic 2 | Story 2.3 |
| FR6 | Epic 2 | Story 2.2 |
| FR7 | Epic 2 | Story 2.1 |
| FR8 | Epic 2 | Story 2.1 |
| FR9 | Epic 3 | Story 3.1 |
| FR10 | Epic 3 | Story 3.1 |
| FR11 | Epic 3 | Story 3.1 |
| FR12 | Epic 4 | Story 4.1 |
| FR13 | Epic 3 | Story 3.2 |
| FR14 | Epic 3 | Story 3.3 |
| FR15 | Epic 3 | Story 3.3 |
| FR16 | Epic 3 | Story 3.3 |
| FR17 | Epic 3 | Story 3.4 |
| FR18 | Epic 3 | Story 3.5 |
| FR19 | Epic 3 | Story 3.7 |
| FR20 | Epic 4 | Story 4.2 |
| FR21 | Epic 4 | Story 4.3 |
| FR22 | Epic 4 | Story 4.4 |
| FR23 | Epic 5 | Story 5.1 |
| FR24 | Epic 5 | Story 5.1 |
| FR25 | Epic 5 | Story 5.2 |
| FR26 | Epic 5 | Story 5.3 |
| FR27 | Epic 3 | Story 3.6 |
| FR28 | Epic 3 | Story 3.6 |
| FR29 | Epic 3 | Story 3.6 |
| FR30 | Epic 3 | Story 3.6 |
| FR31 | Epic 3 | Story 3.7 |
| FR32 | Epic 3 | Story 3.7 |
| FR33 | Epic 3 | Story 3.7 |
| FR34 | Epic 3 | Story 3.7 |
| FR35 | Epic 3 | Story 3.6 |
| FR36 | Epic 3 | Story 3.8 |

---

## Epic List

### Epic 1: Project Foundation
Developers can build and run `zgard` from source; the full development loop (build, test, format) works end-to-end. All package skeletons are in place with correct module boundaries, ready to receive implementation.
**FRs covered:** None directly (enables all FRs via NFR9, NFR12, NFR13, NFR14 and architecture requirements)

### Epic 2: Configuration Loading & Validation
Users can define their entire workspace topology in `config.yaml` and receive complete, actionable validation errors before any filesystem operation occurs.
**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR6, FR7, FR8

### Epic 3: Workspace Initialization (`ws init`)
Users can create their full workspace directory structure and clone all configured repositories in a single command, with real-time progress, dry-run preview, and a summary of results.
**FRs covered:** FR9, FR10, FR11, FR13, FR14, FR15, FR16, FR17, FR18, FR19, FR27, FR28, FR29, FR30, FR31, FR32, FR33, FR34, FR35, FR36

### Epic 4: Workspace Teardown (`ws purge`)
Users can safely remove a targeted workspace's directory structure with confirmation, dry-run preview, and CI-compatible non-interactive mode.
**FRs covered:** FR12, FR20, FR21, FR22

### Epic 5: Repository Synchronization (`ws pull`)
Users can keep all repositories in a targeted workspace up to date with their configured branches, with dry-run preview and parallel execution.
**FRs covered:** FR23, FR24, FR25, FR26, FR36 (pull context)

---

## Epic 1: Project Foundation

Developers can build and run `zgard` from source. The three Go modules (`cmd`, `internal`, `zgard`) are initialised, registered in `go.work`, all dependencies installed, package skeletons in place, and `just build-zgard` produces a runnable binary. The development loop (build, test, format, tidy) works end-to-end.

### Story 1.1: Initialize Three-Module Go Workspace

As a developer,
I want the three Go modules (`cmd/`, `internal/`, `zgard/`) initialized and registered in `go.work`,
So that the monorepo module boundaries are established and the workspace is ready for package implementation.

**Acceptance Criteria:**

**Given** the repository root contains `go.work`
**When** I run `go work use ./cmd ./internal ./zgard`
**Then** all three modules are registered in `go.work` and `go work sync` succeeds without errors

**Given** each module directory exists
**When** I inspect `cmd/go.mod`, `internal/go.mod`, `zgard/go.mod`
**Then** their module paths are `github.com/vhula/grazhda/cmd`, `github.com/vhula/grazhda/internal`, and `github.com/vhula/grazhda/zgard` respectively

**Given** the three modules are registered
**When** I run `go work sync`
**Then** `go.work.sum` is updated and no errors occur

### Story 1.2: Install Dependencies in Each Module

As a developer,
I want all required dependencies (`cobra`, `yaml.v3`, `github.com/fatih/color`) installed in the correct modules,
So that all packages can import their dependencies without version conflicts.

**Acceptance Criteria:**

**Given** the `cmd` module
**When** I run `go get github.com/spf13/cobra@v1.9.1` and `go get github.com/fatih/color`
**Then** `cmd/go.mod` lists both dependencies and `cmd/go.sum` is updated

**Given** the `internal` module
**When** I run `go get gopkg.in/yaml.v3@v3.0.1` and `go get github.com/fatih/color`
**Then** `internal/go.mod` lists both dependencies

**Given** the `zgard` module
**When** I run `go get github.com/vhula/grazhda/cmd` and `go get github.com/vhula/grazhda/internal`
**Then** `zgard/go.mod` lists both as `require` entries resolved via `go.work`

**Given** all modules have dependencies installed
**When** I run `go build ./...` from the repo root
**Then** the build completes without "cannot find module" errors

### Story 1.3: Create Package Skeleton

As a developer,
I want all packages created with correct package declarations and empty placeholder files,
So that the full file tree exists and import paths are valid before any logic is written.

**Acceptance Criteria:**

**Given** the `internal` module
**When** I inspect the package directories
**Then** the following packages exist with valid `package` declarations:
- `internal/config/` (package `config`)
- `internal/targeting/` (package `targeting`)
- `internal/executor/` (package `executor`)
- `internal/workspace/` (package `workspace`)
- `internal/reporter/` (package `reporter`)

**Given** the `cmd` module
**When** I inspect the package directories
**Then** `cmd/ws/` exists with package `ws`

**Given** all package skeleton files exist
**When** I run `go build ./...` from the repo root
**Then** all three modules compile without errors

**Given** the `internal/testdata/` directory
**When** I inspect it
**Then** it contains at least one fixture YAML file: `valid_single_workspace.yaml`

### Story 1.4: Wire Cobra CLI Root and `ws` Subcommand

As a developer,
I want `zgard/main.go` wired with Cobra root and the `ws` parent command registered,
So that `zgard --help` and `zgard ws --help` both work and display correct usage information.

**Acceptance Criteria:**

**Given** `zgard/main.go` exists
**When** I run `./bin/zgard --help`
**Then** output shows `Usage: zgard [command]` and lists `ws` as an available command

**Given** the `ws` parent command is registered
**When** I run `./bin/zgard ws --help`
**Then** output shows `Usage: zgard ws [command]` with `init`, `purge`, and `pull` listed as subcommands

**Given** unknown flags or subcommands
**When** I run `./bin/zgard unknown`
**Then** the command exits with a non-zero code and prints a helpful error message to stderr

### Story 1.5: Set Up Build, Test, and Dev Pipeline

As a developer,
I want `just build-zgard`, `just test`, `just fmt`, and `just tidy` to all work,
So that the development feedback loop is fully operational.

**Acceptance Criteria:**

**Given** the Justfile targets exist
**When** I run `just build-zgard`
**Then** `bin/zgard` is produced and is executable

**Given** the package skeleton with no tests yet
**When** I run `just test`
**Then** `go test ./...` runs across `cmd/`, `internal/`, and `zgard/` and exits `0` (no failures; no tests is a pass)

**Given** any formatting drift
**When** I run `just fmt`
**Then** `gofmt` or `goimports` runs across all modules without errors

**Given** outdated `go.sum` entries
**When** I run `just tidy`
**Then** `go mod tidy` runs across all modules and `go.work.sum` is consistent

---

## Epic 2: Configuration Loading & Validation

Users can define their workspace topology in `$GRAZHDA_DIR/config.yaml` and receive complete, actionable validation errors before any filesystem changes occur. Config loading is independently testable with fixture YAML files.

### Story 2.1: Load Config File and Resolve Default Workspace

As a developer using `zgard`,
I want the tool to load `config.yaml` from `$GRAZHDA_DIR` and resolve the default workspace,
So that I can run commands without specifying a workspace name when a `default` workspace is configured.

**Acceptance Criteria:**

**Given** `$GRAZHDA_DIR/config.yaml` contains a valid config with a workspace named `default`
**When** `config.Load(path)` is called
**Then** it returns a populated `*Config` struct with all workspaces, projects, and repositories parsed

**Given** a config with `name: default` in one workspace
**When** `config.DefaultWorkspace(cfg)` is called
**Then** it returns that workspace without error

**Given** a config with no workspace named `default`
**When** `config.DefaultWorkspace(cfg)` is called
**Then** it returns a descriptive error: `"no default workspace found: add a workspace with name: default or use --name"`

**Given** `$GRAZHDA_DIR` is not set
**When** the `cmd` layer resolves the config path
**Then** it defaults to `$HOME/.grazhda/config.yaml`

**Given** the config file does not exist at the resolved path
**When** `config.Load(path)` is called
**Then** it returns an error with the full path and `os.ErrNotExist`

**Given** a fixture YAML file `valid_single_workspace.yaml` in `internal/testdata/`
**When** `config.Load` is called in a unit test
**Then** the test passes without requiring a real filesystem or `$GRAZHDA_DIR`

### Story 2.2: Validate Config Fields, Uniqueness, and Error Reporting

As a developer using `zgard`,
I want the tool to validate all required fields and uniqueness constraints before touching the filesystem,
So that I receive all errors in a single run with precise field context, not one at a time.

**Acceptance Criteria:**

**Given** a config with a workspace missing a `path` field
**When** `config.Validate(cfg)` is called
**Then** it returns an error slice containing `"workspace[0].path: required field missing"`

**Given** a config with two workspaces sharing the same `name`
**When** `config.Validate(cfg)` is called
**Then** it returns an error containing `"workspace names must be unique: duplicate name 'default'"`

**Given** a config with two projects in the same workspace sharing the same `name`
**When** `config.Validate(cfg)` is called
**Then** it returns an error containing `"workspace[0].projects: duplicate project name 'backend'"`

**Given** a config with multiple validation errors
**When** `config.Validate(cfg)` is called
**Then** it returns all errors collected (not just the first), as a slice of strings

**Given** any config validation error
**When** it is reported via the `cmd` layer
**Then** zero filesystem changes have occurred (NFR6)

**Given** fixture files `duplicate_workspace_names.yaml` and `missing_branch.yaml`
**When** tests call `config.Validate` with these fixtures
**Then** the correct errors are returned in both cases

### Story 2.3: Validate Clone Command Templates

As a developer using `zgard`,
I want the tool to validate that every `clone_command_template` is a valid Go template before running any operations,
So that template errors are caught up-front with clear context rather than failing mid-clone.

**Acceptance Criteria:**

**Given** a workspace with `clone_command_template: "git clone {{.Branch}} {{.RepoName}}"`
**When** `config.Validate(cfg)` is called
**Then** no template error is returned

**Given** a workspace with `clone_command_template: "git clone {{.Unclosed"`
**When** `config.Validate(cfg)` is called
**Then** an error is returned containing `"workspace[0].clone_command_template: invalid template"` and the Go template parse error

**Given** a valid template and a repository with `local_dir_name` set
**When** `config.RenderCloneCmd(template, repo)` is called
**Then** `{{.DestDir}}` is substituted with `local_dir_name`; `{{.RepoName}}` with `repo.name`; `{{.Branch}}` with `repo.branch`

**Given** a repository with no `local_dir_name`
**When** `config.RenderCloneCmd(template, repo)` is called
**Then** `{{.DestDir}}` is substituted with `repo.name`

**Given** fixture file `invalid_template.yaml`
**When** `config.Validate` is called in a unit test
**Then** the test confirms the correct template error is returned

---

## Epic 3: Workspace Initialization (`ws init`)

Users can run `zgard ws init` to create their full workspace directory structure and clone all configured repositories. The command supports targeting (default/named/all), dry-run preview, real-time progress output, operation summary, parallel execution, and continues past individual clone failures without aborting.

### Story 3.1: Workspace Targeting Resolver

As a developer using `zgard`,
I want targeting logic to resolve which workspace(s) to operate on from flags,
So that `ws init`, `ws pull`, and `ws purge` all use a consistent, tested targeting function.

**Acceptance Criteria:**

**Given** no `--name` or `--all` flag and a config with a `default` workspace
**When** `targeting.Resolve(cfg, "", false)` is called
**Then** it returns `[]Workspace` containing only the default workspace

**Given** `--name myws` and a config containing a workspace named `myws`
**When** `targeting.Resolve(cfg, "myws", false)` is called
**Then** it returns `[]Workspace` containing only `myws`

**Given** `--name nonexistent`
**When** `targeting.Resolve(cfg, "nonexistent", false)` is called
**Then** it returns an error: `"workspace 'nonexistent' not found in config"`

**Given** `--all` flag
**When** `targeting.Resolve(cfg, "", true)` is called
**Then** it returns all workspaces from the config

**Given** both `--name` and `--all` are provided
**When** `targeting.Resolve` is called
**Then** it returns an error: `"--name and --all are mutually exclusive"`

**Given** a unit test using `valid_multi_workspace.yaml` fixture
**When** `targeting.Resolve` is called with various flag combinations
**Then** all cases above are covered without filesystem access

### Story 3.2: Workspace and Project Directory Creation

As a developer using `zgard`,
I want `ws init` to create the workspace root and all project subdirectories before cloning,
So that the directory hierarchy matches my config topology.

**Acceptance Criteria:**

**Given** a workspace with `path: ~/ws/myws` and two projects (`backend`, `frontend`)
**When** `workspace.Init` is called with `RunOptions{DryRun: false}`
**Then** `~/ws/myws`, `~/ws/myws/backend`, and `~/ws/myws/frontend` are all created

**Given** the directories already exist
**When** `workspace.Init` is called again
**Then** `os.MkdirAll` is called (idempotent) and no error is returned for existing dirs

**Given** `RunOptions{DryRun: true}`
**When** `workspace.Init` is called
**Then** no directories are created; instead `log.Info("[DRY RUN] would create directory", "path", ...)` is emitted for each

**Given** a path construction
**When** any directory path is built
**Then** `filepath.Join` is always used — no string concatenation (NFR10)

**Given** a `MockExecutor` and a workspace with one project
**When** `workspace.Init` is tested
**Then** the test verifies directory creation calls without touching the real filesystem

### Story 3.3: Clone Command Rendering and Execution

As a developer using `zgard`,
I want each repository cloned using its workspace's rendered `clone_command_template`,
So that all repos are checked out to the correct directory on the correct branch.

**Acceptance Criteria:**

**Given** a repository with `name: api`, `branch: main` and no `local_dir_name`
**When** the clone command template `git clone --branch {{.Branch}} https://github.com/org/{{.RepoName}} {{.DestDir}}` is rendered
**Then** the rendered command is `git clone --branch main https://github.com/org/api <project_path>/api`

**Given** a repository with `local_dir_name: api-service`
**When** the clone command template is rendered
**Then** `{{.DestDir}}` is replaced with `api-service`, not `api`

**Given** `RunOptions{DryRun: false}` and a valid rendered command
**When** `executor.Run(projectPath, renderedCmd)` is called
**Then** `OsExecutor` executes `sh -c <renderedCmd>` with `cmd.Dir = projectPath`

**Given** `RunOptions{Verbose: true}`
**When** the clone command is about to execute
**Then** `  → <rendered command>` is printed to stdout before execution (UX-DR9)

**Given** `RunOptions{DryRun: true}`
**When** the clone would execute
**Then** `log.Info("[DRY RUN] would clone", "repo", name, "cmd", renderedCmd)` is emitted and no subprocess is spawned

**Given** a `MockExecutor` in tests
**When** `workspace.Init` is called with a two-repo workspace
**Then** `mock.Calls` contains exactly two rendered command strings in the expected order

### Story 3.4: Idempotent Skip for Existing Repositories

As a developer using `zgard`,
I want `ws init` to skip cloning when a repo directory already exists,
So that re-running init on an existing workspace is safe and only clones what's missing.

**Acceptance Criteria:**

**Given** a repository whose expected local directory already exists
**When** `workspace.Init` processes that repository
**Then** no clone command is executed; a `⏭` skip result is recorded via `reporter.Record(OpResult{Skipped: true})`

**Given** a workspace with 3 repos where 2 already exist
**When** `workspace.Init` completes
**Then** the summary shows `✓ 1 cloned  ⏭ 2 skipped  ✗ 0 failed`

**Given** `RunOptions{DryRun: true}` and a repo that already exists
**When** `workspace.Init` processes that repository
**Then** `[DRY RUN] ⏭ <name> — already exists, would skip` is emitted

**Given** a test with a temp directory containing a pre-existing repo path
**When** `workspace.Init` is called
**Then** the `MockExecutor` records zero calls for that repo

### Story 3.5: Continue-on-Failure with Atomic Cleanup

As a developer using `zgard`,
I want clone failures to be recorded but not abort the run, and failed partial clones to be cleaned up,
So that I get a full picture of what failed and there is no partial state left behind.

**Acceptance Criteria:**

**Given** a workspace with 3 repos and the second clone fails
**When** `workspace.Init` completes
**Then** the first and third repos are cloned (or attempted); only the second is recorded as `✗ failed`

**Given** a clone fails after partially creating a directory
**When** the failure is recorded
**Then** the partially created directory is removed via `os.RemoveAll` (defer cleanup pattern, NFR5)

**Given** a clone succeeds (directory is fully populated)
**When** the defer runs
**Then** the directory is NOT removed (success flag prevents cleanup)

**Given** a failed repo
**When** the error is reported
**Then** `reporter.Record(OpResult{Repo: "name", Err: err})` captures workspace, project, and repo name along with the error

**Given** tests using `MockExecutor` configured to return an error for a specific call
**When** `workspace.Init` is tested
**Then** cleanup is verified and subsequent repos are still processed

### Story 3.6: Real-Time Progress Output and Run Summary (Reporter)

As a developer using `zgard`,
I want real-time status lines for each operation and a summary at the end,
So that I can see exactly what succeeded, was skipped, or failed, in a consistent structured format.

**Acceptance Criteria:**

**Given** a workspace with 2 projects and 3 repos
**When** `workspace.Init` runs
**Then** stdout contains workspace and project section headers followed by per-repo status lines following the UX-DR2 hierarchy:
```
Workspace: default
  Project: backend
    ✓ api          — cloned (main)
    ✗ auth-service — exit 128: ...
  Project: frontend
    ✓ webapp       — cloned (dev)
```

**Given** the run completes
**When** `reporter.Summary()` is called
**Then** the last stdout output is a blank line followed by `✓ 2 cloned  ⏭ 0 skipped  ✗ 1 failed` (UX-DR4)

**Given** failures occurred
**When** `reporter.Summary()` is called
**Then** stderr receives error detail lines: `      auth-service: exit 128: repository not found` (UX-DR8)

**Given** stdout is piped (non-TTY)
**When** `reporter` writes output
**Then** colour ANSI codes are absent; symbols (`✓`, `✗`, `⏭`) are still present (UX-DR6)

**Given** `RunOptions{Verbose: true}`
**When** an operation runs
**Then** the rendered command is printed to stdout before the status line (UX-DR9)

**Given** a unit test
**When** `reporter` is tested with captured stdout/stderr writers
**Then** output format matches the UX spec exactly — symbols, indentation, summary format

### Story 3.7: Wire `ws init` Command with All Flags and Exit Codes

As a developer using `zgard`,
I want `zgard ws init` to be fully wired with all flags, correct exit codes, and stdout/stderr routing,
So that the command is usable both interactively and in CI pipelines.

**Acceptance Criteria:**

**Given** `zgard ws init` with no flags and a `default` workspace
**When** the command runs successfully
**Then** it exits with code `0` (FR32)

**Given** any repo clone failure
**When** `zgard ws init` completes
**Then** it exits with a non-zero code (FR33)

**Given** `--dry-run` flag
**When** `zgard ws init --dry-run` runs
**Then** every output line is prefixed with `[DRY RUN]`, no directories are created, and exit code is `0` (FR19, FR31)

**Given** `--name myws` flag
**When** `zgard ws init --name myws`
**Then** only the `myws` workspace is initialised (FR10)

**Given** `--all` flag
**When** `zgard ws init --all`
**Then** all workspaces in config are initialised sequentially (FR11)

**Given** a config validation error
**When** `zgard ws init` runs
**Then** the error is printed to stderr, zero filesystem changes occur, and exit code is non-zero (NFR6)

**Given** `--no-confirm` flag
**When** `zgard ws init --no-confirm`
**Then** any confirmation prompts are suppressed (FR34) — note: `ws init` has no confirmation prompts, but the flag must be accepted without error

**Given** `$GRAZHDA_DIR` is unset
**When** `zgard ws init` runs
**Then** config is loaded from `$HOME/.grazhda/config.yaml`

### Story 3.8: Parallel Repository Operations for `ws init`

As a developer using `zgard`,
I want to run `zgard ws init --parallel` to clone repositories concurrently,
So that initialising large workspaces is faster when network bandwidth allows.

**Acceptance Criteria:**

**Given** `--parallel` flag and a workspace with 4 repos
**When** `workspace.Init` runs
**Then** all 4 clone commands are dispatched concurrently via goroutines + `sync.WaitGroup`

**Given** parallel execution
**When** `reporter.Record` is called from multiple goroutines
**Then** a mutex protects the reporter; no data races occur (verified via `go test -race`)

**Given** a repo failure during parallel execution
**When** the WaitGroup completes
**Then** failed repos are recorded and cleanup occurs for each; other repos are not affected

**Given** no `--parallel` flag
**When** `workspace.Init` runs
**Then** repos are processed sequentially (default, NFR3)

**Given** `--parallel` flag
**When** `zgard ws init --parallel --dry-run`
**Then** dry-run output is still produced for all repos; no actual concurrency needed for dry-run (implementation may dispatch serially in dry-run)

---

## Epic 4: Workspace Teardown (`ws purge`)

Users can safely remove a targeted workspace's directory structure. `ws purge` always requires explicit targeting, prompts for confirmation before any deletion, supports dry-run preview, and supports non-interactive mode for CI.

> **Implementation prerequisite:** Stories 3.1 (targeting resolver — `internal/targeting/`) and 3.6 (reporter — `internal/reporter/`) must be complete before implementing Epic 4 stories.

### Story 4.1: Purge Targeting — Explicit Flag Required

As a developer using `zgard`,
I want `ws purge` to require an explicit `--name <name>` or `--all` flag,
So that I can never accidentally purge a workspace by running the command without thinking.

**Acceptance Criteria:**

**Given** `zgard ws purge` with no flags
**When** the command runs
**Then** it exits non-zero and prints to stderr: `"ws purge requires --name <name> or --all"` (FR12)

**Given** `--name myws`
**When** `zgard ws purge --name myws`
**Then** targeting resolves to only `myws`

**Given** `--all`
**When** `zgard ws purge --all`
**Then** targeting resolves to all workspaces in config

**Given** a non-existent workspace name `--name ghost`
**When** `zgard ws purge --name ghost`
**Then** it exits non-zero: `"workspace 'ghost' not found in config"`

**Given** `--parallel` flag
**When** `zgard ws purge --parallel`
**Then** the command rejects the flag with a helpful error: `"--parallel is not supported for ws purge"` (or the flag is simply not registered on this subcommand)

### Story 4.2: Directory Removal

As a developer using `zgard`,
I want `ws purge` to remove the workspace directory tree after confirmation,
So that I can clean up a workspace and all its contents in one command.

**Acceptance Criteria:**

**Given** a workspace at `~/ws/myws` and `--no-confirm` flag (to skip prompt in test)
**When** `workspace.Purge` is called
**Then** `~/ws/myws` and all its contents are removed via `os.RemoveAll`

**Given** the workspace directory does not exist
**When** `workspace.Purge` is called
**Then** it exits with `0` and prints: `⏭ myws — directory not found, skipped`

**Given** removal succeeds
**When** `reporter.Summary()` is called
**Then** output shows `✓ 1 removed`

**Given** removal fails (e.g. permission error)
**When** `workspace.Purge` encounters the error
**Then** the error is recorded; exit code is non-zero; the error is printed to stderr with workspace context

### Story 4.3: Interactive Confirmation Prompt

As a developer using `zgard`,
I want `ws purge` to show me every path it will delete and ask for confirmation,
So that I can never accidentally remove directories I intended to keep.

**Acceptance Criteria:**

**Given** `zgard ws purge --name myws` (no `--no-confirm`)
**When** the command runs
**Then** stdout displays the confirmation prompt listing each directory to be removed, followed by `Confirm? [y/N]:` (UX-DR7)

**Given** the user types `y` at the prompt
**When** confirmation is read
**Then** `confirm()` returns `true` and removal proceeds

**Given** the user types anything other than `y` or `Y` (including pressing Enter)
**When** confirmation is read
**Then** `confirm()` returns `false`, removal is aborted, and exit code is `0`

**Given** `--no-confirm` flag
**When** `workspace.Purge` is called
**Then** the prompt is skipped entirely and removal proceeds (FR34)

**Given** a unit test
**When** `confirm()` is called with a `strings.NewReader("y\n")`
**Then** it returns `true` — the `io.Reader` injection enables testing without a real TTY

### Story 4.4: Dry-Run for `ws purge`

As a developer using `zgard`,
I want to preview what `ws purge` would remove without actually deleting anything,
So that I can verify my targeting before committing to a destructive operation.

**Acceptance Criteria:**

**Given** `zgard ws purge --name myws --dry-run`
**When** the command runs
**Then** every path that would be removed is printed with `[DRY RUN]` prefix; no directories are deleted (FR22)

**Given** dry-run mode
**When** output is produced
**Then** format is structurally identical to a real run with `[DRY RUN]` prefix on each line (UX-DR3)

**Given** dry-run mode
**When** `reporter.Summary()` is called
**Then** summary shows `[DRY RUN] ✓ 1 would remove` and exit code is `0`

**Given** dry-run mode
**When** the confirmation prompt would normally appear
**Then** it is skipped (no confirmation needed for a preview-only operation)

---

## Epic 5: Repository Synchronization (`ws pull`)

Users can run `zgard ws pull` to update all repositories in a targeted workspace to their configured branches. The command supports targeting, skips missing repos gracefully, supports dry-run preview, and supports parallel execution.

> **Implementation prerequisite:** Stories 3.1 (targeting resolver — `internal/targeting/`) and 3.6 (reporter — `internal/reporter/`) must be complete before implementing Epic 5 stories.

### Story 5.1: Execute `git pull --rebase` for Each Repository

As a developer using `zgard`,
I want `ws pull` to run `git pull --rebase` on each repo's configured branch,
So that all repos in my workspace are kept in sync with their upstream.

**Acceptance Criteria:**

**Given** a workspace with repo `api` on branch `main` at `~/ws/myws/backend/api`
**When** `workspace.Pull` is called with `RunOptions{DryRun: false}`
**Then** `executor.Run("~/ws/myws/backend/api", "git pull --rebase origin main")` is called

**Given** a successful pull
**When** `reporter.Record` is called
**Then** `OpResult{Repo: "api", Skipped: false, Err: nil}` is recorded and `✓ api — pulled (main)` is printed

**Given** a pull failure (non-zero exit from git)
**When** `workspace.Pull` processes that repo
**Then** the failure is recorded with error context; subsequent repos continue to be processed (continue-on-failure, matching Epic 3 Story 3.5 behaviour)

**Given** `RunOptions{Verbose: true}`
**When** a pull executes
**Then** `  → git pull --rebase origin main` is printed to stdout before execution

### Story 5.2: Skip Repositories with No Local Directory

As a developer using `zgard`,
I want `ws pull` to silently skip repos that haven't been cloned yet,
So that pulling a partially-initialised workspace doesn't produce errors for repos I haven't set up.

**Acceptance Criteria:**

**Given** a repo whose local directory does not exist
**When** `workspace.Pull` processes it
**Then** it is recorded as skipped: `OpResult{Skipped: true}`; `⏭ api — not present, skipped` is printed (FR25)

**Given** a workspace with 3 repos where 1 is missing
**When** `workspace.Pull` completes
**Then** summary shows `✓ 2 pulled  ⏭ 1 skipped  ✗ 0 failed` and exit code is `0`

**Given** a unit test
**When** `workspace.Pull` is called with a workspace where one repo path doesn't exist
**Then** `MockExecutor.Calls` does not contain a pull command for the missing repo

### Story 5.3: Dry-Run for `ws pull`

As a developer using `zgard`,
I want to preview which repositories `ws pull` would update before running it,
So that I can verify targeting and confirm the expected branches.

**Acceptance Criteria:**

**Given** `zgard ws pull --dry-run`
**When** the command runs
**Then** every repo is printed with `[DRY RUN] ⏭ <name> — would pull (<branch>)` and no git commands are executed (FR26)

**Given** dry-run mode and a repo that doesn't exist locally
**When** `workspace.Pull` processes it
**Then** it prints `[DRY RUN] ⏭ <name> — not present, would skip`

**Given** dry-run output
**When** compared to a real run
**Then** the structure is identical — only `[DRY RUN]` prefix and "would pull" vs "pulled" verb differ (UX-DR3)

**Given** dry-run completes
**When** `reporter.Summary()` is called
**Then** summary shows `[DRY RUN] ✓ N would pull  ⏭ N skipped  ✗ 0 failed` and exit code is `0`

### Story 5.4: Wire `ws pull` Command with All Flags Including `--parallel`

As a developer using `zgard`,
I want `zgard ws pull` fully wired with all supported flags,
So that it behaves consistently with `ws init` for targeting, output, exit codes, and parallel execution.

**Acceptance Criteria:**

**Given** `zgard ws pull` with no flags and a `default` workspace
**When** the command runs successfully
**Then** it exits with code `0`

**Given** `--name myws` or `--all`
**When** `zgard ws pull --name myws` or `zgard ws pull --all`
**Then** targeting resolves correctly (same resolver as `ws init`)

**Given** `--parallel` flag
**When** `zgard ws pull --parallel`
**Then** all repos in the targeted workspace(s) are pulled concurrently via `sync.WaitGroup` (same parallel pattern as `ws init` Story 3.8)

**Given** any pull failure
**When** `zgard ws pull` completes
**Then** it exits non-zero and failure details appear on stderr

**Given** `--verbose` flag
**When** `zgard ws pull --verbose`
**Then** rendered pull commands are printed before each execution

**Given** `$GRAZHDA_DIR` is unset
**When** `zgard ws pull` runs
**Then** config is loaded from `$HOME/.grazhda/config.yaml`

---

## Epic G1 — grazhda upgrade

**Goal:** Enable users to update their Grazhda installation to the latest version with a single command.

**Acceptance Criteria:**
- `grazhda upgrade` succeeds on a freshly installed machine with working internet access.
- `grazhda upgrade` prints step-by-step progress and a success banner.
- `grazhda upgrade` exits non-zero with a clear error on any failure.
- Running `grazhda upgrade` twice produces no errors on the second run.

### Story G1.1 — Implement `grazhda upgrade` command

**As a** developer
**I want to** run `grazhda upgrade`
**So that** my local Grazhda installation is updated to the latest sources and binaries.

**Acceptance Criteria:**

**Given** `$GRAZHDA_DIR/sources` is a valid git repository
**When** `grazhda upgrade` runs
**Then** it pulls the latest sources, rebuilds all binaries with `just build`, and copies them to `$GRAZHDA_DIR/bin/`

**Given** the sources are already up to date
**When** `grazhda upgrade` runs
**Then** it prints "✓ Sources are already up to date." and proceeds with the rebuild

**Given** a missing build dependency (git, go, just, or protoc)
**When** `grazhda upgrade` runs
**Then** it exits with code 1 and names the missing binary

**Given** `$GRAZHDA_DIR/sources` is not a git repository or does not exist
**When** `grazhda upgrade` runs
**Then** it exits with code 1 and advises the user to run the installer

**Given** `just build` fails
**When** `grazhda upgrade` runs
**Then** it exits with code 1 and the build error is visible in the terminal

---

## Epic G2 — grazhda config management

**Goal:** Allow users to open and edit their Grazhda config file without having to remember its location.

**Acceptance Criteria:**
- `grazhda config --edit` opens `config.yaml` in the configured or fallback editor.
- Editor resolution follows the documented precedence chain.
- A missing config file or missing editor produces a clear error.

### Story G2.1 — Add `editor:` field to config template

**As a** developer
**I want** the `config.yaml` to include an `editor:` field
**So that** `grazhda config --edit` uses my preferred editor.

**Acceptance Criteria:**

**Given** a fresh installation
**When** `config.yaml` is created from `config.template.yaml`
**Then** it contains an `editor: vim` field with a comment explaining the resolution order

### Story G2.2 — Implement `grazhda config --edit` command

**As a** developer
**I want to** run `grazhda config --edit`
**So that** my config file opens in my preferred editor immediately.

**Acceptance Criteria:**

**Given** `config.yaml` has `editor: vim`
**When** `grazhda config --edit` runs
**Then** vim opens `$GRAZHDA_DIR/config.yaml`

**Given** `editor:` is not set in `config.yaml`
**When** `grazhda config --edit` runs
**Then** the editor is resolved from `$VISUAL`, then `$EDITOR`, then `vi`

**Given** the resolved editor is not in `$PATH`
**When** `grazhda config --edit` runs
**Then** it exits with code 1 and suggests updating the `editor:` field

**Given** `config.yaml` does not exist
**When** `grazhda config --edit` runs
**Then** it exits with code 1 and advises the user to run the installer

---

## Epic X — Cross-Repository Operations

**Goal:** Extend `zgard` with fan-out commands that execute actions across all repositories in a workspace in a single invocation: `ws exec` (arbitrary shell), `ws stash`, and `ws checkout`.

**Requirements covered:** FR-X1 through FR-X13

**Deliverables:** `ws exec`, `ws stash`, `ws checkout` with `--project-name`/`--repo-name` filtering and parallel execution support.

---

### Story X1 — Extend RunOptions and Executor with filtering and capture

**As a** workspace domain library consumer
**I want** RunOptions to carry `ProjectName`/`RepoName` filter fields and Executor to support stdout capture
**So that** new fan-out operations can filter targets and surface per-repo output.

**Acceptance Criteria:**

**Given** `RunOptions{ProjectName: "backend"}`
**When** passed to any new workspace function
**Then** only repos within the `backend` project are processed

**Given** `RunOptions{RepoName: "api"}` without `ProjectName`
**When** validated at the CLI layer
**Then** the command exits with error `--repo-name requires --project-name`

**Given** `MockExecutor{CaptureOutput: "hello\n"}`
**When** `RunCapture` is called
**Then** it returns `("hello\n", nil)` and records the call

---

### Story X2 — Implement `workspace.Exec`

**As a** developer
**I want** `zgard ws exec <command>` to run a shell command in every repository directory
**So that** I can fan out builds, tests, or any script across my workspace in one step.

**Acceptance Criteria:**

**Given** two repos present on disk
**When** `zgard ws exec "echo hi"`
**Then** the command runs in both repos and their output appears indented under each repo's status line

**Given** one repo is absent from disk
**When** `zgard ws exec "make test"`
**Then** that repo is skipped (⏭) and the command continues in the remaining repos

**Given** one repo's command exits non-zero
**When** `zgard ws exec` runs
**Then** that repo is recorded as failed (✗) and all remaining repos still execute

**Given** `--dry-run`
**When** `zgard ws exec "make test"`
**Then** no commands are executed and `[DRY RUN] would exec: make test` appears per repo

---

### Story X3 — Implement `workspace.Stash`

**As a** developer
**I want** `zgard ws stash` to stash local changes in every repository
**So that** I can cleanly switch branches across the whole workspace.

**Acceptance Criteria:**

**Given** two repos present on disk
**When** `zgard ws stash`
**Then** `git stash push` runs in both and status shows `✓ api — stashed`

**Given** a repo fails to stash (e.g. git error)
**When** `zgard ws stash`
**Then** the failure is recorded and remaining repos are still processed

**Given** `--dry-run`
**When** `zgard ws stash`
**Then** no commands execute and `[DRY RUN] would stash` appears per present repo

---

### Story X4 — Implement `workspace.Checkout`

**As a** developer
**I want** `zgard ws checkout <branch>` to check out a branch in every repository
**So that** I can switch my entire workspace to a feature branch atomically.

**Acceptance Criteria:**

**Given** two repos present on disk
**When** `zgard ws checkout feature-x`
**Then** `git checkout feature-x` runs in both; status shows `✓ api — checked out feature-x`

**Given** a repo does not have the branch
**When** `zgard ws checkout feature-x`
**Then** that repo is recorded as failed (✗) and checkout continues in other repos

**Given** `--dry-run`
**When** `zgard ws checkout main`
**Then** no commands execute and `[DRY RUN] would checkout: main` appears per present repo

---

### Story X5 — Register CLI commands and wire filtering flags

**As a** CLI user
**I want** `zgard ws exec`, `zgard ws stash`, `zgard ws checkout` to appear as subcommands
**So that** I can invoke them from the terminal with all supported flags.

**Acceptance Criteria:**

**Given** `zgard ws --help`
**When** run
**Then** exec, stash, checkout appear in the subcommand list

**Given** `zgard ws exec --project-name backend "make test"`
**When** run
**Then** only repos in the `backend` project are processed

**Given** `zgard ws checkout --repo-name api --project-name backend main`
**When** run
**Then** only `api` in `backend` is targeted

**Given** `zgard ws stash --all`
**When** run
**Then** all workspaces are processed

---

## Epic Y — Universal Targeting System

**Goal:** Unify all workspace/project/repo targeting flags as persistent Cobra flags on the `ws` parent command, add a default-workspace warning, and extend project/repo filtering to `init` and `pull`.

**Requirements covered:** FR-T1 through FR-T12

---

### Story Y1 — Persistent targeting flags on `ws` parent command

**As a** CLI user
**I want** `--name`, `--all`, `--project-name`, and `--repo-name` to be available on every `zgard ws` subcommand
**So that** I don't need to remember which commands accept which targeting flags.

**Acceptance Criteria:**

**Given** any `zgard ws` subcommand
**When** run with `--name myws`
**Then** it targets workspace "myws"

**Given** `zgard ws init --project-name backend`
**When** run
**Then** only repos in the `backend` project are cloned

**Given** `zgard ws pull --repo-name api --project-name backend`
**When** run
**Then** only repo `api` in `backend` is pulled

**Given** `zgard ws exec --repo-name api` (without `--project-name`)
**When** run
**Then** command exits with error `--repo-name requires --project-name`

---

### Story Y2 — Default workspace warning

**As a** CLI user who didn't specify a targeting flag
**I want** to see a yellow warning before output
**So that** I know which workspace is being targeted and can verify it before operations proceed.

**Acceptance Criteria:**

**Given** `zgard ws init` with no targeting flags
**When** run
**Then** stderr contains `Warning: Targeting default workspace: <absolute-path>`

**Given** `zgard ws init --name default`
**When** run
**Then** no warning is printed

**Given** `zgard ws init --all`
**When** run
**Then** no warning is printed

---

### Story Y3 — Shared loadConfig helper

**As a** contributor
**I want** config loading to be a single shared helper
**So that** each command file isn't burdened with 8 lines of identical boilerplate.

**Acceptance Criteria:**

**Given** any `zgard ws` command
**When** `config.yaml` is missing or invalid
**Then** the error message is identical regardless of which command triggered it

---

### Story Y4 — Extend `ws init` and `ws pull` with project/repo filtering

**As a** developer
**I want** `zgard ws init --project-name backend` to clone only the backend project
**So that** I can selectively initialise part of my workspace.

**Acceptance Criteria:**

**Given** a workspace with two projects
**When** `zgard ws init --project-name backend`
**Then** only the `backend` project directory is created and its repos are cloned

**Given** `zgard ws init --project-name nonexistent`
**When** run
**Then** exits with error `project "nonexistent" not found in workspace "..."`

**Given** `zgard ws pull --project-name backend --repo-name api`
**When** run
**Then** only `api` in `backend` is pulled

## Epic Z — Workspace Inspection Suite

### Story Z1 — ws search: content and filename search

Implement `zgard ws search <pattern>` with streaming line-by-line content grep (`bufio.Scanner`) and filename glob mode (`--glob`). Binary files (detected via null-byte scan of first 512 bytes) are silently skipped. `.git` directories are excluded from traversal. Output follows the `[project/repo] filepath:lineno: matched line` format (content) or `[project/repo] filepath` (glob). A summary line `N match(es) across M repo(s)` is always printed. Parallel scanning is supported via `--parallel` / `--parallel`.

**Acceptance criteria:**
- Content search finds matching lines in text files across all resolved repos.
- `--glob` mode finds matching filenames without reading file contents.
- Binary files are silently skipped; no error is emitted.
- `.git` directories are never traversed.
- `--parallel` / `--parallel` flags produce identical results to sequential mode.
- Multi-repo `--repo-name` match emits a yellow warning before results.

### Story Z2 — ws diff: per-repo Git state table

Implement `zgard ws diff` that renders an aligned, colour-coded, project-grouped table showing UNCOMMITTED, AHEAD, and BEHIND counts for each repo. Uncommitted count is derived from `git status --porcelain` line count; ahead/behind from `git rev-list @{u}..HEAD` and `git rev-list HEAD..@{u}`. Repos with no upstream show `--` in those columns. Repos absent from disk render a `(not cloned)` row. A summary line listing dirty/clean/not-cloned counts is printed after all projects.

**Acceptance criteria:**
- UNCOMMITTED column reflects `git status --porcelain` line count.
- AHEAD/BEHIND columns reflect `rev-list` counts; `--` when no upstream tracking branch.
- Missing repos render `(not cloned)` without aborting the run.
- Colour coding: UNCOMMITTED > 0 red; AHEAD/BEHIND > 0 yellow; all-zero clean rows green.
- Summary counts (dirty / clean / not-cloned) printed at end.

### Story Z3 — ws stats: repository metadata table

Implement `zgard ws stats` that renders a project-grouped table with LAST COMMIT (YYYY-MM-DD HH:MM), 30D COMMITS, and CONTRIBUTORS per repo. Last commit timestamp comes from `git log -1 --format="%ci"` truncated to the minute. 30-day commit count comes from `git log --since="30 days ago"`. Contributor count is the number of unique author emails in the full `git log`. Repos absent from disk show a `(not cloned)` row with `-` in all value columns. Parallel execution is supported.

**Acceptance criteria:**
- LAST COMMIT renders as `YYYY-MM-DD HH:MM` from `git log -1 --format="%ci"`.
- 30D COMMITS count derived from `--since="30 days ago"`.
- CONTRIBUTORS = count of unique author emails across full history.
- Missing repos render `(not cloned)` with `-` values without aborting.
- `--parallel` / `--parallel` flags work correctly.

### Story Z4 — Universal targeting integration

Ensure all three commands (`search`, `diff`, `stats`) fully respect the universal targeting flags: `-n`/`--name`, `--all`, `-p`/`--project-name`, `-r`/`--repo-name`. `ValidateFilters` must be called before any git or filesystem operations. The default workspace info message must be shown when no explicit target is provided. `--repo-name` with a partial substring match must emit a yellow multi-match warning. `--all` must not be combinable with `-p`/`-r`.

**Acceptance criteria:**
- `--repo-name` with a name matching multiple repos emits yellow warning and continues.
- `--all` combined with `-p` or `-r` returns a validation error.
- Default workspace info message is shown when no target flags are given.
- Purge safety behaviour is unaffected by these new commands.
- `ValidateFilters` contract (red error on unknown project/repo) is enforced for all three commands.

### Story Z5 — README documentation update

Update `README.md` to document `zgard ws search`, `zgard ws diff`, and `zgard ws stats` under the workspace commands section. Include usage synopsis, flag descriptions, and trimmed example output (matching the UX spec) for each command so that new users can understand the feature without reading internal docs.

## Epic AA — Tag-Based Targeting & IDE Integration

### Story AA1 — Config schema: tags field

Add `tags: []string` to the `Project` and `Repository` YAML structs in `internal/config/config.go`. Update `config.template.yaml` with annotated examples showing tags at both levels.

**Acceptance criteria:**
- Tags load correctly from YAML for both projects and repositories.
- Existing configs without `tags` fields continue to work without change.
- `Validate` behaviour is unchanged.

### Story AA2 — Tag filtering engine

Implement `effectiveTags`, `repoTagsMatch`, and `repoMatchesFilters` helpers in `internal/workspace/targeting.go`. Extend `RunOptions` and `InspectOptions` with a `Tags []string` field. Update `ValidateFilters` and `CountMatchingRepos` to apply tag filtering. Apply `repoMatchesFilters` in all iteration loops: `runOverRepos` (exec/stash/checkout), `Init`, `Pull`, and all three inspect loops (Search, Diff, Stats jobs).

**Acceptance criteria:**
- `--tag` filters repos correctly with OR logic and project-level tag inheritance.
- Zero-match tag filter produces a red error and non-zero exit.
- Combining `--tag` with `-p`/`-r` applies AND logic correctly.
- All existing commands respect `--tag` without behavioural regression.

### Story AA4 — README and config.template.yaml updates

Update `README.md` to document the `--tag` flag in the common flags table. Update `config.template.yaml` to demonstrate `tags` at both project and repository levels.

**Acceptance criteria:**
- README shows `--tag` in the common flags table.
- `config.template.yaml` demonstrates tags at both project and repository levels.
