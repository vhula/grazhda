---
stepsCompleted: [step-01-init, step-02-discovery, step-03-core-experience, step-04-emotional-response, step-05-inspiration, step-06-design-system, step-07-defining-experience, step-08-visual-foundation, step-09-design-directions, step-10-user-journeys, step-11-component-strategy, step-12-ux-patterns, step-13-responsive-accessibility, step-14-complete]
workflowStatus: complete
completedAt: '2026-04-05T12:08:00Z'
inputDocuments:
  - docs/prd.md
  - docs/architecture.md
---

# UX Design Specification — Grazhda (zgard CLI)

**Author:** jake
**Date:** 2026-04-05

---

<!-- UX design content will be appended sequentially through collaborative workflow steps -->

## Executive Summary

### Project Vision

`zgard` is a local CLI automation tool that eliminates the manual, error-prone process of setting up and maintaining developer workspaces. It lets a developer declare their entire workspace topology — directories, repositories, branches — in a single YAML config file, and then bring that topology to life (or tear it down) with a single command. The tool's value is consistency: the same config produces the same workspace on any machine, every time.

### Target Users

- **Solo developers and small teams** managing multi-repo, multi-project workspaces across multiple contexts (e.g., a `main` workspace and a `feature-x` workspace with different branch sets)
- **DevOps / platform engineers** responsible for repeatable, scriptable developer environment setups
- **CI/CD pipeline operators** who need non-interactive automation with predictable exit codes and machine-readable output
- All users are **tech-savvy command-line operators** — comfortable with `git`, shell scripting, and YAML configuration

### Key Design Challenges

1. **Information density vs. clarity** — a single `ws init` can touch dozens of repositories across multiple projects. Users need to track progress, skips, and failures in real time without being overwhelmed by a wall of undifferentiated text.
2. **Safety for destructive operations** — `ws purge` removes directory trees. The confirmation flow, dry-run preview, and `--no-confirm` bypass must all feel trustworthy: never alarming by default, but never something a user can accidentally skip.
3. **Dry-run fidelity** — the `[DRY RUN]` output mode must feel like a faithful, actionable preview — specific enough that a user can verify correctness before committing to the real run.

### Design Opportunities

1. **Structured progressive output** — grouping output by workspace → project → repo with consistent visual hierarchy (indentation, status prefixes) makes even large operations scannable at a glance.
2. **Summary as the hero moment** — an end-of-run summary (`✓ 12 cloned  ⏭ 3 skipped  ✗ 1 failed`) gives users an instant health signal and is the single most important piece of output the tool produces.
3. **Consistent signal language** — a small, coherent vocabulary of status symbols (`[DRY RUN]`, `[SKIP]`, `[OK]`, `[FAIL]`) used everywhere creates a learnable mental model that transfers across all commands.

## Core User Experience

### Defining Experience

The defining experience of `zgard` is **`ws init` on a new machine** — the moment a developer types one command and watches their entire workspace materialise exactly as declared. This is the make-or-break interaction. If it is clear, observable, and lands without surprises, the tool earns immediate trust. Every other command (`ws pull`, `ws purge`) is secondary to this core moment.

The experience must support a natural two-step workflow: `zgard ws init --dry-run` (verify what will happen) followed by `zgard ws init` (commit). Dry-run is not an advanced feature — it is the recommended first step and must feel like a first-class citizen of the interface, not an afterthought.

### Platform Strategy

- **Platform:** POSIX terminal (Linux, macOS) — the shell is the entire UI surface
- **Input model:** keyboard only; no mouse interaction
- **Colour support:** auto-detect TTY — full colour when stdout is a terminal, plain text when piped or redirected (CI-safe by default)
- **Offline:** fully offline after `git` is available — no network calls except the git operations themselves
- **Accessibility:** plain-text output at all times; colour is additive, never the only signal carrier

### Effortless Interactions

- **Zero-flag happy path:** `zgard ws init` (no flags) initialises the default workspace with no required arguments — the simplest possible invocation works
- **Safe-by-default purge:** `ws purge` always requires explicit targeting (`--name` or `--all`) and prompts for confirmation — no way to accidentally purge without intent
- **Idempotent re-init:** running `ws init` on an already-initialised workspace skips existing repos silently — users never need to check state before re-running
- **Re-runnable after failure:** a failed clone leaves no partial state (cleanup on failure) — the user can just re-run without manual cleanup

### Critical Success Moments

1. **First `ws init` completion** — user sees a clean summary: all repos cloned, right branches, zero failures. The workspace is ready.
2. **The `--dry-run` preview** — user sees exactly which directories would be created and which clone commands would run, formatted identically to a real run but clearly marked `[DRY RUN]`. Confidence before commitment.
3. **A failed clone, clearly reported** — user sees `[FAIL] myrepo — exit 128: repository not found` with enough context to act immediately. No log hunting.
4. **`ws purge --dry-run`** — user sees exactly which directories would be removed before confirming. The destructive act is always previewable.

### Experience Principles

1. **Observe, don't hunt** — every operation produces visible, structured output in real time. Users watch progress; they never wonder what is happening.
2. **Preview before commit** — `--dry-run` is always available and always accurate. The tool encourages looking before acting.
3. **Fail loudly, recover cleanly** — failures are immediately visible with actionable context; partial state is never left behind.
4. **Consistency over cleverness** — the same status prefixes, the same output structure, the same flag names across every command. Predictability is the feature.

## Desired Emotional Response

### Primary Emotional Goals

**Primary arc:** Anxious → Certain → Accomplished

**Underlying thread:** Trust earned through consistency — every time `--dry-run` output faithfully mirrors what the real run does, the user's trust in the tool deepens. Trust is not assumed; it is built through repeated fidelity.

**Primary emotional goal: Certainty.** Not serenity — certainty. The user knows exactly what ran, exactly what succeeded or failed, and exactly what to do next. No inference required. This applies equally to the interactive developer and the CI operator reading exit codes at 2am.

### Emotional Journey Mapping

| Stage | Desired feeling |
|---|---|
| First discovery / install | Intrigued — "this solves a real pain" |
| Writing `config.yaml` | In control — "I'm declaring exactly what I want" |
| First `--dry-run` | Anxious-but-reassured — output is readable enough to feel safe committing |
| Watching `ws init` run | Calm trust — steady progress, no surprises |
| Seeing the summary | Accomplished — "done, and I can verify it" |
| A clone fails | Informed, not panicked — actionable failure context immediately visible |
| Re-running after fix | Relieved — idempotent, no manual cleanup |
| 10th time using the tool | Confident trust — dry-run has always matched reality |

### Micro-Emotions

- **Confidence** (not anxiety) before destructive commands → delivered via confirmation prompt + `--dry-run`
- **Trust** (not scepticism) in dry-run fidelity → delivered via output structurally identical to real run
- **Certainty** (not ambiguity) about outcome → delivered via explicit exit codes, stderr/stdout split, always-present summary
- **First-time anxiety is a real starting state** — the design must actively reduce it through explicit, readable dry-run output, not assume pre-existing trust

### Design Implications

- Dry-run output must be structurally identical to real-run output — only `[DRY RUN]` prefix distinguishes them; no different format or level of detail
- Summary is always the last thing printed — it is the emotional resolution of every command run
- stderr is terse and parseable — CI operators must never hunt through stdout to find failures
- Progress output must make it clear the tool is active — users should never wonder "is it stuck?"

### Emotional Design Principles

1. **Certainty over calm** — the goal is zero ambiguity about state, not a gentle UX; explicit signals beat soft reassurances
2. **Trust is earned, not assumed** — consistency between dry-run and real-run is the primary trust mechanism; every deviation destroys it
3. **Acknowledge first-time anxiety** — new users are nervous; clear, scannable dry-run output is the tool's handshake
4. **Fail loudly, recover cleanly** — a failure that is immediately visible and actionable produces relief, not frustration

## UX Pattern Analysis & Inspiration

### Inspiring Products Analysis

**`git`** — The gold standard for developer CLI UX. Consistent subcommand structure (`git remote add`, `git stash pop`), dry-run on destructive ops, terse-but-complete output. The `zgard ws init` / `ws purge` / `ws pull` surface mirrors this subcommand idiom. Onboarding is self-teaching: `zgard --help` and `zgard ws --help` cascade exactly like git.

**`docker compose up`** — Excellent progressive output for multi-entity parallel operations. Each entity logs prefixed lines (`[service-name] ...`); the hierarchy is clear at a glance even with 20+ services. Maps directly to our workspace → project → repo output hierarchy.

**`rsync --dry-run`** — Legendary for fidelity between preview and real run. Output is structurally identical; users trust it completely. This is the exact model for `zgard --dry-run`.

**`terraform plan` / `apply`** — Two-phase workflow (plan = dry-run, apply = commit) is now a well-understood developer mental model. Symbol-prefixed change lines (`+`, `-`, `~`) give instant scannability. Confirms the value of our two-phase dry-run → commit pattern.

**`brew install`** — Clean, calm progress. One status line per entity as it completes. Clear failure messages with actionable context. The emotional template for a `ws init` run.

### Transferable UX Patterns

**Output hierarchy pattern (from docker compose):**
```
Workspace: default
  Project: backend
    ✓ api          — cloned (main)
    ✓ auth-service — cloned (main)
  Project: frontend
    ✗ webapp       — exit 128: repository not found
```

**Two-phase workflow (from terraform):** `--dry-run` first, commit second. Users who know terraform already have this mental model. Reinforce it in docs and `--help` output.

**Symbol vocabulary (adapted from terraform + brew):**
- `✓` — success
- `✗` — failure
- `⏭` — skipped (already exists)
- `[DRY RUN]` — preview prefix on every line

**Summary line (from brew):** Always last. Counts only. Machine-readable if parsed. Example: `✓ 12 cloned  ⏭ 3 skipped  ✗ 1 failed`

### Anti-Patterns to Avoid

- **Spinner-only progress** — hides per-entity status; users can't tell what's running or what failed mid-stream
- **Interleaved parallel output without buffering** — unreadable with many repos; must buffer per-entity blocks when `--parallel` is active
- **Verbose-by-default** — buries the signal; default output should be compact; `--verbose` unlocks full command rendering
- **Different format for `--dry-run`** — if dry-run output looks different from real-run output, users can't verify fidelity
- **Errors only in exit code** — every failure must produce a human-readable stderr line before exit

### Design Inspiration Strategy

| Action | Pattern | Reason |
|---|---|---|
| Adopt | Prefixed-line-per-entity output | Maps naturally to workspace → project → repo hierarchy |
| Adopt | Two-phase dry-run → commit | Proven developer mental model; reduces first-time anxiety |
| Adopt | Summary-as-last-line | Emotional resolution; CI-parseable |
| Adapt | Colour prefixes → monochrome symbols | Colour must be additive; symbols work in plain text and colour alike |
| Avoid | Spinner-only progress | Leaves users blind mid-operation |
| Avoid | Divergent dry-run format | Destroys trust fidelity |

## Design System Foundation

### Design System Choice

**Platform:** POSIX terminal — the only "design system" is the terminal output contract.

**Selected approach: Structured plain-text with additive colour**

No external TUI framework (Bubble Tea, tview) for Phase 1. Output is composed of plain-text lines with a consistent prefix grammar. `fatih/color` provides colour helpers that auto-detect TTY and disable colours gracefully in CI/pipe contexts.

**Rationale:**
- Simplest possible dependency surface for a CLI tool
- Colour is additive — every status is conveyed by symbol/prefix first, colour second
- `fatih/color` handles TTY detection and colour disabling automatically
- Plain-text output is universally parseable by shell scripts and CI systems

### Output Vocabulary (Design Tokens)

| Token | Symbol | Colour (TTY) | Meaning |
|---|---|---|---|
| SUCCESS | `✓` | green | Operation completed successfully |
| FAIL | `✗` | red | Operation failed |
| SKIP | `⏭` | yellow | Skipped (already exists / not applicable) |
| DRY RUN | `[DRY RUN]` | dim/cyan | Preview mode — no changes made |
| INFO | (none) | default | Progress or informational message |
| WARN | `!` | yellow | Non-fatal warning |
| ERROR | `✗` | red → stderr | Fatal or per-operation failure |

### Typography & Spacing

Terminal "typography" is whitespace and indentation:
- **Workspace header:** no indent, bold label: `Workspace: default`
- **Project header:** 2-space indent: `  Project: backend`
- **Repo status line:** 4-space indent: `    ✓ api — cloned (main)`
- **Summary block:** blank line separator before, no indent: `✓ 12 cloned  ⏭ 3 skipped  ✗ 1 failed`
- **Error detail:** 6-space indent under failed repo line: `      exit 128: repository 'xyz' not found`

## Core User Experience (Detailed)

### Defining Experience

The defining experience is **`zgard ws init` completing cleanly on a new machine**. The user types one command, watches a structured stream of per-repo status lines, and sees a summary that confirms every repo is in place on the right branch. The workspace is ready. No manual steps, no branch checks, no missed repos.

### User Mental Model

Users approach `zgard` with a **declaration → execution** mental model (same as terraform/docker compose):

1. Declare intent in `config.yaml`
2. Preview with `--dry-run` to verify intent matches expectation
3. Execute — the tool makes the filesystem match the declaration

The tool never surprises the user. Every action is either previewed or immediately visible in structured output.

### Success Criteria for Core Experience

- Running `zgard ws init` on a clean machine produces exactly the directory and repo structure described in `config.yaml`
- Running it again (idempotent) skips already-present repos and reports `⏭ skipped` — not an error
- Running with `--dry-run` produces output that is structurally identical to the real run, prefixed with `[DRY RUN]`
- Any failure produces an actionable error line and a non-zero exit code; no silent failures

### Novel vs. Established Patterns

| Pattern | Type | Rationale |
|---|---|---|
| `ws init` / `ws purge` / `ws pull` subcommands | Established (git-style) | Familiar; self-documenting |
| `--dry-run` as first-class step | Established (terraform/rsync) | Proven trust-builder |
| Continue-on-failure with summary | Established (brew/CI) | Maximises info per run |
| Workspace → project → repo hierarchy in output | Novel for this tool | Matches config model; makes large workspaces scannable |

## Visual Design Foundation

### Colour System

Implemented via `fatih/color` — colours are auto-detected and downsampled to terminal capability. All colour choices have monochrome fallbacks via symbols.

| Context | Colour (256/truecolor) | Monochrome fallback |
|---|---|---|
| Success `✓` | `#00AF5F` (green) | `✓` symbol alone |
| Failure `✗` | `#FF5F5F` (red) | `✗` symbol alone |
| Skip `⏭` | `#FFAF00` (amber) | `⏭` symbol alone |
| `[DRY RUN]` prefix | `#5FAFFF` (cyan, dim) | `[DRY RUN]` text alone |
| Section headers | bold, default fg | bold text |
| Error detail lines | `#FF5F5F` → stderr | plain text on stderr |

### Spacing & Layout Foundation

- Max useful line width: 120 chars — repo names + status + branch should always fit on one line
- Long paths/commands truncated with `…` at 100 chars in default mode; full in `--verbose`
- Summary block always preceded by a blank line separator
- No ANSI art, no box-drawing characters — pure ASCII symbols only

### No Brand Guidelines

`zgard` has no brand guidelines. The visual foundation is purely functional: signal clarity, hierarchy, and CI compatibility.

## Design Direction Decision

### Selected Direction: Structured Signal

From the three possible directions for a CLI tool's output style:

1. **Minimal** — single-line output per command, summary only
2. **Structured Signal** ← **selected** — per-entity status lines with hierarchy + summary
3. **Rich TUI** — interactive terminal UI with spinners, progress bars, panels

**Structured Signal** is selected because:
- Matches the emotional goal of certainty — users see every operation as it happens
- Compatible with CI (`--no-confirm`, stdout/stderr split, exit codes)
- Consistent with the inspiration analysis (docker compose, brew, terraform)
- Achievable in Phase 1 without a TUI framework dependency

**Rejected: Minimal** — too little information during long-running multi-repo operations; users can't detect what's happening or what failed mid-stream.
**Rejected: Rich TUI** — over-engineered for Phase 1; breaks in CI; requires Bubble Tea dependency.

## User Journey Flows

### Journey 1: New Machine Setup (Happy Path)

```
$ zgard ws init --dry-run

[DRY RUN] Workspace: default
[DRY RUN]   Project: backend
[DRY RUN]     ⏭ api          — would clone (main) → ~/ws/backend/api
[DRY RUN]     ⏭ auth-service — would clone (main) → ~/ws/backend/auth-service
[DRY RUN]   Project: frontend
[DRY RUN]     ⏭ webapp       — would clone (dev)  → ~/ws/frontend/webapp

[DRY RUN] ✓ 3 would clone  ⏭ 0 skipped  ✗ 0 would fail

$ zgard ws init

Workspace: default
  Project: backend
    ✓ api          — cloned (main)
    ✓ auth-service — cloned (main)
  Project: frontend
    ✓ webapp       — cloned (dev)

✓ 3 cloned  ⏭ 0 skipped  ✗ 0 failed
```

### Journey 2: Re-init (Idempotent)

```
$ zgard ws init

Workspace: default
  Project: backend
    ⏭ api          — already exists, skipped
    ⏭ auth-service — already exists, skipped
  Project: frontend
    ✓ webapp       — cloned (dev)

✓ 1 cloned  ⏭ 2 skipped  ✗ 0 failed
```

### Journey 3: Partial Failure

```
$ zgard ws init

Workspace: default
  Project: backend
    ✓ api          — cloned (main)
    ✗ auth-service — exit 128: repository not found
  Project: frontend
    ✓ webapp       — cloned (dev)

✓ 2 cloned  ⏭ 0 skipped  ✗ 1 failed
      auth-service: exit 128: repository not found

exit status 1
```

### Journey 4: Workspace Purge

```
$ zgard ws purge --name default

This will permanently remove:
  ~/ws/backend/api
  ~/ws/backend/auth-service
  ~/ws/frontend/webapp

Confirm? [y/N]: y

  ✓ Removed ~/ws/backend/api
  ✓ Removed ~/ws/backend/auth-service
  ✓ Removed ~/ws/frontend/webapp

✓ 3 removed

$ zgard ws purge --name default --no-confirm   # CI / scripting
  ✓ Removed ~/ws/backend/api
  ...
```

### Journey 5: Sync All Workspaces

```
$ zgard ws pull --all

Workspace: default
  Project: backend
    ✓ api          — pulled (main)
    ⏭ auth-service — not present, skipped
  Project: frontend
    ✓ webapp       — pulled (dev)

Workspace: feature-x
  Project: backend
    ✓ api          — pulled (feature/x)

✓ 3 pulled  ⏭ 1 skipped  ✗ 0 failed
```

## Component Strategy

### Output Components (Terminal "Components")

For a CLI tool, "components" are reusable output format patterns produced by the reporter package.

**Component 1: Section Header**
```
Workspace: <name>
  Project: <name>
```
- Responsibility: `internal/reporter` prints these before iterating entities
- Separator: blank line between workspaces

**Component 2: Status Line**
```
    <symbol> <name> — <message>
```
- `<symbol>`: `✓`, `✗`, `⏭`, or `[DRY RUN]` prefix
- `<name>`: repo name (padded to align columns when possible)
- `<message>`: operation result or "would <action>" in dry-run
- Responsibility: `reporter.Record()` formats and writes

**Component 3: Summary Block**
```
<blank line>
✓ N <verb>  ⏭ N skipped  ✗ N failed
```
- Always last; written by `reporter.Summary()`
- Verbs: `cloned` (init), `pulled` (pull), `removed` (purge)

**Component 4: Error Detail**
```
      <repo>: <error message>
```
- Printed after summary for each failed operation
- Written to stderr

**Component 5: Confirmation Prompt**
```
This will permanently remove:
  <path>
  <path>

Confirm? [y/N]:
```
- Only for `ws purge` without `--no-confirm`
- Reads from `io.Reader` (injectable for tests)
- Default answer is `N` (safe default)

**Component 6: Dry-Run Line**
```
[DRY RUN]     ⏭ <name> — would clone (<branch>) → <path>
```
- Identical structure to real status line, prefixed with `[DRY RUN]`

### Implementation Mapping

| Component | Package | Function |
|---|---|---|
| Section Header | `internal/reporter` | `reporter.PrintHeader()` |
| Status Line | `internal/reporter` | `reporter.Record(OpResult)` |
| Summary Block | `internal/reporter` | `reporter.Summary()` |
| Error Detail | `internal/reporter` | written during `Summary()` to stderr |
| Confirmation Prompt | `internal/workspace/purge.go` | `confirm(prompt, io.Reader) bool` |
| Dry-Run Line | `internal/workspace/*.go` | guarded by `opts.DryRun` check before `executor.Run()` |

## UX Consistency Patterns

### Status Prefix Grammar

Every operational output line follows this grammar (enforced in `reporter.Record()`):

```
<indent><symbol> <entity-name><padding> — <message>
```

- Indent is always 4 spaces for repo-level lines, 2 spaces for project headers
- Symbol is always one of: `✓`, `✗`, `⏭`, `[DRY RUN]` (no mixing)
- Em-dash ` — ` is the separator between entity name and message (always present)
- Message is lowercase, no trailing period

### Flag Consistency

All `ws` subcommands share the same flag vocabulary:

| Flag | Short | Commands | Meaning |
|---|---|---|---|
| `--name <name>` | `-n` | init, purge, pull | Target named workspace |
| `--all` | | init, purge, pull | Target all workspaces |
| `--dry-run` | | init, purge, pull | Preview without executing |
| `--verbose` | `-v` | init, pull | Show rendered commands |
| `--no-confirm` | | purge | Skip Y/N prompt |
| `--parallel` | | init, pull | Clone/pull repos within each project concurrently |
| `--parallel-all` | | init, pull | Clone/pull all repos across all projects concurrently |
| `--clone-delay-seconds=N` | | init | Sleep N seconds after each clone command |

`--parallel` and `--parallel-all` are intentionally absent from `ws purge` (purge is a single directory removal; concurrency is meaningless). `--clone-delay-seconds` is intentionally absent from `ws pull` (pull is fast and in-place; throttling is not needed).

### Error Message Patterns

- Format: `<entity>: <error context>` — e.g. `auth-service: exit 128: repository not found`
- Always written to stderr
- Always accompanied by non-zero exit code
- Always printed in the summary error-detail block (never only mid-stream)

### Dry-Run Consistency Rules

- Every side-effecting call (`os.MkdirAll`, `executor.Run`) is guarded by `if opts.DryRun { log.Info("[DRY RUN] would ..."); return nil }`
- Dry-run output is produced in the same loop as real output — no separate code path
- Summary counts in dry-run use "would clone / would remove" verbs but same format

### Verbose Mode Patterns

When `--verbose` is active:
- Rendered clone/pull command is printed before execution: `  → git clone --branch main https://github.com/org/repo ~/ws/backend/api`
- Commands are printed even in `--dry-run` mode
- No other output changes

## Responsive Design & Accessibility

### Terminal Width Strategy

- **Default:** assume minimum 80 columns; all output fits within 80 chars
- **Preferred:** 120 columns; repo name column can be wider; paths are not truncated
- **Wide terminals:** no special treatment — output does not reflow
- Long repo names or paths are never truncated in normal output; `--verbose` command strings are truncated at 100 chars with `…` to preserve readability

### CI / Non-TTY Accessibility

- When stdout is not a TTY (pipe, redirect, CI): colour is disabled automatically by `fatih/color`
- All status is conveyed by symbols (`✓`, `✗`, `⏭`) and text — never by colour alone
- `--no-confirm` flag makes the tool fully non-interactive for CI pipelines
- Exit codes are the machine-readable status signal: `0` = all success, `1` = any failure or config error

### Platform Accessibility

- Linux (amd64) and macOS (amd64, arm64) — `filepath` package handles OS path separators
- No Windows support in Phase 1 (forward-compatible: no hardcoded `/` separators)
- UTF-8 symbols (`✓`, `✗`, `⏭`) assumed available; fallback to `[OK]`, `[FAIL]`, `[SKIP]` if a future need arises to support restricted terminals

### Scripting Accessibility

- stdout: progress/info lines only — safe to pipe to `grep`, `tee`, `less`
- stderr: error lines + summary failures only — safe to redirect separately
- Exit code: always reflects overall success/failure — safe to use in shell `&&` chains and `if` statements

### Testing Accessibility

- `io.Reader` injection in `confirm()` allows test scripts to simulate user input without a TTY
- `MockExecutor` allows testing all output patterns without spawning real git processes
- Fixture YAML in `internal/testdata/` tests config validation without real filesystem


---

## Phase 2 — dukh Server Commands & zgard ws status

### Overview

Server lifecycle is managed directly by the `dukh` binary (`dukh start`, `dukh stop`, `dukh status`). Workspace health is queried via `zgard ws status`, which connects to the running `dukh` gRPC server on `localhost:50501` (or the configured address). All output follows the same terminal conventions as `zgard ws` — symbols over color alone, stderr for errors.

---

### `dukh stop`

**Purpose:** Send a Stop RPC to `dukh`, triggering graceful shutdown.

#### Success Output (server was running and shut down)

```
✓ dukh stopped — server acknowledged shutdown
```

#### Error Output (server not reachable)

```
✗ dukh not reachable — connection refused (localhost:50501)
```

Exit code `1`; message written to stderr.

#### Color Scheme

| Element | Color |
|---|---|
| `✓` prefix | Green |
| `✗` prefix | Red |
| `dukh stopped` | Default (white/grey) |
| `not reachable` message | Red |

---

### `zgard ws status`

**Purpose:** Query `dukh` for the current health snapshot and render a colored report. If `dukh` is not running, it is automatically started before querying.

#### Auto-Start Behavior

When the initial gRPC call fails (dukh not running), zgard:
1. Prints `⟳ dukh is not running — starting…` in blue.
2. Runs `dukh start` as a subprocess.
3. Polls the gRPC endpoint for up to 10 seconds.
4. Prints `✓ dukh started` in green, then proceeds with the status report.

```
⟳ dukh is not running — starting…
✓ dukh started (pid 12345)
✓ dukh started

Dukh  running  •  uptime: 0s
...
```

If `dukh start` fails or the server does not become ready within 10 seconds, an error is printed and the command exits with code 1.

#### Flags

| Flag | Description |
|---|---|
| `-n, --name <name>` | Show only the named workspace |
| `--rescan` | Trigger a fresh workspace rescan, wait for completion, then render the report |

#### Standard Output (cached snapshot)

```
Dukh  running  •  uptime: 2h 34m

Workspace: default
  Project: backend
    ✓ api          main → main
    ✗ auth         main → feat/login  (branch mismatch)
    ✗ gateway      (missing)
  Project: infra
    ✓ terraform    main → main

Workspace: secondary
  Project: secondary-project1
    ✓ test-repo-1  main → main

✓ 3 aligned  ✗ 1 drifted  ✗ 1 missing
```

#### Output with `--rescan`

When `--rescan` is set, a blue informational line is printed while waiting, then the result follows:

```
⟳ rescanning workspaces…

Dukh  running  •  uptime: 2h 34m

Workspace: default
  Project: backend
    ✓ api          main → main
    ✓ auth         main → main
    ✗ gateway      (missing)
  ...

✓ 4 aligned  ⚠ 0 drifted  ✗ 1 missing
```

The `⟳ rescanning workspaces…` line is printed in blue immediately before the gRPC call so the user knows the CLI is waiting. The gRPC call uses a 60-second timeout; if the scan exceeds that, an error is printed.

#### Layout Rules

- **Header line:** `Dukh  <status>  •  uptime: <Xh Ym>` — two spaces around status, bullet separator.
- **Workspace header:** `Workspace: <name>` — no indent, blue.
- **Project header:** `  Project: <name>` — 2-space indent, default colour.
- **Repo line:** `    <symbol> <name><padding>  <configured_branch> → <actual_branch>` — 4-space indent, name left-aligned in a fixed column (longest repo name + 2 spaces), arrow shows configured vs actual.
- **Missing repo:** `    ✗ <name><padding>  (missing)` — no branch arrows.
- **Summary line:** `✓ N aligned  ✗ N drifted  ✗ N missing` — rendered after a blank line.

#### Colour Scheme

| Element | Colour |
|---|---|
| `Dukh` label | Bold |
| `running` status | Green bold |
| `stopped` / `unknown` status | Yellow bold |
| `Workspace:` header | Blue bold |
| `Project:` header | Default |
| `✓` aligned symbol | Green |
| Aligned repo name | Default |
| `→ <branch>` when aligned | Green |
| `✗` mismatch symbol | Red |
| Mismatch repo name | Default |
| `→ <actual_branch>` when mismatched | Red |
| `(branch mismatch)` annotation | Red italic |
| `(missing)` annotation | Red italic |
| Summary `✓ N aligned` | Green |
| Summary `✗ N drifted` | Red |
| Summary `✗ N missing` | Red |

#### Uptime Formatting

| Duration | Display |
|---|---|
| < 60 seconds | `Xs` |
| < 1 hour | `Xm Ys` |
| ≥ 1 hour | `Xh Ym` |

#### Error States

| Condition | Output | Exit Code |
|---|---|---|
| dukh not running → auto-start succeeds | `⟳ dukh is not running — starting…` + `✓ dukh started` then normal report | 0 |
| dukh not running → auto-start fails | `✗ auto-start dukh: <reason>` (stderr) | 1 |
| dukh not running → ready timeout | `✗ dukh did not become ready: timeout after 10s` (stderr) | 1 |
| Named workspace not found | `✗ workspace "foo" not found` (stderr) | 1 |
| dukh returns empty snapshot (no workspaces configured) | `(no workspaces configured)` (stdout) | 0 |

#### Non-TTY / CI Behaviour

- When stdout is not a TTY, colour is suppressed (via `fatih/color` auto-detection).
- Symbols (`✓`, `✗`) are always present regardless of TTY state.
- `→` arrow is always ASCII-safe (U+2192) and universally supported in UTF-8 terminals.

#### Column Alignment

Repo name column width is computed as `max(len(repoName)) + 2` within each project block, not globally. This keeps short project lists compact while avoiding misaligned output across large workspaces.

---

## Phase 3 — grazhda Management Script UX

### `grazhda upgrade`

#### Progress Output

The upgrade command prints step-by-step progress so the user can track a potentially long-running build:

```
╔═══════════════════════════════════════╗
║         Grazhda Upgrader              ║
╚═══════════════════════════════════════╝

Pulling latest sources...
✓ Sources updated.
  3 files changed, 42 insertions(+), 5 deletions(-)

Rebuilding binaries...
  [just build output]
  ✓ All modules built successfully

Installing updated binaries...
✓ Binaries installed to: /home/user/.grazhda/bin

╔═══════════════════════════════════════╗
║         Upgrade Successful!           ║
╚═══════════════════════════════════════╝

⚠ If you updated grazhda itself, open a new terminal or re-source your shell profile.
```

When sources are already current:

```
✓ Sources are already up to date.
```

The rebuild still proceeds to ensure installed binaries are consistent with sources.

#### Error States

| Condition | Output | Exit Code |
|---|---|---|
| `$GRAZHDA_DIR/sources` not a git repo | `✗ Sources directory not found or not a git repository: /…/sources` + hint to run installer | 1 |
| Missing build dependency (git/go/just/protoc) | `✗ Required binary not found: <name>` + `Please install all missing dependencies and try again.` | 1 |
| `git pull` fails (network, conflict) | `✗ git pull failed:` + raw git output | 1 |
| `just build` fails | just exits non-zero; error output from just/go is forwarded to stderr | 1 |

#### Color Scheme

| Element | Color |
|---|---|
| Section headers | Blue |
| Success messages (`✓`) | Green |
| Warnings (`⚠`) | Yellow |
| Errors (`✗`) | Red |

---

### `grazhda config --edit`

#### Behaviour

The command silently resolves the editor and opens `config.yaml`:

```bash
$ grazhda config --edit
Opening config with: vim
[vim opens config.yaml]
```

The editor launch message appears briefly before the editor takes over the terminal. When the user exits the editor, control returns to the shell with no further output.

#### Error States

| Condition | Output | Exit Code |
|---|---|---|
| `config.yaml` does not exist | `✗ Config file not found: /…/config.yaml` + hint to run installer | 1 |
| Resolved editor binary not in PATH | `✗ Editor not found in PATH: <name>` + `Update the 'editor' field in config.yaml or set $EDITOR.` | 1 |

#### Editor Resolution Display

The resolved editor name is always shown before launching so the user knows what program opened their config:

```
Opening config with: vim
```

This is particularly useful when the user has not set `editor:` in their config and the fallback chain selects `$EDITOR` or `vi`.

---

### `grazhda help` / Unknown Command

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

Unknown commands print `✗ Unknown command: <cmd>` in red followed by the usage block, then exit with code 1.

---

## Phase 4 — Cross-Repository Operations

### Overview

`ws exec`, `ws stash`, and `ws checkout` follow the same structural output as `ws init` and `ws pull`: workspace → project → repo hierarchy with per-repo status symbols. The only addition is per-repo command output for `ws exec`.

### Output: `ws exec`

Command output captured from each repo is printed indented (6 spaces) below the repo's status line. The status line itself uses the same symbol/name/message format as all other commands.

```
Workspace: default
  Project: backend
    ✓ api          — done
      running tests...
      ok  github.com/acme/api
    ✗ auth         — exit 1: make: No rule to make target 'test'
      make: No rule to make target 'test'. Stop.
    ⏭ gateway      — not present, skipped

✓ 1 executed  ⏭ 1 skipped  ✗ 1 failed
```

For parallel execution, each repo's output is buffered and printed atomically (under the reporter mutex) when that goroutine completes. Lines from different repos never interleave.

### Output: `ws stash`

```
Workspace: default
  Project: backend
    ✓ api          — stashed
    ✓ auth         — stashed
    ⏭ gateway      — not present, skipped

✓ 2 stashed  ⏭ 1 skipped  ✗ 0 failed
```

### Output: `ws checkout`

```
Workspace: default
  Project: backend
    ✓ api          — checked out feature-x
    ✗ auth         — pathspec 'feature-x' did not match any file(s) known to git
    ⏭ gateway      — not present, skipped

✓ 1 checked out  ⏭ 1 skipped  ✗ 1 failed
```

### Dry-Run Format

Identical to existing commands: `[DRY RUN]` inline in the message column.

```
    ✓ api          — [DRY RUN] would exec: make test
    ✓ auth         — [DRY RUN] would exec: make test
```

### Filtering Display

When `--project-name` is specified, only the matching project header and its repos are printed. No change to the output format — absent projects are simply not shown.

### Symbol Vocabulary (unchanged)

| Symbol | Meaning |
|---|---|
| `✓` (green) | Operation succeeded |
| `✗` (red) | Operation failed |
| `⏭` (yellow) | Skipped (repo not on disk) |

---

## Phase 5 — Universal Targeting System

### Default Workspace Warning

When no targeting flags are provided and the command falls back to the default workspace, a blue **info** line is printed to stderr before any operation output:

```
Info: Targeting default workspace: /home/alice/workspaces/default
```

The warning uses the same `internal/color.Yellow` helper already used throughout the codebase. It prints to **stderr** so it does not pollute stdout in pipe/script contexts. It appears **once** per invocation, before the `Workspace:` header line.

The warning is intentionally a warning — not an error — because using the default workspace is a valid and expected workflow. It simply reminds the user of what is being targeted when they did not specify.

### When the Warning Does NOT Appear

| Scenario | Warning? |
|---|---|
| `--name myws` provided | No — explicit |
| `--all` provided | No — explicit |
| No flags, commands using `workspace.Resolve` | **Yes** |
| `ws purge` (no flags) | No — blocked with error instead |
| `ws status` (no flags) | No — defaults to all workspaces, not a single default |

### Targeting Flags Reference (for README draft)

```
Targeting Flags (inherited by all ws subcommands):
  -n, --name <name>           Target a named workspace (default: default workspace)
      --all                   Operate on all workspaces
  -p, --project-name <name>   Filter to a specific project
  -r, --repo-name <name>      Filter to a specific repository (requires --project-name)
```

### `ws purge` Safety UI (unchanged)

`ws purge` without explicit targeting prints a red error and exits 1:

```
ws purge requires --name <name> or --all
```

No warning is shown — purge errors immediately, making the safety contract unambiguous.

## Phase 6 — Workspace Inspection Suite

### ws search Output Format

```
[backend/api] src/main.go:42:     func coolHandler() {
[backend/api] src/main.go:56:     // cool init
[backend/auth] pkg/auth.go:12:    cool := true

3 match(es) across 2 repo(s)
```

Glob mode (`--glob`):

```
[backend/api] cmd/server.go
[backend/api] internal/server/server.go

2 match(es) across 1 repo(s)
```

When `--repo-name` matches multiple repos (yellow warning before results):

```
Warning: --repo-name "service" matches 3 repositories
[backend/api-service] ...
```

### ws diff Output Format

```
Workspace: myws
  Project: backend

    REPO              UNCOMMITTED  AHEAD  BEHIND
    ──────────────────────────────────────────────
    api               3            2      0
    auth-service      0            0      1
    gateway           (not cloned)

  Project: frontend

    REPO  UNCOMMITTED  AHEAD  BEHIND
    ─────────────────────────────────
    web   0            0      0

✓ 1 clean  ✗ 2 dirty  ⏭ 1 not cloned
```

Color coding:
- UNCOMMITTED > 0: red
- AHEAD > 0 or BEHIND > 0: yellow
- All zeros and upstream exists: green
- `(not cloned)`: yellow/skipped symbol

### ws stats Output Format

```
Workspace: myws
  Project: backend

    REPO              LAST COMMIT        30D COMMITS  CONTRIBUTORS
    ──────────────────────────────────────────────────────────────
    api               2024-01-15 10:30   47           8
    auth-service      2024-01-14 09:00   12           3
    gateway           (not cloned)       -            -
```

### Common Rules

- Table headers are printed in UPPERCASE.
- Separator uses Unicode `─` (U+2500) matching column width.
- Column padding: 2 spaces between columns.
- Indent: 4 spaces for repo rows under a project (consistent with existing "4-space repo indent" spec).
- `(not cloned)` entries always use yellow ⏭ symbol prefix in summary counts.

## Phase 7 — Tag-Based Targeting & IDE Integration

### Tag Visibility in ws status

Tags are shown after the repo name in `ws status` output, formatted as `[tag1 tag2]` in dim/grey colour (or plain square brackets if colour is disabled):

```
  ✓ api [backend critical]        — clean (main)
  ✓ auth-service [backend]        — clean (main)
```

### Error: Tag filter matches nothing

```
✗ tag filter [legacy] matched no repositories in workspace "myws"
```

Printed in red to stderr.

### Error: IDE not found

```
✗ VS Code CLI 'code' not found on PATH
  Install VS Code and run: Shell Command → Install 'code' command in PATH
```

Printed in red to stderr.

### ws open output

```
Info: Targeting default workspace: /home/user/ws
Opening 3 repo(s) in VS Code...
  ✓ /home/user/ws/backend/api
  ✓ /home/user/ws/backend/auth-service
  ⏭ /home/user/ws/backend/gateway  — not cloned, skipped
```

### Warning: Many windows

```
Warning: 8 IDE windows will open. Proceed? (press Ctrl+C to abort)
```

Printed when count > 5. Always printed; no interactive prompt — the user can abort before the first window appears.

### Common Rules for Phase 7

- `--tag` appears in help text for all `ws` subcommands via persistent flag inheritance.
- Tag values in YAML are free-form strings; the UX never validates casing.
