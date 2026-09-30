# A working introduction to `just`, for building grazhda

This is a from-zero guide to `just`, a command runner. You already know Java, Maven and shell, and you have never used `just`. By the end of this document you will understand not just the commands to type but why each one behaves the way it does, and you will have built up — piece by piece — the real Justfile that drives the build of **grazhda**, a project that ships four things from one repository:

- `zgard` — a CLI written in Go, built from `cli/`
- `dukh` — a background daemon written in Go, speaking gRPC, built from `daemon/`
- `molfar` — a long-running service written in Java 25 / Spring Boot 4.1, built with Maven from `service/`
- `grazhda` — a small Bash toolchain manager, living in `scripts/`

Shared Go code lives in `libs/`, and `cli/`, `daemon/`, and `libs/` are joined into one Go workspace by a `go.work` file at the repository root. The protobuf contracts that `dukh` speaks live in `contracts/` and get compiled into generated Go code under `daemon/proto/`. Everything the build produces lands in `bin/`.

We will use this real project throughout. Nothing here is invented for the sake of an example.

## 1. What problem a task runner solves

Once grazhda has all four components, "building the project" is not one command. It is several:

```
cd cli && go build -o ../bin/zgard .
cd daemon && go build -o ../bin/dukh ./cmd
cd service && mvn -q -B package
protoc --go_out=daemon/proto ... contracts/dukh.proto
bash tests/bash/test_scripts.sh
```

Five commands, three working directories, two toolchains (`go`, `mvn`), a code generator (`protoc`), and a hand-written test harness. Every developer, and every CI job, needs to run some subset of this correctly and in order. That is the actual problem: not "how do I run `go build`" — you already know that — but "how do I give a whole team, including future-you in six months, **one verb per intent** that always does the right thing, in the right place, regardless of which shell they are typing into."

The naive answer is a directory full of shell scripts: `build.sh`, `test.sh`, `clean.sh`. This falls over in familiar ways: scripts accumulate boilerplate (`set -euo pipefail`, argument parsing, usage text) unrelated to the actual task; there is no single place to see "what can I run here" short of `ls scripts/` and reading each file; adding a parameter means adding `getopts` parsing everywhere it is needed; and scripts calling other scripts (`build.sh` shelling out to `generate.sh`) tangle quickly.

The second naive answer, if you come from a C or systems background, is Make — built for exactly this "one verb, several steps" problem, and used for build orchestration for fifty years. It is also the wrong tool here, for reasons worth being precise about, which is section 2.

What we actually want sits between "a folder of scripts" and "a real build system": a single file, checked into the repository, listing named recipes with a description each, callable as `just build`, `just test`, `just fmt`, discoverable with one command, and boring enough that nobody needs to read its source before trusting it. That is `just`.

## 2. `just` versus `make` — the honest comparison

`just` was deliberately modelled on Make's syntax — recipe name, colon, indented body — because that shape is already familiar to a huge number of developers. But `just` is not a build system, and it does not try to be one. Three specific things Make does that `just` deliberately does not:

**No file-timestamp dependency graph.** In a real Makefile, a rule like

```make
bin/zgard: cli/main.go cli/*.go
	go build -o bin/zgard ./cli
```

only runs when `bin/zgard` is older than its listed source files. Make compares modification times and skips work that is already up to date. `just` has no concept of this at all. Every recipe you invoke runs its body, every time, unconditionally. `just build-cli` recompiles the CLI whether or not anything changed since the last run.

**No incremental builds.** Following directly from the above: there is no notion in `just` of "only rebuild the parts that are stale." If `build` depends on `build-cli`, `build-daemon`, and `build-service`, invoking `just build` runs all three, always.

**No implicit rules.** Make has pattern rules — `%.o: %.c` and the like — that let it infer how to produce a file it has never been told about explicitly, based on file extensions and built-in conventions. `just` has nothing like this. Every recipe is exactly the commands you wrote, nothing inferred.

Why is this the right trade for grazhda? None of the four components benefit meaningfully from timestamp-based incrementality at the *orchestration* layer. `go build` and Maven already do their own, much better, incremental compilation internally — `go build` skips recompiling packages whose inputs have not changed, and Maven's dependency and plugin caching does the same. Layering Make's coarse, file-mtime-based staleness checking on top of tools that already do fine-grained incremental compilation buys little, and it buys a real risk: a stale `bin/zgard` that Make decided not to rebuild because some input's mtime did not look newer, even though the content changed — this happens routinely with git checkouts, clock skew, and generated files. For a project whose real goal is "one predictable verb per intent, identical locally and in CI," always-run-everything is simpler to reason about and harder to get subtly wrong.

The warning this section exists to give: if you have used Make, you will instinctively expect `just` recipes to skip themselves when "nothing changed." They never do. `just build` is always a full rebuild of everything `build` depends on. If you want caching, you rely on the underlying tool (`go build`'s package cache, Maven's local repository) doing it for you — `just` has no opinion on the matter. Do not write a Justfile expecting Make's laziness; you will not get it.

## 3. Install and verify

`just` is a single, dependency-free binary. There is no runtime to install alongside it.

```bash
# Linux/WSL — via your distribution's package manager where available:
sudo apt install just
# or, works everywhere, installs into ~/.local/bin:
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin

# macOS, via Homebrew:
brew install just

# Windows, via Scoop or winget:
scoop install just
winget install --id Casey.Just
```

Whichever route you took, confirm it worked:

```bash
$ just --version
just 1.36.0
```

The exact version number does not matter much for this guide, but note it — Justfile syntax has grown new features over time (`[group(...)]` attributes, for instance, are relatively recent), and if a recipe in this guide does something your installed `just` rejects, an out-of-date binary is the first thing to suspect.

Two more commands are worth knowing before you have even written a Justfile, because you will reach for them constantly once you have one. `just --list` prints every recipe in the Justfile found in (or above) the current directory, alongside its one-line description if it has one — the discoverability answer to the "folder of scripts" problem from section 1: new team members run this once and see the whole menu. `just --evaluate` prints every variable `just` knows about — the ones you define with `:=`, plus ones inherited from the environment — and what value it resolved each to, without running any recipe. It is invaluable when a recipe is using the wrong path or version string and you want to know what `just` actually computed before blaming the recipe body.

## 4. Your first Justfile

`just` looks for a file literally named `Justfile` (or `justfile`, lower case — both are recognised) in the current directory, and if it is not there, in each parent directory in turn, the same way `git` finds `.git`. Create one:

```just
greet:
    echo "hello from just"
```

Run it two ways:

```bash
just greet
just
```

The second form — no argument — runs the **first recipe in the file**, which here is also `greet`, so both commands do the same thing. That is the entire rule for defaults: whichever recipe is written first is what bare `just` runs. There is no separate "default" keyword required (though later we will define one explicitly, named `default`, purely for readability — that is a naming convention, not special syntax).

Token by token: `greet` is the **recipe name** — what you type after `just` to run it (`just greet`); recipe names may contain letters, digits, underscores and hyphens. The `:` (colon) separates the recipe's name (and, as section 6 shows, its dependencies) from its body — nothing else may follow it on the same line. The next line, indented, is the recipe's **body**: one shell command. `just` runs this line, and every subsequent line at the same indentation, in a shell. By default `just` also echoes each line before running it:

```
$ just greet
echo "hello from just"
hello from just
```

That double line — the command, then its output — is the default, unless the line starts with `@` (section 6 covers that).

## 5. The single most important thing to understand: one shell per line

This is the idea that trips up nearly every newcomer to `just`, so it gets its own section and a worked failure.

Suppose you write this, reasoning that it looks exactly like a shell script:

```just
build-cli:
    cd cli
    go build -o ../bin/zgard .
```

Run it:

```bash
$ just build-cli
cd cli
go build -o ../bin/zgard .
go: go.mod file not found in current directory or any parent directory
error: Recipe `build-cli` failed on line 3 with exit code 1
```

The `cd cli` succeeded — you can see it ran, with no error — and then the very next line behaved as though it had never happened. This is not a bug and it is not `just` being unusually strict. It is the mental model that governs every recipe you will ever write: **each line of a recipe's body is handed to a brand-new shell process.** The `cd` on line one changed the working directory of *that* shell, which then exited. Line two starts an entirely new shell, inheriting nothing from the first — not its working directory, not any plain (non-`export`) variable assignment, nothing. It starts back in the directory `just` itself was invoked from.

If you have written a `.sh` script before, this will feel wrong, because a shell script is one shell running every line in sequence, so a `cd` persists for the rest of the file. A `just` recipe body is not that: it looks like a script, but behaves like a list of separate, independent shell invocations, one per line.

There are two correct fixes, and which you reach for depends on how much you need to do in that directory.

**Fix one: chain the commands on a single line with `&&`.** If it is a `cd` followed by one command, this is the natural choice — and it is exactly what grazhda's own Justfile does:

```just
build-cli:
    cd cli && go build -o ../bin/zgard .
```

Now `cd cli && go build ...` is one line, handed to one shell, and `&&` means "run the second command only if the first succeeded, in the same shell" — so the `cd` and the `go build` share a working directory, because they are the same shell process.

**Fix two: make the whole recipe one shell.** When a recipe needs several commands in the same directory, chaining everything with `&&` gets unreadable fast. Two ways to get one shell for the whole body:

```just
set shell := ["bash", "-euo", "pipefail", "-c"]

build-cli:
    cd cli
    go mod tidy
    go build -o ../bin/zgard .
```

`set shell := [...]` is a Justfile-wide setting (more in section 6) that changes how *every* recipe's body is invoked — but it still runs each line as a separate invocation of that shell by default, so this alone does not fix the `cd` problem; it changes *which* shell runs each line, not how many shells run the body. What actually merges a recipe body into a single shell invocation is a **shebang recipe**, covered properly in section 8:

```just
build-cli:
    #!/usr/bin/env bash
    set -euo pipefail
    cd cli
    go mod tidy
    go build -o ../bin/zgard .
```

Here the whole indented block, starting from the `#!` line, is written to a temporary script file and executed once, as one process — so `cd cli` on one line really does affect `go build` on a later line, because now they are the same script, the way you are used to.

The rule to keep in your head permanently: **plain recipe lines are independent shells; `&&` joins commands within one line into one shell; a shebang line turns the whole body into one script.** Every `cd`-related surprise you will ever hit in a Justfile traces back to forgetting this.

## 6. Recipe syntax, properly

With the core model in place, here is the rest of the syntax, introduced against grazhda's actual needs.

### Dependencies

A recipe can list other recipes that must run, successfully, before its own body starts:

```just
build: generate
    cd cli && go build -o ../bin/zgard .
```

`just build` first runs the entire `generate` recipe, and only if it exits successfully does it proceed to `build`'s own body. Multiple dependencies are space-separated and run in the order listed:

```just
build: build-cli build-daemon build-service copy-scripts
    @echo "✓ all components built"
```

Remember section 2: this is not a staleness check. `just build` always runs `build-cli`, then `build-daemon`, then `build-service`, then `copy-scripts`, then its own body, every single time, regardless of whether anything changed.

### Parameters with defaults

Recipes can take arguments, given on the command line after the recipe name:

```just
greet name="world":
    echo "hello, {{ name }}"
```

```bash
$ just greet
hello, world
$ just greet Yuri
hello, Yuri
```

`name="world"` declares a parameter with a default value; a parameter written without `=...` is required, and calling the recipe without it is an error. `{{ name }}` — double curly braces — is how you refer to it inside the body. This is **interpolation**, and it is worth contrasting immediately with a plain shell variable, because the two look similar and are not:

```just
show version:
    echo "building {{ version }}"
    echo "shell sees: $version"
```

`{{ version }}` is substituted by `just` itself, before the shell ever sees the line — it is text substitution at the Justfile level. `$version` is a shell variable reference, resolved by the shell that runs that line, and it will be empty unless something has actually exported an environment variable of that name. `just` does not automatically turn a recipe parameter into an environment variable named the same thing. Confusing these two is a common source of "why is my variable blank" bugs; if you wrote `$name` where you meant `{{ name }}`, the shell will happily treat it as an unset environment variable and substitute nothing.

### Variables and `:=`

Justfile-level variables — as opposed to recipe parameters — are declared at the top of the file with `:=`:

```just
bin_dir := "bin"

build-cli:
    mkdir -p {{ bin_dir }}
    cd cli && go build -o ../{{ bin_dir }}/zgard .
```

`:=` (not a plain `=`) is deliberate syntax, and it is evaluated once, at parse time, before any recipe runs — so putting an expensive command substitution in a `:=` variable used by several recipes computes it once, not once per recipe.

### `$(...)` command substitution

Grazhda's build recipes embed a version string computed from git, exactly the way you would in a shell script:

```just
build-cli:
    cd cli && go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/zgard .
```

`$(...)` here is ordinary shell command substitution — it is evaluated by the shell running that line, not by `just`. Because the whole `$(...)` expression sits inside a single recipe line (no `cd` before it in a way that matters), this works exactly as it would in a shell script: `git describe` runs, and if it fails (no tags, not a git repository, etc.), the `|| echo dev` falls back to the literal string `dev`.

### `@` to suppress echoing

By default `just` prints each line before running it (section 4). Prefixing a line with `@` suppresses that echo for that line only:

```just
build-cli:
    @mkdir -p bin
    cd cli && go build -o ../bin/zgard .
    @echo "✓ bin/zgard"
```

```
$ just build-cli
cd cli && go build -o ../bin/zgard .
✓ bin/zgard
```

Notice `mkdir -p bin` never appeared, but `cd cli && go build ...` did, and the final `echo` printed only its own output. Housekeeping and status messages read much better with `@`; the actual, possibly-failing work usually reads better without it, so you can see exactly what ran when something goes wrong.

### `-` to ignore failure

A line prefixed with `-` is allowed to fail without stopping the recipe — `just`'s native equivalent of appending shell's own `|| true`:

```just
fmt:
    -cd service && mvn -q -B spotless:apply
```

If that line exits non-zero, `just` prints a warning and moves on instead of aborting the recipe. Reach for `-` when a step is genuinely optional (a formatter plugin that might not be configured yet); do not reach for it to paper over a build step you actually need to succeed — that turns a build failure into a silent no-op.

### Comments become `--list` descriptions

A `#` comment placed directly above a recipe, with no blank line between them, becomes that recipe's description in `just --list`:

```just
# Regenerate contract code from contracts/
generate:
    @mkdir -p daemon/proto
    protoc ...
```

```
$ just --list
Available recipes:
    generate   # Regenerate contract code from contracts/
```

This is the cheap, low-ceremony version of grazhda's own hand-written `help` recipe from section 9 — `--list` needs no maintenance beyond keeping the comment near the recipe.

### `[private]` recipes and `[group('x')]`

Some recipes exist only to be called by other recipes, not by a human. Mark them private and they still work when invoked directly, but disappear from `just --list`:

```just
[private]
_check-protoc:
    @which protoc >/dev/null || (echo "protoc not found on PATH" && exit 1)

generate: _check-protoc
    protoc ...
```

(A leading underscore in the name is an older convention for the same idea, often used alongside `[private]` — but `[private]` is what actually changes `just`'s behaviour.)

Once a Justfile has fifteen or twenty recipes, a flat `--list` gets hard to scan. Attaching a group to each recipe clusters them in the listing instead:

```just
[group('build')]
build-cli:
    cd cli && go build -o ../bin/zgard .

[group('test')]
test-go:
    cd libs && go test ./...
```

```
$ just --list
Available recipes:
    [build]
        build-cli
    [test]
        test-go
```

For a four-component monorepo like grazhda, groups such as `build`, `test`, and `quality` turn `--list` from a wall of text into something closer to a menu.

## 7. Working directory semantics

By default, every recipe in a Justfile runs with its working directory set to **the directory containing the Justfile**, regardless of which directory you were sitting in when you typed `just`. Run `just build-cli` from inside `daemon/`, and the recipe still starts from the repository root. This is why grazhda's recipes all say `cd cli && go build ...` rather than assuming they are already there — the CLI's `go.mod` lives in `cli/`, not at the root where the Justfile sits.

This matters more than it sounds for a project shaped like grazhda. If `just` instead inherited whatever directory the caller happened to be in, `just build-cli` would build correctly from the repo root but fail, or silently do the wrong thing, from inside `daemon/` or `service/`. Fixing the working directory to the Justfile's own location means `just <anything>` behaves identically no matter where in the tree you invoke it from — exactly the property you want from a single build entry point.

You can override this per-Justfile with a setting:

```just
set working-directory := "service"
```

This changes the starting directory for every recipe to `service/` instead of the Justfile's own directory — usually the wrong choice for a Justfile orchestrating four differently-rooted components like grazhda's, since it would help `service`-only recipes and hurt everything else. It suits a Justfile that genuinely belongs to, and only serves, one subproject.

## 8. Shebang recipes

Section 5 introduced these as the fix for "I need several commands sharing one shell, one working directory, one set of local variables." The full form:

```just
release-notes tag:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building release notes for {{ tag }}"
    prev=$(git describe --tags --abbrev=0 "{{ tag }}^" 2>/dev/null || echo "")
    if [ -n "$prev" ]; then
        git log "$prev".."{{ tag }}" --oneline
    else
        git log "{{ tag }}" --oneline
    fi
```

When a recipe's first body line begins with `#!`, `just` treats the entire block as a self-contained script: it writes the whole thing to a temporary file, makes it executable, and runs it once as a single process — using whatever interpreter the shebang names. That means `set -euo pipefail` actually applies to every subsequent line (unlike a bare Justfile line, where such a `set` would only affect the one shell it ran in and then vanish), local shell variables like `prev` persist across lines, and `cd` behaves the way it does in any script you have ever written.

This is also how you write a recipe in a language other than shell — a Python recipe is identical in shape:

```just
plot:
    #!/usr/bin/env python3
    import statistics
    data = [1, 2, 3, 4, 5]
    print(f"mean: {statistics.mean(data)}")
```

Reach for a shebang recipe whenever a recipe body is more than two or three lines that need to share state (a working directory, a variable, a loop), or whenever you want real control flow (`if`, loops, functions) rather than the more limited world of chaining commands with `&&`, `;`, and `||` on shell lines. For grazhda's short, single-purpose build steps — `cd cli && go build ...` — plain recipe lines are the better fit; reach for shebangs in things like `man` if it grows conditional logic, or in a release-notes-style recipe like the one above.

## 9. Building grazhda's Justfile, step by step

Now the real thing, assembled incrementally rather than dropped in whole. Start a file named `Justfile` at the repository root.

**Step 1 — one recipe, the CLI.**

```just
build-cli:
    @mkdir -p bin
    cd cli && go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/zgard .
    @echo "✓ bin/zgard"
```

`@mkdir -p bin` (section 6's `@`) creates the output directory quietly — `-p` makes it not an error if `bin/` already exists, and `@` hides the housekeeping line since every build recipe repeats it. The `cd cli && go build ...` line is the fix from section 5: `cd` and `go build` share a shell because `&&` joins them into one line. The `-ldflags` value bakes a version string into the binary itself, computed by `$(git describe ...)`, with a `dev` fallback for anyone building outside a git checkout.

**Step 2 — the daemon, same shape.**

```just
build-daemon:
    @mkdir -p bin
    cd daemon && go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/dukh ./cmd
    @echo "✓ bin/dukh"
```

Same pattern, different module and build target (`./cmd`, because `dukh`'s `main` package lives in a `cmd` subdirectory rather than the module root). Nothing conceptually new — that is the payoff of a boring, repeatable pattern.

**Step 3 — the daemon needs generated code first.** `dukh` speaks gRPC defined in `contracts/dukh.proto`, and that has to be compiled into Go source under `daemon/proto/` before `go build` will even find the types it references. This is a real ordering dependency, and it is exactly what recipe dependencies (section 6) are for:

```just
generate:
    @echo "Generating contract code…"
    @mkdir -p daemon/proto
    protoc \
      --go_out=daemon/proto --go_opt=paths=source_relative \
      --go-grpc_out=daemon/proto --go-grpc_opt=paths=source_relative \
      --proto_path=contracts \
      contracts/dukh.proto
    @echo "✓ daemon/proto generated from contracts/dukh.proto"

build-daemon: generate
    @mkdir -p bin
    cd daemon && go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ../bin/dukh ./cmd
    @echo "✓ bin/dukh"
```

`build-daemon: generate` means `just build-daemon` now always runs `generate` first. Remember section 2's warning: this is unconditional — every `just build-daemon` reruns `protoc` in full, whether or not `dukh.proto` changed. A deliberate, acceptable cost here (`protoc` is fast) for the certainty that generated code is never stale.

**Step 4 — the Java service, and a top-level `build`.** `molfar` builds with Maven, and grazhda's Justfile treats that identically to the Go builds — one recipe, `cd` into the module, run the tool, copy the artifact out:

```just
build-service:
    @mkdir -p bin
    cd service && mvn -q -B -DskipTests package
    cp service/target/molfar.jar bin/molfar.jar
    @echo "✓ bin/molfar.jar"
```

`-q` quietens Maven's normally verbose output, `-B` runs it in batch (non-interactive) mode suited to both a terminal and CI, and `-DskipTests` is there because packaging and testing are separate recipes here (`build-service` and `test-java`) — a `just build` should not silently also run the test suite, nor should `just test` rebuild the jar it is about to run.

Now tie the four build steps together:

```just
build: build-cli build-daemon build-service copy-scripts
    @echo "✓ all components built"

copy-scripts:
    @mkdir -p bin
    cp scripts/grazhda scripts/grazhda-init.sh bin/
    chmod +x bin/grazhda bin/grazhda-init.sh
    cp .grazhda.env .grazhda.pkgs.yaml bin/
    @echo "✓ scripts staged in bin/"
```

`build` itself has no interesting body — its job is entirely the dependency list. `just build` now runs, in order, `build-cli`, `build-daemon` (which pulls in `generate` first), `build-service`, and `copy-scripts`, printing its own `✓` only once all four succeed. This one command is now the answer to "build everything," for a developer and CI alike — the promise made in section 1.

**Step 5 — test, one recipe per toolchain, one recipe that runs them all.**

```just
test: test-go test-java test-bash
    @echo "✓ all tests passed"

test-go:
    cd libs && go test ./...
    cd cli && go test ./...
    cd daemon && go test ./...

test-java:
    cd service && mvn -q -B test

test-bash:
    bash tests/bash/test_scripts.sh
```

`test-go` is three separate lines, each its own `cd` plus `go test` — section 5 again — because each line is its own shell, so the three `cd`s never interfere. Here the "one shell per line" rule is actually convenient rather than a trap: three independent test runs, three independent working directories, no shebang needed. `test` composes all three suites the same way `build` composed the four build steps.

**Step 6 — quality: `fmt` and `lint`.**

```just
fmt:
    cd libs && go fmt ./...
    cd cli && go fmt ./...
    cd daemon && go fmt ./...
    cd service && mvn -q -B spotless:apply || true

lint:
    cd libs && go vet ./...
    cd cli && go vet ./...
    cd daemon && go vet ./...
    shellcheck scripts/grazhda scripts/grazhda-init.sh scripts/grazhda-install.sh tests/bash/test_scripts.sh
```

`fmt`'s last line uses plain shell's `|| true` rather than `just`'s `-` prefix from section 6 — either works, and `|| true` is what grazhda's own Justfile happens to use — because not every checkout has the Spotless Maven plugin configured, and a missing formatter should not fail the whole recipe. `lint` runs `go vet` per Go module, then `shellcheck` across every shell script in one line, since `shellcheck` accepts multiple file arguments directly — no `cd`-per-module needed there.

**Step 7 — `clean`.**

```just
clean:
    rm -rf bin man service/target
    @echo "✓ clean"
```

Worth pausing on only to reinforce section 2 once more: there is no `just` equivalent of Make's "only clean if there is something to clean." `just clean` runs `rm -rf` unconditionally, every time — `rm -rf` on a directory that does not exist is simply a no-op, which is bash's behaviour, not `just`'s.

**Step 8 — `help` as the default.** Recall from section 4 that bare `just` runs whichever recipe is written first. Put `default: help` at the very top of the file, before any other recipe, and give it a body that lists everything by hand:

```just
default: help

help:
    @echo "Grazhda — spec-driven build system"
    @echo ""
    @echo "Usage: just <task>"
    @echo ""
    @echo "  generate       Regenerate contract code from contracts/"
    @echo "  build          Build every component into bin/"
    @echo "  build-cli      Build the CLI interface (zgard)"
    @echo "  build-daemon   Build the CLI daemon (dukh)"
    @echo "  build-service  Build the long-term process (molfar)"
    @echo "  copy-scripts   Stage the toolchain manager scripts into bin/"
    @echo "  test           Run every test suite"
    @echo "  test-go        Run the Go test suites"
    @echo "  test-java      Run the molfar test suite"
    @echo "  test-bash      Run the shell test harness"
    @echo "  fmt            Format all sources"
    @echo "  lint           Vet Go sources and shellcheck the scripts"
    @echo "  clean          Remove build output"
```

`default: help` is a recipe named `default` whose only job is to depend on `help` and do nothing else, so both `just` (bare) and `just default` run `help`. This is hand-maintained, unlike `--list`'s comment-derived output from section 6, and grazhda's Justfile keeps both: `--list` for a quick, always-accurate machine-derived view, and this curated `help` as the polished, ordered front door for someone typing bare `just` with no idea where to start. Keeping it in sync as recipes are added or removed is a real, ongoing maintenance cost — worth naming rather than discovering by drift.

At this point the file has grown to roughly what you saw referenced at the top of this guide: `generate`, `build` and its four dependents, `test` and its three, `fmt`, `lint`, `clean`, and `help`/`default` — a complete, working build entry point for a four-language monorepo, built up one recipe at a time, each addition motivated by a concrete need.

## 10. Using it in CI

The entire point of putting this logic in a Justfile rather than scattering it across `.github/workflows/*.yml` is that the same verbs run identically in both places. A GitHub Actions job for grazhda needs only to install `just` and then call the same recipes a developer runs locally:

```yaml
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
      - uses: actions/setup-java@v4
        with:
          distribution: "temurin"
          java-version: "25"
      - run: just build
      - run: just test
      - run: just lint
```

`extractions/setup-just` is a small, widely-used GitHub Action that downloads a pinned `just` release and puts it on `PATH` — pinning the version avoids CI silently drifting to a newer `just` than developers have installed locally.

The symmetry matters concretely: when CI fails on `just test`, you can run that exact command on your own machine and expect the same behaviour — same recipes, same dependency order, same failure output — because there is no separate "CI script" reimplementing the logic with different flags or a different working-directory assumption. Debugging a CI failure becomes "run this one command locally," not "reverse-engineer the YAML."

## 11. Gotchas and debugging

**`just --dry-run`** prints every command a recipe (and its dependencies) would run, without actually running any of them:

```bash
$ just --dry-run build
mkdir -p bin
cd cli && go build -ldflags "..." -o ../bin/zgard .
...
```

Use it before running anything destructive-looking, or when you have just edited a recipe and want to check the expansion of `{{ }}` interpolations and `$(...)` substitutions before committing to actually executing them.

**`just --show <recipe>`** prints the fully-resolved body of one recipe — useful when a recipe depends on others and you want to see exactly what it expands to without wading through the whole file.

**`just --fmt`** reformats the Justfile itself to a canonical style (consistent indentation, spacing). At the time of writing this is still marked unstable in some `just` versions, requiring `just --fmt --unstable`; check `just --help` on your installed version. Run it, then diff the result before committing — like any auto-formatter, it is a convenience, not an oracle.

**Indentation must be consistent within a recipe.** `just`, like Python and unlike most Make implementations, cares about consistent indentation for a recipe body — mixing tabs and spaces, or changing indentation width partway through one recipe, produces a parse error rather than being silently tolerated:

```
error: Recipe line has inconsistent leading whitespace
```

Pick one style (this guide uses four spaces throughout) and let your editor enforce it; do not hand-mix tabs and spaces in a Justfile.

**`{{ }}` interpolation versus shell `$`, restated once more** because it is the second-most common mistake after the one-shell-per-line issue. `{{ name }}` is resolved by `just`, before the shell sees the line, using recipe parameters and Justfile variables. `$name` is resolved by the shell running that line, using environment variables it actually has. They are not interchangeable, and `just` will not warn you if you use the wrong one — a mistyped `$variable` simply expands to an empty string, silently.

**Exported variables.** A Justfile variable declared with `export` becomes an actual environment variable for every shell that runs a recipe line, meaning `$` (not `{{ }}`) can see it:

```just
export GOFLAGS := "-mod=mod"

build-cli:
    cd cli && go build -o ../bin/zgard .
```

Here `GOFLAGS` is set in the environment of the shell that runs `go build`, exactly as if you had run `export GOFLAGS=-mod=mod` yourself before invoking `go build` in a terminal — Go itself reads `GOFLAGS` from the environment, so this line never needs to reference `{{ GOFLAGS }}` explicitly.

**`.env` loading.** `just` can load a `.env` file's contents as environment variables automatically:

```just
set dotenv-load
```

Placed at the top of the Justfile, this makes every `KEY=value` line in a `.env` file next to the Justfile available as an environment variable (`$KEY`) inside every recipe, without writing `export` for each one by hand. Grazhda ships a `.grazhda.env`, but it is a *runtime* configuration file for the `grazhda` toolchain manager itself, copied into `bin/` by `copy-scripts` — not a build-time `.env` that `set dotenv-load` should pick up. Do not conflate "a file with `.env` in its name" with "the file `just`'s dotenv loading looks for," which is specifically named `.env` unless overridden with `set dotenv-filename := "..."`.

**Calling one recipe from another, versus depending on it.** Section 6 covered dependencies (`build: build-cli build-daemon`), which run before the depending recipe's own body, at most once per invocation even if listed by several recipes. There is a second way to invoke a recipe: literally calling `just` again, by name, inside another recipe's body:

```just
release: build
    just test
    git tag "$(cat VERSION)"
```

Here `just test` inside `release`'s body starts an entirely new `just` process partway through `release`'s own execution — unlike `release: build test`, which would run `test` as a dependency before `release`'s body even starts. The distinction rarely bites in a project the size of grazhda, but it matters whenever a recipe needs "run this other recipe partway through my own body, with setup before it and cleanup after" — a shape dependencies cannot express, since dependencies always run first, never in the middle.

## 12. Cheat sheet

```
just                      run the default recipe (the first one in the file)
just <recipe>             run one recipe
just <recipe> arg1 arg2   run a recipe, passing positional arguments
just --list               list every non-private recipe with its description
just --evaluate           print every variable just has resolved
just --dry-run <recipe>   show what would run, without running it
just --show <recipe>      print one recipe's fully-resolved body
just --fmt                reformat the Justfile itself

recipe:                   recipe name, then a colon
    command                indented body line — its own shell, by default

recipe: dep1 dep2         dependencies: run first, in order, once each

recipe param="default":   parameter with a default value
    echo {{ param }}       {{ }} = just-level interpolation (resolved before the shell runs)
    echo $SOME_ENV_VAR      $VAR = shell-level, needs an actual exported env var

name := "value"           Justfile variable, evaluated once at parse time
export NAME := "value"    same, but exported into every recipe's shell environment
set dotenv-load           load ./.env into every recipe's environment automatically
set shell := [...]        change the shell/interpreter used to run recipe lines
set working-directory := "dir"   change the directory every recipe starts in

@command                  suppress echoing this line before it runs
-command                  allow this line to fail without stopping the recipe

# a comment right above a recipe becomes its --list description
[private]                 hide this recipe from --list (still callable directly)
[group('name')]           cluster this recipe under a heading in --list

#!/usr/bin/env bash        shebang recipe: whole body is ONE script, ONE shell,
set -euo pipefail          so cd and local variables persist across lines —
cd somewhere                use this whenever plain lines force too much && chaining
do_a_thing

just other-recipe          inside a body: calls just again, as a NEW process,
                           partway through — different from listing it as a dependency
```

The one rule worth carrying above all the others: every plain line in a recipe body is its own shell. If something that changes shell state — a `cd`, a plain variable assignment, a `source` — needs to affect a later line, either chain them on one line with `&&`, or write the recipe as a shebang script. Everything else in this guide is detail; that one is the trap.
