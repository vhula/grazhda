# Grazhda — from zero to a working, AI-driven, spec-driven project

This is the plan you execute. Not a summary of one — a sequence of phases you
work through yourself, with an explanation at each step of what is happening and
why it is worth doing.

**What you are building.** Grazhda automates the work of a software engineer. It
is delivered as three cooperating processes plus a shell installer:

| Component | Binary | Runtime | What it is for |
| :--- | :--- | :--- | :--- |
| CLI interface | `zgard` | Go | the engineer's hands — every interactive command |
| CLI daemon | `dukh` | Go, gRPC | local eyes and reflexes — watches, answers instantly |
| Long-term process | `molfar` | Java 25 / Spring Boot 4.1 | memory and judgement — plans, runs, records, proposes |
| Toolchain manager | `grazhda` | Bash | installs, upgrades and removes the other three |

**How you are building it.** Spec-driven development with
[GitHub spec-kit](https://github.com/github/spec-kit): every behaviour is
specified before it is written, an AI agent drafts the implementation from the
specification, and you accept or reject the result against that specification.

**Before you start, read the guides**, in whichever order suits you:
`guides/01-just.md`, `guides/02-go.md`, `guides/03-spec-kit.md` and — when you
reach Phase 15 — `guides/04-spring-boot.md`. Nothing in this plan assumes you
already know `just`, Go, spec-kit or Spring; the guides are where those are
taught properly. This document assumes you have them open in another window.

If you are driving the loop with Claude Code, section 12 of the spec-kit guide
covers the tool-specific mechanics: where the slash commands come from, what
belongs in `CLAUDE.md` versus the constitution, the permission allow-list worth
committing, and the context discipline that protects specification quality.

---

## How to use this plan

Each phase has the same shape:

- **Goal** — what exists at the end that did not exist at the start.
- **Why** — what this phase buys you, and what goes wrong if you skip it.
- **Do this** — the actual commands and decisions.
- **Check** — how you know the phase is finished.
- **Pitfalls** — what people get wrong here.

Work them in order. Phases 0 to 7 are setup and are mostly linear. Phase 8 is
the teaching phase: one feature taken all the way through the loop, slowly, with
commentary on every gate. Phases 9 to 14 run that same loop six more times, so
they do not re-explain it — instead each one covers what is *different* about
that feature: why it comes when it does, what its specification has to pin down,
and what to look at hardest in review.

Appendix D carries a worked example of every artefact you have to produce cold —
the product definition, the capability inventory, the parity matrix, the
constitution skeleton, and the parts of a specification people get wrong. Read a
phase, then read its template.

### The division of labour between you and the agent

This is the single most important idea in the document, so it goes first.

| You own | The agent owns |
| :--- | :--- |
| The constitution | Drafting `plan.md` from a spec |
| Every `spec.md` — what and why | Drafting `tasks.md` from a plan |
| Deciding a spec is good enough to plan from | Writing the code and the tests |
| Reviewing the plan against the constitution | Running the build and fixing what it breaks |
| Accepting or rejecting the implementation | Proposing spec corrections for your approval |

The agent is fast and tireless and has no stake in whether the result is the
thing you wanted. The specification is how you keep a stake in it. If you let
the agent write the specs as well as the code, you have not adopted spec-driven
development; you have adopted a very confident autocomplete.

The failure mode to watch for, all the way through: the agent quietly edits a
specification so that it matches code it has already written. Treat any change
to a `spec.md` inside an implementation step as a defect until you have read it
and agreed.

### Effort

Rough, for one person working through it without rushing.

| Phase | What | Effort |
| :--- | :--- | :--- |
| 0 | Decide what you are actually building | half a day |
| 1 | Prepare the machine | 1–2 hours |
| 2 | Baseline the old project | half to one day |
| 3 | Initialise spec-kit | 1 hour |
| 4 | Write the constitution | 2–4 hours |
| 5 | Architecture decisions | 2–3 hours |
| 6 | Skeleton and build system | half a day |
| 7 | The daemon contract | 2–3 hours |
| 8 | Feature 001, the full loop | 1–2 days |
| 9–14 | Features 002–007 | 1–2 days each |
| 15 | `molfar` | 1–2 weeks |
| 16 | Close-out | 2–3 days |

The point of phases 0 to 7 taking almost a week before a single feature exists
is that they are what make features cheap afterwards. Resist the urge to start
coding in phase 2.

---

## Phase 0 — Decide what you are actually building

**Goal.** One page of prose, in your own words, that a stranger could read and
correctly predict what the software does.

**Why.** Spec-kit will help you specify the wrong product with tremendous rigour.
It has no opinion about whether the thing is worth building. The old project was
a multi-repository workspace manager; you are re-aiming it at automating an
engineer's work. That is a change of purpose, not a refactor, and it needs to be
settled in your head before it can be settled in a repository.

**Do this.** Create the repository directory and answer five questions in
writing. Do not skip to the tooling.

```bash
mkdir -p /mnt/e/proj/grazhda-spec && cd /mnt/e/proj/grazhda-spec
git init
mkdir -p docs
```

Write `docs/product.md` answering:

1. **Whose work, exactly?** "A software engineer" is too broad to build for. Is
   it you, working across many repositories? A team lead? Someone maintaining
   thirty services?
2. **Which chores?** List ten real tasks you did in the last month that you
   resented. Be specific: "bump the logging library across every backend repo
   and confirm each still builds" beats "dependency management".
3. **What does the system produce?** A pull request? A report? A fixed working
   tree? Decide now — it shapes every interface.
4. **What must it never do?** Push without asking, touch a repository outside
   scope, run for an unbounded time. These become constitutional principles in
   phase 4.
5. **How will you know it worked?** Not "it feels useful" — something you can
   count. Hours saved, chores completed unattended, proposals accepted without
   edits.

Read `IDEAS.md` alongside this. It argues that the answer to question 2 is where
this project lives or dies, and suggests where the genuinely defensible value is.

**Check.** You can describe the product in two sentences without using the word
"AI".

**Pitfalls.** Writing "automate everything an engineer does". That is not a
product definition, it is a mood. Narrow it until it is uncomfortable, then
build that.

---

## Phase 1 — Prepare the machine

**Goal.** Every tool the project needs, installed and verified.

**Why.** Half of the frustration in a polyglot project is discovering a missing
toolchain three layers into a build. Get them all in, confirm each prints a
version, then never think about it again.

**Do this.**

| Tool | Why this project needs it | Verify with |
| :--- | :--- | :--- |
| `git` | source control, and the thing the CLI orchestrates | `git --version` |
| Go 1.26+ | `zgard` and `dukh` | `go version` |
| `just` | the single build entry point (guide 01) | `just --version` |
| `protoc` + Go plugins | generating the daemon contract | `protoc --version` |
| `uv` | installs the spec-kit CLI (guide 03) | `uv --version` |
| Java 25 | `molfar` | `java -version` |
| Maven | `molfar`'s build | `mvn -v` |
| `shellcheck` | keeps the Bash installer honest | `shellcheck --version` |

On Debian/Ubuntu or WSL:

```bash
# Go — do not use the distribution package, it lags
curl -fsSLO https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc && source ~/.bashrc

# just
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh \
  | bash -s -- --to ~/.local/bin

# protoc and the Go plugins
sudo apt-get install -y protobuf-compiler shellcheck
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# Java 25 and Maven — SDKMAN is the least painful route
curl -s "https://get.sdkman.io" | bash && source ~/.sdkman/bin/sdkman-init.sh
sdk install java 25-tem
sdk install maven
```

**Check.** Run every command in the "Verify with" column. All eight print a
version.

**Pitfalls.** `$HOME/go/bin` not on `PATH`, so `protoc-gen-go` is installed but
`protoc` cannot find it — the error is a confusing "program not found or is not
executable". Also: a distribution-packaged Go that is two versions behind will
reject `go.work` files written for 1.26.

---

## Phase 2 — Baseline the old project

**Goal.** `docs/v1-baseline/` containing the old project's documentation, and
two new files: `docs/capability-inventory.md` and `docs/parity-matrix.md`.

**Why.** This is a brownfield rebuild. There is a working implementation at
`/mnt/e/proj/grazhda` whose behaviour is the accumulated result of many
decisions, some deliberate and some accidental. You cannot specify a system you
have not inventoried, and you cannot tell when the rebuild is finished without a
checklist. The parity matrix is your definition of done.

The subtle reason this matters: when you later write a specification, you will
be tempted to transcribe the old code. Do not. A specification describes
*intent*. The inventory is what lets you separate intent from accident — you can
look at a behaviour and ask "was this designed, or did it just end up this way?"

**Do this.**

```bash
cd /mnt/e/proj/grazhda-spec
mkdir -p docs/v1-baseline
cp /mnt/e/proj/grazhda/guides/{CLI,CONFIG,GAPS,QUICK-START,DEVELOPMENT,feature-ideas}.md docs/v1-baseline/
cp /mnt/e/proj/grazhda/config.template.yaml docs/v1-baseline/
cp /mnt/e/proj/grazhda/proto/dukh.proto docs/v1-baseline/
cp /mnt/e/proj/grazhda/README.md docs/v1-baseline/README-v1.md
```

Then build the old binaries and walk every command yourself:

```bash
cd /mnt/e/proj/grazhda && just build
./bin/zgard --help
./bin/zgard ws --help
./bin/zgard ws init --help      # …and so on for every subcommand
./bin/dukh --help
```

Write `docs/capability-inventory.md` as one row per capability: the command, its
flags, what it does, and — the important column — **is this intended behaviour or
an accident?** The old project's own `docs/v1-baseline/GAPS.md` is unusually
honest and already lists five known discrepancies; use it as a starting point
and add your own.

Then write `docs/parity-matrix.md`: the same list, with a column for the owning
feature and separate status columns for **specified**, **domain implemented**,
**CLI implemented** and **tested**, all empty. Appendix D.3 gives the shape.

Splitting implementation into domain and command surface is worth the extra
column. It is the difference between a matrix that says a feature is "done" and
one that can tell you the logic exists but nothing exposes it — which is exactly
the state a rebuild drifts into if you let the adapters lag.

**Check.** Every command in `docs/v1-baseline/CLI.md` appears as a row in the
parity matrix, and you have made a deliberate keep-or-close decision on each
known gap.

**Pitfalls.** Copying the old source into the new repository "for reference".
Don't — you will end up porting it line by line and the specifications will
become fiction written after the fact. The old project stays where it is, read
from but never copied.

---

## Phase 3 — Initialise spec-kit

**Goal.** `.specify/` and the agent integration exist and are committed.

**Why.** Everything after this point runs through spec-kit's templates and
scripts. Getting it in early means every later artefact is created the standard
way rather than hand-rolled and slightly wrong.

**Do this.** Guide `03-spec-kit.md` explains what each part of the output is
for — and its section 12 explains how the installed skills reach Claude Code.
Here you just install it.

```bash
uv tool install specify-cli --from git+https://github.com/github/spec-kit.git
specify --version

cd /mnt/e/proj/grazhda-spec
specify init . --integration claude
```

Then spend twenty minutes reading what it created. At minimum open:

- `.specify/memory/constitution.md` — currently a template; phase 4 fills it
- `.specify/templates/spec-template.md` — the shape every spec will take
- `.specify/templates/plan-template.md` — including the Constitution Check gate
- `.specify/scripts/bash/create-new-feature.sh` — how feature directories are made

```bash
git add -A && git commit -m "chore: initialise spec-kit scaffolding"
```

**Check.** `ls .specify/templates/` lists the spec, plan, tasks and constitution
templates, and your agent's slash commands or skills are installed.

**Pitfalls.** Running `specify init` inside a non-empty directory without
understanding what it will overwrite. Commit first, always.

---

## Phase 4 — Write the constitution

**Goal.** `.specify/memory/constitution.md`, in your words, ratified at 1.0.0.

**Why.** This is the highest-leverage document in the repository and the one
most people write badly. It is the set of constraints that every later decision
is checked against. Every `plan.md` includes a Constitution Check gate; a plan
that violates a principle must either change or justify itself in writing.

A good principle has one property: **it would sometimes stop you doing something
you would otherwise do.** "Write clean code" fails that test — nobody was going
to write dirty code on purpose, so it constrains nothing. "Domain logic never
lives in a command handler; an import of the CLI framework inside the domain
library is a build-visible violation" passes: it forbids a specific, tempting,
otherwise-reasonable shortcut, and you can check it mechanically.

**Do this.** Draft it yourself first, in bullet form, then let the agent turn
your bullets into the document — not the other way round. For this project the
principles worth considering:

1. **Specification is the source of truth.** No behaviour exists until specified.
   Code that contradicts its spec is a defect in one of them. Every requirement
   carries a stable `FR-###` identifier that tests and commits reference.
2. **Three processes, one domain.** Domain logic lives in a shared library, never
   in a command handler, an RPC method or a REST controller. Those are adapters:
   they parse, delegate and render.
3. **Everything declarative and reproducible.** Workspaces, packages and
   automation jobs are declared in YAML and validated before use. Destructive
   operations support a dry run and require an explicit target.
4. **Test-first, contract-first.** Contracts are written before the code that
   satisfies them. Shell-outs go through an injected interface with a
   deterministic fake, so tests never touch real `git`, network or home
   directory.
5. **The terminal is a user interface.** Fixed status symbols, colour that
   degrades under `NO_COLOR`, errors to stderr and data to stdout, a summary line
   on every multi-target operation, exit 0 when nothing failed and 1 when
   anything did.
6. **Automation must be accountable.** Every automated run records what it was
   asked to do, what it did, what changed and why it stopped. The system proposes;
   a human approves. Runs are bounded by explicit scope and budget. Secrets never
   enter logs, prompts or records.
7. **Simplicity, then capability.** Every dependency, process and abstraction is
   justified in the plan's complexity table.

Add a Technology Constraints section (Go 1.26+, Java 25 / Spring Boot 4.1,
`just` as the single entry point, configuration at `$GRAZHDA_DIR/config.yaml`),
a Development Workflow section describing the loop, and a Governance section
with an amendment log and a version number.

**Check.** Read each principle and ask "when would this stop me?" If you cannot
name a case, cut it or sharpen it. Seven sharp principles beat fifteen vague ones.

**Pitfalls.** Writing aspirations instead of constraints. Also: forgetting to
version it. When you later change your mind — and you will — you want an
amendment log showing what changed and why, or the constitution becomes just
another document nobody trusts.

---

## Phase 5 — Decide the architecture

**Goal.** `docs/architecture.md` — the technical design every feature plan will
refer to instead of re-deciding.

**Why.** If each `plan.md` re-litigates the technology, you get eight
inconsistent designs. Decide the cross-cutting things once, write down the
alternatives you rejected and why, and let the feature plans be about features.

**Do this.** Write, at minimum:

- **The three processes** and what each is for, with lifetimes. The argument that
  matters: a CLI is invoked dozens of times an hour, so JVM start-up is
  disqualifying — hence Go. A process that must run for weeks benefits from a
  mature scheduler, persistence and observability stack, and pays start-up once —
  hence Spring Boot. Write that reasoning down; it is the answer to "why two
  languages?" forever after.
- **The repository layout**: `libs/` (shared Go domain), `cli/`, `daemon/`,
  `service/`, `contracts/`, `scripts/`, `specs/`, `docs/`, joined by `go.work`.
- **Layering rules**, stated so they can be checked: the domain knows nothing
  about terminals, flags or transports; adapters do not branch on domain state;
  cross-process coupling goes only through `contracts/`.
- **A technology decision table** with a rejected alternative for every row.
- **Deliberate non-goals** — no multi-tenant operation, no credential storage, no
  autonomous publication.

**Check.** Someone who reads only this file could lay out the repository
correctly and would not have to ask you a single "why is it like that?" question.

**Pitfalls.** Designing for a scale you do not have. This runs on one engineer's
machine.

---

## Phase 6 — Skeleton and build system

**Goal.** An empty but complete repository where `just build` and `just test`
both succeed.

**Why.** Get the loop green before there is anything in it to break. If the build
works when the project does nothing, then the first time it fails you know
exactly what caused it.

**Do this.** Guide `02-go.md` explains modules and workspaces; guide `01-just.md`
explains the Justfile. Here is the sequence.

```bash
cd /mnt/e/proj/grazhda-spec
mkdir -p libs cli daemon service scripts contracts tests/bash bin

cd libs   && go mod init github.com/vhula/grazhda/libs   && cd ..
cd cli    && go mod init github.com/vhula/grazhda/cli    && cd ..
cd daemon && go mod init github.com/vhula/grazhda/daemon && cd ..

go work init ./libs ./cli ./daemon
```

Add intra-repository `replace` directives so each module also builds on its own,
without the workspace. Note what the `cli` → `daemon` dependency is for: the CLI
consumes only the daemon's *generated client*, which happens to live in the
daemon module. It must never reach into the daemon's server internals, and the
architecture document should say so.

```
# cli/go.mod
replace (
	github.com/vhula/grazhda/daemon => ../daemon
	github.com/vhula/grazhda/libs   => ../libs
)
```

Write a minimal `main.go` in `cli/` and `daemon/cmd/` that prints a version and
exits. Then write the Justfile — start with `build-cli` and `build-daemon` only,
and grow it in later phases:

```just
default: help

help:
    @just --list

build-cli:
    @mkdir -p bin
    cd cli && go build -o ../bin/zgard .

build-daemon:
    @mkdir -p bin
    cd daemon && go build -o ../bin/dukh ./cmd

build: build-cli build-daemon

test:
    cd libs && go test ./...
    cd cli && go test ./...
    cd daemon && go test ./...
```

Add a `.gitignore` (at least `/bin/`, `service/target/`, `go.work.sum`) and a CI
workflow that runs `just build` and `just test` on push.

**Check.**

```bash
just build && ./bin/zgard --version && ./bin/dukh --version
just test          # passes trivially; there are no tests yet
```

**Pitfalls.** The one that bites everyone: **each line of a `just` recipe runs in
its own shell**, so `cd cli` on one line does not affect the next. Chain with
`&&` on one line, as above. Guide 01 covers this properly.

---

## Phase 7 — The daemon contract

**Goal.** `contracts/dukh.proto`, generated Go committed at `daemon/proto/`, and
a `just generate` recipe.

**Why.** The contract is the only coupling between the CLI, the daemon and later
the automation service. Writing it before either side exists forces you to decide
what the daemon actually knows, rather than letting the answer emerge from
whatever the implementation happened to expose. Committing the generated code
means a fresh clone builds without `protoc` installed.

**Do this.** Define three operations — report health, trigger a rescan, shut down
gracefully — and the health payload: per workspace, per project, per repository,
with the configured branch, the actual branch, whether the repository exists and
whether the branch is aligned. Include `option java_package` as well as
`option go_package`, because `molfar` will generate a client from the same file
in phase 15.

Add to the Justfile:

```just
generate:
    @mkdir -p daemon/proto
    protoc \
      --go_out=daemon/proto --go_opt=paths=source_relative \
      --go-grpc_out=daemon/proto --go-grpc_opt=paths=source_relative \
      --proto_path=contracts \
      contracts/dukh.proto
```

Make CI run `just generate` and then `git diff --exit-code -- daemon/proto`, so
a stale generated file fails the build rather than drifting quietly.

**Check.** `just generate && just build` succeeds and `git status` is clean
immediately after.

**Pitfalls.** Reserving field numbers. When you later remove a field, mark it
`reserved` so the number is never reused — otherwise an old client and a new
server will disagree about what a field means, silently.

---

## Phase 8 — Feature 001, the full loop

**Goal.** `specs/001-config-core/` complete, the configuration domain
implemented and tested, and the spec reconciled with what shipped.

This is the phase where you learn the loop. Go slowly. Everything after it is
this same sequence with less commentary.

**The feature.** Configuration is the declarative description of the engineer's
working world — workspaces, projects, repositories, daemon settings — plus
everything needed to read, validate, query and safely change it. It is first
because nothing else can run without it.

### 8.1 Specify

```bash
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name config-core --number 1 "configuration core"
```

Then invoke `/speckit-specify` with a description of the feature in your own
words. The agent drafts `specs/001-config-core/spec.md` from the template.

**Now read it properly.** This is the most valuable half-hour in the phase. You
are checking for:

- **Does it say *how* anywhere?** A specification that mentions YAML libraries,
  Go types or file names has failed. It should be readable by someone who does
  not know what language you are using. Send it back if it leaks implementation.
- **Are the user stories independently testable?** Each should be a slice you
  could build alone and demonstrate.
- **Is every requirement numbered and singular?** `FR-012` should assert one
  thing. If it contains "and", split it.
- **Are the success criteria measurable?** "Fast" is not; "a config of 100
  repositories loads and validates in under 100 ms" is.
- **Are the edge cases real?** An empty config file, a config with no default
  workspace, a repository name containing slashes, a backup that cannot be
  written. If the edge-case list is generic, the agent has not thought about
  your domain.

### 8.2 Clarify

Run `/speckit-clarify`. It finds under-specified areas and asks you structured
questions. Answer them — this is you doing the work only you can do. Every
`NEEDS CLARIFICATION` marker must be gone before planning starts; a marker that
survives into implementation becomes an assumption the agent makes silently.

### 8.3 Plan

Run `/speckit-plan`. Now technology is allowed to appear. The output includes a
**Constitution Check** section, and this is a real gate, not decoration. Read it
and ask: does anything here violate a principle? If it does, the plan changes or
the violation gets justified in writing in the complexity table. If you wave one
through, the constitution is dead from that moment on.

### 8.4 Tasks

Run `/speckit-tasks`. The output is grouped by user story rather than by layer,
so each group is independently completable. Skim for tasks with no file path or
no acceptance signal — those are where the agent is guessing.

### 8.5 Implement

Run `/speckit-implement`. Then, before reading a line of the code:

```bash
just build && just test
```

Then review the code against the spec, not against your taste. Useful questions:

- Is there a test for each `FR-###`? Grep for the identifiers.
- Does the domain code import anything it should not?
- Does any test touch the real filesystem outside a temporary directory, or run
  a real `git`?
- Did any `spec.md` change during implementation? If so, read the diff carefully.

### 8.6 Reconcile

The step everyone skips. The implementation always teaches you something the
spec got wrong. Update `spec.md` so it describes what actually shipped, change
its status from `Draft` to `Implemented`, and tick the rows in
`docs/parity-matrix.md`.

A repository whose specifications lie is worse than one with no specifications,
because people trust it.

```bash
git add -A && git commit -m "feat(001): configuration core

Implements FR-001..FR-034 of specs/001-config-core/spec.md."
```

**Check.** `just test` green; every `FR-###` in the spec traceable to a test;
parity matrix updated; spec status `Implemented`.

**Pitfalls.** Accepting a spec you have only skimmed. Everything downstream
inherits its mistakes, and the agent will implement a vague requirement
confidently rather than asking again.

---

## Phases 9–14 — The rest of parity

Six features, the same loop each time. By now you know the moves, so what follows
is not a re-explanation of the loop — it is, for each feature, the thing that is
different: why it comes when it does, what its specification has to pin down that
a loose one would miss, and what to look at hardest in review.

Each of these phases delivers **both** halves of its feature: the domain code in
`libs/` and the command surface in `cli/` that exposes it. Do not defer the CLI
to a later pass. A feature whose domain is written but whose command does not
exist cannot be exercised by hand, and a feature you cannot exercise by hand is a
feature whose specification you have not really tested.

### Why this order

The dependencies are real, not aesthetic:

| Feature | Depends on | Because |
| :--- | :--- | :--- |
| 001 configuration | — | nothing can run without a loaded, validated config |
| 002 targeting | 001 | it filters the structures 001 defines |
| 003 lifecycle | 001, 002 | it operates on the repositories 002 selects |
| 004 cross-repo | 003 | it reuses the fan-out and reporting 003 establishes |
| 005 package registry | 001 | independent of workspaces; only needs paths and config |
| 006 daemon | 001, 002 | it must locate repositories exactly as 003 would |
| 007 toolchain manager | all | it installs the binaries the others produce |
| 008 `molfar` | all | it drives the whole substrate |

005 is the one you could move. It touches nothing the workspace features touch,
so if you want a change of scenery after three workspace features, do it earlier.
Everything else is ordered by necessity.

---

### Phase 9 — Feature 002, workspace targeting

**Goal.** `specs/002-targeting/` complete; the targeting and destination-path
rules implemented in `libs/workspace`; the shared targeting flags mounted on the
`zgard ws` parent command.

**Why now.** Every command in features 003, 004 and 006 inherits this flag set.
Specify it once, here, and the later features are about what they *do* rather
than about which repositories they do it to. Get it wrong and you re-litigate the
same six flags five more times.

**What the specification must pin down.**

*Workspace selection.* `--name`/`-n` selects one; `--all` selects every one; with
neither, the default workspace is used and an informational line naming the
resolved path is printed **to stderr** — so it never contaminates piped output.
The default is the first workspace marked `default: true`, or failing that the
first one named `default`, and when neither exists the error must name `--name`
as the way out.

*Cross-flag constraints,* checked before any work begins: `--name` and `--all`
are mutually exclusive; `--repo-name` requires `--project-name`; `--all` combines
with neither `--project-name` nor `--repo-name`. Each violation exits immediately
with its own message.

*Filtering.* `--project-name` is an exact match. `--repo-name` is a
case-sensitive **substring** match against the full configured repository name,
regardless of structure mode — so `cool` matches `ORG/PACK/my-cool-backend`, and
the spec should say so with that example, because "matches a repository" is
otherwise read as exact by every reader. `--tag` is repeatable with OR semantics,
and project-level tags are inherited by every repository in the project, merged
with the repository's own, deduplicated, order preserving.

*The four outcomes of a filter,* which is where a loose spec does real damage.
Matching several repositories is legal and produces a warning before work starts.
Matching none is an error — but the spec must say *which* none, because the
conditions differ: a named project that does not exist; a repository filter that
matched nothing, which only reports as such when a project name was also given;
a tag filter that matched nothing within a repository-filtered set; and a tag
filter that matched nothing in the workspace at all. Four branches, four
messages. Write them out.

*Destination resolution.* An explicit `local_dir_name` always wins. Otherwise
`tree` keeps the full configured name so slashes become nested directories, and
`list` uses the last `/`-delimited segment with a trailing `.git` stripped. Two
repositories colliding on the same last segment in `list` mode are resolved by
the engineer with `local_dir_name`; without it the second is reported as already
existing. Structure precedence is project, then workspace, then `tree`.

*Path expansion.* A leading `~` expands to the home directory. A `~` anywhere
else does not.

**What to watch in review.** Whether the four no-match branches survived as four
distinct errors or got collapsed into one generic "no matches" — collapsing them
is the single most likely simplification, and it makes the tool much worse to
use. And whether targeting stayed **pure**: it validates configuration structure
only and must never touch the filesystem. A repository that is merely not cloned
yet is not a targeting error.

**Check.** Table-driven tests covering every flag combination and all four
no-match branches. By hand: `zgard ws list -t backend` narrows correctly, a
filter matching nothing exits 1 with a message naming the filter, and
`-r api` without `-p` refuses before doing anything.

**Pitfalls.** Writing "an unmatched filter is an error" and considering the
requirement specified. It is not; it is four requirements wearing one coat.

---

### Phase 10 — Feature 003, workspace lifecycle

**Goal.** `specs/003-lifecycle/` complete; `ws init`, `ws pull` and `ws purge`
working end to end; and — the part that outlives this feature — the reporting
contract and the execution seam that every later command uses.

**Why now.** This is the first feature that touches the disk, and it establishes
two things four later features depend on: how progress is reported, and how the
domain shells out. Both are much cheaper to get right once than to change after
seven commands inherit them.

**What the specification must pin down.**

*Cloning.* Project directories are created **before** any cloning starts, so
concurrent workers never race on directory creation. A repository whose
destination already exists is skipped, not re-cloned. The clone command comes
from rendering the workspace template. A clone that fails leaves **no** partial
directory behind — the destination is removed. `--clone-delay-seconds` pauses
after each successful clone, in sequential and parallel mode alike.

*Pulling.* A rebase pull against the repository's effective branch. Repositories
not present on disk are skipped rather than failing.

*Purging.* The safety contract, and it is worth stating as a principle rather
than a behaviour: purge **never** falls back to the default workspace. It
requires `--name` or `--all` explicitly, confirms unless told not to, reports a
skip when the directory is already absent, and records a failure rather than
aborting the run.

*Dry run.* A banner, a "would …" line per repository including the rendered clone
command, and a summary line prefixed to match. The requirement that makes this
testable: a dry run performs **zero** command executions and **zero** filesystem
writes. Say that explicitly, because it is what the test asserts.

*Parallel.* One worker pool across every project, not one per project, with a
running done-of-total counter.

*Reporting.* Per-repository status lines with the fixed symbols, optional
captured output indented beneath, a closing summary counting successes, skips and
failures using the operation's own verb, failure details on stderr, and exit 0
when nothing failed and 1 when anything did. Multi-target operations continue
past a failure and report them all at the end.

*The message strings themselves are part of the contract* — `already exists,
skipped`, `not present, skipped`, `cloned (main)`. Put them in the spec. They are
what the tests assert and what the parity matrix checks.

*Cancellation.* An interrupt stops between repositories and reports cancellation
rather than continuing.

*The execution seam.* Everything that shells out goes through one interface with
one real implementation and one deterministic fake. The spec states the
behaviour — tests never invoke real `git`, real network or a real home directory
— and the plan chooses the mechanism.

**What to watch in review.** Three things, in order. Does the dry-run test assert
zero executor calls, or does it just check the output looks right? Does a failed
clone actually remove the directory, and is there a test that proves it? Does the
parallel path visit every repository exactly once — a `sync.WaitGroup` misuse
here produces a test that passes and a tool that silently skips repositories
under load.

**Check.** `just test` green; a dry run over a workspace where some repositories
exist and some do not produces correct "would" lines and touches nothing; a
deliberately failing clone leaves no directory; `ws purge` with no target exits
non-zero without prompting.

**Pitfalls.** Letting the reporting contract emerge from the implementation
rather than specifying it. If you do, features 004 and 006 will each copy
whatever accident got written here, and by the time you notice, seven commands
disagree about what a summary line looks like.

---

### Phase 11 — Feature 004, cross-repository operations

**Goal.** `specs/004-crossrepo/` complete; `ws exec`, `stash`, `checkout`,
`search`, `diff`, `stats` and `list` working, plus the shared table renderer.

**Why now.** Every one of these is an addition on top of the fan-out and
reporting from 003. None of them needs anything the previous features did not
already provide, which makes this the phase where you feel the investment pay
off.

**What the specification must pin down.** This is a wide feature, so the spec
grows accordingly. Per command:

*exec* runs an arbitrary command in each resolved repository, captures its
output, and prints it indented beneath that repository's status line. Not-present
repositories are skipped. Failures do not stop the run.

*stash* and *checkout* are thin: a stash push, and a checkout of a branch taken
as a required positional argument.

*search* has three modes and the spec must state their precedence, not leave it
to the implementation: substring content search by default, filename glob with
`--glob`, regular expression with `--regex`, and **glob wins when both are
given**. Version-control metadata directories are skipped entirely. Binary files
are skipped, detected by a null byte in the first 512 bytes — state the rule, not
"skips binary files". Output is `[project/repo] file:line: content` for content
matches and `[project/repo] file` for filename matches, closing with
`N match(es) across M repo(s)`. An invalid regular expression fails **before any
repository is walked**.

*diff* renders a per-project table with columns REPO, UNCOMMITTED, AHEAD, BEHIND.
Two distinct placeholders, and this is the detail most likely to be got wrong: a
repository that is not cloned renders `(not cloned)` in all three value columns,
while a cloned repository with no upstream renders `--` in AHEAD and BEHIND only.
Rows are coloured by state, and the table closes with counts of clean, dirty and
not-cloned.

*stats* renders REPO, LAST COMMIT, 30D COMMITS, CONTRIBUTORS. Last commit is
normalised to `YYYY-MM-DD HH:MM`. An empty history renders `--`; a repository not
on disk renders `(not cloned)` with `-` in the numeric columns. Contributors are
unique author addresses compared case-insensitively after trimming.

*list* is a tree view grouped by project, each repository with its resolved local
path and whether it exists, closing with a cloned/not-cloned summary.

*The shared table renderer:* column widths sized to the widest cell including the
header, a bold header row, a rule beneath it, two spaces between columns, a
consistent indent, and correct output when colour is disabled.

*Ordering.* Output preserves the order declared in the configuration even when
the work was collected concurrently. This is a requirement, not an implementation
detail — say it in the spec, because it constrains the design.

**What to watch in review.** The two diff placeholders, which the old project's
own documentation gets wrong and its code gets right — decide deliberately and
record the decision. Order preservation under `--parallel`, which is easy to
implement by accident and easy to break by accident. And whether the binary
detection is the specified rule or something the agent invented.

**Check.** Golden-output tests for both tables. A parallel run and a sequential
run over the same workspace produce byte-identical output. A search over a
directory containing a binary file finds nothing in it and does not error.

**Pitfalls.** Specifying "show a table of repository state" and moving on. Every
implementation of that sentence differs, and none of them is wrong, which means
you have no way to tell whether the one you got is the one you wanted.

---

### Phase 12 — Feature 005, the declarative package registry

**Goal.** `specs/005-pkgreg/` complete; `zgard pkg install`, `purge`, `register`
and `unregister`; the dependency resolver; the generated shell environment file.

**Why now.** It depends only on configuration and paths, so it can slot in
anywhere after 001 — but it must come before 007, because the installer seeds the
global registry file this feature reads.

**What the specification must pin down.**

*Two registry layers.* A distribution-managed global registry, replaced on every
upgrade, and a user-managed local one that survives upgrades. A missing local
registry behaves as an empty one, not as an error.

*Merge semantics,* which are exact and easy to describe wrongly. Identity is the
pair (name, version). A local entry with the same identity as a global one
replaces it. A local entry with the *same name but a different version* is an
additional, separate entry — not a replacement. State both halves.

*Dependency references.* A reference is a bare name or `name@version`. A bare
name resolves to the sole package with that name; when several versions exist it
resolves to the unversioned one; and when that is ambiguous it is an error whose
message tells the engineer to qualify the reference. A versioned reference that
no entry satisfies is an error naming the version.

*Resolution.* Transitive dependencies are expanded automatically. Installation
order is a **deterministic** topological order — deterministic matters, because a
resolver that returns a different valid order each run makes every failure
irreproducible. A cycle is detected and reported naming the packages involved.
Purge order is the exact reverse.

*The install lifecycle,* in this order, per package: create the package directory
if requested; write the `<name>:pre` environment block and re-source the
environment file so the install script sees it; run the install script with the
environment file sourced first; write the `<name>:post` block and re-source again
so later packages see it. Each phase script runs with a documented set of
variables available to it: the project root, the package directory, the package
name, the version.

*The purge lifecycle:* the optional purge script, then removal of the package
directory if it was pre-created, then excision of both environment blocks.

*The environment file* is a set of named, lexically delimited, **idempotent**
blocks. Writing the same block twice replaces it in place rather than appending.
Removing a block leaves no doubled or leading blank lines. Removing a block that
is not there, and operating on a file that does not exist, are both no-ops.

*Presentation.* Script output is suppressed behind a progress indicator by
default; `--verbose` streams it instead.

*Registration* is interactive: every field prompted, dependencies chosen by
number from the merged package list, multi-line fields terminated by a blank
line. *Unregistration* removes by name, by exact name and version, or all local
entries — and does **not** uninstall anything. Say that explicitly; it is the
thing a user will assume wrongly.

**What to watch in review.** The merge semantics and the phase ordering, because
both read as obvious and are not. Also: every test's phase scripts must be
trivial local shell — `echo`, `true`, `exit 1`. A test that installs a real SDK
is not a test.

**Check.** A cycle is detected and named. A bare reference with two versions
available produces the qualify-it error. The same environment block written twice
produces a byte-identical file. A purge of a package whose directory was never
created succeeds.

**Pitfalls.** Letting the resolver be non-deterministic. It will pass its tests
and produce install failures you cannot reproduce.

---

### Phase 13 — Feature 006, the health daemon

**Goal.** `specs/006-daemon/` complete; `dukh start`, `stop`, `status` and
`scan`; the monitoring loop; the gRPC service; and `zgard ws status` reading from
it.

**Why now.** It needs configuration from 001 and destination resolution from 002,
because it must locate every repository exactly as `ws init` would — if the two
disagree, the daemon reports repositories as missing that are sitting on disk.

**What the specification must pin down.**

*Starting.* The start command returns control to the engineer without leaving a
foreground process, and prints the new process id. Specify that as behaviour;
leave the mechanism to the plan.

*Process identity.* A file under the project's `run/` directory recording the
running process, written at startup and removed on clean shutdown. A stale file
whose process is gone is cleaned up by the status command rather than reported as
a running daemon.

*The three contract operations.* Report the health snapshot, optionally scoped to
one workspace and optionally forcing a fresh scan first and waiting for it;
trigger a rescan without waiting; request graceful shutdown. Specify the payload:
per workspace, per project, per repository — name, resolved path, configured
branch, actual branch, whether it exists, whether the branch is aligned — plus
the server's version and its uptime in seconds.

*The monitoring loop.* A scan at startup, then a scan every configured period.
The period is re-read from configuration **before each wait**, so changing it
takes effect at the next tick without a restart. A documented default applies
when it is absent or non-positive.

*Rescan requests.* Fire-and-forget requests are dropped when one is already
queued. A synchronous request waits for that scan to finish and gives up if the
caller's own deadline passes first.

*Shutdown.* On request or on a termination signal: stop accepting work, finish
what is in flight, remove the identity file, exit.

*The report.* The workspace/project/repository tree with alignment symbols, drift
shown as `configured → actual`, a closing count of aligned, drifted and missing,
and a header line carrying the daemon's uptime.

*Auto-start.* When the engineer asks for health and the daemon is not running, it
is started for them, waited for until it answers, then queried — unless
auto-start is disabled in configuration, in which case the failure is explicit
rather than silent.

*Logging.* A rotating file under `logs/`, mirrored to standard error, bounded in
size and in number of compressed backups.

**What to watch in review.** **Testability, above everything.** A daemon that
determines a repository's branch by running `git` directly is a daemon you cannot
test. The specification should require the behaviour; the plan should introduce
an injectable seam in front of it. If the plan does not, send it back — this is
the one feature where a missing seam makes the whole thing untestable rather than
merely awkward.

Second: no test may bind a fixed port. A test suite that fails when you happen to
have the daemon running is a test suite you will start ignoring.

**Check.** Tests run with no `git` on the path and no fixed port. A stale identity
file is cleaned up. A cancelled synchronous rescan returns promptly with a
cancellation error. Changing the monitoring period in configuration takes effect
without restarting.

**Pitfalls.** Specifying "reads the current branch of each repository" and
discovering at implementation time that this makes the feature untestable. The
specification is where you notice that, if you are reading it properly.

---

### Phase 14 — Feature 007, the toolchain manager

**Goal.** `specs/007-toolchain/` complete; `grazhda-install.sh`,
`grazhda-init.sh` and the `grazhda` management script; the Bash test harness.

**Why now.** Last, because it installs everything the other features produce. It
cannot be meaningfully tested until there are binaries to install.

**What the specification must pin down.**

*Install.* Create the project root; clone or update the sources into it; verify
every required build tool and, when any is missing, report **all** of them rather
than the first; build; place the binaries; seed the configuration file from the
template **only when one does not already exist**; seed the global package
registry; and wire the shell integration into the engineer's profile
idempotently. The installer refuses to run when sourced rather than executed, and
works when piped from a download.

*Shell integration.* Export the project root when unset; prepend the binary
directory to the search path **exactly once**; create the packages directory;
source the generated environment file when it is readable.

*Upgrade.* Refuse when the sources directory is not a repository. Verify build
tools. Update the sources and report whether anything changed. Stop a running
daemon gracefully before replacing its binary, escalating to a forced stop after
a bounded wait, and clean up a stale identity file. Rebuild, install, refresh the
global registry from the sources, and tell the engineer plainly that the daemon
is **not** restarted for them.

*Uninstall.* Confirm; stop the daemon; remove the shell-profile lines; delete
everything under the root **except** the configuration file, which is
deliberately preserved so a reinstall reuses it. *Purge* is the same but deletes
the root entirely, with an explicit warning that it is irreversible.

*Configuration editing.* An editor resolved in a documented order — the
configured value, then the two conventional environment variables, then a final
fallback — with the path quoted so spaces are safe, and a clear failure when no
configuration file exists.

*Output and flags.* A verbosity flag accepted in any argument position. Quiet by
default, but a failed step prints the output it captured, so a failure is never
silent. Consistent information, success, warning and error lines. A usage
command. An unknown command produces an error, the usage text, and a non-zero
exit.

**What to watch in review.** Idempotency, everywhere. Install twice and the shell
profile must be byte-identical to after the first install. Uninstall and it must
be byte-identical to before. This is the property that makes the installer safe
to run when you are not sure whether you already ran it — which is the only state
anyone is ever in.

**Check.** `shellcheck` clean. The harness runs install, install again, then
uninstall, all against a temporary home directory, and asserts the profile is
unchanged at the end. A missing build tool produces a list, not a single
complaint.

**Pitfalls.** A test harness that touches your real home directory or your real
shell profile. Set `HOME` to a temporary directory in the harness and never
assume; the first time this goes wrong you will be restoring your own `.bashrc`
from memory.

---

**At the end of Phase 14** you have parity: every capability of the old system,
specified, implemented, tested and traceable — and `docs/parity-matrix.md`
carries no empty cells. Everything from here is new.

---

## Phase 15 — Feature 008, `molfar`, the long-term process

**Goal.** The automation service: it accepts a chore, plans it, runs it across
your repositories, records everything it did, and hands back a proposal you
approve — or do not.

**Why this is last, and separate.** It is the only genuinely new component, it
depends on everything before it, and its failure modes are qualitatively
different from the rest of the project. A bug in `ws diff` prints a wrong number.
A bug here pushes a commit you did not authorise, or runs for six hours, or
writes a credential into a log file. It gets its own milestone and its own
sub-phases.

**Read `guides/04-spring-boot.md` first.** It teaches Spring from zero and targets
4.1 specifically. One thing from it before you write a line: in Spring Boot 4
auto-configuration is modular, so putting a library such as `flyway-core` on the
classpath gives you the library with no Spring wiring and **no error at all** —
total silence, no migrations, no log line. Depend on the starter, always. The
starter is the contract.

**Read `IDEAS.md` too.** It argues for several things that belong in this
specification rather than being bolted on later: a determinism ladder, a
verification oracle on every step, treating repository content as untrusted
input, and shipping in observe-only mode first.

### 15.1 Specify it, properly

This is the one specification where thoroughness is not a matter of taste. The
requirements that carry the most weight:

- Every run carries a **recorded budget** — wall clock, step count, files changed,
  and money — and stops when any of them is exhausted, in a terminal state that
  names which one.
- Nothing is pushed, merged, tagged or published without a **recorded approval**.
- Every executed step is **reconstructable from the record** after a restart:
  the command, the working directory, the exit status, the captured output, the
  timing, the files changed.
- **Credentials are redacted** on the way into every record, proposal and
  notification.
- A step that would act **outside the task's declared scope** is refused and
  fails the run.
- Repository content may influence the **content** of a proposal but never the
  **policy** — scope, budget, approval requirement. This is the prompt-injection
  defence, and it has to be structural, not advisory.

Run `/speckit-clarify` twice on this one. It is worth it.

### 15.2 The contract, before any controller

Write `contracts/molfar-openapi.yaml` first. Contract-first is a constitutional
principle and the CLI's client is generated from the same file. Deciding the
resource shapes — chore, plan, run, step, proposal, approval — before writing a
controller forces you to settle what the service actually exposes rather than
letting it accrete.

### 15.3 The framework-free core

Put the task state machine, budget accounting, scope checking, plan validation
and credential redaction in a plain-Java `core` package with **no framework
dependencies**, unit-tested with plain JUnit and no Spring context.

This is the same layering rule the Go side follows, and it matters more here, not
less: these are the rules that keep the system safe, and you want to be able to
test them exhaustively in milliseconds rather than through a Spring context that
takes eight seconds to start. Guide `04-spring-boot.md` part 10 shows the package
layout and how to enforce the rule with an architecture test, so it fails the
build rather than eroding.

### 15.4 Persistence and the journal

The journal is not a feature of this service; it is the product. Design it as
queryable dimensions rather than free text, because the questions you will
actually ask are "what did it try on this repository", "which proposals did I
reject and why", and "how long does this chore usually take".

Flyway migrations from the start, even for the first table. The embedded database
choice and its trade-offs are covered in the Spring guide.

### 15.5 The execution engine

Action kinds behind a registry, each declaring its rung on the determinism
ladder and carrying its own verification oracle. Two requirements from `IDEAS.md`
that belong in the code from day one rather than being retrofitted: a step with
no oracle is marked unverified and surfaced separately in the proposal, and a
repository with uncommitted changes is refused unless the task explicitly opted
in.

### 15.6 The API and the CLI bridge

The controllers, then `zgard agent submit`, `list`, `follow` and `approve`.
`follow` exits with a status reflecting the run's own terminal state, so it
composes in a shell script.

### 15.7 Observe-only, then hands

Ship it in shadow mode: it receives events, plans, and writes what it *would*
have done into the journal, executing nothing. Run it against your real work for
a few weeks, then read the journal and count how often the plan was the right
one.

That number is the only honest answer to "is this ready to be given hands", and
there is no way to get it except by waiting. Resist skipping this. It also
produces the first entries in the evaluation corpus, which is what lets you tell
whether tomorrow's change to the planner helped.

**Check.** A chore submitted from the command line produces a proposal you can
read; the journal survives a restart with the run intact; nothing has reached a
remote without your approval; and a run given an impossible budget stops and says
which budget it ran out of.

---

## Phase 16 — Close out

**Goal.** Nothing left that only exists in your head.

**Do this.**

- Walk `docs/v1-baseline/CLI.md` command by command against the built binaries.
  Every row in `docs/parity-matrix.md` goes green or gets an explicit note
  saying why it will not.
- Re-read every `spec.md` against its implementation one final time. Statuses
  move to `Implemented`.
- Generate man pages and shell completions.
- Finish the release workflow so a tag produces binaries.
- Write the README as if for someone who has never seen the project.

**Check.** A stranger can clone, run `just build`, and use the tool from the
README alone.

---

## Appendix A — Working with the agent

**Give it the spec, not a paraphrase.** "Implement `specs/003-lifecycle/spec.md`,
requirements FR-001 through FR-018" beats describing what you want in chat. The
spec is the contract; the chat is not.

**Make it show its reasoning at the gates.** Before implementation, ask it to
list which requirement each planned file satisfies. Gaps show up immediately.

**Review against the spec, not against your instincts.** If you find yourself
objecting to something the spec permits, the spec is what needs changing — change
it deliberately and re-run, rather than negotiating in chat.

**Watch for spec drift.** Any change to `specs/**` during an implementation step
is suspect. Read the diff.

**Keep the loop tight.** One feature per session. An agent asked to implement
four features at once will produce four half-features and a coherent-sounding
summary.

**When it gets stuck, suspect the specification.** An agent thrashing on a
requirement is usually a sign the requirement is ambiguous or contradicts
another one. Go and read it rather than re-prompting.

---

## Appendix B — Troubleshooting

| Symptom | Cause | Fix |
| :--- | :--- | :--- |
| `cd` on one Justfile line has no effect on the next | each recipe line is its own shell | chain with `&&`, or use a shebang recipe |
| `protoc-gen-go: program not found` | `$HOME/go/bin` not on `PATH` | add it, then re-run `just generate` |
| `go: cannot find module providing package …/libs` | module not in `go.work`, or no `replace` | `go work use ./libs`, add the `replace` |
| CI fails on `git diff --exit-code -- daemon/proto` | generated code is stale | run `just generate` and commit |
| A test passes locally, fails in CI | it depends on your home directory or real `git` | inject the seam; use `t.TempDir()` |
| `just test` green but the feature is wrong | the spec was wrong, or the tests test the code rather than the requirements | trace each `FR-###` to a test |
| `go.work.sum` differs between your machine and a teammate's, or CI complains about it | it records checksums for the *workspace's* module resolution, not a portable artefact | never hand-edit it; regenerate with `go build ./...` from the repo root; keep it out of git (Appendix E) |
| `go build` fails with "missing go.sum entry" after adding a `replace` | the `replace` fixed the import path but the module's own `go.sum` was never updated | run `go mod tidy` inside the specific module directory, not at the repo root |
| A `replace` directive works for you but a fresh clone or `go install` of the module fails elsewhere | `replace` is local to the module file; it is never published or honoured by a consumer outside this repository | keep `replace` for intra-repo development only; never rely on it for a module you expect to be imported externally |
| `protoc` runs but generated Go lands in an unexpected package, or imports fail | `.proto` has no `option go_package`, or `--go_opt=paths=source_relative` was dropped | set `option go_package` explicitly in `contracts/dukh.proto` and keep the flag in the `just generate` recipe |
| `protoc-gen-go-grpc: unknown flag` or similar plugin errors after a `go install …@latest` | the plugin version drifted out of sync with `protoc` or the `grpc` module version | pin exact versions (`@vX.Y.Z`) for both plugins in the setup instructions, not `@latest` |
| A recipe you know exists does not show up in `just --list` | it is marked `[private]`, or has no comment line directly above it | remove `[private]`, or add a `#` comment immediately above the recipe |
| A `just` recipe keeps going after a command inside it clearly failed | a leading `-` on that line, or a shebang recipe missing `set -euo pipefail` | drop the `-`; add `set -euo pipefail` as the first line of shebang recipes |
| `molfar` builds and starts, but a feature that depends on a library silently does nothing — no error, no log line | Spring Boot 4's auto-configuration is modular; a raw library such as `flyway-core` on the classpath gets no Spring wiring at all | depend on the starter artefact (`spring-boot-starter-…`), never the bare library, for anything that needs auto-configuration |
| `mvn -v` reports an older Java version than 25, or the build uses the wrong JDK | the SDKMAN environment was not sourced in the shell that invoked `mvn` (a fresh terminal, or CI) | source `~/.sdkman/bin/sdkman-init.sh`, or set `JAVA_HOME` explicitly, and verify with `mvn -v` before building |
| `molfar`'s test suite fails intermittently in CI with a bind or bean-conflict error, but passes alone | Spring test contexts sharing a fixed port or an application context across modules | use `server.port=0` in test configuration; reserve `@DirtiesContext` for tests that actually need a fresh context |
| `/speckit-plan` or `/speckit-tasks` refuses to run, reporting a missing prerequisite | `check-prerequisites.sh` cannot find the artefact the phase depends on, usually because the active feature pointer is stale | confirm `.specify/feature.json` names the feature you mean; re-run `create-new-feature.sh` or `/speckit-specify` if it does not |
| A new feature is numbered `009` when you expected `008` | `create-new-feature.sh` derives the next number by scanning `specs/` for the highest existing prefix, and a renamed or deleted directory throws that off | pass `--number` explicitly, or check the directory names under `specs/` before running the command |
| An edit to a file under `.specify/templates/` has no effect, or reverts after an upgrade | templates are regenerated against recorded hashes on a refresh; direct edits are overwritten | put project-specific changes in `.specify/templates/overrides/` instead of editing the template in place |
| `just build` and `just test` pass on your machine but fail on a fresh clone or in CI | a generated or locally-cached file was stale and only worked by accident, or a toolchain version differs from CI's | rebuild from a fresh clone before pushing; pin exact tool versions in the CI workflow rather than trusting whatever is installed |
| The `daemon/proto` diff check fails only sometimes, not on every push | the `protoc` or plugin versions used locally differ from the versions CI installs | pin exact `protoc` and plugin versions in one place both environments read from, instead of installing "latest" in either |

---

---

## Appendix C — Glossary

**Adapter** — code that translates between the outside world and the domain:
command handlers, RPC servers, HTTP controllers. Contains no business rules.

**Constitution** — the project's governing principles, checked by every plan.

**Domain** — the code that knows the business rules and nothing about terminals,
transports or frameworks.

**FR-###** — a numbered functional requirement in a specification; the unit of
traceability between spec, test and commit.

**Oracle** — the check that tells you an automated step actually worked: a build
that passes, a test that goes green, a pattern that disappears.

**Parity matrix** — the table tracking, for every capability of the old system,
whether it is specified, implemented in the domain, exposed by a command, and
tested. Your definition of done.

**Reconciliation** — updating a specification after implementation so that it
describes what actually shipped.

**SC-###** — a numbered, measurable success criterion.

---

## Appendix D — Worked templates

Several phases above tell you to write a document without showing you one. Each
example here is a starting *shape*, not a form to fill in: the value of these
documents is in the thinking, and a template followed thoughtlessly produces one
that is complete, correctly structured and read by nobody — including you, three
weeks later, when it would have saved you an afternoon. Work out why each says
what it says, then write your own; if yours come out generic, you filled in a
form rather than answering a question.

### D.1 — `docs/product.md` (Phase 0)

What makes this good is that it is *narrow enough to be wrong*: a definition you
cannot disagree with has not defined anything. The five questions are ordered so
each constrains the next — name whose work it is and the chores narrow, name the
chores and the output follows.

```markdown
# Grazhda — what it is for

## 1. Whose work

One engineer — me — maintaining roughly thirty repositories across four
workspaces, grouped into projects that mostly share a branch and a build. Not a
team, not a platform group. It runs on my machine, reads my `config.yaml`, and
acts with my Git credentials.

## 2. Which chores

Ten things I did last month and resented, in the order I would hand them over:

1. Bump the logging library across every repository tagged `backend`, confirming
   each still builds.
2. Correct CI workflow drift: eleven repositories disagree about runner image.
3. Migrate calls to a deprecated internal HTTP client to its replacement.
4. Adopt a new lint rule and fix every violation it produces.
5. Add or correct licence headers after a copyright-line change.
6. Rename a configuration key in every service that reads it, plus its docs.
7. Triage overnight CI failures into flake, regression and infrastructure, with
   the evidence for each classification.
8. Assemble release notes for a coordinated release across six repositories.
9. Answer "what changed in these six repositories since I last looked".
10. Chase stale pull requests and summarise what each is blocked on.

Chores 7 to 10 write no code at all: lower risk, immediately useful, and they
build the integrations the code-changing chores will need anyway.

## 3. What the system produces

A **proposal**: a branch pushed to a remote, or a diff held locally, with the
record of how it was produced — what was asked and planned, what each step did,
which check confirmed it, where it stopped. That is the unit I review. Not
merged code, not a deployment, not a chat transcript.

## 4. What it must never do

1. Never push, merge or deploy without a recorded approval from me.
2. Never touch a repository outside the run's declared scope — a workspace,
   project, repository or tag set, resolved before the first step.
3. Never run unbounded: every run carries a budget in time, steps and money.
4. Never let a secret reach a log line, a prompt or the run record.

## 5. How I will know it worked

- **SC-A**: In a quarter, at least twelve chores complete unattended and produce
  proposals I approve with no edits, counted from the run record.
- **SC-B**: Median time from an overnight CI failure to a triaged classification
  with evidence, under thirty minutes against the CI timestamp.
- **SC-C**: Unapproved writes to any remote: zero, for the life of the project.
```

**Question two is where the project lives or dies**, so put the failure beside
the success:

```markdown
Bad:  Dependency management. Code quality. Repository maintenance.
      Keeping things up to date. Reducing toil.

Good: Bump the logging library across every repository tagged `backend`
      and confirm each still builds.
```

The bad answer names categories, and categories cannot be planned, budgeted or
checked — nothing tells you whether "code quality" succeeded. The good one names
a scope, a transformation and an oracle. Specify the first and the agent
produces something plausible and useless.

### D.2 — `docs/capability-inventory.md` (Phase 2)

The inventory lists what the old system does, and its indispensable column is
**intended or accidental**. Without it you will transcribe accidents into
specifications and defend them for a year. Fill it by asking of each row:
designing this today, would I do it this way? The accidental rows below come
from the baseline's own `GAPS.md`; yours should add what you find yourself.

```markdown
# Capability Inventory — grazhda v1

One row per observable behaviour of the baseline. The last two columns are the
point: *intended* behaviour is a candidate requirement, *accidental* behaviour
is a decision waiting to be made.

| Capability | Flags | Behaviour | Intended or accidental? | Decision |
| :--- | :--- | :--- | :--- | :--- |
| `ws init` | `-n`, `--all`, `-p`, `-r`, `-t`, `--dry-run`, `--parallel` | Clones every repository in the resolved target; skips ones already on disk; continues past a failure and reports all failures at the end. | Intended — the continue-past-failure contract is stated in `CLI.md` and relied on. | Specify as-is (003). |
| Default-workspace fallback | none | With no targeting flag, resolves the workspace marked `default` and prints a blue notice to stderr naming the path. | Intended — the notice exists so the target is never a surprise. | Specify as-is (002). |
| `-r/--repo-name` matching | `-r` | Case-sensitive substring match against the full config name, so `-r cool` matches `ORG/PACK/my-cool-backend-lol`; multiple matches warn, zero matches error. | Intended, but under-documented — the substring semantics surprise people. | Specify as-is; state the rule explicitly (002). |
| `ws status` targeting | `-p`, `-r`, `-t` | Flags inherited from the `ws` parent are accepted and then silently ignored; only `-n` filters the daemon snapshot. | **Accidental** — inherited flags nobody wired up. | Close: 006 applies the same filters as every other `ws` command. |
| `ws search` mode selection | `--glob`, `--regex` | Both may be supplied together; glob wins because the regex path is skipped when `--glob` is set. Nothing rejects the ambiguity. | **Accidental** — an ordering artefact, not a decision. | Carry: 004 specifies the precedence explicitly rather than erroring, so existing scripts keep working. |
| `workspace.structure` values | n/a | Any value other than `list` is treated as `tree`; validation does not reject unknown values, so `treee` silently works. | **Accidental** — a fall-through default. | Close: 001 rejects unknown values at validation. |
| `zgard.*` and `general.*` config | n/a | Parsed, validated and queryable through `config get`; no workspace or package operation reads them. | **Accidental** in origin, defensible in effect. | Carry: 001 documents them as inspection metadata in its Assumptions. |
```

The `Decision` column matters as much as the honesty column. "Close" means the
rebuild fixes it; "Carry" means it reproduces the behaviour deliberately, with a
spec that says so. What you must not have is a third state, where a behaviour
survives because nobody looked.

### D.3 — `docs/parity-matrix.md` (Phase 2)

This is your definition of done. It separates specified, implemented and tested
because one "done" column hides the failure you are guarding against: a
capability can be implemented and untested, tested and unspecified, or specified
and quietly never built — three problems with three fixes. A fourth column earns
its place on the Go commands, where domain and adapter ship separately: targeting
can be tested in `libs/` while the `zgard` command exposing it does not exist.

```markdown
# Parity Matrix

Parity with the pre-specification implementation (`docs/v1-baseline/`) is a
standing obligation until every row reads ✅ (constitution § Governance).
Legend — **spec**: a specification covers the behaviour · **domain impl**: the
logic exists in `libs/` · **CLI impl**: a command exposes it · **test**:
automated tests cover it. **✅** done · **🟡** partial · **⬜** not started.
Last updated: 2026-08-30.

## `zgard ws` — workspace operations

| Baseline capability | Feature | Spec | Domain impl | CLI impl | Test |
| :--- | :--- | :---: | :---: | :---: | :---: |
| `ws init` (clone, skip existing, continue on failure) | 003 | ✅ | ✅ | ⬜ | ✅ |
| `ws pull` (rebase, skip uncloned) | 003 | ✅ | ✅ | ⬜ | ✅ |
| `ws purge` (explicit target, confirm, dry-run) | 003 | ✅ | ✅ | ⬜ | 🟡 |
| `ws exec <cmd>` (capture + indent output) | 004 | ✅ | ✅ | ⬜ | ✅ |
| `ws diff` (uncommitted / ahead / behind table) | 004 | ✅ | ✅ | ⬜ | ✅ |
| `ws stats` (last commit / 30d / contributors) | 004 | ✅ | ✅ | ⬜ | ✅ |
| `ws list` (tree view, clone status) | 004 | ✅ | ✅ | ⬜ | 🟡 |
| `ws status` (health from the daemon) | 006 | ✅ | ✅ | ⬜ | ✅ |

### Shared targeting and flags

| Baseline capability | Feature | Spec | Domain impl | CLI impl | Test |
| :--- | :--- | :---: | :---: | :---: | :---: |
| `-n/--name`, `--all`, default fallback + notice | 002 | ✅ | ✅ | ⬜ | 🟡 |
| `-t/--tag`, OR logic, project→repo inheritance | 002 | ✅ | ✅ | ⬜ | ✅ |
| Cross-flag constraints (`-r` requires `-p`) | 002 | ✅ | ✅ | ⬜ | 🟡 |
| Summary line, stderr failures, exit code 0/1 | 003 | ✅ | ✅ | ⬜ | ✅ |
```

Close the file with a table carrying the keep-or-close decision for every row of
`GAPS.md`, so the two cannot drift apart. And watch the 🟡 rows: a matrix with
no partials usually means somebody is rounding up.

### D.4 — `.specify/memory/constitution.md` (Phase 4)

The skeleton below is the whole document, one principle at full strength and the
rest reduced to a line each: the specificity of Principle I is what to copy, not
the section list. It names an artefact path, defines what counts as a violation,
fixes the identifier every requirement carries and says when the loop closes.

```markdown
# Grazhda Constitution

Grazhda automates the work of a software engineer, as three cooperating
processes — a CLI interface (`zgard`), a CLI daemon (`dukh`) and a long-term
orchestration process (`molfar`, Spring Boot) — plus a shell toolchain manager
(`grazhda`) that installs them. This constitution governs every specification,
plan, task list and line of code here. It supersedes habit and prior art.

## Core Principles

### I. Specification Is the Source of Truth (NON-NEGOTIABLE)

No behaviour exists until it is specified. Every feature begins as a
`specs/###-name/spec.md` describing *what* and *why* in technology-neutral
language, then `plan.md` (*how*), `tasks.md` (*in what order*), then code.

- Code that contradicts its spec is a defect in the code **or** the spec; the
  loop is not closed until one of them is corrected.
- Every functional requirement carries a stable `FR-###` and every measurable
  outcome an `SC-###`. Tests, tasks and commit messages reference them.
- Specs are living documents: after implementation the spec is reconciled with
  what was built, before the feature counts as done.

### II–VII — one line each here; each expands to the depth of I

- **II. Three Processes, One Domain** — domain logic in shared libraries;
  handlers, gRPC methods and controllers parse, delegate and render.
- **III. Declarative and Reproducible** — YAML validated before use; `--dry-run`
  and an explicit target on anything destructive.
- **IV. Test-First, Contract-First** — contracts before implementations; a
  failing test per requirement; shell-outs only through an injected executor.
- **V. The Terminal Is a User Interface** — fixed symbols (`✓` `✗` `⚠` `⏭`),
  `NO_COLOR` honoured, errors to stderr and data to stdout.
- **VI. Automation Must Be Accountable** — every run records what it was asked,
  did and why it stopped; propose, never publish; bounded scope and budget.
- **VII. Simplicity, Then Capability** — the smallest thing that satisfies the
  requirement; every dependency justified in Complexity Tracking.

## Technology Constraints
Go 1.26+ over a `go.work` workspace for `zgard` and `dukh`; Java 25 and Spring
Boot 4.1 for `molfar`; shellcheck-clean Bash for the toolchain manager; `just`
as the single entry point; configuration at `$GRAZHDA_DIR/config.yaml` with
flags > environment > file; no unit test may require the network.

## Development Workflow
Specify, clarify, plan, tasks, implement, reconcile — and the rule that the loop
is not complete until the build is green *and* the specification describes the
shipped behaviour.

### Quality Gates

| Gate | Requirement |
| :--- | :--- |
| Plan approved | Constitution Check passes with no unjustified violation |
| Implementation complete | `just build` and `just test` green; every `FR-###` covered by a test |
| Feature done | Spec reconciled; parity matrix updated; no `NEEDS CLARIFICATION` left |
| Release | Contracts versioned; migration notes written for any breaking change |

## Governance
Amendments require a stated reason, the principle text before and after, and a
version bump. Every plan includes a Constitution Check; an unjustified violation
blocks the feature. Parity with `docs/v1-baseline/` is a standing obligation
until `docs/parity-matrix.md` reports full coverage.

### Amendment log

| Version | Date | Change | Reason |
| :--- | :--- | :--- | :--- |
| 1.0.0 | 2026-08-30 | Ratified. | — |

**Version**: 1.0.0 | **Ratified**: 2026-08-30 | **Last Amended**: 2026-08-30
```

**Versioning a constitution.** Semantic versioning applies to the promises the
document makes, not its word count. MAJOR when a principle is removed or
redefined, because plans approved under the old text may no longer be valid;
MINOR when a principle or section is added, because nothing already approved
becomes wrong; PATCH for clarification that changes no obligation. Each log entry
states a reason and enough before-and-after that a reader of an old plan can tell
which version it was checked against:

```markdown
| 1.1.0 | 2026-09-12 | Technology Constraints: the long-term process moves from Java 21 / Spring Boot 3.x to **Java 25 / Spring Boot 4.1**; the CLI and daemon floor is corrected from Go 1.24+ to **Go 1.26+** to match `go.work` and every `go.mod`. | Java 25 is the current long-term-support release and Spring Boot 4.1 the current line; adopting them now avoids a migration before `molfar` has shipped a feature. The Go correction removes a constraint the repository already contradicted. |
```

That is MINOR, not MAJOR: no principle changed meaning, the constraints section
gained tighter values. Had Principle VI been rewritten to permit unattended
pushes it would be 2.0.0, and every plan approved under 1.x would need
re-checking. It is shown here as an example of the form — Phase 1 has you
install Java 25 and Go 1.26 before you ratify, so your own 1.0.0 should already
carry those values and your first amendment will be about something else.

### D.5 — A good `spec.md`, in fragments (Phase 8.1)

Reproducing a whole specification teaches you section headings, which spec-kit
already gives you. What people get wrong is smaller — four places, drawn from
`specs/001-config-core/spec.md`.

**A user story needs a priority, a reason for it, and a test that depends on no
other story.** The independent test is what gets skipped, and what makes the
story schedulable.

```markdown
Before:
### User Story 2 - Validation
The config should be validated properly and give good errors.

After:
### User Story 2 - Catch mistakes before they cost me an hour (Priority: P1)

Before any operation touches the filesystem, the engineer can validate the
configuration and see every problem at once — not the first one only.

**Why this priority**: A malformed config that fails halfway through a clone of
forty repositories is worse than one that refuses to start. Validation is what
makes every later operation safe.

**Independent Test**: Feed the validator files with a missing name, a duplicate
workspace name, a broken template and an unknown structure value; confirm each
produces a distinct, field-addressed message and a valid file produces none.
```

**A requirement asserts one thing and keeps its number forever.** If it contains
"and", it is two requirements and a test will cover half of it. The number ties
spec to test to commit, so it is never reused: a withdrawn FR-014 stays
withdrawn and the next requirement written is FR-035.

```markdown
Before:
- **FR-011**: Validation must check names, paths, templates, branches and
  structure values, report all errors, and address fields by path.

After:
- **FR-011**: Validation MUST return every problem found in a single pass
  rather than stopping at the first.
- **FR-016**: Validation MUST reject any `structure` value other than `tree` or
  `list` at both workspace and project level, naming the permitted values.
- **FR-017**: Every validation message MUST address the offending field by its
  path, including indices (for example `workspace[1].projects[0].branch`).
```

**A success criterion is a number somebody could measure without asking what you
meant.** The test: two people given the shipped system reach the same verdict.

```markdown
Before:  **SC-001**: Configuration loading is fast and reliable.

After:   **SC-001**: A configuration file describing 3 workspaces, 10 projects
         and 100 repositories loads and validates in under 100 ms.
```

**Edge cases show whether anyone thought about the domain.** Generic filler —
"invalid input", "network error" — means the agent produced a list rather than
reasoning about your system. Real ones name a grazhda condition and its outcome:

```markdown
- The config file exists but is empty → it loads as a configuration with no
  workspaces; validation reports nothing, and commands that need a default
  workspace fail with "no default workspace found".
- No workspace is marked `default: true` and none is named `default` → resolving
  the default target fails with an actionable message naming `--name`.
- `~` appears at the start of a workspace path → it expands to the engineer's
  home directory; a `~` elsewhere in the path is left alone.
```

Each cost a minute of thought about how the configuration is used, and each
prevents a bug the agent would otherwise resolve by guessing.

### D.6 — The Constitution Check in a `plan.md` (Phase 8.3)

"Justify it in writing" means nothing until you have seen a justification that
would be accepted. A passing check is short — each principle, and the specific
thing in *this* plan that satisfies it. "Complies with all principles" is a wave,
not a check.

```markdown
## Constitution Check

| Principle | Status | How this plan satisfies it |
| :--- | :--- | :--- |
| I. Spec is source of truth | Pass | Every task in `tasks.md` names the `FR-###` it implements; no behaviour is planned that FR-001..FR-034 do not describe. |
| II. Three processes, one domain | Pass | Load, validate, query and update live in `libs/config/`; `cli/cfg/` only parses flags and renders. No YAML type crosses into `cli/`. |
| IV. Test-first, contract-first | Pass | The config schema is fixed in `contracts/` before the loader; every FR has a failing test first (SC-002). |
| VI. Accountable automation | n/a | This feature performs no automated action. |
| VII. Simplicity, then capability | Pass | One new dependency (a YAML library); no new process or abstraction layer. |

**Result**: pass, no entries in Complexity Tracking.
```

When something does violate a principle, record it rather than rewording the
plan until the violation is invisible. A justification is accepted when it names
the principle, states what is bought and — the part people omit — says which
simpler alternative was rejected, and why:

```markdown
## Complexity Tracking

| Violation | Principle | Why it is needed | Simpler alternative rejected because |
| :--- | :--- | :--- | :--- |
| Two registry layers (global + user-local) with merge semantics, rather than one source of truth. | III — state lives in one declared place; VII — simplicity | Parity is a standing obligation (§ Governance) and the local layer is how every existing installation overrides tool versions per machine. | A single registry silently drops every user-local override on upgrade. Folding them into the global file makes a machine-specific value shared, which is wrong. |
| `zgard` starts `dukh` transparently when the daemon is not running. | VI — the system acts only on explicit instruction | `ws status` is unusable if the engineer must remember to start a daemon first; the baseline behaves this way and `autostart: false` exists to opt out. | Failing with "start dukh first" was tried in the baseline and reverted; requiring an explicit `dukh start` in every shell profile moves the problem rather than solving it. |
```

A justification you would be embarrassed to read back in six months is the plan
telling you to change the plan. The table makes the cost visible, not payable.

### D.7 — Commit message convention

A commit is the third place a requirement identifier appears, after the spec and
the test. Once all three agree, `git log --grep=FR-027` answers "why is this like
this?" without relying on memory. The convention is small:

```
<type>(<feature number>): <summary in the imperative>

<body: which FR-### or SC-### identifiers this touches, and what changed>
```

Types: `spec`, `plan`, `tasks`, `feat`, `fix`, `docs`, `chore`; the number is
the three-digit directory prefix under `specs/`. An implementation commit names
the requirements it implements, a reconciliation commit those that *changed*.
Four commits from feature 002:

```
spec(002): workspace targeting

Adds FR-001..FR-019 and SC-001..SC-004 covering --name/--all, project and
repository filters, tag OR-logic with project-to-repository inheritance, and the
cross-flag constraints. Resolves the four /speckit-clarify questions; no NEEDS
CLARIFICATION markers remain. No code.

plan(002): workspace targeting

Constitution Check passes. Targeting resolves in libs/workspace/target.go;
cli/ws/ only binds flags. One Complexity Tracking entry: substring matching on
--repo-name is carried from the baseline deliberately.

feat(002): workspace targeting

Implements FR-001..FR-019 of specs/002-ws-targeting/spec.md. Tests in
libs/workspace/target_test.go name each FR; SC-002 measured against a
100-repository fixture at 41 ms.

docs(002): reconcile spec with shipped targeting

FR-014 corrected: a --tag matching nothing is an error, not a warning — the
baseline errors and scripts depend on the exit code (docs/v1-baseline/CLI.md).
FR-020 added for the multi-match warning, which implementation revealed was
never specified. Status Draft -> Implemented; the four 002 rows in the parity
matrix ticked for spec, domain impl and test.
```

The fourth keeps the repository honest. A reconciliation commit saying only
"update spec" is how a specification quietly becomes fiction; one naming the
requirement and the reason is a record you can still trust a year later.

---

## Appendix E — Git workflow

The plan so far shows one `git commit`, at the end of Phase 8, and says
nothing about branches. That is not an oversight to leave standing once you
are running the loop for real: the commit history is the audit trail for work
an agent did on your behalf, and it is only useful if you build it on purpose.

### The commit sequence within a feature

Section 12.6 of `guides/03-spec-kit.md` walks a full worked session; this is
the discipline behind it, stated as a rule rather than a transcript. Each
artefact a feature produces gets its own commit, in the order it was produced:

```
spec(001): configuration core
spec(001): answer clarification questions
plan(001): configuration core
tasks(001): configuration core
feat(001): configuration core

Implements FR-001..FR-034 of specs/001-config-core/spec.md.
docs(001): reconcile spec against shipped implementation, update parity matrix
```

Six artefacts, six commits, in the order Phase 8 names them: specify, clarify,
plan, tasks, implement, reconcile. Do not squash them together afterwards,
even though a single combined commit looks tidier in `git log`. The reason to
keep them apart is not tidiness, it is that separation is what makes `git diff
-- specs/` mean something. If a specification and an implementation land in
the same commit, that command shows you both at once and you have lost the
ability to ask "did the spec change during implementation?" without reading
the whole diff by hand. With them separate, the question has a one-line
answer, and the history itself shows which phase introduced which change —
useful today when you are reviewing, and useful in six months when you are
trying to remember why a requirement reads the way it does.

The same logic applies within a phase. If `/speckit-clarify` changes
`spec.md`, that is its own commit, distinct from the one that created the spec
— the clarification session answered questions you asked, and its diff is
worth being able to see on its own.

### Branching

One branch per feature, named to match its spec directory: `001-config-core`,
`002-workspace-target`, and so on. This gives you a branch name that means
something when you are choosing what to review or what to merge, and it
matches the number the rest of the toolchain already uses.

The wrinkle worth knowing before you rely on it: spec-kit does not read the
git branch to decide which feature is active. It reads
`.specify/feature.json`, a one-line pointer written by `create-new-feature.sh`
(`guides/03-spec-kit.md` section 4 covers the file). `SPECIFY_FEATURE` and
`SPECIFY_FEATURE_DIRECTORY` can override it, but nothing in the loop consults
`git branch
--show-current`. In practice this means you can run `/speckit-plan` against feature 003 while sitting
on branch `004-cross-repo-ops`, and it will happily do that, because nothing
checks. The branch is a convention for you and for pull requests; it is not
wired to the tool. Get into the habit of checking `.specify/feature.json` when
something targets the wrong feature — see Appendix B.

### What must never be committed together

A specification change and a source change must never land in the same commit.
The reason is self-review: a solo developer's only defence against an agent
quietly rewriting a spec to match code it has already written (the failure
mode Appendix A and guide 03 section 11 both name) is being able to look at a
specification diff in isolation and ask whether it is something you actually
approved. Mixed into a larger commit with source changes, that diff is easy to
skim past.

Enforce it mechanically rather than by memory, with a pre-commit hook:

```bash
#!/usr/bin/env bash
# .git/hooks/pre-commit — refuse commits that mix specification and source
staged=$(git diff --cached --name-only)
if grep -q '^specs/' <<<"$staged" && grep -qE '^(libs|cli|daemon|service)/' <<<"$staged"; then
    echo "refusing: this commit changes both a specification and source." >&2
    echo "commit the specification change on its own so it can be reviewed." >&2
    exit 1
fi
```

This is the same hook guide `03-spec-kit.md` section 12.8 introduces for the
Claude Code loop specifically; it belongs here too because it is a property of
the repository, not of any one agent — it fires identically whether the staged
change came from you, from Claude Code, or from a different tool entirely.

### What belongs in `.gitignore`

```
/bin/
service/target/
go.work.sum
.specify/feature.json
*.bak
.idea/
.vscode/
```

- **`/bin/`** — build output. Nothing under it is source; committing it means every build produces a
  diff of binaries nobody reads.
- **`service/target/`** — Maven's build directory, the Java equivalent of `/bin/`, for the same
  reason.
- **`go.work.sum`** — it records checksums for how the *workspace* resolved modules on your machine,
  not a portable dependency lock. It regenerates from `go.work` and the per-module `go.sum` files, so
  committing it invites merge conflicts that carry no information.
- **`.specify/feature.json`** — per-checkout state, not a decision worth sharing: it says which
  feature *your* working copy currently has active, which is meaningless to anyone else's checkout.
  Spec-kit's own scaffolding already ignores it; this entry is here so the reason is written down
  once rather than rediscovered.
- **`*.bak`** — anything an editor or a hand-edit left behind. It should never have been staged, but
  a blanket rule is cheaper than remembering to check.
- **IDE directories** (`.idea/`, `.vscode/`) — personal editor configuration, not project
  configuration; committing it turns every contributor's toolbar preferences into a shared file.

Two things belong in git that are easy to mistake for build output and exclude
by reflex. **Per-module `go.sum` files** (`libs/go.sum`, `cli/go.sum`,
`daemon/go.sum`) are committed — each is the actual dependency lock for that
module, and without it a fresh clone cannot verify what it downloads.
**`daemon/proto/`**, the generated protobuf Go, is committed too, precisely so that a fresh clone
builds with `just build` before `protoc` is even installed; Phase 7 says this
and CI's `git diff --exit-code -- daemon/proto` check exists to catch the case
where it drifts from `contracts/dukh.proto` and nobody regenerated it.

### Tags and releases

Tag the repository at milestones that matter to a reader of the history, not
at every commit: the end of Phase 8 once the loop itself is proven, the end of
each parity phase if you want checkpoints to return to, and a real semantic
version at Phase 16 close-out once release binaries exist.

```bash
git tag -a v0.3.0 -m "parity: workspace lifecycle (phases 8-10)"
git push origin v0.3.0
```

Guide `01-just.md` section 11 shows a `release` recipe that calls `just test`
and `git tag` from inside a Justfile; that is the mechanical half of this. The
other half is restraint — a tag is a promise that the thing at that commit was
in a state worth naming, so do not tag mid-feature, and do not tag anything
`git diff -- specs/` has not already told you is clean.

### Reviewing your own agent-assisted work

A solo developer has no second reviewer, which means `git` itself has to do
part of that job. Three commands, run in this order, before you read a single
line of generated code:

```bash
git diff --stat
```

Shape before content. A feature that touches forty files when the tasks list
named twelve is worth stopping over before you read any of the forty.

```bash
git diff -- specs/
```

The single highest-value review action available to you, and it takes ten
seconds because it is almost always empty. If it is not empty during an
implementation step, that is section 11's failure mode happening in front of
you — read it before anything else.

```bash
git log --oneline
```

Scoped to one feature's commits (`git log --oneline -- specs/001-config-core`,
or filter by the `(001)` tag in the message), this should show exactly the
sequence from the first section of this appendix: spec, clarify, plan, tasks,
implement, reconcile. A phase that skipped one — reconciliation is the one
everyone skips — shows up here as a gap, not as something you have to remember
to check.

---

## Appendix F — When a phase goes wrong

Every phase in this plan ends with a Check, and the Check is usually enough.
This appendix is for when it fails and the reason is not obvious — the
situations the plan's per-phase Pitfalls sections do not cover because they
are not specific to one phase.

**The build broke and you do not know which change did it.** If the commit discipline in Appendix E
held, this is cheaper than it sounds: the last few commits are individually
small and individually labelled by artefact, so the first move is `git diff
--stat HEAD~1` to see what the most recent commit actually touched, and often
simply reverting the last artefact commit (`git revert HEAD`, or the specific
commit if something landed after it) gets you back to green while you work out
what went wrong at leisure. When the break is further back than one commit,
`git bisect` earns its keep — it is cheap precisely because the history is
made of small, single-purpose commits rather than one large squash:

```bash
git bisect start
git bisect bad HEAD
git bisect good <last known-good commit>
# just build && just test at each step bisect checks out, then:
git bisect good   # or: git bisect bad
```

**The specification turned out to be wrong after implementation.** This will happen — the whole point
of reconciliation in Phase 8.6 is that implementation teaches you things the
spec did not anticipate. The legitimate path is to amend the spec
deliberately, in its own commit, with a note in the commit message saying what
was wrong and why the change is correct now, and then re-run implementation
against the corrected spec for whatever the change affects. The illegitimate
path is editing the spec quietly so it matches whatever the code already does,
with no commit message explaining why and no moment where you actually decided
the new behaviour was right. The difference is not cosmetic: the first
produces a specification you can still trust the next time you read it; the
second produces the outcome Appendix A warns about, a spec that lies, which
section 11 of guide `03-spec-kit.md` treats as worse than having no spec at
all.

**The agent produced a large diff you cannot review.** Do not push through it by reading faster.
Revert the implementation commit, reduce `tasks.md` — or, if the scope problem
started earlier, the plan — to a single user story, and re-run
`/speckit-implement` against that narrower scope. An unreviewable diff is a
process failure, not a review failure: the fix is not a more careful read, it
is a smaller unit of work, because the entire method depends on every
implementation step being small enough that you actually read it against its
requirements. Guide `03-spec-kit.md` section 12.7 makes the same point from
the other direction — user stories exist so that a group of tasks can be
completed, and reviewed, independently.

**A phase's Check does not pass and you cannot see why.** Work backwards through the phase's own
artefacts rather than guessing at the code. Distrust them in this order:
`tasks.md` first, since it is generated from the plan and is the artefact most
likely to contain a task with no file path or no acceptance signal (Phase 8.4
tells you to look for exactly this); then `plan.md`, specifically its
Constitution Check section, since a plan that quietly violated a principle
produces implementation that technically satisfies the tasks but not the
intent; and only last the `spec.md` itself, because if the specification was
actually ambiguous the agent would have surfaced that at `/speckit-clarify`,
not at implementation — a Check failing after clarify usually means something
downstream of the spec went wrong, not the spec itself.

**You have lost the thread after a break of weeks.** Read three things, in this order, before you
touch anything. First, `docs/parity-matrix.md` — it tells you what is done,
what is left, and, if you kept it current, what you were in the middle of.
Second, `git log --oneline` for the last few commits, which tells you the
actual last artefact produced, as opposed to what you remember producing.
Third, the `spec.md` of whichever feature the log says was in progress — read
it fully, not skimmed, because after weeks away you are effectively reviewing
it for the first time again. Guide `03-spec-kit.md` section 12.2 makes the
same point about `HANDOFF.md` and `CLAUDE.md`: the artefacts on disk are the
memory, and this is what "reading the memory back" looks like in practice.

**You want to abandon a feature mid-flight.** Do it in a way that leaves nothing lying. Set the
spec's Status field to `Abandoned` and add one line saying why, in its own
commit — this is the same discipline as reconciliation, applied to a feature
that will not ship rather than one that did. Update the corresponding row in
`docs/parity-matrix.md` to say explicitly that it was decided against, rather
than leaving it blank, which reads later as "not started" instead of
"considered and declined." Do not merge whatever partial implementation exists
onto a long-lived branch; either delete the feature branch or leave it
unmerged as a record, but keep it out of anything that ships. A parity matrix
with an honest "abandoned" row is still trustworthy; one with a silent gap is
not, for the same reason a specification that quietly stops matching its code
is not.

---

## What "done" means

Not "the code works". Done is:

- `just build` and `just test` green,
- every row of `docs/parity-matrix.md` accounted for,
- every `spec.md` describing the behaviour that actually shipped,
- and a chore you used to do by hand now handed to `molfar` and coming back as a
  proposal you were happy to approve.

The last one is the only one that matters to anybody but you. Keep it in sight
through the six weeks of everything that comes before it.
