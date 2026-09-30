# Spec-Driven Development with GitHub Spec Kit

A working guide for rebuilding Grazhda from specifications.

This assumes you are an experienced engineer who has never used spec-kit and is sceptical of
methodologies. That is the right posture. What follows tries to earn the method rather than assert
it, and is explicit about where it costs more than it returns.

Everything stated about the tool is grounded in two sources: the upstream repository at
`https://github.com/github/spec-kit`, and a real `specify init` scaffolding — the `.specify/`
directory and the installed agent skills — inspected file by file. Where the two disagree, the
disagreement is called out rather than smoothed over. Version numbers move quickly here, so when
this guide and the scaffolding on your disk differ, believe your disk.

---

## 1. The problem with how software normally gets specified

The usual life of a specification: someone writes a document before the work starts, it is reviewed
and approved, implementation begins, and within two weeks something turns up that the document did
not anticipate. A sensible decision gets made, written in a ticket comment, and the team moves on.
That happens forty more times. The code ships.

Six months later someone asks why workspace purge refuses to run without an explicit target. The
document says nothing. The ticket comment is in a thread nobody can find. The only remaining
authority is the code, and the code tells you *what* it does but never *why* — you can see the guard
clause, but not the incident that caused it. The specification has become archaeology: a record of
what a group of people believed before they knew anything.

The failure is not that the document was wrong when written. It is that nothing in the process ever
brought it back into agreement with reality. Two common responses are both bad: stop writing specs
and let the code be the truth, which works until you need to change something and cannot tell which
behaviours are intentional and which are accidents; or wrap the document in more process, which
raises the cost of correcting it and therefore guarantees it gets corrected less often.

Spec-driven development takes a third position: the specification is the artefact you maintain, the
code is derived from it, and a divergence is a defect that must be closed on one side or the other.
Grazhda's constitution commits to the hard half of that:

> ### I. Specification Is the Source of Truth (NON-NEGOTIABLE)
>
> No behaviour exists until it is specified. [...]
>
> - Code that contradicts its spec is a defect in the code **or** the spec; the loop is not closed
>   until one of them is corrected.

Note the "or". The principle does not claim the spec is always right. It claims a divergence is
never allowed to just sit there.

The practical difference shows up with an AI coding agent. An agent given a codebase and a vague
instruction produces something plausible. An agent given a specification with thirty-four numbered
requirements produces something you can check. The spec is not documentation that happens to help
the agent; it is the input the agent works from and the standard you judge its output against.

---

## 2. What spec-driven development claims — and what it costs

The claims, plainly.

**Intent survives.** The reason a behaviour exists lives beside the behaviour, in language that does
not require reading Go. Revisiting feature 002 in a year, `spec.md` tells you `--tag` uses OR logic
and inherits project-to-repository because that is what engineers expected, not because it was
easier to implement.

**Review moves earlier and gets cheaper.** Arguing whether merge should replace lists wholesale or
element-wise costs one sentence in a spec. It costs a day once implemented, tested and depended on.
Grazhda's FR-027 settles it before any code exists.

**Work becomes decomposable.** A spec with prioritised, independently testable user stories can be
sliced and shipped a story at a time. The task generator depends on this.

**Agent output becomes checkable** — the most valuable property in practice. You do not review agent
code by reading every line and forming an aesthetic judgement; you ask whether FR-011 through FR-017
each have a failing-then-passing test, and whether anything in the diff was asked for.

Now the costs, which are real.

**The front-loading is not free.** Feature 001's spec — five user stories, thirty-four requirements,
five success criteria, seven edge cases — is a serious morning's work that ends with no running
code. A one-line flag fix does not need a spec directory.

**The loop has latency.** Nine commands, each reading files, thinking, writing files. For
exploratory work where you do not yet know what you want, that latency is pure loss; prototype
first, then specify what the prototype taught you.

**Specs rot too, and rotten specs are worse than none.** The central risk. A stale spec actively
misleads: a reader trusts `FR-023` and builds on a promise the code stopped keeping. Section 9 is
about this, and it is the step everyone skips.

**The tooling is young.** The scaffolding here records `1.0.2.dev0`. Upstream docs and installed
templates already disagree on command naming (section 3). Expect movement.

**It relocates judgement rather than removing it.** You still decide whether merge replaces lists.
The method makes you decide it in a sentence rather than in a function, and makes the decision
visible. If a team is bad at deciding, structured documents will not save it.

The honest summary: this pays off on work that is large, long-lived, done with an agent, or crossing
several people. It does not pay off on small, short-lived, solo, exploratory work. Grazhda — three
processes, four languages, a rebuild to parity — is squarely in the first category.

---

## 3. Installing spec-kit

Spec Kit ships as a Python CLI, `specify-cli`, on PyPI and installable from git. The recommended
installer is `uv`.

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh   # if you don't have uv
uv tool install specify-cli                        # from PyPI
# or pinned, which is what you want on a project you intend to keep:
uv tool install specify-cli --from git+https://github.com/github/spec-kit.git@vX.Y.Z
```

Keep the `v` prefix on the tag. `pipx` works too.

**Checking the installation** is split across subcommands, and documentation lags here. The current
upstream README documents `specify self check` ("Check whether a newer release is available
(read-only)"), `specify self upgrade` (with `--dry-run` and `--tag vX.Y.Z`),
`specify integration list`, and `specify version`. Older releases shipped a bare `specify check`
that verified your environment had the tools it needed; if a tutorial tells you to run it and it
errors, those are the replacements. Run `specify --help` on the version you installed.

### Initialising

To scaffold into an existing directory, which is what a brownfield rebuild needs:

```bash
specify init . --integration claude
```

`--integration` selects which coding agent the scaffolding is written for. Spec Kit supports "30+ AI
coding agents", and the flag decides where the command files land and what form they take. The
`.specify/` directory is identical regardless; only the agent-facing layer changes. Upstream notes
that non-interactive sessions such as CI default to GitHub Copilot unless you pass `--integration`,
so pass it. A run of this command records its choices in `.specify/init-options.json`:

```json
{
  "ai": "claude", "ai_skills": true, "feature_numbering": "sequential",
  "here": true, "integration": "claude", "script": "sh", "speckit_version": "1.0.2.dev0"
}
```

`here: true` means init ran against an existing directory; `script: "sh"` selects the Bash helpers
over PowerShell; `feature_numbering: "sequential"` gives `001-`, `002-` prefixes, where the
alternative `"timestamp"` gives directories like `20260319-143022-user-auth`.

### Skills mode

`ai_skills: true` is the interesting one. Some integrations can install the Spec Kit commands as
**agent skills** rather than slash-command prompt files:

```bash
specify init . --integration claude --integration-options="--skills"
```

In prompt mode you get command files invoked as `/speckit.specify`. In skills mode you get a
directory per command, each holding a `SKILL.md` with frontmatter, which the agent loads as a skill.
This project is in skills mode; `.claude/skills/` holds exactly ten: `speckit-analyze`,
`speckit-checklist`, `speckit-clarify`, `speckit-constitution`, `speckit-converge`,
`speckit-implement`, `speckit-plan`, `speckit-specify`, `speckit-tasks`, `speckit-taskstoissues`.

**The naming discrepancy matters and upstream is quiet about it.** The README writes commands with a
dot (`/speckit.specify`); the installed skills use a hyphen (`speckit-specify`). Neither is a typo.
`.specify/integration.json` records the separator as a per-integration setting:

```json
"integration_settings": { "claude": { "script": "sh", "invoke_separator": "-" } }
```

So the dotted form is the canonical command *identity*, and the separator your agent accepts is an
integration property. The installed templates follow the local convention — `plan-template.md` says
"This template is filled in by the `/speckit-plan` command". Upstream's quickstart acknowledges the
mess only in passing, noting invocation "varies by agent type (e.g., `/speckit.*`, `$speckit-*`, or
`/skill:speckit-*`)" and telling you to "use whichever form your agent exposes." This guide uses the
hyphenated form, because that is what is installed here.

### Extensions, presets and bundles

Recent versions add three composition mechanisms. You will not need them on day one, but they
explain the directory layout. **Extensions** add capability (`specify extension search|add`) — every
installed skill checks `.specify/extensions.yml` for hooks under keys like `hooks.before_specify`
and `hooks.before_converge`, and git branch creation in `/speckit-specify` is now delivered by a
hook rather than built in. **Presets** change how existing workflows behave without adding
capability (`specify preset add`), overriding templates and command text. **Bundles** package
curated sets of both (`specify bundle search|info|install|list|update|remove`, plus
`bundle validate|build --path` for authoring), resolving from a "priority-ordered catalog stack
(project > user > built-in)".

The consequence is that templates are resolved at runtime through a stack, in the order
`.specify/templates/overrides/`, `.specify/presets/templates/`, `.specify/extensions/templates/`,
`.specify/templates/`. That is why `resolve-template.sh` exists.

---

## 4. A guided tour of what `specify init` puts on disk

```
.specify/
├── templates/      spec, plan, tasks, checklist, constitution skeletons
├── scripts/bash/   common, create-new-feature, check-prerequisites,
│                   setup-plan, setup-tasks, resolve-template
├── memory/         constitution.md, .constitution-template.json
├── integrations/   claude.manifest.json, speckit.manifest.json
├── workflows/      workflow-registry.json, speckit/workflow.yml
├── integration.json  init-options.json  feature.json  .gitignore
.claude/skills/speckit-*/SKILL.md   (ten skills)
specs/                              (created as you go)
```

**`.specify/templates/`** holds the skeletons every artefact is generated from — the thing you are
most likely to want to change, though the correct way is an override or preset rather than editing
in place, since a refresh overwrites them (their hashes are recorded in
`.specify/integrations/speckit.manifest.json`).

`spec-template.md` (131 lines) defines prioritised user stories with acceptance scenarios, edge
cases, `FR-###` requirements, key entities, `SC-###` success criteria, and assumptions. Its HTML
comments are instructions to the agent and are worth reading:

> IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance. Each user
> story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them, you
> should still have a viable MVP.

`plan-template.md` (113 lines) is where technology enters: a Technical Context block, a
**Constitution Check** gate, a project structure section, and a Complexity Tracking table.
`tasks-template.md` (252 lines) defines setup, foundational, one phase per user story, then polish,
with task lines in the strict format `[ID] [P?] [Story] Description`. `constitution-template.md`
(50 lines) is a placeholder skeleton. `checklist-template.md` (45 lines) generates reviewer-owned
quality checklists and is pointed about semantics: "`[x]` means the criterion has been reviewed and
satisfied for requirements quality. It does not mean implementation work is complete."

**`.specify/scripts/bash/`** holds deterministic helpers the agent shells out to, so path resolution
and numbering are not left to a language model's judgement. You will rarely edit these; read
`create-new-feature.sh` once. That script (407 lines) computes the next feature number by scanning
`specs/` for the highest numeric prefix and forces base 10 with `$((10#$BRANCH_NUMBER))` so `010`
does not become 8. It supports `--short-name`, `--number` (a *preference* — if taken, it warns and
picks the next free one), `--timestamp`, `--dry-run` and `--json`, and truncates branch names to
GitHub's 244-byte limit with a warning. `check-prerequisites.sh` locates the current feature and
asserts required files exist; flags like `--require-tasks --include-tasks` are how analyze,
implement and converge refuse to run out of order. `setup-plan.sh` and `setup-tasks.sh` prepare
those artefacts and emit JSON paths; `resolve-template.sh` implements the resolution stack;
`common.sh` (918 lines) holds the shared machinery.

**`.specify/memory/`** is the project's long-term memory: `constitution.md`. The file you own most
directly and the one with the widest blast radius. Every other command loads it — the specify, plan,
tasks, clarify, analyze, checklist, implement, converge and taskstoissues skills each contain some
variant of "**IF EXISTS**: Load `.specify/memory/constitution.md` for project principles and
governance constraints."

**State files.** `integration.json` records installed integrations and their settings;
`init-options.json` the init flags; `feature.json` a one-line pointer to the active feature,
`{"feature_directory":"specs/008-automation"}`. **Spec Kit tracks the active feature through that
file, not through your git branch** — upstream's quickstart says so explicitly, and
`SPECIFY_FEATURE` / `SPECIFY_FEATURE_DIRECTORY` override it. It is git-ignored because it is
"per-checkout state rather than something to share."

**`.specify/workflows/`** carries a bundled workflow, `speckit`, "Runs specify → plan → tasks →
implement with review gates". Its `workflow.yml` chains the four core commands with human gates:

```yaml
  - id: review-spec
    type: gate
    message: "Review the generated spec before planning."
    options: [approve, reject]
    on_reject: abort
```

Those gates are the method in miniature. The value is not the automation; it is the enforced pause
where a human reads the spec before anything is planned.

**`.claude/skills/`** is the agent integration directory: ten `SKILL.md` files, 112 to 386 lines
each, 2,476 in total, each with skill frontmatter (`name`, `description`, `user-invocable: true`).
These are generated, their hashes recorded in `claude.manifest.json`, and a refresh rewrites them.
Do not edit them; use an override or preset.

**Which files will you actually edit?** `.specify/memory/constitution.md` constantly, and everything
under `specs/`. Occasionally a template, via `.specify/templates/overrides/`. Almost never the
scripts or skills.

---

## 5. The constitution

The constitution is the one document that outlives every feature. It exists because otherwise the
same architectural argument gets re-litigated in every plan, and the agent — which has no memory
between features — decides differently each time.

Concretely it is a gate. `plan-template.md` contains a Constitution Check section marked
"*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*" The plan skill instructs
the agent to "Fill Constitution Check section from constitution", "Evaluate gates (ERROR if
violations unjustified)", then "Re-evaluate Constitution Check post-design". A violation is either
removed or justified in the plan's Complexity Tracking table.

**What belongs**: principles that constrain every future decision, still true in a year, whose
violation you would want a plan to justify in writing. **What does not**: anything specific to one
feature, anything you expect to change with a dependency release, anything that is really a
preference (style belongs in a linter). The boundary is genuinely hard, and Grazhda shows the
tension: it has a **Technology Constraints** section naming Go 1.26+, Java 25, Spring Boot 4.1 and
Maven. Strictly that is implementation detail; it is there because on this project the runtime split
*is* a governing constraint. Note the price — the constitution had to be amended at 1.1.0 when the
Java version moved.

### Three principles, dissected

**Principle II — Three Processes, One Domain.** After a table assigning `zgard`, `dukh` and `molfar`
their roles:

> Domain logic lives in shared libraries, never in a command handler, a gRPC service method, or a
> REST controller. Those three are adapters; they parse, delegate, and render.
>
> A capability may be exposed by more than one process, but it is implemented once.

Good, because it is *checkable* and it *constrains future decisions*. When feature 006 exposes
health monitoring through both `zgard ws status` and the daemon, this decides where the logic goes
before anyone argues. It is falsifiable — an import of `cobra` inside `libs/` violates it visibly.
Contrast a weaker version, "keep the code well layered", which is unfalsifiable and therefore
constrains nothing.

**Principle III — Everything Is Declarative and Reproducible.** Two of its bullets:

> Any operation that removes data requires an explicit target and an explicit confirmation
> (`--no-confirm` to bypass in CI). Destructive commands never fall back to a default target.
>
> Re-running an operation that has already been applied is safe and reports "already satisfied"
> rather than failing or duplicating work.

The first has a visible scar on it. "Destructive commands never fall back to a default target" is
what someone writes after `ws purge` deleted the wrong workspace — the constitution carrying a
lesson forward so it need not be learned twice. `CLAUDE.md` repeats it as a hard rule, which tells
you what it cost. The second is a broad idempotence requirement that silently shapes dozens of
future specs: cloning in 003, package installs in 005, toolchain upgrades in 007 all inherit it
without re-arguing.

**Principle V — The Terminal Is a User Interface.** Includes:

> Errors name the thing that failed, the reason, and the next action. They go to stderr; data goes
> to stdout.
>
> Every command reports a summary line. Multi-target operations continue past a failure and report
> all failures at the end rather than aborting on the first.

Watch it propagate. Feature 001's FR-011 — "Validation MUST return every problem found in a single
pass rather than stopping at the first" — is not an independent decision; it is principle V applied
to configuration validation. The spec author did not have to think about it and the reviewer can
check it without a discussion.

Some principles are marked `(NON-NEGOTIABLE)`. Use that sparingly; if everything is non-negotiable,
nothing is.

### Amendment and versioning

The constitution skill mandates semantic versioning of the document: MAJOR for "Backward
incompatible governance/principle removals or redefinitions", MINOR for "New principle/section added
or materially expanded guidance", PATCH for "Clarifications, wording, typo fixes". It also requires
a **Sync Impact Report** prepended as an HTML comment, recording the version change, modified
principles, added and removed sections, and deferred placeholders.

Grazhda keeps a visible amendment log as well. Its 1.1.0 entry moves the Java target to 25 / Spring
Boot 4.1 and corrects the Go floor from 1.24+ to 1.26+, giving this reason for the second half: "The
Go correction removes a constraint the repository already contradicted." That is the most
instructive line in the document. The constitution said 1.24+ while every `go.mod` said 1.26; rather
than pretend the document was right, the document was corrected — the same discipline section 9 asks
you to apply to specs. Bump the version whenever you change the text; an unversioned amendment is
indistinguishable from a document nobody maintains.

---

## 6. The loop, command by command

Nine commands. Seven core, three optional, and one of the optional ones (`clarify`) you should treat
as mandatory.

**`/speckit-constitution`** — *consumes* your intent, the existing constitution, the resolved
template; *produces* `.specify/memory/constitution.md` plus a Sync Impact Report. Run once at
project start, then only when a principle genuinely changes. It resolves the template through
`resolve-template.sh` and refuses to continue if resolution fails — it will not fall back to a
single layer. It preserves existing project values, computes the version bump, and validates that no
`[PLACEHOLDER]` tokens remain. *Skip when* nothing about your principles changed; most features
never touch it.

**`/speckit-specify`** — *consumes* a plain-language description; *produces*
`specs/###-name/spec.md`, `specs/###-name/checklists/requirements.md`, and an updated
`feature.json`. The command you spend the most human effort on. It generates a 2-4 word short name,
resolves the feature directory, copies the resolved template, and fills it in. Two constraints
deserve attention. On ambiguity:

> - Make informed guesses based on context and industry standards
> - Only mark with [NEEDS CLARIFICATION: specific question] if: [...] No reasonable default exists
> - **LIMIT: Maximum 3 [NEEDS CLARIFICATION] markers total**
> - Prioritize clarifications by impact: scope > security/privacy > user experience > technical
>   details

The agent is told to decide rather than interrogate, and record what it decided in Assumptions.
Right default — but it means you must **read the Assumptions section carefully**, because that is
where the agent's guesses live and a wrong guess there is silent. Second, the command
self-validates: it writes `checklists/requirements.md` with items like "No implementation details
(languages, frameworks, APIs)" and "Success criteria are technology-agnostic", reviews the spec
against them, and iterates up to three times before warning you. *Skip when* never, for a real
feature.

**`/speckit-clarify`** — *consumes* `spec.md` and the constitution; *produces* edits to `spec.md`.
Nominally optional; treat as required. Its goal: "Detect and reduce ambiguity or missing decision
points in the active feature specification and record the clarifications directly in the spec file."
It scans a taxonomy — functional scope, domain and data model, interaction flow, non-functional
attributes, integrations — marking each Clear, Partial or Missing, then asks at most **five**
questions, one at a time, prioritised by impact, writing each answer into the spec as it goes, and
finishes with a coverage summary classifying every category as Resolved, Deferred, Clear or
Outstanding. If you say you are skipping it, it proceeds "but must warn that downstream rework risk
increases." *Skip when* a throwaway spike, or a spec so mechanical there is nothing to decide.

**`/speckit-plan`** — *consumes* `spec.md`, the constitution, the plan template; *produces*
`plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`. Where technology first
appears. Fill Technical Context (unknowns marked `NEEDS CLARIFICATION`), evaluate the Constitution
Check, generate `research.md` in Phase 0 resolving every unknown, generate design artefacts in
Phase 1, then **re-evaluate the Constitution Check against the design just produced**. That
re-evaluation is the part people drop: a design that passed as a sketch can violate the constitution
once it has concrete file paths. *Skip when* you cannot — tasks, analyze, implement and converge all
require it.

**`/speckit-tasks`** — *consumes* `plan.md` and `spec.md` (both required), plus `data-model.md`,
`contracts/`, `research.md` if present; *produces* `tasks.md`. Generates setup, foundational work
("Core infrastructure that MUST be complete before ANY user story can be implemented"), one phase
per story in priority order, then polish. Each task carries an ID, an optional `[P]` parallel
marker, a story tag, and an exact file path. One default surprises people:

> **Tests**: The examples below include test tasks. Tests are OPTIONAL - only include them if
> explicitly requested in the feature specification.

On Grazhda that is overridden by constitutional principle IV, which requires "at least one automated
test that fails before the implementation exists and passes after" for every functional requirement.
The constitution outranks the template — exactly the sort of conflict it exists to settle.
*Skip when* never, if you intend to use implement.

**`/speckit-analyze`** — *consumes* all three artefacts and the constitution; *produces* a report in
the conversation and **no files**. Must run after tasks produced a complete `tasks.md`. Severities:

> - **CRITICAL**: Violates constitution MUST, missing core spec artifact, or requirement with zero
>   coverage that blocks baseline functionality
> - **HIGH**: Duplicate or conflicting requirement, ambiguous security/performance attribute,
>   untestable acceptance criterion
> - **MEDIUM**: Terminology drift, missing non-functional task coverage, underspecified edge case
> - **LOW**: Style/wording improvements, minor redundancy

The most valuable output is the coverage table (Requirement Key, Has Task?, Task IDs, Notes), a
direct answer to "does every `FR-###` have work assigned", plus an "Unmapped Tasks" list answering
the inverse. *Skip when* the feature is small enough that you have read all three files yourself.

**`/speckit-checklist`** — *consumes* the feature artefacts, the constitution, and your description;
*produces* a file under `specs/###-name/checklists/`. Generates a custom requirements-quality
checklist for a theme you name, after asking up to three (escalating to five) clarifying questions.
Read the semantics carefully: these are **requirements-quality** checklists reviewed by a human, not
implementation progress trackers. They are also a gate — `/speckit-implement` scans every file under
`checklists/`, counts checked and unchecked items, prints a status table, and if anything is
unchecked it stops and asks whether to proceed, treating markers as read-only. *Skip when* most
features; reach for it when a specific risk dimension needs a named reviewer.

**`/speckit-implement`** — *consumes* `tasks.md` and `plan.md` (required), plus the design artefacts
and constitution; *produces* code and ticked checkboxes. Runs the checklist gate, then executes
tasks in dependency order respecting `[P]` markers and phase boundaries. On Grazhda the constitution
adds a hard stop: `just build` and `just test` must be green. *Skip when* you would rather drive
implementation yourself — `tasks.md` is a perfectly good work queue for a human.

**`/speckit-converge`** — *consumes* the three artefacts, the constitution, and the current state of
the code; *produces* new tasks appended to `tasks.md`. The newest and least understood command:

> Close the gap between what a feature's specification, plan, and tasks call for and what the
> codebase currently implements. [...] **append each piece of remaining work as a new, traceable
> task** at the bottom of `tasks.md`.
>
> This is **not** a diff tool and does **not** track changes. It assesses the present state of the
> code relative to the feature's artifacts — no git, no branch comparison, no history.

It must run after implement has run at least once; new IDs continue as `T{M+1:03d}`. Converge is the
closest thing spec-kit has to answering "is this feature actually done", and on a brownfield rebuild
it is the workhorse — it grinds a partially implemented feature toward its spec without regenerating
`tasks.md` and losing your completed checkboxes. *Skip when* implement finished cleanly and analyze
reports full coverage.

**`/speckit-taskstoissues`** — *consumes* `tasks.md` and your git remote; *produces* GitHub issues
titled `T001: <description>`, deduplicated by matching existing issue titles against `\bT\d{3,}\b`.
It refuses to run unless the remote is a GitHub URL. *Skip when* you do not track work in issues.

Upstream documents a short path (constitution, specify, plan, tasks, implement) for smaller features
and a full path adding clarify, checklist, analyze and converge. Grazhda's constitution mandates its
own sequence and ends with a step no command performs: "**Reconcile** — update `spec.md` and
`plan.md` to match what shipped, record deviations, and check the feature off in
`docs/parity-matrix.md`."

---

## 7. How to write a specification that is actually useful

This is the heart of it. The commands are mechanism; this is the skill.

### What and why, never how

A specification describes observable behaviour and the reason it is wanted, not mechanism. The test:
if a sentence would have to change because you swapped a library, it belongs in the plan.

A bad requirement:

> **FR-0XX**: The loader MUST use `gopkg.in/yaml.v3` to unmarshal the config into a `Config` struct,
> and MUST call `validator.New().Struct(cfg)` to check struct tags.

Everything there is mechanism. It names a library, a type, an approach; it tells a reviewer nothing
about what the system does; it forecloses choices the plan should be making. Worse, it cannot fail
review — whatever the code does, the sentence describes it. How Grazhda's real spec covers the same
territory:

> - **FR-004**: The system MUST report a YAML parse failure with the file path and the parser's
>   reason.
> - **FR-011**: Validation MUST return every problem found in a single pass rather than stopping at
>   the first.
> - **FR-017**: Every validation message MUST address the offending field by its path, including
>   indices (for example `workspace[1].projects[0].branch`).

Three sentences, no libraries, each one a test you could write today. FR-017 makes a UX promise
precise enough to assert on without saying how paths are constructed.

The same rule governs success criteria, which the template says "must be technology-agnostic and
measurable": *"**SC-001**: A configuration file describing 3 workspaces, 10 projects and 100
repositories loads and validates in under 100 ms."* Satisfiable in Go, Rust or Python. That is the
point.

Neutrality stops at the domain boundary. "YAML" appears in Grazhda's spec because the file format is
a user-facing contract — engineers edit that file by hand. Naming the YAML *parser* would be
mechanism; naming YAML is domain. Judge by whether the user can observe it.

### Prioritised, independently testable user stories

The template demands priorities (P1, P2, P3) and independence. Grazhda's configuration feature has
five stories, and the priorities do real work. Two are P1, with justifications like:

> **Why this priority**: A malformed config that fails halfway through a clone of forty repositories
> is worse than one that refuses to start. Validation is what makes every later operation safe.

That is an argument, not a label. It says *why* validation cannot be deferred, in terms of
consequence to a user. When someone later proposes shipping without it, the counter-argument is
already written.

Every story also carries an Independent Test, which is what makes the slicing real:

> **Independent Test**: Write a `config.yaml` describing two workspaces with projects and
> repositories, load it, and confirm every field is available with the documented defaults applied —
> no other feature required.

"No other feature required" is the whole claim. If you cannot write that sentence for a story, the
story is not independent and your priorities are decoration. Acceptance scenarios stay concrete:
*"**Given** `$GRAZHDA_DIR` is set, **When** the configuration path is resolved, **Then** it is
`$GRAZHDA_DIR/config.yaml`."* The shape carries: a targeting story ("Act on a subset of my
repositories without naming each one", tested by confirming `-t backend` selects exactly the
expected set), a registry story ("Install a package and everything it needs, once", tested with a
three-level dependency chain installed exactly once in dependency order).

### Stable `FR-###` identifiers, and why numbering matters

The identifiers are the join key across the whole system. Grazhda's constitution requires that
"Tests, tasks, and commit messages reference those identifiers." `/speckit-analyze` builds its
coverage table on them. `/speckit-converge` loads "Functional Requirements (FR-###)" specifically. A
quality gate reads "every `FR-###` covered by a test". Without identifiers none of that is
mechanisable — you would be matching on prose.

**Never renumber.** If FR-019 is wrong, fix its text; if obsolete, mark it obsolete rather than
deleting and shuffling up. A renumbering silently invalidates every test comment, task and commit
message referencing the old number.

**Group without renumbering.** Grazhda organises thirty-four requirements under bold subheadings —
"Loading and path resolution", "Configuration model", "Validation", "Clone command rendering",
"Querying", "Safe modification", "Environment overrides" — while the numbers run straight through
FR-001 to FR-034. Headings give structure; numbers stay stable.

**One assertion per requirement.** FR-019 says exactly one thing: *"`Branch` MUST resolve to the
repository's branch when set and the project's branch otherwise."* A requirement asserting three
things cannot be marked covered, because two of the three might be tested and one not.

### Measurable success criteria

Requirements say what the system does; success criteria say how you know the feature was worth
building. They are usually written as noise — "the system is fast" — and are then useless. Grazhda's
cover four different kinds of measure:

> - **SC-002**: Every one of the validation rules FR-012 through FR-017 is demonstrated by at least
>   one test that fails before the rule is implemented.
> - **SC-004**: No sequence of replace or merge operations, including ones that fail validation or
>   are interrupted mid-write, can leave the engineer without either a valid configuration or a
>   backup of the previous one.
> - **SC-005**: 100% of the configuration surface documented in `docs/v1-baseline/CONFIG.md` is
>   expressible in the new model.

SC-001, quoted above, is a performance budget with a stated workload — the workload is what makes it
checkable. SC-002 makes principle IV's test-first rule verifiable here and names the range it
covers. SC-004 is a safety invariant quantified over *all* sequences including interrupted ones, far
stronger than "writes are atomic". SC-005 is a parity criterion pointing at a baseline document.
Note that `/speckit-converge` includes "only items requiring buildable work" and excludes
"post-launch outcome metrics and business KPIs" — "reduce support tickets by 50%" is legitimate but
cannot be converged toward.

### Edge cases as a first-class section

The template gives edge cases their own mandatory section, and this is where a spec earns its keep.
Requirements describe the happy path competently; edge cases are where the design decisions actually
live. Grazhda names a condition and states the resolved behaviour with an arrow:

> - The config file exists but is empty → it loads as a configuration with no workspaces; validation
>   reports nothing, and commands that need a default workspace fail with "no default workspace
>   found".
> - No workspace is marked `default: true` and none is named `default` → resolving the default
>   target fails with an actionable message naming `--name`.
> - A repository name contains `/` → its destination directory is decided by the effective structure
>   mode (specified in feature 002), not by the loader.
> - `~` appears at the start of a workspace path → it expands to the engineer's home directory; a
>   `~` elsewhere in the path is left alone.
> - The backup file cannot be written (read-only directory) → replace and merge both abort before
>   touching the original.

Three things to copy. Every entry has a *decided* outcome; none says "TBD". The third explicitly
hands a case to another feature and names it — that is how you keep a spec bounded without leaving a
hole. The last describes a failure of the safety mechanism itself, exactly the case an
implementation forgets. A useful generator: for each requirement, ask what happens when the input is
empty, malformed, duplicated, oversized, concurrent, or when the resource it needs is unavailable.
Most produce a shrug; the rest are your edge cases.

### Assumptions instead of ambiguity

When the description does not settle something, the choice is between leaving it vague and deciding
explicitly. Vagueness always loses: it gets resolved during implementation, invisibly, by whoever is
typing.

> - The configuration file is edited by a single engineer on a single machine; concurrent writers
>   are out of scope, and the atomic-rename guarantee is sufficient protection.
> - The `zgard` and `general` blocks are inspection metadata: they are parsed, validated as
>   well-formed and queryable, but no operation reads them. This is inherited deliberately from the
>   baseline (see `docs/v1-baseline/GAPS.md`).
> - The default daemon port `50501` is inherited from the baseline.

Each is falsifiable and carries scope. The first draws a boundary and says why the weaker guarantee
is acceptable. The second is remarkable: it documents that part of the configuration surface *does
nothing*, cites the baseline gap analysis, and marks the inheritance as deliberate. Without that
sentence the next engineer either wires those fields up (breaking parity) or deletes them (breaking
users). Read this section first when reviewing an agent-drafted spec — it is where the guesses are,
by design.

### Resolve every `NEEDS CLARIFICATION` before planning

A marker is a hole; the template shows the form: *"**FR-006**: System MUST authenticate users via
[NEEDS CLARIFICATION: auth method not specified - email/password, SSO, OAuth?]"*. Grazhda's
constitution makes clearing them a step in the loop and a quality gate: "Feature done | Spec
reconciled; parity matrix updated; no `NEEDS CLARIFICATION` left".

Resolve them *before* planning rather than during, because otherwise the plan makes the decision for
you, in technical terms, buried in `research.md`, where nobody who reads only the spec will ever see
it. The generated `checklists/requirements.md` includes "No [NEEDS CLARIFICATION] markers remain" as
an explicit item. The finished feature 001 spec contains zero. That is what a plan-ready spec looks
like.

---

## 8. What the plan and the tasks add

**The plan is where technology appears.** `plan.md` is the first artefact allowed to name a
language, framework or library. Its Technical Context block is a fixed set of slots —
Language/Version, Primary Dependencies, Storage, Testing, Target Platform, Project Type, Performance
Goals, Constraints, Scale/Scope — with unknowns marked `NEEDS CLARIFICATION` and resolved in
Phase 0. For Grazhda's configuration feature the block writes itself from the constitution (Go
1.26+, a Go workspace over `libs/`, `cli/` and `daemon/`, no network in unit tests), which is
exactly the labour a constitution is supposed to save. The plan also produces `research.md`
(decisions and rationale), `data-model.md` (entities as designed, versus the spec's
technology-free Key Entities), `contracts/` and `quickstart.md`.

**The Constitution Check gate** runs twice: before Phase 0 and after Phase 1. Violations are removed
or justified in a table whose columns are pointed — Violation | Why Needed | Simpler Alternative
Rejected Because. The third column does the work. Justifying complexity in the abstract is easy;
writing down why the simple option specifically fails is hard. Most unnecessary complexity dies in
that cell. Grazhda makes an unjustified violation blocking.

**Tasks are grouped by user story, not by layer.** This runs against habit. The obvious grouping is
all the models, then all the services, then all the commands. The template refuses:

> Tasks MUST be organized by user story so each story can be:
> - Implemented independently
> - Tested independently
> - Delivered as an MVP increment

Setup, then foundational ("No user story work can begin until this phase is complete"), then one
phase per story in priority order, then polish — each story phase ending in a checkpoint: "At this
point, User Story 1 should be fully functional and testable independently."

Layer grouping produces a long stretch where nothing is demonstrable and every task is equally
unfinished. Story grouping produces a sequence where each phase ends in something a user could use.
On a rebuild-to-parity project that matters twice over: you measure progress against the baseline
feature by feature instead of guessing at percentages. The `[P]` marker identifies tasks touching
different files with no dependency between them.

---

## 9. The step everyone skips: reconciliation

Implementation always teaches you something the spec did not know. The clone template needs a
variable nobody anticipated. A validation rule turns out unenforceable as written. An edge case
resolves differently once you see real data.

Three choices at that moment, two legitimate. **Change the code to match the spec**, when the spec
was right and the implementation drifted. **Change the spec to match the code**, when implementation
revealed the spec was wrong — explicitly, recording what changed and why. **Do nothing and
remember** — the one everybody takes, and how the whole method fails.

Grazhda's constitution makes reconciliation step 7 of the loop and repeats it as a quality gate.
`CLAUDE.md` states the consequence in a sentence worth memorising:

> A feature is not done because the code works. It is done when the code works *and* the
> specification describes it.

**Why lying specs are worse than no specs.** With no specification, a reader knows they must read
the code — expensive but honest, the cost visible and paid up front. With a specification that lies,
a reader believes they have understood the system without reading the code. They build on `FR-023`
and discover in production that the error message changed two releases ago. The cost is invisible,
deferred, and paid by someone who had no way to know. The failure is self-reinforcing: the first
person burned stops trusting all specs, so nobody maintains them, so they all go stale.

There is a second-order cost with agents. If your agent reads `spec.md` as its source of truth and
the spec is wrong, it will confidently generate code that contradicts the working system, with the
fluency that makes wrong answers convincing.

**How to do it.** Re-read the spec against the diff. For each `FR-###` confirm the shipped behaviour
matches the sentence; for each `SC-###` confirm you can demonstrate it; for each edge case confirm
the stated outcome is what happens. Where they differ, decide which side is wrong and fix that side,
keeping the identifier and changing the text. Add what implementation taught you — new edge cases
belong in the Edge Cases section, not a commit message; wrong assumptions get corrected, not
deleted. Update the status line (feature 001 is headed `**Status**: Implemented`, which tells a
reader whether they are looking at a promise or a description) and the parity matrix.

`/speckit-converge` helps with half of this: it finds work the code has not done. It does not find
the inverse — behaviour the code has that the spec never described. That direction is still a human
reading a diff.

---

## 10. Brownfield: specifying a system that already exists

Grazhda has a working implementation and the rebuild must reach parity. That changes the method in
four specific ways.

**Baseline the current behaviour first.** Before writing a single spec, write down what the existing
system does. Grazhda keeps this at `docs/v1-baseline/`: `README-v1.md`, `CLI.md`, `CONFIG.md`,
`QUICK-START.md`, `DEVELOPMENT.md`, `dukh.proto`, the old config and package templates, plus
`GAPS.md` and `feature-ideas.md`. The critical property is that a baseline is **descriptive, not
aspirational**. `GAPS.md` says so itself: "It intentionally describes the current behavior, not the
desired future behavior." That discipline produces entries like:

> | `workspace.structure` validation | Runtime treats any value other than `list` as `tree`;
> validation does not reject unknown values. | A typo silently falls back to tree-style paths. |
>
> | `zgard ws status` targeting | `--name` filters the dukh snapshot by workspace. Inherited
> `--project-name`, `--repo-name`, and `--tag` flags are accepted but not applied. | Users may
> expect the same filters as `init`, `pull`, `exec`... |

Current behaviour plus consequence, no opinion about what should happen. That neutrality is what
makes the next step possible.

**Write specs that describe intent, not transcribe the old code.** The temptation is to turn each
old function into a requirement. Resist it: that produces a worse copy of the code, inheriting every
accident, unreviewable because there is nothing to review it against. Ask instead what each
behaviour was *for*. The baseline silently treats anything that is not `list` as `tree`; the new
spec does not transcribe that:

> - **FR-016**: Validation MUST reject any `structure` value other than `tree` or `list` at both
>   workspace and project level, naming the permitted values.

A deliberate behaviour change, visible *because* the baseline recorded the old behaviour first.
Without `GAPS.md` you could not tell whether FR-016 was a fix or an unnoticed regression.

**Keep a parity matrix.** `docs/parity-matrix.md` tracks every baseline capability against the
rebuild:

| Baseline capability | Feature | Spec | Domain impl | CLI impl | Test |
| :--- | :--- | :---: | :---: | :---: | :---: |
| `ws init` (clone, skip existing, continue on failure) | 003 | ✅ | ✅ | ⬜ | ✅ |
| `ws purge` (explicit target, confirm, dry-run) | 003 | ✅ | ✅ | ⬜ | 🟡 |
| `-t/--tag` with OR logic and project→repo inheritance | 002 | ✅ | ✅ | ⬜ | 🟡 |

Two choices make this useful rather than decorative. Separating spec from implementation from test
lets a row read honestly as ✅/✅/⬜/✅ — specified, domain logic built, CLI not wired, tested — a real
state a single percentage would hide. And the legend (✅ done, 🟡 partial, ⬜ not started) forces
"partial" to be its own answer rather than an optimistic ✅. The constitution makes it binding, and
updating it is part of done.

**Decide deliberately which quirks to carry and which to close.** Every existing system has
behaviours nobody designed. Three options, and the only wrong move is not choosing. *Carry it, and
say why*: Grazhda keeps the default daemon port `50501` and the inert `zgard` and `general` config
blocks, both recorded in Assumptions as inherited, because the next reader cannot otherwise tell an
inheritance from an oversight. *Close it, and specify the new behaviour*: FR-016 closes the
`structure` gap; FR-032 closes a parsing gap — "A `DUKH_PORT` value that is not an integer MUST be
ignored rather than applied." *Defer it, and record the deferral*: the `ws status` targeting gap can
sit in the parity matrix as 🟡 until a feature addresses it.

One more brownfield advantage: you have a running oracle. Ambiguity in a greenfield spec must be
resolved by argument; here you can go and look. When you cannot decide whether `--tag` should be OR
or AND, run the old binary — then write down which you chose and why, because "it matches v1" is a
reason, and a reason recorded is a reason that survives.

---

## 11. Working with an AI agent inside this loop

Spec-driven development and agent-assisted coding are separately useful and jointly much more
useful, because a specification is the review standard agent output otherwise lacks.

### Who owns what

**The human owns the constitution.** One wrong sentence propagates into every future feature. An
agent can draft the wording; the principles must be yours.

**The human owns the specification.** Not necessarily the typing — `/speckit-specify` drafts well —
but the content, and specifically the decisions: priorities, edge case resolutions, assumptions,
success thresholds. When the agent writes "System MUST retain data for 90 days", that number came
from nowhere. Confirm it or change it.

**The human owns acceptance.** Whether the feature is done is not the agent's call.

**The agent owns drafting the plan** — deriving Technical Context, research, data model and
contracts is exactly the thorough, tedious work agents are good at; review the Constitution Check
and Complexity Tracking and skim the rest. **The agent owns the task breakdown**, which is
mechanical: check that every `FR-###` appears somewhere and that story grouping was respected.
**And the agent owns the code**, under review against the spec.

### Reviewing agent output against a spec

Do not review agent-written code the way you review a colleague's pull request; you will read for
style and idiom and miss what matters. Review against the specification instead.

Work requirement by requirement: for each `FR-###`, find the code that implements it and the test
that proves it. If you cannot find the test, the requirement is not done regardless of what the
checkbox says. That is exactly Grazhda's gate — "every `FR-###` covered by a test" — and it is
mechanical, which is what makes it reliable when you are tired.

Then work the other direction. Read the diff and ask, for each meaningful chunk, which requirement
asked for this. Anything unattributable is scope creep, an implicit requirement that should be
written down, or a misunderstanding. All three want action. Then check the edge cases by hand:
agents implement the happy path reliably and the seventh edge case unreliably, and "the backup file
cannot be written → replace and merge both abort before touching the original" is precisely the kind
of case that gets a plausible implementation and no test. Run `/speckit-analyze` for the coverage
table, and treat its "Unmapped Tasks" section as a list of things to explain.

### The failure mode: the agent edits the spec to match its code

The most important paragraph here. An agent asked to make the spec and the code agree will sometimes
take the cheaper path and change the spec. It is not deception; it is an optimiser finding the
shortest route to a stated goal. But the result is a specification quietly rewritten to describe
whatever was built, which destroys the entire value of the method — you now have a document that can
never disagree with the code, and therefore can never catch anything.

`/speckit-converge` is explicitly designed against this. It treats the spec as read-only intent —
"Read `spec.md`, `plan.md`, and `tasks.md` as the **sole source of intent**" — and its output is
always new *tasks*, never edits to requirements. The gap gets closed by changing the code, unless a
human decides otherwise.

Practical defences. **Keep spec edits in their own commits**: a commit touching `spec.md` and
`libs/` together hides a requirement change inside an implementation change. **Diff the spec before
merging** — `git diff -- specs/` on every feature branch, ten seconds, the single highest-value
review action in this method. **Watch for requirements getting weaker**; the signature is softening,
where "MUST" becomes "SHOULD", a specific threshold becomes "reasonable performance", a decided edge
case becomes "behaviour is undefined". **Watch for silently renumbered requirements**: if FR-019 has
different content than last week, either it was legitimately amended or something was inserted and
everything shifted, and the second is a bigger problem than it looks. **Be specific in your
instruction** — "make the implementation satisfy FR-019" and "make FR-019 and the code agree" are
different instructions with different outcomes; say which side you expect to move.

### Practical habits

Give the agent the spec, not a summary; it reads faster than you. When you disagree with generated
code, check first whether the spec was ambiguous — often it was, and the durable fix is in the spec.
Do not let the agent run the whole loop unattended; the gate steps in `workflow.yml` exist because
the cost of a wrong spec compounds through every later stage. Keep a `CLAUDE.md` for the operational
rules that are not principles — Grazhda's carries "`libs/` is the domain. It must never import
`cobra`" and "Do not run `go mod tidy` casually — the dependency set is pinned deliberately" — and
note it states its own subordination: the constitution "supersedes habit, preference and prior art,
including this file."

---

## 12. Running the loop in Claude Code

Section 11 was about working with any agent. This section is about the specific mechanics of doing
it in Claude Code: how spec-kit reaches the tool, where to put which instructions, how to configure
permissions so the loop runs without babysitting, and — the part that most affects the quality of
your specifications — how to manage context.

### 12.1 How spec-kit reaches Claude Code

```bash
specify init . --integration claude
```

The Claude integration installs one `SKILL.md` per command under `.claude/skills/`:

```
.claude/skills/
  speckit-constitution/SKILL.md
  speckit-specify/SKILL.md
  speckit-clarify/SKILL.md
  speckit-plan/SKILL.md
  speckit-tasks/SKILL.md
  speckit-analyze/SKILL.md
  speckit-checklist/SKILL.md
  speckit-implement/SKILL.md
  speckit-converge/SKILL.md
  speckit-taskstoissues/SKILL.md
```

A skill is a directory containing a `SKILL.md` of instructions. Claude Code exposes each one as a
slash command named after the directory, so `speckit-specify/SKILL.md` becomes `/speckit-specify`.
Custom commands and skills are the same mechanism now — a file at `.claude/commands/deploy.md` and a
skill at `.claude/skills/deploy/SKILL.md` both create `/deploy`. The skill form is what spec-kit
uses because it allows a directory of supporting files and frontmatter controlling whether you or
Claude invokes it.

**The name separator will confuse you once.** Upstream documentation writes `/speckit.specify` with
a dot. The installed Claude integration uses a hyphen: `/speckit-specify`. The dotted form is the
command's canonical identity; the separator is a property of the integration, recorded in
`.specify/integration.json` as `invoke_separator`. Other agents use other forms again. Believe your
disk over any document, including this one: type `/spec` and let completion show you what exists.

**What actually happens when you run one.** Nothing magical, which is worth understanding because it
is what you debug when it misbehaves. The skill body is loaded as instructions for that turn. Those
instructions tell Claude to run the bash scripts in `.specify/scripts/bash/` — `create-new-feature.sh`
to make the feature directory, `check-prerequisites.sh` to confirm the artefacts a phase depends on
exist, `setup-plan.sh` and `setup-tasks.sh` to resolve paths — then read the matching template from
`.specify/templates/`, and write the result into `specs/<feature>/`. The scripts do path resolution
and file creation; the model does the thinking; the templates constrain the shape of the output.

That means a spec-kit command failing is usually a script failing, and the error will say so. It
also means you can run those scripts yourself when you want to see what they do.

### 12.2 `CLAUDE.md` and the constitution are different documents

Claude Code automatically loads `CLAUDE.md` from the repository root into every session. The
constitution is loaded only when something reads it. That difference should decide what goes where.

**The constitution holds principles** — the constraints that every future decision is checked
against, versioned, amended deliberately, and referenced by every `plan.md`'s Constitution Check.

**`CLAUDE.md` holds operational facts** — the things an agent needs in front of it on every turn to
avoid wasting a cycle. For Grazhda that means the reading order, the build commands, and the handful
of rules that are expensive to rediscover:

```markdown
# Working in this repository

Read before changing anything, in this order:
1. `HANDOFF.md` — where the rebuild stands and what to do next
2. `.specify/memory/constitution.md` — the governing principles; it supersedes
   habit, preference and prior art, including this file
3. `docs/architecture.md` — the three-process design and layering rules
4. `specs/###-*/spec.md` — the specification for whatever you are touching

## Hard rules

- `libs/` is the domain. It must never import `cobra`, and must never shell out
  except through `executor.Executor`. A test that needs real `git`, real network
  or a real `$HOME` means the design is wrong.
- Do not run `go mod tidy` casually — the dependency set is pinned deliberately.
- Output symbols are fixed: ✓ ok, ✗ failed, ⚠ drifted, ⏭ skipped.

## Commands

just build   just test   just lint   just generate
```

Two disciplines keep this useful. **Keep it short** — it is paid for on every single turn, so a
thousand-line `CLAUDE.md` is a thousand lines of tax on every question you ask. Point at the
constitution rather than duplicating it. And **state its subordination explicitly**, as the example
above does: when the two documents disagree, the constitution wins, and saying so in the file that
is always loaded is what makes that stick.

`/init` will generate a starting `CLAUDE.md` by inspecting the repository. Treat its output as a
draft. It describes what the code does; you want the rules that are not visible in the code.

### 12.3 Permissions, so the loop runs without babysitting

By default Claude Code asks before running commands and editing files. During a spec-kit loop that
means a prompt for every script invocation, which trains you to approve without reading — the worst
possible outcome.

Settings files are read in this precedence, highest first:

| Level | File | Scope |
| :--- | :--- | :--- |
| Managed | `managed-settings.json` (MDM or console) | Your organisation |
| Command line | `claude --settings …` | This session |
| Project local | `.claude/settings.local.json` | You, this project |
| Shared project | `.claude/settings.json` | Everyone who clones |
| User | `~/.claude/settings.json` | You, every project |

Commit `.claude/settings.json` with the allow-list this project's loop actually needs, and keep
personal exceptions in `.claude/settings.local.json`, which is git-ignored:

```json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "permissions": {
    "allow": [
      "Bash(bash .specify/scripts/bash/*)",
      "Bash(just *)",
      "Bash(go build *)",
      "Bash(go test *)",
      "Bash(go vet *)",
      "Bash(mvn -B *)",
      "Bash(git status)",
      "Bash(git diff *)"
    ],
    "deny": [
      "Read(./.env)",
      "Read(./.env.*)"
    ]
  }
}
```

Settings files are strict JSON — a trailing comma or a `//` comment is a syntax error and the file
is reported as broken at start-up. Run `/status` inside a session to confirm what loaded.

Notice what is *not* in the allow-list: `git commit`, `git push`, and anything that writes outside
the repository. Let those keep asking. The method depends on you seeing what changed, and the
commits are where you see it.

Resist the blanket bypass. It is available and it feels like productivity, but a loop whose whole
value is a human checking the output against a specification does not benefit from removing the
human. Note also that the `auto` and `bypassPermissions` values for `permissions.defaultMode` do not
take effect from project or local settings — they must be set in user or managed settings, or passed
per session — which is a deliberate guard rail worth leaving in place.

### 12.4 Context discipline is what protects specification quality

This is the most important subsection here, and the least obvious.

A specification drafted in a session that is already full of implementation detail will leak
implementation detail. Not because the model is careless, but because everything in the context
window is evidence about what you want. If the last forty minutes were spent debugging a YAML
decoder, the spec you draft next will mention YAML decoding — and section 7's rule about
technology-neutral language fails not through ignorance but through contamination.

The defence is mechanical: **one phase per session, `/clear` between phases.**

```
/clear   →  /speckit-specify …     →  read spec.md, revise      →  commit
/clear   →  /speckit-clarify       →  answer the questions      →  commit
/clear   →  /speckit-plan          →  read the Constitution Check  →  commit
/clear   →  /speckit-tasks         →  skim for unattributed tasks  →  commit
/clear   →  /speckit-implement     →  just test, review the diff   →  commit
/clear   →  reconcile the spec     →  update the parity matrix     →  commit
```

This works precisely because the artefacts on disk are the memory. That is the whole point of the
method: `spec.md`, `plan.md` and `tasks.md` carry everything the next phase needs, so the context
window does not have to. A loop that requires a long conversation to stay coherent is a loop whose
artefacts are too thin — and that is useful diagnostic information, not an argument for keeping the
conversation.

`/compact` is the wrong tool at a phase boundary. It preserves a summary of the conversation, which
is exactly the contamination you are trying to drop. Use it mid-phase when a single long
implementation is running out of room; use `/clear` between phases.

When you need a specific file in context, reference it explicitly rather than relying on the model
to find it: `@specs/001-config-core/spec.md`. Being explicit costs one line and removes a whole class
of "it was looking at the wrong feature" confusion.

### 12.5 Plan mode at the implementation gate

For a large feature, run `/speckit-implement` in plan mode first. You get the intended approach —
which files, in what order, satisfying which requirements — before anything on disk changes. Reading
that against `tasks.md` takes two minutes and catches the misunderstandings that are expensive to
unpick after twelve files exist.

This is the same gate section 6 describes, expressed in the tool: the artefact to review before
approving is the plan, not the diff.

### 12.6 A worked session

What one feature actually looks like, from an empty `specs/` directory to a committed,
reconciled feature. Everything here is either a thing you type or a thing you read.

```bash
# Session 1 — specify
/clear
/speckit-specify configuration core: the declarative description of an
  engineer's workspaces, projects, repositories and daemon settings, plus
  everything needed to read, validate, query and safely change it
```

Read the generated `specs/001-config-core/spec.md` in an editor, not in the terminal. Check it
against the section 7 rules: no implementation detail, independently testable stories, one assertion
per `FR-###`, measurable `SC-###`, real edge cases. Edit it yourself where it is wrong — you own
this document, and hand-editing is not cheating.

```bash
git add specs/001-config-core/spec.md
git commit -m "spec(001): configuration core"

# Session 2 — clarify
/clear
/speckit-clarify
```

Answer every question. This is the part only you can do. Commit again.

```bash
# Session 3 — plan
/clear
/speckit-plan
```

Read the Constitution Check section first, before anything else in the file. If it says a principle
is violated, the plan changes or the violation is justified in writing in the Complexity Tracking
table. Waving one through here is how a constitution dies. Commit.

```bash
# Session 4 — tasks
/clear
/speckit-tasks
```

Skim for tasks with no file path and tasks that map to no requirement. Commit.

```bash
# Session 5 — implement
/clear
@specs/001-config-core/spec.md
@specs/001-config-core/tasks.md
/speckit-implement
```

Then, before reading any code:

```bash
just build && just test
git diff --stat
git diff -- specs/          # must be empty
```

That third command is ten seconds and it is the single highest-value review action in the method. If
the specification changed during implementation, read the change before anything else.

Then review requirement by requirement, as section 11 describes. Commit the implementation on its
own.

```bash
# Session 6 — reconcile
/clear
```

Update `spec.md` to describe what actually shipped, move its Status from `Draft` to `Implemented`,
tick the rows in `docs/parity-matrix.md`, and commit that separately. Six sessions, six commits, and
a history in which you can see exactly which phase introduced which change.

### 12.7 Subagents: when to fan out

Claude Code can dispatch subagents that work in parallel with their own context. Custom ones live in
`.claude/agents/`. They are genuinely useful in this loop, and genuinely harmful in the wrong place.

**Worth fanning out.** Independent user stories, *after* the plan and tasks are fixed and you have
read them — the task list is grouped by story precisely so the groups do not collide. Verification
passes, which parallelise well and have no side effects: "for every `FR-###` in this spec, find the
test that covers it and report the ones with none" is an excellent subagent job. Research, such as
reading the old implementation and reporting what a command actually does.

**Not worth fanning out.** Drafting the specification — it needs your judgement and one coherent
voice, and merging two drafts is worse than writing one. Anything where two agents would touch the
same package: Go and Java both fail in confusing ways when two writers race on one file, and the
failure looks like a compilation error rather than a coordination error.

The practical rule: fan out on *reading*, serialise on *writing*.

### 12.8 Making the gates mechanical with hooks

Everything in section 12.6 is a discipline, and disciplines erode. Some of them can be made
automatic.

Claude Code supports hooks configured in `.claude/settings.json`, which fire at defined points in a
session. Check the hooks reference for the current event names and configuration shape before
writing one — this is an area that moves.

Two that pay for themselves here, in increasing order of how much you should trust me about them:

**A pre-commit hook, in git, not in Claude Code.** This one needs no tool-specific knowledge and
cannot break:

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

That single hook enforces the discipline that makes `git diff -- specs/` meaningful, and it works
regardless of which agent, or which human, produced the change.

**A session-stop hook that runs `just test`.** More convenient, more tool-specific, and worth
setting up once you are comfortable with the hooks reference. The value is that a session cannot end
quietly with a red build.

### 12.9 When it goes wrong

**The agent is thrashing on one requirement.** Do not re-prompt harder. Go and read the requirement.
An agent looping on a task is usually a signal that the requirement is ambiguous, contradicts
another one, or depends on something that was never specified. The fix is upstream, in a document,
not in a better prompt.

**The diff contains things no requirement asked for.** Scope creep, an implicit requirement that
should be written down, or a misunderstanding. All three want action, and all three are invisible
unless you read the diff by chunk and ask "which `FR-###` is this?".

**The specification changed and you did not ask for it.** Section 11 covers why this happens. In
Claude Code the defence is the commit discipline in 12.6 and the pre-commit hook in 12.8: make the
change impossible to hide, and it stops mattering whether it was deliberate.

**The session has gone on too long.** Symptoms: the model starts re-reading files it already read,
answers get vaguer, it forgets a constraint you stated twenty minutes ago. That is a full context
window, and the fix is `/clear` plus a pointer to the artefacts — which works, because the artefacts
are the memory. If it does not work, your artefacts are too thin, and that is the real finding.

---

## 13. One-page reference

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
uv tool install specify-cli          # or --from git+https://github.com/github/spec-kit.git@vX.Y.Z
specify version                      # confirm what you installed
specify self check                   # newer release available?
specify integration list             # integrations this version supports
specify init . --integration claude --integration-options="--skills"
```

Older releases had a bare `specify check`; current docs use `specify self check`. Run
`specify --help` on your version. Commands are written `/speckit.name` upstream and `/speckit-name`
in a hyphen-separator integration such as this one.

| # | Command | Consumes | Produces | Optional |
| :--- | :--- | :--- | :--- | :--- |
| 0 | `/speckit-constitution` | intent, template | `.specify/memory/constitution.md` | once per project |
| 1 | `/speckit-specify` | description | `spec.md`, `checklists/requirements.md` | no |
| 2 | `/speckit-clarify` | `spec.md` | edits to `spec.md` (≤5 questions) | in name only |
| 3 | `/speckit-plan` | `spec.md`, constitution | `plan.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md` | no |
| 4 | `/speckit-checklist` | feature artefacts | `checklists/<theme>.md` | yes |
| 5 | `/speckit-tasks` | `plan.md`, `spec.md` | `tasks.md` | no |
| 6 | `/speckit-analyze` | all three | a report, no files | recommended |
| 7 | `/speckit-implement` | `tasks.md`, `plan.md` | code, ticked tasks | no |
| 8 | `/speckit-converge` | artefacts + code state | new tasks appended | when gaps remain |
| — | `/speckit-taskstoissues` | `tasks.md`, git remote | GitHub issues | yes |
| 9 | **Reconcile** | code, `spec.md` | corrected `spec.md`, parity matrix | **no** |

Step 9 is not a command. Nothing will do it for you.

```
.specify/
  templates/[overrides/]  artefact skeletons; put your changes in overrides/
  scripts/bash/           create-new-feature, check-prerequisites, setup-*, resolve-template
  memory/constitution.md  the governing document — you own this
  integrations/ workflows/ integration.json init-options.json
  feature.json            active feature pointer (git-ignored, not your branch)
.claude/skills/           ten speckit-* SKILL.md files (generated; do not edit)
specs/001-config-core/
  spec.md                 what and why — the artefact you maintain
  plan.md                 how — where technology first appears
  research.md data-model.md contracts/ quickstart.md
  tasks.md                ordered work, grouped by user story
  checklists/requirements.md   spec-quality gate, kept by specify/clarify
docs/
  v1-baseline/            descriptive record of the old system (brownfield)
  parity-matrix.md        baseline capability × spec/impl/test
```

### Is this feature done?

- [ ] Every mandatory spec section filled: user stories, edge cases, requirements, success criteria,
      assumptions
- [ ] Zero `[NEEDS CLARIFICATION]` markers remain
- [ ] Every user story has a priority and a stated Independent Test
- [ ] Every `FR-###` states exactly one testable thing; every `SC-###` is measurable and
      technology-agnostic
- [ ] No implementation detail leaked into `spec.md`
- [ ] `plan.md` Constitution Check passes, or violations are justified in Complexity Tracking with
      the simpler alternative explicitly rejected
- [ ] Every `FR-###` maps to at least one task; every task maps back to a requirement
- [ ] Every `FR-###` has at least one automated test
- [ ] Every edge case has a test or a documented reason it has none
- [ ] `just build` and `just test` are green
- [ ] The diff contains nothing no requirement asked for
- [ ] `git diff -- specs/` reviewed: no requirement quietly weakened, renumbered, or rewritten to
      match the code
- [ ] **`spec.md` reconciled with what actually shipped**
- [ ] Deviations from the baseline recorded, not silent
- [ ] `docs/parity-matrix.md` updated
- [ ] Spec `Status` header reflects reality

The ones that get skipped are the last four. They decide whether this repository has specifications
or archaeology.
