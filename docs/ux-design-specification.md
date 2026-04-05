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
- **Safe-by-default purge:** `ws purge` always requires explicit targeting (`--ws` or `--all`) and prompts for confirmation — no way to accidentally purge without intent
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

No external TUI framework (Bubble Tea, tview) for Phase 1. Output is composed of plain-text lines with a consistent prefix grammar. `charmbracelet/log v2` provides colour-aware structured logging that auto-detects TTY and downgrades gracefully in CI/pipe contexts.

**Rationale:**
- Simplest possible dependency surface for a CLI tool
- Colour is additive — every status is conveyed by symbol/prefix first, colour second
- `charmbracelet/log v2` handles TTY detection, colour downsampling, and log level routing automatically
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

Implemented via `charmbracelet/log v2` — colours are auto-detected and downsampled to terminal capability. All colour choices have monochrome fallbacks via symbols.

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
$ zgard ws purge --ws default

This will permanently remove:
  ~/ws/backend/api
  ~/ws/backend/auth-service
  ~/ws/frontend/webapp

Confirm? [y/N]: y

  ✓ Removed ~/ws/backend/api
  ✓ Removed ~/ws/backend/auth-service
  ✓ Removed ~/ws/frontend/webapp

✓ 3 removed

$ zgard ws purge --ws default --no-confirm   # CI / scripting
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
| `--ws <name>` | `-w` | init, purge, pull | Target named workspace |
| `--all` | | init, purge, pull | Target all workspaces |
| `--dry-run` | | init, purge, pull | Preview without executing |
| `--verbose` | `-v` | init, pull | Show rendered commands |
| `--no-confirm` | | purge | Skip Y/N prompt |
| `--parallel` | | init, pull | Run repo ops concurrently |

`--parallel` is intentionally absent from `ws purge` (purge is single-directory; parallelism is meaningless).

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

- When stdout is not a TTY (pipe, redirect, CI): colour is disabled automatically by `charmbracelet/log v2`
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
