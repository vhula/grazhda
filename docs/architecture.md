---
stepsCompleted: [step-01-init, step-02-context, step-03-starter, step-04-decisions, step-05-patterns, step-06-structure, step-07-validation, step-08-complete]
workflowStatus: complete
completedAt: '2026-04-05T11:50:23Z'
inputDocuments: [docs/prd.md, README.md, config.template.yaml, grazhda.sh, grazhda-init.sh, grazhda, Justfile]
workflowType: 'architecture'
project_name: 'Grazhda — zgard CLI'
user_name: 'jake'
date: '2026-04-05'
---

# Architecture Decision Document

> **⚠️ Implementation Note:** This document is the original planning-phase architecture. The final implementation diverged in several areas:
>
> | Plan | Implementation |
> | :--- | :--- |
> | Three modules: `cmd/`, `internal/`, `zgard/` | Two modules: `internal/`, `zgard/` — `cmd/` was eliminated; its code lives in `zgard/ws/` |
> | `charmbracelet/log` for output | `fatih/color` for coloured terminal output; no logging library |
> | `internal/targeting/` as its own package | `internal/workspace/targeting.go` (file within `workspace` package) |
> | `internal/reporter/` described in rough terms | Fully implemented with `Record`, `Summary`, `ExitCode`, `PrintLine` |
> | `internal/executor/` with `Executor` interface | Implemented as described; `OsExecutor` captures stderr for rich error messages |
> | No colour output mentioned | `internal/color/` package added wrapping `fatih/color` |
>
> **For the current accurate source layout see [README.md](../README.md) and [STUDY.md](../STUDY.md).**

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements (36 total across 7 areas):**
- Configuration Management (FR1–8): Load, parse, and validate `$GRAZHDA_DIR/config.yaml` up-front; enforce uniqueness, required fields, and valid Go templates before any filesystem work.
- Workspace Targeting (FR9–12): Resolve default/named/all workspaces; purge requires explicit targeting.
- Workspace Initialization (FR13–19): Create directory hierarchy; render and execute clone command templates; skip-if-exists idempotency; continue-on-failure.
- Workspace Teardown (FR20–22): Remove workspace directory with confirmation and dry-run.
- Repository Synchronization (FR23–26): `git pull --rebase` per repo; skip missing repos silently; dry-run support.
- Operation Feedback & Reporting (FR27–31): Real-time per-op progress; end-of-run summary; `[DRY RUN]` prefix; verbose mode.
- Automation & Scripting (FR32–36): Exit codes 0/non-zero; stderr/stdout split; `--no-confirm`; `--parallel`.

**Non-Functional Requirements (14 total across 4 areas):**
- Performance: Config validation <500ms; dir creation <1s; sequential default; `--parallel` opt-in.
- Reliability: No partial artifacts on failure (NFR5); zero filesystem changes on validation failure (NFR6); dry-run accuracy (NFR7); no silent errors (NFR8).
- Portability: Linux amd64 + macOS amd64/arm64; `filepath` package only; no runtime deps beyond `git` in `$PATH` (NFR9–11).
- Testability: Domain logic separate from CLI entry (NFR12); injectable executor interface (NFR13); config validation testable with fixture YAML (NFR14).

**Scale & Complexity:**
- Primary domain: CLI / local filesystem + subprocess execution
- Complexity level: Medium
- Estimated architectural components: 3 Go modules (`zgard`, `cmd`, `internal`), 5–7 packages (config, workspace, executor, reporter, targeting, optional parallel runner)

### Technical Constraints & Dependencies

- Language: Go; CLI framework: Cobra; logging: charmbracelet/log; YAML: gopkg.in/yaml.v3
- Binary output path: `$GRAZHDA_DIR/bin/zgard` (or `bin/zgard` during dev via Justfile)
- Config location: `$GRAZHDA_DIR/config.yaml` — no alternatives, no `--config` flag
- `zgard/`, `cmd/`, and `internal/` are each a separate Go module in a `go.work` monorepo workspace alongside `dukh/`
- Clone/pull operations: `os/exec` subprocess calls; `git` must be in `$PATH`
- No gRPC, no server process, no persistence layer for Phase 1

### Cross-Cutting Concerns Identified

- **Config pipeline**: load → parse → validate → resolve template — runs once at startup, shared by all commands
- **Workspace targeting resolver**: default/`--name`/`--all` logic shared across all three commands
- **Dry-run mode**: must thread through every operation without executing side effects
- **Output/logging**: charmbracelet/log with INFO/WARN/ERROR levels; `--verbose` flag; all output level-gated
- **Error accumulation**: continue-on-failure requires a result collector that aggregates per-repo outcomes for end-of-run summary and exit code
- **Executor interface**: clone and pull operations behind an interface for testability and future parallel fan-out

## Starter Template Evaluation

### Primary Technology Domain

Go CLI tool — no scaffolding generator applicable. Foundation established via three `go mod init` calls within the existing `go.work` monorepo: `zgard/`, `cmd/`, and `internal/` are each a separate Go module.

### Starter Options Considered

No canonical "starter template" exists for Go CLI tools analogous to web framework starters. The architectural foundation is established through module initialization and canonical Go project layout conventions.

### Selected Foundation: Three-module Go workspace layout

**Rationale:** `cmd/` and `internal/` are sibling modules alongside `zgard/` in the `go.work` workspace. This decouples the CLI entry point from domain logic and Cobra wiring, allows `dukh/` to import shared `internal/` packages in a future phase without depending on the binary module, and keeps each module's concerns strictly separated. NFR12 (domain logic separate from CLI entry) and NFR13 (injectable executor) are both fully satisfied.

**Go's `internal` directory restriction still applies:** because `internal/` is a direct child of the repository root, Go's toolchain allows only packages rooted at the repo parent (i.e., the monorepo itself) to import from it. External modules cannot. All modules in `go.work` (`zgard`, `cmd`, `dukh`) sit inside the repo root and are permitted importers.

**Initialization:**

```bash
# Initialize each module (only needed on first setup):
cd cmd      && go mod init github.com/vhula/grazhda/cmd
cd internal && go mod init github.com/vhula/grazhda/internal
cd zgard    && go mod init github.com/vhula/grazhda/zgard

# go.work already exists; add new modules:
go work use ./cmd ./internal ./zgard
```

**Canonical Package Layout:**

```
grazhda/                          # repo root (go.work)
│
├── cmd/                          # module: github.com/vhula/grazhda/cmd
│   ├── go.mod
│   └── ws/
│       ├── ws.go                 # `ws` parent command
│       ├── init.go               # `ws init` subcommand (flags → calls workspace pkg)
│       ├── purge.go              # `ws purge` subcommand
│       └── pull.go               # `ws pull` subcommand
│
├── internal/                     # module: github.com/vhula/grazhda/internal
│   ├── go.mod
│   ├── config/
│   │   ├── config.go             # Config structs + YAML loading
│   │   └── validate.go           # Up-front validation logic
│   ├── workspace/
│   │   ├── init.go               # ws init domain logic
│   │   ├── purge.go              # ws purge domain logic
│   │   └── pull.go               # ws pull domain logic
│   ├── executor/
│   │   ├── executor.go           # Executor interface + os/exec implementation
│   │   └── mock.go               # Mock executor for tests
│   ├── targeting/
│   │   └── resolver.go           # Default/--name/--all targeting logic
│   ├── reporter/
│   │   └── reporter.go           # Result accumulator + summary output
│   └── testdata/
│       └── *.yaml                # Config fixture files for unit tests
│
└── zgard/                        # module: github.com/vhula/grazhda/zgard
    ├── go.mod                    # requires: cmd, internal
    ├── go.sum
    └── main.go                   # Cobra root cmd setup only; no business logic
```

**Core Dependency Versions:**

| Dependency | Version | Module(s) |
|---|---|---|
| Go | 1.26.1 | all |
| github.com/spf13/cobra | v1.9.1 | `cmd`, `zgard` |
| gopkg.in/yaml.v3 | v3.0.1 | `internal` |
| charm.land/log/v2 | v2.x | `internal`, `cmd` |

**Architectural Decisions Established by Layout:**

- `zgard/main.go` is a thin entrypoint — Cobra wiring only, no business logic (NFR12)
- All domain logic lives in the `internal` module — satisfies NFR12
- `executor.Executor` interface defined in `internal/executor/` — satisfies NFR13
- Config fixtures in `internal/testdata/` — satisfies NFR14
- `internal/targeting/` resolver shared by all three commands — satisfies FR9–12
- `internal/reporter/` accumulates results across all operations for summary output
- `dukh/` may import from `internal/` in a future phase without touching `zgard/` or `cmd/`

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
- Executor interface shape — Option A (command-string, `sh -c`)
- Dry-run threading — `RunOptions` struct passed through call stack
- Result accumulator model — `OpResult` slice + `Reporter` struct
- Partial artifact cleanup — defer-based `os.RemoveAll` on failure

**Important Decisions (Shape Architecture):**
- Parallel execution — `sync.WaitGroup` + goroutines (uncapped, Phase 1)
- Distribution — source build only via Justfile + grazhda.sh installer

**Deferred Decisions (Post-MVP):**
- Worker pool with configurable concurrency (Phase 2 `--parallel` enhancement)
- Pre-built binary releases / GitHub Releases (Phase 2+)

### Execution & Command Model

- **Executor interface:** Command-string model — rendered Go template string passed to `sh -c` via `os/exec`. Interface: `Run(dir string, command string) error`.
- **Real implementation:** `OsExecutor` wraps `exec.Command("sh", "-c", command)` with `cmd.Dir = dir`.
- **Mock implementation:** `MockExecutor` records calls for assertion in tests.

### Operation Options (Cross-Cutting)

All workspace functions receive a `RunOptions` struct — no Cobra imports in `internal/`:

```go
type RunOptions struct {
    DryRun    bool
    Verbose   bool
    Parallel  bool
    NoConfirm bool
}
```

### Result Accumulation & Reporting

```go
type OpResult struct {
    Workspace string
    Project   string
    Repo      string
    Skipped   bool
    Err       error
}

type Reporter struct {
    log     *log.Logger   // charm.land/log/v2
    results []OpResult
}

func (r *Reporter) Record(res OpResult)
func (r *Reporter) Summary()      // end-of-run counts + failure list to stdout/stderr
func (r *Reporter) ExitCode() int // 0 = all success; 1 = any failure
```

### Reliability Patterns

- **Partial artifact cleanup:** defer-based `os.RemoveAll(destPath)` guarded by a `success` boolean — fires on any clone failure or panic.
- **Config-first guarantee:** validation returns all errors before any `os.Mkdir` call.
- **Dry-run accuracy:** workspace functions check `opts.DryRun` before every side-effecting call; log with `[DRY RUN]` prefix instead of executing.

### Parallel Execution

- Sequential by default; `opts.Parallel = true` enables concurrent fan-out.
- Implementation: `sync.WaitGroup` + goroutines, results collected via mutex-protected `Reporter`.
- Phase 1: uncapped concurrency (workspace sizes expected to be ≤ 50 repos).

### Infrastructure & Distribution

- **Dev build:** `just build-zgard` → `bin/zgard`
- **User install:** `grazhda.sh` clones repo, runs `just build`, copies to `$GRAZHDA_DIR/bin/`
- **Phase 1:** source-build only — no pre-built binaries or GitHub Releases

## Implementation Patterns & Consistency Rules

### Critical Conflict Points Identified

6 areas where AI agents could diverge without explicit rules: naming conventions, error handling, logging usage, config struct design, test structure, and dry-run implementation.

### Naming Patterns

**Go Identifier Conventions:**
- Exported types/functions: `PascalCase` — `RunOptions`, `OpResult`, `OsExecutor`
- Unexported identifiers: `camelCase` — `resolveTarget`, `buildDestPath`
- File names: `snake_case.go` — `config.go`, `validate.go`, `mock_executor.go`
- Test files: co-located `<file>_test.go` — `config_test.go`, `resolver_test.go`
- Constants: `PascalCase` for exported, `camelCase` for unexported — not `ALL_CAPS`
- Error variables: always named `err` for single errors; contextual names (`cloneErr`, `mkdirErr`) only when multiple errors in scope

**Package Naming:**
- All package names: lowercase singular — `config`, `workspace`, `executor`, `targeting`, `reporter`
- NO plural package names (`configs`, `workspaces`)
- Package name matches directory name

**CLI Flag Naming:**
- All flags: `kebab-case` — `--dry-run`, `--no-confirm`, `--name`
- Short flags: single letter where established — `-v` for verbose, `-n` for `--name`
- Flag variables in Cobra commands: `camelCase` — `dryRun`, `noConfirm`, `wsName`

### Structure Patterns

**Test Location:**
- ALL tests: co-located `_test.go` files alongside the package they test
- NO separate `tests/` directory
- Test fixtures: `internal/testdata/*.yaml` — loaded via `os.Open("testdata/...")` in tests

**Package Responsibility Boundaries:**
- `cmd/ws/*.go` — flag parsing, `RunOptions` construction, call to `internal/workspace`, exit code via `os.Exit(reporter.ExitCode())`. NO business logic.
- `internal/config/` — YAML load + validation only. NO filesystem ops.
- `internal/workspace/` — orchestration only. Calls executor + reporter. NO direct `os/exec` calls.
- `internal/executor/` — subprocess execution only. NO logging.
- `internal/targeting/` — workspace resolution only. Pure function, no side effects.
- `internal/reporter/` — logging + result accumulation only. NO filesystem ops.

### Error Handling Patterns

**Error message format:**
- Lowercase, no trailing period: `"workspace not found: default"`
- Wrap with context: `fmt.Errorf("operation: %w", err)`
- NO sentinel errors for internal use — return descriptive wrapped errors
- Validation errors: collect ALL errors into a `[]error` slice, return as joined multi-line message — never return after first validation failure (FR6)

**Error flow:**
```
executor.Run() → error
  ↓ workspace pkg wraps with context
  reporter.Record(OpResult{Err: err})  ← never propagated up; always recorded
  ↓
cmd layer: reporter.ExitCode() → os.Exit(code)
```
- Workspace functions NEVER return errors from individual repo operations — they record them and continue (FR18). Only config/setup errors are returned directly.

### Logging Patterns

**charm.land/log/v2 usage:**

```go
// INFO: normal progress
log.Info("cloning repository", "repo", repo.Name, "project", proj.Name)

// INFO (dry-run): literal [DRY RUN] prefix in message
log.Info("[DRY RUN] would clone", "repo", repo.Name, "cmd", renderedCmd)

// WARN: non-fatal issues (skips)
log.Warn("repository already exists, skipping", "path", destPath)

// Errors logged by reporter.Record() only — workspace functions do NOT call log.Error()
```

**Rules:**
- ALL log calls use structured key-value pairs — never format strings for data
- `[DRY RUN]` is a literal string prefix in the message, not a separate field
- Workspace functions do NOT call `log.Error()` — reporter handles error logging
- Verbose output gated: check `opts.Verbose` before calling log for command details

**stdout vs stderr:**
- `log.Info` / `log.Warn` → stdout logger
- `log.Error` + end-of-run failure summary → stderr logger
- Reporter uses two separate `log.Logger` instances: one for stdout, one for stderr

### Config Struct Patterns

```go
type Config struct {
    Workspaces []Workspace `yaml:"workspaces"`
}

type Workspace struct {
    Name                 string    `yaml:"name"`
    Path                 string    `yaml:"path"`
    CloneCommandTemplate string    `yaml:"clone_command_template"`
    Structure            string    `yaml:"structure"`  // "tree" (default) or "list"
    Projects             []Project `yaml:"projects"`
}

type Project struct {
    Name         string       `yaml:"name"`
    Repositories []Repository `yaml:"repositories"`
}

type Repository struct {
    Name         string `yaml:"name"`
    Branch       string `yaml:"branch"`
    LocalDirName string `yaml:"local_dir_name,omitempty"`
}
```

- All fields: value types (no pointers) — use zero-value checks in validation
- Optional fields: `omitempty` YAML tag — `LocalDirName` only; `Structure` defaults to `"tree"` when empty
- Each type is a named top-level type — no embedded structs
- `StructureTree = "tree"` and `StructureList = "list"` constants defined in `internal/config`
- `workspace.ResolveDestName(projPath, repoName, localDirName, structure)` computes the final local directory name: for `list` mode it tries the shortest unique trailing suffix of `repoName` (split on `/`), falling back to longer suffixes and then the full name

### Process Patterns

**Dry-Run Implementation:**
```go
if opts.DryRun {
    log.Info("[DRY RUN] would create directory", "path", dirPath)
    return nil
}
if err := os.MkdirAll(dirPath, 0755); err != nil { ... }
```
- Every side-effecting call has a corresponding dry-run branch that logs and returns

**Idempotency Check Pattern:**
```go
if _, err := os.Stat(destPath); err == nil {
    reporter.Record(OpResult{..., Skipped: true})
    return
}
```

**Defer Cleanup Pattern:**
```go
success := false
defer func() {
    if !success { os.RemoveAll(destPath) }
}()
// ... run clone ...
success = true
```

### Testing Patterns

**Table-driven tests (mandatory for all domain logic):**
```go
func TestResolveTarget(t *testing.T) {
    tests := []struct {
        name    string
        wsFlag  string
        allFlag bool
        cfg     *config.Config
        want    []config.Workspace
        wantErr bool
    }{
        {"default workspace", "", false, cfgWithDefault, ...},
        {"named workspace", "myws", false, cfg, ...},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { ... })
    }
}
```

### Enforcement Guidelines

**All AI Agents MUST:**
- Import `charm.land/log/v2` — NOT `github.com/charmbracelet/log`
- Pass `RunOptions` struct to ALL workspace functions — never read Cobra flags inside `internal/`
- Record repo-level failures via `reporter.Record()` — never `return err` from repo loops
- Check `opts.DryRun` before EVERY side-effecting filesystem or exec call
- Use `fmt.Errorf("context: %w", err)` for all error wrapping
- Use `filepath.Join(...)` — NEVER `path.Join(...)` (NFR10)
- Write table-driven tests for all non-trivial domain functions

**Anti-Patterns (NEVER do these):**
- Calling `log.Error(...)` in workspace pkg — use reporter instead
- `if dryRun { return }` without logging what would have happened
- Returning after first validation error — always collect all errors first
- Calling `exec.Command(...)` directly — always use `executor.Executor` interface
- Reading Cobra flags inside the `internal` module
- Importing Cobra (`github.com/spf13/cobra`) from the `internal` module
- Importing `github.com/vhula/grazhda/cmd` from the `internal` module (would create a cycle)

## Project Structure & Boundaries

### Complete Project Directory Structure

```
grazhda/                               # repo root (go.work)
│
├── cmd/                               # module: github.com/vhula/grazhda/cmd
│   ├── go.mod
│   └── ws/
│       ├── ws.go                      # `zgard ws` parent command registration
│       ├── init.go                    # `zgard ws init` — flags → RunOptions → workspace.Init()
│       ├── purge.go                   # `zgard ws purge` — flags → RunOptions → workspace.Purge()
│       └── pull.go                    # `zgard ws pull` — flags → RunOptions → workspace.Pull()
│
├── internal/                          # module: github.com/vhula/grazhda/internal
│   ├── go.mod
│   │
│   ├── config/
│   │   ├── config.go                  # Config/Workspace/Project/Repository structs; Load()
│   │   ├── validate.go                # Validate() — collects all errors, returns []string
│   │   ├── template.go                # RenderCloneCmd() — Go template rendering
│   │   ├── config_test.go
│   │   ├── validate_test.go
│   │   └── template_test.go
│   │
│   ├── targeting/
│   │   ├── resolver.go                # Resolve() — returns []config.Workspace from flags
│   │   └── resolver_test.go
│   │
│   ├── executor/
│   │   ├── executor.go                # Executor interface; OsExecutor implementation
│   │   ├── mock.go                    # MockExecutor; records Calls []string for test assertions
│   │   └── executor_test.go
│   │
│   ├── workspace/
│   │   ├── options.go                 # RunOptions struct definition
│   │   ├── init.go                    # Init(cfg, workspaces, exec, reporter, opts)
│   │   ├── purge.go                   # Purge(cfg, workspaces, reporter, opts)
│   │   ├── pull.go                    # Pull(cfg, workspaces, exec, reporter, opts)
│   │   ├── init_test.go
│   │   ├── purge_test.go
│   │   └── pull_test.go
│   │
│   ├── reporter/
│   │   ├── reporter.go                # Reporter struct; Record(); Summary(); ExitCode()
│   │   └── reporter_test.go
│   │
│   └── testdata/
│       ├── valid_single_workspace.yaml
│       ├── valid_multi_workspace.yaml
│       ├── missing_branch.yaml
│       ├── duplicate_workspace_names.yaml
│       ├── invalid_template.yaml
│       └── no_default_workspace.yaml
│
└── zgard/                             # module: github.com/vhula/grazhda/zgard
    ├── go.mod                         # requires: github.com/vhula/grazhda/cmd, github.com/vhula/grazhda/internal
    ├── go.sum
    └── main.go                        # Cobra root + global flags; calls os.Exit(reporter.ExitCode())
```

### FR Category → Package Mapping

| FR Category | FRs | Package(s) |
|---|---|---|
| Configuration Management | FR1–8 | `internal/config/` (config.go, validate.go, template.go) |
| Workspace Targeting | FR9–12 | `internal/targeting/resolver.go` |
| Workspace Initialization | FR13–19 | `internal/workspace/init.go` + `internal/executor/` |
| Workspace Teardown | FR20–22 | `internal/workspace/purge.go` |
| Repository Synchronization | FR23–26 | `internal/workspace/pull.go` + `internal/executor/` |
| Operation Feedback & Reporting | FR27–31 | `internal/reporter/reporter.go` |
| Automation & Scripting | FR32–36 | `cmd/ws/*.go` (flags) + `internal/reporter/` (exit codes) |

### Architectural Boundaries

**Package input/output contracts:**

```
cmd/ws/init.go                         # module: github.com/vhula/grazhda/cmd
  IN:  Cobra flags (--name, --all, --dry-run, --verbose, --no-confirm, --parallel)
  OUT: os.Exit(reporter.ExitCode())

github.com/vhula/grazhda/internal/config/
  IN:  $GRAZHDA_DIR/config.yaml path (string)
  OUT: *Config or []error

github.com/vhula/grazhda/internal/targeting/
  IN:  *Config, wsFlag string, allFlag bool
  OUT: []Workspace or error

github.com/vhula/grazhda/internal/workspace/init.go
  IN:  []Workspace, Executor, *Reporter, RunOptions
  OUT: (side effects: mkdir + clone); results via reporter.Record()

github.com/vhula/grazhda/internal/executor/
  IN:  dir string, command string
  OUT: error

github.com/vhula/grazhda/internal/reporter/
  IN:  OpResult structs via Record()
  OUT: Summary() to stdout/stderr; ExitCode() int
```

### Data Flow

```
User invokes: zgard ws init --name myws --dry-run

zgard/main.go
  └─► cmd/ws/init.go              Parse flags → build RunOptions{DryRun:true, WsName:"myws"}
        └─► internal/config.Load()     Read $GRAZHDA_DIR/config.yaml
        └─► internal/config.Validate() Collect all errors; abort if any
        └─► internal/targeting.Resolve() → []Workspace{myws}
        └─► internal/workspace.Init(workspaces, executor, reporter, opts)
              For each workspace → project → repo:
                └─► internal/config.RenderCloneCmd()   → rendered shell string
                └─► opts.DryRun check                  → log "[DRY RUN]" and skip exec
                └─► internal/executor.Run(dir, cmd)    → error (real run only)
                └─► reporter.Record(OpResult{...})
        └─► reporter.Summary()    → stdout/stderr
        └─► os.Exit(reporter.ExitCode())
```

### Integration Points

**External integrations (Phase 1):**
- `git` binary via `$PATH` — invoked by `OsExecutor` via `sh -c <rendered_clone_cmd>`
- `$GRAZHDA_DIR` env var — read in `cmd/ws/*.go` to locate `config.yaml`

**Internal communication — all via function parameters (no globals):**
- `Reporter` instance created in `cmd/ws/*.go`, passed to `internal/workspace.*` functions
- `Executor` instance created in `cmd/ws/*.go` (real or mock), passed down
- `RunOptions` struct threaded through all workspace function calls

No global state. No package-level variables. All dependencies are injected.

### Development Workflow Integration

- **Build:** `just build-zgard` → compiles `zgard/` → outputs `bin/zgard`
- **Test:** `just test` → `go test ./...` across `cmd/`, `internal/`, `zgard/`
- **Format/Tidy:** `just fmt` + `just tidy` → runs per-module
- **Install:** `grazhda.sh` → clone → `just build` → copy `bin/zgard` to `$GRAZHDA_DIR/bin/`

## Architecture Validation Results

### Coherence Validation ✅

All technology decisions are compatible and version-safe. Package responsibilities are non-overlapping. Patterns (RunOptions, Executor interface, Reporter accumulator, defer cleanup) are mutually consistent. Structure maps directly to requirement categories.

### Requirements Coverage ✅

All 36 FRs and 14 NFRs have architectural support. See FR Category → Package Mapping in Project Structure section for full traceability.

### Gaps Addressed

**Gap 1 — $GRAZHDA_DIR resolution (RESOLVED):**
`$GRAZHDA_DIR` env var is read in `cmd/ws/*.go` (CLI layer only). The resolved path string is passed to `config.Load(path string)`. `internal/config` never reads env vars. If `$GRAZHDA_DIR` is unset, `cmd` layer defaults to `$HOME/.grazhda`.

**Gap 2 — Confirmation prompt location (RESOLVED):**
Interactive Y/N confirmation lives in `internal/workspace/purge.go`. A small inline `confirm(prompt string, r io.Reader) bool` helper reads from an `io.Reader` parameter (defaults to `os.Stdin` in real use; injectable for tests). `opts.NoConfirm` skips it. No separate `prompt/` package needed for Phase 1.

**Gap 3 — --parallel on ws purge (RESOLVED):**
`--parallel` is **not registered** on `zgard ws purge` — purge operates on a single workspace directory; parallelism is meaningless. The flag is only registered on `ws init` and `ws pull`. Intentional deviation from PRD FR36's "shared flags" statement.

### Architecture Completeness Checklist

**✅ Requirements Analysis**
- [x] Project context thoroughly analyzed
- [x] Scale and complexity assessed
- [x] Technical constraints identified
- [x] Cross-cutting concerns mapped

**✅ Architectural Decisions**
- [x] Critical decisions documented with verified versions
- [x] Executor interface contract defined
- [x] RunOptions struct specified
- [x] Result accumulator model defined
- [x] Partial cleanup pattern defined
- [x] Parallel execution model selected

**✅ Implementation Patterns**
- [x] Go naming conventions established
- [x] Package responsibility boundaries defined
- [x] Error handling flow specified
- [x] Logging patterns with charm.land/log/v2 documented
- [x] Config struct with YAML tags specified
- [x] Dry-run implementation pattern defined
- [x] Testing patterns (table-driven, mocks) specified
- [x] Anti-patterns explicitly listed

**✅ Project Structure**
- [x] Complete file-level directory tree defined
- [x] FR category → package mapping complete
- [x] Package input/output contracts documented
- [x] Data flow diagram (init command walkthrough)
- [x] Integration points documented

### Architecture Readiness Assessment

**Overall Status: READY FOR IMPLEMENTATION**

**Confidence Level: High** — all 36 FRs and 14 NFRs have traceable architectural support; patterns are specific enough to prevent agent conflicts.

**Key Strengths:**
- Clean dependency injection throughout (no globals, no Cobra in `internal/`)
- Explicit package boundary contracts prevent scope creep
- All cross-cutting concerns (dry-run, targeting, reporting) are centralized
- Data flow is linear and easy to trace

**Areas for Future Enhancement (Post-MVP):**
- Worker pool with configurable concurrency for `--parallel`
- Shell completion via Cobra's built-in completion commands
- `zgard config init` scaffolding command

## Dukh Server Architecture Vision

This section describes the planned architecture of `dukh` — the second Go component in the Grazhda system — to provide system-wide context and ensure `zgard`'s Phase 1 design decisions remain compatible with the full architecture.

### System-Level Context

Grazhda is a three-tier system:

```
┌─────────────────────────────────────────────┐
│  Tier 3: Brain / Interface                   │
│  molfar (Java) — orchestrates workflows      │
│  molf    (Java) — human operator CLI         │
└───────────────┬─────────────────────────────┘
                │  gRPC
┌───────────────▼─────────────────────────────┐
│  Tier 2: Workers (Go)                        │
│  zgard — workspace setup & lifecycle CLI     │
│  dukh  — process lifecycle gRPC server       │
└───────────────┬─────────────────────────────┘
                │  shared config + filesystem
┌───────────────▼─────────────────────────────┐
│  Tier 1: Workspace & Processes               │
│  $GRAZHDA_DIR/config.yaml                    │
│  ~/ws/<workspace>/<project>/<repo>/          │
└─────────────────────────────────────────────┘
```

`zgard` and `dukh` are independent workers: each can be used standalone from the command line, or driven programmatically by `molfar` via gRPC.

### Dukh Module Structure

`dukh/` is a separate Go module in the `go.work` workspace:

```
dukh/                              # module: github.com/vhula/grazhda/dukh
├── go.mod                         # requires: github.com/vhula/grazhda/internal
├── go.sum
├── main.go                        # gRPC server startup; reads config; listens on host:port
├── server/
│   └── dukh.go                    # DukhServer implements the generated gRPC service interface
├── process/
│   ├── manager.go                 # ProcessManager: start/stop/status for managed processes
│   └── manager_test.go
└── proto/
    └── dukh.proto                 # gRPC service definition (generates dukh/gen/)
```

### gRPC Service Interface

```protobuf
service Dukh {
  rpc Start  (StartRequest)  returns (StartResponse);
  rpc Stop   (StopRequest)   returns (StopResponse);
  rpc Status (StatusRequest) returns (StatusResponse);
}
```

### Shared Internal Module Usage

`dukh` imports `github.com/vhula/grazhda/internal` for:
- `internal/config` — reads the same `$GRAZHDA_DIR/config.yaml`; workspace/project/process config structs are shared
- `internal/reporter` — optional: consistent output formatting if `dukh` exposes a local CLI

This is the primary architectural reason `internal/` is a standalone module at the repo root rather than embedded inside `zgard/`.

### Config Integration

`dukh` reads its own `dukh:` section from `config.yaml`:

```yaml
dukh:
  host: localhost
  port: 50501
```

Process definitions (planned) will be declared inside workspace/project entries — same config file, no duplication.

### Compatibility Constraints on zgard

The following Phase 1 decisions are made with `dukh` compatibility in mind:

| Decision | Rationale |
|---|---|
| `internal/` is a standalone module | `dukh` can import shared config/reporter without taking on `zgard` or `cmd` as a dependency |
| `internal/config.Config` is the canonical workspace model | `dukh` uses the same struct to reason about workspaces — no divergent models |
| No env-var reading inside `internal/` | Both `zgard` and `dukh` resolve `$GRAZHDA_DIR` in their own entry points and pass a path into `internal/config.Load()` |
| Executor interface is injectable | Future: `dukh` may use the same `Executor` abstraction for process execution |

### Implementation Handoff

**AI Agent Guidelines:**
- Follow all architectural decisions exactly as documented
- Use implementation patterns consistently across all components
- Respect package responsibility boundaries — no cross-boundary direct calls
- Refer to this document for all architectural questions

**First Implementation Story:** Initialize three Go modules (`cmd/`, `internal/`, `zgard/`) with `go mod init`, register them in `go.work`, add dependencies (`go get`) to each, create the package skeleton (empty files with correct package declarations), and verify `just build-zgard` succeeds.

---

## Phase 2 — dukh Architecture

### Overview

`dukh` is a standalone Go module (`github.com/vhula/grazhda/dukh`) registered in the `go.work` workspace. It runs as a long-running background gRPC server, polling workspace health and exposing it via a well-defined protobuf API.

### Directory Structure

```
proto/
  dukh.proto              # source of truth — do not edit generated output here

dukh/
  cmd/
    main.go               # cobra CLI: dukh start (only subcommand)
  server/
    server.go             # gRPC server struct, Start/Stop lifecycle, DukhService impl
    monitor.go            # workspace polling goroutine, branch detection, health snapshot
    log.go                # charmbracelet/log + lumberjack writer setup
  proto/                  # generated protobuf Go code — do not edit manually
    dukh.pb.go
    dukh_grpc.pb.go
  go.mod                  # module: github.com/vhula/grazhda/dukh
  go.sum
```

### gRPC Contract

Proto source: `proto/dukh.proto`

```protobuf
syntax = "proto3";
package dukh;
option go_package = "github.com/vhula/grazhda/dukh/proto";

service DukhService {
  rpc Stop   (StopRequest)   returns (StopResponse);
  rpc Status (StatusRequest) returns (StatusResponse);
}

message StopRequest   {}
message StopResponse  { string message = 1; }

message StatusRequest {
  string workspace_name = 1;
  bool   rescan         = 2;  // when true: perform synchronous rescan before returning snapshot
}
message StatusResponse {
  repeated WorkspaceStatus workspaces    = 1;
  string                   server_version = 2;
  int64                    uptime_seconds = 3;
}

message WorkspaceStatus {
  string                 name     = 1;
  string                 path     = 2;
  repeated ProjectStatus projects = 3;
}

message ProjectStatus {
  string              name         = 1;
  repeated RepoStatus repositories = 2;
}

message RepoStatus {
  string name              = 1;
  string path              = 2;
  string configured_branch = 3;
  string actual_branch     = 4;
  bool   exists            = 5;
  bool   branch_aligned    = 6;
}
```

### Monitoring Logic

`monitor.go` runs a goroutine that sleeps 30 seconds between polls. On each cycle:

1. Load the current config from `internal/config` (re-read to pick up changes).
2. For each workspace → project → repository:
   - Compute `path = workspace.Path + "/" + project.Name + "/" + repo.LocalDirName()`.
   - Set `exists = os.Stat(path) == nil`.
   - If `exists`, run `git -C <path> symbolic-ref --short HEAD` to get `actual_branch`; on non-zero exit, set `actual_branch = ""`.
   - Set `branch_aligned = exists && actual_branch == configuredBranch`.
3. Build a new `[]WorkspaceStatus` snapshot.
4. Acquire write lock on `sync.RWMutex`, replace stored snapshot, release lock.

The `Status` RPC acquires the read lock, copies the snapshot, and returns it. Filtering by `workspace_name` happens in the handler before returning.

### Logging Setup (`server/log.go`)

```go
lj := &lumberjack.Logger{
    Filename:   filepath.Join(grazhda dir, "logs", "dukh.log"),
    MaxSize:    5,   // MiB
    MaxBackups: 3,
    Compress:   true,
}
logger := log.New(lj)  // charmbracelet/log
```

All structured log calls use `logger.Info(...)`, `logger.Error(...)` with key-value pairs. The logger instance is passed through the server struct; no global logger is used.

### PID File Management

- On `dukh start`: `os.MkdirAll($GRAZHDA_DIR/run, 0755)`, write PID to `$GRAZHDA_DIR/run/dukh.pid`.
- On graceful shutdown (Stop RPC or OS signal): `os.Remove($GRAZHDA_DIR/run/dukh.pid)`.
- On startup: if PID file exists and `os.FindProcess(pid)` succeeds with signal 0, exit with error "dukh already running (pid N)".

### Graceful Shutdown

`server.go` registers `signal.NotifyContext` for `SIGTERM` and `SIGINT`. The Stop RPC calls the same internal shutdown function:

1. Call `grpcServer.GracefulStop()`.
2. Remove PID file.
3. Flush and close log writer.
4. Cancel context to unblock the main goroutine.

### Build: Code Generation

`Justfile` recipe:

```
generate:
    protoc --go_out=dukh/proto --go-grpc_out=dukh/proto \
           --go_opt=paths=source_relative \
           --go-grpc_opt=paths=source_relative \
           proto/dukh.proto
```

Run with `just generate` before building `dukh`.

### Dependencies

| Package | Purpose |
|---|---|
| `google.golang.org/grpc` | gRPC runtime |
| `google.golang.org/protobuf` | protobuf serialization |
| `github.com/charmbracelet/log` | structured, levelled logging |
| `gopkg.in/natefinish/lumberjack.v2` | log file rotation |
| `github.com/spf13/cobra` | CLI framework for `dukh start` |
| `github.com/vhula/grazhda/internal` | shared config + workspace model |

### Config Integration

`dukh` reads from the same `$GRAZHDA_DIR/config.yaml` using `internal/config.Load()`. It additionally reads:

```yaml
dukh:
  host: localhost   # default: localhost
  port: 50501       # default: 50501
```

### Synchronous Rescan Flow (`--rescan`)

When `zgard dukh status --rescan` is called the following happens:

1. zgard creates a 60-second `context.WithTimeout` and builds `StatusRequest{Rescan: true}`.
2. dukh's `Status()` handler detects `req.Rescan == true` and calls `monitor.TriggerScanAndWait(ctx)`.
3. `TriggerScanAndWait` creates a `reply chan struct{}` and sends it on `triggerScan chan chan struct{}` (capacity 1).
4. The monitor loop receives the reply channel, calls `scan()`, then closes the reply channel.
5. `TriggerScanAndWait` unblocks when the reply channel is closed and returns `nil`.
6. The handler reads the (now fresh) snapshot and builds the `StatusResponse`.
7. zgard receives the response and renders the health report.

The gRPC context deadline propagates all the way into `TriggerScanAndWait`, so if the scan exceeds 60 seconds the call is cancelled and zgard prints an error — no hanging.

### Architectural Invariants

- `dukh/proto/` is generated code — never edited by hand.
- No business logic in `cmd/main.go`; it only wires cobra → `server.Start()`.
- `monitor.go` does not import `server.go`; the server struct owns the monitor goroutine via a channel or `context.Context` for cancellation.
- `internal/config` is the single source of truth for workspace structure; `dukh` does not duplicate config parsing logic.
