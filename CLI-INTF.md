# CLI-INTF — the `zgard` and `dukh` command interface

An analysis of the command-line surface of the v1 implementation at
`/mnt/e/proj/grazhda`, read from source (authoritative) and cross-checked
against the built binaries in its `bin/` and its own `guides/CLI.md`.

Three things in one document: **what the interface is today** (visualised),
**where it is not uniform** (a numbered catalogue with file-and-line evidence),
and **what it should become** (a proposed target surface, with a mapping table
and a migration plan).

Nothing in the old project was modified. Every claim carries a reference in
Appendix A so you can check it rather than take my word.

---

# 1. Executive summary — skim this

## 1.1 The verdict

**The bones are good and the surface is uneven.** The targeting model — five
filter flags mounted once on the `ws` parent, cross-validated before any work
begins, with four *distinct* messages for the four ways a filter can match
nothing — is better than most multi-repository tools manage. `--dry-run` with a
banner, continue-past-failure with a closing summary, `NO_COLOR` and TTY
detection, dynamic shell completion for workspace and tag names, a typed
`ExitError` instead of `os.Exit` inside a handler: all correct, all deliberate.

What is uneven is everything *around* those decisions. The same flag letter
means three different things depending on depth. The same word — `purge` —
names three different destructions. Warnings go to stdout and notices go to
stderr, which is backwards. There is no machine-readable output anywhere, so
nothing here composes into a script except by parsing colour codes. And the
help text is hand-maintained prose rendered through a Markdown engine that
turns `$GRAZHDA_DIR` into `$GRAZHDA_****DIR` on screen.

And the third member of the toolchain barely behaves like the other two. Where
`zgard` and `dukh` are Cobra trees with help, completion, versions and a shared
persistent flag set, `grazhda` is a hand-rolled `case` router that cannot report
its own version, ignores `NO_COLOR`, offers no dry run before deleting your
installation, and carries three reproduced defects — one of which corrupts an
upgrade while it is running.

None of that is hard to fix. Almost all of it is cheaper to fix *now*, in the
specifications of features 002–007, than after seven commands have inherited it.

## 1.2 The whole surface, on one screen

```
zgard                                dukh                grazhda
├── ws        11 subcommands         ├── start           ├── upgrade
│   ├── init      pull      purge    ├── stop            ├── uninstall
│   ├── exec      stash     checkout ├── status          ├── purge
│   ├── list      status    diff     └── scan            └── help
│   └── stats     search
├── config     7 subcommands         globals:            globals:
│   ├── path      validate  list      --no-color          --verbose/-v
│   ├── get       edit      replace   --help              --help
│   └── merge                         --version           (no --version,
└── pkg        4 subcommands          (no --quiet)         no --no-color)
    ├── install   purge
    └── register  unregister         + grazhda-install.sh   no flags; 2 env vars
                                     + grazhda-init.sh      sourced, not run

26 commands · 26 distinct flags · 2 binaries + 3 shell files
0 machine-readable output modes · 1 command documented in three places that
does not exist
```

## 1.3 The ten findings that matter

| # | Severity | Finding | Where |
| :-- | :--- | :--- | :--- |
| **F-01** | 🔴 High | `ws exec` **swallows the child command's flags**: `zgard ws exec go test -v ./...` loses `-v` to zgard's own `--verbose`. No `--` convention is documented or enforced. | `zgard/ws/exec.go:54` |
| **F-02** | 🔴 High | `-v` means `--version` at the root and `--verbose` one level down. The same letter, two meanings, decided by depth. | `zgard/root.go:31` vs every subcommand |
| **F-03** | 🔴 High | `-n/--name` is a *workspace* in `ws`, a *package ref* in `pkg install/purge`, and a *package name* with no shorthand in `pkg unregister`. | `zgard/ws/ws.go:65`, `pkg/install.go:85`, `pkg/unregister.go:76` |
| **F-04** | 🔴 High | **No machine-readable output.** No `--json`, no `--output`, no `--plain`. Every reporting command emits coloured, human-laid-out text only. | everywhere |
| **F-05** | 🟠 Medium | Warnings print to **stdout**, notices print to **stderr** — the opposite of the convention. `--repo-name matches 3 repositories` contaminates piped data. | `internal/workspace/inspect.go:52` vs `zgard/ws/flags.go:45` |
| **F-06** | 🟠 Medium | `ws status` prints `✗` (red) for a **drifted** branch, then counts it under `⚠` in the summary of the same output. | `zgard/ws/status.go:110` vs `:130` |
| **F-07** | 🟠 Medium | `pkg purge` deletes package trees with **no confirmation and no `--dry-run`**, while the equally destructive `ws purge` demands both plus an explicit target. | `zgard/pkg/purge.go` |
| **F-08** | 🟠 Medium | Exit status is only ever **0 or 1**. A usage error, an invalid config, a filter that matched nothing and a failed clone are indistinguishable to a script. | `internal/reporter/reporter.go:196` |
| **F-09** | 🟠 Medium | Help is hand-written Markdown rendered by glamour: `$GRAZHDA_DIR` displays as `$GRAZHDA_****DIR`, tables reflow, width is hard-coded to 100 columns, and the root's catalogue has already drifted from the real command list. | `internal/ui/ui.go:26`, observed output |
| **F-52** | 🔴 High | `grazhda upgrade` **overwrites its own running file with `cp`**. `cp` truncates in place; bash reads scripts lazily by byte offset, so the rest of the upgrade executes corrupted content. Reproduced. | `grazhda:200-205` |
| **F-42** | 🔴 High | **`grazhda config --edit` does not exist.** It is documented in `CLI.md`, in `config.template.yaml`'s comment and in `zgard`'s help; the router knows four commands and this is not one of them. | `grazhda:310-350` |
| **F-53** | 🔴 High | `grazhda upgrade`'s `git pull` error handler is **dead code** — `set -e` exits on the assignment before the status is tested. Reproduced: exit 128, no message. | `grazhda:177-185` |
| **F-10** | 🟡 Low | Documentation contradicts code in six places, including dukh's default port (docs 50051, code **50501**) and its scan-interval config key. | `dukh/cmd/main.go:60` vs `internal/grpcdial/dial.go:13` |

Full catalogue — 67 findings — in [§5](#5-findings--where-the-interface-is-not-uniform).

## 1.4 The seven changes worth making

Ordered by value per unit of effort, not by size.

| Priority | Change | Why | Cost |
| :--- | :--- | :--- | :--- |
| **P0** | `ws exec … -- <cmd>` and pass argv as a list, not a joined string | The command is unusable with any child flag today | S |
| **P0** | Diagnostics to stderr, data to stdout — **all** of them | Makes every command pipeable; one-line fix per site | S |
| **P0** | One flag registry: `-v` = `--verbose` everywhere, `--version` loses its shorthand; `-n` = workspace only | Removes the two collisions a user actually trips over | S |
| **P0** | Four one-line fixes in `grazhda`: self-replace via `mv`, test `git pull` with `if !`, marker-delimited profile block, `set -euo pipefail` | Three are reproduced defects; one corrupts an upgrade mid-run | S |
| **P1** | Enumerated exit codes (0/1/2/3/4/5/130) | Turns `zgard` into something you can write `if` around | M |
| **P1** | `-o, --output text\|json` on every reporting command | The single biggest capability gap; unlocks scripting, and `molfar` needs it | M |
| **P1** | Symmetric safety: `--dry-run` + confirmation + `--yes` on every destructive command | `pkg purge` is currently the sharpest unguarded edge | M |
| **P2** | `-j, --jobs N` replacing boolean `--parallel`; `--clone-delay <duration>` | Concurrency you can tune; unit out of the flag name | M |

`-v` is the letter to fix first for a reason beyond tidiness: `grazhda` already reads it as `--verbose`, so aligning `zgard` and `dukh` to that is the choice that invalidates the fewest working command lines.

Two changes deliberately **not** recommended: renaming `ws`/`pkg` to full words
(you type them constantly, and the abbreviation is not the confusing part), and
renaming `zgard ws status` or `dukh status` (both are already correct under a
noun-verb grammar — the overlap is conversational, not structural).

## 1.5 What this costs, and who owns it

Every change below maps onto a feature specification in `PLAN.md`. None of it
is a separate project.

| Feature | Owns | Findings |
| :--- | :--- | :--- |
| 002 targeting | flag names, shorthands, no-match exit code | F-02, F-03, F-11…F-15 |
| 003 lifecycle | the reporting contract: streams, symbols, summary, exit codes, `--jobs`, `--yes` | F-05, F-08, F-20…F-27 |
| 004 cross-repo | `exec --`, search modes, table contract, `-o json` | F-01, F-04, F-16, F-28…F-31 |
| 005 package registry | `pkg` verb pairs, `pkg list`, dry-run, confirmation | F-07, F-17, F-18, F-19 |
| 006 daemon | dukh's flag set, `--foreground`, `restart`, `logs`, drift symbol | F-06, F-32…F-36 |
| 007 toolchain | the whole shell surface: self-replacement, the profile round trip, `--dry-run`/`-y`, the phantom `config --edit`, `status`/`doctor` | F-11, F-42…F-66 |
| constitution | Principle V gains the flag registry, stream and exit-code rules | all |

---

# 2. How to read this document

- **§3** visualises the interface as it exists: command trees, a flag
  application matrix, argument shapes, stream and exit-code maps.
- **§4** is the per-command reference — synopsis, operands, flags, behaviour,
  output, exit status — for all 22 commands.
- **§5** is the findings catalogue, numbered `F-01`…`F-41`, each with severity,
  evidence and impact. This is the part to argue with.
- **§6** is the best-practice research, with each rule scored against the
  current surface.
- **§7** is the proposed target interface.
- **§8** maps every old command and flag to its replacement.
- **§9** is the migration and compatibility plan.
- **§10** connects it to the spec-driven rebuild.

Conventions in this document: `<operand>` is required, `[optional]`,
`flag…` is repeatable. 🔴 High severity means a user hits it in normal use;
🟠 Medium means it bites in scripts or under load; 🟡 Low is a paper cut or a
documentation defect.

---

# 3. The current interface, visualised

## 3.1 `zgard` — the full command tree

Flags shown against the command that *defines* them. Inherited flags are shown
once, at the level they are mounted.

```
zgard                                            Workspace lifecycle manager
│
│  ── persistent, inherited by everything ───────────────────────────────────
│     --no-color                 disable ANSI colour (overrides NO_COLOR)
│     -q, --quiet                suppress stdout progress; errors still print
│  ── local to the root command only ─────────────────────────────────────────
│     -h, --help
│     -v, --version              ⚠ -v means --verbose one level down
│
├── ws                                                Workspace operations
│   │
│   │  ── persistent, inherited by all 11 subcommands ───────────────────────
│   │     -n, --name <ws>        target one workspace   (default: the default ws)
│   │         --all              every workspace
│   │     -p, --project-name <p> exact project match
│   │     -r, --repo-name <r>    case-sensitive SUBSTRING match; requires -p
│   │     -t, --tag <tag>…       repeatable, OR logic, project tags inherited
│   │
│   │  ── cross-flag rules, checked in PersistentPreRunE before any work ─────
│   │     -r requires -p  ·  --all ⊕ -p  ·  --all ⊕ -r  ·  -n ⊕ --all
│   │
│   ├── init                     clone every repo; skip those present
│   │     --dry-run  -v/--verbose  --parallel
│   │     --clone-delay-seconds <int>
│   │     --no-confirm           ⚠ accepted, but init never prompts (no-op)
│   │
│   ├── pull                     git pull --rebase; skip uncloned
│   │     --dry-run  -v/--verbose  --parallel
│   │
│   ├── purge                    rm -rf the workspace directories
│   │     --dry-run  -v/--verbose  --no-confirm
│   │     ⚠ requires -n or --all — never falls back to default  ← safety contract
│   │     ⚠ inherits -p/-r/-t and silently ignores all three
│   │
│   ├── exec <command> [args…]   fan a shell command across repos
│   │     --dry-run  -v/--verbose  --parallel
│   │     ⚠ child flags are eaten by zgard: `exec go test -v ./...` loses -v
│   │
│   ├── stash                    git stash push; skip clean trees
│   │     --dry-run  -v/--verbose  --parallel
│   │
│   ├── checkout <branch>        git checkout <branch>; skip uncloned
│   │     --dry-run  -v/--verbose  --parallel
│   │
│   ├── status                   repo health, read from the dukh daemon
│   │     --rescan               force a fresh scan and wait for it
│   │     ⚠ inherits -p/-r/-t and silently ignores all three (GAPS.md)
│   │     ⚠ auto-starts dukh unless dukh.autostart: false
│   │
│   ├── search <pattern>         content / filename / regex search
│   │     --glob  --regex  --parallel
│   │     ⚠ --glob and --regex may both be passed; glob silently wins
│   │
│   ├── diff                     table: REPO UNCOMMITTED AHEAD BEHIND
│   │     --parallel  -v/--verbose
│   │
│   ├── stats                    table: REPO LAST-COMMIT 30D CONTRIBUTORS
│   │     --parallel  -v/--verbose
│   │
│   └── list                     tree view + on-disk clone status
│         (no local flags — no --parallel, no --verbose)
│
├── config                          Inspect and validate the configuration
│   │                               resolved from $GRAZHDA_DIR/config.yaml
│   │                               else ~/.grazhda/config.yaml
│   ├── path                     print the resolved config path
│   ├── validate                 report every validation error; exit 1 if any
│   ├── list                     workspaces + projects, from config only
│   ├── get <key>                dotted path, e.g. dukh.port, workspaces.0.name
│   ├── edit                     $editor → $VISUAL → $EDITOR → vi
│   ├── replace --from-file <f>  backup, validate, atomically replace
│   └── merge   --from-file <f>  backup, deep-merge, validate, write
│
└── pkg                             Declarative package manager in $GRAZHDA_DIR
    │                               global registry ∪ local registry, merged
    ├── install                  topological order; env blocks written + sourced
    │     -n, --name <ref>       ⚠ -n is a WORKSPACE one subtree over
    │         --all
    │     -v, --verbose          ⚠ here it means "stream script output"
    │
    ├── purge                    reverse topological order; excise env blocks
    │     -n, --name <ref>  --all  -v/--verbose
    │     ⚠ no --dry-run, no confirmation prompt
    │
    ├── register                 interactive; writes the local registry
    │     (no flags at all — every field is prompted)
    │
    └── unregister               registry only; does NOT uninstall
          --name <name>          ⚠ no -n shorthand here, unlike install/purge
          --version <v>          requires --name
          --all
```

## 3.2 `dukh` — the full command tree

```
dukh                                        Workspace health monitor (daemon)
│
│  ── persistent ─────────────────────────────────────────────────────────────
│     --no-color
│  ── local to root ──────────────────────────────────────────────────────────
│     -h, --help
│     -v, --version
│  ⚠ no --quiet, so the two binaries do not share a global flag set
│
├── start        detach, write $GRAZHDA_DIR/run/dukh.pid, serve gRPC, scan loop
│                  (no flags at all)
│                  ⚠ foreground mode exists but is reachable only via the
│                    undocumented internal env var DUKH_DAEMON=1
│                  ⚠ no --port / --config / --foreground → hard to test
│
├── stop         graceful shutdown over gRPC   (no flags, no --timeout)
│
├── status       process view: ● running (pid, uptime) / ○ not running
│                  (no flags)  ⚠ uses ● ○, not the project's ✓ ✗ ⚠ ⏭
│                  cleans up a stale pid file
│
└── scan         fire-and-forget rescan        (no flags, no --wait)
                   ⚠ `zgard ws status --rescan` does the same thing and waits
```

## 3.3 `grazhda` — the toolchain manager, and its two shell companions

Three shell files, one of which is a command surface and two of which are not.

```
grazhda [--verbose|-v] <command>        Installs, upgrades and removes the rest
│
│  ── global, stripped from ANY argument position ───────────────────────────
│     --verbose, -v      print build and git output in real time
│  ⚠ -v is --verbose HERE and --version in zgard and dukh
│  ⚠ no --version, no --no-color, no --quiet, no --dry-run, no --yes
│
├── upgrade      git pull → verify tools → stop dukh → just build → install
│                  ⚠ overwrites its own running file with `cp` (verified fatal)
│                  ⚠ the git-pull error handler is unreachable under `set -e`
│
├── uninstall    stop dukh → strip profile lines → delete all but config.yaml
│                  ⚠ prompt reads /dev/tty unconditionally; no --yes, no --dry-run
│
├── purge        the same, but deletes config.yaml too
│                  ⚠ third meaning of the verb `purge` in this toolchain
│
├── help | --help | -h        usage
└── (no args)                 usage, exit 0        unknown → usage, exit 1

   ✗ config --edit            DOCUMENTED IN THREE PLACES, DOES NOT EXIST

grazhda-install.sh                       Bootstrap — no command surface at all
│  interface = two environment variables:
│     GRAZHDA_DIR              install root         (default $HOME/.grazhda)
│     LOCAL_GRAZHDA_REPO_DIR   install from a local checkout instead of GitHub
│  ⚠ no flags whatsoever: no --help, --dir, --ref, --yes, --verbose
│  ⚠ no way to install a specific tag or ref — always default-branch HEAD
│  runs when executed OR when piped from a download; prompts via /dev/tty

grazhda-init.sh                          Sourced by the shell profile, not run
│  export GRAZHDA_DIR if unset · prepend $GRAZHDA_DIR/bin to PATH exactly once
│  mkdir -p $GRAZHDA_DIR/pkgs · source .grazhda.env and the legacy env file
│  ✅ idempotent, guarded, and covered by a test
```

**The shape worth noticing.** `zgard` and `dukh` are Cobra command trees with
help, completion, version and a persistent flag set. `grazhda` is a hand-rolled
`case` router with one flag and no version. They are presented to the user as
one toolchain, and the third member does not behave like the first two.

## 3.4 How a flag reaches a command

Cobra's inheritance is doing real work here, and it is the reason the same
letter can mean two things. Read this diagram as "what is in scope at depth N".

```
   depth 0            depth 1             depth 2
   ────────           ────────            ────────
   zgard      ──────► ws         ───────► init
   │                  │                   │
   │ persistent:      │ persistent:       │ local:
   │  --no-color      │  -n --name        │  --dry-run
   │  -q --quiet      │  --all            │  -v --verbose   ◄── collides with
   │                  │  -p --project-name│  --parallel         zgard -v
   │ local (root      │  -r --repo-name   │  --clone-delay-seconds
   │  command only):  │  -t --tag         │  --no-confirm
   │  -h --help       │                   │
   │  -v --version ───┼───────────────────┘  not inherited: a LOCAL flag on the
   └──────────────────┘                      root command, so `zgard ws init -v`
                                             resolves to --verbose, and
                                             `zgard -v` resolves to --version
```

The collision is legal, silent, and exactly the failure the
"one global flag registry" rule exists to prevent. Cobra assigns `-v` to
`--version` automatically whenever the root command has the letter free.

## 3.5 The flag application matrix

The table people actually want: **which flag applies to which command**, and
whether it does anything.

**● applies · ○ accepted but ignored · — not available**

### `zgard ws`

| | init | pull | purge | exec | stash | checkout | status | search | diff | stats | list |
| :--- | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: |
| `-n --name` | ● | ● | ●* | ● | ● | ● | ● | ● | ● | ● | ● |
| `--all` | ● | ● | ● | ● | ● | ● | ● | ● | ● | ● | ● |
| `-p --project-name` | ● | ● | **○** | ● | ● | ● | **○** | ● | ● | ● | ● |
| `-r --repo-name` | ● | ● | **○** | ● | ● | ● | **○** | ● | ● | ● | ● |
| `-t --tag` | ● | ● | **○** | ● | ● | ● | **○** | ● | ● | ● | ● |
| `--dry-run` | ● | ● | ● | ● | ● | ● | — | — | — | — | — |
| `--parallel` | ● | ● | — | ● | ● | ● | — | ● | ● | ● | **—** |
| `-v --verbose` | ● | ● | ● | ● | ● | ● | — | **—** | ● | ● | **—** |
| `--no-confirm` | **○** | — | ● | — | — | — | — | — | — | — | — |
| `--clone-delay-seconds` | ● | — | — | — | — | — | — | — | — | — | — |
| `--rescan` | — | — | — | — | — | — | ● | — | — | — | — |
| `--glob` / `--regex` | — | — | — | — | — | — | — | ● | — | — | — |
| `-q --quiet` | ● | ● | ● | ● | ● | ● | **○** | **○** | **○** | **○** | **○** |
| `--no-color` | ● | ● | ● | ● | ● | ● | ● | ● | ● | ● | ● |

\* `ws purge` requires `-n` or `--all`; it never falls back to the default.

**What the holes tell you.** Four separate patterns of absence, none of them
principled: `list` has no flags at all though it walks the filesystem;
`search` is the only fan-out command with no `--verbose`; `--no-confirm` is
declared where it does nothing and absent where the equivalent guard would
help; and `-q/--quiet` is a root-level promise that five of eleven commands
quietly do not keep.

### `zgard config` and `zgard pkg`

| | path | validate | list | get | edit | replace | merge | ‖ | install | purge | register | unregister |
| :--- | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: |
| operand | — | — | — | `<key>` | — | — | — | ‖ | — | — | — | — |
| `-n --name` | — | — | — | — | — | — | — | ‖ | ● | ● | — | **●** no `-n` |
| `--version` (pkg) | — | — | — | — | — | — | — | ‖ | — | — | — | ● |
| `--all` | — | — | — | — | — | — | — | ‖ | ● | ● | — | ● |
| `-v --verbose` | — | — | — | — | — | — | — | ‖ | ● | ● | — | — |
| `--from-file` | — | — | — | — | — | ● | ● | ‖ | — | — | — | — |
| `--dry-run` | — | — | — | — | — | **—** | **—** | ‖ | **—** | **—** | — | **—** |
| confirmation | — | — | — | — | — | backup | backup | ‖ | — | **none** | — | none |

### `dukh`

| | start | stop | status | scan |
| :--- | :-: | :-: | :-: | :-: |
| `--no-color` | ● | ● | ● | ● |
| `-q --quiet` | **—** | **—** | **—** | **—** |
| any local flag | **—** | **—** | **—** | **—** |

### `grazhda`

| | upgrade | uninstall | purge | help |
| :--- | :-: | :-: | :-: | :-: |
| `-v --verbose` | ● | **—**\* | **—**\* | — |
| `--version` | **—** | **—** | **—** | — |
| `--no-color` / `NO_COLOR` | **—** | **—** | **—** | **—** |
| `--dry-run` | — | **—** | **—** | — |
| `-y --yes` | — | **—** | **—** | — |
| confirmation | none | ● `/dev/tty` | ● `/dev/tty` | — |

\* `-v` is stripped from any argument position and sets `VERBOSE=1` globally,
but only `upgrade` reads it — `uninstall` and `purge` never call `verbose()` or
`run_cmd`, so the flag is accepted and inert on two of the three commands.

## 3.6 Argument shapes — three ways to say "the one thing I am operating on"

```
positional, exactly one     zgard ws checkout <branch>
                            zgard ws search   <pattern>
                            zgard config get  <key>

positional, one or more     zgard ws exec <command> [args…]     ← joined with " "
   (variadic, then                                                and run via sh -c
    shell-evaluated)

required flag               zgard config replace --from-file <path>
                            zgard config merge   --from-file <path>
                            zgard pkg install    --name <ref>
                            zgard pkg unregister --name <name> [--version <v>]
```

Three grammars for one concept. `config get` takes its key positionally while
`config replace` takes its file by flag; `pkg install` takes its package by
flag while `ws checkout` takes its branch positionally. Nothing distinguishes
the cases — the same command family disagrees with itself.

## 3.7 Output streams and exit codes, as they are

```
                        ┌─────────── stdout ───────────┐  ┌──── stderr ────┐
zgard ws init           │ ✓/✗/⏭ per-repo status lines  │  │ Info: default  │
                        │ [DRY RUN] banner              │  │   workspace…   │
                        │ Warning: -r matches N repos  ◄┼──┤ failure detail │
                        │ summary: ✓ n cloned …        │  │ lines          │
                        └───────────────────────────────┘  └────────────────┘
                                     ▲                            ▲
                                     └── warnings on the DATA      └── notices on
                                         stream: wrong way round       the DIAGNOSTIC
                                                                        stream: right

zgard ws status         │ ⟳ rescanning workspaces…      │  │ ✗ connect err  │
                        │ the whole health report       │  │                │
                        └── progress on the data stream ┘

exit codes, all commands:      0  nothing failed
                               1  something failed, config invalid, usage error,
                                  filter matched nothing, daemon unreachable,
                                  regex invalid, workspace not found …
                                  ── one code for eight different situations
```

`ws diff`, `ws stats`, `ws list` and `ws search` **always exit 0**, even when
every repository is missing and even when a search finds nothing. `grep` exits
1 on no match; `zgard ws search` does not, so it cannot stand in an `if`.

## 3.8 The symbol and colour vocabulary

| Symbol | Meaning | Colour | Used by |
| :-: | :--- | :--- | :--- |
| `✓` | success | green | reporter, `ws list`, `ws status`, `config validate`, `pkg` |
| `✗` | failure | red | reporter, `ws list` (not cloned), `ws status` (**and drift**) |
| `⏭` | skipped | yellow | reporter only |
| `⚠` | drifted | yellow | **`ws status` summary only** — never on a row |
| `●` `○` | daemon up / down | green / yellow | `dukh status` only |
| `⟳` | in progress | blue | `ws status --rescan`, dukh autostart |
| `ℹ` | information | blue | `config replace`, `config merge` |
| `→` | mapping | blue | `config list`, `ws status` branch drift |

Seven glyphs, two of them undocumented, and one contradiction: a drifted
repository is drawn `✗` red on its row and counted as `⚠` yellow two lines
later. `CLAUDE.md` in the new repository already fixes the vocabulary at four
symbols — `✓ ✗ ⚠ ⏭` — which makes `●`/`○` and this contradiction spec
violations on day one.

## 3.9 Where configuration comes from

```
   flags  ──────────────────────────────────────────────►  highest precedence
     │
     ├── environment: GRAZHDA_DIR, GRAZHDA_EDITOR, NO_COLOR, VISUAL, EDITOR
     │
     └── $GRAZHDA_DIR/config.yaml   ── else ──►  ~/.grazhda/config.yaml
             editor:                              dukh: {host, port, autostart,
             packages: […]                              monitoring.period_mins}
             workspaces: […]                      zgard: {config: {}}   ← inert
                                                  general: {…}          ← inert
```

Correct precedence, and one gap: there is no `--config <path>` flag on either
binary, so the *only* way to point at a different configuration is the
`GRAZHDA_DIR` environment variable. That is the reason the daemon is awkward to
test, and the reason `zgard` cannot be pointed at a second machine's config.

---

# 4. Detailed command reference

Behaviour is taken from source, not from `guides/CLI.md`; where the two
disagree, the disagreement is noted and appears again in §5.

## 4.1 `zgard ws` — workspace operations

### Shared targeting semantics

All eleven subcommands inherit the same five flags and the same resolution
order. This is the strongest part of the design and worth stating precisely.

**Workspace selection.** `-n` selects one. `--all` selects every one. With
neither, the *default* workspace is used — the first marked `default: true`,
else the first named `default` — and an informational line naming the resolved
path is printed **to stderr**, so it never contaminates piped output.

**Filtering.** `-p` is an exact project-name match. `-r` is a **case-sensitive
substring** match against the full configured repository name, so `-r cool`
matches `ORG/PACK/my-cool-backend-lol`; more than one match is legal and
produces a warning. `-t` is repeatable with OR semantics, and project-level
tags are inherited by every repository in the project.

**The four no-match branches** — genuinely well done, and the thing most tools
collapse into one generic message:

| Condition | Message |
| :--- | :--- |
| named project absent | `project %q not found in workspace %q` |
| repo filter matched nothing (only when `-p` given) | `repository filter %q matched no repositories in project %q` |
| tag filter matched nothing inside a repo-filtered set | `tag filter %v matched no repositories matching %q in project %q` |
| tag filter matched nothing in the workspace | `tag filter %v matched no repositories in workspace %q` |

**Cross-flag constraints**, evaluated in `PersistentPreRunE` before any work:
`-r` requires `-p`; `--all` excludes `-p`; `--all` excludes `-r`; `-n` excludes
`--all` (this last one is checked later, in `workspace.Resolve`, not in the
pre-run hook — an inconsistency in *where* validation happens, see F-13).

---

**`zgard ws init`** — clone every repository in the resolved target.

| | |
| :--- | :--- |
| Operands | none |
| Flags | `--dry-run` `-v/--verbose` `--parallel` `--clone-delay-seconds <int>` `--no-confirm` |
| Behaviour | Project directories are created before cloning starts, so parallel workers never race. A destination that already exists is skipped, not re-cloned. A failed clone removes the partial directory. Continues past failures; reports them all at the end. |
| Output | `✓ <repo> — cloned (main)` / `⏭ <repo> — already exists, skipped` / `✗ <repo> — <last stderr line>` |
| Exit | 0 if nothing failed, 1 if anything did |
| Defects | `--no-confirm` is accepted and does nothing — `init` has no prompt |

**`zgard ws pull`** — `git pull --rebase` on every repository.

| | |
| :--- | :--- |
| Operands | none |
| Flags | `--dry-run` `-v/--verbose` `--parallel` |
| Behaviour | Repositories not present on disk are skipped, not failed |
| Output | `✓ <repo> — pulled` / `⏭ <repo> — not present, skipped` |

**`zgard ws purge`** — remove workspace directories from disk.

| | |
| :--- | :--- |
| Operands | none |
| Flags | `--dry-run` `-v/--verbose` `--no-confirm` |
| Safety | **Requires `-n` or `--all`.** Never falls back to the default workspace. Prompts with the list of directories unless `--no-confirm` or `--dry-run`. |
| Output | prompt → `Confirm? [y/N]:` on stdout; then `✓/✗` rows and a `removed` summary |
| Defects | Inherits `-p`, `-r`, `-t` and ignores all three silently. The confirmation prompt is written to stdout and read from stdin with no TTY check — under a pipe it reads whatever arrives and defaults to "no". |

**`zgard ws exec <command> [args…]`** — fan a shell command across repositories.

| | |
| :--- | :--- |
| Operands | one or more; joined with a single space and executed through `sh -c` |
| Flags | `--dry-run` `-v/--verbose` `--parallel` |
| Behaviour | Output is captured and printed indented beneath each repository's status line. Not-present repositories are skipped. Failures do not stop the run. |
| Defects | **The command's own flags are consumed by cobra.** `zgard ws exec go test -v ./...` runs `go test ./...` with zgard in verbose mode. `--` works by accident (cobra stops parsing) but is documented nowhere. Joining argv with spaces and re-evaluating through `sh -c` means quoting is applied twice. |

**`zgard ws stash`** / **`zgard ws checkout <branch>`** — thin git wrappers.

| | |
| :--- | :--- |
| Operands | `stash`: none · `checkout`: exactly one branch |
| Flags | `--dry-run` `-v/--verbose` `--parallel` |
| Output | `✓ <repo> — stashed` / `✓ <repo> — checked out <branch>` |

**`zgard ws status`** — repository health, read from the `dukh` daemon.

| | |
| :--- | :--- |
| Operands | none |
| Flags | `--rescan` (force a fresh scan, wait up to 60 s) |
| Behaviour | Connects over gRPC; if dukh is unreachable and `dukh.autostart` is true it starts the daemon, waits, and retries. With `autostart: false` it fails explicitly. |
| Output | header with uptime, then a workspace → project → repository tree, then `✓ n aligned  ⚠ n drifted  ✗ n missing` |
| Defects | Ignores `-p`, `-r`, `-t` (GAPS.md). Draws drift as red `✗` but counts it as `⚠`. Writes through package-level `fmt.Printf` rather than the command's writer, so it cannot be captured in a test. Prints `⟳ rescanning…` to stdout. |

**`zgard ws search <pattern>`** — content, filename or regex search.

| | |
| :--- | :--- |
| Operands | exactly one pattern |
| Flags | `--glob` `--regex` `--parallel` |
| Modes | default: case-sensitive substring of file contents · `--glob`: filename glob · `--regex`: Go regular expression against contents |
| Behaviour | `.git` and binary files are skipped. An invalid regex fails before any repository is walked. |
| Output | `[project/repo] file:line: content`, closing `N match(es) across M repo(s)` |
| Exit | **always 0**, even with zero matches |
| Defects | `--glob --regex` together is accepted; glob wins because the regex branch is skipped. Mode selection is two booleans where it is really one enumeration. |

**`zgard ws diff`** / **`zgard ws stats`** — the two tables.

| | |
| :--- | :--- |
| Flags | `--parallel` `-v/--verbose` |
| `diff` columns | `REPO` `UNCOMMITTED` `AHEAD` `BEHIND` |
| `stats` columns | `REPO` `LAST COMMIT` `30D COMMITS` `CONTRIBUTORS` |
| Placeholders | not cloned → `(not cloned)` in every value column · cloned but no upstream → `--` in `AHEAD`/`BEHIND` only |
| Exit | always 0 |

**`zgard ws list`** — declared hierarchy plus on-disk clone status.

| | |
| :--- | :--- |
| Flags | none |
| Behaviour | Stats every resolved destination path. A filter error is printed to stderr and the walk *continues to the next workspace* rather than aborting — unique to this command. |
| Output | `workspace:` / `project:` / `✓ <repo> <path>` or `✗ <repo> (not cloned)`, closing `✓ n cloned  ✗ n not cloned` |
| Defects | No `--parallel` despite doing filesystem work; no `--verbose`; `-q` ignored; documented output in `CLI.md` does not match the implementation |

## 4.2 `zgard config` — configuration inspection

| Command | Operand / flag | Behaviour | Exit |
| :--- | :--- | :--- | :--- |
| `path` | — | print the resolved config path | 0 |
| `validate` | — | print **every** error with its field path | 0 / 1 |
| `list` | — | workspaces, projects, branch, repo counts — config only, no filesystem | 0 |
| `get <key>` | positional | dotted path with numeric indices (`workspaces.0.name`); scalars plain, maps/lists as compact YAML | 0 / 1 |
| `edit` | — | `config.editor` → `$VISUAL` → `$EDITOR` → `vi`; path is shell-quoted | 0 / 1 |
| `replace` | `--from-file <f>` **required** | backup → validate → atomic write; original untouched on failure | 0 / 1 |
| `merge` | `--from-file <f>` **required** | backup → deep-merge (maps key-by-key, slices replaced) → validate → write | 0 / 1 |

**Missing:** there is no `config set <key> <value>`. Changing one value means
opening an editor or writing a patch file — the only two commands that write
take whole files.

## 4.3 `zgard pkg` — declarative package manager

Two registries are merged for every operation: a distribution-managed global
registry replaced on upgrade, and a user-managed local one that survives.
Identity is `(name, version)`; a local entry with the same identity replaces
the global one, and a local entry with the same name but a different version is
an additional entry.

| Command | Flags | Behaviour | Guards |
| :--- | :--- | :--- | :--- |
| `install` | `-n/--name <ref>` `--all` `-v/--verbose` | transitive deps expanded, deterministic topological order; per package: `pre` env block → source → install script → `post` env block → source | none |
| `purge` | `-n/--name <ref>` `--all` `-v/--verbose` | exact reverse order; purge script → remove directory → excise both env blocks | **none** |
| `register` | *(no flags)* | fully interactive; deps chosen by number from the merged list; multi-line fields end at a blank line | n/a |
| `unregister` | `--name` `--version` `--all` | registry only — **does not uninstall** | none |

`--name` accepts `<name>` or `<name>@<version>`; a bare name resolves to the
sole entry, else the unversioned one, else an ambiguity error.

**Missing:** no `pkg list`. You can install, purge, register and unregister
packages, but you cannot ask what is in the registry or what is installed
without reading two YAML files by hand.

## 4.4 `dukh` — the health daemon

| Command | Flags | Behaviour |
| :--- | :--- | :--- |
| `start` | none | re-executes itself with `DUKH_DAEMON=1` in a detached process group, sleeps 300 ms, prints `✓ dukh started (pid N)`. The daemon writes `$GRAZHDA_DIR/run/dukh.pid`, opens gRPC on `localhost:50501`, scans at startup and on the configured period, logs to `$GRAZHDA_DIR/logs/dukh.log`. |
| `stop` | none | gRPC `Stop`; finishes in-flight work, removes the pid file, exits |
| `status` | none | reads the pid file, checks liveness, removes it if stale, then a 1-second gRPC probe for uptime |
| `scan` | none | gRPC `Scan`, fire-and-forget |

**Defects.** No `--config`, `--port` or `--foreground`, so the daemon can only
be run one way and only where `GRAZHDA_DIR` is set — which is exactly why it is
hard to test. No `restart`, though `grazhda upgrade` ends by telling you to
restart it manually. No `logs`, though it writes a rotating log you have to know
the path of. No `--quiet`. `scan` cannot wait; `zgard ws status --rescan` can,
so the same operation has two entry points with different semantics and no
shared flag.

## 4.5 `grazhda` — the toolchain manager

Bash, 357 lines, a hand-rolled `case` router. It installs, upgrades and removes
the other three components, so by construction it cannot depend on them.

### `grazhda upgrade`

| | |
| :--- | :--- |
| Flags | `--verbose`/`-v` (any position) |
| Sequence | refuse if `$GRAZHDA_DIR/sources` is not a git repository → verify `git`, `go`, `just`, `protoc`, reporting **all** missing at once → `git pull` → stop dukh gracefully → `just build` → copy four artefacts into `bin/` → refresh the global package registry → advise that a new terminal may be needed |
| dukh handling | reads `run/dukh.pid`; removes it if stale; `dukh stop`, then polls for up to 10 s, then `SIGKILL`; on a failed `stop`, `SIGTERM` and 2 s |
| Exit | 0, or 1 on a missing sources directory / missing tool / failed build |

**Two verified defects.**

*It overwrites itself while running.* `cp bin/grazhda "$BIN_DIR/grazhda"`
replaces the very file being executed — and `cp` truncates in place, keeping the
inode. Bash reads a script lazily, by byte offset, from a descriptor it still
holds, so the remainder of the upgrade is read out of the new file at the old
offset. Reproduced here with a synthetic script: the run died at
`unexpected EOF while looking for matching '"'` and its remaining steps never
executed. Replacing `cp X Y` with `cp X Y.tmp && mv -f Y.tmp Y` fixes it — an
atomic rename gives the new content a new inode and leaves the running process
reading the old one; verified to both complete the run *and* apply the update.

*The `git pull` error handler is unreachable.* Under `set -e`:

```bash
pull_output=$(git pull 2>&1)     # assignment carries git's exit status
pull_exit=$?                     # ← never reached when git fails
if [ "$pull_exit" -ne 0 ]; then  # ← dead code
    error "git pull failed:"
```

Reproduced: the script exits with git's status (128) and prints nothing. The
carefully written diagnostic never appears. The fix is
`if ! pull_output=$(git pull 2>&1); then …` or a local `set +e` block.

### `grazhda uninstall` and `grazhda purge`

| | `uninstall` | `purge` |
| :--- | :--- | :--- |
| Removes | everything under `$GRAZHDA_DIR` **except `config.yaml`** | `$GRAZHDA_DIR` entirely |
| Rationale | a reinstall reuses your workspace configuration | irreversible, stated in the warning |
| Prompt | `Are you sure…? (y/n)` read from `/dev/tty` | same |
| Guards | `${GRAZHDA_DIR:?}` on both the `find` and the `rm -rf` | same |
| Missing | `--dry-run`, `--yes`, a TTY check, a profile backup | same |

Both call `remove_shell_profile_lines`, which picks `~/.bashrc.user` if it
exists and `~/.bashrc` otherwise, then deletes every line matching
`grazhda-init\.sh` or `export GRAZHDA_DIR=` and copies the result back.

**The round trip does not close.** The installer appends with
`printf '\n%s\n'` — a blank separator line plus two content lines — and the
uninstaller removes only the two content lines. Reproduced:

```
original            after install              after uninstall
──────────          ──────────────             ───────────────
export FOO=1        export FOO=1               export FOO=1
alias ll="ls -l"    alias ll="ls -l"           alias ll="ls -l"
                    ⏎                          ⏎          ← left behind
                    export GRAZHDA_DIR=…
                    [[ -s … ]] && source …
```

Installing twice **is** idempotent — the `grep -q` guard works, verified. But
`PLAN.md` Phase 14 names the byte-identical round trip as the check for this
feature, and v1 fails the second half of it: every install/uninstall cycle
leaves one more blank line in the user's profile.

### `grazhda-install.sh`

| | |
| :--- | :--- |
| Interface | **no flags.** `GRAZHDA_DIR` and `LOCAL_GRAZHDA_REPO_DIR` are the entire surface |
| Invocation | works executed *or* piped from a download — the `BASH_SOURCE`/`$0 == bash` guard is the correct idiom, and prompts read `/dev/tty` so piping still works |
| Sequence | log setup → verify four tools, reporting all missing → prompt if an install exists → clone (or copy a local checkout) → `just build` → copy `bin/*` → seed `config.yaml` **only if absent** → seed the global registry → append the profile snippet if not already present |
| Logging | everything to `$GRAZHDA_DIR/logs/install.log`; on failure, the last 20 lines to stderr; `trap 'on_error $LINENO' ERR` |
| Strictness | `set -eo pipefail` — the management script uses only `set -e`, the test harness `set -euo pipefail` |

The profile snippet is written with the home directory left *unexpanded* —
`export GRAZHDA_DIR="$HOME/.grazhda"` as a literal — so the line stays correct
if the user's home moves. That is a deliberate and good detail.

Three gaps. There is **no `--ref`/`--tag`**, so `git clone` always takes
default-branch `HEAD` and no specific release can be installed. The log is
truncated at the start of every run, so retrying a failed install destroys the
log of the failure you were about to read. And only bash profiles are wired —
`.bashrc.user` or `.bashrc` — although the script carries a Darwin branch for
`sed`, and zsh is macOS's default shell.

### `grazhda-init.sh`

Sixteen lines, sourced from the shell profile, and the most correct file of the
three: `GRAZHDA_DIR` exported only when unset, the `bin` directory prepended to
`PATH` **exactly once** via a `case` guard, `pkgs/` created, and both the legacy
and canonical env files sourced only when readable. It is also the only shell
component whose idempotency is actually covered by a test.

### The Bash test harness

`tests/bash/test_scripts.sh`, four tests, run by `just test-bash`. It sets
`HOME` to a temporary directory — the pitfall `PLAN.md` Phase 14 warns about,
correctly avoided.

What it covers: `grazhda-init.sh` PATH idempotency; the registry copy; config
seeding and preservation; and `cmd_upgrade`'s registry refresh against four fake
tools on `PATH`, with `stop_dukh_if_running` overridden after sourcing — a neat
injection technique.

What it does not cover: **`cmd_uninstall`, `cmd_purge` and
`remove_shell_profile_lines` have no tests at all.** The two destructive
commands and the profile round trip — the exact thing Phase 14 defines as the
acceptance check — are untested, which is why the blank-line defect above is
still present. The tests also leak `HOME`, `GRAZHDA_DIR` and `PATH` between
cases, giving them an undeclared ordering dependency that test 4 works around by
re-exporting `PATH`.

---

# 5. Findings — where the interface is not uniform

Forty-one findings in nine groups. Each carries an ID you can reference from a
specification, a severity, the evidence, and what it costs the user.

## 5.1 What is already right — do not "fix" these

Worth naming first, because a uniformity pass is exactly how good decisions get
flattened by accident.

| # | Keep | Why |
| :--- | :--- | :--- |
| **K-1** | Targeting flags mounted once on the `ws` parent | One definition, one validation site, eleven consumers |
| **K-2** | Cross-flag validation in `PersistentPreRunE`, before any work | The user learns the flags are wrong before a clone starts |
| **K-3** | Four distinct no-match messages | Collapsing these into one generic "no matches" is the single most likely regression of a cleanup pass |
| **K-4** | Default-workspace notice on **stderr** | The one stream decision that is already correct |
| **K-5** | `ws purge` refuses to default | A safety contract, not a convenience |
| **K-6** | Continue past failure, report at the end, exit 1 | Right for multi-target operations |
| **K-7** | Typed `ExitError` instead of `os.Exit` in `RunE` | Keeps exit-code policy in one place |
| **K-8** | `SilenceErrors` + `SilenceUsage` | Errors print once, without a wall of usage text |
| **K-9** | Dynamic completion for `--name`, `--project-name`, `--tag` | Genuinely good, and undocumented in `CLI.md` |
| **K-10** | `NO_COLOR` and non-TTY handled by `fatih/color` | Correct by default, no code required |
| **K-11** | `-r` substring semantics with a multi-match warning | Surprising but useful; the fix is documentation, not behaviour |
| **K-12** | `grazhda`'s `--verbose` accepted in **any** argument position | Exactly what `PLAN.md` Phase 14 requires |
| **K-13** | `check_binary` reports **all** missing tools, not the first | Also a Phase 14 requirement, already met |
| **K-14** | `run_cmd`/`run_logged` are quiet by default but print captured output on failure | A failed step is never silent |
| **K-15** | `${GRAZHDA_DIR:?}` guards every `rm -rf` and `find … -exec rm` | The one place a typo would be unrecoverable |
| **K-16** | Installer seeds `config.yaml` only when absent; profile append is guarded | Reinstall reuses your configuration |
| **K-17** | Installer works when piped from a download, prompting via `/dev/tty` | The correct idiom for `curl … \| bash` |
| **K-18** | `trap 'on_error $LINENO' ERR` with a log tail | Unexpected failures name their line and show context |
| **K-19** | `grazhda-init.sh` prepends `PATH` exactly once, sources only readable files | Idempotent, and the one shell file with a test proving it |
| **K-20** | `stop_dukh_if_running` escalates `dukh stop` → wait 10 s → `SIGKILL`, and clears a stale pid file | The daemon lifecycle handled properly before its binary is replaced |

## 5.2 Grammar and naming

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-11** | 🟡 | `purge` names three unrelated destructions: `ws purge` (delete cloned directories), `pkg purge` (uninstall packages), `grazhda purge` (delete the installation *including* config). `grazhda uninstall` is a fourth, gentler one. | `ws/purge.go`, `pkg/purge.go`, `grazhda` |
| **F-12** | 🟡 | `install`/`purge` and `register`/`unregister` are two verb pairs where only the second is symmetric. `install` should pair with `uninstall`. | `pkg/pkg.go:60` |
| **F-13** | 🟠 | Validation lives in two places: `-r requires -p`, `--all ⊕ -p`, `--all ⊕ -r` in `PersistentPreRunE`; `-n ⊕ --all` in `workspace.Resolve`. Same class of error, two code paths, two moments. | `ws/ws.go:52` vs `internal/workspace/targeting.go:139` |
| **F-14** | 🟡 | Abbreviation policy is unstated: `ws` and `pkg` are abbreviated, `config` is not; flags are fully spelled (`--project-name`) but their nouns are not. | tree |
| **F-15** | 🟡 | `list` names three different things (`ws list` = repos + disk, `config list` = declared hierarchy, `pkg` has none), and the two that exist overlap. | `ws/list.go`, `cfg/cfgcmd.go:130` |
| **F-16** | 🟠 | `pkg` has no `list`. The registry is only inspectable by reading two YAML files. | `pkg/pkg.go:60` |
| **F-17** | 🟠 | `config` has no `set`. The only writes take a whole file (`replace`, `merge`) or an editor. | `cfg/cfgcmd.go:44` |

## 5.3 Flags — the registry problem

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-02** | 🔴 | `-v` is `--version` at the root and `--verbose` at depth 2. Cobra assigns it automatically when the letter is free on the root command. `CLI.md` compounds this by documenting `-v, --version` as a *global* flag; it is root-local and not inherited. | `zgard/root.go:31`, `guides/CLI.md:107` |
| **F-03** | 🔴 | `-n/--name` is a workspace in `ws`, a package ref in `pkg install`/`pkg purge`, and a package name in `pkg unregister` — where the `-n` shorthand is missing entirely. | `ws/ws.go:63`, `pkg/install.go:75`, `pkg/purge.go:63`, `pkg/unregister.go:78` |
| **F-18** | 🟠 | `--verbose` has three meanings: elapsed timings per repository (`ws` mutating), an inspect option (`ws diff`/`stats`), streaming script output (`pkg`). `CLI.md` documents a fourth — "print the rendered command" — which matches none of them. | `ws/run.go:40`, `reporter.go:140`, `pkg/install.go:77`, `CLI.md:640` |
| **F-19** | 🟠 | `--all` means "all workspaces" under `ws` and "all packages" under `pkg`. Defensible — same shape, scoped noun — but never stated, so it reads as a collision. | `ws/ws.go:64`, `pkg/install.go:76` |
| **F-20** | 🟡 | `-n` is conventionally `--dry-run` in Unix tooling (`make -n`, `rsync -n`). Here `--dry-run` has no shorthand and `-n` is taken. | convention |
| **F-21** | 🟠 | `--no-confirm` is declared on `ws init`, which never prompts. A flag that does nothing teaches users the flag set is unreliable. | `ws/init.go:70`, `GAPS.md` |
| **F-22** | 🟡 | `--no-confirm` is a negative boolean naming an affirmative intent. `--yes`/`-y` (or `--force`) is the convention; `--no-*` should negate an existing affirmative flag. | `ws/purge.go:99` |
| **F-23** | 🟠 | `--parallel` is a boolean. There is no way to bound concurrency, so forty repositories mean forty simultaneous clones against one host. | `ws/init.go:68` |
| **F-24** | 🟡 | `--clone-delay-seconds` bakes the unit into the flag name instead of taking a duration. | `ws/init.go:69` |
| **F-25** | 🟠 | No boolean flag has a `--no-` negation, so once a default moves into `config.yaml` there is no way to turn it off for one invocation. | all |
| **F-26** | 🟠 | `-q/--quiet` is a root-level persistent promise honoured only through the reporter. `ws list`, `diff`, `stats`, `search`, `status` and every `config` and `pkg` command ignore it. | `root.go:115`, `run.go:41` |
| **F-27** | 🟠 | `dukh` has `--no-color` but no `--quiet`, so the two binaries expose different global flag sets. | `dukh/cmd/main.go:77` |
| **F-28** | 🟡 | No `--config <path>` on either binary. `GRAZHDA_DIR` is the only way to select a configuration — which is the root cause of the daemon's testability problem. | `internal/path` |
| **F-29** | 🟠 | `ws purge` accepts `-p`, `-r`, `-t` and discards them: `RunOptions` is built without those fields. Not in `GAPS.md`. | `ws/purge.go:78` |
| **F-30** | 🟠 | `ws status` accepts `-p`, `-r`, `-t` and discards them. In `GAPS.md`; decision recorded as *close* in 006. | `ws/status.go:45` |

## 5.4 Arguments

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-01** | 🔴 | `ws exec` parses the child command's flags as its own. `zgard ws exec go test -v ./...` silently drops `-v` into zgard's `--verbose`. `DisableFlagParsing` is not set and `--` is documented nowhere. | `ws/exec.go:54` |
| **F-31** | 🟠 | `ws exec` joins argv with spaces and runs the result through `sh -c`, so quoting is evaluated twice. `exec "make test"` and `exec make test` take different paths to the same place, and `exec echo "a b"` does not do what it looks like. | `ws/exec.go:56`, `internal/executor/executor.go:39` |
| **F-32** | 🟠 | The "one required operand" is positional in `ws checkout`, `ws search` and `config get`, but a required flag in `config replace`, `config merge` and `pkg install`. | §3.6 |

## 5.5 Output streams

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-05** | 🟠 | The multi-match warning goes to **stdout**; the default-target notice goes to **stderr**. Both are diagnostics. The warning contaminates piped data. | `inspect.go:51,227,380`, `workspace.go:117,332`, `exec.go:33` vs `ws/flags.go:45` |
| **F-33** | 🟡 | That warning is duplicated verbatim at six sites. A message string that is part of the contract exists six times. | as above |
| **F-34** | 🟠 | `ws status` writes `⟳ rescanning workspaces…` — a progress message — to stdout. | `ws/status.go:76` |
| **F-35** | 🟠 | `ws status` renders through package-level `fmt.Printf` rather than `cmd.OutOrStdout()`, so its output cannot be captured by a test. Every other command writes to an injected writer. | `ws/status.go:88` |
| **F-36** | 🟡 | `ws purge`'s confirmation prompt is written to stdout and read from stdin with no TTY check. Piped, it consumes a line of the caller's data and answers "no". | `ws/purge.go:71`, `ws/confirm.go` |

## 5.6 Symbols, colour, layout

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-06** | 🟠 | A drifted repository is drawn `✗` red on its row and counted as `⚠` yellow in the summary of the same report. | `ws/status.go:105` vs `:128` |
| **F-37** | 🟡 | `dukh status` uses `●`/`○`, outside the four-symbol vocabulary the rebuild has already fixed in `CLAUDE.md`. | `dukh/cmd/status.go:70` |
| **F-38** | 🟡 | Three layouts render the same workspace → project → repository hierarchy: `ws list`'s tree, `ws status`'s tree, and the `diff`/`stats` tables. Column widths, indents and summary wording all differ. | `ws/list.go`, `ws/status.go`, `internal/format` |

## 5.7 Exit codes

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-08** | 🟠 | Only 0 and 1 exist. Usage error, missing config, invalid config, no-match, failed operation, unreachable daemon and invalid regex are indistinguishable. | `reporter.go:196`, `root.go:100` |
| **F-39** | 🟠 | `ws diff`, `ws stats`, `ws list` and `ws search` always exit 0. A search with no matches succeeds, so it cannot be used in a conditional the way `grep` can. | `ws/search.go:70`, `ws/diff.go:73` |

## 5.8 Machine readability

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-04** | 🔴 | No `--json`, no `--output`, no `--plain`, no `-` stdin/stdout convention. Every reporting command emits coloured, hand-laid-out text. Anything programmatic must parse ANSI. | all |
| **F-40** | 🟠 | This is not only a user-facing gap. `molfar` (feature 008) is specified to drive this substrate; without a stable machine format it will either parse tables or bypass the CLI entirely, and the second choice duplicates the domain. | `PLAN.md` §15 |

## 5.9 Help, and documentation drift

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-09** | 🟠 | Help is hand-written Markdown rendered through glamour. Observed consequences: `**$GRAZHDA_DIR/config.yaml**` renders as `$GRAZHDA_****DIR/config.yaml` and `~/.grazhda` as `~****/.grazhda`, because underscores and tildes are Markdown emphasis. Width is hard-coded to 100 columns regardless of `$COLUMNS`. | `internal/ui/render.go:25`, live `zgard config --help` |
| **F-41** | 🟠 | Because the catalogue in each `Long` string is hand-maintained, it drifts. The root's `ws` table lists ten subcommands and omits `list`; the `config` table in the shipped binary omits `edit`, `replace` and `merge`, which exist in source. A stray `</p>` sits in the `ws` help text. | `root.go:38`, `ws/ws.go:44` |

**Documentation-versus-code drift**, each one a thing a user will believe and be wrong about:

| Claim in docs | Reality in code |
| :--- | :--- |
| dukh gRPC port defaults to **50051** (`dukh/cmd/main.go:64`, `start.go:41`) | **50501** (`internal/grpcdial/dial.go:13`) |
| dukh scan period is `dukh.scan_interval: 30s` (`main.go:65`) | `dukh.monitoring.period_mins` (`internal/config/config.go:57`) |
| `-v, --version` is a global flag (`CLI.md:107`) | root-local; `-v` is `--verbose` everywhere below |
| `--no-confirm` applies to `init` (`CLI.md:640`) | accepted, no effect |
| `--verbose` "prints the rendered command" (`CLI.md:640`) | prints elapsed time, or streams script output |
| `ws list` prints `Summary: 3 cloned, 2 not cloned (5 total)` (`CLI.md:390`) | prints `✓ 3 cloned  ✗ 2 not cloned` |
| `grazhda config --edit` exists (`CLI.md:16`, `config.template.yaml:5`, `zgard config edit` help) | **no such command** — `Unknown command: config`, exit 1 |
| `grazhda`'s usage lists three commands (`CLI.md:14-20`) | four: `upgrade`, `uninstall`, `purge`, `help` |

## 5.10 The toolchain manager

`grazhda`, `grazhda-install.sh` and `grazhda-init.sh`. Three findings in this
group were reproduced rather than inferred; they are marked **verified**.

### Command surface

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-42** | 🔴 | **`grazhda config --edit` does not exist.** It is documented in `guides/CLI.md`, named in `config.template.yaml`'s own comment, and referenced from `zgard config edit`'s help. The router handles `upgrade`, `uninstall`, `purge` and `help` only, so the documented command prints `Unknown command: config` and exits 1. | `grazhda:310-350`, `CLI.md:14-20`, `config.template.yaml:5` |
| **F-43** | 🟡 | `CLI.md`'s quoted usage block is stale in both directions at once: it lists `config --edit`, which is absent, and omits `uninstall` and `purge`, which exist. | `CLI.md:14-20` vs `grazhda:89-107` |
| **F-44** | 🟠 | No `grazhda --version`. The component that installs and upgrades everything cannot report its own version. | `grazhda:310` |
| **F-45** | 🟠 | No `grazhda status` or `doctor`. Nothing answers "what is installed, where, at what version, is dukh running" — the first question after an upgrade, and the one the toolchain manager is best placed to answer. | — |
| **F-46** | 🟡 | `purge` is the third meaning of that verb in one toolchain, sitting next to `uninstall`, from which it differs only by deleting one file. | `grazhda:246`, `:277` |

### Flags

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-47** | 🔴 | `-v` now has **four bindings across one toolchain**: `--verbose` in `grazhda`, `--version` in `zgard`, `--version` in `dukh`, `--verbose` in every `zgard` subcommand. Two meanings, no rule, decided by which binary and which depth. | `grazhda:316`, `zgard/root.go:31` |
| **F-48** | 🟠 | `uninstall` and `purge` have no `--yes`/`--force`, and read the prompt from `/dev/tty` unconditionally. With no controlling terminal — CI, a cron job — `read` fails, `set -e` fires, and the script exits with no message at all. clig's rule is the opposite: fail with the flag that would have answered the prompt. | `grazhda:230`, `:262` |
| **F-49** | 🟠 | No `--dry-run` on `uninstall` or `purge`, the two most destructive operations in the toolchain. `ws purge`, which deletes strictly less, has one. | `grazhda:222-300` |
| **F-50** | 🟠 | The installer has **no flags at all** — no `--help`, `--dir`, `--ref`, `--yes`, `--verbose`. Its entire interface is two environment variables, one of them (`LOCAL_GRAZHDA_REPO_DIR`) undocumented in `CLI.md`. | `grazhda-install.sh:11`, `:239` |
| **F-51** | 🟠 | No way to install a named version: `git clone` always takes default-branch `HEAD`. With F-44 that means you can neither choose nor discover which version you are running. | `grazhda-install.sh:135` |
| **F-52a** | 🟡 | `-v` is stripped globally but only `upgrade` acts on it; on `uninstall` and `purge` it is accepted and inert — the same class of defect as `ws init --no-confirm` (F-21). | `grazhda:311-317` |

### Correctness — reproduced

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-52** | 🔴 | **`grazhda upgrade` overwrites itself with `cp` while running.** `cp` truncates in place and keeps the inode; bash reads a script lazily by byte offset from a descriptor it still holds. **Verified**: a running script replaced by `cp` mid-execution died with `unexpected EOF while looking for matching '"'` and never ran its remaining steps. `cp X Y.tmp && mv -f Y.tmp Y` was verified to complete the run *and* apply the update. | `grazhda:200-205` |
| **F-53** | 🔴 | **The `git pull` error handler is dead code.** `pull_output=$(git pull 2>&1)` carries git's exit status, so `set -e` exits before `pull_exit=$?` is read. **Verified**: exit 128, no message; the `error "git pull failed:"` branch never runs. | `grazhda:177-185` |
| **F-54** | 🟠 | **Install → uninstall is not byte-identical.** Install appends `printf '\n%s\n'`; uninstall removes the two content lines and leaves the blank separator. Every cycle adds one blank line to the user's profile. **Verified**, including that install-twice *is* idempotent. This is precisely the check `PLAN.md` Phase 14 defines for feature 007, and v1 fails half of it. | `grazhda-install.sh:246-249` vs `grazhda:66-84` |
| **F-55** | 🟠 | Profile lines are matched by pattern, not delimited by markers. Any line containing `grazhda-init.sh` or `export GRAZHDA_DIR=` is deleted, including one the user wrote for their own reasons. The convention — sdkman, conda, rustup — is a `# >>> grazhda >>>` … `# <<< grazhda <<<` block removed as a unit. | `grazhda:73-79` |
| **F-56** | 🟠 | The shell profile is rewritten with **no backup**, while `zgard config replace` backs up `config.yaml` before touching it. The more dangerous edit has the weaker guard. | `grazhda:77-80` |
| **F-57** | 🟡 | If filtering leaves the profile empty, the pipeline returns 1 and `set -e` aborts the uninstall *after* the profile edit and *before* the file removal, leaving a half-uninstalled system. Only bites a profile containing nothing but grazhda lines. | `grazhda:75-78` |
| **F-58** | 🟡 | Only bash profiles are wired (`.bashrc.user`, else `.bashrc`). No zsh, no fish — though the installer carries a Darwin branch for `sed`, and zsh is macOS's default shell. | `grazhda-install.sh:20-24`, `grazhda:63-69` |
| **F-59** | 🟡 | The install log is truncated at the start of every run, so retrying a failed install destroys the log of the failure. | `grazhda-install.sh:39` |
| **F-60** | 🟡 | Three strictness levels in one toolchain: `set -e` (`grazhda`), `set -eo pipefail` (installer), `set -euo pipefail` (tests). Neither shipped script sets `-u`. | file headers |

### Output

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-61** | 🟠 | Hardcoded ANSI with **no `NO_COLOR`, no TTY detection and no `--no-color`**, while `zgard` and `dukh` honour all three. `grazhda upgrade > install.log` writes escape codes into the log. | `grazhda:5-9`, `grazhda-install.sh:5-9` |
| **F-62** | 🟡 | Only `error()` writes to stderr. Info, success, warnings — including the destructive-action warnings — and the confirmation prompt all go to stdout. | `grazhda:25-28` |
| **F-63** | ✅ | Symbols are `✓ ✗ ⚠`, already inside the four-symbol vocabulary. No change needed. | `grazhda:25-28` |

### Testing

| # | Sev | Finding | Evidence |
| :--- | :-- | :--- | :--- |
| **F-64** | 🟠 | The harness has **zero coverage of the destructive paths**: `cmd_uninstall`, `cmd_purge` and `remove_shell_profile_lines` are untested. The profile round trip that Phase 14 names as the acceptance check is exactly what is missing — which is why F-54 survives. | `tests/bash/test_scripts.sh:158-165` |
| **F-65** | 🟡 | Tests leak `HOME`, `GRAZHDA_DIR` and `PATH` between cases, creating an undeclared ordering dependency; test 4 re-exports `PATH` to work around it. | `tests/bash/test_scripts.sh:44-48` |
| **F-66** | — | `shellcheck` could not be run here (not installed on this machine); `just lint` includes it, so CI presumably covers style. Everything above comes from reading the source and, for F-52/53/54, from reproducing the behaviour. | — |

---

# 6. What good CLIs do — the research, scored

Sources are listed in Appendix B. The rules below are the ones that actually
bear on this surface; each carries a verdict on the current implementation.

## 6.1 Grammar

> "Pick one subcommand scheme consistently: **noun verb** (`tool repo clone`) —
> scales best when nouns multiply — or **verb noun** (`tool get pods`).
> Bad: `tool repo clone` next to `tool list-users` next to `tool prune repos`.
> Three grammars = user must memorize every command."
> — *SOTA CLI UX rules*

> "If a command has subcommands, the command should function as an area or
> grouping identifier for the subcommands, rather than specify an action."
> — *Microsoft, System.CommandLine design guidance*

> Use conventional verbs: `list, get, create, delete, update, add, remove, run,
> start, stop, status, init, config, login, logout, version, completion, help`.

**Verdict: ✅ mostly compliant, and worth defending.** `zgard <noun> <verb>` is
consistent — `ws init`, `config get`, `pkg install` — and `ws`, `config`, `pkg`
are grouping identifiers that do nothing on their own, exactly as the guidance
asks. The failures are lexical, not structural: `purge` used for three
different destructions (F-11), a verb pair that is not a pair (F-12), and a
missing `list` in the one noun that has none (F-16).

## 6.2 One flag means one thing

> "Same flag, same meaning, everywhere. `-o` must not mean `--output` in one
> subcommand and `--organization` in another. **Maintain a global flag
> registry; audit collisions in review.**" — *SOTA CLI UX rules*

> "Every flag has a long form (`--output`); short forms (`-o`) only for the
> handful used constantly." … "kebab-case (`--dry-run`, not `--dryRun`)"

> Standard names: `-h/--help`, `--version`, `-v` verbose, `-q/--quiet`,
> `-o/--output`, `-f/--force`, `-a/--all`, `-n/--dry-run`, `--no-input`,
> `--json`, `--plain`. — *clig.dev*

**Verdict: ❌ two live collisions (F-02, F-03) and three meanings for
`--verbose` (F-18).** This is the single clearest gap between the current
surface and every guideline consulted, and it is also the cheapest to close.

## 6.3 stdout is data, stderr is everything else

> "**stdout**: primary output and all machine-readable data. **stderr**: log
> messages, errors, warnings, status updates." — *clig.dev*

**Verdict: ⚠️ half right.** The default-target notice is correctly on stderr
(K-4); the multi-match warning and the rescan progress line are on stdout
(F-05, F-34). One rule, applied inconsistently within the same command.

## 6.4 Machine-readable output is a first-class mode

> "Detect TTY to distinguish interactive terminal from piped usage. Default to
> human-readable formatting when writing to a terminal. Provide `--json` for
> structured machine output. Use `--plain` to disable human-friendly formatting
> that breaks line-based tools." — *clig.dev*

> "Changing human-readable output is acceptable **if** users use `--plain` or
> `--json` for scripts — these remain stable." — *clig.dev*

**Verdict: ❌ absent entirely (F-04).** Note the second quote: without a stable
machine format you can never change the human format either, because somebody
is parsing it. Adding `-o json` is what buys the freedom to improve the tables.

## 6.5 Exit codes carry meaning

> "Zero on success; non-zero mapped to failure modes." — *clig.dev*
> "Meaningful exit codes … predictable CLI UX." — *All Days Tech*

**Verdict: ❌ one code for eight situations (F-08), and four commands that
never fail at all (F-39).**

## 6.6 Destructive actions are guarded proportionally

> "Mild changes: optional confirmation. Moderate: prompt `y/n` interactively;
> require `--force` in scripts. Severe: require the user to type the name of the
> thing being deleted, or pass `--confirm=thing-name`."
> "Only prompt when stdin is a TTY. If not a TTY and input is required, fail
> with clear instructions on which flag to pass." — *clig.dev*

**Verdict: ⚠️ excellent in one place, absent in another.** `ws purge` is a
model: explicit target, listed paths, prompt, `--dry-run` (K-5). `pkg purge`
deletes trees and runs purge scripts with neither (F-07). And neither prompt
checks for a TTY (F-36).

## 6.7 Configuration precedence

> "Exactly this chain, highest wins: flags → environment variables →
> project config → user config → system config → built-in defaults."
> — *SOTA CLI UX rules*; the same order in *clig.dev*

**Verdict: ✅ correct, with one gap.** Flags beat `GRAZHDA_*` env vars beat
`config.yaml` beat defaults. The gap is that there is no `--config` flag, so
the top of the chain cannot address the file itself (F-28).

## 6.8 Help is generated, complete, and plain

> "Extensive help on `-h`/`--help` … feature description, usage examples
> (common cases first), common flags, link to full documentation." — *clig.dev*

**Verdict: ⚠️ rich but fragile.** Per-command `Example` blocks are genuinely
good — six to eight worked examples each, common case first. The delivery is
the problem: hand-maintained catalogues that drift (F-41) and a Markdown
renderer that mangles the very environment variables the text is explaining
(F-09).

## 6.9 Arguments versus flags

> "Prefer flags over positional arguments for clarity and future
> extensibility." … "Multiple arguments acceptable only for simple, primary
> operations." — *clig.dev*
> "Rather than passing arguments directly, CLIs should use flags to label
> arguments, so users don't need to memorize argument order." — *Atlassian*

**Verdict: ⚠️ inconsistent (F-32), and one case where the guidance points the
other way.** For `exec`, the operand *is* another program's whole command line,
and the correct convention is the POSIX `--` terminator plus no re-parsing —
which is what `kubectl exec`, `docker run` and `git bisect run` all do.

## 6.10 Future-proofing

> "Keep changes additive when possible (add new flags rather than change
> behavior). Warn before breaking changes: tell users it will change, how to
> adapt now, and detect when they've adapted to silence the warning."
> — *clig.dev*

**Verdict: n/a today, decisive for §9.** It is what makes the migration plan
below a sequence of additive steps with a single, announced breaking release
rather than a rename.

---

# 7. The proposed interface

## 7.1 Ten rules the surface should satisfy

These are written as constraints — each one forbids something specific and
tempting — so they can go into the constitution's Principle V as-is.

| | Rule |
| :-- | :--- |
| **R1** | **One grammar.** Every command is `<binary> <noun> <verb> [operand…] [flags]`. A noun does nothing on its own but list its verbs. |
| **R2** | **One flag registry.** A flag name and a shorthand mean exactly one thing across both binaries. A collision is a build-breaking test failure, not a review comment. |
| **R3** | **Long form always; shorthand rarely.** Every flag has a kebab-case long form. Shorthands are reserved for the nine in §7.2 and are never reassigned. |
| **R4** | **stdout is data; stderr is everything else** — notices, warnings, progress, prompts, errors. No exceptions. |
| **R5** | **Enumerated exit codes.** Seven meanings, listed in §7.6, used by every command. |
| **R6** | **Every reporting command supports `-o json`.** When `-o json` is active, stdout is a single valid document and every diagnostic goes to stderr. |
| **R7** | **Destruction is guarded proportionally**: explicit target, `--dry-run`, a TTY-checked prompt, and `--yes` to skip it. Every destructive command, not some. |
| **R8** | **Flags that can be defaulted from config have a `--no-` negation.** |
| **R9** | **Help is generated from the command tree**, plain text, wrapped to the terminal. No hand-maintained catalogue in a `Long` string. |
| **R10** | **Message strings and symbols are contract.** `✓ ✗ ⚠ ⏭`, the skip and success strings, and the summary line change only deliberately, with the spec updated in the same commit. |

## 7.2 The global flag registry

The whole point of the table is that it is *the* table: one row per shorthand,
valid in both binaries, at every depth.

| Short | Long | Type | Meaning — identical everywhere |
| :-: | :--- | :--- | :--- |
| `-h` | `--help` | bool | help for this command |
| | `--version` | bool | print version and exit — **no shorthand** |
| `-v` | `--verbose` | count | more diagnostics on stderr; repeatable (`-vv`) |
| `-q` | `--quiet` | bool | suppress non-error diagnostics |
| `-o` | `--output` | enum | `text` (default) \| `json` |
| `-n` | `--name` | string | **the workspace name, and nothing else** |
| `-p` | `--project` | string | project filter (alias: `--project-name`) |
| `-r` | `--repo` | string | repository filter (alias: `--repo-name`) |
| `-t` | `--tag` | string… | tag filter, repeatable, OR |
| `-j` | `--jobs` | int | max concurrent workers; `0` = one per CPU |
| `-y` | `--yes` | bool | assume yes; skip confirmation prompts |
| | `--all` | bool | every item of the enclosing noun |
| | `--dry-run` | bool | show what would happen; change nothing |
| | `--no-color` | bool | disable colour |
| | `--no-input` | bool | never prompt; fail if input is required |
| | `--config` | path | configuration file to use |

**Changes from today, and why each one:**

- `--version` loses `-v`; `-v` becomes `--verbose` everywhere. This is the
  convention (`git`, `docker`, `kubectl` all read `-v` as verbosity) and it
  resolves F-02 in the direction that breaks the *fewest* existing habits,
  because eleven subcommands already read `-v` that way.
- `-n` is a workspace, permanently. `pkg install` takes its package as a
  positional operand instead (F-03).
- `--project-name` → `--project`, `--repo-name` → `--repo`, with the old
  spellings kept as hidden aliases forever. Nobody types `-name`.
- `-j/--jobs` replaces boolean `--parallel`, which stays as a deprecated alias
  for `-j 0` (F-23).
- `-y/--yes` replaces `--no-confirm`, which stays as a deprecated alias (F-22).

## 7.3 Proposed `zgard`

```
zgard [global flags] <noun> <verb> [operand…] [flags]

global: -h --help  --version  -v --verbose  -q --quiet  -o --output
        --no-color  --no-input  --config <path>

ws  [-n <ws> | --all] [-p <proj>] [-r <repo>] [-t <tag>…]
├── init         [--dry-run] [-j N] [--clone-delay <dur>]
├── pull         [--dry-run] [-j N]
├── purge         --dry-run | (-n <ws> | --all) [-y]      ← target still mandatory
├── exec         [--dry-run] [-j N] [--shell] -- <cmd> [args…]
├── stash        [--dry-run] [-j N]
├── checkout     [--dry-run] [-j N] <branch>
├── list         [-j N] [--declared]        ← --declared replaces `config list`
├── status       [--refresh] [--timeout <dur>]     ← was --rescan
├── diff         [-j N]
├── stats        [-j N] [--since <dur>]     ← 30d window becomes a parameter
└── search       [-j N] --mode content|name|regex <pattern>

config
├── path
├── validate     [<file>]                   ← optional: check before replacing
├── get          <key>
├── set          <key> <value> [--dry-run]  ← NEW: closes the read/write loop
├── edit
├── replace      <file> [-y]                ← positional, was --from-file
└── merge        <file> [-y]                ← positional, was --from-file

pkg
├── list         [--installed] [--available]        ← NEW
├── install      [<ref>…] [--all] [--dry-run]       ← positional refs, was -n
├── uninstall    [<ref>…] [--all] [--dry-run] [-y]  ← was `purge`
├── register     [--from-file <f>]                  ← non-interactive path added
└── unregister   [<name>[@<version>]…] [--all] [-y] ← positional, was --name
```

**Notable, beyond the flag registry:**

- **`ws exec … -- <cmd>`.** Everything after `--` is passed through untouched,
  as an argv **list**, not a joined string. `--shell` opts back into `sh -c`
  when you genuinely want a pipeline. Fixes F-01 and F-31 together.
- **`ws list --declared`** skips the filesystem and prints the configured
  hierarchy — which is exactly what `config list` does today. One concept, one
  command, a flag for the axis that actually varies (F-15).
- **`ws search --mode`** replaces two booleans whose interaction is currently an
  accident. See §9.3 — `PLAN.md` records a deliberate decision to *carry* the
  glob-wins precedence, and this proposal disagrees with it; the trade-off is
  laid out there rather than assumed here.
- **`ws status --refresh`** and **`dukh scan --wait`** are the same operation
  named consistently from both sides.
- **`config set`** and **`pkg list`** are the two genuinely missing verbs. Both
  are new capability, not parity, and belong in their feature's spec as such.

## 7.4 Proposed `dukh`

```
dukh [global flags] <verb> [flags]        ← same global set as zgard

start    [--foreground] [--port <n>] [--config <path>]
stop     [--timeout <dur>]
restart  [--timeout <dur>]                 ← NEW
status   [-o json]
scan     [--wait] [--timeout <dur>]        ← --wait is new
logs     [-f/--follow] [-n/--lines <n>]    ← NEW
```

Each addition closes a specific defect rather than adding surface for its own
sake:

- **`--foreground`** promotes the internal `DUKH_DAEMON=1` re-exec switch to a
  documented flag. This is the single change that makes the daemon testable:
  a test can start it in-process on an ephemeral port instead of shelling out.
- **`--port`/`--config`** let a test — or a second machine — run a daemon
  without `GRAZHDA_DIR`. Feature 006's specification requires that no test bind
  a fixed port; without these flags that requirement cannot be met.
- **`restart`** exists because `grazhda upgrade` currently ends by telling the
  user to do it by hand.
- **`logs`** exists because the daemon writes a rotating log whose path you
  otherwise have to know.
- **`scan --wait`** gives `dukh` the semantics `zgard ws status --rescan`
  already has.

## 7.5 The output contract

```
stdout   the requested data, and nothing else
         ├─ -o text : symbols, tables, trees, colour when a TTY
         └─ -o json : exactly one JSON document, no colour, no progress

stderr   the default-workspace notice        Info: …
         the multi-match warning             Warning: …
         progress and spinners               ⟳ …
         confirmation prompts                Confirm? [y/N]
         per-failure detail lines            ✗ <repo>: <reason>
         everything --verbose adds
```

Symbols stay at four — `✓` ok, `✗` failed, `⚠` drifted, `⏭` skipped — which
means `dukh status` moves from `●`/`○` to `✓`/`⏭`, and a drifted repository is
drawn `⚠` on its row as well as in its summary (F-06, F-37).

The JSON envelope is versioned so the shape can evolve:

```json
{
  "schema": "grazhda.v1.ws.diff",
  "workspaces": [
    { "name": "default",
      "projects": [
        { "name": "backend",
          "repositories": [
            { "name": "api", "path": "/home/me/ws/backend/api",
              "cloned": true, "uncommitted": 3, "ahead": 2, "behind": 0 }
          ] } ] } ],
  "summary": { "clean": 0, "dirty": 2, "not_cloned": 1 },
  "failures": []
}
```

Three properties worth fixing in the specification: `null` is used for "not
applicable" (no upstream) and the key is absent for "not measured" (not
cloned), so the two placeholders `--` and `(not cloned)` survive the trip into
JSON as distinct facts; ordering follows the configuration, not completion
order, exactly as the text mode requires; and the document is emitted whole at
the end, so a partial run never produces truncated JSON.

## 7.6 The exit-code contract

| Code | Meaning | Example |
| :-: | :--- | :--- |
| `0` | success; nothing failed | every repository cloned or skipped |
| `1` | partial or total failure of the requested work | two of forty clones failed |
| `2` | usage error | `-r` without `-p`; unknown flag; missing operand |
| `3` | configuration missing or invalid | no `config.yaml`; validation errors |
| `4` | nothing matched the target or filter | `-t nosuchtag`; `search` found nothing |
| `5` | a required service is unavailable | dukh not running and autostart disabled |
| `130` | interrupted | Ctrl-C during a fan-out |

`4` is the one that earns its keep: it makes `zgard ws search` usable in a
conditional the way `grep` is, and it distinguishes "your filter was wrong"
from "the work failed" — which today are both `1`.

## 7.7 The safety contract

| Command | Explicit target | `--dry-run` | Prompt | Skip with |
| :--- | :-: | :-: | :-: | :--- |
| `ws purge` | **required** | ✅ | ✅ | `-y` |
| `pkg uninstall` | required unless `--all` | ✅ *(new)* | ✅ *(new)* | `-y` |
| `pkg unregister` | required unless `--all` | ✅ *(new)* | ✅ *(new)* | `-y` |
| `config replace` / `merge` | file operand | ✅ *(new)* | ✅ *(new)* | `-y` |
| `config set` | key + value | ✅ *(new)* | — | — |
| `grazhda uninstall` | — | ✅ *(new)* | ✅ TTY-checked *(new)* | `-y` |
| `grazhda purge` | — | ✅ *(new)* | ✅ **type the directory** *(new)* | `--confirm <dir>` |
| `ws init` / `pull` / `stash` / `checkout` / `exec` | — | ✅ | — | — |

Plus two rules that apply to every prompt: **prompt only when stdin is a TTY**,
and when it is not, fail with the flag that would have answered it — never
consume a line of piped data (F-36). `--no-input` makes that failure explicit
for scripts.

## 7.8 Help

Generated from the command tree, not hand-written prose: drop the glamour
renderer for `--help` and keep per-command `Example` blocks, which are the good
part. Wrap to `$COLUMNS`, fall back to 80. Plain text so `zgard ws --help |
grep -- --parallel` works. Keep the Markdown long-form for `zgard <cmd> --man`
or the generated man pages, where a renderer is appropriate and the width is
the page's, not the terminal's — this preserves the effort already invested
without paying for it on every `--help`.

## 7.9 Proposed `grazhda`

The toolchain manager should look like the tools it installs. Same global flag
set, same symbols, same exit codes, same safety contract.

```
grazhda [global flags] <command> [flags]

global: -h --help  --version  -v --verbose  -q --quiet  -o --output
        --no-color  --no-input  --yes/-y  --dry-run

install     [--ref <tag|branch>] [--dir <path>] [--from-local <dir>] [-y]
upgrade     [--ref <tag|branch>] [--dry-run]
uninstall   [--dry-run] [-y]              keeps config.yaml
purge       [--dry-run] [-y] [--confirm <dir>]   deletes everything
status                                     ← NEW: what is installed, and where
doctor                                     ← NEW: verify the install is sound
config edit                                ← NEW: the documented command, real
version
help
```

**What each change buys.**

- **`-v` becomes `--verbose` and `--version` loses its shorthand** — the same
  resolution as `zgard`, so the letter finally means one thing across all three
  binaries (F-47).
- **`--version`, `status`, `doctor`.** `status` prints the install root, the
  version of each of the four components, whether dukh is running, and where the
  config and registries live. `doctor` checks the four build tools, the profile
  wiring, `PATH`, and that the binaries actually execute. Between them they
  answer the two questions the toolchain manager is uniquely placed to answer and
  currently cannot (F-44, F-45).
- **`install` becomes a subcommand of `grazhda`**, with `grazhda-install.sh`
  kept as the bootstrap that fetches and delegates. That gives the installer a
  flag surface — `--ref`, `--dir`, `--from-local`, `-y` — instead of two
  environment variables (F-50), and makes pinning a release possible (F-51).
- **`--dry-run` and `-y` on `uninstall` and `purge`** (F-48, F-49), with the
  prompt on stderr, a TTY check, and `--no-input` failing loudly instead of
  hanging or dying silently.
- **`purge --confirm <dir>`** applies clig's severity ladder: the irreversible
  operation asks you to name the directory being deleted, rather than accepting
  a bare `y`.
- **`config edit`** stops being a phantom. Either implement it as a thin
  delegation to `zgard config edit`, or delete every reference to it — but not
  the current third state, where three documents describe a command that exits 1
  (F-42).

**And four fixes that are not interface changes at all**, listed here because
they are the highest-severity items in this document:

| Fix | Replaces |
| :--- | :--- |
| `cp X Y.tmp && mv -f Y.tmp Y` for self-replacement | `cp` over the running script (F-52) |
| `if ! pull_output=$(git pull 2>&1); then …` | the unreachable error handler (F-53) |
| Marker-delimited profile block, removed as a unit, with a backup | pattern-matched line deletion (F-54, F-55, F-56) |
| `set -euo pipefail` in every shell file | three strictness levels (F-60) |

**The one test that would have caught three of them**, and which Phase 14
already specifies:

```bash
HOME=$(mktemp -d)                      # never the real one
cp fixture.bashrc "$HOME/.bashrc"
grazhda install --from-local . -y
grazhda install --from-local . -y      # idempotent?
cmp "$HOME/.bashrc" install1.bashrc    # must be identical
grazhda uninstall -y
cmp "$HOME/.bashrc" fixture.bashrc     # must be byte-identical  ← fails today
```


---

# 8. Old → new, every command and flag

## 8.1 Commands

| Today | Proposed | Change |
| :--- | :--- | :--- |
| `zgard ws init` | unchanged | `-j` replaces `--parallel`; `--clone-delay` |
| `zgard ws pull` | unchanged | `-j` |
| `zgard ws purge` | unchanged | `-y` replaces `--no-confirm`; rejects `-p/-r/-t` |
| `zgard ws exec <cmd…>` | `zgard ws exec -- <cmd…>` | `--` required when the child takes flags; argv passed as a list |
| `zgard ws stash` | unchanged | `-j` |
| `zgard ws checkout <br>` | unchanged | `-j` |
| `zgard ws list` | unchanged | gains `-j`, `--declared`, `-o json` |
| `zgard ws status` | unchanged | `--rescan` → `--refresh`; applies `-p/-r/-t`; `-o json` |
| `zgard ws diff` | unchanged | `-j`, `-o json` |
| `zgard ws stats` | unchanged | `-j`, `--since`, `-o json` |
| `zgard ws search <pat>` | unchanged | `--mode` replaces `--glob`/`--regex`; exit 4 on no match |
| `zgard config path` | unchanged | `-o json` |
| `zgard config validate` | `validate [<file>]` | optional operand |
| `zgard config list` | `zgard ws list --declared` | **deprecated**; one concept, one command |
| `zgard config get <key>` | unchanged | `-o json` |
| — | `zgard config set <k> <v>` | **new** |
| `zgard config edit` | unchanged | — |
| `zgard config replace --from-file <f>` | `zgard config replace <f>` | positional |
| `zgard config merge --from-file <f>` | `zgard config merge <f>` | positional |
| — | `zgard pkg list` | **new** |
| `zgard pkg install --name <ref>` | `zgard pkg install <ref>…` | positional, multiple |
| `zgard pkg purge --name <ref>` | `zgard pkg uninstall <ref>…` | **renamed**; gains `--dry-run`, prompt |
| `zgard pkg register` | unchanged | gains `--from-file` for non-interactive use |
| `zgard pkg unregister --name <n> [--version <v>]` | `zgard pkg unregister <name>[@<ver>]…` | positional, `name@version` matching `install` |
| `dukh start` | unchanged | gains `--foreground`, `--port`, `--config` |
| `dukh stop` | unchanged | gains `--timeout` |
| `dukh status` | unchanged | `-o json`; `✓`/`⏭` replace `●`/`○` |
| `dukh scan` | unchanged | gains `--wait`, `--timeout` |
| — | `dukh restart` | **new** |
| — | `dukh logs` | **new** |
| `grazhda upgrade` | unchanged | gains `--dry-run`, `--ref`; self-replacement via `mv` |
| `grazhda uninstall` | unchanged | gains `--dry-run`, `-y`; marker-delimited profile edit with backup |
| `grazhda purge` | unchanged | gains `--dry-run`, `-y`, `--confirm <dir>` |
| `grazhda config --edit` | `grazhda config edit` | **currently non-existent** — implement or delete every reference |
| `grazhda help` | unchanged | — |
| — | `grazhda install` | **new** — `grazhda-install.sh` becomes a bootstrap that delegates |
| — | `grazhda status` | **new** |
| — | `grazhda doctor` | **new** |
| — | `grazhda --version` | **new** |
| `GRAZHDA_DIR=… grazhda-install.sh` | `grazhda install --dir <path>` | env var kept |
| `LOCAL_GRAZHDA_REPO_DIR=… grazhda-install.sh` | `grazhda install --from-local <dir>` | env var kept |

## 8.2 Flags

| Today | Proposed | Kind of change |
| :--- | :--- | :--- |
| `-v, --version` (root) | `--version` | **breaking** — `-v` becomes verbose |
| `-v, --verbose` (subcommands) | `-v, --verbose` everywhere, count-valued | additive |
| `-n, --name` (`ws`) | unchanged | — |
| `-n, --name` (`pkg install`/`purge`) | positional operand | **breaking**, aliased for one release |
| `--name` (`pkg unregister`) | positional operand | **breaking**, aliased |
| `-p, --project-name` | `-p, --project` | additive; old spelling kept as a hidden alias |
| `-r, --repo-name` | `-r, --repo` | additive; old spelling kept as a hidden alias |
| `--parallel` | `-j, --jobs N` | additive; `--parallel` ≡ `-j 0`, deprecated |
| `--no-confirm` | `-y, --yes` | additive; `--no-confirm` deprecated |
| `--clone-delay-seconds N` | `--clone-delay <dur>` | additive; old flag deprecated |
| `--glob` / `--regex` | `--mode name` / `--mode regex` | additive; both-given becomes an error (§9.3) |
| `--rescan` | `--refresh` | additive; `--rescan` deprecated |
| `--from-file <f>` | positional `<f>` | additive; flag deprecated |
| — | `-o, --output text\|json` | **new**, everywhere |
| — | `--no-input` | **new**, both binaries |
| — | `--config <path>` | **new**, both binaries |
| `-q, --quiet` | honoured by **every** command | bug fix |
| `grazhda -v` (verbose) | `-v, --verbose` | unchanged — the *other* binaries move to match it |
| — | `grazhda --version`, `--no-color`, `-y`, `--dry-run`, `-o` | **new**; brings the third binary into the shared flag set |

---

# 9. Migration

## 9.1 The sequence

Additive first, breaking last, in an announced release. Nothing below requires
a flag day.

**Step 0 — the reproduced defects, first and separately.** `grazhda`'s
self-replacement (`mv`, not `cp`), the unreachable `git pull` handler, the
profile round trip, and `set -euo pipefail` across all three shell files. These
are not interface changes, they are corrections, and two of them can corrupt an
upgrade or a user's shell profile — so they ship ahead of everything else, with
the Phase 14 round-trip test added in the same commit.

**Step 1 — pure bug fixes, no interface change.** F-05 warnings to stderr,
F-34 progress to stderr, F-35 write through the command's writer, F-06 drift
symbol, F-29/F-30 stop silently ignoring filters (error rather than discard),
F-21 remove the no-op `--no-confirm` from `init`, and the six documentation
corrections in §5.9. None of this changes a command line that works today.

**Step 2 — additive flags.** `-o/--output`, `-j/--jobs`, `-y/--yes`,
`--project`, `--repo`, `--config`, `--no-input`, `--refresh`, `--mode`,
positional operands *alongside* the existing required flags, `dukh --foreground`
/`--port`/`--restart`/`--logs`, `pkg list`, `config set`. Old spellings keep
working and are marked deprecated in help.

**Step 3 — deprecation warnings.** Every deprecated spelling prints one line to
stderr naming its replacement. clig's rule applies: *tell users it will change,
how to adapt now, and detect when they have adapted so the warning stops.*

**Step 4 — the one breaking release.** `-v` becomes `--verbose` at the root;
`pkg install -n` stops accepting a package; `ws exec` requires `--` before a
command that takes flags; exit codes 2–5 begin to be used. Announce these four
together, in one version, with the deprecation warnings having run for at least
one release before it.

## 9.2 What breaks, honestly

| Change | Who notices | Mitigation |
| :--- | :--- | :--- |
| `-v` at root | anyone scripting `zgard -v` for the version | `--version` has worked all along; deprecation warning first |
| `pkg install -n jdk` | any install script | positional accepted from Step 2; `-n` warns, then errors |
| `ws exec` argv as a list | anyone relying on shell expansion inside `exec` | `--shell` restores the old behaviour exactly |
| exit `2`–`5` | scripts testing `== 1` specifically | rare — most test `!= 0`; called out in release notes |
| `config list` | anyone using it | kept as a deprecated alias for `ws list --declared` indefinitely |

## 9.3 The one place this proposal disagrees with `PLAN.md`

`--glob` and `--regex` may both be given today, and glob wins because the regex
branch is skipped — an ordering artefact, not a decision (`GAPS.md`).
`PLAN.md` Phase 11 and the capability inventory record the resolution as
**Carry**: specify the precedence explicitly rather than erroring, so existing
scripts keep working.

`--mode content|name|regex` is the better interface — one enumeration for one
choice, with an unambiguous error instead of a silent precedence — but adopting
it means overriding a decision the owner already made deliberately.

**Both can be had.** Add `--mode` as the canonical spelling; keep `--glob` and
`--regex` as aliases that set it; and make *supplying both booleans* the only
case that still resolves by precedence, with a deprecation warning naming
`--mode`. Scripts keep working, the ambiguity gets a name, and the carried
behaviour is written down as `PLAN.md` requires — but this is the owner's call,
not mine, and it belongs in feature 004's specification either way.

---

# 10. What this means for the rebuild

The value of doing this analysis *now* is that none of it has been built yet.
Every finding above is a sentence in a specification rather than a refactor.

## 10.1 Constitution — Principle V

Principle V currently reads, in `PLAN.md`'s draft: *fixed symbols, `NO_COLOR`
honoured, errors to stderr and data to stdout.* Three additions make it
enforceable rather than aspirational:

- **One flag registry.** A flag name or shorthand means one thing across both
  binaries. *This is checkable*: a test that walks both command trees and fails
  on a shorthand bound to two different long names.
- **stdout is data; stderr is every diagnostic** — notices, warnings, progress
  and prompts included, not only errors.
- **Every reporting command supports `-o json`**, and under `-o json` stdout is
  a single valid document.

Each of these would sometimes stop you doing something tempting, which is the
test `PLAN.md` sets for a principle.

## 10.2 Which feature owns what

| Feature | Findings it must resolve | The requirement to write |
| :--- | :--- | :--- |
| **002 targeting** | F-02, F-03, F-13, F-19, F-20, F-29, F-30 | The flag registry table (§7.2) as numbered requirements; `-p/-r/-t` on a command that cannot use them is a **usage error**, not a silent discard; all cross-flag validation in one place; exit `4` for no-match |
| **003 lifecycle** | F-05, F-08, F-18, F-22, F-23, F-24, F-25, F-26, F-33, F-36 | The reporting contract: which stream each line class goes to; `-j/--jobs` semantics; `-y`; the exit-code table; prompts require a TTY |
| **004 cross-repo** | F-01, F-04, F-31, F-32, F-38, F-39 | `exec` takes argv after `--` and does not re-shell it; the JSON envelope and its two placeholder rules; one table renderer; `search` exits 4 |
| **005 package registry** | F-07, F-12, F-16, F-17 | `pkg list`; `install`/`uninstall` pairing; dry-run and confirmation on everything destructive |
| **006 daemon** | F-06, F-27, F-28, F-34, F-35, F-37, and the port/interval drift | `--foreground`, `--port`, `--config` — *stated as behaviour*, because they are what make the daemon testable at all; `⚠` for drift; the daemon's own flag set matches `zgard`'s |
| **007 toolchain** | F-11, F-42…F-66 | Self-replacement is atomic; the shell-profile block is marker-delimited, backed up, and removed byte-for-byte; `uninstall`/`purge` take `--dry-run` and `-y`, prompt only on a TTY, and fail loudly under `--no-input`; the installer gains a flag surface including `--ref`; `grazhda` reports its own version; `config edit` either exists or is deleted from every document; `purge` does not become the toolchain's third meaning of the verb |
| **008 molfar** | F-40 | `molfar` consumes `zgard -o json`, not scraped text |

## 10.3 Parity-matrix consequences

Three rows in `docs/parity-matrix.md` need a decision recorded before their
feature is specified, because each one is a deliberate divergence from v1:

| Row | Decision |
| :--- | :--- |
| `ws status` targeting | **Close** — apply `-p/-r/-t` (already recorded in `GAPS.md`) |
| `ws purge` targeting | **Close** — *reject* `-p/-r/-t` as a usage error. Not currently in `GAPS.md`; add it (F-29) |
| `ws search` mode selection | **Carry**, with §9.3's compromise |
| `config list` | **Carry as an alias**, superseded by `ws list --declared` |
| `pkg purge` → `pkg uninstall` | **Close** — rename, alias the old name |
| `grazhda config --edit` | **Decide** — it has never worked. Either implement it (as a delegation to `zgard config edit`) or record it as declined and strike it from `CLI.md` and `config.template.yaml`. Today it is a documented capability with no row in the matrix at all |
| Shell-profile round trip | **Close** — byte-identical after uninstall is the Phase 14 acceptance check and v1 fails it |
| `grazhda upgrade` self-replacement | **Close** — a correctness defect, not a behaviour to carry |

## 10.4 The one test that keeps this true

Everything in §7 decays without a mechanical check. One test, in `cli/`, walking
both command trees:

```
for every command in zgard and dukh:
    every flag has a long form, kebab-case
    every shorthand maps to exactly one long name, across both binaries
    every long name has exactly one help string, across both binaries
    every destructive command declares --dry-run and -y
    every reporting command declares -o
```

That is roughly forty lines, runs in milliseconds, and fails the build the day
somebody adds `-n` to a package command at the end of a long afternoon — which
is exactly how the current surface got here.

`grazhda` is not a Cobra tree, so it cannot be walked the same way. Its
equivalent is a declared flag table in the script plus a test that asserts
`grazhda help` lists exactly the commands the router handles — which is the
check that would have caught the phantom `config --edit` and the stale usage
block on the day either one appeared.

---

# Appendix A — Evidence index

Every claim in this document, by file and line in `/mnt/e/proj/grazhda`.
Read-only reference; nothing there was modified.

| Subject | Location |
| :--- | :--- |
| root command, `-v/--version`, persistent flags | `zgard/root.go:26-31`, `:114-115` |
| glamour help renderer | `zgard/root.go:120`, `internal/ui/render.go:25` |
| hand-written root catalogue | `zgard/root.go:38-101` |
| `ws` parent, targeting flags, cross-flag rules | `zgard/ws/ws.go:20-70` |
| stray `</p>` in help | `zgard/ws/ws.go:44` |
| dynamic completions | `zgard/ws/ws.go:72-130` |
| default-target notice → stderr | `zgard/ws/flags.go:42-48`, `zgard/ws/run.go:33` |
| shared run helper, `--verbose` → elapsed | `zgard/ws/run.go:22-58`, `:40` |
| `init` flags incl. no-op `--no-confirm` | `zgard/ws/init.go:66-71` |
| `purge` explicit-target guard; filters discarded | `zgard/ws/purge.go:46-49`, `:78-83` |
| confirmation prompt, no TTY check | `zgard/ws/purge.go:64-74`, `zgard/ws/confirm.go` |
| `exec` arg handling | `zgard/ws/exec.go:54-57` |
| `status` ignores filters; `✗` for drift; `⚠` in summary; `fmt.Printf` | `zgard/ws/status.go:45`, `:105`, `:128`, `:88` |
| `search` mode precedence | `zgard/ws/search.go:70-77`, `internal/workspace/inspect.go:56` |
| `list`, no flags, continue-on-filter-error | `zgard/ws/list.go:96-101` |
| `config` subcommands, `--from-file` | `zgard/cfg/cfgcmd.go:44-51`, `:262`, `:316` |
| `pkg` flags, `-n` collision | `zgard/pkg/install.go:75`, `purge.go:63`, `unregister.go:78` |
| `pkg` has no confirmation or dry-run | `zgard/pkg/purge.go` (whole file) |
| four no-match branches | `internal/workspace/targeting.go:118-130` |
| `-n ⊕ --all` checked elsewhere | `internal/workspace/targeting.go:139` |
| multi-match warning, six sites, on stdout | `internal/workspace/inspect.go:51,227,380`; `workspace.go:117,332`; `exec.go:33` |
| reporter symbols, summary, exit code | `internal/reporter/reporter.go:118-124`, `:186-193`, `:196` |
| `sh -c` execution | `internal/executor/executor.go:36-39` |
| colour / TTY handling | `internal/color/color.go` |
| default gRPC port 50501 | `internal/grpcdial/dial.go:13` |
| scan period config key | `internal/config/config.go:57`, `config.template.yaml:20-21` |
| dukh root, no `--quiet` | `dukh/cmd/main.go:19-24`, `:77` |
| dukh docs say 50051 / `scan_interval` | `dukh/cmd/main.go:64-65`, `dukh/cmd/start.go:41` |
| `DUKH_DAEMON=1` re-exec | `dukh/cmd/start.go:22`, `:70-95` |
| dukh `●`/`○` | `dukh/cmd/status.go:70-95` |
| autostart behaviour | `dukh/client/client.go:40-63` |
| known gaps, v1's own list | `guides/GAPS.md` |
| documented CLI surface | `guides/CLI.md` |
| `grazhda` router, usage, `-v` stripping | `grazhda:89-107`, `:310-350` |
| colour constants, no NO_COLOR/TTY | `grazhda:5-9`, `grazhda-install.sh:5-9` |
| `run_cmd` quiet-then-show-on-failure | `grazhda:36-50` |
| `check_binary` reports all missing | `grazhda:53-59`, `:163-172` |
| profile line removal, no markers, no backup | `grazhda:63-84` |
| `stop_dukh_if_running` escalation | `grazhda:110-152` |
| `git pull` handler defeated by `set -e` | `grazhda:177-185` |
| self-replacement with `cp` | `grazhda:200-205` |
| `uninstall` / `purge`, `/dev/tty` prompts, `${GRAZHDA_DIR:?}` guards | `grazhda:222-300` |
| installer: env-var-only interface | `grazhda-install.sh:11-16`, `:239-243` |
| installer: profile snippet, unexpanded `$HOME` | `grazhda-install.sh:26-31`, `:246-252` |
| installer: log truncation, ERR trap | `grazhda-install.sh:39`, `:66-74` |
| installer: clone with no ref selection | `grazhda-install.sh:131-142` |
| `grazhda-init.sh` idempotent PATH guard | `grazhda-init.sh:7-10` |
| Bash harness, four tests, no destructive coverage | `tests/bash/test_scripts.sh:158-165` |

# Appendix B — Sources

- [Command Line Interface Guidelines (clig.dev)](https://clig.dev/) — the
  primary reference: philosophy, flags, output streams, TTY detection,
  confirmation proportionality, configuration precedence, future-proofing.
- [SOTA CLI UX — commands, flags and config rules](https://github.com/martinholovsky/sota-skills/blob/main/skills/sota-cli-ux/rules/01-commands-flags-config.md)
  — the global flag registry rule, one-grammar rule, the precedence chain.
- [Command-line design guidance for System.CommandLine (Microsoft)](https://learn.microsoft.com/en-us/dotnet/standard/commandline/design-guidance)
  — parent commands as grouping identifiers; the conventional verb set.
- [10 design principles for delightful CLIs (Atlassian)](https://www.atlassian.com/blog/it-teams/10-design-principles-for-delightful-clis)
  — flags as labels for arguments.
- [Azure CLI command guidelines](https://github.com/Azure/azure-cli/blob/dev/doc/command_guidelines.md)
  — noun-verb consistency at scale.
- [Naming: Best Practices — Building Command-Line Utilities](https://cli-guide.readthedocs.io/en/latest/design/naming.html)
  — subcommand naming.
- [Building a Simple CLI Tool: argument parsing, exit codes, UX basics](https://alldaystech.com/guides/programming-languages/build-simple-cli-tool-argument-parsing-exit-codes)
  — meaningful exit codes.
- [The Poetics of CLI Command Names (smallstep)](https://smallstep.com/blog/the-poetics-of-cli-command-names/)
  — verb ambiguity, which is the `purge` problem.
- In-repository: `/mnt/e/proj/grazhda/guides/CLI.md`, `guides/GAPS.md`, and
  `/mnt/e/proj/grazhda-spec/PLAN.md`, `CLAUDE.md`, `docs/CONTEXT.md`.
