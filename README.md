# Grazhda — the plan and the groundwork

This directory is not the project yet. It is everything you need to build the
project yourself, deliberately, from an empty folder.

**What you are going to build.** Grazhda automates the work of a software
engineer, as three cooperating processes plus a shell installer:

| Component | Binary | Runtime | What it is for |
| :--- | :--- | :--- | :--- |
| CLI interface | `zgard` | Go | the engineer's hands — every interactive command |
| CLI daemon | `dukh` | Go, gRPC | local eyes and reflexes — watches, answers instantly |
| Long-term process | `molfar` | Java 25 / Spring Boot 4.1 | memory and judgement — plans, runs, records, proposes |
| Toolchain manager | `grazhda` | Bash | installs, upgrades and removes the other three |

**How you are going to build it.** Spec-driven development with
[GitHub spec-kit](https://github.com/github/spec-kit): every behaviour is
specified before it is written, an AI agent drafts the implementation from the
specification, and you accept or reject the result against that specification.
You own the specifications. The agent owns the typing.

---

## What is here

| File | What it is | When to read it |
| :--- | :--- | :--- |
| **`CLAUDE.md`** | The rules an AI agent working in this repository must follow. Auto-loaded by Claude Code every session, so it is deliberately short. | It reads itself; read it once so you know what it commits you to |
| **`docs/CONTEXT.md`** | What was decided before this repository existed and why, plus research findings worth not rediscovering. | Once, before Phase 0 |
| **`PLAN.md`** | The seventeen phases, from installing Go to a finished project. Every phase says what to do, why it matters, how to check it and what people get wrong. | Now, all of it, before starting |
| **`IDEAS.md`** | A critique of the automation idea and fifteen concrete changes worth making. Argues where the defensible value actually is. | Now — several points belong in the constitution, and it is much cheaper to include them than to retrofit |
| **`MYTHOLOGY.md`** | Where the names come from: the Hutsul homestead, the linked amulet, the house spirit, the weather-worker. Plus a lexicon for naming what does not exist yet. | Any time. Read it before you name a new component |
| **`guides/01-just.md`** | `just` and the Justfile, from zero, built up through this project's own build | Before Phase 6 |
| **`guides/02-go.md`** | Go from zero for someone who knows Java. The longest of the four, and the one you will lean on most | Before Phase 6, and keep it open through Phases 8–14 |
| **`guides/03-spec-kit.md`** | Spec-driven development and spec-kit, from zero: the method, the commands, how to write a specification that is actually useful, and — section 12 — how to run the whole loop inside Claude Code | Before Phase 3, and re-read section 7 before every spec you write |
| **`guides/04-spring-boot.md`** | Spring Boot from zero for someone who knows Java. Targets Spring Boot 4.1 specifically, including the modular auto-configuration trap that will otherwise cost you an afternoon | Before Phase 15 |

---

## Suggested order

**Day one — read.** `PLAN.md` end to end, then `IDEAS.md`. Do not install
anything yet. The point of the first day is to decide whether you agree with the
plan and the product definition, because everything after that is expensive to
change.

**Day two — Phase 0.** Write your own `docs/product.md`. Answer the five
questions in Phase 0 honestly, especially the second one: which ten chores do
you actually want back? That answer determines whether this project is worth six
weeks of your evenings.

**Then work the phases.** Read `guides/03-spec-kit.md` before Phase 3,
`guides/01-just.md` and `guides/02-go.md` before Phase 6, and
`guides/04-spring-boot.md` before Phase 15. Phase 8 is the teaching phase — go
slowly through it once and the remaining features will move three times faster.

---

## The three things most likely to go wrong

**You skip Phase 0 and start installing tools.** Tooling is comforting and
progress-shaped. Deciding what the product is, is neither. The plan front-loads
the uncomfortable part on purpose.

**You accept a specification you only skimmed.** Everything downstream inherits
its mistakes, and an AI agent will implement a vague requirement confidently
rather than asking you again. The half-hour spent reading a spec properly is the
highest-return half-hour in the whole project.

**You skip reconciliation.** After the code works there is a step where you go
back and correct the specification to describe what actually shipped. It feels
like paperwork. Skip it three times and the specifications become fiction, which
is worse than having none — because people trust them.

---

## A note on what this is not

This is not a plan for me to execute. It is written for you to work through by
hand, with an AI agent as the thing that types, not the thing that decides.
Every phase names the checks you run yourself, because the whole method rests on
a human being who can tell whether the result is right.
