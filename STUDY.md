# Grazhda — Study Guide for Junior Go Developers

This guide explains the Grazhda project from first principles. You do not need to know Go before reading this. By the end you will understand how the repository is organised, what every file does, why the code is written the way it is, and a broad set of Go language concepts — all illustrated with the actual project code.

---

## Table of Contents

1. [What is Grazhda?](#1-what-is-grazhda)
2. [The Go Module System](#2-the-go-module-system)
3. [Project Directory Layout](#3-project-directory-layout)
4. [Go Language Fundamentals](#4-go-language-fundamentals)
   - [Variables and Types](#41-variables-and-types)
   - [Functions](#42-functions)
   - [Structs](#43-structs)
   - [Pointers](#44-pointers)
   - [Interfaces](#45-interfaces)
   - [Slices and Maps](#46-slices-and-maps)
   - [Error Handling](#47-error-handling)
   - [Goroutines and Concurrency](#48-goroutines-and-concurrency)
   - [defer](#49-defer)
   - [init()](#410-init)
   - [Visibility (exported vs unexported)](#411-visibility-exported-vs-unexported)
5. [internal/config — Loading Configuration](#5-internalconfig--loading-configuration)
6. [internal/executor — Running Shell Commands](#6-internalexecutor--running-shell-commands)
7. [internal/reporter — Showing Progress](#7-internalreporter--showing-progress)
8. [internal/workspace — The Core Domain](#8-internalworkspace--the-core-domain)
   - [options.go](#81-optionsgo)
   - [targeting.go — Choosing Workspaces](#82-targetinggo--choosing-workspaces)
   - [workspace.go — Init, Purge, Pull](#83-workspacego--init-purge-pull)
9. [zgard — The CLI Entry Point](#9-zgard--the-cli-entry-point)
   - [main.go and root.go](#91-maingo-and-rootgo)
   - [zgard/ws — Workspace Commands](#92-zgardws--workspace-commands)
10. [How Data Flows End-to-End](#10-how-data-flows-end-to-end)
11. [Testing in Go](#11-testing-in-go)
12. [Build System (Justfile)](#12-build-system-justfile)
13. [Configuration Reference](#13-configuration-reference)
14. [Go Concepts Glossary](#14-go-concepts-glossary)
15. [dukh — The Background Workspace Monitor](#15-dukh--the-background-workspace-monitor)
16. [gRPC and Protocol Buffers](#16-grpc-and-protocol-buffers)
17. [dukh/server — Server Packages](#17-dukhserver--server-packages)
    - [server.go — gRPC Lifecycle](#171-servergo--grpc-lifecycle)
    - [monitor.go — Configurable Background Polling](#172-monitorrgo--configurable-background-polling)
    - [log.go — Structured Logging with Rotation](#173-loggo--structured-logging-with-rotation)
18. [dukh CLI Commands](#18-dukh-cli-commands)
    - [dukh start — Self-Daemonizing](#181-dukh-start--self-daemonizing-background-process)
    - [dukh scan](#182-dukh-scan--triggering-an-immediate-rescan)
    - [dukh status — Process Health](#183-dukh-status--process-health)
    - [zgard ws status --rescan](#184-zgard-ws-status---rescan--synchronous-scan-before-status)
    - [Auto-Start — Transparent dukh Bootstrapping](#185-auto-start--transparent-dukh-bootstrapping)
    - [Configurable Monitoring Period](#186-configurable-monitoring-period)
19. [Updated Project Layout](#19-updated-project-layout)
20. [New Go Concepts Introduced in Phase 2](#20-new-go-concepts-introduced-in-phase-2)

---

## 1. What is Grazhda?

Grazhda is a **command-line tool** (CLI) called `zgard` that manages software development workspaces on your local machine.

A *workspace* in Grazhda terms is a directory that contains one or more *projects*, each of which contains one or more git repositories. You describe the structure once in a YAML file and then use `zgard` to:

- **`zgard ws init`** — create the directory structure and clone all repositories
- **`zgard ws pull`** — pull the latest changes in every repository
- **`zgard ws purge`** — delete a workspace directory

```
~/.grazhda/config.yaml describes:
  workspace "default"
    project "backend"
      repository "api"        → cloned to ~/ws/backend/api
      repository "auth"       → cloned to ~/ws/backend/auth
    project "frontend"
      repository "dashboard"  → cloned to ~/ws/frontend/dashboard
```

---

## 2. The Go Module System

### Packages

Every `.go` file begins with `package <name>`. A **package** is a group of files in the same directory. All files sharing a package can see each other's types and functions directly — no import needed.

```
internal/reporter/
├── reporter.go       ← package reporter
└── reporter_test.go  ← package reporter_test  (external test package)
```

### Modules

A **module** is a directory tree with a `go.mod` file at its root. The `go.mod` declares:
1. The module's canonical import path (used by other modules to import it)
2. The Go version being used
3. External dependency versions

Grazhda has two active modules:

| Directory | Module path (`go.mod` module line) |
| :--- | :--- |
| `internal/` | `github.com/vhula/grazhda/internal` |
| `zgard/` | `github.com/vhula/grazhda/zgard` |

To import the `reporter` package from `zgard/`, you write:

```go
import "github.com/vhula/grazhda/internal/reporter"
```

Go uses the module path as a namespace — `internal/reporter/reporter.go` says `package reporter` but is imported using the full path `github.com/vhula/grazhda/internal/reporter`.

### go.work — Local Multi-Module Workspace

Normally, when module A imports module B, Go downloads B from the internet (from a registry like `pkg.go.dev`). The `go.work` file at the project root tells Go to use local copies instead:

```
go 1.26.1

use (
    ./internal   ← use the local internal/ directory as the module
    ./zgard
    ./dukh
)
```

This lets the two modules reference each other during development without publishing anything online.

### The `internal/` Convention

Go has a language-enforced rule: any package inside a directory named `internal` can only be imported by code **within the parent of that `internal/` directory**. Since `internal/` is at the repository root, only code in this same repository can import it. This prevents external consumers of the module from depending on your private implementation details.

### go.mod and go.sum

`go.mod` lists direct and indirect dependencies with their versions:

```
require (
    github.com/spf13/cobra v1.9.1
    gopkg.in/yaml.v3 v3.0.1
)
```

`go.sum` is a lock file that stores cryptographic checksums of every downloaded dependency. You should commit both files; never edit `go.sum` by hand.

---

## 3. Project Directory Layout

```
grazhda/
│
├── go.work                      ← ties local modules together
├── Justfile                     ← `just build-zgard`, `just test`, etc.
├── config.template.yaml         ← example config; copy to ~/.grazhda/config.yaml
├── README.md
├── STUDY.md                     ← this file
│
├── internal/                    ← module: github.com/vhula/grazhda/internal
│   ├── go.mod
│   ├── go.sum
│   │
│   ├── config/
│   │   ├── config.go            ← Load, Validate, DefaultWorkspace, RenderCloneCmd
│   │   └── config_test.go       ← 14 unit tests
│   │
│   ├── executor/
│   │   ├── executor.go          ← Executor interface + OsExecutor
│   │   └── mock.go              ← MockExecutor for tests
│   │
│   ├── color/
│   │   └── color.go             ← Green/Red/Yellow/Blue helpers (wraps fatih/color)
│   │
│   ├── reporter/
│   │   ├── reporter.go          ← Reporter: ✓/⏭/✗ output + summary
│   │   └── reporter_test.go     ← 9 unit tests
│   │
│   ├── workspace/
│   │   ├── options.go           ← RunOptions struct
│   │   ├── targeting.go         ← Resolve: picks workspaces from flags
│   │   ├── workspace.go         ← Init, Purge, Pull
│   │   ├── workspace_test.go    ← 12 unit tests
│   │   └── targeting_test.go    ← 7 unit tests
│   │
│   └── testdata/                ← YAML fixtures used by tests
│       ├── valid_single_workspace.yaml
│       ├── valid_multi_workspace.yaml
│       ├── duplicate_workspace_names.yaml
│       ├── missing_required_fields.yaml
│       ├── missing_branch.yaml
│       └── invalid_template.yaml
│
└── zgard/                       ← module: github.com/vhula/grazhda/zgard
    ├── go.mod
    ├── go.sum
    ├── main.go                  ← func main() — program entry point
    ├── root.go                  ← zgard root Cobra command + Execute()
    └── ws/
        ├── ws.go                ← "zgard ws" parent command
        ├── init.go              ← "zgard ws init" with all flags
        ├── purge.go             ← "zgard ws purge" with all flags
        ├── pull.go              ← "zgard ws pull" with all flags
        ├── config.go            ← resolveConfigPath() helper
        ├── confirm.go           ← confirm() prompt helper
        └── ws_test.go           ← placeholder test
```

**Logical grouping rationale:**

| Package | Why it is separate |
| :--- | :--- |
| `config` | Config concerns are standalone — load YAML, validate, render templates. No dependency on executor or reporter. |
| `executor` | Shell execution is a generic concern, reusable for any future command (not just workspace). |
| `reporter` | Progress output is a generic concern, reusable for any future zgard command. |
| `workspace` | Domain logic — orchestrates executor and reporter to implement Init/Purge/Pull. |
| `zgard/ws` | CLI layer only — parse flags, resolve config, call workspace functions, call `os.Exit`. |

---

## 4. Go Language Fundamentals

### 4.1 Variables and Types

Go is **statically typed** — every variable has a fixed type known at compile time.

```go
// var keyword with explicit type
var name string = "grazhda"

// Short declaration — type is inferred
path := "/home/jake/ws"

// Multiple assignment
count, err := someFunction()

// Zero values — Go initialises all variables
var n int      // 0
var s string   // ""
var b bool     // false
var p *int     // nil (a nil pointer)
```

**Built-in types used in Grazhda:**

| Type | Description | Example |
| :--- | :--- | :--- |
| `string` | Immutable sequence of UTF-8 bytes | `"api"` |
| `bool` | True or false | `true`, `false` |
| `int` | Platform-width integer (64-bit on 64-bit OS) | `42` |
| `error` | Built-in interface for errors | `nil` or `errors.New("msg")` |
| `[]T` | Slice: a variable-length list of T | `[]string{"a", "b"}` |
| `map[K]V` | Hash map from K to V | `map[string]bool{}` |
| `*T` | Pointer to a value of type T | `&config.Config{}` |

### 4.2 Functions

```go
// Basic function: name, parameters (name type), return types
func add(a int, b int) int {
    return a + b
}

// Multiple return values — Go's idiomatic way to return results + errors
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

// Calling a function with multiple returns
result, err := divide(10, 3)
if err != nil {
    // handle error
}

// Discarding a return value with _
result, _ := divide(10, 3)  // we don't care about the error
```

**Named return values** (used occasionally):

```go
func split(s string) (head, tail string) {
    head = s[:1]
    tail = s[1:]
    return   // "naked return" — returns head and tail
}
```

### 4.3 Structs

A **struct** groups named fields into a composite type:

```go
type Workspace struct {
    Name    string
    Path    string
    Default bool
}

// Creating a struct value
ws := Workspace{Name: "default", Path: "/home/jake/ws", Default: true}

// Accessing fields
fmt.Println(ws.Name)  // "default"
ws.Path = "/tmp/ws"   // modify a field
```

**Struct tags** are string metadata attached to fields. The YAML library reads the `yaml:"..."` tag to know which YAML key maps to which field:

```go
type Repository struct {
    Name         string `yaml:"name"`
    Branch       string `yaml:"branch,omitempty"`   // optional field
    LocalDirName string `yaml:"local_dir_name,omitempty"`
}
```

Without the tag, `yaml.Unmarshal` would look for a key named `LocalDirName` (the field name) instead of `local_dir_name`.

### 4.4 Pointers

By default Go passes values **by copy**. A pointer holds the memory *address* of a value:

```go
// *Config means "a pointer to a Config"
func Load(path string) (*Config, error) {
    var cfg Config
    // ... fill in cfg ...
    return &cfg, nil   // & takes the address of cfg
}

cfg, err := Load("config.yaml")
// cfg is *Config — a pointer
fmt.Println(cfg.Name)   // Go auto-dereferences: cfg.Name is short for (*cfg).Name
```

**When to use pointers:**
- The function needs to *modify* the caller's value (e.g. `Reporter.Record`)
- The struct is large and copying it would be expensive
- `nil` is a meaningful "no value" (e.g. optional config fields)

**Pointer receiver vs value receiver:**

```go
// Value receiver — gets a copy; modifications don't affect the original
func (ws Workspace) DisplayName() string {
    return ws.Name
}

// Pointer receiver — gets the original; can modify it; also more efficient for large structs
func (r *Reporter) Record(res OpResult) {
    r.results = append(r.results, res)  // modifies the Reporter
}
```

As a rule, if *any* method on a type uses a pointer receiver, all methods should use a pointer receiver for consistency.

### 4.5 Interfaces

An **interface** defines behaviour as a set of method signatures. Any type that implements those methods *automatically* satisfies the interface — no explicit declaration required:

```go
// Interface declaration
type Executor interface {
    Run(dir string, command string) error
}

// OsExecutor satisfies Executor because it has a Run method with the right signature
type OsExecutor struct{}

func (e OsExecutor) Run(dir string, command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = dir
    return cmd.Run()
}

// MockExecutor also satisfies Executor
type MockExecutor struct{ Calls []string }

func (m *MockExecutor) Run(dir string, command string) error {
    m.Calls = append(m.Calls, command)
    return nil
}

// A function accepting the interface works with EITHER type
func Init(ws config.Workspace, exec Executor, ...) error {
    // exec can be OsExecutor or MockExecutor — Init doesn't know or care
}
```

This is how Go achieves **testability without mocking frameworks** — define an interface, write a real implementation and a fake implementation, inject the fake in tests.

**The standard library uses interfaces everywhere:**

```go
type io.Writer interface {
    Write(p []byte) (n int, err error)
}

// os.File satisfies io.Writer (you can write to a file)
// strings.Builder satisfies io.Writer (you can write to a string buffer)
// bytes.Buffer satisfies io.Writer

// Reporter accepts io.Writer, so it works with both:
rep := reporter.NewReporter(os.Stdout, os.Stderr)          // production
rep := reporter.NewReporter(&strings.Builder{}, &strings.Builder{})  // tests
```

### 4.6 Slices and Maps

**Slices** are dynamically-sized views over arrays:

```go
// Literal
repos := []string{"api", "auth", "dashboard"}

// Length and access
fmt.Println(len(repos))   // 3
fmt.Println(repos[0])     // "api"
fmt.Println(repos[1:])    // ["auth", "dashboard"] — sub-slice

// append — returns a new slice (may allocate new backing array)
repos = append(repos, "gateway")

// make — creates a slice with a given length and capacity
errs := make([]string, 0, 10)  // empty slice, capacity 10 — avoids reallocation

// Range — iterate over a slice
for i, repo := range repos {
    fmt.Println(i, repo)
}

// Ignore the index
for _, repo := range repos {
    fmt.Println(repo)
}
```

**Maps** are hash tables:

```go
// make — creates an empty map
seenWS := make(map[string]bool)

// Set
seenWS["default"] = true

// Get — returns the value and an "ok" bool
val, ok := seenWS["default"]   // val=true, ok=true
val, ok = seenWS["other"]      // val=false, ok=false

// Simpler read (zero value if key missing)
if seenWS["default"] {
    // already seen
}

// Delete
delete(seenWS, "default")

// Range
for name, seen := range seenWS {
    fmt.Println(name, seen)
}
```

### 4.7 Error Handling

Go has no exceptions. Functions return `error` as the last return value:

```go
// nil means "no error happened"
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        // Wrap the error with context using %w
        return nil, fmt.Errorf("config file %q: %w", path, err)
    }
    return &cfg, nil
}

// The caller must check the error
cfg, err := config.Load(cfgPath)
if err != nil {
    return err   // propagate up
}
```

**`fmt.Errorf` with `%w`** wraps the original error. Callers can later use `errors.Is` or `errors.As` to inspect the chain:

```go
if errors.Is(err, os.ErrNotExist) {
    // the original error was "file not found"
}
```

**When to return vs when to log:** In library code (`internal/`), always return errors — let the caller decide what to do. In CLI code (`zgard/ws/`), errors from `RunE` are printed by Cobra automatically.

### 4.8 Goroutines and Concurrency

A **goroutine** is a lightweight thread managed by the Go runtime. You can run thousands simultaneously:

```go
// Start a goroutine with the go keyword
go func() {
    fmt.Println("running concurrently")
}()

// The anonymous function executes concurrently with the caller
```

**Why goroutines?** For `--parallel`, Grazhda clones repositories at the same time to save wall-clock time. Without goroutines, clones would run one after another.

**sync.WaitGroup** — waiting for goroutines to finish:

```go
var wg sync.WaitGroup

for _, repo := range proj.Repositories {
    wg.Add(1)          // increment counter BEFORE starting goroutine
    repo := repo       // capture: give each goroutine its own copy of repo
    go func() {
        defer wg.Done()  // decrement counter when goroutine finishes
        cloneRepo(repo)
    }()
}

wg.Wait()  // block until counter reaches zero
```

**sync.Mutex** — preventing data races:

When multiple goroutines read and write the same memory simultaneously, the result is undefined — this is called a **data race** and is a serious bug. A `Mutex` (mutual exclusion lock) ensures only one goroutine can be in a critical section at a time:

```go
type MockExecutor struct {
    mu    sync.Mutex
    Calls []string
}

func (m *MockExecutor) Run(dir, command string) error {
    m.mu.Lock()           // acquire the lock — other goroutines block here
    m.Calls = append(m.Calls, command)
    m.mu.Unlock()         // release the lock
    return nil
}
```

A simpler pattern using `defer`:

```go
func (r *Reporter) Record(res OpResult) {
    r.mu.Lock()
    defer r.mu.Unlock()   // guaranteed to run when Record returns
    r.results = append(r.results, res)
    // ... do more work — mutex is still held
}
```

**Race detector:** Run `go test -race ./...` to have the Go toolchain instrument your binary and detect races at runtime.

**Loop variable capture (a common gotcha):**

```go
for _, repo := range proj.Repositories {
    go func() {
        fmt.Println(repo.Name)  // BUG in older Go: all goroutines share the same `repo`
    }()
}

// Fix: shadow the variable inside the loop
for _, repo := range proj.Repositories {
    repo := repo   // new `repo` variable scoped to this iteration
    go func() {
        fmt.Println(repo.Name)  // each goroutine captures its own copy
    }()
}
```

Go 1.22+ fixed this for `for` loops, but the explicit capture is still idiomatic for clarity.

### 4.9 defer

`defer` schedules a function call to run when the surrounding function returns — regardless of how it returns (normal return, early return, panic):

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()   // will run when processFile returns, no matter what

    // ... work with f ...
    return nil
}
```

**Common patterns:**

```go
// 1. Unlock a mutex
r.mu.Lock()
defer r.mu.Unlock()

// 2. Close resources
defer f.Close()
defer conn.Close()

// 3. Rollback on failure (used in cloneRepo)
var success bool
defer func() {
    if !success {
        os.RemoveAll(repoPath)  // clean up partial clone
    }
}()
// ... try to clone ...
success = true   // if we reach here, defer does nothing
```

Multiple defers run in **LIFO** (last in, first out) order — the last deferred function runs first.

### 4.10 init()

`init()` is a special function that the Go runtime calls automatically after package-level variables are initialised, before `main()`:

```go
// zgard/root.go
var rootCmd = &cobra.Command{...}   // package-level variable, initialised first

func init() {
    rootCmd.AddCommand(ws.NewCmd())   // then init() runs
}

func main() {
    Execute()   // then main() runs
}
```

Rules:
- Multiple files in the same package can each have an `init()` function — all of them run
- `init()` cannot be called explicitly
- `init()` cannot take parameters or return values
- Use it for one-time setup that depends on other package-level variables being initialised first

### 4.11 Visibility (exported vs unexported)

Go uses naming convention for access control — no `public`/`private` keywords:

| First letter | Visibility | Example |
| :--- | :--- | :--- |
| Uppercase | **Exported** — visible outside the package | `NewReporter`, `OpResult`, `Executor` |
| Lowercase | **Unexported** — private to the package | `expandHome`, `cloneRepo`, `mu` |

This applies to functions, types, struct fields, constants, and variables:

```go
type Reporter struct {
    out     io.Writer    // unexported — tests can't access this directly
    errOut  io.Writer    // unexported
    mu      sync.Mutex  // unexported
    results []OpResult  // unexported
}

// Exported constructor — the only way to create a Reporter from outside the package
func NewReporter(out, errOut io.Writer) *Reporter {
    return &Reporter{out: out, errOut: errOut}
}

// Exported methods — form the public API
func (r *Reporter) Record(res OpResult) { ... }
func (r *Reporter) Summary(label string, dryRun bool) { ... }
func (r *Reporter) ExitCode() int { ... }
```

**Test packages:** A file with `package reporter_test` (note the `_test` suffix) is an *external test package*. It can only access exported identifiers, just like any other package. This is useful for testing the public API in isolation. Files with `package reporter` inside a `_test.go` file are *internal* test files and can access unexported identifiers.

---

## 5. internal/config — Loading Configuration

**File:** `internal/config/config.go`

### The Type Hierarchy

The configuration file describes a tree: Config → Workspaces → Projects → Repositories. The Go types mirror this exactly:

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

The `yaml:"..."` struct tags tell the `gopkg.in/yaml.v3` library which YAML key to look for. Without the `local_dir_name` tag, the decoder would look for a key `LocalDirName` (the exact field name) which would not match the snake_case YAML convention.

`omitempty` means: when marshalling to YAML, omit the field if it is the zero value (empty string). When unmarshalling, the field is simply not set if the key is absent.

### Load

```go
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("config file %q: %w", path, err)
    }
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config %q: %w", path, err)
    }
    return &cfg, nil
}
```

`os.ReadFile` reads the entire file into a `[]byte` (a slice of bytes). `yaml.Unmarshal` parses those bytes and fills the `cfg` struct by matching YAML keys to struct field tags. `&cfg` passes a pointer so the decoder can write into it.

### DefaultWorkspace

```go
func DefaultWorkspace(cfg *Config) (*Workspace, error) {
    for i := range cfg.Workspaces {
        ws := &cfg.Workspaces[i]   // pointer to slice element, not a copy
        if ws.Default || ws.Name == "default" {
            return ws, nil
        }
    }
    return nil, fmt.Errorf("no default workspace found: add a workspace with name: default or use --name")
}
```

Why `for i := range` instead of `for _, ws := range`? Because `for _, ws := range cfg.Workspaces` gives you a **copy** of each element, and `&ws` would give the address of the copy, not the original slice element. Using `&cfg.Workspaces[i]` gives the address of the actual slice element.

### Validate

Validation runs up-front — all errors are collected and reported before any filesystem change:

```go
func Validate(cfg *Config) []string {
    var errs []string
    seenWS := make(map[string]bool)

    for i, ws := range cfg.Workspaces {
        if ws.Name == "" {
            errs = append(errs, fmt.Sprintf("workspace[%d].name: required field missing", i))
        } else if seenWS[ws.Name] {
            errs = append(errs, fmt.Sprintf("workspace names must be unique: duplicate name %q", ws.Name))
        } else {
            seenWS[ws.Name] = true
        }
        // ... more checks for path, clone_command_template, projects, branches ...
    }
    return errs
}
```

Returning `[]string` (a slice of error messages) rather than `error` allows reporting *all* problems at once instead of stopping at the first one. In the CLI commands, the caller prints each message then returns a single summary error:

```go
if errs := config.Validate(cfg); len(errs) > 0 {
    for _, e := range errs {
        fmt.Fprintln(os.Stderr, e)
    }
    return fmt.Errorf("configuration is invalid")
}
```

### RenderCloneCmd

This function takes a Go template string like:

```
git clone --branch {{.Branch}} https://github.com/org/{{.RepoName}} {{.DestDir}}
```

and fills it in with real values for a specific repository. The `destDir` (full filesystem path) is now computed by the workspace layer (see `ResolveDestName` below) and passed in directly:

```go
func RenderCloneCmd(tmplStr string, proj Project, repo Repository, destDir string) (string, error) {
    branch := repo.Branch
    if branch == "" {
        branch = proj.Branch   // fall back to project-level branch
    }
    data := CloneTemplateData{
        Branch:   branch,
        RepoName: repo.Name,
        DestDir:  destDir,     // pre-computed by workspace.ResolveDestName
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

`text/template` is a standard library package. `{{.Branch}}` refers to the `Branch` field of the data struct. `bytes.Buffer` is an in-memory byte buffer — `t.Execute` writes rendered output into it, and `buf.String()` converts it to a string.

### ResolveDestName — Workspace Structure Modes

Repository names sometimes contain slashes — e.g. `org/team/repo` for namespaced package registries. The `structure` field on a workspace controls how such names are mapped to local directories.

```go
// StructureTree: "org/team/repo" → <project>/org/team/repo  (nested dirs)
// StructureList: "org/team/repo" → <project>/repo           (last segment, .git stripped)
func ResolveDestName(_ /*projPath*/ string, repoName, localDirName, structure string) string {
    if localDirName != "" {
        return localDirName   // explicit override always wins
    }
    if structure == config.StructureList {
        return lastSegment(repoName)  // strips .git, returns after last "/"
    }
    return repoName           // tree mode (default): use full name as-is
}

func lastSegment(name string) string {
    name = strings.TrimSuffix(name, ".git")
    if idx := strings.LastIndex(name, "/"); idx >= 0 {
        return name[idx+1:]
    }
    return name
}
```

**`list` mode** always returns the last `/`-delimited segment, stripping any `.git` suffix:

| Input `repo.Name` | Result |
|---|---|
| `"org/pack/repo"` | `"repo"` |
| `"org/pack/repo.git"` | `"repo"` |
| `"repo"` (no slash) | `"repo"` |

If two repos share the same last segment (e.g. `org/api` and `other/api`), both resolve to `api`. The second clone will be **skipped** as "already exists" by `cloneRepo`. Use `local_dir_name` to resolve naming conflicts explicitly.

`localDirName` always overrides `structure`.

---

## 6. internal/executor — Running Shell Commands

**Files:** `internal/executor/executor.go`, `internal/executor/mock.go`

The executor package is deliberately generic — it knows nothing about Grazhda's workspace concept. It can run any shell command in any directory.

### The Executor Interface

```go
type Executor interface {
    Run(dir string, command string) error
}
```

One method. Any type with `Run(dir string, command string) error` satisfies `Executor`. This interface is the *contract* between:
- The workspace package (the *consumer*) which needs to run shell commands
- The executor package (the *provider*) which actually runs them

### OsExecutor

```go
type OsExecutor struct{}

func (e OsExecutor) Run(dir string, command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.Dir = dir
    return cmd.Run()
}
```

`exec.Command("sh", "-c", command)` creates a `*exec.Cmd`. The `sh -c` invocation runs the full command string through the shell, so glob patterns, environment variables, and pipes all work.

`cmd.Dir = dir` sets the working directory for the subprocess.

`cmd.Run()` starts the process and waits for it to exit. It returns:
- `nil` if the process exited with code 0 (success)
- An `*exec.ExitError` if the process exited with a non-zero code

### MockExecutor

```go
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

`MockExecutor` records every call in `Calls`. Tests inspect this slice to assert that the right commands were run.

Two ways to inject errors:
- `mock.Err = errors.New("clone failed")` — every call fails
- `mock.ErrFn = func(call int) error { ... }` — per-call control (e.g. fail only the first call)

The mutex protects `Calls` because `--parallel` mode calls `Run` from multiple goroutines simultaneously. The `errFn` is read *inside* the lock and invoked *outside* — deliberately, so the error function itself is not constrained from acquiring other locks.

---

## 7. internal/reporter — Showing Progress

**Files:** `internal/reporter/reporter.go`, `internal/reporter/reporter_test.go`

The reporter package is also generic — it can display progress for any operation that produces a stream of named results. Like executor, it knows nothing about workspaces specifically.

### OpResult

```go
type OpResult struct {
    Workspace string
    Project   string
    Repo      string
    Skipped   bool
    Err       error
    Msg       string
}
```

Every repository operation produces one `OpResult`. The `Err` field distinguishes failure from success; `Skipped` marks a repository that was intentionally not operated on (e.g. already cloned).

### Reporter

```go
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

The constructor accepts `io.Writer` for both output streams. In production:

```go
rep := reporter.NewReporter(os.Stdout, os.Stderr)
```

In tests:

```go
var out, errOut strings.Builder
rep := reporter.NewReporter(&out, &errOut)
// after the operation, inspect out.String() and errOut.String()
```

This pattern is called **dependency injection** — the dependencies (`os.Stdout`, `os.Stderr`) are *injected* from outside instead of hard-coded. It makes the code testable without forking processes or capturing stdout.

### Record

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

    displayMsg := res.Msg
    if res.Err != nil && displayMsg == "" {
        displayMsg = res.Err.Error()
    }

    fmt.Fprintf(r.out, "    %s %-14s — %s\n", symbol, res.Repo, displayMsg)
}
```

`fmt.Fprintf(w, format, args...)` writes a formatted string to any `io.Writer` — in this case the reporter's `out`.

Format specifiers:
- `%s` — string
- `%d` — integer
- `%q` — quoted string (adds surrounding `"` and escapes special chars)
- `%-14s` — left-aligned string padded to 14 characters minimum

`res.Err.Error()` calls the `Error() string` method on the `error` interface, returning the error message as a plain string.

### Summary

```go
func (r *Reporter) Summary(successLabel string, dryRun bool) {
    r.mu.Lock()
    defer r.mu.Unlock()

    var success, skipped, failed int
    for _, res := range r.results {
        switch {
        case res.Err != nil:
            failed++
        case res.Skipped:
            skipped++
        default:
            success++
        }
    }

    prefix := ""
    if dryRun {
        prefix = "[DRY RUN] "
    }
    fmt.Fprintf(r.out, "\n%s✓ %d %s  ⏭ %d skipped  ✗ %d failed\n",
        prefix, success, successLabel, skipped, failed)

    for _, res := range r.results {
        if res.Err != nil {
            fmt.Fprintf(r.errOut, "      %s: %s\n", res.Repo, res.Err.Error())
        }
    }
}
```

The `successLabel` parameter allows the summary to read naturally for each command:

| Command | label (normal) | label (dry-run) |
| :--- | :--- | :--- |
| `ws init` | `"cloned"` | `"would clone"` |
| `ws pull` | `"pulled"` | `"would pull"` |
| `ws purge` | `"removed"` | `"would remove"` |

Failure details go to `errOut` (stderr) — following the Unix convention that diagnostics and errors go to stderr while normal output goes to stdout. This lets callers pipe stdout without capturing error messages.

---

## 8. internal/workspace — The Core Domain

The workspace package orchestrates the executor and reporter to implement three operations: Init, Purge, and Pull.

### 8.1 options.go

```go
type RunOptions struct {
    DryRun    bool
    Verbose   bool
    Parallel  bool
    NoConfirm bool
}
```

A **value object** — carries options from the CLI layer down to domain functions. Because it is a struct (not individual parameters), adding a new option in the future requires only changing this struct, not every function signature.

### 8.2 targeting.go — Choosing Workspaces

```go
func Resolve(cfg *config.Config, wsName string, all bool) ([]config.Workspace, error) {
    if wsName != "" && all {
        return nil, fmt.Errorf("--name and --all are mutually exclusive")
    }

    if all {
        return cfg.Workspaces, nil
    }

    if wsName != "" {
        for _, ws := range cfg.Workspaces {
            if ws.Name == wsName {
                return []config.Workspace{ws}, nil   // wrap in a 1-element slice
            }
        }
        return nil, fmt.Errorf("workspace %q not found in config", wsName)
    }

    // Default: no flags given — use the default workspace
    ws, err := config.DefaultWorkspace(cfg)
    if err != nil {
        return nil, err
    }
    return []config.Workspace{*ws}, nil   // dereference pointer to get value
}
```

`[]config.Workspace{ws}` is a **slice literal** — creates a slice with one element. All three CLI commands iterate over the returned slice, so they handle the single-workspace and all-workspaces cases identically.

`*ws` dereferences the pointer: `ws` is `*config.Workspace` (a pointer), so `*ws` is the `config.Workspace` value it points to.

### 8.3 workspace.go — Init, Purge, Pull

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

`path[0]` is the first **byte** of the string (Go strings are byte sequences, not rune sequences). `path[1:]` is a string slice from index 1 to the end — everything after the `~`.

#### Init

```go
func Init(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) error {
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
                repo := repo           // capture loop variable
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

`os.MkdirAll(path, perm)` creates the directory and all missing parent directories. `0o755` is an **octal integer literal** (the `0o` prefix) for Unix permission bits: owner can read/write/execute; group and others can read and execute.

#### cloneRepo — the defer/success pattern

```go
func cloneRepo(ws config.Workspace, proj config.Project, projPath string, repo config.Repository, exec executor.Executor, rep *reporter.Reporter, opts RunOptions) {
    // Resolve dest directory using workspace structure mode
    destName := workspace.ResolveDestName(projPath, repo.Name, repo.LocalDirName, ws.Structure)
    repoPath := filepath.Join(projPath, destName)

    // Skip if already cloned
    if _, err := os.Stat(repoPath); err == nil {
        rep.Record(reporter.OpResult{..., Skipped: true, Msg: "already exists, skipped"})
        return
    }

    cmd, err := config.RenderCloneCmd(ws.CloneCommandTemplate, proj, repo, repoPath)
    if err != nil { /* record error, return */ }

    if opts.Verbose {
        rep.PrintLine(fmt.Sprintf("  → %s", cmd))
    }

    // Rollback guard
    var success bool
    defer func() {
        if !success {
            os.RemoveAll(repoPath)   // delete partial clone on any failure
        }
    }()

    if err := exec.Run(projPath, cmd); err != nil {
        rep.Record(reporter.OpResult{..., Err: err})
        return   // success remains false — defer will clean up
    }

    success = true   // we reached here — defer does nothing
    rep.Record(reporter.OpResult{..., Msg: fmt.Sprintf("cloned (%s)", branch)})
}
```

The defer/success pattern is a clean way to implement transactional-style cleanup: if anything goes wrong after the directory is created, the partial clone is removed so the workspace is left in a consistent state.

`os.Stat(path)` returns info about a file/directory. When the path does not exist, it returns an error (specifically `*fs.PathError`). The `if _, err := os.Stat(repoPath); err == nil` idiom checks "does this path exist?" — `err == nil` means the stat succeeded, i.e. the path exists.

#### Purge

```go
func Purge(ws config.Workspace, rep *reporter.Reporter, opts RunOptions) error {
    wsPath := expandHome(ws.Path)

    if opts.DryRun {
        rep.PrintLine(fmt.Sprintf("[DRY RUN] would remove: %s", wsPath))
        rep.Record(reporter.OpResult{Workspace: ws.Name, Repo: ws.Name,
            Msg: fmt.Sprintf("[DRY RUN] would remove %s", wsPath)})
        return nil
    }

    if _, err := os.Stat(wsPath); os.IsNotExist(err) {
        rep.Record(reporter.OpResult{..., Skipped: true, Msg: "directory not found, skipped"})
        return nil
    }

    if err := os.RemoveAll(wsPath); err != nil {
        rep.Record(reporter.OpResult{..., Err: err})
        return nil
    }

    rep.Record(reporter.OpResult{..., Msg: fmt.Sprintf("removed %s", wsPath)})
    return nil
}
```

`os.IsNotExist(err)` is a helper that returns true if the error means "path does not exist". This is the idiomatic way to check for a missing file.

`os.RemoveAll(path)` deletes the entire directory tree — equivalent to `rm -rf`. It returns `nil` if the path does not exist (so you could skip the `os.Stat` check, but the explicit check allows reporting "skipped" vs "removed").

Note: `Purge` does *not* take an `Executor`. It uses only standard library file functions. This is simpler and more predictable — no shell parsing involved.

#### Pull

Pull mirrors Init but skips non-existent repos and runs `git pull --rebase`:

```go
cmd := fmt.Sprintf("git pull --rebase origin %s", branch)
if err := exec.Run(repoPath, cmd); err != nil {
    rep.Record(reporter.OpResult{..., Err: err})
    return
}
```

The pull command is constructed as a plain string, not a template, because `git pull` is always the same command structure — unlike clone which varies by hosting provider.

---

## 9. zgard — The CLI Entry Point

Grazhda uses [Cobra](https://github.com/spf13/cobra), a popular Go library for building CLIs. Cobra organises commands as a tree:

```
zgard                  ← root command (rootCmd)
└── ws                 ← parent command (ws.NewCmd())
    ├── init           ← subcommand (newInitCmd())
    ├── purge          ← subcommand (newPurgeCmd())
    └── pull           ← subcommand (newPullCmd())
```

Each command is a `*cobra.Command` struct with fields like `Use`, `Short`, `Long`, `RunE`, and a set of registered flags.

### 9.1 main.go and root.go

**`zgard/main.go`** — the entire file:

```go
package main

func main() {
    Execute()
}
```

`package main` is special — the Go toolchain looks for `func main()` in a `package main` file to start the program. `main.go` is intentionally empty of logic; all setup lives in `root.go`.

**`zgard/root.go`:**

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

`var rootCmd = &cobra.Command{...}` is a **package-level variable**. It is initialised before `main()` runs. Package-level variables cannot reference values that are computed at runtime (no function calls in the initialiser — except function calls that return a pointer to a constant-like value, which `&cobra.Command{...}` is).

`func init()` runs after package-level variables are set and before `main()`. It calls `ws.NewCmd()` to get the configured `ws` command and registers it under `rootCmd`.

`rootCmd.Execute()` parses `os.Args` (the command-line arguments), finds the matching command, validates flags, and calls the command's `RunE` function. If it returns an error, Cobra prints it automatically.

### 9.2 zgard/ws — Workspace Commands

#### ws.go

```go
package ws

func NewCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "ws",
        Short: "Workspace operations",
        Long:  "Manage workspace lifecycle: initialize, purge, or pull repositories.",
    }
    cmd.AddCommand(newInitCmd())
    cmd.AddCommand(newPurgeCmd())
    cmd.AddCommand(newPullCmd())
    return cmd
}
```

`NewCmd` is exported (uppercase). The three `new*Cmd()` functions are unexported — they are internal construction details of the `ws` package. This is a clean public API: callers get one function to call, and the internals are hidden.

#### config.go — resolving the config path

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

`os.Getenv("GRAZHDA_DIR")` returns the value of the environment variable, or `""` if it is not set. The `_` discards the error from `UserHomeDir` — if the home directory cannot be resolved, `dir` is empty and the path will be relative to the current directory (an acceptable degraded behaviour).

`filepath.Join(home, ".grazhda")` produces `$HOME/.grazhda` on any platform.

#### confirm.go — the interactive prompt

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

`bufio.NewScanner(reader)` wraps an `io.Reader` to read line by line. `scanner.Scan()` reads the next line from the input, returning `false` at EOF (or if the reader is a terminal and the user pressed Ctrl+D).

`strings.TrimSpace` removes leading/trailing whitespace (spaces, tabs, newlines). `strings.EqualFold` compares strings case-insensitively — `"y"`, `"Y"`, `"yes"`, `"YES"` all return true.

By accepting `io.Writer` and `io.Reader` as parameters instead of using `os.Stdout`/`os.Stdin` directly, `confirm` is testable: inject `strings.NewReader("y\n")` to simulate typing `y` + Enter.

#### init.go — a complete Cobra command

```go
func newInitCmd() *cobra.Command {
    // Flag variables declared in the enclosing scope
    var dryRun bool
    var verbose bool
    var parallel bool
    var wsName string
    var all bool

    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize a workspace by cloning all repositories",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Resolve config path
            cfgPath := resolveConfigPath()

            // 2. Load and validate config
            cfg, err := config.Load(cfgPath)
            if err != nil {
                return err
            }
            if errs := config.Validate(cfg); len(errs) > 0 {
                for _, e := range errs {
                    fmt.Fprintln(os.Stderr, e)
                }
                return fmt.Errorf("configuration is invalid")
            }

            // 3. Resolve target workspaces
            workspaces, err := workspace.Resolve(cfg, wsName, all)
            if err != nil {
                return err
            }

            // 4. Run the operation
            exec := executor.OsExecutor{}
            rep := reporter.NewReporter(os.Stdout, os.Stderr)
            opts := workspace.RunOptions{
                DryRun: dryRun, Verbose: verbose, Parallel: parallel,
            }
            for _, ws := range workspaces {
                if err := workspace.Init(ws, exec, rep, opts); err != nil {
                    return err
                }
            }

            // 5. Print summary and exit
            label := "cloned"
            if dryRun {
                label = "would clone"
            }
            rep.Summary(label, dryRun)
            os.Exit(rep.ExitCode())
            return nil
        },
    }

    // Register flags — each BoolVar/StringVarP writes into the local variable
    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
    cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
    cmd.Flags().BoolVar(&parallel, "parallel", false, "Clone repositories concurrently")
    cmd.Flags().StringVarP(&wsName, "ws", "w", "", "Target workspace name")
    cmd.Flags().BoolVar(&all, "all", false, "Operate on all workspaces")

    return cmd
}
```

**Closures and flag variables:** `dryRun`, `verbose`, etc. are declared *outside* the `RunE` function but used *inside* it. The inner function *closes over* those variables — this is a **closure**. When `RunE` runs (after Cobra has parsed the command line), the flag variables already hold the parsed values written there by Cobra via the `&dryRun` pointer.

**`BoolVar(&dryRun, "dry-run", false, "...")`** registers a `--dry-run` flag:
- `&dryRun` — pointer to the variable that receives the value
- `"dry-run"` — the long flag name
- `false` — the default value
- `"..."` — the help text

**`BoolVarP(&verbose, "verbose", "v", false, "...")`** — the `P` variant adds a short form `-v`.

**`os.Exit(rep.ExitCode())`** terminates the process immediately with the given code. It *bypasses* all deferred functions, so it must only be called after all cleanup is done (Cobra cannot run more code after it). Exit code 0 = success; 1 = at least one repo operation failed.

---

## 10. How Data Flows End-to-End

Here is the complete call chain for `zgard ws init --name myws --parallel`:

```
os.Args = ["zgard", "ws", "init", "--name", "myws", "--parallel"]

main()
└─ Execute()
     └─ rootCmd.Execute()                  [Cobra parses os.Args]
          └─ ws.RunE called
               │
               ├─ resolveConfigPath()
               │    └─ os.Getenv("GRAZHDA_DIR") or $HOME/.grazhda/config.yaml
               │
               ├─ config.Load(path)
               │    ├─ os.ReadFile(path)   → []byte
               │    └─ yaml.Unmarshal(...)  → *Config
               │
               ├─ config.Validate(cfg)     → [] (no errors)
               │
               ├─ workspace.Resolve(cfg, "myws", false)
               │    └─ finds ws with Name=="myws" → []Workspace{myws}
               │
               ├─ executor.OsExecutor{}
               ├─ reporter.NewReporter(os.Stdout, os.Stderr)
               ├─ workspace.RunOptions{Parallel: true}
               │
               └─ workspace.Init(myws, exec, rep, opts)
                    ├─ rep.PrintLine("Workspace: myws")
                    │
                    └─ for project "backend":
                         ├─ rep.PrintLine("  Project: backend")
                         ├─ os.MkdirAll("~/ws/backend", 0o755)
                         │
                         ├─ goroutine → cloneRepo(api)
                         │    ├─ os.Stat(repoPath)        → not found, proceed
                         │    ├─ config.RenderCloneCmd()  → "git clone --branch main ... api"
                         │    ├─ exec.Run(projPath, cmd)  → runs git clone
                         │    └─ rep.Record(OpResult{Repo:"api", Msg:"cloned (main)"})
                         │
                         └─ goroutine → cloneRepo(auth)
                              └─ ... (same, branch="dev")
                    wg.Wait()
               rep.Summary("cloned", false)
               → prints "✓ 2 cloned  ⏭ 0 skipped  ✗ 0 failed"
               os.Exit(0)
```

---

## 11. Testing in Go

### Test File Conventions

Test files end in `_test.go`. They are compiled only during `go test` — never in production builds.

```
reporter_test.go     ← only compiled when testing
```

A test file can declare one of two packages:

| Package | Access | Typical use |
| :--- | :--- | :--- |
| `package reporter` | Can access unexported identifiers | Testing internals |
| `package reporter_test` | Only exported identifiers | Testing the public API |

Grazhda uses `package reporter_test`, `package workspace_test`, etc. — testing the public API only. This ensures tests remain valid even if internal implementation changes.

### Test Functions

```go
func TestRecord_Success(t *testing.T) {
    // Arrange
    var out, errOut strings.Builder
    rep := reporter.NewReporter(&out, &errOut)

    // Act
    rep.Record(reporter.OpResult{Repo: "api", Msg: "cloned (main)"})

    // Assert
    if !strings.Contains(out.String(), "✓") {
        t.Errorf("expected success symbol, got: %q", out.String())
    }
}
```

All test functions must be named `Test<Something>` and take `*testing.T`:

| Method | Behaviour |
| :--- | :--- |
| `t.Fatalf("msg", ...)` | Fail test immediately and stop this test function |
| `t.Errorf("msg", ...)` | Record failure but continue running the test |
| `t.Helper()` | Mark this function as a helper so failures point to the caller |

**`strings.Builder`** is a string buffer that implements `io.Writer`. It is the idiomatic way to capture written output in tests.

### Table-Driven Tests

A common Go pattern for testing many inputs:

```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantErrs int
    }{
        {"valid", "valid.yaml", 0},
        {"missing name", "missing_name.yaml", 1},
        {"duplicate names", "duplicate.yaml", 1},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            cfg, _ := config.Load("../testdata/" + tc.input)
            errs := config.Validate(cfg)
            if len(errs) != tc.wantErrs {
                t.Errorf("expected %d errors, got %d: %v", tc.wantErrs, len(errs), errs)
            }
        })
    }
}
```

`t.Run(name, func)` creates a named sub-test. You can run a specific sub-test with `go test -run TestValidate/duplicate_names`.

### Helper Functions

```go
func makeWorkspace(t *testing.T) (config.Workspace, string) {
    t.Helper()   // marks this as a helper — failures point to the caller, not here
    tmp := t.TempDir()   // creates a temporary directory, cleaned up after the test
    ws := config.Workspace{
        Name:                 "test-ws",
        Path:                 tmp,
        CloneCommandTemplate: "echo clone {{.RepoName}} {{.DestDir}}",
        Projects: []config.Project{
            {Name: "backend", Branch: "main",
             Repositories: []config.Repository{
                 {Name: "api"},
                 {Name: "auth", Branch: "dev"},
             }},
        },
    }
    return ws, tmp
}
```

`t.TempDir()` creates a unique temporary directory that is automatically deleted when the test finishes. Never use hard-coded `/tmp/grazhda-test` — it breaks when tests run in parallel.

### Running Tests

```bash
go test ./...                     # all packages, from module root
go test ./workspace/...           # workspace package only
go test -run TestInit             # only tests matching "TestInit"
go test -race ./...               # with race detector
go test -v ./...                  # verbose: show each test name
go test -count=1 ./...            # disable test result cache (force re-run)
```

The `-race` flag instruments the binary to detect data races at runtime. It is slower (10–20x) but should be run before every commit in code with goroutines.

### Test Fixtures

`internal/testdata/*.yaml` contains example config files. Using fixture files:
- Makes test intent visible — you can read the YAML to understand what is being tested
- Avoids cluttering test code with multi-line strings
- Can be shared across multiple test functions

---

## 12. Build System (Justfile)

`Justfile` uses the `just` command runner — similar to `make` but simpler syntax. Each named recipe contains shell commands:

```just
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
    cd internal && go mod tidy
    cd zgard && go mod tidy -e
```

| Command | What it does |
| :--- | :--- |
| `just build-zgard` | Compiles `bin/zgard` (the production binary) |
| `just build` | Builds `bin/zgard` + copies `grazhda-install.sh` bash scripts to `bin/` |
| `just test` | Runs `go test ./...` for both modules |
| `just fmt` | Auto-formats all `.go` files with `gofmt` (the standard formatter) |
| `just tidy` | `go mod tidy` for each module — removes unused dependencies and adds missing ones |
| `just clean` | Removes the `bin/` directory |

**`go build -o ../bin/zgard .`** compiles the package in the current directory (`.`) and writes the binary to `../bin/zgard`. The `-o` flag specifies the output file name.

**`go fmt ./...`** formats every `.go` file in the current module. Gofmt is opinionated — it enforces a single canonical style so there is no debate about formatting. In Go, formatting is not style; it is mandatory.

**`go mod tidy`** removes unused dependencies from `go.mod` and `go.sum`, and adds any missing ones. Run it after adding or removing imports. The `zgard` module uses `go mod tidy -e` (tolerate errors) because it depends on the local `internal` module which is resolved via `go.work` and not published to the Go module registry.

---

## 13. Configuration Reference

The config file location is resolved in this order:
1. `$GRAZHDA_DIR/config.yaml` (when the environment variable is set)
2. `~/.grazhda/config.yaml` (the default)

### Full annotated example

```yaml
workspaces:
  # First workspace: the default — uses tree structure (nested dirs for slashed names)
  - name: default                        # unique identifier, used with --name
    default: true                        # mark as default (or just name it "default")
    path: /home/jake/ws                  # workspace root directory
    structure: tree                      # "tree" (default) or "list" — see below
    clone_command_template: >            # YAML folded scalar — newlines become spaces
      git clone --branch {{.Branch}}
      https://github.com/myorg/{{.RepoName}}
      {{.DestDir}}
    projects:
      - name: backend                    # project directory: /home/jake/ws/backend
        branch: main                     # default branch for repos in this project
        repositories:
          - name: api                    # cloned to /home/jake/ws/backend/api
          - name: auth
            branch: dev                  # overrides project branch for this repo only
          - name: api
            local_dir_name: api-v2       # cloned to /home/jake/ws/backend/api-v2
          - name: org/pack/repo          # tree mode → /home/jake/ws/backend/org/pack/repo

  # Second workspace: personal projects, list structure (flat clone dirs)
  - name: personal
    path: /home/jake/personal
    structure: list                      # last URL segment used as dest dir
    clone_command_template: "git clone git@github.com:jake/{{.RepoName}} {{.DestDir}}"
    projects:
      - name: tools
        branch: main
        repositories:
          - name: dotfiles
          - name: scripts
          - name: org/pack/repo          # list mode → /home/jake/personal/tools/repo
```

### Template Variables

| Variable | Resolves to |
| :--- | :--- |
| `{{.Branch}}` | `repository.branch` if set, otherwise `project.branch` |
| `{{.RepoName}}` | `repository.name` (full value, slashes included) |
| `{{.DestDir}}` | Full filesystem path to the clone destination (computed by `ResolveDestName`) |

### Workspace Structure Modes

| `structure` value | Behaviour | Example: `org/pack/repo.git` |
| :--- | :--- | :--- |
| `tree` *(default)* | Full name as nested dirs | `<project>/org/pack/repo.git` |
| `list` | Last `/`-delimited segment, `.git` stripped | `<project>/repo` |

`local_dir_name` on a repository always wins over `structure`. In `list` mode, two repos with the same last segment resolve to the same path; the second will be skipped as "already exists". Use `local_dir_name` to disambiguate.

---

## 14. Go Concepts Glossary

| Term | Plain-English meaning |
| :--- | :--- |
| **package** | A directory of `.go` files sharing a package name; files in the same package can see each other's unexported identifiers |
| **module** | A tree of packages with a `go.mod` at the root; the unit of versioning and distribution |
| **import path** | The string used in `import "..."` — for external modules, this is the module path from `go.mod` plus any sub-package path |
| **go.work** | A workspace file that maps module paths to local directories, enabling multi-module development without publishing |
| **internal/** | A directory name enforced by Go: packages inside `internal/` can only be imported by code in the parent of `internal/` |
| **interface** | A named set of method signatures; any type with those methods satisfies the interface automatically |
| **struct** | A composite type grouping named fields |
| **struct tag** | Backtick-quoted metadata on a struct field, read by reflection (e.g. `yaml:"name"`) |
| **pointer** (`*T`) | A value holding the memory address of a `T`; `&x` takes the address of `x`; `*p` dereferences pointer `p` |
| **slice** (`[]T`) | A dynamically-sized list; backed by an array but with its own length and capacity |
| **map** (`map[K]V`) | A hash table; zero value is `nil`; must be initialised with `make(map[K]V)` before writing |
| **goroutine** | A lightweight concurrent task started with `go func(){}()`; managed by the Go runtime |
| **channel** | A typed pipe for communicating between goroutines; not used in this project |
| **sync.Mutex** | A mutual exclusion lock; only one goroutine can hold it at a time |
| **sync.WaitGroup** | A counter that blocks `Wait()` until all goroutines have called `Done()` |
| **data race** | A bug where two goroutines read/write the same memory concurrently without synchronisation; detected with `go test -race` |
| **defer** | Schedules a call to run when the enclosing function returns, in LIFO order |
| **closure** | A function that captures variables from its surrounding scope |
| **`init()`** | A special function that runs automatically after package-level variable initialisation, before `main()` |
| **exported** | An identifier starting with an uppercase letter; visible outside the package |
| **unexported** | An identifier starting with a lowercase letter; private to the package |
| **receiver** | The `(r *Reporter)` part of a method definition; like `this` or `self` in other languages |
| **nil** | The zero value for pointers, interfaces, slices, maps, channels, and functions; represents "no value" |
| **zero value** | The default value of a type when not explicitly set: `0` for ints, `""` for strings, `false` for bools, `nil` for pointers/interfaces |
| **`error`** | A built-in interface with one method: `Error() string`; `nil` means no error |
| **`fmt.Errorf`** | Creates a formatted error; `%w` wraps an existing error so it can be unwrapped later |
| **`io.Writer`** | Standard interface for write targets (`os.Stdout`, `strings.Builder`, `bytes.Buffer`, files…) |
| **`io.Reader`** | Standard interface for read sources (`os.Stdin`, `strings.NewReader`, files…) |
| **dependency injection** | Passing dependencies (like `io.Writer`) as parameters rather than hard-coding them; enables testability |
| **`:=`** | Short variable declaration — declares and assigns in one step; type is inferred |
| **`_`** | Blank identifier — discards a return value that must otherwise be used |
| **`go test -race`** | Runs tests with the race detector enabled; ~10x slower but catches concurrent bugs |
| **`go mod tidy`** | Removes unused deps and adds missing ones to `go.mod` and `go.sum` |
| **`go fmt`** | Auto-formats `.go` files to the canonical gofmt style |
| **`./...`** | A wildcard pattern meaning "this package and all sub-packages"; used with `go test`, `go build`, etc. |
| **octal literal** | `0o755` — integer written in base 8; used for Unix file permission bits |
| **`t.TempDir()`** | Creates a unique temp directory in tests that is cleaned up automatically after the test |
| **`t.Helper()`** | Marks a function as a test helper so error messages point to the calling test, not the helper |


---

## 15. dukh — The Background Workspace Monitor

### 15.1 What dukh Does

`dukh` is the second binary in the Grazhda ecosystem. While `zgard` performs workspace operations on demand (clone, pull, purge), `dukh` runs continuously in the background and answers one question: **is my workspace in the state I declared in config.yaml?**

Every few minutes dukh walks through every workspace, project, and repository in `config.yaml` and checks:

1. Does the repository directory exist on disk?
2. If it does, what git branch is currently checked out?
3. Does that branch match what `config.yaml` says?

This health snapshot is held in memory and served instantly over gRPC when `zgard ws status` asks for it.

### 15.2 Why a Separate Process?

`dukh` is a **long-running daemon** — a background process that never exits on its own. Daemons cannot be implemented as a simple command (like `zgard ws init`) because they need to keep running after the terminal is closed. Instead:

- `dukh start` self-daemonizes: it re-execs itself with `DUKH_DAEMON=1` in a detached child process
- The child runs independently, logs to a file, and exposes a gRPC port
- `dukh stop/scan` connect to that port to communicate with the running server
- `dukh status` reads the PID file and checks process liveness
- `zgard ws status` uses the gRPC Status RPC to report workspace health

---

## 16. gRPC and Protocol Buffers

### 16.1 What is gRPC?

**gRPC** (Google Remote Procedure Call) is a framework for making function calls across process boundaries — even across a network. Instead of a REST API where you send JSON over HTTP, gRPC:

- Defines services and message types in a `.proto` file (the *contract*)
- Generates Go code from that contract (`protoc` tool)
- Uses HTTP/2 and binary encoding — faster and more type-safe than JSON

For Grazhda, gRPC means `zgard` can call functions on the running `dukh` process as if they were local Go functions.

### 16.2 The proto File

`proto/dukh.proto` is the single source of truth for dukh's API:

```proto
service DukhService {
  rpc Stop(StopRequest)     returns (StopResponse);
  rpc Status(StatusRequest) returns (StatusResponse);
  rpc Scan(ScanRequest)     returns (ScanResponse);
}
```

- `service` defines the server and its callable methods
- `rpc` means "remote procedure call" — a method that can be called from another process
- `message` types define what data is sent and received (like Go structs)

**Key message types:**

```proto
message StatusResponse {
  repeated WorkspaceStatus workspaces = 1;  // a slice of WorkspaceStatus
  string server_version                = 2;  // e.g. "0.1.0"
  int64  uptime_seconds                = 3;  // seconds since dukh started
}
```

The numbers (`= 1`, `= 2`, …) are *field tags* — unique identifiers used in the binary encoding. They never change once defined, which allows adding new fields without breaking older clients.

### 16.3 Code Generation

`just generate` runs:

```bash
protoc \
  --go_out=dukh/proto --go_opt=paths=source_relative \
  --go-grpc_out=dukh/proto --go-grpc_opt=paths=source_relative \
  --proto_path=proto \
  proto/dukh.proto
```

This produces two files in `dukh/proto/`:

| File | Contents |
|---|---|
| `dukh.pb.go` | Go structs for every `message` type; marshal/unmarshal logic |
| `dukh_grpc.pb.go` | `DukhServiceServer` interface (implement in the server); `DukhServiceClient` (used by zgard) |

**Never edit these files by hand.** They are regenerated every time `just generate` runs.

### 16.4 The Client–Server Pattern

```
zgard ws status               dukh start
       |                           |
  dial("localhost:50501")     grpc.NewServer()
       |                           |
  client.Status(req)  ──RPC──>  server.Status(req)
       |                           |
  resp <──────────────────────  return resp
       |
  renderStatus(resp)
```

`zgard` is the **client** — it connects, makes one call, disconnects.
`dukh` is the **server** — it listens forever, handles calls, returns responses.

---

## 17. dukh/server — Server Packages

### 17.1 server.go — gRPC Lifecycle

`server.go` contains the `Server` struct:

```go
type Server struct {
    dukhpb.UnimplementedDukhServiceServer  // satisfies the interface
    grpcServer *grpc.Server
    monitor    *Monitor
    logger     *log.Logger
    startedAt  time.Time
    stopped    atomic.Bool
}
```

**`UnimplementedDukhServiceServer`** is an embedded struct generated by protoc. It implements every RPC method with a "not implemented" error. By embedding it, `Server` automatically satisfies the `DukhServiceServer` interface even if you only override some methods. This is a common Go pattern for forward-compatibility.

**`atomic.Bool`** is a boolean that is safe to read and write from multiple goroutines simultaneously without a mutex. The `sync/atomic` package provides a small set of lock-free operations for simple types.

The `ListenAndServe` method starts the gRPC server:

```go
func (s *Server) ListenAndServe(addr string) error {
    lis, err := net.Listen("tcp", addr)   // open a TCP port
    s.grpcServer = grpc.NewServer()
    dukhpb.RegisterDukhServiceServer(s.grpcServer, s)  // wire up our handler
    return s.grpcServer.Serve(lis)        // blocks until GracefulStop is called
}
```

`grpc.NewServer()` creates a gRPC server. `Serve()` blocks, accepting incoming connections and dispatching RPCs to the registered handler (`s`). When `GracefulStop()` is called (e.g. from the `Stop` RPC), `Serve()` returns.

**RPC handlers** are plain Go methods that satisfy the generated interface:

```go
func (s *Server) Stop(_ context.Context, _ *dukhpb.StopRequest) (*dukhpb.StopResponse, error) {
    go func() {
        time.Sleep(100 * time.Millisecond)
        s.GracefulStop()
    }()
    return &dukhpb.StopResponse{Message: "dukh shutting down"}, nil
}
```

The goroutine trick (`go func() { ... }()`) lets us return the response *before* the server shuts down. Without it, the client would never receive the response because the server would close the connection first.

```go
func (s *Server) Scan(_ context.Context, _ *dukhpb.ScanRequest) (*dukhpb.ScanResponse, error) {
    s.monitor.TriggerScan()
    return &dukhpb.ScanResponse{Message: "rescan initiated"}, nil
}
```

`Scan` just calls `TriggerScan()` on the monitor and returns immediately. The actual rescan happens in the monitor's goroutine — `Scan` does not wait for it to complete.

### 17.2 monitor.go — Configurable Background Polling

The `Monitor` struct manages the polling loop:

```go
type Monitor struct {
    mu          sync.RWMutex
    snapshot    []WorkspaceHealth
    configPath  string
    logger      *log.Logger
    stopCh      chan struct{}
    doneCh      chan struct{}
    triggerScan chan struct{}
}
```

**`sync.RWMutex`** is a reader/writer mutex. Multiple goroutines can read the snapshot simultaneously (shared read lock) but only one can write at a time (exclusive write lock). This is more efficient than a plain `sync.Mutex` when reads are frequent and writes are rare.

```go
// Reading (many readers can hold this simultaneously):
m.mu.RLock()
defer m.mu.RUnlock()
return m.snapshot

// Writing (exclusive — blocks all readers and other writers):
m.mu.Lock()
m.snapshot = result
m.mu.Unlock()
```

**Channels for lifecycle control:**

```go
stopCh      chan struct{}      // closed by Stop() to signal the loop to exit
doneCh      chan struct{}      // closed by the loop when it exits; Stop() waits on it
triggerScan chan chan struct{}  // capacity-1 buffer; nil = fire-and-forget, non-nil = wait for completion
```

`chan chan struct{}` is a channel that carries other channels as values. This lets callers optionally pass a "reply" channel so the loop can signal them when a scan finishes.

**The polling loop uses `time.Timer`, not `time.Ticker`:**

```go
func (m *Monitor) loop() {
    defer close(m.doneCh)
    m.scan()

    for {
        period := m.loadPeriod()         // re-read from config every cycle
        timer := time.NewTimer(period)

        select {
        case <-timer.C:                   // normal tick
            m.scan()
        case reply := <-m.triggerScan:    // manual scan requested
            timer.Stop()
            m.scan()
            if reply != nil {
                close(reply)              // unblock any TriggerScanAndWait caller
            }
        case <-m.stopCh:                  // shutdown requested
            timer.Stop()
            return
        }
    }
}
```

`time.Ticker` fires at a fixed interval set at creation. `time.Timer` fires once, after a duration set at creation. By creating a new `Timer` every iteration, the period is re-read from `config.yaml` on every cycle — so changing `dukh.monitoring.period_mins` takes effect at the next tick with no restart needed.

**`select`** waits on multiple channel operations simultaneously and executes whichever case is ready first. If multiple are ready, one is chosen at random. This is Go's core concurrency primitive for coordinating goroutines.

**`TriggerScan` uses a non-blocking send (fire-and-forget):**

```go
func (m *Monitor) TriggerScan() {
    select {
    case m.triggerScan <- nil:  // nil reply = fire-and-forget
    default:                    // channel full — scan already queued, skip
    }
}
```

The channel has capacity 1. If a scan signal is already queued (channel full), `default` runs instead of blocking. This ensures `TriggerScan` is always instant regardless of server load.

**`TriggerScanAndWait` blocks until the scan completes:**

```go
func (m *Monitor) TriggerScanAndWait(ctx context.Context) error {
    reply := make(chan struct{})
    select {
    case m.triggerScan <- reply:  // send a non-nil reply channel
    case <-ctx.Done():
        return ctx.Err()
    }
    select {
    case <-reply:   // loop closes this when the scan finishes
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

This is the **done-channel pattern** — a common Go idiom for synchronous callbacks over channels. The caller creates a fresh `chan struct{}`, sends it as the payload, then waits for it to be closed. The loop closes the channel after `scan()` returns, which unblocks the caller. Used by `zgard ws status --rescan` so the CLI waits for fresh data before printing.

**`loadPeriod` reads the configured polling interval:**

```go
func (m *Monitor) loadPeriod() time.Duration {
    cfg, err := config.Load(m.configPath)
    if err != nil || cfg.Dukh.Monitoring.PeriodMins <= 0 {
        return defaultPeriod  // 5 minutes
    }
    return time.Duration(cfg.Dukh.Monitoring.PeriodMins) * time.Minute
}
```

`time.Duration` is an `int64` counting nanoseconds. `time.Minute` is a constant equal to `60 * time.Second`. Multiplying an integer by `time.Minute` gives a `time.Duration`.

### 17.3 log.go — Structured Logging with Rotation

```go
func InitLogger(grazhdaDir string) (*log.Logger, func(), error) {
    rotator := &lumberjack.Logger{
        Filename:   filepath.Join(logDir, "dukh.log"),
        MaxSize:    5,     // MiB before rotation
        MaxBackups: 3,     // keep 3 old files
        Compress:   true,  // gzip old files
    }
    multi := io.MultiWriter(rotator, os.Stderr)
    logger := log.New(multi)
    logger.SetLevel(log.InfoLevel)
    return logger, func() { _ = rotator.Close() }, nil
}
```

**`lumberjack.Logger`** is a writer that automatically rotates log files. When `dukh.log` reaches 5 MiB it is renamed to `dukh-2026-04-06T09:00:00.log.gz` and a new `dukh.log` is started. Up to 3 old files are kept.

**`io.MultiWriter`** returns a writer that mirrors every write to multiple destinations simultaneously — both the rotating file and `os.Stderr`. This means log lines appear both in the log file and in the terminal if you run `dukh start` directly.

**`charmbracelet/log`** is a structured logger. "Structured" means log entries carry typed key-value pairs alongside the message:

```go
logger.Info("monitor: scan complete", "workspaces", 3)
// Output: INFO monitor: scan complete workspaces=3
```

The return value includes a `cleanup` function (the `func()` return type). The caller defers it:

```go
logger, cleanup, _ := server.InitLogger(grazhdaDir)
defer cleanup()
```

This is a common Go pattern for resource cleanup: return a function the caller can defer, so cleanup happens automatically when the function returns.

---

## 18. dukh CLI Commands

### 18.1 dukh start — Self-Daemonizing Background Process

`dukh start` uses the **self-re-exec pattern**: instead of requiring a separate launcher to spawn it, the `dukh` binary spawns itself in detached mode.

```go
func runStart(_ *cobra.Command, _ []string) error {
    if os.Getenv(daemonEnv) == "1" {
        return runServer()   // already the daemon — run the server
    }
    return launchDaemon()    // launcher mode — re-exec self and exit
}

func launchDaemon() error {
    exe, _ := os.Executable()
    cmd := exec.Command(exe, "start")
    cmd.Env = append(os.Environ(), "DUKH_DAEMON=1")
    cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
    setDetach(cmd)           // Setsid=true on Unix
    cmd.Start()
    fmt.Printf("✓ dukh started (pid %d)\n", cmd.Process.Pid)
    return nil
}
```

`os.Executable()` returns the absolute path to the currently running binary. This is more reliable than using `os.Args[0]` (which may be relative or a shell wrapper).

**`DUKH_DAEMON=1`** is the flag that distinguishes "launcher mode" from "daemon mode". When dukh detects this environment variable, it skips the re-exec and goes straight to running the server.

**`setDetach(cmd)`** is in `detach_unix.go`:

```go
//go:build !windows

func setDetach(cmd *exec.Cmd) {
    cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
```

`Setsid: true` calls the `setsid` Unix system call, which starts the process in a new session with no controlling terminal. This makes the child process fully independent: it survives when the launcher exits and is not killed by Ctrl-C in the terminal.

### 18.2 dukh scan — Triggering an Immediate Rescan

```go
func runScan(_ *cobra.Command, _ []string) error {
    conn, client, err := dial()
    defer conn.Close()
    resp, err := client.Scan(context.Background(), &dukhpb.ScanRequest{})
    fmt.Println("✓ " + resp.Message)
}
```

The scan itself happens asynchronously in dukh's monitor goroutine. `dukh scan` only guarantees the signal was delivered, not that the scan finished. If you need to wait for the scan to complete, use `zgard ws status --rescan` instead.

### 18.3 dukh status — Process Health

`dukh status` is about **process health** — not workspace health. It answers: is the daemon running?

```go
func runDukhStatus(_ *cobra.Command, _ []string) error {
    pid, err := readPIDFile(grazhdaDir)   // read $GRAZHDA_DIR/run/dukh.pid
    if !isProcessAlive(pid) {
        fmt.Println("○  dukh: not running")
        return nil
    }
    uptime := tryGetUptime()              // optional gRPC probe
    fmt.Printf("●  dukh: running  (pid %d, uptime: %s)\n", pid, uptime)
}
```

`isProcessAlive(pid)` on Unix uses `syscall.Kill(pid, 0)`. Signal 0 is never actually delivered — it only checks whether the process exists and is reachable. Returns `nil` error if alive.

This is distinct from `zgard ws status` — `dukh status` reports the daemon's process state.

### 18.4 zgard ws status --rescan — Synchronous Scan Before Status

`zgard ws status --rescan` combines a scan trigger and a status report into one atomic operation. The user sees fresh data instead of the last cached snapshot.

```bash
zgard ws status --rescan          # all workspaces
zgard ws status --rescan --name myws  # one workspace
```

**How it works end-to-end:**

1. zgard builds `StatusRequest{Rescan: true}` and sends it over gRPC with a 60-second context
2. dukh's `Status()` handler detects `req.Rescan == true`
3. It calls `monitor.TriggerScanAndWait(ctx)` — sends a reply channel, then blocks
4. The monitor loop picks up the trigger, calls `scan()`, then closes the reply channel
5. `TriggerScanAndWait` returns `nil`; the handler then reads the (now fresh) snapshot
6. dukh returns `StatusResponse` to zgard
7. zgard prints `⟳ rescanning workspaces…` then renders the health report

The gRPC client context timeout of 60 seconds propagates to `TriggerScanAndWait(ctx)`. If dukh takes longer than 60 seconds to finish scanning, the RPC is cancelled and zgard prints an error. This prevents the CLI from hanging indefinitely on large workspaces.

**Key difference vs `dukh scan`:**

| | `dukh scan` | `zgard ws status --rescan` |
|---|---|---|
| Waits for scan to finish | ✗ (fire-and-forget) | ✓ (blocks until done) |
| Returns health data | ✗ | ✓ |
| Useful when | you want to trigger a background refresh | you need fresh data right now |

### 18.5 Auto-Start — Transparent dukh Bootstrapping

`zgard ws status` transparently starts `dukh` when it is not already running. This eliminates the need for users to manually run `dukh start` before checking workspace health.

**How it works:**

1. zgard creates a lazy gRPC client via `grpc.NewClient` (no immediate connection).
2. The first `Status` RPC call fails because no server is listening.
3. zgard catches the error and prints `⟳ dukh is not running — starting…`.
4. It resolves the `dukh` binary — first checking `$GRAZHDA_DIR/bin/dukh`, then falling back to `$PATH`.
5. It runs `dukh start` as a subprocess, which handles its own daemonization (see §18.6).
6. zgard polls the gRPC endpoint at 500 ms intervals for up to 10 seconds.
7. Once the `Status` RPC succeeds, it prints `✓ dukh started` and renders the health report.

```
⟳ dukh is not running — starting…
✓ dukh started (pid 12345)
✓ dukh started

Dukh  running  •  uptime: 0s
...
```

**Why retry instead of one-shot?** The `dukh start` subprocess returns as soon as the daemon process is forked, but the gRPC listener may take a moment to bind. Polling with a short interval and reasonable timeout ensures zgard waits just long enough.

**Why resolve the binary?** In a standard Grazhda installation, `dukh` lives in `$GRAZHDA_DIR/bin/`. If `$GRAZHDA_DIR` is set, zgard checks there first. This ensures the auto-start uses the same version of dukh that was installed, even if another version exists on `$PATH`.

### 18.6 Configurable Monitoring Period

`config.yaml` controls how often dukh scans:

```yaml
dukh:
  host: localhost
  port: 50501
  monitoring:
    period_mins: 5
```

The Go struct that maps to this YAML is in `internal/config/config.go`:

```go
type DukhConfig struct {
    Host       string           `yaml:"host"`
    Port       int              `yaml:"port"`
    Monitoring MonitoringConfig `yaml:"monitoring"`
}

type MonitoringConfig struct {
    PeriodMins int `yaml:"period_mins"`
}
```

Each struct field has a **struct tag** — the backtick string after the type. `yaml:"period_mins"` tells the YAML parser to map the key `period_mins` in the YAML file to the `PeriodMins` field. Without the tag, the parser would look for a key named `PeriodMins` (case-insensitive match, but explicit tags are clearer).

The monitor reads this on every cycle so you can tune the interval without restarting dukh.

---

## 19. Updated Project Layout

```
grazhda/
├── go.work                 top-level workspace — links all three modules
├── Justfile                build recipes: generate, build-zgard, build-dukh, test, fmt, tidy
├── proto/
│   └── dukh.proto          source of truth for the gRPC contract (edit this, not the generated files)
├── config.template.yaml    template copied to $GRAZHDA_DIR/config.yaml on install
├── internal/               shared Go module (imported by both zgard and dukh)
│   ├── color/              Green/Red/Yellow/Blue helpers wrapping fatih/color
│   ├── config/             Load(), Validate(), DukhConfig, MonitoringConfig, Workspace hierarchy
│   ├── executor/           Executor interface + OsExecutor (captures stderr for good error messages)
│   ├── reporter/           Record(), Summary(), ExitCode() — progress output for ws commands
│   └── workspace/          Init, Purge, Pull domain logic + targeting resolver
├── zgard/                  zgard CLI module
│   ├── main.go             entry point: calls Execute()
│   ├── root.go             root Cobra command; wires ws subcommand group
│   └── ws/                 zgard ws command group
│       ├── ws.go           NewCmd() — creates the group
│       ├── init.go         zgard ws init
│       ├── pull.go         zgard ws pull
│       ├── purge.go        zgard ws purge
│       ├── status.go       zgard ws status — gRPC Status RPC; coloured health report
│       ├── config.go       shared config loading for ws commands
│       └── confirm.go      interactive confirmation prompt
└── dukh/                   dukh gRPC server module
    ├── cmd/
    │   ├── main.go         entry point; Cobra root with start/stop/status/scan commands
    │   ├── start.go        dukh start — self-re-exec daemonization; runServer()
    │   ├── stop.go         dukh stop — gRPC Stop RPC
    │   ├── status.go       dukh status — PID file + process liveness check
    │   ├── scan.go         dukh scan — gRPC Scan RPC (immediate rescan)
    │   ├── dial.go         shared gRPC dial helper
    │   ├── detach_unix.go  setDetach() — Setsid=true; only compiled on non-Windows
    │   ├── detach_windows.go setDetach() no-op for Windows
    │   ├── pid_unix.go     isProcessAlive() via syscall.Kill(pid, 0)
    │   └── pid_windows.go  isProcessAlive() via os.FindProcess
    ├── proto/              generated protobuf Go code — DO NOT EDIT
    │   ├── dukh.pb.go      message structs (StopRequest, StatusResponse, RepoStatus, …)
    │   └── dukh_grpc.pb.go DukhServiceServer interface + DukhServiceClient
    └── server/
        ├── server.go       Server struct; ListenAndServe; Stop/Status/Scan handlers; PID file
        ├── monitor.go      Monitor struct; configurable polling loop; TriggerScan; health snapshot
        └── log.go          InitLogger: charmbracelet/log + lumberjack 5 MiB rotation
```

---

## 20. New Go Concepts Introduced in Phase 2

| Concept | Explanation |
|---|---|
| **gRPC** | Remote procedure call framework; methods defined in `.proto` files, called across processes |
| **protobuf** | Protocol Buffers — binary serialisation format used by gRPC; more compact than JSON |
| **`protoc`** | The Protocol Buffer compiler; generates Go code from `.proto` files |
| **`//go:build`** | Build constraint; controls which files are compiled based on OS, architecture, or tags |
| **`time.Timer`** | Fires once after a duration; create a new one each loop to support variable intervals |
| **`time.Ticker`** | Fires repeatedly at a fixed interval; less flexible than Timer for variable periods |
| **`sync.RWMutex`** | Reader/writer mutex; many readers OR one writer, not both simultaneously |
| **`io.MultiWriter`** | Fan-out writer; copies every write to multiple destinations at once |
| **`exec.Command`** | Prepares an OS command; `Start()` launches it without waiting; `Run()` waits |
| **`syscall.SysProcAttr`** | Low-level process attributes; `Setsid: true` starts a new session (detaches from terminal) |
| **`atomic.Bool`** | A boolean safe for concurrent read/write without a mutex |
| **buffered channel** | `make(chan T, n)` — holds up to n items without blocking the sender |
| **non-blocking send** | `select { case ch <- v: default: }` — sends only if channel has space, otherwise skips |
| **struct tag** | Backtick annotation on a field: `` `yaml:"field_name"` `` — controls marshalling behaviour |
| **`context.Background()`** | Root context with no deadline; used for operations that should not time out |
| **`context.WithTimeout`** | Creates a context that is cancelled after a duration; good for network calls |
| **lumberjack** | Third-party library for log rotation by file size |
| **`charmbracelet/log`** | Structured logger with key-value pairs; outputs to any `io.Writer` |
| **daemon / background process** | A long-running process with no controlling terminal; started with `Setsid` |
| **`grpc.NewServer()`** | Creates a gRPC server; `Serve(lis)` blocks accepting connections |
| **`UnimplementedDukhServiceServer`** | Generated embedded struct; provides default "not implemented" for all RPCs |

---

## 21. grazhda Management Script

The `grazhda` Bash script (installed to `$GRAZHDA_DIR/bin/grazhda`) provides self-management for your Grazhda installation.

### File Origin

```
project root: grazhda          ← source file
  ↓ just copy-scripts
bin/grazhda                    ← compiled into bin/ during build
  ↓ install_from_sources
$GRAZHDA_DIR/bin/grazhda       ← active script on user's PATH
```

### Command Router Pattern

The script uses a Bash `case` statement to dispatch subcommands — the same pattern used by tools like `git`, `docker`, and `kubectl`:

```bash
main() {
    case "$1" in
        upgrade) cmd_upgrade ;;
        config)  cmd_config  ;;
        *)       usage; exit 1 ;;
    esac
}
main "$@"
```

`"$@"` forwards all positional parameters; quoting prevents word-splitting if a parameter contains spaces.

### `grazhda upgrade` — How It Works

```bash
grazhda upgrade
```

1. Verifies `git`, `go`, `just`, `protoc` are all in `$PATH`
2. `cd $GRAZHDA_DIR/sources` → `git pull`
3. `export PATH="$(go env GOPATH)/bin:$PATH"` → ensures `protoc-gen-go` is reachable
4. `just build` → runs generate + builds zgard + builds dukh + copies scripts
5. `cp bin/* $GRAZHDA_DIR/bin/` → installs updated binaries

**Self-update safety:** When the script replaces itself (`cp bin/grazhda $GRAZHDA_DIR/bin/grazhda`), the already-running Bash process is unaffected — it has the old file loaded. The new version is active on the next invocation.

### `grazhda config --edit` — How It Works

```bash
grazhda config --edit
```

1. Checks `$GRAZHDA_DIR/config.yaml` exists
2. Reads `editor:` from `config.yaml` using `grep` + `sed` (no YAML parser needed)
3. Falls back through `$VISUAL` → `$EDITOR` → `vi`
4. `exec "$editor" "$CONFIG_FILE"` — replaces the shell process with the editor

**Why `exec`?** Using `exec` instead of just running the editor command means the shell process is replaced by the editor. Signals (Ctrl+C, window resize) go directly to the editor, not to a wrapper shell. The shell's return code is the editor's return code.

### YAML Parsing in Bash

A minimal helper extracts flat top-level scalar keys from YAML:

```bash
yaml_get() {
    local key="$1" file="$2"
    grep -E "^${key}:[[:space:]]*" "$file" \
        | sed "s/^${key}:[[:space:]]*//" \
        | tr -d '"'"'" \
        | xargs
}
```

- `grep -E "^key:\s*"` — finds the line starting with `key:`
- `sed` strips the key prefix
- `tr -d '"'"'"` — removes any surrounding quotes (single or double)
- `xargs` trims leading/trailing whitespace

This is intentionally limited. It only works for flat scalar keys. Nested YAML (like `dukh.host`) must be parsed in Go.

### The `editor:` Config Field

`config.yaml` gains a top-level `editor:` field:

```yaml
# Editor used by `grazhda config --edit`.
# Resolved in order: this field → $VISUAL → $EDITOR → vi
editor: vim
```

**Resolution chain** (first non-empty wins):

| Priority | Source |
|---|---|
| 1 | `editor:` in `$GRAZHDA_DIR/config.yaml` |
| 2 | `$VISUAL` environment variable |
| 3 | `$EDITOR` environment variable |
| 4 | `vi` (hardcoded fallback) |

### New Shell Concepts in Phase 3

| Concept | Explanation |
|---|---|
| **`exec cmd`** | Replaces the current shell process with `cmd`; no return |
| **`command -v name`** | Checks if `name` is a reachable command; returns 0 if found, 1 if not |
| **`xargs`** | Reads stdin and trims leading/trailing whitespace when used with no arguments |
| **`tr -d chars`** | Deletes every character in `chars` from stdin |
| **`grep -E`** | Extended regex grep; no need to escape `+`, `|`, `()`, `{}`  |
| **`${var:-default}`** | Uses `default` if `var` is unset or empty |
| **`set -e`** | Exit immediately if any command returns non-zero |
| **`git pull` exit codes** | 0 = success (with or without changes); non-zero = network/merge error |
