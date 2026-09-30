# STEPS — the executable form of `PLAN.md`

`PLAN.md` tells you *what each phase is for* and *what goes wrong if you skip
it*. This document tells you **which key to press**. It is the same seventeen
phases, broken into numbered steps small enough that none of them requires a
decision you have not already been told how to make, each with the command that
performs it, the technology it involves, and the reason it is there.

**`PLAN.md` is authoritative.** Where the two disagree, the plan wins and this
file is the thing with the bug. A handful of steps here are not in `PLAN.md` at
all — mechanical consequences of things it says elsewhere, or traps in this
specific working copy. Every one of those is marked **`[not in PLAN.md]`** so
you can tell my additions from the owner's plan.

---

## How to read a step

Every step has the same four parts:

| Part | What it is |
| :--- | :--- |
| **What** | The change to the world, in one sentence |
| **Why** | What this buys you, and the technology involved |
| **Do** | The command, exactly, to paste into a Linux shell |
| **Check** | How you know the step worked before moving on |

Two conventions in the command blocks:

- Every block assumes you are at the repository root, `/mnt/e/proj/grazhda-spec`,
  unless the block itself `cd`s somewhere. When in doubt, `cd /mnt/e/proj/grazhda-spec`.
- A command written as `heredoc` — `cat > file <<'EOF' … EOF` — writes a file
  from the terminal. The quotes around `'EOF'` matter: they stop the shell from
  expanding `$VARIABLES` and `` `backticks` `` inside the text, so what lands in
  the file is what you see on screen. Drop the quotes only when you *want*
  substitution.

**Steps prefixed 🧠 are yours and cannot be delegated.** They are the ones where
you decide something. Everything else is typing, and typing is what the agent is
for.

---

## The session ritual

`PLAN.md` and `guides/03-spec-kit.md` §12.4 both say one phase per session with
`/clear` between phases. Concretely, every time you sit down:

```bash
cd /mnt/e/proj/grazhda-spec
git log --oneline | head -5        # what did I actually last produce?
cat docs/parity-matrix.md          # (once it exists) what is left?
git status                         # is the tree clean?
```

Then in Claude Code: `/clear`, name the phase you are on, and reference the files
you need with `@path` rather than hoping they get found. At the end of the
session, commit. An uncommitted working tree is the one state from which you
cannot tell what you did.

The reason for `/clear` rather than `/compact` at a phase boundary is not
tidiness. Everything in the context window is evidence about what you want, so a
specification drafted in a session full of YAML-decoder debugging will mention
YAML decoding. `/compact` preserves a summary — which is exactly the
contamination you are dropping. The artefacts on disk are the memory.

---

## Where this working copy actually stands

Before Step 0.1, the facts about *this* directory, because two of them are traps:

- `git init` has already happened. `PLAN.md` Phase 0 opens with it; you can skip
  that line.
- **Every planning file is untracked.** `PLAN.md`, `IDEAS.md`, `MYTHOLOGY.md`,
  `README.md`, `CLAUDE.md`, `docs/CONTEXT.md` and `guides/` have never been
  committed. The last commit, `e0c126b`, *emptied* the tree of the old
  implementation.
- **The `.gitignore` is the old project's** and contains `.claude/` and
  `*.tar.gz`. Both are wrong for what you are about to do — the first would hide
  the spec-kit skills you need committed, the second is currently hiding a
  `_to_delete/` directory full of tarballs. Phase 3 deals with both.

---

# Phase 0 — Decide what you are actually building

**Goal:** one page of prose, `docs/product.md`, that a stranger could read and
correctly predict what the software does.

**Effort:** half a day, almost all of it thinking. Nothing is installed here.

---

### Step 0.1 — Confirm where you are

**What.** Establish that the shell is in the right directory and the git
repository is the one you think it is.

**Why.** Every command after this is relative to this directory. Two minutes of
confirmation now prevents an afternoon of "why did that land in the wrong repo".
`git rev-parse --show-toplevel` prints the root of the repository containing the
current directory, which is the unambiguous answer to "where am I".

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
git rev-parse --show-toplevel
git log --oneline -3
git status --short
```

**Check.** The first command prints `/mnt/e/proj/grazhda-spec`; the log shows
`e0c126b Rewrite with speckit: start from 0` at the top; the status lists the
planning files as untracked (`??`).

---

### Step 0.2 — 🧠 Read the planning material, in order

**What.** Read four documents before touching anything.

**Why.** This is the highest-return half-day in the project and there is no
shortcut. `PLAN.md` is the plan you are executing; `IDEAS.md` argues about
whether it is worth executing; `docs/CONTEXT.md` records decisions that look
arbitrary without their reasoning; `MYTHOLOGY.md` is what you consult before
naming anything. Reading them in this order means each one lands with the
context of the last.

**Do.**

```bash
less PLAN.md          # all of it, today
less IDEAS.md         # all of it, today
less docs/CONTEXT.md  # once
less MYTHOLOGY.md     # skim; read properly before you name something new
```

**Check.** You can state, without looking: which four components exist, which
language each is in, and why the CLI is not Java.

---

### Step 0.3 — Create `docs/product.md` from the skeleton

**What.** Write the file with its five questions in place and the answers blank.

**Why.** Separating "create the structure" from "answer the questions" is
deliberate. The structure is mechanical and takes ten seconds; the answers are
the work. If you try to do both at once you will write the structure and stop.
The five questions are ordered so each constrains the next — name whose work it
is and the chores narrow; name the chores and the output follows.

**Do.**

```bash
mkdir -p docs
cat > docs/product.md <<'EOF'
# Grazhda — what it is for

Status: DRAFT — answers below are placeholders until Step 0.4 replaces them.

## 1. Whose work

<!-- One engineer? A team lead? Someone maintaining thirty services?
     "A software engineer" is too broad to build for. Name the person,
     the number of repositories, and whose credentials it acts with. -->

## 2. Which chores

<!-- Ten real tasks you did in the last month and resented, specific enough
     to be planned, budgeted and checked. See PLAN.md Appendix D.1 for the
     difference between a chore and a category. -->

1.
2.
3.
4.
5.
6.
7.
8.
9.
10.

## 3. What the system produces

<!-- A pull request? A report? A fixed working tree? This shapes every
     interface in the project, so decide it before Phase 4. -->

## 4. What it must never do

<!-- These become constitutional principles in Phase 4. Push without asking?
     Touch a repository outside scope? Run unbounded? Leak a secret? -->

1.
2.
3.
4.

## 5. How I will know it worked

<!-- Not "it feels useful" — something countable, with an SC-### identifier,
     so Phase 4 and every later spec can reference it. -->

- **SC-A**:
- **SC-B**:
- **SC-C**:
EOF
```

**Check.** `cat docs/product.md` shows the five headings.

---

### Step 0.4 — 🧠 Answer the five questions

**What.** Replace every placeholder comment with prose. Open the file in your
editor; this is not a command-line step.

**Why.** Question 2 is where the project lives or dies. `PLAN.md` Appendix D.1
gives the worked example and, more usefully, the failure beside it:

```
Bad:  Dependency management. Code quality. Repository maintenance.
Good: Bump the logging library across every repository tagged `backend`
      and confirm each still builds.
```

The bad answer names categories, and a category cannot be planned, budgeted or
checked — nothing tells you whether "code quality" succeeded. The good one names
a **scope**, a **transformation** and an **oracle**. Specify the first kind and
the agent will produce something plausible and useless.

Two decisions `docs/CONTEXT.md` §6 says to settle here rather than defer:
whether grazhda integrates with an issue tracker, CI and a code forge (`IDEAS.md`
§8 argues the highest-value early chores are the non-code ones, and those need
integrations); and whether the ten chores in question 2 are weighted toward
reading or toward changing code.

**Do.**

```bash
${EDITOR:-vi} docs/product.md
```

**Check.** Read `PLAN.md`'s own test back to yourself: you can describe the
product in two sentences **without using the word "AI"**. If you cannot, the
definition is a mood, not a product. Narrow it until it is uncomfortable.

---

### Step 0.5 — Commit the planning material and the product definition

**What.** The first commit of this repository's new life.

**Why.** Everything currently in the tree is untracked. Committing it now means
Phase 3's `specify init` — which writes into an existing directory — has
something to diff against if it overwrites something you wanted. `PLAN.md`
Phase 3's pitfall is exactly this: "Running `specify init` inside a non-empty
directory without understanding what it will overwrite. Commit first, always."

**Do.**

```bash
git add PLAN.md IDEAS.md MYTHOLOGY.md README.md CLAUDE.md STEPS.md docs/ guides/
git commit -m "docs: plan, guides and product definition

The planning material produced before Phase 0, plus docs/product.md
answering the five Phase 0 questions."
git log --oneline -2
```

**Check.** `git status --short` no longer lists those files.

---

# Phase 1 — Prepare the machine

**Goal:** eight tools installed, each printing a version.

**Effort:** 1–2 hours, most of it downloads.

**Why the whole phase exists.** Half the frustration in a polyglot project is
discovering a missing toolchain three layers into a build, where the error
message is about something else entirely. Get them all in, confirm each prints a
version, then never think about it again.

---

### Step 1.1 — Verify `git`

**What.** Confirm git is present and recent enough.

**Why.** Git is both your source control and the thing `zgard` orchestrates —
every workspace operation shells out to it. Anything from the last few years is
fine.

**Do.**

```bash
git --version
```

**Check.** Prints a version. If not: `sudo apt-get install -y git`.

---

### Step 1.2 — Install Go 1.26

**What.** Install the Go toolchain from upstream tarball into `/usr/local/go`.

**Why.** Go is the language for `zgard` and `dukh`. It compiles to a single
static binary with no runtime on the engineer's machine, starts in
milliseconds — which is the whole reason the CLI is not Java — and has
first-class concurrency for fanning out across repositories.

**Do not use the distribution package.** `apt`'s Go typically lags by two or more
minor versions, and a Go older than 1.26 will reject the `go.work` file you write
in Phase 6 with an error that does not mention the version. The upstream tarball
is the supported installation route: unpack it at `/usr/local/go` and put
`/usr/local/go/bin` on `PATH`.

**Do.**

```bash
curl -fsSLO https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
rm go1.26.1.linux-amd64.tar.gz
```

**Check.**

```bash
/usr/local/go/bin/go version    # go version go1.26.1 linux/amd64
```

---

### Step 1.3 — Put Go and its installed binaries on `PATH`

**What.** Add `/usr/local/go/bin` and `$HOME/go/bin` to `PATH`, permanently.

**Why.** Two directories, two reasons. `/usr/local/go/bin` holds `go` itself.
`$HOME/go/bin` is where `go install` places compiled tools — and this is the one
that causes `PLAN.md`'s named pitfall. `protoc` finds its code-generator plugins
by looking for executables named `protoc-gen-*` **on `PATH`**. If `$HOME/go/bin`
is missing from it, `protoc-gen-go` is installed and invisible, and the error is
a confusing "program not found or is not executable" that says nothing about
`PATH`.

**Do.**

```bash
grep -q '/usr/local/go/bin' ~/.bashrc || \
  echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
```

The `grep -q … ||` guard makes this idempotent — run it twice and `~/.bashrc`
gains one line, not two. That habit matters more than it looks: Phase 14's
installer has the same requirement written into its specification.

**Check.**

```bash
go version
echo "$PATH" | tr ':' '\n' | grep -E 'go/bin'
```

Both `/usr/local/go/bin` and `<your home>/go/bin` appear.

---

### Step 1.4 — Install `just`

**What.** Install the `just` command runner into `~/.local/bin`.

**Why.** `just` is the single build entry point for a repository that contains
Go, Java and Bash. Without it, "how do I build this?" has three different
answers depending on which subdirectory you are in, and CI reimplements all three
with slightly different flags. With it, `just build` means the same thing to you
and to CI, which is what makes a CI failure reproducible locally.

It is `make` without the parts of `make` that exist for C: no implicit rules, no
file-timestamp dependency checking (recipes always run), and no tab-versus-spaces
trap. `guides/01-just.md` teaches it properly.

**Do.**

```bash
mkdir -p ~/.local/bin
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh \
  | bash -s -- --to ~/.local/bin
grep -q '.local/bin' ~/.bashrc || echo 'export PATH=$HOME/.local/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
```

**Check.**

```bash
just --version
```

---

### Step 1.5 — Install `protoc` and the two Go plugins

**What.** Install the protocol-buffer compiler and the two generators that turn
a `.proto` file into Go.

**Why.** `contracts/dukh.proto` (Phase 7) is the only coupling between the CLI,
the daemon and later `molfar`. `protoc` compiles it; `protoc-gen-go` emits the
message structs; `protoc-gen-go-grpc` emits the client and server interfaces.
They are separate binaries because `protoc` is a plugin host — it parses the
`.proto` and hands a descriptor to whichever `protoc-gen-<lang>` you name on the
command line.

**Pin the plugin versions.** `PLAN.md` Appendix B lists two separate failure
modes caused by `@latest`: a plugin that drifts out of sync with the `grpc`
module version, and a `daemon/proto` diff check in CI that fails only sometimes,
because your machine and CI installed "latest" on different days. `@latest` is
shown in the plan's Phase 1 snippet; pinning is what the troubleshooting table
tells you to do once it bites. Do it now.

**Do.**

```bash
sudo apt-get update
sudo apt-get install -y protobuf-compiler shellcheck

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
```

*(Check the current release of each before pinning — `go list -m -versions
google.golang.org/protobuf` lists them. The point is that a specific version is
written down somewhere, not that these two numbers are eternal.)*

**Check.**

```bash
protoc --version              # libprotoc 3.x or later
which protoc-gen-go           # ~/go/bin/protoc-gen-go
which protoc-gen-go-grpc      # ~/go/bin/protoc-gen-go-grpc
shellcheck --version
```

If either `which` prints nothing, Step 1.3 did not take effect in this shell.

---

### Step 1.6 — Install `uv`

**What.** Install Astral's `uv`, a Python package and tool installer.

**Why.** Spec-kit ships as a Python CLI called `specify-cli`. You are not writing
Python; `uv tool install` is simply the supported way to get that one command
onto your `PATH` in its own isolated environment, without touching your system
Python. It is the recommended installer in spec-kit's own documentation.

**Do.**

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
source ~/.bashrc
```

**Check.**

```bash
uv --version
```

---

### Step 1.7 — Install Java 25 and Maven via SDKMAN

**What.** Install SDKMAN, then a Java 25 JDK and Maven through it.

**Why.** `molfar` is Java 25 on Spring Boot 4.1 — the component that must run for
weeks and therefore benefits from a mature scheduler, persistence, transaction
and observability stack, paying its start-up cost once. SDKMAN is the least
painful route because it manages several JDKs side by side and switches between
them per shell, which matters the first time you need to reproduce a build on a
different version.

Spring Boot 4.1's baseline is JDK 17, so Java 25 is comfortably in range; the
reason for 25 is that it is the current long-term-support release and adopting it
now avoids a migration before `molfar` has shipped a feature.

**Do.**

```bash
curl -s "https://get.sdkman.io" | bash
source ~/.sdkman/bin/sdkman-init.sh
sdk install java 25-tem
sdk install maven
```

`25-tem` is Eclipse Temurin, the build most CI images use — matching it locally
removes a class of "works on my machine" difference.

**Check.**

```bash
java -version    # openjdk version "25"
mvn -v           # must report Java 25, not just any Java
```

**The trap here** is in `PLAN.md` Appendix B: `mvn -v` reporting an older Java
than 25 means the SDKMAN environment was not sourced in the shell that invoked
`mvn` — a fresh terminal, or CI. Confirm `mvn -v` before every Java build until
`~/.bashrc` sources `sdkman-init.sh` reliably.

---

### Step 1.8 — Verify the whole toolchain in one command

**What.** Run all eight verification commands together.

**Why.** `PLAN.md`'s Phase 1 check is "run every command in the Verify with
column; all eight print a version". Doing it as one block means you see all eight
answers on one screen and can spot the one that silently reports the wrong
version — which is the failure mode that actually happens, rather than a command
being missing outright.

**Do.**

```bash
for c in "git --version" "go version" "just --version" "protoc --version" \
         "uv --version" "java -version" "mvn -v" "shellcheck --version"; do
  printf '\n--- %s\n' "$c"; eval "$c" 2>&1 | head -3
done
```

**Check.** Eight blocks, eight versions, no "command not found". Go is 1.26+,
Java is 25, and `mvn -v` agrees about Java.

---

### Step 1.9 — Record the versions **`[not in PLAN.md]`**

**What.** Write down what you just installed.

**Why.** `PLAN.md` Appendix B twice tells you to pin exact tool versions "in one
place both environments read from" — once for the protoc plugins, once for CI
generally. This file is that place. It costs a minute now and answers "what
changed?" in three months when a generated file starts differing between your
machine and CI.

**Do.**

```bash
{
  echo "# Toolchain versions"
  echo
  echo "Recorded $(date +%Y-%m-%d). CI pins these; see .github/workflows/ci.yml."
  echo
  {
    git --version; go version; just --version; protoc --version
    uv --version; java -version 2>&1 | head -1
    mvn -v | head -1; shellcheck --version | sed -n 2p
    echo "protoc-gen-go      v1.36.6"
    echo "protoc-gen-go-grpc v1.5.1"
  } | sed 's/^/    /'
} > docs/toolchain.md
git add docs/toolchain.md && git commit -m "docs: record pinned toolchain versions"
```

The inner `sed 's/^/    /'` indents the version lines by four spaces, which is
Markdown's indented-code-block form — a fenced block would need backticks inside
a shell string, which is exactly the kind of quoting fight not worth having.

**Check.** `cat docs/toolchain.md` reads back the eight versions.

---

# Phase 2 — Baseline the old project

**Goal:** `docs/v1-baseline/` holding the old project's documentation, plus
`docs/capability-inventory.md` and `docs/parity-matrix.md`.

**Effort:** half a day to a day.

**Why the whole phase exists.** This is a brownfield rebuild. There is a working
implementation at `/mnt/e/proj/grazhda` whose behaviour is the accumulated result
of many decisions, some deliberate and some accidental. You cannot specify a
system you have not inventoried, and you cannot tell when the rebuild is finished
without a checklist. The parity matrix is your definition of done.

**The rule that governs this whole phase:** the old project is read-only
reference material. Read it freely; never modify a file in it; **never copy its
source into this repository**. Porting it line by line produces specifications
written after the fact, which is precisely the failure this rebuild exists to
avoid. Documentation and contracts are copied; `.go` files are not.

---

### Step 2.1 — Create the baseline directory

**What.** One directory for everything copied from v1.

**Why.** Keeping it under one path makes the boundary visible: anything under
`docs/v1-baseline/` is history to be honest about, not code to be reused.

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
mkdir -p docs/v1-baseline
```

**Check.** `ls -d docs/v1-baseline`.

---

### Step 2.2 — Copy the old documentation and contracts

**What.** Six guides, the config template, the proto contract and the README.

**Why.** These four kinds of file are the *stated intent* of the old system.
`CLI.md` is the command surface you owe parity to. `CONFIG.md` and
`config.template.yaml` are the configuration schema feature 001 must cover.
`GAPS.md` is the old project's own honest list of places where its documentation
and its behaviour disagree — the seed for the inventory's honesty column.
`dukh.proto` is the contract Phase 7 rewrites deliberately rather than copies.

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
cp /mnt/e/proj/grazhda/guides/{CLI,CONFIG,GAPS,QUICK-START,DEVELOPMENT,feature-ideas}.md \
   docs/v1-baseline/
cp /mnt/e/proj/grazhda/config.template.yaml docs/v1-baseline/
cp /mnt/e/proj/grazhda/proto/dukh.proto      docs/v1-baseline/
cp /mnt/e/proj/grazhda/README.md             docs/v1-baseline/README-v1.md
ls -1 docs/v1-baseline/
```

If a path does not exist, find its equivalent rather than skipping it:

```bash
ls /mnt/e/proj/grazhda/guides/ /mnt/e/proj/grazhda/*.md /mnt/e/proj/grazhda/proto/
```

**Check.** `ls -1 docs/v1-baseline/` lists nine files and **no `.go` files**:

```bash
find docs/v1-baseline -name '*.go' | wc -l    # must print 0
```

---

### Step 2.3 — Build the old binaries

**What.** Build v1 in its own directory so you can run its commands.

**Why.** Documentation is what a system was supposed to do; the binary is what it
does. The difference between them is exactly what `GAPS.md` is about, and the
only way to find the gaps `GAPS.md` does not list is to run the thing.

**Do.**

```bash
cd /mnt/e/proj/grazhda && just build
ls -l bin/
```

**Check.** `bin/zgard` and `bin/dukh` exist and are executable. If the build
fails, it is a v1 problem — do not fix it here; work from the documentation and
note that you could not run it.

---

### Step 2.4 — Capture every command's help text

**What.** Walk the whole command tree and save the output.

**Why.** You are going to write one inventory row per capability, and the help
text is the fastest complete enumeration of flags. Capturing it to a file rather
than reading it in the terminal means you can grep it while writing the inventory
and diff it against `zgard --help` at Phase 16 close-out.

**Do.**

```bash
cd /mnt/e/proj/grazhda
{
  for c in "" "ws" "ws init" "ws pull" "ws purge" "ws exec" "ws stash" \
           "ws checkout" "ws search" "ws diff" "ws stats" "ws list" "ws status" \
           "pkg" "pkg install" "pkg purge" "pkg register" "pkg unregister" \
           "config" "config get" "config edit"; do
    printf '\n===== zgard %s --help\n' "$c"
    ./bin/zgard $c --help 2>&1
  done
  printf '\n===== dukh --help\n'; ./bin/dukh --help 2>&1
} > /mnt/e/proj/grazhda-spec/docs/v1-baseline/help-output.txt

cd /mnt/e/proj/grazhda-spec
grep -c '=====' docs/v1-baseline/help-output.txt
```

Adjust the command list to what v1 actually has — the list above is derived from
`PLAN.md` phases 9–14 and may not match exactly. A command that errors with
"unknown command" is itself a finding worth an inventory row.

**Check.** `less docs/v1-baseline/help-output.txt` shows one block per command,
each with its flags.

---

### Step 2.5 — 🧠 Read `GAPS.md` and decide on each entry

**What.** Read the old project's own list of discrepancies and make a
keep-or-close decision on every one.

**Why.** `docs/CONTEXT.md` §5 says `GAPS.md` is "unusually honest" and lists five
real discrepancies between documentation and behaviour. Each one is a decision
waiting to be made, and the decision has exactly two legitimate answers:

- **Close** — the rebuild fixes it, and the fix goes into a specification.
- **Carry** — the rebuild reproduces the behaviour deliberately, with a spec that
  says why (usually because scripts depend on it).

What you must not have is a third state, where a behaviour survives because
nobody looked.

**Do.**

```bash
less docs/v1-baseline/GAPS.md
```

**Check.** You can name, for each gap, which of Close or Carry it is and which
feature number (001–007) will own it.

---

### Step 2.6 — Create `docs/capability-inventory.md`

**What.** One row per observable behaviour of v1, with an honesty column.

**Why.** The indispensable column is **intended or accidental**. Without it you
will transcribe accidents into specifications and defend them for a year. Fill it
by asking of each row: *designing this today, would I do it this way?*
`PLAN.md` Appendix D.2 has a filled example; the rows below are its examples,
kept as a calibration reference for the ones you add.

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
cat > docs/capability-inventory.md <<'EOF'
# Capability Inventory — grazhda v1

One row per observable behaviour of the baseline (`docs/v1-baseline/`). The last
two columns are the point: *intended* behaviour is a candidate requirement,
*accidental* behaviour is a decision waiting to be made. **Close** = the rebuild
fixes it. **Carry** = the rebuild reproduces it deliberately, and says so in a
spec.

Source material: `docs/v1-baseline/CLI.md`, `CONFIG.md`, `GAPS.md`,
`help-output.txt`, and the binaries themselves.

| Capability | Flags | Behaviour | Intended or accidental? | Decision |
| :--- | :--- | :--- | :--- | :--- |
| `ws init` | `-n`, `--all`, `-p`, `-r`, `-t`, `--dry-run`, `--parallel` | Clones every repository in the resolved target; skips ones already on disk; continues past a failure and reports all failures at the end. | Intended — the continue-past-failure contract is stated in `CLI.md` and relied on. | Specify as-is (003). |
| Default-workspace fallback | none | With no targeting flag, resolves the workspace marked `default` and prints a notice to **stderr** naming the path. | Intended — the notice exists so the target is never a surprise. | Specify as-is (002). |
| `-r/--repo-name` matching | `-r` | Case-sensitive **substring** match against the full config name, so `-r cool` matches `ORG/PACK/my-cool-backend-lol`; multiple matches warn, zero matches error. | Intended, but under-documented — the substring semantics surprise people. | Carry; state the rule explicitly with that example (002). |
| `ws status` targeting | `-p`, `-r`, `-t` | Flags inherited from the `ws` parent are accepted and then silently ignored; only `-n` filters the daemon snapshot. | **Accidental** — inherited flags nobody wired up. | Close: 006 applies the same filters as every other `ws` command. |
| `ws search` mode selection | `--glob`, `--regex` | Both may be supplied together; glob wins because the regex path is skipped when `--glob` is set. Nothing rejects the ambiguity. | **Accidental** — an ordering artefact, not a decision. | Carry: 004 specifies the precedence explicitly so existing scripts keep working. |
| `workspace.structure` values | n/a | Any value other than `list` is treated as `tree`; validation does not reject unknown values, so `treee` silently works. | **Accidental** — a fall-through default. | Close: 001 rejects unknown values at validation. |
| `zgard.*` and `general.*` config | n/a | Parsed, validated and queryable through `config get`; no workspace or package operation reads them. | **Accidental** in origin, defensible in effect. | Carry: 001 documents them as inspection metadata in its Assumptions. |
| *(add one row per command in `help-output.txt`)* | | | | |

## `GAPS.md` decisions

Every entry in `docs/v1-baseline/GAPS.md`, with its resolution, so the two
documents cannot drift apart.

| Gap (from `GAPS.md`) | Close or Carry | Owning feature | Reason |
| :--- | :--- | :--- | :--- |
| | | | |
EOF
${EDITOR:-vi} docs/capability-inventory.md
```

**Check.** Every command block in `docs/v1-baseline/help-output.txt` has at
least one row, and no row's honesty column is blank.

---

### Step 2.7 — Create `docs/parity-matrix.md`

**What.** The same capability list, with four empty status columns.

**Why.** This is your definition of done, and the four columns exist because one
"done" column hides the failure you are guarding against. A capability can be
implemented and untested, tested and unspecified, or specified and quietly never
built — three problems with three different fixes.

The fourth column, **CLI impl**, is the one people leave out, and it is the one
that matters most here: `docs/CONTEXT.md` §5 records that a full exploratory
build of this project produced a complete domain and daemon and **no CLI source
at all**. A matrix without that column would have reported those features done.

**Do.**

```bash
cat > docs/parity-matrix.md <<'EOF'
# Parity Matrix

Parity with the pre-specification implementation (`docs/v1-baseline/`) is a
standing obligation until every row reads ✅ (constitution § Governance).

Legend — **Spec**: a specification covers the behaviour · **Domain impl**: the
logic exists in `libs/` · **CLI impl**: a command exposes it · **Test**:
automated tests cover it.
**✅** done · **🟡** partial · **⬜** not started · **➖** deliberately declined.

Last updated: <!-- date -->

## `zgard ws` — workspace operations

| Baseline capability | Feature | Spec | Domain impl | CLI impl | Test |
| :--- | :--- | :---: | :---: | :---: | :---: |
| `ws init` (clone, skip existing, continue on failure) | 003 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws pull` (rebase, skip uncloned) | 003 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws purge` (explicit target, confirm, dry-run) | 003 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws exec <cmd>` (capture + indent output) | 004 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws stash` / `ws checkout` | 004 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws search` (content / glob / regex) | 004 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws diff` (uncommitted / ahead / behind table) | 004 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws stats` (last commit / 30d / contributors) | 004 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws list` (tree view, clone status) | 004 | ⬜ | ⬜ | ⬜ | ⬜ |
| `ws status` (health from the daemon) | 006 | ⬜ | ⬜ | ⬜ | ⬜ |

## Shared targeting and flags

| Baseline capability | Feature | Spec | Domain impl | CLI impl | Test |
| :--- | :--- | :---: | :---: | :---: | :---: |
| `-n/--name`, `--all`, default fallback + stderr notice | 002 | ⬜ | ⬜ | ⬜ | ⬜ |
| `-p/--project-name` exact match | 002 | ⬜ | ⬜ | ⬜ | ⬜ |
| `-r/--repo-name` substring match | 002 | ⬜ | ⬜ | ⬜ | ⬜ |
| `-t/--tag`, OR logic, project→repo inheritance | 002 | ⬜ | ⬜ | ⬜ | ⬜ |
| Cross-flag constraints (`-r` requires `-p`, etc.) | 002 | ⬜ | ⬜ | ⬜ | ⬜ |
| The four no-match branches, four distinct messages | 002 | ⬜ | ⬜ | ⬜ | ⬜ |
| Destination resolution (`tree`/`list`, `local_dir_name`, `~`) | 002 | ⬜ | ⬜ | ⬜ | ⬜ |
| Summary line, stderr failures, exit code 0/1 | 003 | ⬜ | ⬜ | ⬜ | ⬜ |

## `zgard config` and `zgard pkg`

| Baseline capability | Feature | Spec | Domain impl | CLI impl | Test |
| :--- | :--- | :---: | :---: | :---: | :---: |
| Config load, validate, query | 001 | ⬜ | ⬜ | ⬜ | ⬜ |
| `pkg install` / `purge` (lifecycle, env blocks) | 005 | ⬜ | ⬜ | ⬜ | ⬜ |
| `pkg register` / `unregister` | 005 | ⬜ | ⬜ | ⬜ | ⬜ |
| Two-layer registry merge semantics | 005 | ⬜ | ⬜ | ⬜ | ⬜ |
| Deterministic topological install order | 005 | ⬜ | ⬜ | ⬜ | ⬜ |

## `dukh` and `grazhda`

| Baseline capability | Feature | Spec | Domain impl | CLI impl | Test |
| :--- | :--- | :---: | :---: | :---: | :---: |
| `dukh start` / `stop` / `status` / `scan` | 006 | ⬜ | ⬜ | ⬜ | ⬜ |
| Monitoring loop, period re-read each tick | 006 | ⬜ | ⬜ | ⬜ | ⬜ |
| gRPC contract (status / scan / stop) | 006 | ⬜ | ⬜ | ⬜ | ⬜ |
| `grazhda-install.sh` (idempotent shell wiring) | 007 | ⬜ | ⬜ | ⬜ | ⬜ |
| `grazhda upgrade` / `uninstall` / `purge` | 007 | ⬜ | ⬜ | ⬜ | ⬜ |
EOF
```

**Check.** Every command in `docs/v1-baseline/CLI.md` appears as a row. A matrix
with no 🟡 rows once work starts usually means somebody is rounding up.

---

### Step 2.8 — Commit the baseline

**Do.**

```bash
git add docs/
git commit -m "docs: baseline v1, capability inventory and parity matrix

Copies v1 documentation and contracts into docs/v1-baseline/, inventories
every observable behaviour with an intended-or-accidental judgement, and
opens the parity matrix that defines done."
```

**Check.** `git status --short` clean; `git show --stat HEAD` lists no `.go`
files.

---

# Phase 3 — Initialise spec-kit

**Goal:** `.specify/` and the Claude integration exist, are committed, and the
loop's guard rails (git ignores, permissions, pre-commit hook) are in place.

**Effort:** 1 hour.

**Why the whole phase exists.** Everything after this point runs through
spec-kit's templates and scripts. Installing it early means every later artefact
is created the standard way rather than hand-rolled and slightly wrong. Its
templates and gate scripts encode discipline you would otherwise have to
remember on your own — which is the entire reason to use the real tool rather
than imitate the method by hand.

---

### Step 3.1 — Replace the `.gitignore` **`[not in PLAN.md's Phase 3]`**

**What.** Swap the old project's `.gitignore` for the one in `PLAN.md`
Appendix E.

**Why — and this is the trap in this working copy.** The current `.gitignore`
was inherited from v1 and contains two lines that will actively break the next
few steps:

- **`.claude/`** — spec-kit's Claude integration writes ten `SKILL.md` files
  under `.claude/skills/`, and Step 3.6 writes `.claude/settings.json`. Both are
  meant to be committed so a fresh clone gets the same loop and the same
  permission allow-list. With `.claude/` ignored they would silently never be
  staged, and you would discover it weeks later on another machine.
- **`*.tar.gz`** — currently hiding a `_to_delete/` directory of archives. Remove
  the pattern and they become visible to `git add -A`, which is how a hundred
  megabytes of tarballs end up in a repository's history permanently.

Appendix E explains each entry of the replacement. Two things that look like
build output and must **not** be ignored: the per-module `go.sum` files, which
are the actual dependency locks, and `daemon/proto/`, the generated protobuf Go,
committed so that a fresh clone builds before `protoc` is installed.

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
rm -rf _to_delete            # or: mv _to_delete ~/grazhda-v1-archives
cat > .gitignore <<'EOF'
/bin/
service/target/
go.work.sum
.specify/feature.json
.claude/settings.local.json
*.bak
.idea/
.vscode/
EOF
git add .gitignore && git commit -m "chore: replace v1 .gitignore per PLAN.md Appendix E

Drops .claude/ so the spec-kit skills and the shared permission allow-list
are committed, and *.tar.gz so nothing large is hidden from review."
```

**Check.**

```bash
git check-ignore -v .claude/settings.json    # prints nothing → not ignored. Good.
git check-ignore -v bin/zgard                # prints the /bin/ rule. Good.
```

---

### Step 3.2 — Install the spec-kit CLI

**What.** Install `specify-cli` as an isolated `uv` tool.

**Why.** `specify` is the scaffolding command: it writes `.specify/` (templates,
bash helper scripts, the constitution slot, state files) and the agent-facing
integration layer. Pin it from git at a tag on a project you intend to keep, so a
future `specify init` or refresh does not arrive with different templates than
the ones your existing specs were written against.

**Do.**

```bash
uv tool install specify-cli --from git+https://github.com/github/spec-kit.git
specify version
specify --help
```

To pin (recommended once you know which release you want — keep the `v` prefix):

```bash
# uv tool install specify-cli --from git+https://github.com/github/spec-kit.git@v1.0.2
```

**Check.** `specify --help` lists subcommands. Note that older tutorials mention
a bare `specify check`; current releases split that into `specify self check` and
`specify integration list`. Believe `--help` over any document.

---

### Step 3.3 — Confirm the tree is clean, then initialise

**What.** Run `specify init` into this existing directory.

**Why.** `init` writes into a non-empty directory, and `PLAN.md`'s pitfall for
this phase is running it without knowing what it will overwrite. A clean tree
before, and `git status` after, turns "what did it change?" into a one-command
answer.

`--integration claude` selects which coding agent the scaffolding is written for.
The `.specify/` directory is identical whichever you choose; only the agent-facing
layer differs. Pass it explicitly — a non-interactive session defaults to GitHub
Copilot. `--integration-options="--skills"` installs the commands as **agent
skills** (a directory per command holding a `SKILL.md`) rather than plain prompt
files; Claude Code exposes each skill directory as a slash command named after
it, so `speckit-specify/SKILL.md` becomes `/speckit-specify`.

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
git status --short                 # must be empty
specify init . --integration claude --integration-options="--skills"
git status --short                 # everything specify just wrote
```

**Check.** `git status` shows new files under `.specify/` and `.claude/skills/`
and — critically — **no modifications to files you wrote**. If `CLAUDE.md` or
`README.md` show as modified, read that diff before doing anything else.

---

### Step 3.4 — 🧠 Read what it created

**What.** Spend twenty minutes reading the scaffolding.

**Why.** Four files determine the shape of every artefact you produce for the
next six weeks. Reading them once now is much cheaper than inferring their rules
from output later.

- `.specify/templates/spec-template.md` — prioritised user stories with
  acceptance scenarios, edge cases, `FR-###` requirements, key entities, `SC-###`
  success criteria, assumptions. Its HTML comments are instructions to the agent
  and are worth reading in their own right.
- `.specify/templates/plan-template.md` — the Technical Context block, the
  **Constitution Check** gate, and the Complexity Tracking table.
- `.specify/scripts/bash/create-new-feature.sh` — how feature directories are
  numbered. It scans `specs/` for the highest numeric prefix and forces base 10,
  so `010` does not become 8; `--number` is a *preference*, and if that number is
  taken it warns and picks the next free one.
- `.specify/memory/constitution.md` — currently a placeholder skeleton. Phase 4
  fills it.

**Do.**

```bash
ls -1 .specify/templates/ .specify/scripts/bash/ .claude/skills/
cat .specify/integration.json
cat .specify/init-options.json
less .specify/templates/spec-template.md
less .specify/templates/plan-template.md
less .specify/scripts/bash/create-new-feature.sh
```

**Check.** You can answer three questions: what shape a `spec.md` has, where the
Constitution Check lives, and how a feature gets its number.

---

### Step 3.5 — Confirm the slash-command separator

**What.** Find out whether your installed skills are `/speckit-specify` or
`/speckit.specify`.

**Why.** This will confuse you exactly once. Upstream documentation writes the
dotted form; the Claude integration installs the hyphenated one. Neither is a
typo — the dotted form is the command's canonical *identity*, and the separator
is a per-integration property recorded in `.specify/integration.json` as
`invoke_separator`. Believe your disk over any document, including this one.

**Do.**

```bash
ls -1 .claude/skills/
grep -o '"invoke_separator": *"[^"]*"' .specify/integration.json
```

Then, in Claude Code, type `/spec` and let completion show you what exists.

**Check.** Ten skill directories: `speckit-analyze`, `speckit-checklist`,
`speckit-clarify`, `speckit-constitution`, `speckit-converge`,
`speckit-implement`, `speckit-plan`, `speckit-specify`, `speckit-tasks`,
`speckit-taskstoissues`.

---

### Step 3.6 — Write the shared permission allow-list

**What.** Create `.claude/settings.json` with the commands this loop needs.

**Why.** By default Claude Code asks before running each command. During a
spec-kit loop that means a prompt for every script invocation, which trains you
to approve without reading — the worst possible outcome. Allow-list the
mechanical commands so the prompts you *do* get mean something.

Settings are read highest-precedence-first: managed, `claude --settings`,
`.claude/settings.local.json`, `.claude/settings.json`, `~/.claude/settings.json`.
Commit the shared file; keep personal exceptions in the local one, which Step 3.1
already git-ignored.

**Notice what is deliberately absent:** `git commit`, `git push`, and anything
that writes outside the repository. The method depends on you seeing what
changed, and the commits are where you see it.

Settings files are **strict JSON** — a trailing comma or a `//` comment makes the
file a syntax error reported at start-up.

**Do.**

```bash
mkdir -p .claude
cat > .claude/settings.json <<'EOF'
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "permissions": {
    "allow": [
      "Bash(bash .specify/scripts/bash/*)",
      "Bash(just *)",
      "Bash(go build *)",
      "Bash(go test *)",
      "Bash(go vet *)",
      "Bash(go fmt *)",
      "Bash(gofmt *)",
      "Bash(protoc *)",
      "Bash(mvn -B *)",
      "Bash(shellcheck *)",
      "Bash(git status)",
      "Bash(git status *)",
      "Bash(git diff *)",
      "Bash(git log *)"
    ],
    "deny": [
      "Read(./.env)",
      "Read(./.env.*)"
    ]
  }
}
EOF
python3 -m json.tool .claude/settings.json > /dev/null && echo "valid JSON"
```

**Check.** `valid JSON` prints. Inside a Claude Code session, `/status` confirms
what actually loaded.

---

### Step 3.7 — Install the pre-commit hook

**What.** A git hook that refuses any commit mixing a specification change with a
source change.

**Why.** This is the single most important guard rail in the repository, and it
enforces the rule `CLAUDE.md` states and `PLAN.md` Appendix E explains: a
solo developer's only defence against an agent quietly rewriting a spec to match
code it has already written is being able to look at a specification diff *in
isolation* and ask whether it is something you approved. Mixed into a commit with
source changes, that diff is easy to skim past.

The hook lives in `.git/hooks/`, which is not version-controlled, so it must be
installed per clone — that is why Step 3.8 also stores a copy under `scripts/`.
It fires identically whether the staged change came from you, from Claude Code,
or from a different tool entirely.

**Do.**

```bash
cat > .git/hooks/pre-commit <<'EOF'
#!/usr/bin/env bash
# refuse commits that mix specification and source
staged=$(git diff --cached --name-only)
if grep -q '^specs/' <<<"$staged" && grep -qE '^(libs|cli|daemon|service)/' <<<"$staged"; then
    echo "refusing: this commit changes both a specification and source." >&2
    echo "commit the specification change on its own so it can be reviewed." >&2
    exit 1
fi
EOF
chmod +x .git/hooks/pre-commit
```

**Check.** Test it before you rely on it:

```bash
mkdir -p specs/000-hook-test libs
echo x > specs/000-hook-test/spec.md; echo x > libs/x.go
git add specs/000-hook-test/spec.md libs/x.go
git commit -m "should be refused"      # expect the refusal message, exit 1
git reset && rm -rf specs/000-hook-test libs/x.go
```

The commit is refused. If it succeeds, the hook is not executable.

---

### Step 3.8 — Keep a committed copy of the hook **`[not in PLAN.md]`**

**What.** Store the hook under `scripts/` with a one-line installer.

**Why.** `.git/hooks/` is not committed, so a fresh clone — or the day you move
this to another machine — silently has no guard rail. A committed copy plus an
install line makes the omission visible.

**Do.**

```bash
mkdir -p scripts
cp .git/hooks/pre-commit scripts/pre-commit.hook
cat >> scripts/README.md <<'EOF'
# scripts/

`pre-commit.hook` — refuses commits that mix `specs/` with source. Install with:

    install -m 0755 scripts/pre-commit.hook .git/hooks/pre-commit

`.git/hooks/` is not version-controlled, so this must be run once per clone.
EOF
```

**Check.** `ls scripts/`.

---

### Step 3.9 — Commit the scaffolding

**Do.**

```bash
git add -A
git status --short          # read this list before committing
git commit -m "chore: initialise spec-kit scaffolding

specify init . --integration claude --integration-options=--skills, plus the
shared permission allow-list and the spec/source pre-commit hook."
```

**Check.**

```bash
ls .specify/templates/                   # spec, plan, tasks, checklist, constitution
git ls-files .claude | head              # the skills ARE tracked
git ls-files .specify/feature.json       # prints nothing — correctly ignored
```

---

# Phase 4 — Write the constitution

**Goal:** `.specify/memory/constitution.md`, in your words, ratified at 1.0.0.

**Effort:** 2–4 hours.

**Why the whole phase exists.** This is the highest-leverage document in the
repository and the one most people write badly. It is the set of constraints
every later decision is checked against: every `plan.md` contains a Constitution
Check gate, and a plan that violates a principle must either change or justify
itself in writing in its Complexity Tracking table.

Concretely it is a gate, not a mission statement. From `.specify/memory/` it is
loaded by *every* spec-kit command — specify, plan, tasks, clarify, analyze,
checklist, implement, converge — each of which contains some form of "IF EXISTS:
load `.specify/memory/constitution.md` for project principles."

**The test a principle must pass:** *it would sometimes stop you doing something
you would otherwise do.* "Write clean code" fails — nobody was going to write
dirty code on purpose, so it constrains nothing. "Domain logic never lives in a
command handler; an import of the CLI framework inside the domain library is a
build-visible violation" passes: it forbids a specific, tempting, otherwise
reasonable shortcut, and you can check it mechanically.

---

### Step 4.1 — 🧠 Draft your principles as bullets, yourself

**What.** Write the principles in bullet form, in a scratch file, before the
agent sees anything.

**Why.** `PLAN.md` is explicit about the direction: *you* draft, then the agent
turns your bullets into the document — not the other way round. If the agent
drafts the principles, you end up reviewing a plausible document instead of
deciding what you actually believe, and the constitution becomes something you
inherited rather than something you wrote.

Four items from `IDEAS.md` are, per `docs/CONTEXT.md` §7, much cheaper to write
in now than to retrofit into `molfar` in Phase 15: an **oracle on every step**;
**repository content is untrusted input** — it may influence the *content* of a
proposal, never the *policy* (scope, budget, approval requirement), enforced
structurally rather than with a warning; **observe-only first**; and an
**evaluation corpus**.

**Do.**

```bash
cat > docs/constitution-draft.md <<'EOF'
# Constitution bullets — my words, before the agent sees them

For each principle: state it, then answer "when would this stop me?" in one
line. If you cannot name a case, cut the principle or sharpen it.

## I. Specification is the source of truth
- No behaviour exists until specified. Code contradicting its spec is a defect
  in one of them, and the loop is not closed until one is corrected.
- Every requirement carries a stable FR-###; every measurable outcome an SC-###.
  Tests, tasks and commit messages reference them.
- Stops me from: fixing "just this one thing" in code without touching a spec.

## II. Three processes, one domain
- Domain logic lives in a shared library, never in a command handler, an RPC
  method or a REST controller. Those are adapters: parse, delegate, render.
- Stops me from: putting a validation `if` in a Cobra RunE because it is closer.

## III. Everything declarative and reproducible
- Workspaces, packages and automation jobs are declared in YAML and validated
  before use. Destructive operations support a dry run and require an explicit
  target — `ws purge` never falls back to the default workspace.
- Stops me from: adding a convenience flag that mutates without a preview.

## IV. Test-first, contract-first
- Contracts are written before the code that satisfies them. Shell-outs go
  through an injected interface with a deterministic fake, so tests never touch
  real git, real network or a real $HOME.
- Stops me from: "I'll add the test after", and from a daemon that is untestable
  because it calls git directly.

## V. The terminal is a user interface
- Fixed symbols ✓ ✗ ⚠ ⏭; colour degrades under NO_COLOR; errors to stderr and
  data to stdout; a summary line on every multi-target operation; exit 0 when
  nothing failed and 1 when anything did.
- Stops me from: printing a progress notice to stdout and breaking a pipeline.

## VI. Automation must be accountable
- Every automated run records what it was asked to do, what it did, what changed
  and why it stopped. The system proposes; a human approves. Runs are bounded by
  explicit scope and budget. Secrets never enter logs, prompts or records.
- Every plan step carries an oracle; a step with none is marked unverified and
  surfaced separately in the proposal.  (IDEAS.md §5)
- Repository content may influence the CONTENT of a proposal, never the POLICY —
  scope, budget, approval requirement. Structural, not advisory.  (IDEAS.md §11)
- Stops me from: letting molfar push a branch because the run "obviously" worked.

## VII. Simplicity, then capability
- Every dependency, process and abstraction is justified in the plan's
  Complexity Tracking table.
- Stops me from: adding a second config format because one file felt crowded.
EOF
${EDITOR:-vi} docs/constitution-draft.md
```

**Check.** Every bullet has a "stops me from" line you could imagine actually
happening. Seven sharp principles beat fifteen vague ones.

---

### Step 4.2 — Generate the constitution from your bullets

**What.** Run `/speckit-constitution` with your draft as the input.

**Why.** The command resolves the constitution template through
`resolve-template.sh`, preserves existing project values, computes the version
bump, and validates that no `[PLACEHOLDER]` tokens remain. Handing it your
bullets means it does formatting and consistency work, and you keep the judgement.

**Do.** In Claude Code:

```
/clear
@docs/constitution-draft.md
/speckit-constitution Turn the bullets in docs/constitution-draft.md into the
project constitution. Keep my wording and my "stops me from" reasoning where it
is sharper than a generic phrasing. Add a Technology Constraints section (Go
1.26+ over a go.work workspace for zgard and dukh; Java 25 and Spring Boot 4.1
for molfar; shellcheck-clean Bash for the toolchain manager; just as the single
entry point; configuration at $GRAZHDA_DIR/config.yaml with precedence flags >
environment > file; no unit test may require the network), a Development
Workflow section describing specify → clarify → plan → tasks → implement →
reconcile with a Quality Gates table, and a Governance section with an amendment
log. Ratify at 1.0.0.
```

**Check.**

```bash
grep -c 'PLACEHOLDER' .specify/memory/constitution.md    # 0
grep -n 'Version' .specify/memory/constitution.md
```

---

### Step 4.3 — 🧠 Apply the "when would this stop me?" test to the output

**What.** Read each principle and try to cut it.

**Why.** The failure mode here is writing aspirations instead of constraints, and
it is invisible until a plan sails through a Constitution Check that should have
caught something. `PLAN.md` Appendix D.4 shows the target specificity: Principle I
at full strength "names an artefact path, defines what counts as a violation,
fixes the identifier every requirement carries, and says when the loop closes."

**Do.**

```bash
less .specify/memory/constitution.md
${EDITOR:-vi} .specify/memory/constitution.md
```

For each principle ask: *what would I have done last month that this forbids?*
No answer means cut it or sharpen it.

**Check.** Every principle names something checkable. At least one names a file
path or an identifier format.

---

### Step 4.4 — Verify the sections and the amendment log exist

**What.** Confirm the four structural sections landed.

**Why.** Governance is what makes the document survive its first disagreement.
Semantic versioning applies to the *promises* the document makes: MAJOR when a
principle is removed or redefined, because plans approved under the old text may
no longer be valid; MINOR when one is added, because nothing already approved
becomes wrong; PATCH for clarification that changes no obligation. Without a log,
a constitution becomes just another document nobody trusts.

**Do.**

```bash
grep -nE '^#{1,3} (Core Principles|Technology Constraints|Development Workflow|Governance|Quality Gates|Amendment log)' \
  .specify/memory/constitution.md
```

**Check.** All six headings present, and the file ends with a line carrying
**Version**, **Ratified** and **Last Amended**.

---

### Step 4.5 — Commit, and clean up the draft

**Do.**

```bash
rm docs/constitution-draft.md      # or keep it; it is a record of your thinking
git add -A
git commit -m "docs: ratify the constitution at 1.0.0

Seven principles, technology constraints, the development workflow with its
quality gates, and governance with an amendment log."
```

**Check.** `git show --stat HEAD`.

---

# Phase 5 — Decide the architecture

**Goal:** `docs/architecture.md` — the technical design every feature plan refers
to instead of re-deciding.

**Effort:** 2–3 hours.

**Why the whole phase exists.** If each `plan.md` re-litigates the technology you
get eight inconsistent designs. Decide the cross-cutting things once, write down
the alternatives you rejected and why, and let the feature plans be about
features.

---

### Step 5.1 — Create the file from its skeleton

**What.** Write `docs/architecture.md` with its five required sections.

**Why.** The five sections are the five questions a feature plan would otherwise
ask you again. The most valuable of them is the **technology decision table with
a rejected alternative on every row** — a decision without its rejected
alternative is indistinguishable from an accident six months later.

**Do.**

```bash
cat > docs/architecture.md <<'EOF'
# Architecture

The technical design every feature plan refers to instead of re-deciding.
Governed by `.specify/memory/constitution.md`; where the two disagree, the
constitution wins.

## 1. The three processes and their lifetimes

| Process | Binary | Language | Lifetime | Role |
| :--- | :--- | :--- | :--- | :--- |
| CLI interface | `zgard` | Go 1.26 | milliseconds, dozens of times an hour | the engineer's hands |
| CLI daemon | `dukh` | Go 1.26, gRPC | days; restarted on upgrade | local eyes and reflexes |
| Long-term process | `molfar` | Java 25 / Spring Boot 4.1 | weeks | memory and judgement |
| Toolchain manager | `grazhda` | Bash | invoked | installs, upgrades, removes the rest |

**Why two languages.** A CLI invoked dozens of times an hour cannot pay JVM
start-up, which disqualifies Java for `zgard`; Go gives a single static binary,
no runtime on the engineer's machine, and first-class concurrency for fanning out
across repositories. A process that must run for weeks benefits from a mature
scheduler, persistence, transaction and observability stack and pays start-up
once, which is `molfar`. `dukh` shares one domain library with the CLI, so the
same argument puts it in Go. `grazhda` installs the other three and therefore
cannot depend on them, so it is Bash.

**The honest cost** of the Spring Boot choice is several hundred megabytes
resident on a laptop already running an IDE and a browser. Three planned
mitigations: `dukh` starts `molfar` on demand and stops it when idle; the domain
stays in a framework-free `core` package so a native-image build remains viable;
running it as a container stays open.

## 2. Repository layout

    libs/        shared Go domain — config, workspace, pkgman, reporter, executor
    cli/         zgard — Cobra command tree; parse, delegate, render
    daemon/      dukh — cmd/ (entry point), server/ (gRPC), proto/ (generated)
    service/     molfar — Maven, Spring Boot
    contracts/   dukh.proto, molfar-openapi.yaml — the only cross-process coupling
    scripts/     grazhda, grazhda-init.sh, grazhda-install.sh
    specs/       one directory per feature
    docs/        this file, product.md, architecture, parity matrix, v1-baseline
    tests/bash/  the shell harness
    bin/         build output (git-ignored)

Joined by `go.work` at the root listing `./libs ./cli ./daemon`.

## 3. Layering rules — stated so they can be checked

- The domain (`libs/`) knows nothing about terminals, flags or transports. It
  never imports `cobra`, never calls `os/exec` (it calls `executor.Executor`),
  never writes to `os.Stdout` (it writes to an `io.Writer`).
- Adapters (`cli/`, `daemon/cmd`, `daemon/server`, molfar's controllers) parse,
  delegate and render. They contain no domain branching.
- Cross-process coupling goes through `contracts/` only.
- `cli` depends on `daemon` **solely for the generated gRPC client**, which
  happens to live in the daemon module. It must never reach into
  `daemon/server`.
- Go forbids import cycles at the compiler level, with no escape hatch, so the
  one-way direction `cli, daemon → libs` is enforced by the build rather than by
  review convention. That is why `libs` is a separate module.
- Two tests confirm the line is in the right place: could a second frontend (a
  TUI, an HTTP API) be added without touching `libs/`? Can every domain test run
  offline, in parallel, in under a second, with no real git?

## 4. Technology decisions

| Decision | Chosen | Rejected alternative | Why |
| :--- | :--- | :--- | :--- |
| CLI language | Go 1.26 | Java for all three | JVM start-up on a command run dozens of times an hour is disqualifying |
| Long-term process | Java 25 / Spring Boot 4.1 | Go for all three | would mean rebuilding scheduling, persistence and migration machinery by hand |
| Go module layout | three modules + `go.work` | one module, many packages | a module boundary is a hard boundary; `libs` should not carry gRPC and log-rotation dependencies |
| Cross-process contract | protobuf / gRPC | hand-written JSON over HTTP | one contract, two generated clients (Go and Java) |
| Generated code | committed | generated at build time | a fresh clone builds without `protoc`; CI regenerates and diffs |
| Build entry point | `just` | `make`, or per-language scripts | one verb per action across four toolchains; no timestamp semantics to misread |
| Spring dependencies | starters only | bare libraries | Boot 4 auto-config is modular: a bare library gets no wiring and no error at all |
| *(add your own rows)* | | | |

## 5. Deliberate non-goals

- No multi-tenant operation. This runs on one engineer's machine.
- No credential storage. It uses the engineer's existing git credentials.
- No autonomous publication. The system proposes; a human approves.
- No scale engineering for a fleet. Designing for a scale you do not have is the
  named pitfall of this phase.
EOF
${EDITOR:-vi} docs/architecture.md
```

**Check.** `PLAN.md`'s test: someone who reads only this file could lay out the
repository correctly and would not have to ask you a single "why is it like
that?" question.

---

### Step 5.2 — Point `CLAUDE.md` at it **`[not in PLAN.md]`**

**What.** Add `docs/architecture.md` to the reading order in `CLAUDE.md`.

**Why.** `CLAUDE.md` is auto-loaded every session and currently names four files
to read; the architecture document is now the fifth, and the one that answers
layering questions. The version in `guides/03-spec-kit.md` §12.2 lists it
explicitly.

**Do.**

```bash
${EDITOR:-vi} CLAUDE.md    # add docs/architecture.md to "Read before doing anything"
git add -A
git commit -m "docs: architecture — three processes, layout, layering rules"
```

**Check.** `grep -n architecture CLAUDE.md`.

---

# Phase 6 — Skeleton and build system

**Goal:** an empty but complete repository where `just build` and `just test`
both succeed.

**Effort:** half a day.

**Why the whole phase exists.** Get the loop green before there is anything in it
to break. If the build works when the project does nothing, then the first time
it fails you know exactly what caused it.

**Read `guides/02-go.md` Parts 2 and 6, and `guides/01-just.md`, before this
phase.** They teach modules, workspaces and recipes properly; the steps below
assume you have.

---

### Step 6.1 — Create the directory skeleton

**What.** The eight top-level directories the architecture names.

**Why.** In Go a **package is a directory** — the unit of compilation and
visibility — and a **module is a tree of packages with a `go.mod`** — the unit of
versioning and dependency. Creating the directories first makes the module
boundaries you are about to declare visible as structure rather than as three
`go mod init` invocations you have to remember the shape of.

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
mkdir -p libs cli daemon/cmd service scripts contracts tests/bash bin
ls -1
```

**Check.** All eight exist. `bin/` is git-ignored already (Step 3.1).

---

### Step 6.2 — Initialise the three Go modules

**What.** A `go.mod` in `libs/`, `cli/` and `daemon/`.

**Why three modules rather than one.** The default that serves most projects is
one module with many packages, and nothing about having several binaries requires
several modules. Grazhda takes three for three reasons: the dependency sets
differ enough that they should not be merged (`daemon` pulls in gRPC, protobuf, a
logging framework and a log rotator; `libs` should carry none of that); they
version at different rates; and — the architectural reason — **a module boundary
is a hard boundary**. `libs` cannot import `cli`, because `libs/go.mod` does not
require it and adding that requirement would create a cycle the Go tooling
rejects outright. The layering is enforced by the build, not by review.

The module path is the import path other code uses. `github.com/vhula/grazhda/libs`
does not need to exist as a URL for local development, but it must be globally
unique and stable, so it is conventionally a repository URL.

**Do.**

```bash
cd libs   && go mod init github.com/vhula/grazhda/libs   && cd ..
cd cli    && go mod init github.com/vhula/grazhda/cli    && cd ..
cd daemon && go mod init github.com/vhula/grazhda/daemon && cd ..
head -3 libs/go.mod cli/go.mod daemon/go.mod
```

**Check.** Three `go.mod` files, each with the right `module` line and a `go`
directive of 1.26 or later.

---

### Step 6.3 — Create the Go workspace

**What.** `go.work` at the repository root listing the three modules.

**Why.** Multi-module repositories have an obvious problem: `cli` imports
`github.com/vhula/grazhda/libs`, which by default resolves to a *published*
version from the module proxy — so editing `libs` would not affect the CLI build
until you tagged a release. `go.work` (Go 1.18+) makes any `go` command inside
the tree resolve those module paths to the local directories instead. The Maven
analogy is a parent POM's `<modules>` list, with one difference: `go.work` is
about *resolution*, not aggregation. It declares no shared versions and no build
order.

**Commit it.** The original advice was to git-ignore `go.work` as a developer's
local view, but in a repository whose modules only make sense together —
which this one is — committing it means a fresh clone builds with no setup.
`go.work.sum` is a different matter and stays ignored (Step 3.1): it records how
the *workspace* resolved modules on your machine, not a portable lock, so
committing it invites merge conflicts that carry no information.

**Do.**

```bash
go work init ./libs ./cli ./daemon
cat go.work
```

**Check.** `go.work` lists all three under `use (…)`.

---

### Step 6.4 — Add `replace` directives so each module builds standalone

**What.** Point each module's inter-repository dependencies at sibling
directories in its own `go.mod`.

**Why.** `go.work` covers your machine but not `GOWORK=off`, and not someone
running `go install github.com/vhula/grazhda/cli@latest`. A `replace` in a
module's own `go.mod` states the local relationship independently of the
workspace, which is what lets CI verify that each module still builds on its own.

The `v0.0.0` version is a placeholder that is never resolved, because the
replacement intercepts it. Three properties worth knowing: a `replace` in the
**main** module applies while one in a *dependency* is ignored, so you cannot
force a replacement on consumers; a module with a filesystem `replace` cannot be
`go get`-ed by an outside consumer; and both of those are fine here, because
these are programs, not a library meant for external consumption.

**Note what the `cli` → `daemon` dependency is for:** the CLI consumes only the
daemon's *generated gRPC client*, which happens to live in the daemon module. It
must never reach into `daemon/server`. `docs/architecture.md` §3 says so, and
this is the step where the temptation is created.

**Do.**

```bash
cat >> cli/go.mod <<'EOF'

require (
	github.com/vhula/grazhda/daemon v0.0.0
	github.com/vhula/grazhda/libs   v0.0.0
)

replace (
	github.com/vhula/grazhda/daemon => ../daemon
	github.com/vhula/grazhda/libs   => ../libs
)
EOF

cat >> daemon/go.mod <<'EOF'

require github.com/vhula/grazhda/libs v0.0.0

replace github.com/vhula/grazhda/libs => ../libs
EOF

cat cli/go.mod daemon/go.mod
```

**Check.** Both files parse: `cd cli && go list ./... ; cd ..` prints no error.
(Unused `require` lines are tolerated until there is code; `go mod tidy` inside
the specific module directory — never at the repository root — is the fix if one
complains.)

---

### Step 6.5 — Write the CLI entry point

**What.** A minimal `cli/main.go` that prints a version and exits.

**Why.** Two things at once. It gives `just build-cli` something to compile, and
it establishes the `-ldflags "-X main.version=…"` mechanism the Justfile uses to
bake a version string into the binary at link time. `main.version` is a package
level variable the linker overwrites; the `dev` fallback is what a build outside
a git checkout gets.

Cobra does not appear yet. It arrives in Phase 8 with the first feature, and
keeping it out now means `just build` proves the *toolchain* works before it
proves any dependency does.

**Do.**

```bash
cat > cli/main.go <<'EOF'
// Command zgard is the grazhda command-line interface.
package main

import (
	"fmt"
	"os"
)

// version is set at link time: -ldflags "-X main.version=…"
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println("zgard", version)
		return
	}
	fmt.Fprintln(os.Stderr, "zgard: no commands yet; see specs/")
	os.Exit(1)
}
EOF
```

**Check.**

```bash
cd cli && go build -o /tmp/zgard . && /tmp/zgard --version && cd ..
```

Prints `zgard dev`.

---

### Step 6.6 — Write the daemon entry point

**What.** The same, at `daemon/cmd/main.go`.

**Why.** `dukh`'s `main` package lives in a `cmd` subdirectory rather than the
module root, which is why the Justfile builds `./cmd` rather than `.`. The
separation exists because the daemon module will also hold `server/` and
`proto/` packages that the binary imports — a module root full of a `main`
package and library packages is legal but harder to read.

**Do.**

```bash
cat > daemon/cmd/main.go <<'EOF'
// Command dukh is the grazhda health daemon.
package main

import (
	"fmt"
	"os"
)

// version is set at link time: -ldflags "-X main.version=…"
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println("dukh", version)
		return
	}
	fmt.Fprintln(os.Stderr, "dukh: no commands yet; see specs/")
	os.Exit(1)
}
EOF
```

**Check.**

```bash
cd daemon && go build -o /tmp/dukh ./cmd && /tmp/dukh --version && cd ..
```

---

### Step 6.7 — Write the Justfile

**What.** The build entry point, with only the recipes that make sense today.

**Why, and the one thing that bites everyone.** **Each line of a `just` recipe
runs in its own shell.** `cd cli` on one line does not affect the next — so
`cd cli` followed by `go build` on separate lines builds the wrong directory,
silently. Chain them with `&&` on one line, as below. That single rule is the
most common `just` mistake and `guides/01-just.md` §5 is devoted to it.

Two smaller conventions in the file: `@` at the start of a line suppresses the
echo of the command itself (used for housekeeping like `mkdir -p`), and
`default: help` at the very top makes bare `just` print the help, because `just`
with no argument runs whichever recipe is written first.

`build-service` and the Java and Bash test recipes are deliberately absent —
`service/` and `tests/bash/` have no content yet, and a recipe that fails because
its subject does not exist is worse than no recipe. They arrive in Phases 14 and
15.

**Do.**

```bash
cat > Justfile <<'EOF'
default: help

# Show the available tasks
help:
    @echo "Grazhda — spec-driven build system"
    @echo ""
    @echo "Usage: just <task>"
    @echo ""
    @echo "  build          Build every component into bin/"
    @echo "  build-cli      Build the CLI interface (zgard)"
    @echo "  build-daemon   Build the CLI daemon (dukh)"
    @echo "  test           Run every test suite"
    @echo "  test-go        Run the Go test suites"
    @echo "  fmt            Format all sources"
    @echo "  lint           Vet Go sources"
    @echo "  clean          Remove build output"

# Build the CLI interface (zgard)
build-cli:
    @mkdir -p bin
    cd cli && go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/zgard .
    @echo "✓ bin/zgard"

# Build the CLI daemon (dukh)
build-daemon:
    @mkdir -p bin
    cd daemon && go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/dukh ./cmd
    @echo "✓ bin/dukh"

# Build every component into bin/
build: build-cli build-daemon
    @echo "✓ all components built"

# Run the Go test suites
test-go:
    cd libs && go test ./...
    cd cli && go test ./...
    cd daemon && go test ./...

# Run every test suite
test: test-go
    @echo "✓ all tests passed"

# Format all sources
fmt:
    cd libs && go fmt ./...
    cd cli && go fmt ./...
    cd daemon && go fmt ./...

# Vet Go sources
lint:
    cd libs && go vet ./...
    cd cli && go vet ./...
    cd daemon && go vet ./...

# Remove build output
clean:
    rm -rf bin
    @echo "✓ clean"
EOF
```

**Check.**

```bash
just --list          # every recipe, with its comment as the description
just --evaluate      # parses the file without running anything
```

---

### Step 6.8 — Run the build and the tests

**What.** The phase's actual goal.

**Why.** This is the moment the loop goes green with nothing in it. From here,
any failure of `just build` is caused by the change you just made.

**Do.**

```bash
just build
./bin/zgard --version
./bin/dukh --version
just test
```

**Check.** Two binaries, two version lines, and `test` passing trivially — there
are no tests yet, and `go test ./...` on a package with none reports
`no test files`, which is a pass.

**If `cd` seems to have no effect** — a build lands in the wrong directory, or
`go build` reports no Go files — re-read Step 6.7's warning. That is the
one-shell-per-line rule.

---

### Step 6.9 — Add the CI workflow

**What.** `.github/workflows/ci.yml` running the same recipes you just ran.

**Why.** The entire point of putting build logic in a Justfile rather than
scattering it across workflow YAML is that the same verbs run identically in both
places. When CI fails on `just test`, you run that exact command locally and
expect the same behaviour — same recipes, same dependency order, same failure
output — because there is no separate CI script reimplementing the logic with
different flags. Debugging becomes "run this one command", not
"reverse-engineer the YAML".

Pin the tool versions, for the reason in `PLAN.md` Appendix B: a CI that installs
"latest" while you installed something else produces failures that appear only
sometimes.

**Do.**

```bash
mkdir -p .github/workflows
cat > .github/workflows/ci.yml <<'EOF'
name: ci
on: [push, pull_request]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: extractions/setup-just@v2
        with:
          just-version: "1.36.0"
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26.1"
      - run: just build
      - run: just test
      - run: just lint
EOF
```

*(The Java setup step and the `just generate` diff check are added in Phases 15
and 7 respectively, when there is something for them to do.)*

**Check.** `python3 -c "import yaml,sys; yaml.safe_load(open('.github/workflows/ci.yml'))"`
parses without error, if PyYAML is available; otherwise read it.

---

### Step 6.10 — Format, then commit the skeleton

**What.** Run `go fmt` before the first Go commit.

**Why.** Formatting in Go is not a matter of taste — `gofmt` has no options, and
every Go project is formatted identically. Running it before the first commit
means no later diff is polluted by whitespace churn.

**Do.**

```bash
just fmt
just lint
git add -A
git status --short
git commit -m "chore: repository skeleton and build system

Three Go modules joined by go.work with intra-repo replace directives,
minimal zgard and dukh entry points, the Justfile as the single build
entry point, and CI running the same recipes."
```

**Check.**

```bash
just clean && just build && just test && git status --short   # clean tree
```

---

# Phase 7 — The daemon contract

**Goal:** `contracts/dukh.proto`, generated Go committed at `daemon/proto/`, and
a `just generate` recipe with a CI check that catches staleness.

**Effort:** 2–3 hours.

**Why the whole phase exists.** The contract is the only coupling between the
CLI, the daemon and later the automation service. Writing it *before* either side
exists forces you to decide what the daemon actually knows, rather than letting
the answer emerge from whatever the implementation happened to expose.
Contract-first is a constitutional principle, and this is where it first bites.

---

### Step 7.1 — 🧠 Read the v1 contract, then decide rather than copy

**What.** Read `docs/v1-baseline/dukh.proto` and decide what the new one says.

**Why.** The old contract is evidence about what the daemon needed to expose, not
a specification you owe parity to at the field level. `PLAN.md` Phase 7 names the
three operations — report health, trigger a rescan, shut down gracefully — and
the payload: per workspace, per project, per repository, with the configured
branch, the actual branch, whether the repository exists and whether the branch is
aligned. That is the decision; the old file is the sanity check.

**Do.**

```bash
less docs/v1-baseline/dukh.proto
```

**Check.** You can name the three RPCs and every field of the health payload
without looking.

---

### Step 7.2 — Write `contracts/dukh.proto`

**What.** The contract file.

**Why each piece is there.**

- **`syntax = "proto3"`** — the current protobuf language version.
- **`package grazhda.dukh.v1`** — the protobuf namespace. The trailing `v1` is
  how you get a second, incompatible version later without renaming anything.
- **`option go_package`** — where the generated Go lands and what its package is
  called (`;dukhpb` sets the package name). Omit it and generated code arrives in
  an unexpected package with imports that do not resolve; `PLAN.md` Appendix B
  lists this as a named failure.
- **`option java_package`** — because `molfar` generates a **second** client from
  the same file in Phase 15. One contract, two generated clients, is much of the
  argument for protobuf over hand-written JSON.
- **`reserved`** — the pitfall this phase names. When you later remove a field,
  mark its number `reserved` so it is never reused; otherwise an old client and a
  new server will disagree about what a field means, **silently**. Reserving a
  block ahead of time in messages you expect to grow costs nothing.

**Do.**

```bash
cat > contracts/dukh.proto <<'EOF'
syntax = "proto3";

package grazhda.dukh.v1;

option go_package   = "github.com/vhula/grazhda/daemon/proto;dukhpb";
option java_package = "dev.grazhda.contract.dukh.v1";

// DukhService is the only interface dukh exposes. zgard and molfar are both
// clients of it; neither reaches into the daemon any other way.
service DukhService {
  // Status reports the health snapshot, optionally scoped to one workspace and
  // optionally forcing a fresh scan first and waiting for it.
  rpc Status(StatusRequest) returns (StatusResponse);
  // Scan requests a rescan without waiting for it. Fire and forget.
  rpc Scan(ScanRequest) returns (ScanResponse);
  // Stop requests graceful shutdown: stop accepting work, finish what is in
  // flight, remove the identity file, exit.
  rpc Stop(StopRequest) returns (StopResponse);
}

message StatusRequest {
  string workspace_name = 1;  // empty means every workspace
  bool   rescan         = 2;  // scan first and wait for it
  reserved 3 to 15;
}

message StatusResponse {
  repeated WorkspaceHealth workspaces     = 1;
  string                   server_version = 2;
  int64                    uptime_seconds = 3;
  reserved 4 to 15;
}

message WorkspaceHealth {
  string                 name     = 1;
  repeated ProjectHealth projects = 2;
  reserved 3 to 7;
}

message ProjectHealth {
  string                    name         = 1;
  repeated RepositoryHealth repositories = 2;
  reserved 3 to 7;
}

message RepositoryHealth {
  string name              = 1;  // the full configured name
  string path              = 2;  // resolved local path, exactly as ws init would
  string configured_branch = 3;
  string actual_branch     = 4;  // empty when the repository is not present
  bool   exists            = 5;
  bool   aligned           = 6;  // actual_branch == configured_branch
  reserved 7 to 15;
}

message ScanRequest  { string workspace_name = 1; reserved 2 to 7; }
message ScanResponse { bool queued = 1; reserved 2 to 7; }

message StopRequest  { reserved 1 to 7; }
message StopResponse { bool accepted = 1; reserved 2 to 7; }
EOF
```

**Check.** It parses:

```bash
protoc --proto_path=contracts --descriptor_set_out=/dev/null contracts/dukh.proto \
  && echo "proto OK"
```

---

### Step 7.3 — Add the `generate` recipe and wire it into `build-daemon`

**What.** A Justfile recipe that runs `protoc`, and a dependency making the
daemon build run it first.

**Why.** Code generation in Go is a build **step**, not a library call — nothing
in `go.mod` triggers it, because Go's build has no plugin phases. Less magic than
Maven's protobuf plugin, more explicit steps, and a real risk of a stale
generated file, which is what Step 7.6 makes CI catch.

`--go_opt=paths=source_relative` tells the generator to place output alongside
the input's relative path rather than reconstructing a directory tree from the
`go_package` option. Drop it and files land somewhere you did not expect.

`build-daemon: generate` makes the dependency explicit and unconditional — every
`just build-daemon` reruns `protoc` in full whether or not the proto changed.
That is a deliberate cost (`protoc` is fast) bought for the certainty that
generated code is never stale. `just` has no timestamp-based skipping; recipes
always run.

**Do.**

```bash
python3 - <<'PY'
import re, pathlib
p = pathlib.Path("Justfile"); s = p.read_text()
gen = '''# Regenerate contract code from contracts/
generate:
    @mkdir -p daemon/proto
    protoc \\
      --go_out=daemon/proto --go_opt=paths=source_relative \\
      --go-grpc_out=daemon/proto --go-grpc_opt=paths=source_relative \\
      --proto_path=contracts \\
      contracts/dukh.proto
    @echo "✓ daemon/proto generated from contracts/dukh.proto"

'''
s = s.replace("# Build the CLI interface (zgard)", gen + "# Build the CLI interface (zgard)", 1)
s = s.replace("build-daemon:\n", "build-daemon: generate\n", 1)
s = s.replace('    @echo "  build          Build every component into bin/"',
              '    @echo "  generate       Regenerate contract code from contracts/"\n'
              '    @echo "  build          Build every component into bin/"', 1)
p.write_text(s)
PY
grep -n -A9 '^generate:' Justfile
grep -n '^build-daemon' Justfile
```

*(Editing by hand in `$EDITOR` is equally fine — the script is only here so the
step is reproducible.)*

**Check.** `just --list` shows `generate`, and `build-daemon` lists it as a
dependency.

---

### Step 7.4 — Generate the Go, and add the runtime dependencies

**What.** Run the generator, then let Go resolve the imports it produced.

**Why.** `protoc-gen-go` emits the message structs; `protoc-gen-go-grpc` emits
the client and server interfaces. Both import `google.golang.org/protobuf` and
`google.golang.org/grpc` at runtime, which must therefore appear in
`daemon/go.mod`. `go get` inside the specific module directory — never at the
repository root — adds them.

**Do.**

```bash
just generate
ls -1 daemon/proto/
cd daemon && go get google.golang.org/grpc google.golang.org/protobuf && cd ..
cd daemon && go mod tidy && cd ..
```

**Check.**

```bash
ls daemon/proto/dukh.pb.go daemon/proto/dukh_grpc.pb.go
head -12 daemon/proto/dukh.pb.go     # note the "Code generated … DO NOT EDIT." header
cd daemon && go build ./... && cd ..
```

---

### Step 7.5 — Commit the generated code deliberately

**What.** Stage `daemon/proto/` as source.

**Why.** Generated Go is **committed** — the normal Go convention — precisely so
that a fresh clone builds with `just build` before `protoc` is even installed.
`PLAN.md` Appendix E calls this out as one of the two things that look like build
output and must not be ignored (the other being the per-module `go.sum` files).

**Do.**

```bash
git add contracts/ daemon/ Justfile
git status --short
git commit -m "feat(contract): dukh gRPC contract and generated Go

Three operations — Status, Scan, Stop — and the health payload: per
workspace, per project, per repository, with configured and actual branch,
existence and alignment. java_package is set because molfar generates a
second client from this file in Phase 15."
```

**Check.** `git ls-files daemon/proto | wc -l` is 2.

---

### Step 7.6 — Make CI fail on stale generated code

**What.** Add a regenerate-and-diff step to the workflow.

**Why.** Committed generated code has exactly one failure mode: somebody edits
the `.proto` and forgets to regenerate, so the contract and the code disagree and
nothing says so. `git diff --exit-code` returns non-zero when there are
differences, which turns that silence into a build failure. This is the check
`PLAN.md` Phase 7 requires.

Pin the plugin versions here to the same values as Step 1.5, for the reason in
Appendix B: an unpinned CI and a pinned laptop produce a diff that appears only
sometimes.

**Do.**

```bash
python3 - <<'PY'
import pathlib
p = pathlib.Path(".github/workflows/ci.yml"); s = p.read_text()
s = s.replace("      - run: just build", """      - name: Install protoc and plugins
        run: |
          sudo apt-get update && sudo apt-get install -y protobuf-compiler
          go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
          go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
      - name: Generated code is up to date
        run: |
          just generate
          git diff --exit-code -- daemon/proto
      - run: just build""", 1)
p.write_text(s)
PY
cat .github/workflows/ci.yml
```

**Check.** Reproduce what CI does, locally:

```bash
just generate && git diff --exit-code -- daemon/proto && echo "generated code current"
```

---

### Step 7.7 — The phase check

**Do.**

```bash
just generate && just build
git status --short          # must be empty
./bin/dukh --version
git add -A && git commit -m "ci: fail the build on stale generated contract code"
```

**Check.** `git status` is clean **immediately after** `just generate && just
build`. A dirty tree here means generation is not deterministic — usually a
plugin version difference, and the reason Step 1.5 pinned them.

---

# Phase 8 — Feature 001, the full loop

**Goal:** `specs/001-config-core/` complete, the configuration domain implemented
and tested, and the spec reconciled with what shipped.

**Effort:** 1–2 days. Go slowly. Everything after this phase is this same
sequence with less commentary.

**The feature.** Configuration is the declarative description of the engineer's
working world — workspaces, projects, repositories, daemon settings — plus
everything needed to read, validate, query and safely change it. It is first
because nothing else can run without it.

**Six sessions, six commits.** Each of the steps below is its own Claude Code
session, opened with `/clear`. The reason is in `guides/03-spec-kit.md` §12.4 and
it is worth restating: everything in the context window is evidence about what
you want, so a specification drafted after forty minutes of debugging a decoder
will mention decoders. The artefacts on disk are the memory; the conversation is
not.

---

### Step 8.1 — Create the feature directory

**What.** Run spec-kit's own script to make `specs/001-config-core/`.

**Why use the script rather than `mkdir`.** It computes the next feature number
by scanning `specs/` for the highest numeric prefix (forcing base 10, so `010`
does not become 8), writes `.specify/feature.json` — **the pointer spec-kit uses
to decide which feature is active** — and emits the resolved paths as JSON.
`--number` is a *preference*: if that number is taken it warns and takes the next
free one.

**Spec-kit does not read your git branch.** Current versions key entirely off
`.specify/feature.json`; branch creation is an optional hook. Nearly all
third-party writing about spec-kit says otherwise. This matters the first time
something targets the wrong feature.

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
git checkout -b 001-config-core
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name config-core --number 1 "configuration core"
cat .specify/feature.json
ls specs/001-config-core/
```

**Check.** `.specify/feature.json` reads
`{"feature_directory":"specs/001-config-core"}` and the directory exists.

---

### Step 8.2 — Session 1: specify

**What.** Draft `spec.md` from a description in your own words.

**Why.** `/speckit-specify` copies the resolved spec template and fills it in. Two
behaviours of the command to know before you read its output. First, on
ambiguity it is told to *decide rather than interrogate* — "make informed guesses
based on context and industry standards", with a hard limit of three
`[NEEDS CLARIFICATION]` markers — and to record what it decided in an
**Assumptions** section. That is the right default, and it means **the
Assumptions section is where the agent's guesses live**, and a wrong guess there
is silent. Read it first. Second, the command self-validates: it writes
`checklists/requirements.md` with items like "No implementation details" and
reviews its own output against them up to three times.

**Do.** In Claude Code:

```
/clear
/speckit-specify configuration core: the declarative description of an
engineer's workspaces, projects, repositories and daemon settings, plus
everything needed to read, validate, query and safely change it. Validation
must report every problem in one pass, address each offending field by its
path including indices, and reject unknown structure values rather than
falling through to a default. A missing config file, an empty config file and
a config with no default workspace are all distinct outcomes.
```

**Check.**

```bash
ls specs/001-config-core/
wc -l specs/001-config-core/spec.md
```

---

### Step 8.3 — 🧠 Read the specification properly

**What.** The most valuable half-hour in the phase. Read it in an editor, not in
the terminal.

**Why.** Everything downstream inherits this document's mistakes, and the agent
will implement a vague requirement confidently rather than asking you again.
There are five things to check, and one of them people skip.

1. **Does it say *how* anywhere?** A spec that mentions YAML libraries, Go types
   or file names has failed. The test: if a sentence would have to change because
   you swapped a library, it belongs in the plan. Send it back if it leaks.
2. **Are the user stories independently testable?** Each should be a slice you
   could build alone and demonstrate. The "Independent Test" line is the one that
   gets skipped, and it is what makes a story schedulable.
3. **Is every requirement numbered and singular?** `FR-012` asserts one thing. If
   it contains "and", it is two requirements and a test will cover half of it.
   Numbers are never reused: a withdrawn FR-014 stays withdrawn.
4. **Are the success criteria measurable?** "Fast" is not. "A configuration of 3
   workspaces, 10 projects and 100 repositories loads and validates in under
   100 ms" is. The test: two people given the shipped system reach the same
   verdict.
5. **Are the edge cases real?** Generic filler — "invalid input", "network
   error" — means the agent produced a list rather than reasoning about your
   domain. Real ones name a grazhda condition and its outcome: an empty config
   file loads as a configuration with no workspaces; no workspace marked
   `default: true` and none named `default` fails with a message naming `--name`;
   a leading `~` expands and a `~` elsewhere does not.

**Hand-editing is not cheating.** You own this document.

**Do.**

```bash
${EDITOR:-vi} specs/001-config-core/spec.md
grep -c '^- \*\*FR-' specs/001-config-core/spec.md
grep -n 'and' specs/001-config-core/spec.md | grep 'FR-'   # candidates to split
grep -n -A20 '## Assumptions' specs/001-config-core/spec.md # the guesses
```

**Check.** No implementation detail; every `FR-###` singular; every `SC-###` a
number; the Assumptions section read line by line and corrected.

---

### Step 8.4 — Commit the specification, alone

**Why alone.** A specification change and a source change must never land in the
same commit — that is what makes `git diff -- specs/` mean something, and the
Step 3.7 hook enforces it. The type prefix `spec` and the `(001)` feature number
are the convention from `PLAN.md` Appendix D.7; once spec, test and commit all
carry the same identifier, `git log --grep=FR-027` answers "why is this like
this?" without relying on memory.

**Do.**

```bash
git add specs/001-config-core/
git commit -m "spec(001): configuration core

Adds FR-001..FR-0NN and SC-001..SC-00N covering load, validate, query and
safe update of the workspace configuration. No code."
```

---

### Step 8.5 — Session 2: clarify

**What.** Run `/speckit-clarify` and answer its questions.

**Why.** Nominally optional; treat as mandatory. It scans a taxonomy — functional
scope, domain and data model, interaction flow, non-functional attributes,
integrations — marks each Clear, Partial or Missing, then asks **at most five**
questions, one at a time, writing each answer into the spec as it goes.

This is you doing the work only you can do. Every `NEEDS CLARIFICATION` marker
must be gone before planning starts; a marker that survives into implementation
becomes an assumption the agent makes silently.

**Do.**

```
/clear
/speckit-clarify
```

Then:

```bash
grep -c 'NEEDS CLARIFICATION' specs/001-config-core/spec.md    # must be 0
git diff -- specs/                                             # read what it changed
git add specs/001-config-core/
git commit -m "spec(001): answer clarification questions"
```

**Check.** Zero markers, and the coverage summary at the end of the session
classified every category as Resolved, Deferred or Clear.

---

### Step 8.6 — Session 3: plan

**What.** Run `/speckit-plan`, then read the Constitution Check **first**.

**Why.** Now technology is allowed to appear. The command produces `plan.md` and
usually `research.md`, `data-model.md`, `contracts/` and `quickstart.md`. It
evaluates the Constitution Check, resolves unknowns in a research phase, produces
design artefacts, and then — the part people drop — **re-evaluates the
Constitution Check against the design it just produced**. A design that passed as
a sketch can violate the constitution once it has concrete file paths.

This is a real gate, not decoration. A passing check is short and specific:
each principle, and the thing in *this* plan that satisfies it. "Complies with all
principles" is a wave, not a check. When something genuinely does violate a
principle, it goes in the Complexity Tracking table with three parts — what is
bought, and *which simpler alternative was rejected and why*. A justification you
would be embarrassed to read back in six months is the plan telling you to change
the plan.

If you wave one through, the constitution is dead from that moment on.

**Do.**

```
/clear
/speckit-plan
```

Then:

```bash
sed -n '/## Constitution Check/,/^## /p' specs/001-config-core/plan.md
sed -n '/## Complexity Tracking/,/^## /p' specs/001-config-core/plan.md
${EDITOR:-vi} specs/001-config-core/plan.md
git add specs/001-config-core/
git commit -m "plan(001): configuration core

Constitution Check passes. Load, validate, query and update live in
libs/config/; cli/ only parses flags and renders."
```

**Check.** Every principle has a row naming something concrete in this plan.

---

### Step 8.7 — Session 4: tasks

**What.** Run `/speckit-tasks`, then skim for guesses.

**Why.** The output is grouped **by user story** rather than by layer, so each
group is independently completable — setup, then foundational work that must
precede any story, then one phase per story in priority order, then polish. Task
lines carry an ID, an optional `[P]` parallel marker, a story tag and an exact
file path.

**One default will surprise you:** the tasks template says tests are OPTIONAL and
included only if the spec explicitly asks. Your constitution's Principle IV
overrides that — every functional requirement needs a test that fails before the
implementation and passes after. The constitution outranks the template; that is
exactly the sort of conflict it exists to settle. If the generated `tasks.md` has
no test tasks, say so and have them added.

Skim for **tasks with no file path** and **tasks that map to no requirement**.
Those are where the agent is guessing.

**Do.**

```
/clear
/speckit-tasks
```

Then:

```bash
grep -c '^- \[ \]' specs/001-config-core/tasks.md
grep -n '_test' specs/001-config-core/tasks.md | head        # test tasks exist?
grep -nE '^- \[ \].*(T[0-9]{3})' specs/001-config-core/tasks.md | grep -v '\.go' | head
git add specs/001-config-core/
git commit -m "tasks(001): configuration core"
```

**Check.** Every task has a file path; every `FR-###` appears somewhere.

---

### Step 8.8 — Optional session: analyze

**What.** Run `/speckit-analyze` for a coverage report.

**Why.** It consumes all three artefacts plus the constitution and produces a
report in the conversation and **no files**. The valuable part is the coverage
table — Requirement Key, Has Task?, Task IDs, Notes — a direct answer to "does
every `FR-###` have work assigned", plus an "Unmapped Tasks" list answering the
inverse. Severities run CRITICAL (violates a constitutional MUST, or a
requirement with zero coverage) down to LOW.

Skip it only when the feature is small enough that you have read all three files
yourself. On feature 001, run it.

**Do.**

```
/clear
/speckit-analyze
```

**Check.** No CRITICAL findings. Any requirement with zero task coverage is fixed
in `tasks.md` before implementation, not during it.

---

### Step 8.9 — Session 5: implement

**What.** Run `/speckit-implement`, in plan mode first for a feature this size.

**Why plan mode first.** You get the intended approach — which files, in what
order, satisfying which requirements — before anything on disk changes. Reading
that against `tasks.md` takes two minutes and catches the misunderstandings that
are expensive to unpick after twelve files exist. The artefact to review before
approving is the plan, not the diff.

`/speckit-implement` runs the checklist gate first (it scans every file under
`checklists/`, counts checked and unchecked items, and stops to ask if anything
is unchecked), then executes tasks in dependency order respecting `[P]` markers
and phase boundaries.

**Do.**

```
/clear
@specs/001-config-core/spec.md
@specs/001-config-core/tasks.md
/speckit-implement
```

**Check.** Nothing yet — the next step is the check.

---

### Step 8.10 — Verify before reading a line of code

**What.** Three commands, in this order, before any code review.

**Why.** A solo developer has no second reviewer, so `git` has to do part of that
job.

- `just build && just test` — because reviewing code that does not compile is
  wasted effort.
- `git diff --stat` — **shape before content**. A feature that touches forty
  files when `tasks.md` named twelve is worth stopping over before you read any
  of the forty.
- `git diff -- specs/` — **the single highest-value review action in the
  method**, and it takes ten seconds because it is almost always empty. If it is
  *not* empty during an implementation step, that is the failure mode the whole
  method guards against happening in front of you: an agent editing a
  specification so it matches code it has already written. Read it before
  anything else.

**Do.**

```bash
just build && just test
git diff --stat
git diff -- specs/          # must be empty
```

Then trace requirements to tests:

```bash
grep -o 'FR-[0-9]\{3\}' specs/001-config-core/spec.md | sort -u > /tmp/frs.txt
grep -rho 'FR-[0-9]\{3\}' libs/ cli/ --include='*_test.go' | sort -u > /tmp/tested.txt
comm -23 /tmp/frs.txt /tmp/tested.txt        # requirements with no test naming them
```

And check the layering rules mechanically:

```bash
grep -rn 'spf13/cobra' libs/            # must print nothing
grep -rn 'os/exec'     libs/            # only inside the executor implementation
grep -rn 'os.Getenv("HOME")\|os.UserHomeDir' libs/ --include='*_test.go'   # nothing
go test ./... -race -C libs             # the race detector, from the beginning
```

**Check.** Build green, tests green, `git diff -- specs/` empty, every `FR-###`
named by at least one test, no forbidden import in `libs/`.

**Why `-race` from the beginning:** a previous exploratory build of this project
found two genuine concurrency bugs in the v1 code that would otherwise have been
ported verbatim — a data race in the package manager's spinner `Stop`, and one in
the daemon server's shutdown path.

---

### Step 8.11 — 🧠 Review the code against the spec

**What.** Read the diff requirement by requirement.

**Why.** Review against the specification, not against your taste. If you find
yourself objecting to something the spec permits, **the spec is what needs
changing** — change it deliberately in its own commit and re-run, rather than
negotiating in chat.

Useful questions: which `FR-###` is this chunk? Does any test touch the real
filesystem outside `t.TempDir()`, or run a real `git`? Does the domain import
anything it should not?

**If the diff is too large to review, do not read faster.** Revert the
implementation commit, reduce `tasks.md` to a single user story, and re-run
against that narrower scope. An unreviewable diff is a process failure, not a
review failure.

**Do.**

```bash
git diff | less
git add -A
git commit -m "feat(001): configuration core

Implements FR-001..FR-0NN of specs/001-config-core/spec.md. Tests in
libs/config/*_test.go name each FR; SC-001 measured against a 100-repository
fixture."
```

---

### Step 8.12 — Session 6: reconcile

**What.** Update the spec to describe what actually shipped, then tick the parity
matrix.

**Why — this is the step everyone skips.** The implementation always teaches you
something the spec got wrong. A repository whose specifications lie is worse than
one with no specifications, because people trust it.

The legitimate path is to amend the spec deliberately, in its own commit, with a
message saying what was wrong and why the change is correct now. The illegitimate
path is editing it quietly so it matches whatever the code already does, with no
moment where you decided the new behaviour was right. The difference is not
cosmetic: the first produces a specification you can still trust; the second
produces one that lies.

A reconciliation commit saying only "update spec" is how a specification quietly
becomes fiction. One naming the requirement and the reason is a record you can
still trust a year later.

**Do.**

```
/clear
```

```bash
${EDITOR:-vi} specs/001-config-core/spec.md      # Status: Draft -> Implemented
${EDITOR:-vi} docs/parity-matrix.md              # tick the 001 rows
git add specs/ docs/
git commit -m "docs(001): reconcile spec against shipped implementation

FR-0NN corrected: <what was wrong and why the new text is right>.
FR-0NN added for <behaviour implementation revealed was never specified>.
Status Draft -> Implemented; the 001 rows in the parity matrix ticked for
spec, domain impl, CLI impl and test."
```

**Check.** `just test` green; every `FR-###` traceable to a test; parity matrix
updated; spec status `Implemented`.

---

### Step 8.13 — Merge and tag

**What.** Merge the feature branch and tag the proven loop.

**Why tag here.** A tag is a promise that the thing at that commit was in a state
worth naming. The end of Phase 8 qualifies: it is the point at which the *loop
itself* is proven, not just one feature. Do not tag mid-feature.

**Do.**

```bash
git log --oneline -- specs/001-config-core   # spec, clarify, plan, tasks, feat, reconcile
git checkout main && git merge --no-ff 001-config-core
git tag -a v0.1.0 -m "the loop is proven: feature 001, configuration core"
```

**Check.** `git log --oneline` shows the six-artefact sequence. A phase that
skipped one — reconciliation is the one everyone skips — shows up here as a gap
rather than as something you have to remember to check.

---

## The Loop, as a reusable procedure

Phases 9–14 are this same sequence six more times, so it is written once here.
Whenever a later phase says **"run the Loop"**, it means:

| # | Session | Command | What you do afterwards |
| :-- | :--- | :--- | :--- |
| L1 | — | `git checkout -b NNN-short-name` then `create-new-feature.sh --json --short-name <name> --number N "<title>"` | confirm `.specify/feature.json` |
| L2 | specify | `/clear`, `/speckit-specify <description>` | read against the five checks in Step 8.3; hand-edit; commit `spec(NNN)` |
| L3 | clarify | `/clear`, `/speckit-clarify` | answer every question; `grep -c 'NEEDS CLARIFICATION'` is 0; commit |
| L4 | plan | `/clear`, `/speckit-plan` | read the Constitution Check **first**; commit `plan(NNN)` |
| L5 | tasks | `/clear`, `/speckit-tasks` | skim for tasks with no file path or no requirement; commit `tasks(NNN)` |
| L6 | analyze | `/clear`, `/speckit-analyze` | fix any zero-coverage requirement in `tasks.md`, not in code |
| L7 | implement | `/clear`, `@spec.md`, `@tasks.md`, `/speckit-implement` | plan mode first on a big feature |
| L8 | — | `just build && just test`, `git diff --stat`, `git diff -- specs/` | in that order, before reading code |
| L9 | — | trace `FR-###` to tests; check layering greps; `go test -race` | see Step 8.10 |
| L10 | — | read the diff requirement by requirement | commit `feat(NNN)` |
| L11 | reconcile | `/clear` | amend the spec to what shipped; tick the parity matrix; commit `docs(NNN)` |
| L12 | — | `git checkout main && git merge --no-ff NNN-…` | tag if the phase is a milestone |

**Both halves, every time.** Each of phases 9–14 delivers the domain code in
`libs/` **and** the command surface in `cli/` that exposes it. Do not defer the
CLI to a later pass. A feature whose domain is written but whose command does not
exist cannot be exercised by hand, and a feature you cannot exercise by hand is a
feature whose specification you have not really tested. This is not hypothetical:
a full exploratory build of this project produced a complete domain and daemon
and no CLI source at all.

---

# Phase 9 — Feature 002, workspace targeting

**Goal:** `specs/002-targeting/` complete; the targeting and destination-path
rules in `libs/workspace`; the shared targeting flags mounted on the `zgard ws`
parent command.

**Why now.** Every command in features 003, 004 and 006 inherits this flag set.
Specify it once here and the later features are about what they *do* rather than
which repositories they do it to. Get it wrong and you re-litigate the same six
flags five more times.

---

### Step 9.1 — Create the feature

```bash
cd /mnt/e/proj/grazhda-spec
git checkout main && git checkout -b 002-targeting
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name targeting --number 2 "workspace targeting"
cat .specify/feature.json
```

---

### Step 9.2 — Run the Loop, L2 first

```
/clear
/speckit-specify workspace targeting: how every ws command decides which
repositories it acts on. Workspace selection by name, by --all, or by falling
back to the default. Filtering by project name, repository name and tag.
Cross-flag constraints checked before any work begins. Destination path
resolution from the configured structure mode. Targeting is pure: it validates
configuration structure only and never touches the filesystem.
```

---

### Step 9.3 — 🧠 What this specification must pin down

Check the drafted `spec.md` against every line of this list. A missing item is a
requirement you will discover as a bug five features later.

**Workspace selection.** `--name`/`-n` selects one; `--all` selects every one;
with neither, the default workspace is used **and an informational line naming
the resolved path is printed to stderr** — stderr specifically, so it never
contaminates piped output. The default is the first workspace marked
`default: true`, failing that the first named `default`, and when neither exists
the error must name `--name` as the way out.

**Cross-flag constraints, checked before any work begins.** `--name` and `--all`
are mutually exclusive; `--repo-name` requires `--project-name`; `--all` combines
with neither `--project-name` nor `--repo-name`. Each violation exits immediately
with its own message.

**Filtering.** `--project-name` is an exact match. `--repo-name` is a
case-sensitive **substring** match against the full configured repository name,
regardless of structure mode — so `cool` matches `ORG/PACK/my-cool-backend`, and
**the spec must say so with that example**, because "matches a repository" is
read as exact by every reader. `--tag` is repeatable with OR semantics, and
project-level tags are inherited by every repository in the project, merged with
the repository's own, deduplicated, order preserving.

**The four outcomes of a filter** — this is where a loose spec does real damage.
Matching several repositories is legal and produces a warning before work starts.
Matching none is an error, but the spec must say *which* none, because the
conditions differ:

1. a named project that does not exist;
2. a repository filter that matched nothing — which only reports as such when a
   project name was also given;
3. a tag filter that matched nothing within a repository-filtered set;
4. a tag filter that matched nothing in the workspace at all.

Four branches, four messages. Write them out. "An unmatched filter is an error"
is not a specified requirement; it is four requirements wearing one coat.

**Destination resolution.** An explicit `local_dir_name` always wins. Otherwise
`tree` keeps the full configured name so slashes become nested directories, and
`list` uses the last `/`-delimited segment with a trailing `.git` stripped. Two
repositories colliding on the same last segment in `list` mode are resolved by
the engineer with `local_dir_name`; without it the second is reported as already
existing. Structure precedence is project, then workspace, then `tree`.

**Path expansion.** A leading `~` expands to the home directory. A `~` anywhere
else does not.

---

### Step 9.4 — What to watch in review (L10)

Two things specifically.

**Did the four no-match branches survive as four distinct errors,** or did they
get collapsed into one generic "no matches"? Collapsing them is the single most
likely simplification and it makes the tool much worse to use.

```bash
grep -rn 'no match\|not found\|no repositories' libs/workspace/ | wc -l   # expect 4 distinct messages
```

**Did targeting stay pure?** It validates configuration structure only and must
never touch the filesystem. A repository that is merely not cloned yet is not a
targeting error.

```bash
grep -rn 'os.Stat\|os.Open\|filepath.Walk' libs/workspace/target*.go   # expect nothing
```

---

### Step 9.5 — Check

```bash
just build && just test
go test ./... -race -C libs

# by hand, against a real config:
./bin/zgard ws list -t backend            # narrows correctly
./bin/zgard ws list -t nosuchtag; echo $?  # exits 1, message names the filter
./bin/zgard ws list -r api;      echo $?   # refuses: -r requires -p, before doing anything
```

Table-driven tests cover every flag combination and all four no-match branches.

Then L11 (reconcile) and L12 (merge).

---

# Phase 10 — Feature 003, workspace lifecycle

**Goal:** `specs/003-lifecycle/` complete; `ws init`, `ws pull` and `ws purge`
working end to end; and — the part that outlives this feature — the **reporting
contract** and the **execution seam** every later command uses.

**Why now.** This is the first feature that touches the disk, and it establishes
two things four later features depend on: how progress is reported, and how the
domain shells out. Both are much cheaper to get right once than to change after
seven commands inherit them.

---

### Step 10.1 — Create the feature

```bash
git checkout main && git checkout -b 003-lifecycle
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name lifecycle --number 3 "workspace lifecycle"
```

---

### Step 10.2 — 🧠 What this specification must pin down

**Cloning.** Project directories are created **before** any cloning starts, so
concurrent workers never race on directory creation. A repository whose
destination already exists is skipped, not re-cloned. The clone command comes from
rendering the workspace template. **A clone that fails leaves no partial directory
behind** — the destination is removed. `--clone-delay-seconds` pauses after each
successful clone, in sequential and parallel mode alike.

**Pulling.** A rebase pull against the repository's effective branch. Repositories
not present on disk are skipped rather than failing.

**Purging — the safety contract,** and state it as a principle rather than a
behaviour: purge **never** falls back to the default workspace. It requires
`--name` or `--all` explicitly, confirms unless told not to, reports a skip when
the directory is already absent, and records a failure rather than aborting the
run.

**Dry run.** A banner, a "would …" line per repository including the rendered
clone command, and a summary line prefixed to match. The requirement that makes
this testable: a dry run performs **zero** command executions and **zero**
filesystem writes. Say that explicitly, because it is what the test asserts.

**Parallel.** One worker pool across every project, not one per project, with a
running done-of-total counter.

**Reporting.** Per-repository status lines with the fixed symbols (`✓` ok, `✗`
failed, `⚠` drifted, `⏭` skipped), optional captured output indented beneath, a
closing summary counting successes, skips and failures **using the operation's
own verb**, failure details on stderr, exit 0 when nothing failed and 1 when
anything did. Multi-target operations continue past a failure and report them all
at the end.

**The message strings themselves are part of the contract** — `already exists,
skipped`, `not present, skipped`, `cloned (main)`. Put them in the spec. They are
what the tests assert and what the parity matrix checks.

**Cancellation.** An interrupt stops between repositories and reports
cancellation rather than continuing.

**The execution seam.** Everything that shells out goes through one interface
with one real implementation and one deterministic fake. The spec states the
behaviour — tests never invoke real `git`, real network or a real home directory —
and the **plan** chooses the mechanism.

---

### Step 10.3 — What to watch in review, in order

1. **Does the dry-run test assert zero executor calls,** or does it just check
   the output looks right? Only the first is a test of the requirement.
2. **Does a failed clone actually remove the directory,** and is there a test that
   proves it?
3. **Does the parallel path visit every repository exactly once?** A
   `sync.WaitGroup` misuse here produces a test that passes and a tool that
   silently skips repositories under load.

```bash
grep -rn 'Calls\|calls' libs/workspace/*dry*_test.go | head    # the zero-call assertion
grep -rn 'RemoveAll' libs/workspace/                            # the failed-clone cleanup
go test ./... -race -C libs
```

---

### Step 10.4 — Check

```bash
just build && just test

# a dry run over a workspace where some repositories exist and some do not:
./bin/zgard ws init -n <workspace> --dry-run     # correct "would" lines, touches nothing
./bin/zgard ws purge; echo $?                    # exits non-zero WITHOUT prompting
```

A deliberately failing clone leaves no directory behind.

**Pitfall.** Letting the reporting contract emerge from the implementation rather
than specifying it. If you do, features 004 and 006 will each copy whatever
accident got written here, and by the time you notice, seven commands disagree
about what a summary line looks like.

---

# Phase 11 — Feature 004, cross-repository operations

**Goal:** `specs/004-crossrepo/` complete; `ws exec`, `stash`, `checkout`,
`search`, `diff`, `stats` and `list`, plus the shared table renderer.

**Why now.** Every one of these is an addition on top of the fan-out and
reporting from 003. None needs anything the previous features did not already
provide, which makes this the phase where you feel the investment pay off.

---

### Step 11.1 — Create the feature

```bash
git checkout main && git checkout -b 004-crossrepo
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name crossrepo --number 4 "cross-repository operations"
```

---

### Step 11.2 — 🧠 What this specification must pin down, per command

**`exec`** runs an arbitrary command in each resolved repository, captures its
output, and prints it indented beneath that repository's status line.
Not-present repositories are skipped. Failures do not stop the run.

**`stash`** and **`checkout`** are thin: a stash push, and a checkout of a branch
taken as a required positional argument.

**`search`** has three modes and the spec must state their **precedence** rather
than leaving it to the implementation: substring content search by default,
filename glob with `--glob`, regular expression with `--regex`, and **glob wins
when both are given**. Version-control metadata directories are skipped
entirely. Binary files are skipped, **detected by a null byte in the first 512
bytes** — state the rule, not "skips binary files". Output is
`[project/repo] file:line: content` for content matches and `[project/repo] file`
for filename matches, closing with `N match(es) across M repo(s)`. An invalid
regular expression fails **before any repository is walked**.

**`diff`** renders a per-project table with columns REPO, UNCOMMITTED, AHEAD,
BEHIND. **Two distinct placeholders**, and this is the detail most likely to be
got wrong: a repository that is **not cloned** renders `(not cloned)` in all
three value columns, while a **cloned repository with no upstream** renders `--`
in AHEAD and BEHIND only. Rows are coloured by state; the table closes with
counts of clean, dirty and not-cloned. The old project's documentation gets this
wrong and its code gets it right — decide deliberately and record the decision.

**`stats`** renders REPO, LAST COMMIT, 30D COMMITS, CONTRIBUTORS. Last commit is
normalised to `YYYY-MM-DD HH:MM`. An empty history renders `--`; a repository not
on disk renders `(not cloned)` with `-` in the numeric columns. Contributors are
unique author addresses compared case-insensitively after trimming.

**`list`** is a tree view grouped by project, each repository with its resolved
local path and whether it exists, closing with a cloned/not-cloned summary.

**The shared table renderer:** column widths sized to the widest cell **including
the header**, a bold header row, a rule beneath it, two spaces between columns, a
consistent indent, and correct output when colour is disabled.

**Ordering.** Output preserves the order declared in the configuration even when
the work was collected concurrently. This is a requirement, not an implementation
detail — say it in the spec, because it constrains the design.

---

### Step 11.3 — Check

```bash
just build && just test

# byte-identical output, parallel vs sequential:
./bin/zgard ws diff -n <workspace> > /tmp/seq.txt
./bin/zgard ws diff -n <workspace> --parallel > /tmp/par.txt
diff /tmp/seq.txt /tmp/par.txt && echo "order preserved"

# a search over a tree containing a binary file finds nothing in it, and does not error:
./bin/zgard ws search somestring; echo $?

NO_COLOR=1 ./bin/zgard ws diff -n <workspace>   # table still aligns
```

Golden-output tests exist for both tables.

**Pitfall.** Specifying "show a table of repository state" and moving on. Every
implementation of that sentence differs, and none is wrong, which means you have
no way to tell whether the one you got is the one you wanted.

---

# Phase 12 — Feature 005, the declarative package registry

**Goal:** `specs/005-pkgreg/` complete; `zgard pkg install`, `purge`, `register`
and `unregister`; the dependency resolver; the generated shell environment file.

**Why now.** It depends only on configuration and paths, so it can slot in
anywhere after 001 — but it must come **before 007**, because the installer seeds
the global registry file this feature reads. If you want a change of scenery
after three workspace features, this is the one you can move earlier.

---

### Step 12.1 — Create the feature

```bash
git checkout main && git checkout -b 005-pkgreg
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name pkgreg --number 5 "declarative package registry"
```

---

### Step 12.2 — 🧠 What this specification must pin down

**Two registry layers.** A distribution-managed global registry, replaced on
every upgrade, and a user-managed local one that survives upgrades. A missing
local registry behaves as an **empty** one, not as an error.

**Merge semantics** — exact, and easy to describe wrongly. Identity is the pair
(name, version). A local entry with the same identity as a global one **replaces**
it. A local entry with the *same name but a different version* is an
**additional, separate entry** — not a replacement. State both halves.

**Dependency references.** A reference is a bare name or `name@version`. A bare
name resolves to the sole package with that name; when several versions exist it
resolves to the unversioned one; and when that is ambiguous it is an error whose
message tells the engineer to qualify the reference. A versioned reference that
no entry satisfies is an error naming the version.

**Resolution.** Transitive dependencies are expanded automatically. Installation
order is a **deterministic** topological order — deterministic matters, because a
resolver that returns a different valid order each run makes every failure
irreproducible. A cycle is detected and reported naming the packages involved.
Purge order is the exact reverse.

**The install lifecycle, in this order, per package:** create the package
directory if requested; write the `<name>:pre` environment block and **re-source
the environment file** so the install script sees it; run the install script with
the environment file sourced first; write the `<name>:post` block and re-source
again so later packages see it. Each phase script runs with a documented set of
variables: the project root, the package directory, the package name, the version.

**The purge lifecycle:** the optional purge script, then removal of the package
directory if it was pre-created, then excision of both environment blocks.

**The environment file** is a set of named, lexically delimited, **idempotent**
blocks. Writing the same block twice replaces it in place rather than appending.
Removing a block leaves no doubled or leading blank lines. Removing a block that
is not there, and operating on a file that does not exist, are both no-ops.

**Presentation.** Script output is suppressed behind a progress indicator by
default; `--verbose` streams it instead.

**Registration is interactive:** every field prompted, dependencies chosen by
number from the merged package list, multi-line fields terminated by a blank
line. **Unregistration** removes by name, by exact name and version, or all local
entries — and does **not** uninstall anything. Say that explicitly; it is the
thing a user will assume wrongly.

---

### Step 12.3 — What to watch in review

The merge semantics and the phase ordering, because both read as obvious and are
not. And: **every test's phase scripts must be trivial local shell** — `echo`,
`true`, `exit 1`. A test that installs a real SDK is not a test.

```bash
grep -rn 'curl\|wget\|https://' libs/pkgman/*_test.go     # expect nothing
```

---

### Step 12.4 — Check

```bash
just build && just test
```

Specifically: a cycle is detected and named; a bare reference with two versions
available produces the qualify-it error; the same environment block written twice
produces a byte-identical file; a purge of a package whose directory was never
created succeeds.

**Pitfall.** Letting the resolver be non-deterministic. It will pass its tests
and produce install failures you cannot reproduce.

---

# Phase 13 — Feature 006, the health daemon

**Goal:** `specs/006-daemon/` complete; `dukh start`, `stop`, `status` and
`scan`; the monitoring loop; the gRPC service; and `zgard ws status` reading from
it.

**Why now.** It needs configuration from 001 and destination resolution from 002,
because **it must locate every repository exactly as `ws init` would**. If the two
disagree, the daemon reports repositories as missing that are sitting on disk.

---

### Step 13.1 — Create the feature

```bash
git checkout main && git checkout -b 006-daemon
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name daemon --number 6 "health daemon"
```

---

### Step 13.2 — 🧠 What this specification must pin down

**Starting.** The start command returns control to the engineer without leaving a
foreground process, and prints the new process id. Specify that as *behaviour*;
leave the mechanism to the plan.

**Process identity.** A file under the project's `run/` directory recording the
running process, written at startup and removed on clean shutdown. A stale file
whose process is gone is **cleaned up by the status command** rather than reported
as a running daemon.

**The three contract operations** (already fixed in `contracts/dukh.proto` at
Phase 7): report the health snapshot, optionally scoped to one workspace and
optionally forcing a fresh scan first and waiting for it; trigger a rescan without
waiting; request graceful shutdown. The payload: per workspace, per project, per
repository — name, resolved path, configured branch, actual branch, whether it
exists, whether the branch is aligned — plus the server's version and uptime in
seconds.

**The monitoring loop.** A scan at startup, then a scan every configured period.
**The period is re-read from configuration before each wait**, so changing it
takes effect at the next tick without a restart. A documented default applies when
it is absent or non-positive.

**Rescan requests.** Fire-and-forget requests are **dropped when one is already
queued**. A synchronous request waits for that scan to finish and gives up if the
caller's own deadline passes first.

**Shutdown.** On request or on a termination signal: stop accepting work, finish
what is in flight, remove the identity file, exit.

**The report.** The workspace/project/repository tree with alignment symbols,
drift shown as `configured → actual`, a closing count of aligned, drifted and
missing, and a header line carrying the daemon's uptime.

**Auto-start.** When the engineer asks for health and the daemon is not running,
it is started for them, waited for until it answers, then queried — **unless
auto-start is disabled in configuration**, in which case the failure is explicit
rather than silent.

**Logging.** A rotating file under `logs/`, mirrored to standard error, bounded
in size and in number of compressed backups.

---

### Step 13.3 — What to watch in review: testability, above everything

A daemon that determines a repository's branch by running `git` directly is a
daemon you cannot test. The **specification** should require the behaviour; the
**plan** should introduce an injectable seam in front of it. If the plan does
not, send it back — this is the one feature where a missing seam makes the whole
thing untestable rather than merely awkward.

Second: **no test may bind a fixed port.** A test suite that fails when you happen
to have the daemon running is a test suite you will start ignoring.

```bash
grep -rn ':50051\|:8080\|net.Listen("tcp", "' daemon/ --include='*_test.go'   # expect nothing
grep -rn 'exec.Command' daemon/server/                                        # expect nothing
```

`net.Listen("tcp", "127.0.0.1:0")` — port zero — is the correct idiom: the kernel
picks a free port and the test reads it back from the listener.

---

### Step 13.4 — Check

```bash
just build && just test

# tests must pass with no git on the path at all:
env PATH=/usr/bin:/bin go test ./... -C daemon

./bin/dukh start && ./bin/zgard ws status && ./bin/dukh stop
```

A stale identity file is cleaned up. A cancelled synchronous rescan returns
promptly with a cancellation error. Changing the monitoring period in
configuration takes effect without restarting.

**Pitfall.** Specifying "reads the current branch of each repository" and
discovering at implementation time that this makes the feature untestable. The
specification is where you notice that, if you are reading it properly.

---

# Phase 14 — Feature 007, the toolchain manager

**Goal:** `specs/007-toolchain/` complete; `grazhda-install.sh`,
`grazhda-init.sh` and the `grazhda` management script; the Bash test harness.

**Why now.** Last, because it installs everything the other features produce. It
cannot be meaningfully tested until there are binaries to install.

---

### Step 14.1 — Create the feature

```bash
git checkout main && git checkout -b 007-toolchain
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name toolchain --number 7 "toolchain manager"
```

---

### Step 14.2 — 🧠 What this specification must pin down

**Install.** Create the project root; clone or update the sources into it; verify
every required build tool and, when any is missing, report **all** of them rather
than the first; build; place the binaries; seed the configuration file from the
template **only when one does not already exist**; seed the global package
registry; and wire the shell integration into the engineer's profile
**idempotently**. The installer refuses to run when sourced rather than executed,
and works when piped from a download.

**Shell integration.** Export the project root when unset; prepend the binary
directory to the search path **exactly once**; create the packages directory;
source the generated environment file when it is readable.

**Upgrade.** Refuse when the sources directory is not a repository. Verify build
tools. Update the sources and report whether anything changed. **Stop a running
daemon gracefully before replacing its binary**, escalating to a forced stop after
a bounded wait, and clean up a stale identity file. Rebuild, install, refresh the
global registry from the sources, and tell the engineer plainly that the daemon is
**not** restarted for them.

**Uninstall.** Confirm; stop the daemon; remove the shell-profile lines; delete
everything under the root **except the configuration file**, which is deliberately
preserved so a reinstall reuses it. **Purge** is the same but deletes the root
entirely, with an explicit warning that it is irreversible.

**Configuration editing.** An editor resolved in a documented order — the
configured value, then the two conventional environment variables, then a final
fallback — with the path quoted so spaces are safe, and a clear failure when no
configuration file exists.

**Output and flags.** A verbosity flag accepted **in any argument position**.
Quiet by default, but a failed step prints the output it captured, so a failure is
never silent. Consistent information, success, warning and error lines. A usage
command. An unknown command produces an error, the usage text, and a non-zero
exit.

---

### Step 14.3 — What to watch in review: idempotency, everywhere

Install twice and the shell profile must be **byte-identical** to after the first
install. Uninstall and it must be byte-identical to before. This is the property
that makes the installer safe to run when you are not sure whether you already ran
it — which is the only state anyone is ever in.

---

### Step 14.4 — Add the Bash recipes to the Justfile

**What.** `test-bash`, `copy-scripts`, and `shellcheck` in `lint`.

**Why.** The shell scripts are now real components with real tests, so they join
the same entry point as everything else. `shellcheck` accepts multiple file
arguments on one line, so no `cd`-per-module is needed.

**Do.**

```bash
python3 - <<'PY'
import pathlib
p = pathlib.Path("Justfile"); s = p.read_text()
s = s.replace("""# Run every test suite
test: test-go""", """# Run the shell test harness
test-bash:
    bash tests/bash/test_scripts.sh

# Stage the toolchain manager scripts into bin/
copy-scripts:
    @mkdir -p bin
    cp scripts/grazhda scripts/grazhda-init.sh bin/
    chmod +x bin/grazhda bin/grazhda-init.sh
    @echo "✓ scripts staged in bin/"

# Run every test suite
test: test-go test-bash""", 1)
s = s.replace("""    cd daemon && go vet ./...""", """    cd daemon && go vet ./...
    shellcheck scripts/grazhda scripts/grazhda-init.sh scripts/grazhda-install.sh tests/bash/test_scripts.sh""", 1)
s = s.replace("build: build-cli build-daemon", "build: build-cli build-daemon copy-scripts", 1)
p.write_text(s)
PY
just --list
```

---

### Step 14.5 — Check

```bash
just lint                      # shellcheck clean
just test-bash
```

The harness runs install, install again, then uninstall, **all against a
temporary home directory**, and asserts the profile is unchanged at the end. A
missing build tool produces a list, not a single complaint.

**Pitfall, and it is the expensive one.** A test harness that touches your real
home directory or your real shell profile. Set `HOME` to a temporary directory in
the harness and never assume; the first time this goes wrong you will be restoring
your own `.bashrc` from memory.

```bash
grep -n 'HOME=' tests/bash/test_scripts.sh      # must set HOME to a temp dir, early
```

---

### Step 14.6 — Parity

**At the end of Phase 14 you have parity**: every capability of the old system,
specified, implemented, tested and traceable.

```bash
grep -c '⬜' docs/parity-matrix.md     # must be 0
grep -n '➖' docs/parity-matrix.md     # every declined row has a stated reason
git checkout main && git merge --no-ff 007-toolchain
git tag -a v0.7.0 -m "parity: every v1 capability specified, implemented and tested"
```

**Check.** `docs/parity-matrix.md` carries no empty cells. Everything from here is
new.

---

# Phase 15 — Feature 008, `molfar`, the long-term process

**Goal:** the automation service — it accepts a chore, plans it, runs it across
your repositories, records everything it did, and hands back a proposal you
approve, or do not.

**Effort:** 1–2 weeks.

**Why this is last, and separate.** It is the only genuinely new component, it
depends on everything before it, and its failure modes are qualitatively
different. A bug in `ws diff` prints a wrong number. A bug here pushes a commit
you did not authorise, or runs for six hours, or writes a credential into a log
file.

**Read `guides/04-spring-boot.md` first, and `IDEAS.md` again.** One thing from
the Spring guide before you write a line: **in Spring Boot 4 auto-configuration is
modular**, so putting a library such as `flyway-core` on the classpath gives you
the library with no Spring wiring and **no error at all** — total silence, no
migrations, no log line. Depend on the **starter**, always. *The starter is the
contract.*

---

### Step 15.1 — Create the feature and specify it, properly

```bash
git checkout main && git checkout -b 008-molfar
bash .specify/scripts/bash/create-new-feature.sh --json \
  --short-name molfar --number 8 "automation service"
```

**🧠 This is the one specification where thoroughness is not a matter of taste.**
The requirements that carry the most weight:

- Every run carries a **recorded budget** — wall clock, step count, files changed,
  and money — and stops when any is exhausted, in a terminal state that **names
  which one**.
- Nothing is pushed, merged, tagged or published without a **recorded approval**.
- Every executed step is **reconstructable from the record after a restart**: the
  command, the working directory, the exit status, the captured output, the
  timing, the files changed.
- **Credentials are redacted** on the way into every record, proposal and
  notification.
- A step that would act **outside the task's declared scope** is refused and fails
  the run.
- **Repository content may influence the *content* of a proposal but never the
  *policy*** — scope, budget, approval requirement. This is the prompt-injection
  defence, and it has to be structural, not advisory.

**Run `/speckit-clarify` twice on this one.** It is worth it.

```
/clear
/speckit-specify …
/clear
/speckit-clarify
/clear
/speckit-clarify
```

---

### Step 15.2 — Write the contract before any controller

**What.** `contracts/molfar-openapi.yaml`, first.

**Why.** Contract-first is a constitutional principle and the CLI's client is
generated from the same file. Deciding the resource shapes — chore, plan, run,
step, proposal, approval — *before* writing a controller forces you to settle what
the service actually exposes rather than letting it accrete one endpoint at a
time.

**Do.**

```bash
${EDITOR:-vi} contracts/molfar-openapi.yaml
git add contracts/ && git commit -m "spec(008): molfar HTTP contract, before any controller"
```

**Check.** Every resource in the spec's Key Entities section has a shape here.

---

### Step 15.3 — Generate the Spring Boot project

**What.** Create `service/` from start.spring.io.

**Why not write the `pom.xml` by hand,** even though you know Maven: the
generator is the authoritative source for which starter artifact ids exist at
your exact Boot version, and Boot 4 renamed several. Selecting the gRPC starter
also makes it add the protobuf Maven plugin for you, which is fiddly by hand.

**Do.**

```bash
cd /mnt/e/proj/grazhda-spec
curl -s https://start.spring.io/starter.zip \
  -d type=maven-project -d language=java \
  -d bootVersion=4.1.0 -d javaVersion=25 \
  -d groupId=com.grazhda -d artifactId=molfar -d name=molfar \
  -d packageName=com.grazhda.molfar -d packaging=jar \
  -d dependencies=web,validation,data-jpa,flyway,actuator,h2 \
  -o /tmp/molfar.zip
unzip -o /tmp/molfar.zip -d service/
ls service/
```

Then open `https://start.spring.io` in a browser to confirm the current
dependency ids and the **gRPC Client** starter, which the CLI form above may not
expose under a stable id. Add it by hand if needed:
`spring-boot-starter-grpc-client`, group `org.springframework.boot`, no version.

**Check.** `service/pom.xml` exists and its parent is
`spring-boot-starter-parent` at 4.1.x.

---

### Step 15.4 — Read the pom, element by element

**Why each part is there.**

- **The parent POM** does three jobs: imports `spring-boot-dependencies`, a BOM
  with managed versions for several hundred artifacts, which is why no dependency
  below carries a `<version>`; configures plugin defaults; and defines the
  `java.version` property, which it translates into `maven.compiler.release`.
  Setting `<java.version>25</java.version>` is therefore the *whole* of "target
  Java 25" — do not also configure `maven-compiler-plugin`.
- **The empty `<relativePath/>`** tells Maven not to look for the parent on disk.
  Without it Maven checks `../pom.xml`, and `service/` sits under a repository
  root with no parent POM, so that check is pure noise.
- **Six starters and a driver.** `webmvc` brings Spring MVC, Tomcat 11, Jackson 3
  and the web auto-configuration. `validation` is a **separate** starter — if
  `@Valid` appears to do nothing, this is what is missing. `data-jpa` brings
  Hibernate 7.1 and HikariCP. `flyway` is the starter-is-the-contract lesson made
  concrete. `actuator` for health and metrics; `grpc-client` for talking to
  `dukh`. The H2 driver is a plain library at `runtime` scope, which is correct.
- **`<finalName>molfar</finalName>`** is what makes the Justfile's
  `cp service/target/molfar.jar bin/molfar.jar` work without a version in the
  path.

**Do.**

```bash
less service/pom.xml
python3 - <<'PY'
import pathlib,re
p=pathlib.Path("service/pom.xml"); s=p.read_text()
if "<finalName>" not in s:
    s=s.replace("<build>","<build>\n\t\t<finalName>molfar</finalName>",1)
    p.write_text(s)
PY
grep -n 'finalName\|java.version\|artifactId>spring-boot-starter' service/pom.xml
```

**Check.** `cd service && mvn -q -B -DskipTests package && ls target/molfar.jar`

**One Boot 4.1 wrinkle:** `-DskipTests` no longer skips **AOT processing**. It
still skips *running* tests, but AOT work proceeds, so packaging may be slower
than expected and can fail on something unrelated. `-Dmaven.test.skip=true` skips
test compilation and execution both.

---

### Step 15.5 — Build the framework-free `core`

**What.** The task state machine, budget accounting, scope checking, plan
validation and credential redaction in a plain-Java `core` package with **no
framework dependencies**, unit-tested with plain JUnit and no Spring context.

**Why.** This is the same layering rule the Go side follows, and it matters more
here, not less: these are the rules that keep the system safe, and you want to
test them exhaustively in milliseconds rather than through a Spring context that
takes eight seconds to start.

`MolfarApplication` must live in the **root package**, `com.grazhda.molfar`,
because `@ComponentScan`'s base package is implicit and everything else must sit
beneath it. A class in a sibling package is invisible to the container, and
nothing tells you.

**Do.**

```bash
cd service/src/main/java/com/grazhda/molfar
mkdir -p core/{chore,plan,run,proposal,health,port} \
         adapter/{web,jpa,grpc,git,yaml} application scheduling config
cd /mnt/e/proj/grazhda-spec
find service/src/main/java -type d | sort
```

---

### Step 15.6 — Enforce the layering with an ArchUnit test

**What.** Two tests, no Spring context, sub-second.

**Why.** A rule that lives only in a document decays. This one fails in CI the
moment somebody adds `@Entity` to a domain class at the end of a long day, which
is worth more than any amount of documentation.

**Do.** Add the ArchUnit test dependency to `service/pom.xml`
(`com.tngtech.archunit:archunit-junit5`, test scope), then:

```bash
mkdir -p service/src/test/java/com/grazhda/molfar
cat > service/src/test/java/com/grazhda/molfar/LayeringTest.java <<'EOF'
package com.grazhda.molfar;

import com.tngtech.archunit.core.domain.JavaClasses;
import com.tngtech.archunit.core.importer.ClassFileImporter;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;

import static com.tngtech.archunit.lang.syntax.ArchRuleDefinition.noClasses;
import static com.tngtech.archunit.lang.syntax.ArchRuleDefinition.noFields;

class LayeringTest {

    private static final JavaClasses classes =
            new ClassFileImporter().importPackages("com.grazhda.molfar");

    @Test
    void coreDependsOnNothingButJavaAndItself() {
        noClasses().that().resideInAPackage("..core..")
                .should().dependOnClassesThat()
                .resideInAnyPackage("org.springframework..", "jakarta..", "tools.jackson..",
                                    "com.fasterxml..", "io.grpc..", "org.hibernate..")
                .check(classes);
    }

    @Test
    void noFieldInjection() {
        noFields().should().beAnnotatedWith(Autowired.class).check(classes);
    }
}
EOF
cd service && mvn -q -B test && cd ..
```

**Check.** Both tests pass. Deliberately add a `@Service` annotation to a `core`
class and confirm the first test **fails** — a guard rail you have not seen fail
is a guard rail you do not have.

---

### Step 15.7 — Persistence and the journal

**What.** Flyway migrations from the very first table.

**Why.** The journal is not a feature of this service; **it is the product**.
Design it as queryable dimensions rather than free text, because the questions you
will actually ask are "what did it try on this repository", "which proposals did I
reject and why", and "how long does this chore usually take" — none of which a
text blob answers.

Flyway from the start, even for the first table, because retrofitting migrations
onto a schema that was created by Hibernate's auto-DDL means reverse-engineering
your own database.

**Do.**

```bash
mkdir -p service/src/main/resources/db/migration
cat > service/src/main/resources/db/migration/V1__journal.sql <<'EOF'
-- The run journal. Queryable dimensions, not free text.
CREATE TABLE chore (
  id            VARCHAR(36)  PRIMARY KEY,
  submitted_at  TIMESTAMP    NOT NULL,
  intent        VARCHAR(512) NOT NULL,
  scope         VARCHAR(512) NOT NULL,
  status        VARCHAR(32)  NOT NULL
);
-- …run, step, proposal, approval follow, per the spec's Key Entities.
EOF
```

**Check.** Start the service and confirm Flyway logs the migration. **If there is
no log line at all, that is the modular auto-configuration trap** — you have
`flyway-core` rather than `spring-boot-starter-flyway`:

```bash
grep -n 'flyway' service/pom.xml     # must say spring-boot-starter-flyway
```

---

### Step 15.8 — Generate the Java gRPC client from the same `.proto`

**What.** Point molfar's protobuf plugin at `contracts/dukh.proto`.

**Why.** One contract, two generated clients, is much of the argument for
protobuf. The `option java_package` you set in Phase 7 exists precisely for this
moment.

The channel property in Spring gRPC is **singular**:
`spring.grpc.client.channel.<name>.target`. The plural `channels…address` form
was Spring gRPC 1.0 and is out of date — this will cost you an hour if you follow
an older tutorial.

**Do.**

```bash
grep -rn 'proto' service/pom.xml | head
${EDITOR:-vi} service/pom.xml          # point protoSourceRoot at ../contracts
cat >> service/src/main/resources/application.yaml <<'EOF'

spring:
  grpc:
    client:
      channel:
        dukh:
          target: "localhost:50051"
EOF
cd service && mvn -q -B compile && cd ..
```

**Check.** Generated Java appears under `service/target/generated-sources/`.

---

### Step 15.9 — The execution engine

**What.** Action kinds behind a registry, each declaring its rung on the
determinism ladder and carrying its own **verification oracle**.

**Why.** Two requirements from `IDEAS.md` belong in the code from day one rather
than retrofitted: a step with **no oracle** is marked unverified and surfaced
separately in the proposal — so the proposal says out loud "these four changes
were verified; these two could not be" — and a repository with **uncommitted
changes** is refused unless the task explicitly opted in.

---

### Step 15.10 — The API and the CLI bridge

**What.** The controllers, then `zgard agent submit`, `list`, `follow` and
`approve`.

**Why `follow` matters:** it exits with a status reflecting the run's own terminal
state, so it composes in a shell script.

---

### Step 15.11 — Add the Java recipes to the Justfile and CI

**Do.**

```bash
python3 - <<'PY'
import pathlib
p=pathlib.Path("Justfile"); s=p.read_text()
s=s.replace("""# Build every component into bin/
build: build-cli build-daemon copy-scripts""",
"""# Build the long-term process (molfar)
build-service:
    @mkdir -p bin
    cd service && mvn -q -B -DskipTests package
    cp service/target/molfar.jar bin/molfar.jar
    @echo "✓ bin/molfar.jar"

# Build every component into bin/
build: build-cli build-daemon build-service copy-scripts""",1)
s=s.replace("""# Run every test suite
test: test-go test-bash""",
"""# Run the molfar test suite
test-java:
    cd service && mvn -q -B test

# Run every test suite
test: test-go test-java test-bash""",1)
p.write_text(s)
PY

python3 - <<'PY'
import pathlib
p=pathlib.Path(".github/workflows/ci.yml"); s=p.read_text()
s=s.replace("""      - name: Install protoc""","""      - uses: actions/setup-java@v4
        with:
          distribution: "temurin"
          java-version: "25"
      - name: Install protoc""",1)
p.write_text(s)
PY
just --list && just build && just test
```

**Check.** `bin/` holds `zgard`, `dukh`, `molfar.jar` and the shell scripts.

---

### Step 15.12 — 🧠 Observe-only, then hands

**What.** Ship it in shadow mode: it receives events, plans, and writes what it
*would* have done into the journal, **executing nothing**. Run it against your
real work for a few weeks.

**Why.** Then read the journal and count how often the plan was the right one.
That number is the only honest answer to "is this ready to be given hands", and
there is no way to get it except by waiting. It also produces the first entries in
the **evaluation corpus** — frozen repository states plus tasks plus known-good
outcomes — which is what lets you tell whether tomorrow's change to the planner
helped, rather than making every planner change a matter of opinion.

Resist skipping this.

**Check.** A chore submitted from the command line produces a proposal you can
read; the journal survives a restart with the run intact; nothing has reached a
remote without your approval; and a run given an impossible budget stops and says
**which** budget it ran out of.

---

# Phase 16 — Close out

**Goal:** nothing left that only exists in your head.

**Effort:** 2–3 days.

---

### Step 16.1 — Walk the baseline command by command

**What.** Every command in `docs/v1-baseline/CLI.md`, run against the new
binaries.

**Why.** The parity matrix records what you believe; this step checks it. Every
row goes green or gets an explicit note saying why it will not.

**Do.**

```bash
just build
less docs/v1-baseline/CLI.md          # side by side with a terminal
./bin/zgard --help
diff <(./bin/zgard ws --help) <(sed -n '/ws --help/,/^====/p' docs/v1-baseline/help-output.txt)
${EDITOR:-vi} docs/parity-matrix.md
```

**Check.** `grep -c '⬜' docs/parity-matrix.md` is 0, and every `➖` carries a
reason. An honest "declined" row is still trustworthy; a silent gap is not.

---

### Step 16.2 — Re-read every specification against its implementation

**Do.**

```bash
for f in specs/*/spec.md; do echo "== $f"; grep -n 'Status:' "$f"; done
```

**Check.** Every status reads `Implemented` (or `Abandoned` with a one-line
reason). Any spec still `Draft` is either unfinished work or an unreconciled
feature.

---

### Step 16.3 — Generate man pages and shell completions

**Why.** Cobra generates both from the command tree you already built, so they
cost a recipe rather than a document — and a man page that is generated cannot
drift from the flags.

**Do.**

```bash
python3 - <<'PY'
import pathlib
p=pathlib.Path("Justfile"); s=p.read_text()
s+="""
# Generate man pages and shell completions
docs-gen:
    @mkdir -p man completions
    ./bin/zgard completion bash > completions/zgard.bash
    ./bin/zgard completion zsh  > completions/_zgard
    ./bin/zgard man --out man
    @echo "✓ man/ and completions/"
"""
p.write_text(s)
PY
just build && just docs-gen && ls man completions
```

**Check.** `man ./man/zgard.1` renders.

---

### Step 16.4 — Finish the release workflow

**Why.** A tag should produce binaries without a human building them, so the
thing people download is the thing CI tested.

**Do.**

```bash
${EDITOR:-vi} .github/workflows/release.yml    # on: push: tags: ['v*']; just build; upload bin/
git add -A && git commit -m "ci: release workflow — a tag produces binaries"
```

---

### Step 16.5 — Write the README for a stranger

**Why.** `README.md` currently describes *the plan*. At close-out it must describe
*the product*, for someone who has never seen either.

**Do.**

```bash
${EDITOR:-vi} README.md
```

**Check — the phase's real test.** A stranger can clone, run `just build`, and
use the tool from the README alone. Test it by doing exactly that in a clean
directory:

```bash
cd /tmp && rm -rf grazhda-check && git clone /mnt/e/proj/grazhda-spec grazhda-check
cd grazhda-check && just build && ./bin/zgard --help
```

Note that this clone has **no `protoc` requirement** — the generated code is
committed, which is what Phase 7 bought.

---

### Step 16.6 — Tag the release

```bash
cd /mnt/e/proj/grazhda-spec
just build && just test && just lint
git status --short                      # clean
git tag -a v1.0.0 -m "parity plus molfar in observe-only mode"
git push origin v1.0.0
```

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

---

# Appendix 1 — Every command, on one page

**Build and test**

```bash
just                 # help (bare just runs the first recipe, which is `default: help`)
just --list          # every recipe, from the comment above it
just --evaluate      # parse the Justfile without running anything
just build           # zgard, dukh, molfar.jar, staged scripts
just test            # go + java + bash suites
just lint            # go vet per module, shellcheck the scripts
just fmt             # gofmt per module
just generate        # regenerate daemon/proto from contracts/dukh.proto
just clean           # rm -rf bin
```

**Go, per module — never at the repository root**

```bash
cd libs && go test ./... -race       # the race detector; use it from the start
cd libs && go test ./... -run TestX -v
cd libs && go mod tidy               # only when the dependency set genuinely changed
GOWORK=off go build ./...            # does this module still build standalone?
go work use ./newmodule              # add a module to the workspace
```

**Spec-kit**

```bash
bash .specify/scripts/bash/create-new-feature.sh --json --short-name x --number N "title"
bash .specify/scripts/bash/check-prerequisites.sh --require-tasks --include-tasks
cat .specify/feature.json            # WHICH FEATURE IS ACTIVE — not the git branch
SPECIFY_FEATURE=003-lifecycle …      # override the pointer for one command
specify self check                   # is a newer spec-kit available
```

**The slash commands, in loop order**

```
/speckit-constitution   once, at project start
/speckit-specify        what and why, technology-neutral
/speckit-clarify        at most five questions; treat as mandatory
/speckit-plan           technology enters; READ THE CONSTITUTION CHECK FIRST
/speckit-tasks          grouped by user story, each with a file path
/speckit-analyze        coverage table; no files written
/speckit-checklist      a requirements-quality checklist for one named risk
/speckit-implement      runs the checklist gate, then the tasks
/speckit-converge       appends new tasks for the gap between spec and code
```

**Review, in this order, every time**

```bash
just build && just test
git diff --stat                      # shape before content
git diff -- specs/                   # must be empty during implementation
git log --oneline -- specs/001-…     # spec, clarify, plan, tasks, feat, reconcile
```

**Requirement tracing**

```bash
grep -o 'FR-[0-9]\{3\}' specs/NNN-*/spec.md | sort -u > /tmp/frs.txt
grep -rho 'FR-[0-9]\{3\}' libs/ cli/ daemon/ --include='*_test.go' | sort -u > /tmp/tested.txt
comm -23 /tmp/frs.txt /tmp/tested.txt        # requirements no test names
git log --grep=FR-027                        # why is this like this?
```

**Layering, checked mechanically**

```bash
grep -rn 'spf13/cobra' libs/                 # nothing
grep -rn 'os/exec' libs/                     # only the executor implementation
grep -rn 'fmt.Println' libs/                 # nothing; take an io.Writer
cd service && mvn -q -B test -Dtest=LayeringTest
```

---

# Appendix 2 — When a step fails

Most failures are in `PLAN.md` Appendix B, which is a table of symptom → cause →
fix. The five you are most likely to hit, with the fix inline:

| Symptom | Fix |
| :--- | :--- |
| `cd` on one Justfile line has no effect on the next | Each recipe line is its own shell. Chain with `&&` on one line, or use a shebang recipe with `set -euo pipefail`. |
| `protoc-gen-go: program not found` | `$HOME/go/bin` is not on `PATH`. Step 1.3. |
| `go: cannot find module providing package …/libs` | Module not in `go.work`, or the `replace` is missing. `go work use ./libs`; Step 6.4. |
| CI fails on `git diff --exit-code -- daemon/proto` | Generated code is stale. `just generate` and commit. If it only fails *sometimes*, your `protoc` plugin versions differ from CI's — Step 1.5. |
| `molfar` starts but a feature does nothing, with no error and no log line | Boot 4's modular auto-configuration. You have a bare library where a starter belongs. |

Four situations `PLAN.md` Appendix F covers that no per-step check will catch:

**The build broke and you do not know which change did it.** If the commit
discipline held, the last few commits are small and individually labelled by
artefact. `git diff --stat HEAD~1`, then `git revert HEAD` to get back to green
while you work out why at leisure. Further back, `git bisect` earns its keep —
cheap precisely because the history is small single-purpose commits rather than
one squash.

**The specification turned out to be wrong after implementation.** Legitimate:
amend it deliberately, in its own commit, with a message saying what was wrong and
why the new text is right, then re-run implementation for whatever the change
affects. Illegitimate: editing it quietly to match the code. The first produces a
specification you can still trust; the second produces one that lies.

**The agent produced a diff you cannot review.** Do not push through by reading
faster. Revert the implementation commit, reduce `tasks.md` — or the plan — to a
single user story, and re-run against that narrower scope. An unreviewable diff is
a process failure, not a review failure.

**You have lost the thread after a break of weeks.** Read three things in this
order before touching anything: `docs/parity-matrix.md` (what is done, what is
left, what you were in the middle of), `git log --oneline` (the *actual* last
artefact produced, as opposed to what you remember), and the `spec.md` of whichever
feature the log says was in progress — fully, not skimmed, because after weeks away
you are effectively reviewing it for the first time again.

**And the general rule:** when the agent thrashes on a requirement, **suspect the
requirement**. An agent looping on a task usually means the requirement is
ambiguous, contradicts another one, or depends on something never specified. The
fix is upstream, in a document, not in a better prompt.

---

# Appendix 3 — Phase checklist

Tick a phase only when its Check passed, not when its steps were performed.

| Phase | Done when |
| :--: | :--- |
| 0 | `docs/product.md` describes the product in two sentences with no mention of AI |
| 1 | Eight tools print a version; `mvn -v` agrees Java is 25 |
| 2 | Every command in `CLI.md` is a parity-matrix row; every `GAPS.md` entry is Close or Carry |
| 3 | `.specify/templates/` lists five templates; the ten skills are **tracked**; the pre-commit hook refuses a mixed commit |
| 4 | Every principle answers "when would this stop me?"; no `[PLACEHOLDER]`; version 1.0.0 |
| 5 | A reader of `docs/architecture.md` alone could lay out the repository |
| 6 | `just build && ./bin/zgard --version && just test` |
| 7 | `just generate && just build` leaves `git status` clean |
| 8 | Six commits in the log; every `FR-###` traceable to a test; spec status `Implemented` |
| 9 | Four no-match branches, four distinct messages; targeting touches no filesystem |
| 10 | Dry run asserts **zero** executor calls; a failed clone leaves no directory; `ws purge` with no target exits non-zero without prompting |
| 11 | Parallel and sequential output byte-identical; both diff placeholders correct |
| 12 | A cycle is named; the same env block written twice is byte-identical |
| 13 | Tests pass with no `git` on `PATH` and no fixed port |
| 14 | `shellcheck` clean; install-install-uninstall leaves the profile byte-identical |
| 15 | A budget-exhausted run names **which** budget; the journal survives a restart |
| 16 | A stranger clones, runs `just build`, and uses the tool from the README alone |
