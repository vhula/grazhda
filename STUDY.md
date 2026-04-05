# Grazhda — Study Guide for Junior Go Developers

This guide explains the Grazhda project structure and its Go source code from first principles. No prior Go knowledge is assumed. By the end you will understand how the project is organised, what each file does, and why Go code is written the way it is.

---

## Table of Contents

1. [What is Grazhda?](#1-what-is-grazhda)
2. [The Go Module System](#2-the-go-module-system)
3. [Project Directory Layout](#3-project-directory-layout)
4. [Reading a Go File — Fundamentals](#4-reading-a-go-file--fundamentals)
5. [internal/config — Loading Configuration](#5-internalconfig--loading-configuration)
6. [internal/workspace — The Core Domain](#6-internalworkspace--the-core-domain)
   - [options.go — RunOptions struct](#61-optionsgo--runoptions-struct)
   - [executor.go — Running Shell Commands](#62-executorgo--running-shell-commands)
   - [mock.go — Faking Execution in Tests](#63-mockgo--faking-execution-in-tests)
   - [reporter.go — Showing Progress](#64-reportergo--showing-progress)
   - [targeting.go — Choosing Workspaces](#65-targetinggo--choosing-workspaces)
   - [workspace.go — Init, Purge, Pull](#66-workspacego--init-purge-pull)
7. [zgard — The CLI Entry Point](#7-zgard--the-cli-entry-point)
   - [main.go — The Program Starts Here](#71-maingo--the-program-starts-here)
   - [root.go — The Root Command](#72-rootgo--the-root-command)
   - [zgard/ws — Workspace Subcommands](#73-zgardws--workspace-subcommands)
8. [How Data Flows — End-to-End](#8-how-data-flows--end-to-end)
9. [Testing in Go](#9-testing-in-go)
10. [Build System (Justfile)](#10-build-system-justfile)
11. [Go Concepts Glossary](#11-go-concepts-glossary)

---

## 1. What is Grazhda?

Grazhda is a **command-line tool** (CLI) that manages software development workspaces on your computer. A *workspace* is a directory that holds several git repositories, organised into projects.

When you type `zgard ws init`, Grazhda:
1. Reads a YAML configuration file to understand which repositories belong to the workspace.
2. Creates the directory structure on disk.
3. Runs `git clone` for each repository.
4. Prints a live progress report and a summary.

There is only one user-facing binary: **`zgard`**.

---

## 2. The Go Module System

Before looking at source code, you need to understand how Go organises code.

### Packages

Every Go source file starts with `package <name>`. A **package** is a group of `.go` files in the same directory. All files in a directory share the same package name and can see each other's types and functions directly.

```go
// This file is part of the "config" package.
package config
```

### Modules

A **module** is a collection of packages with a root `go.mod` file that declares the module's name (called its *import path*). Grazhda has two active modules:

| Directory | Module path |
| :--- | :--- |
| `internal/` | `github.com/vhula/grazhda/internal` |
| `zgard/` | `github.com/vhula/grazhda/zgard` |

The module path is used to import packages from *other* modules:

```go
import "github.com/vhula/grazhda/internal/config"
```

### Go Workspace (`go.work`)

The top-level `go.work` file tells Go that both modules live locally together — so they can reference each other without being published online:

```
go 1.26.1

use (
    ./internal
    ./zgard
    ./dukh    // placeholder for future component
)
```

### The `internal/` Convention

Any package under a directory named `internal` can only be imported by code in the **parent** of that `internal/` directory. Here, `internal/config` can only be imported by `zgard/` (its sibling) or other code in the same repository. This prevents external users of the module from depending on your private implementation details.

---

## 3. Project Directory Layout

```
grazhda/
├── go.work                  ← ties modules together locally
├── Justfile                 ← build/test/fmt shortcuts (like a Makefile)
├── config.template.yaml     ← copy this to ~/.grazhda/config.yaml
├── README.md
├── STUDY.md                 ← this file
│
├── internal/                ← module: github.com/vhula/grazhda/internal
│   ├── go.mod               ← module declaration + dependencies
│   ├── config/
│   │   ├── config.go        ← config types, Load, Validate, RenderCloneCmd
│   │   └── config_test.go   ← 14 tests for config logic
│   ├── workspace/
│   │   ├── options.go       ← RunOptions struct (dry-run, verbose, parallel, etc.)
│   │   ├── executor.go      ← Executor interface + OsExecutor (real shell runner)
│   │   ├── mock.go          ← MockExecutor (fake runner for tests)
│   │   ├── reporter.go      ← Reporter: prints ✓/⏭/✗ lines and summary
│   │   ├── targeting.go     ← Resolve: picks which workspaces to operate on
│   │   ├── workspace.go     ← Init, Purge, Pull — the main operations
│   │   ├── workspace_test.go
│   │   ├── reporter_test.go
│   │   └── targeting_test.go
│   └── testdata/            ← YAML fixture files used by tests
│       ├── valid_single_workspace.yaml
│       ├── valid_multi_workspace.yaml
│       └── ...
│
├── zgard/                   ← module: github.com/vhula/grazhda/zgard
│   ├── go.mod
│   ├── main.go              ← program entry point: calls Execute()
│   ├── root.go              ← defines the "zgard" Cobra command
│   └── ws/
│       ├── ws.go            ← "zgard ws" parent command
│       ├── init.go          ← "zgard ws init" subcommand
│       ├── purge.go         ← "zgard ws purge" subcommand
│       ├── pull.go          ← "zgard ws pull" subcommand
│       ├── config.go        ← resolveConfigPath() helper
│       └── confirm.go       ← confirm() prompt helper
│
└── dukh/                    ← future gRPC server (placeholder only)
```

**Key insight:** `internal/workspace` contains *all* the business logic — the rules about how workspaces are managed. The `zgard/ws/` packages contain only the CLI glue — parsing flags, reading config, calling workspace functions, and calling `os.Exit`.

---

## 4. Reading a Go File — Fundamentals

Here is the skeleton of any Go source file:

```go
package mypackage          // 1. Package declaration

import (                   // 2. Imports — what other packages we use
    "fmt"
    "os"
    "github.com/vhula/grazhda/internal/config"
)

// MyStruct is a type with named fields.
type MyStruct struct {
    Name string
    Age  int
}

// NewMyStruct creates a MyStruct. By convention, constructors are named New<Type>.
func NewMyStruct(name string, age int) *MyStruct {
    return &MyStruct{Name: name, Age: age}
}

// Greet is a method on *MyStruct. The (m *MyStruct) part is the "receiver".
func (m *MyStruct) Greet() string {
    return fmt.Sprintf("Hello, I am %s, age %d", m.Name, m.Age)
}
```

### Pointers (`*` and `&`)

Go passes values by copy. When a function needs to *modify* a value — or when a struct is large — you use a **pointer**:

- `*MyStruct` means "a pointer to a MyStruct" (the memory address of one)
- `&value` means "take the address of value"
- When you see `return &MyStruct{...}`, the function returns the address of the new struct, not a copy

### Error Handling

Go does not have exceptions. Instead, functions return an `error` value as the last return:

```go
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("config file %q: %w", path, err)
    }
    // success — return the value and nil for the error
    return &cfg, nil
}
```

The caller always checks `if err != nil`. The `%w` verb in `fmt.Errorf` *wraps* the original error so it can be unwrapped later.

### Interfaces

An **interface** describes behaviour — a set of methods a type must have. Any type that has those methods automatically satisfies the interface (no explicit `implements` keyword):

```go
type Executor interface {
    Run(dir string, command string) error
}

// OsExecutor has a Run method, so it satisfies Executor automatically.
type OsExecutor struct{}

func (e OsExecutor) Run(dir string, command string) error { ... }
```

Interfaces are how Go achieves testability — you can swap the real `OsExecutor` for a `MockExecutor` in tests.

---

## 5. internal/config — Loading Configuration

**File:** `internal/config/config.go`

### The Type Hierarchy

Grazhda's config maps directly to Go types using struct *tags* that tell the YAML decoder which field name to look for:

```go
type Config struct {
    Workspaces []Workspace `yaml:"workspaces"`
}

type Workspace struct {
    Name                 string    `yaml:"name"`
    Default              bool      `yaml:"default"`
    Path                 string    `yaml:"path"`
    CloneCommandTemplate string    `yaml:"clone_command_template"`
    Projects             []Project `yaml:"projects"`
}

type Project struct {
    Name         string       `yaml:"name"`
    Branch       string       `yaml:"branch"`
    Repositories []Repository `yaml:"repositories"`
}

type Repository struct {
    Name         string `yaml:"name"`
    Branch       string `yaml:"branch,omitempty"`
    LocalDirName string `yaml:"local_dir_name,omitempty"`
}
```

`[]Workspace` means a **slice** (a dynamically-sized list) of Workspace values. The `,omitempty` tag means the field is optional in the YAML file.

### Load

```go
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)    // read raw bytes from disk
    if err != nil {
        return nil, fmt.Errorf("config file %q: %w", path, err)
    }
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {  // parse YAML into cfg
        return nil, fmt.Errorf("parsing config %q: %w", path, err)
    }
    return &cfg, nil
}
```

`yaml.Unmarshal` fills in the `cfg` struct by matching YAML keys to struct field tags.

### Validate

`Validate` checks for mistakes *before* any filesystem operation runs. It returns a `[]string` — a slice of human-readable error messages:

```go
func Validate(cfg *Config) []string {
    var errs []string
    seenWS := make(map[string]bool)   // map to detect duplicates

    for i, ws := range cfg.Workspaces {
        if ws.Name == "" {
            errs = append(errs, fmt.Sprintf("workspace[%d].name: required field missing", i))
        } else if seenWS[ws.Name] {
            errs = append(errs, fmt.Sprintf("workspace names must be unique: duplicate name %q", ws.Name))
        } else {
            seenWS[ws.Name] = true
        }
        // ... more checks
    }
    return errs
}
```

`make(map[string]bool)` creates an empty map from string keys to boolean values. `seenWS[ws.Name] = true` records that we've seen a name; `seenWS[ws.Name]` returns `true` if the key exists, `false` otherwise.

### RenderCloneCmd

This function fills in a Go template (e.g. `git clone --branch {{.Branch}} https://github.com/org/{{.RepoName}} {{.DestDir}}`) with real values:

```go
func RenderCloneCmd(tmplStr, projectPath string, proj Project, repo Repository) (string, error) {
    branch := repo.Branch
    if branch == "" {
        branch = proj.Branch   // repo inherits project branch if not set
    }
    destName := repo.LocalDirName
    if destName == "" {
        destName = repo.Name
    }
    data := CloneTemplateData{
        Branch:   branch,
        RepoName: repo.Name,
        DestDir:  filepath.Join(projectPath, destName),
    }
    t, err := template.New("clone").Parse(tmplStr)
    if err != nil {
        return "", fmt.Errorf("parsing clone template: %w", err)
    }
    var buf bytes.Buffer
    if err := t.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("rendering clone template: %w", err)
    }
    return buf.String(), nil
}
```

`bytes.Buffer` is an in-memory buffer that implements `io.Writer`. `template.Execute` writes the rendered output into the buffer, and `buf.String()` converts it to a regular string.

---

## 6. internal/workspace — The Core Domain

Everything in the `workspace` package works together. All six files share `package workspace` so they can use each other's types and functions directly without an import.

### 6.1 options.go — RunOptions struct

```go
package workspace

// RunOptions controls the behaviour of workspace operations.
type RunOptions struct {
    DryRun    bool
    Verbose   bool
    Parallel  bool
    NoConfirm bool
}
```

This is a simple **struct** — a collection of named fields. All four fields are `bool` (true/false). In Go, `bool` fields default to `false`, so a zero-value `RunOptions{}` means "normal run, no flags set".

This struct is passed to `Init`, `Purge`, and `Pull` so they share a consistent set of options.

### 6.2 executor.go — Running Shell Commands

```go
package workspace

import "os/exec"

type Executor interface {
    Run(dir string, command string) error
}

type OsExecutor struct{}

func (e OsExecutor) Run(dir string, command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = dir
    return cmd.Run()
}
```

`exec.Command("sh", "-c", command)` creates a new system process that runs the shell command. `cmd.Dir = dir` sets the working directory. `cmd.Run()` executes it and blocks until it exits, returning an error if the process failed.

The `Executor` interface has only one method: `Run`. This makes it easy to swap implementations:
- In production: `OsExecutor` — runs real shell commands
- In tests: `MockExecutor` — records calls, never touches the filesystem

### 6.3 mock.go — Faking Execution in Tests

```go
package workspace

import "sync"

type MockExecutor struct {
    mu    sync.Mutex
    Calls []string
    Err   error
    ErrFn func(callIndex int) error
}

func (m *MockExecutor) Run(dir string, command string) error {
    m.mu.Lock()
    m.Calls = append(m.Calls, command)
    idx := len(m.Calls)
    errFn := m.ErrFn
    m.mu.Unlock()

    if errFn != nil {
        return errFn(idx)
    }
    return m.Err
}
```

**Why the mutex?** When `--parallel` is used, multiple goroutines call `Run` at the same time. `sync.Mutex` ensures only one goroutine modifies `m.Calls` at a time — otherwise two goroutines writing to the slice simultaneously would cause a *data race* (a bug where the result depends on timing).

`m.mu.Lock()` acquires the lock; `m.mu.Unlock()` releases it. The `defer` pattern (used in reporter below) is more idiomatic, but here the lock is released early on purpose — the error function is called *outside* the lock to avoid holding it longer than necessary.

`ErrFn` is a **function field**: `func(callIndex int) error` means "a function that takes an int and returns an error". Tests set this to simulate the first clone failing:

```go
mock.ErrFn = func(call int) error {
    if call == 1 { return errors.New("clone failed") }
    return nil
}
```

### 6.4 reporter.go — Showing Progress

```go
package workspace

type Reporter struct {
    out     io.Writer
    errOut  io.Writer
    mu      sync.Mutex
    results []OpResult
}

func NewReporter(out, errOut io.Writer) *Reporter {
    return &Reporter{out: out, errOut: errOut}
}
```

`io.Writer` is a standard library interface with one method: `Write(p []byte) (n int, err error)`. Both `os.Stdout` and `strings.Builder` implement it. By accepting `io.Writer` instead of writing directly to `os.Stdout`, the Reporter works identically in tests (writing to a `strings.Builder`) and in production (writing to the terminal).

The unexported fields (`out`, `errOut`, `mu`, `results`) start with lowercase letters — Go's convention for package-private visibility. They can only be accessed from within the `workspace` package.

#### Record

```go
func (r *Reporter) Record(res OpResult) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.results = append(r.results, res)

    symbol := "✓"
    if res.Err != nil {
        symbol = "✗"
    } else if res.Skipped {
        symbol = "⏭"
    }
    fmt.Fprintf(r.out, "    %s %-14s — %s\n", symbol, res.Repo, displayMsg)
}
```

`defer r.mu.Unlock()` schedules `Unlock` to run when the function returns, regardless of how it exits. This is Go's idiomatic way to ensure a lock is always released.

`%-14s` in the format string means: format as a string, left-aligned, padded to at least 14 characters. This makes repo names line up in columns.

#### Summary

```go
func (r *Reporter) Summary(successLabel string, dryRun bool) {
    // count success, skipped, failed from r.results
    fmt.Fprintf(r.out, "\n%s✓ %d %s  ⏭ %d skipped  ✗ %d failed\n",
        prefix, success, successLabel, skipped, failed)
    for _, res := range r.results {
        if res.Err != nil {
            fmt.Fprintf(r.errOut, "      %s: %s\n", res.Repo, res.Err.Error())
        }
    }
}
```

The `successLabel` string varies by command: `"cloned"`, `"pulled"`, `"removed"`, or `"would clone"` in dry-run. This avoids duplicating the Summary function for each command.

Failures are printed to `errOut` (stderr) while the summary line goes to `out` (stdout) — following the Unix convention that diagnostics go to stderr.

### 6.5 targeting.go — Choosing Workspaces

```go
func Resolve(cfg *config.Config, wsName string, all bool) ([]config.Workspace, error) {
    if wsName != "" && all {
        return nil, fmt.Errorf("--ws and --all are mutually exclusive")
    }
    if all {
        return cfg.Workspaces, nil
    }
    if wsName != "" {
        for _, ws := range cfg.Workspaces {
            if ws.Name == wsName {
                return []config.Workspace{ws}, nil
            }
        }
        return nil, fmt.Errorf("workspace %q not found in config", wsName)
    }
    ws, err := config.DefaultWorkspace(cfg)
    if err != nil {
        return nil, err
    }
    return []config.Workspace{*ws}, nil
}
```

`Resolve` handles four cases in priority order:
1. `--ws` and `--all` together → error (mutually exclusive)
2. `--all` → return every workspace
3. `--ws <name>` → return the named workspace
4. Neither → return the default workspace

`[]config.Workspace{ws}` creates a one-element slice literal. `*ws` dereferences the pointer returned by `DefaultWorkspace` to get the value.

### 6.6 workspace.go — Init, Purge, Pull

This is the largest file — the main orchestration logic.

#### expandHome

```go
func expandHome(path string) string {
    if len(path) == 0 || path[0] != '~' {
        return path
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return path
    }
    return filepath.Join(home, path[1:])
}
```

`path[0]` is the first byte of the string. `path[1:]` is a **slice** of the string starting from index 1 (everything after the `~`). `filepath.Join` concatenates path segments with the OS-appropriate separator.

#### Init

```go
func Init(ws config.Workspace, exec Executor, rep *Reporter, opts RunOptions) error {
    rep.PrintLine("Workspace: " + ws.Name)
    wsPath := expandHome(ws.Path)

    for _, proj := range ws.Projects {
        rep.PrintLine("  Project: " + proj.Name)
        projPath := filepath.Join(wsPath, proj.Name)

        if !opts.DryRun {
            if err := os.MkdirAll(projPath, 0o755); err != nil {
                return fmt.Errorf("creating directory %s: %w", projPath, err)
            }
        }

        if opts.Parallel {
            var wg sync.WaitGroup
            for _, repo := range proj.Repositories {
                wg.Add(1)
                repo := repo                  // capture loop variable
                go func() {
                    defer wg.Done()
                    cloneRepo(ws, proj, projPath, repo, exec, rep, opts)
                }()
            }
            wg.Wait()
        } else {
            for _, repo := range proj.Repositories {
                cloneRepo(ws, proj, projPath, repo, exec, rep, opts)
            }
        }
    }
    return nil
}
```

**`sync.WaitGroup`** is a counter for goroutines:
- `wg.Add(1)` increments the counter before launching each goroutine
- `wg.Done()` (via defer) decrements it when the goroutine finishes
- `wg.Wait()` blocks until the counter reaches zero — i.e., all goroutines are done

**`go func() { ... }()`** launches a goroutine — a lightweight concurrent task. Goroutines run concurrently (potentially in parallel on multi-core machines).

**`repo := repo`** (the "capture" line) is a Go idiom for loop variables before Go 1.22. Without it, all goroutines would reference the same `repo` variable and might all see the last value of the loop. Creating a local copy ensures each goroutine sees its own value.

**`0o755`** is an octal literal for Unix permissions: owner can read/write/execute; group and others can read/execute.

#### cloneRepo — the defer/success pattern

```go
func cloneRepo(...) {
    // ...

    var success bool
    defer func() {
        if !success {
            os.RemoveAll(repoPath)  // cleanup if clone failed
        }
    }()

    if err := exec.Run(projPath, cmd); err != nil {
        rep.Record(OpResult{..., Err: err})
        return   // success remains false → defer cleans up
    }

    success = true   // clone succeeded → defer does nothing
    rep.Record(OpResult{..., Msg: "cloned"})
}
```

The `defer` here implements *rollback on failure*: if the clone command fails, the partially-created directory is removed so the next run will retry from a clean state.

#### Purge

```go
func Purge(ws config.Workspace, rep *Reporter, opts RunOptions) error {
    wsPath := expandHome(ws.Path)

    if _, err := os.Stat(wsPath); os.IsNotExist(err) {
        rep.Record(OpResult{..., Skipped: true, Msg: "directory not found, skipped"})
        return nil
    }

    if err := os.RemoveAll(wsPath); err != nil {
        rep.Record(OpResult{..., Err: err})
        return nil
    }
    // ...
}
```

`os.Stat` returns info about a file/directory. If the path does not exist, `os.IsNotExist(err)` returns `true`. `os.RemoveAll` deletes a directory and all its contents — the equivalent of `rm -rf`.

Note that Purge *does not take an Executor* — it only calls standard library file functions, not shell commands. This makes it simpler and more predictable.

#### Pull

`Pull` mirrors `Init`'s structure but calls `pullRepo` instead of `cloneRepo`. It skips repositories that have not been cloned yet (the directory does not exist):

```go
if _, err := os.Stat(repoPath); os.IsNotExist(err) {
    rep.Record(OpResult{..., Skipped: true, Msg: "not present, skipped"})
    return
}
cmd := fmt.Sprintf("git pull --rebase origin %s", branch)
exec.Run(repoPath, cmd)
```

The pull command is constructed as a plain string, not a template — it is always `git pull --rebase origin <branch>` with no customisation needed.

---

## 7. zgard — The CLI Entry Point

The `zgard` module uses the [Cobra](https://github.com/spf13/cobra) library to build the CLI. Cobra organises commands as a tree: `zgard` → `ws` → `init | purge | pull`.

### 7.1 main.go — The Program Starts Here

```go
package main

func main() {
    Execute()
}
```

`package main` is special — the Go toolchain looks for `func main()` here to start the program. This file is deliberately minimal; all logic is in `root.go`.

### 7.2 root.go — The Root Command

```go
package main

import (
    "os"

    "github.com/spf13/cobra"
    "github.com/vhula/grazhda/zgard/ws"
)

var rootCmd = &cobra.Command{
    Use:   "zgard",
    Short: "Workspace lifecycle manager",
    Long:  "zgard manages local workspace lifecycle — init, purge, and pull repositories.",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.AddCommand(ws.NewCmd())
}
```

`var rootCmd = &cobra.Command{...}` declares a **package-level variable** — it lives for the duration of the program.

`func init()` is a special Go function that runs automatically before `main()`. Go allows multiple `init()` functions across files; they run in the order the files are imported. Here it registers the `ws` subcommand tree.

`ws.NewCmd()` is a constructor from the `zgard/ws` package that returns the fully-configured `ws` command.

### 7.3 zgard/ws — Workspace Subcommands

#### ws.go

```go
package ws

func NewCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "ws",
        Short: "Workspace operations",
    }
    cmd.AddCommand(newInitCmd())
    cmd.AddCommand(newPurgeCmd())
    cmd.AddCommand(newPullCmd())
    return cmd
}
```

`NewCmd` is exported (capital N) — callable from other packages. `newInitCmd`, `newPurgeCmd`, `newPullCmd` are unexported (lowercase n) — they are implementation details of the `ws` package.

#### config.go

```go
func resolveConfigPath() string {
    dir := os.Getenv("GRAZHDA_DIR")
    if dir == "" {
        home, _ := os.UserHomeDir()
        dir = filepath.Join(home, ".grazhda")
    }
    return filepath.Join(dir, "config.yaml")
}
```

`os.Getenv` reads an environment variable. The `_` discards the error from `UserHomeDir` (if the home directory cannot be found, `dir` will be empty and the path will be relative — acceptable fallback).

#### confirm.go

```go
func confirm(out io.Writer, reader io.Reader, msg string, paths []string) bool {
    fmt.Fprintln(out, msg)
    for _, p := range paths {
        fmt.Fprintf(out, "  %s\n", p)
    }
    fmt.Fprint(out, "Confirm? [y/N]: ")

    scanner := bufio.NewScanner(reader)
    if scanner.Scan() {
        answer := strings.TrimSpace(scanner.Text())
        return strings.EqualFold(answer, "y")
    }
    return false
}
```

`bufio.NewScanner` wraps a reader to read line by line. `scanner.Scan()` reads the next line, returning `false` at EOF. `strings.EqualFold` compares strings case-insensitively — so `y`, `Y`, `yes` all work.

By accepting `io.Writer` and `io.Reader` as parameters instead of using `os.Stdout`/`os.Stdin` directly, this function is fully testable: pass a `strings.Builder` for output and `strings.NewReader("y\n")` for input.

#### init.go — a complete command

```go
func newInitCmd() *cobra.Command {
    var dryRun bool
    var verbose bool
    var parallel bool
    var wsName string
    var all bool

    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize a workspace by cloning all repositories",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfgPath := resolveConfigPath()
            cfg, err := config.Load(cfgPath)
            if err != nil { return err }

            if errs := config.Validate(cfg); len(errs) > 0 {
                for _, e := range errs { fmt.Fprintln(os.Stderr, e) }
                return fmt.Errorf("configuration is invalid")
            }

            workspaces, err := workspace.Resolve(cfg, wsName, all)
            if err != nil { return err }

            exec := workspace.OsExecutor{}
            rep := workspace.NewReporter(os.Stdout, os.Stderr)
            opts := workspace.RunOptions{DryRun: dryRun, Verbose: verbose, Parallel: parallel}

            for _, ws := range workspaces {
                if err := workspace.Init(ws, exec, rep, opts); err != nil {
                    return err
                }
            }

            label := "cloned"
            if dryRun { label = "would clone" }
            rep.Summary(label, dryRun)
            os.Exit(rep.ExitCode())
            return nil
        },
    }

    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
    cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
    cmd.Flags().BoolVar(&parallel, "parallel", false, "Clone repositories concurrently")
    cmd.Flags().StringVarP(&wsName, "ws", "w", "", "Target workspace name")
    cmd.Flags().BoolVar(&all, "all", false, "Operate on all workspaces")

    return cmd
}
```

**Closure over local variables:** `dryRun`, `verbose`, etc. are declared *outside* the `RunE` function but referenced *inside* it. When `RunE` runs, it reads the values of those variables at that moment — after Cobra has already parsed the command-line flags and written into them via `cmd.Flags().BoolVar(&dryRun, ...)`. This is why the flag registration passes `&dryRun` (the address of the variable): Cobra writes the parsed flag value directly into that variable.

**`RunE` vs `Run`:** Cobra commands can use `Run` (no error return) or `RunE` (returns error). `RunE` is preferred because Cobra automatically prints the error and exits with a non-zero code if RunE returns a non-nil error.

**`os.Exit(rep.ExitCode())`:** `os.Exit` terminates the process immediately with the given code. It bypasses all deferred functions — so it's called only at the very end, after cleanup is complete. Exit code 0 = success; 1 = at least one repository failed.

---

## 8. How Data Flows — End-to-End

Here is what happens when you run `zgard ws init --ws myws --parallel`:

```
main()
  └─ Execute()
       └─ rootCmd.Execute()       [Cobra parses args and flags]
            └─ ws init RunE()
                 ├─ resolveConfigPath()           → "$HOME/.grazhda/config.yaml"
                 ├─ config.Load(path)             → *Config
                 ├─ config.Validate(cfg)          → [] (no errors)
                 ├─ workspace.Resolve(cfg, "myws", false) → []Workspace{myws}
                 ├─ workspace.OsExecutor{}
                 ├─ workspace.NewReporter(stdout, stderr)
                 └─ workspace.Init(myws, exec, rep, opts{Parallel:true})
                      ├─ rep.PrintLine("Workspace: myws")
                      ├─ for project in myws.Projects:
                      │    ├─ rep.PrintLine("  Project: backend")
                      │    ├─ os.MkdirAll(projPath)
                      │    └─ [goroutine] cloneRepo(api)
                      │       [goroutine] cloneRepo(auth)
                      │       wg.Wait()
                      └─ return nil
                 rep.Summary("cloned", false)   → prints "✓ 2 cloned  ⏭ 0 skipped  ✗ 0 failed"
                 os.Exit(0)
```

---

## 9. Testing in Go

### Test Files

Test files have the suffix `_test.go`. They are compiled only during `go test`, never in production builds.

A test file can be in the same package as the code it tests (`package workspace`) — giving access to unexported functions — or in an external test package (`package workspace_test`) — testing only the public API. Grazhda uses `package workspace_test` for its workspace tests, which is the preferred style for testing public behaviour.

### Test Functions

```go
func TestInit_ClonesRepos(t *testing.T) {
    ws, _ := makeWorkspace(t)
    var out, errOut strings.Builder
    rep := workspace.NewReporter(&out, &errOut)
    mock := &workspace.MockExecutor{}

    err := workspace.Init(ws, mock, rep, workspace.RunOptions{})
    if err != nil {
        t.Fatalf("Init error: %v", err)
    }

    if len(mock.Calls) != 2 {
        t.Errorf("expected 2 clone calls, got %d", len(mock.Calls))
    }
}
```

- `t *testing.T` is the test context provided by the framework
- `t.Fatalf` fails the test immediately and prints the message
- `t.Errorf` records a failure but continues running the test
- `strings.Builder` captures what the Reporter writes so we can assert on it
- `MockExecutor.Calls` records which commands were "run" without touching disk

### Running Tests

```bash
cd internal && go test ./...    # all packages under internal/
cd zgard    && go test ./...    # all packages under zgard/
# or use the Justfile shortcut:
just test
```

`./...` is a wildcard that matches the current package and all sub-packages.

### Race Detector

```bash
cd internal && go test -race ./...
```

The `-race` flag instruments the binary to detect **data races** — bugs where two goroutines access the same memory concurrently without synchronisation. The `MockExecutor` mutex was added specifically because the race detector caught a concurrent write to `Calls` during parallel tests.

### Test Fixtures

`internal/testdata/*.yaml` contains example config files used by config tests. Using real files rather than inline strings makes tests readable and lets you validate realistic configs.

---

## 10. Build System (Justfile)

`Justfile` is like a Makefile but uses the `just` tool. Each recipe is a named set of shell commands:

```
build-zgard:
    mkdir -p bin
    cd zgard && go build -o ../bin/zgard .

test:
    cd internal && go test ./...
    cd zgard && go test ./...

fmt:
    cd internal && go fmt ./...
    cd zgard && go fmt ./...

tidy:
    go work sync
    cd internal && go mod tidy
    cd zgard && go mod tidy
```

| Command | What it does |
| :--- | :--- |
| `just build-zgard` | Compiles `bin/zgard` |
| `just test` | Runs all unit tests |
| `just fmt` | Auto-formats Go source (`gofmt`) |
| `just tidy` | Syncs `go.work` and updates `go.mod`/`go.sum` |

---

## 11. Go Concepts Glossary

| Term | Plain-English meaning |
| :--- | :--- |
| **package** | A directory of `.go` files that share a package name and can see each other's unexported identifiers |
| **module** | A tree of packages with a `go.mod` file declaring the module path and dependencies |
| **interface** | A named set of method signatures; any type with those methods satisfies the interface |
| **struct** | A composite type grouping named fields, like a record or object |
| **pointer** (`*T`) | A value that holds the memory address of a `T`; used to share and mutate values |
| **slice** (`[]T`) | A dynamically-sized list of `T` values |
| **map** (`map[K]V`) | A hash table mapping keys of type `K` to values of type `V` |
| **goroutine** | A lightweight concurrent task started with `go func()` |
| **channel** | A typed pipe for communicating between goroutines (not used in this project) |
| **`sync.Mutex`** | A mutual exclusion lock — only one goroutine can hold it at a time |
| **`sync.WaitGroup`** | A counter that blocks until all goroutines have called `Done()` |
| **`defer`** | Schedules a function call to run when the surrounding function returns |
| **`error`** | A built-in interface; `nil` means no error; non-nil means something went wrong |
| **`io.Writer`** | Standard interface for anything that can accept bytes (stdout, files, buffers…) |
| **`io.Reader`** | Standard interface for anything that can produce bytes (stdin, strings, files…) |
| **`:=`** | Short variable declaration — declares and assigns in one step |
| **`_`** | Blank identifier — discards a value the compiler would otherwise require you to use |
| **struct tag** (`yaml:"name"`) | Metadata on a struct field read by libraries like the YAML decoder |
| **`init()`** | A special function that runs automatically before `main()` |
| **exported identifier** | Any identifier starting with an uppercase letter — visible outside the package |
| **unexported identifier** | Any identifier starting with a lowercase letter — private to the package |
