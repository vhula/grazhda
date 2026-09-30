# Context — what was decided before this repository existed

Written 2026-09-03, at the end of the session that produced `PLAN.md` and the
guides. Read this once at the start of a session. It explains choices that look
arbitrary from the outside, and records findings that cost real effort to
establish so nobody has to establish them twice.

`CLAUDE.md` carries the rules. This file carries the reasoning behind them.

---

## 1. Where this comes from

`/mnt/e/proj/grazhda` is a working Go project: a multi-repository workspace
lifecycle toolkit. You describe workspaces in YAML, then use `zgard` to clone,
pull, inspect and operate on many repositories at once, `dukh` to monitor their
health in the background, and a `grazhda` shell script to install and upgrade
the toolchain. It has a declarative package manager, a gRPC contract between the
CLI and the daemon, and roughly 7,000 lines of production Go with a comparable
amount of tests.

**The new purpose is different.** Grazhda is being re-aimed at *automating the
work of a software engineer*. The workspace machinery is no longer the product;
it is the substrate. On top of it sits a new long-running process that accepts a
chore, plans it, runs it across many repositories, records everything, and hands
back a proposal a human approves.

That is a change of purpose, not a refactor — which is why the rebuild is
spec-first rather than an incremental port.

---

## 2. The decisions, and why

### Three processes, two languages

| Process | Language | Reasoning |
| :--- | :--- | :--- |
| `zgard`, the CLI | Go | invoked dozens of times an hour; JVM start-up is disqualifying. Single static binary, no runtime on the engineer's machine, first-class concurrency for fanning out across repositories |
| `dukh`, the daemon | Go | shares one domain library with the CLI; the same argument applies |
| `molfar`, the long-term process | Java 25 / Spring Boot 4.1 | must run for weeks; benefits from a mature scheduler, persistence, transaction and observability stack, and pays start-up cost once |
| `grazhda`, the installer | Bash | it installs the others; it cannot depend on them |

**Rejected:** Java for all three (unacceptable CLI latency); Go for all three
(would mean rebuilding scheduling, persistence and migration machinery by hand).

The honest cost of the Spring Boot choice: several hundred megabytes resident on
a laptop that is already running an IDE and a browser. Three mitigations were
planned rather than left to be discovered — let `dukh` start `molfar` on demand
and stop it when idle; keep the domain in a framework-free `core` package so a
native-image build stays viable; consider running it as a container.

### Scope: superset, not pivot

Everything the old project does stays and is reframed as the substrate.
`molfar` adds the automation on top. **Parity is milestone one; automation is
milestone two.** The parity obligation is tracked in `docs/parity-matrix.md`, and
it is what stops the rebuild quietly losing features that took years to get right.

### Spec-kit, for real

The real `specify` CLI (`specify init . --integration claude`), not a hand-rolled
imitation of the method. The reason to use the actual tool is that its templates
and gate scripts encode discipline the owner would otherwise have to remember.

### Ownership

The owner writes the specifications and accepts the results. An agent types.
This is stated in `CLAUDE.md` as a rule because it is the thing that erodes
first under time pressure, and because the entire value of the method collapses
without it.

---

## 3. Naming

The names are Hutsul — the highlanders of the Ukrainian Carpathians — and each
one was chosen because its meaning describes the component.

- **grazhda** (ґражда): the fortified Carpathian homestead, a closed square of
  dwelling and outbuildings around one courtyard, with blank outer walls and a
  single gate. A boundary drawn around everything that belongs together. It is
  why packages install into the project's own directory rather than the host
  system.
- **zgard** (згарда): the neck ornament of many small cast crosses strung on one
  cord, worn as adornment and as an amulet. Many repositories on one thread — and
  a protective object, which is why the CLI's most-used features are `--dry-run`,
  the confirmation prompt, and the refusal to purge without an explicit target.
- **dukh** (дух): spirit, breath. The household spirit that watches the homestead
  and makes disorder known. The Unix word *daemon* comes from the Greek δαίμων,
  a benevolent attendant spirit — both traditions land on the same idea.
- **molfar** (мольфар): the Hutsul weather-worker and seer; knowledge accumulated
  over years, passed by apprenticeship, used for or against the community. The
  moral distinction in the folklore is exactly why the design insists the system
  *proposes* and a human *approves*.

`MYTHOLOGY.md` has the full account, a lexicon for naming future components
(**trembita** for the event bus, **vatra** for the run journal, **strunka** for
the work queue, **koliba** for an ephemeral sandbox), and the discipline that
keeps this usable: mythological names go on processes and subsystems, never on
commands or flags, and every one gets a one-line gloss where it first appears.

---

## 4. Findings worth not rediscovering

These cost research time. They are current as of September 2026 and should be
re-verified if a lot of time has passed.

### Spring Boot 4.1

- Released 10 June 2026. Spring Framework 7.0.8, Spring Security 7.1.0,
  Hibernate 7.1, Tomcat 11, Jakarta Persistence 3.2. Java baseline is still
  JDK 17, so Java 25 is comfortably in range.
- **Auto-configuration is modular, and the failure mode is silence.** In Boot 3,
  putting a library on the classpath was enough to get it wired. In Boot 4 each
  auto-configuration lives in its own module. Adding raw `flyway-core` gives you
  the library with **no Spring wiring and no error** — no migrations, no log
  line, nothing. Depend on `spring-boot-starter-flyway`. Generalised: *the
  starter is the contract; a raw library gives you silence.*
- **gRPC auto-configuration is new in 4.1** — server and client, Netty and
  Servlet HTTP/2 transports, `@GrpcAdvice` for centralised exception handling,
  automatic observations. This is `molfar` talking to `dukh` handed to you.
  Channel property is singular: `spring.grpc.client.channel.<name>.target`. The
  plural `channels…address` form was Spring gRPC 1.0 and is out of date.
- Jackson 3 is the default; Jackson 2 is deprecated. Derby support and
  layertools were removed.
- Also worth knowing: HTTP Service Clients, API versioning auto-configuration,
  SSRF mitigation through a configurable `InetAddressFilter`, `RestTestClient`,
  an OpenTelemetry starter, and automatic context propagation for `@Async`.

### spec-kit

- Commands are installed as skills at `.claude/skills/speckit-*/SKILL.md`, and
  Claude Code exposes each as a slash command named after the directory.
- **The separator is a hyphen, not a dot.** Upstream documentation writes
  `/speckit.specify`; the Claude integration uses `/speckit-specify`. The dotted
  form is the command's canonical identity; the separator is recorded per
  integration in `.specify/integration.json` as `invoke_separator`. Believe the
  disk.
- **The active feature is `.specify/feature.json`, not the git branch.** Current
  spec-kit does not key off branches; branch creation is an optional hook. Nearly
  all third-party writing about spec-kit still says otherwise.
- `/speckit-converge` is the only command that reads the codebase as input, and
  it is deliberately built so the spec is read-only intent: its output is always
  new *tasks*, never edits to requirements. That is the tool's own defence
  against an agent rewriting a specification to match its code.

### Claude Code

- Settings precedence, highest first: managed settings, `claude --settings`,
  `.claude/settings.local.json`, `.claude/settings.json`, `~/.claude/settings.json`.
  Commit the shared file with the allow-list this loop needs; keep personal
  exceptions in the local one.
- `permissions.defaultMode` values `auto` and `bypassPermissions` do not take
  effect from project or local settings — user or managed only. A deliberate
  guard rail; leave it in place.
- Custom commands and skills are the same mechanism: `.claude/commands/x.md` and
  `.claude/skills/x/SKILL.md` both produce `/x`.

`guides/03-spec-kit.md` §12 and `guides/04-spring-boot.md` carry all of this in
teaching form, with sources.

---

## 5. What an exploratory build already demonstrated

Before this plan was written, one full automated attempt at the rebuild was run
end to end and then deleted, deliberately, so the owner could do it by hand. The
code is gone; the findings are not.

**The approach works.** Constitution, eight specifications with roughly 370
numbered requirements, the shared Go domain and the daemon were all produced and
were green: `go build`, `go vet` and about 250 tests passing, per module and in
the workspace.

**The CLI was the gap.** Domain code and daemon were complete; `cli/` had no Go
source at all. That is exactly the drift `PLAN.md` now guards against by
requiring each feature phase to ship both the domain and the command surface,
and by giving the parity matrix separate columns for the two.

**Two real concurrency bugs in the old code** were found by the race detector
and would otherwise have been ported verbatim: a data race in the package
manager's spinner `Stop`, and one in the daemon server's shutdown path. Run
`go test -race` from the beginning.

**The old project's own `GAPS.md` is unusually honest** and lists five real
discrepancies between its documentation and its behaviour. Use it as the seed
for the capability inventory in Phase 2, and make a deliberate keep-or-close
decision on each.

**Details established by reading the old source,** worth having on hand when
specifying features 002 to 007: the four distinct no-match filter branches; the
exact skip and success message strings; the two different "not cloned" versus
"no upstream" placeholders in `ws diff`, where the old documentation and the old
code disagree and the code is right; glob taking precedence over regex in
`ws search`; and the fact that the daemon must resolve repository paths exactly
as `ws init` does or it reports present repositories as missing. `PLAN.md` phases
9 to 14 carry these.

---

## 6. Decisions deliberately deferred

- **Whether grazhda integrates with an issue tracker, CI and a code forge.**
  `IDEAS.md` §8 argues the highest-value early automations are the non-code ones
  — assembling a standup, triaging overnight CI failures, drafting a pull-request
  description — and those need integrations. Purely local is a legitimate choice
  but caps the ceiling. Settle it in Phase 0.
- **Which embedded database `molfar` uses.** H2 in file mode, SQLite, or a
  containerised Postgres. Trade-offs in `guides/04-spring-boot.md` part 7.
- **Whether `ws list` gains a `--parallel` flag** it never had in the old
  project. Decide, then make the spec and the binary agree.
- **How the planner is implemented.** The specification is written so a plan may
  come from a deterministic template or from a language model; the interface is
  specified in terms of what a plan must contain, not how it is derived.

---

## 7. The parts of `IDEAS.md` that belong in the constitution

`IDEAS.md` is a critique with fifteen suggestions. Four of them are the ones
worth defending, and they are cheaper to write into the constitution in Phase 4
than to retrofit into `molfar` in Phase 15:

1. **An oracle on every step.** A plan step with no way to check it worked is a
   step that cannot be trusted, and the proposal should say so out loud —
   "these four changes were verified; these two could not be".
2. **Repository content is untrusted input.** A README or a dependency changelog
   can contain instructions. Repo content may influence the *content* of a
   proposal, never the *policy* — scope, budget, approval requirement. Enforce it
   structurally, not with a warning.
3. **Observe-only first.** Ship `molfar` in shadow mode, planning and recording
   but executing nothing, for weeks. The proportion of plans that were right is
   the only honest answer to "is it ready to be given hands".
4. **An evaluation corpus.** Frozen repository states plus tasks plus known-good
   outcomes. Without one, every planner change is a matter of opinion.

---

## 8. How this repository got here

The `.git` history begins with a commit called *"Rewrite with speckit: start
from 0"*, which emptied the working tree of the old implementation. Everything
currently present — `PLAN.md`, `IDEAS.md`, `MYTHOLOGY.md`, `README.md`, the four
guides and this file — is planning material. No project source exists yet.

`PLAN.md` Phase 0 begins with `git init`; that has already happened, and the
`.gitignore` from the old project is still in place and will need replacing per
Appendix E.
