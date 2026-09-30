# Working in this repository

## What this directory is right now

`grazhda-spec` currently contains **a plan and four teaching guides, and no
code**. The project described below has not been built yet. It will be built
here, by the repository owner, following `PLAN.md` phase by phase.

Do not scaffold, initialise or generate ahead of the plan. If the owner is on
Phase 4, do not helpfully create `libs/`. The phases are ordered deliberately and
each one has a check that the next depends on.

## Read before doing anything

1. `PLAN.md` — the seventeen phases. Find which one is current before acting.
2. `docs/CONTEXT.md` — what was decided before this repository existed, and why.
   Read it once per session; it explains choices that look arbitrary otherwise.
3. `.specify/memory/constitution.md` — once it exists (Phase 4), it governs
   everything and supersedes this file.
4. `IDEAS.md` — before any work on `molfar` (Phase 15).
5. `MYTHOLOGY.md` — before naming anything new.

## What is being built

Grazhda automates the work of a software engineer. Three cooperating processes
plus a shell installer:

| Component | Binary | Runtime | Role |
| :--- | :--- | :--- | :--- |
| CLI interface | `zgard` | Go 1.26 | the engineer's hands |
| CLI daemon | `dukh` | Go 1.26, gRPC | local eyes and reflexes |
| Long-term process | `molfar` | Java 25 / Spring Boot 4.1 | memory and judgement |
| Toolchain manager | `grazhda` | Bash | installs, upgrades, removes the rest |

Method: spec-driven development with GitHub spec-kit. Every behaviour is
specified before it is written.

## Division of labour — this is not negotiable

The owner owns the constitution, every `spec.md`, and acceptance of the result.
You own drafting plans and task lists, writing code and tests, running the build,
and proposing spec corrections **for approval**.

Never edit a file under `specs/` during an implementation step. If the code and
the specification disagree, say so and ask which one should move. An agent that
quietly rewrites a specification to match its own code destroys the entire method.

## The old project

`/mnt/e/proj/grazhda` is the previous implementation. It is **read-only
reference material** for reaching behavioural parity. Read it freely. Never
modify a file in it, and never copy its source into this repository — porting it
line by line produces specifications written after the fact, which is the
failure this rebuild exists to avoid.

## Rules that bind once code exists

- `libs/` is the domain. It must never import `cobra`, and must never shell out
  except through `executor.Executor`. A test that needs real `git`, real network
  or a real `$HOME` means the design is wrong — fix the design, not the test.
- `cli/`, `daemon/cmd`, `daemon/server` and molfar's controllers are adapters:
  parse, delegate, render. No domain branching.
- Cross-process coupling goes through `contracts/` only. Generated code is
  committed; CI regenerates and diffs it.
- Message strings are part of the contract (`already exists, skipped`,
  `not present, skipped`, `cloned (main)`). Change one only deliberately, and
  update its specification in the same commit.
- Output symbols are fixed: `✓` ok, `✗` failed, `⚠` drifted, `⏭` skipped.
  Errors to stderr, data to stdout, colour degrades under `NO_COLOR`.
- Exit code 0 when nothing failed, 1 when anything did. Multi-target operations
  continue past failures and report them all at the end.
- `ws purge` never falls back to the default workspace. Safety contract.
- Do not run `go mod tidy` casually once the dependency set is pinned.
- Spring Boot 4: **depend on the starter, never the bare library.** Auto-config
  is modular, so `flyway-core` alone gives you the library with no wiring and no
  error at all — silence. See `guides/04-spring-boot.md`.
- Never commit a specification change and a source change in the same commit.
  See `PLAN.md` Appendix E.

## Commands

Once Phase 6 is done: `just build` · `just test` · `just lint` · `just generate`

## Working with the owner

One phase per session. `/clear` between phases — the artefacts on disk are the
memory, and a specification drafted in a context full of implementation detail
will leak implementation detail into the specification.

When you get stuck on a requirement, suspect the requirement. Thrashing usually
means it is ambiguous or contradicts another one. Go and read it rather than
re-prompting yourself.
