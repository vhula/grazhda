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
21. [grazhda Management Script](#21-grazhda-management-script)
22. [internal/format — Duration Formatting](#22-internalformat--duration-formatting)
23. [internal/grpcdial — gRPC Connection Helper](#23-internalgrpcdial--grpc-connection-helper)
24. [internal/ui — Terminal Rendering](#24-internalui--terminal-rendering)
25. [internal/pkgman — Package Management Domain](#25-internalpkgman--package-management-domain)
    - [registry.go — Data Model and Persistence](#251-registrygo--data-model-and-persistence)
    - [Dual Registry Architecture](#252-dual-registry-architecture)
    - [resolver.go — Dependency Resolution (Kahn's Algorithm)](#253-resolvergo--dependency-resolution-kahns-algorithm)
    - [env.go — Marker-Based Environment Blocks](#254-envgo--marker-based-environment-blocks)
    - [installer.go — Package Installation Lifecycle](#255-installergo--package-installation-lifecycle)
    - [purger.go — Package Removal Lifecycle](#256-purgergo--package-removal-lifecycle)
    - [runner.go — Script Execution Engine](#257-runnergo--script-execution-engine)
    - [spinner.go — Progress Indicator](#258-spinnergo--progress-indicator)
26. [zgard/cfg — Configuration Commands](#26-zgardcfg--configuration-commands)
27. [zgard/pkg — Package Management CLI](#27-zgardpkg--package-management-cli)
    - [install.go — Installing Packages](#271-installgo--installing-packages)
    - [purge.go — Removing Packages](#272-purgego--removing-packages)
    - [register.go — Interactive Package Registration](#273-registergo--interactive-package-registration)
    - [unregister.go — Removing from the Local Registry](#274-unregistergo--removing-from-the-local-registry)
    - [registry_load.go — Shared Registry Loader](#275-registry_loadgo--shared-registry-loader)
28. [Shell Scripts Deep Dive](#28-shell-scripts-deep-dive)
    - [grazhda-init.sh — Shell Startup Initialization](#281-grazhda-initsh--shell-startup-initialization)
    - [grazhda-install.sh — First-Run Installer](#282-grazhda-installsh--first-run-installer)
29. [Testing Patterns and Strategies](#29-testing-patterns-and-strategies)
    - [Go Testing Patterns](#291-go-testing-patterns)
    - [Bash Script Testing](#292-bash-script-testing)
    - [Testing Cobra CLI Commands](#293-testing-cobra-cli-commands)
30. [Complete Project Layout (Phase 4)](#30-complete-project-layout-phase-4)
31. [New Go Concepts Introduced in Phases 3–4](#31-new-go-concepts-introduced-in-phases-34)

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

Grazhda has three active modules:

| Directory | Module path (`go.mod` module line) |
| :--- | :--- |
| `internal/` | `github.com/vhula/grazhda/internal` |
| `zgard/` | `github.com/vhula/grazhda/zgard` |
| `dukh/` | `github.com/vhula/grazhda/dukh` |

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

### `grazhda uninstall` — How It Works

```bash
grazhda uninstall
```

1. Prints a warning banner and prompts `Are you sure? (y/n)`; exits 0 on non-`y`
2. Stops `dukh` gracefully via `stop_dukh_if_running` (same helper as `upgrade`)
3. Detects `~/.bashrc.user` or `~/.bashrc` (same logic as the installer)
4. Removes lines matching `grazhda-init\.sh` and `export GRAZHDA_DIR=` using `grep -v` into a temp file, then copies back
5. Runs `find $GRAZHDA_DIR -mindepth 1 -maxdepth 1 ! -name 'config.yaml' -exec rm -rf {} +` — deletes all top-level entries inside `$GRAZHDA_DIR` except `config.yaml`

**Why keep `config.yaml`?** The user's workspace layout is non-trivial to recreate. Preserving it means `grazhda-install.sh` can be re-run and immediately pick up the existing configuration without any manual editing.

### `grazhda purge` — How It Works

```bash
grazhda purge
```

Same flow as `uninstall`, but the final step is `rm -rf "${GRAZHDA_DIR:?}"` — the entire directory is removed, including `config.yaml`. Intended for a clean slate when Grazhda is no longer needed.

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

---

## 22. internal/format — Duration Formatting

This tiny utility package wraps `time.Duration` into human-friendly strings:

```go
package format

func Uptime(d time.Duration) string
```

Usage:

```go
format.Uptime(2*time.Hour + 15*time.Minute)   // "2h 15m"
format.Uptime(3*time.Minute + 42*time.Second)  // "3m 42s"
format.Uptime(17 * time.Second)                // "17s"
```

**Design:** The function cascades through three bands (hours, minutes, seconds) and always shows at most two units. This keeps status output concise — "2h 15m" is more readable than "2h15m42s".

### Why a dedicated package?

Formatting a duration as "2h 15m" is needed in multiple places (dukh status output, reporter timing, etc.). Extracting it into `internal/format` means a single implementation serves every caller.

### Go concept: `time.Duration`

`time.Duration` is an `int64` alias measuring nanoseconds. The standard library provides constants like `time.Second`, `time.Minute`, and `time.Hour` for arithmetic. Methods like `d.Hours()`, `d.Minutes()`, and `d.Seconds()` return `float64` values, so the function casts them to `int` to drop decimal fractions.

---

## 23. internal/grpcdial — gRPC Connection Helper

This package centralises how zgard and dukh CLI commands connect to the dukh gRPC server:

```go
package grpcdial

func Addr(cfg *config.Config) string         // address from config or defaults
func DefaultAddr() string                    // "localhost:50501"
func Dial(addr string) (*grpc.ClientConn, error) // open a lazy connection
```

### Fallback chain

```
config.yaml → dukh.host + dukh.port
    ↓ (if not set)
defaults: "localhost" + 50501
```

`Addr()` inspects `cfg.Dukh.Host` and `cfg.Dukh.Port` and falls back to the hardcoded defaults. This means the dukh server works out of the box with no configuration — you only need to set `dukh.host` / `dukh.port` if you want a non-standard address.

### Lazy connections

```go
func Dial(addr string) (*grpc.ClientConn, error) {
    return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
```

`grpc.NewClient()` creates a connection that doesn't actually dial the server until the first RPC call. This is called "lazy connection establishment". The caller must eventually call `conn.Close()` (usually via `defer conn.Close()`).

### Go concept: error wrapping with `%w`

```go
return fmt.Errorf("dial %q: %w", addr, err)
```

The `%w` verb in `fmt.Errorf` wraps the original error. Callers can then use `errors.Is(err, someErr)` or `errors.As(err, &target)` to inspect the cause chain. This is Go's standard pattern for enriching errors with context while preserving the original.

---

## 24. internal/ui — Terminal Rendering

This package renders markdown help text for the terminal using the [glamour](https://github.com/charmbracelet/glamour) library:

```go
package ui

func Render(md string) string
```

### How it works

1. If the input is empty, return `""`
2. Check `color.IsDisabled()` — if colour is off, use the "notty" style (plain text)
3. Otherwise, auto-detect terminal background (dark/light) and pick the matching glamour style
4. Render the markdown; word-wrap at 100 characters
5. If anything fails, **return the raw markdown**

This last point is the key design decision: **fail open**. If the terminal doesn't support colours, or glamour encounters an edge-case markdown construct, the user still sees the help text — just without styling.

### Why render help as markdown?

Cobra's built-in help formatter produces plain text with basic indentation. By writing `rootLong` as Markdown and rendering it with glamour, the help output gets headings, bullet lists, code blocks, and colour — all from a single source string. The same markdown also reads well on GitHub if you paste it into a README.

### Go concept: conditional imports

The `ui` package imports both `glamour` and the internal `color` package. This is not "conditional" in the C `#ifdef` sense — Go always compiles all imports. Instead, the package uses runtime branching:

```go
if color.IsDisabled() {
    // plain style
} else {
    // coloured style
}
```

Go encourages keeping all code in the binary and branching at runtime rather than using build tags for optional features.

---

## 25. internal/pkgman — Package Management Domain

This is the largest package in the project and contains the complete domain logic for declarative package management. It handles:

- **Registry** — loading, saving, and merging YAML package definitions
- **Resolver** — topological dependency sorting with cycle detection
- **Env** — marker-based shell environment block management
- **Installer / Purger** — orchestrating multi-step package lifecycles
- **Runner** — executing shell scripts with environment overlays
- **Spinner** — progress feedback for non-verbose mode

### 25.1. registry.go — Data Model and Persistence

The core data types:

```go
type Registry struct {
    Packages []Package `yaml:"registry"`
}

type Package struct {
    Name           string   `yaml:"name"`
    Version        string   `yaml:"version,omitempty"`
    PreCreateDir   bool     `yaml:"pre_create_dir,omitempty"`
    DependsOn      []string `yaml:"depends_on,omitempty"`
    PreInstallEnv  string   `yaml:"pre_install_env,omitempty"`
    Install        string   `yaml:"install,omitempty"`
    PostInstallEnv string   `yaml:"post_install_env,omitempty"`
    Purge          string   `yaml:"purge,omitempty"`
}
```

Each field maps directly to a YAML key. The `omitempty` tag tells the YAML marshaller to skip empty fields when writing — so a package with no version produces `name: sdkman` rather than `name: sdkman\nversion: ""`.

**Package identity** is the tuple `(Name, Version)`. Two packages with the same name but different versions are distinct entries. The helper function `samePkgIdentity(a, b)` checks:

```go
func samePkgIdentity(a, b Package) bool {
    return a.Name == b.Name && a.Version == b.Version
}
```

**Path helpers** construct canonical file locations:

| Function | Returns |
|---|---|
| `RegistryPath(dir)` | `dir + "/.grazhda.pkgs.yaml"` |
| `LocalRegistryPath(dir)` | `dir + "/registry.pkgs.local.yaml"` |
| `EnvPath(dir)` | `dir + "/.grazhda.env"` |
| `PkgDir(dir, name)` | `dir + "/pkgs/" + name` |

**Loading and saving** use `gopkg.in/yaml.v3`:

```go
func LoadRegistry(path string) (*Registry, error) {
    data, err := os.ReadFile(path)     // read the whole file
    var reg Registry
    err = yaml.Unmarshal(data, &reg)   // parse YAML into the struct
    return &reg, nil
}
```

`LoadLocalRegistry(path)` differs in one crucial way: if the file does not exist, it returns an empty registry instead of an error. This is because the local registry is optional — a user may never create one.

### 25.2. Dual Registry Architecture

Grazhda uses two package registries:

```
$GRAZHDA_DIR/
├── .grazhda.pkgs.yaml          ← GLOBAL (shipped with Grazhda, replaced on upgrade)
└── registry.pkgs.local.yaml    ← LOCAL  (user-managed, never touched by upgrade)
```

**Global registry:** Defines packages that ship with the project. It is overwritten during `grazhda upgrade` so the maintainer can add, remove, or update package definitions. Users should not edit this file.

**Local registry:** Users can register their own packages or override global ones. It uses the exact same YAML schema but lives in a separate file that `grazhda upgrade` never replaces.

**Merge semantics** (`MergeRegistries()`):

1. Start with all global packages
2. For each local package, compare `(Name, Version)` against global entries
3. If an exact match exists, the local entry **replaces** the global one
4. If no match, the local entry is **appended**
5. The merged result is used for install/purge/resolve operations

```go
func MergeRegistries(global, local *Registry) *Registry {
    merged := &Registry{Packages: append([]Package{}, global.Packages...)}
    for _, lp := range local.Packages {
        found := false
        for i, gp := range merged.Packages {
            if samePkgIdentity(gp, lp) {
                merged.Packages[i] = lp
                found = true
                break
            }
        }
        if !found {
            merged.Packages = append(merged.Packages, lp)
        }
    }
    return merged
}
```

**Important:** local packages can depend on global packages. The `depends_on` field references packages by name (or `name@version`), and the resolver looks up dependencies in the merged registry. This means a local package like `my-tool` can declare `depends_on: [sdkman]` where `sdkman` is defined in the global registry.

### 25.3. resolver.go — Dependency Resolution (Kahn's Algorithm)

The resolver turns an unordered list of packages into a safe installation order where every package's dependencies are installed before the package itself.

**Algorithm overview (Kahn's topological sort):**

```
1. Build an in-degree map:  for each package, count how many dependencies it has
2. Seed queue: packages with in-degree 0 (no dependencies) go into a queue
3. Loop: dequeue a package, add it to the output, decrement in-degree of
   packages that depend on it. If any reach 0, enqueue them.
4. Detect cycles: if the output is smaller than the selected set, the
   remaining nodes form a cycle.
```

Here is the core loop:

```go
for len(queue) > 0 {
    cur := queue[0]
    queue = queue[1:]
    ordered = append(ordered, selected[cur])

    for _, dependent := range deps[cur] {
        inDegree[dependent]--
        if inDegree[dependent] == 0 {
            queue = append(queue, dependent)
            sortStrings(queue)
        }
    }
}
```

**Why sort the queue?** Kahn's algorithm is non-deterministic — when multiple packages have zero in-degree, the algorithm can pick any of them. By sorting the queue after each insertion, Grazhda always produces the same output for the same input. This is important for reproducible builds and predictable test assertions.

**Version-aware references** — the `depends_on` field supports two formats:

| depends_on entry | Meaning |
|---|---|
| `"sdkman"` | Refers to the only package named `sdkman` (any version, or unversioned) |
| `"sdkman@1.2.3"` | Refers specifically to the package with `name: sdkman` and `version: 1.2.3` |

The `parseDep()` function splits these:

```go
func parseDep(s string) (name, version string) {
    if idx := strings.Index(s, "@"); idx >= 0 {
        return s[:idx], s[idx+1:]
    }
    return s, ""
}
```

**Multi-version ambiguity** — if the registry has `jdk@17.0.8-tem` and `jdk@21.0.1-tem`, and a depends_on entry says just `"jdk"`, the resolver returns an error:

```
package "jdk" has multiple versions; use "jdk@<version>" in depends_on
```

The exception: if one of the candidates has an empty version field, it is treated as the "default" and selected automatically.

**Transitive closure** — the `closure()` function walks the dependency graph depth-first using a stack. Starting from the seed packages, it follows every `depends_on` edge and collects all reachable packages. This ensures that installing `maven` also pulls in `jdk` and `sdkman` even though you only asked for `maven`.

**Reverse resolution** for purging:

```go
func ResolveReverse(reg *Registry, names []string) ([]Package, error) {
    ordered, err := Resolve(reg, names)
    // ... reverse ordered in place ...
    return ordered, nil
}
```

Reversing the topological order means dependents come first. When purging `maven → jdk → sdkman`, the purger removes maven before jdk before sdkman, so no package is removed while something still depends on it.

### Go concept: Kahn's algorithm

Kahn's algorithm (1962) is one of the two classical approaches to topological sort (the other being DFS-based). It has the advantage of naturally detecting cycles: after processing, any node with a non-zero in-degree is part of a cycle. This makes cycle reporting trivial.

### Go concept: `strings.Cut`

```go
k, _, _ := strings.Cut(kv, "=")
```

`strings.Cut` splits a string on the first occurrence of a separator and returns `(before, after, found)`. It was added in Go 1.18 as a cleaner alternative to `strings.SplitN(s, "=", 2)`. When you don't need the boolean, `_, _, _` discards it cleanly.

### 25.4. env.go — Marker-Based Environment Blocks

The `.grazhda.env` file stores shell environment variable exports for installed packages. Instead of appending lines that are impossible to remove later, pkgman uses **named blocks** with sentinel markers:

```bash
# === BEGIN GRAZHDA: sdkman:pre ===
export SDKMAN_DIR="$GRAZHDA_DIR/pkgs/sdkman"
# === END GRAZHDA: sdkman:pre ===

# === BEGIN GRAZHDA: sdkman:post ===
[[ -s "$SDKMAN_DIR/bin/sdkman-init.sh" ]] && source "$SDKMAN_DIR/bin/sdkman-init.sh"
# === END GRAZHDA: sdkman:post ===
```

Each package can have two blocks:
- `<name>:pre` — written **before** installation (so the install script can use the exported variables)
- `<name>:post` — written **after** installation (so the user's shell picks up the new tool)

**Three operations:**

| Function | Behaviour |
|---|---|
| `UpsertBlock(path, name, content)` | Write block; replace if already present, append if not |
| `RemoveBlock(path, name)` | Delete the named block; no-op if missing |
| `HasBlock(path, name)` | Check existence without modifying the file |

**Idempotency:** calling `UpsertBlock` twice with identical arguments produces the same file. This is critical because `zgard pkg install --all` might be re-run after a partial failure — previously written blocks must not be duplicated.

**How `UpsertBlock` works:**

```
1. Read the file (or start with empty if ENOENT)
2. findBlock() scans line-by-line for the begin/end markers
3. If found: splice out the old lines, insert the new block
4. If not found: append a blank line separator + the new block
5. Write back to the file
```

The `findBlock()` helper returns `(startIdx, endIdx)` line indices, or `(-1, -1)` if the block is not present:

```go
func findBlock(lines []string, beginMarker, endMarker string) (startIdx, endIdx int) {
    startIdx = -1
    for i, l := range lines {
        trimmed := strings.TrimSpace(l)
        if startIdx < 0 && trimmed == beginMarker {
            startIdx = i
        } else if startIdx >= 0 && trimmed == endMarker {
            return startIdx, i
        }
    }
    return -1, -1
}
```

**Blank line management** — when removing a block, excess blank lines before and after the deleted region are trimmed, and a single blank separator is kept between remaining blocks. This prevents the file from accumulating whitespace after repeated install/purge cycles.

### Go concept: slice surgery

```go
lines = append(rawLines[:startIdx:startIdx], newBlock...)
lines = append(lines, rawLines[endIdx+1:]...)
```

The three-index slice `rawLines[:startIdx:startIdx]` sets both length and capacity to `startIdx`. This prevents `append` from overwriting elements beyond the cut point. Without the capacity limit, Go would reuse the underlying array and corrupt the tail.

### 25.5. installer.go — Package Installation Lifecycle

`Installer` orchestrates the full installation of one or more packages:

```go
type Installer struct {
    grazhdaDir string
    reg        *Registry
    out        io.Writer
    errOut     io.Writer
    verbose    bool
}
```

**Per-package lifecycle** (`installOne`):

```
1. Print "▶ Installing sdkman" (blue)
2. If pre_create_dir: mkdir $GRAZHDA_DIR/pkgs/sdkman
3. If pre_install_env:
   a. UpsertBlock(.grazhda.env, "sdkman:pre", content)
   b. source .grazhda.env         ← makes exports visible to the install script
4. Run install script:
   a. Prepend `source .grazhda.env` to the script
   b. In verbose mode: stream stdout/stderr line by line
   c. In quiet mode: show a braille spinner, print "✓ done" / "✗ failed"
5. If post_install_env:
   a. UpsertBlock(.grazhda.env, "sdkman:post", content)
   b. source .grazhda.env         ← makes new exports visible to subsequent packages
6. Print "✓ sdkman installed" (green)
```

**Why source .grazhda.env before the install script?** Consider installing `jdk` which depends on `sdkman`. The sdkman installation writes `SDKMAN_DIR` into its `:pre` block and sdkman-init into its `:post` block. When the installer reaches `jdk`, it sources `.grazhda.env` before running `sdk install java ...`. This makes `sdk` (provided by sdkman-init) available to the jdk install script, even though sdkman was installed moments ago in the same session.

**Verbose vs quiet mode:**

- **Verbose** (`--verbose`): script output lines are streamed to the terminal prefixed with `│ ` for visual grouping
- **Quiet** (default): output goes to `io.Discard`, and a braille `Spinner` animates on stderr to show progress

### 25.6. purger.go — Package Removal Lifecycle

`Purger` is the mirror image of `Installer`:

```go
type Purger struct {
    grazhdaDir string
    reg        *Registry
    out        io.Writer
    errOut     io.Writer
    verbose    bool
}
```

**Per-package lifecycle** (`purgeOne`):

```
1. Print "▶ Purging sdkman" (yellow)
2. Run optional purge script (e.g., `sdk uninstall java ...`)
3. If pre_create_dir: rm -rf $GRAZHDA_DIR/pkgs/sdkman
4. Remove env blocks:
   a. RemoveBlock(.grazhda.env, "sdkman:pre")
   b. RemoveBlock(.grazhda.env, "sdkman:post")
5. Print "✓ sdkman purged" (green)
```

**Colour convention:** Install uses **blue** (▶), Purge uses **yellow** (▶). Both use green (✓) for success and red (✗) for failure. This gives the user an instant visual cue about which operation is running.

**Idempotent removal:** `removeBlockIfPresent()` checks `HasBlock()` before calling `RemoveBlock()`. Missing directories/blocks don't cause errors. This means you can safely run `zgard pkg purge --all` twice.

### 25.7. runner.go — Script Execution Engine

`Runner` executes a single shell script phase (install, purge, source-env) for a specific package:

```go
type Runner struct {
    grazhdaDir string
    pkg        Package
    out        io.Writer
    errOut     io.Writer
}
```

**Key method:**

```go
func (r *Runner) RunPhase(ctx context.Context, phase, script string) error
```

**How it works:**

1. Skip empty scripts (no-op)
2. Launch `bash -c <script>` with `exec.CommandContext`
3. Set environment: inherit `os.Environ()` + overlay `GRAZHDA_DIR`, `PKG_DIR`, `PKG_NAME`, `VERSION`
4. Pipe stdout and stderr through `streamLines()` which prefixes every line with `│ `
5. Wait for the process to exit; non-zero exit code becomes a Go error

**Environment overlay pattern:**

```go
func (r *Runner) buildEnv() []string {
    overlay := []string{
        "GRAZHDA_DIR=" + r.grazhdaDir,
        "PKG_DIR=" + pkgDir,
        "PKG_NAME=" + r.pkg.Name,
        "VERSION=" + r.pkg.Version,
    }
    // For each os.Environ() entry, replace it with the overlay if the key matches.
    // This ensures overlays win without duplicating keys.
}
```

The overlay merge is important because Go's `exec.Cmd.Env` replaces the entire environment. If you just set `Env = overlay`, the process loses `PATH`, `HOME`, etc. By starting from `os.Environ()` and only replacing matching keys, the child process inherits the full environment plus the grazhda-specific overrides.

### Go concept: `exec.CommandContext`

```go
cmd := exec.CommandContext(ctx, "bash", "-c", script)
```

`exec.CommandContext` ties the child process to a `context.Context`. When the context is cancelled (e.g., Ctrl+C), Go sends SIGKILL to the child process. This ensures that long-running install scripts are interrupted when the user cancels.

### Go concept: goroutines for pipe draining

```go
done := make(chan struct{})
go func() {
    defer close(done)
    streamLines(stdoutPipe, "    │ ", r.out)
}()
streamLines(stderrPipe, "    │ ", r.errOut)
<-done
```

Stdout and stderr must be drained concurrently. If the process writes more than the OS pipe buffer to stdout while nobody reads it, the process blocks. By reading stdout in a goroutine and stderr in the main goroutine, both pipes are drained simultaneously. The `<-done` channel receive ensures the goroutine finishes before `cmd.Wait()` is called.

### 25.8. spinner.go — Progress Indicator

```go
type Spinner struct {
    mu      sync.Mutex
    msg     string
    out     io.Writer
    stop    chan struct{}
    stopped bool
}
```

The spinner renders braille animation frames (`⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏`) at 100 ms intervals, giving a smooth rotating effect:

```
  ⠹ [install] running…
```

**Thread safety:** The spinner runs in its own goroutine and can be stopped from any goroutine. A `sync.Mutex` protects the message field, and a `chan struct{}` signals the goroutine to exit. The `stopped` flag prevents double-close panics.

**Line clearing:** `Stop()` first writes `\r%-80s\r` to overwrite the spinner line with spaces, then prints the final status. This avoids leftover spinner characters in the terminal.

### Go concept: `sync.Mutex`

A `sync.Mutex` is the simplest synchronisation primitive. `mu.Lock()` acquires exclusive access; `mu.Unlock()` releases it. Here it protects the `msg` and `stopped` fields which are read by the ticker goroutine and written by the caller's goroutine.

### Go concept: channel close as broadcast

`close(s.stop)` wakes up every goroutine blocked in `<-s.stop` or `select { case <-s.stop: }`. This is the canonical Go pattern for one-shot "shutdown" signals — closing a channel is effectively a broadcast to all listeners.

---

## 26. zgard/cfg — Configuration Commands

The `cfg` package (originally named `cfgcmd`, then renamed) provides the `zgard config` subcommand tree:

```
zgard config
├── path       Print the resolved config file path
├── validate   Load and validate config, report errors
├── list       Display workspace/project hierarchy with colours
└── get <key>  Look up a value by dotted YAML path
```

### Command wiring

```go
func NewCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "config", Short: "Manage configuration"}
    cmd.AddCommand(newPathCmd(), newValidateCmd(), newListCmd(), newGetCmd())
    return cmd
}
```

The parent `config` command has no `Run` function — running `zgard config` alone prints help. Each subcommand is created by an unexported factory (e.g., `newPathCmd()`) that returns a `*cobra.Command`.

### resolveConfigPath()

```go
func resolveConfigPath() string
```

1. Check `$GRAZHDA_DIR` — if set, return `$GRAZHDA_DIR/config.yaml`
2. Otherwise, return `$HOME/.grazhda/config.yaml`

This fallback chain mirrors the convention used across all Grazhda scripts and commands.

### Dotted-path access (the `get` subcommand)

`zgard config get dukh.port` uses `config.GetByPath()` to traverse the YAML tree using dot-separated keys. Array indices are supported as numeric segments, e.g., `workspaces.0.name` retrieves the name of the first workspace.

### Go concept: `reporter.ExitError`

```go
return reporter.ExitError{Code: 1}
```

Cobra normally calls `os.Exit(1)` on error, but Grazhda's root command intercepts the error from `cmd.Execute()` and checks if it's an `ExitError`. If so, it uses the error's `Code` field as the process exit code. This keeps the exit code logic out of individual commands and centralises it in the root.

---

## 27. zgard/pkg — Package Management CLI

The `pkg` package provides the `zgard pkg` subcommand tree:

```
zgard pkg
├── install      Install one or all packages
├── purge        Remove one or all packages
├── register     Interactively add a package to the local registry
└── unregister   Remove a package from the local registry
```

### `grazhdaDir()` — shared helper

Every subcommand needs the installation directory. This helper reads `$GRAZHDA_DIR` or falls back to `$HOME/.grazhda`:

```go
func grazhdaDir() string {
    if d := os.Getenv("GRAZHDA_DIR"); d != "" {
        return d
    }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".grazhda")
}
```

### 27.1. install.go — Installing Packages

**Flags:**

| Flag | Short | Type | Description |
|---|---|---|---|
| `--name` | `-n` | string | Package ref (`<name>` or `<name>@<version>`) |
| `--all` | | bool | Install all packages |
| `--verbose` | `-v` | bool | Stream script output to the terminal |

**Validation:** exactly one of `--name` or `--all` must be set. Both or neither is an error.

**Flow:**

```go
func runInstall(cmd *cobra.Command, args []string) error {
    reg, err := loadMergedRegistry(grazhdaDir())  // global + local merged
    inst := pkgman.NewInstaller(dir, reg, os.Stdout, os.Stderr, verbose)
    return inst.Install(cmd.Context(), names)
}
```

### 27.2. purge.go — Removing Packages

Identical flag structure to install. Uses `pkgman.NewPurger()` with `ResolveReverse()` internally for safe removal order.

### 27.3. register.go — Interactive Package Registration

This command has no flags — it walks the user through an interactive dialog:

```
$ zgard pkg register
Name (required): my-tool
Version (optional): 1.0.0
Pre-create directory? (y/N): y
Dependencies (choose from existing packages):
  1. sdkman
  2. jdk@17.0.8-tem
  3. maven@3.9.4
Select (space-separated numbers, or empty): 1 2
Pre-install env (multi-line, empty line to finish):
  export MY_TOOL_HOME="$PKG_DIR"

Install script (multi-line, empty line to finish):
  curl -L https://... | tar xz -C "$PKG_DIR"

Post-install env (multi-line, empty line to finish):
  export PATH="$MY_TOOL_HOME/bin:$PATH"

Purge script (multi-line, empty line to finish):
  rm -rf "$PKG_DIR"

✓ Registered my-tool@1.0.0 in local registry
```

**Prompt helpers:**

```go
func promptRequired(in *bufio.Reader, out io.Writer, label string) (string, error)
func promptOptional(in *bufio.Reader, out io.Writer, label string) (string, error)
func promptBool(in *bufio.Reader, out io.Writer, label string) (bool, error)
func promptMultiline(in *bufio.Reader, out io.Writer, label string) (string, error)
func promptDependsOn(in *bufio.Reader, out io.Writer, reg *pkgman.Registry) ([]string, error)
```

`promptRequired` keeps asking until a non-empty value is given. `promptMultiline` collects lines until the user enters a blank line. `promptDependsOn` displays a numbered list of all packages in the merged registry (not just the local one), so users can depend on global packages.

### Go concept: `bufio.Reader` for interactive input

```go
reader := bufio.NewReader(os.Stdin)
line, _ := reader.ReadString('\n')
```

`bufio.NewReader` wraps `os.Stdin` with a buffer. `ReadString('\n')` reads until a newline, including the newline itself. You must `strings.TrimSpace()` the result to remove it. This pattern avoids the more complex `fmt.Scanln` which chokes on spaces in input.

### 27.4. unregister.go — Removing from the Local Registry

**Flags:**

| Flag | Type | Description |
|---|---|---|
| `--name` | string | Package name to remove |
| `--version` | string | Optional: remove only this version |
| `--all` | bool | Clear the entire local registry |

**Three modes:**

```
zgard pkg unregister --name my-tool           ← remove all versions
zgard pkg unregister --name my-tool --version 1.0.0  ← exact match only
zgard pkg unregister --all                    ← empty the local registry
```

The `--all` flag is mutually exclusive with `--name` and `--version`. The command validates this and returns a clear error message.

### 27.5. registry_load.go — Shared Registry Loader

A small helper used by `install`, `purge`, `register`, and `unregister`:

```go
func loadMergedRegistry(grazhdaDir string) (*pkgman.Registry, error) {
    global, err := pkgman.LoadRegistry(pkgman.RegistryPath(grazhdaDir))
    local, err  := pkgman.LoadLocalRegistry(pkgman.LocalRegistryPath(grazhdaDir))
    return pkgman.MergeRegistries(global, local), nil
}
```

This three-line function exists as a separate file because:

1. It avoids duplicating the load-and-merge logic in four commands
2. It provides a single place to change if the merge strategy ever changes
3. It's independently testable

---

## 28. Shell Scripts Deep Dive

### 28.1. grazhda-init.sh — Shell Startup Initialization

This 14-line script is sourced every time you open a terminal (via `.bashrc`):

```bash
#!/bin/bash
GRAZHDA_DIR="${GRAZHDA_DIR:-$HOME/.grazhda}"

# Idempotent PATH prepend
case ":$PATH:" in
    *":$GRAZHDA_DIR/bin:"*) ;;
    *) export PATH="$GRAZHDA_DIR/bin:$PATH" ;;
esac

mkdir -p "$GRAZHDA_DIR/pkgs"

# Source env (with fallback for legacy filename)
[ -f "$GRAZHDA_DIR/.grazhda.env" ] && source "$GRAZHDA_DIR/.grazhda.env"
source "$GRAZHDA_DIR/grazhda-env.sh" 2>/dev/null || true
```

**Idempotent PATH prepend:**

The `case ":$PATH:" in *":$GRAZHDA_DIR/bin:"*)` pattern checks if the directory is already in PATH. The colons ensure exact matching: without them, `/usr/local/bin` would falsely match a search for `/usr/local/bi`. By wrapping PATH in colons, every entry is bordered by colons, making the match precise.

**Why `2>/dev/null || true`?** The legacy `grazhda-env.sh` may not exist. Redirecting stderr to `/dev/null` suppresses the "file not found" message, and `|| true` prevents `set -e` (if active) from aborting the shell.

**Why no `set -e`?** This script is `source`d into the user's shell. If `set -e` were active and any command failed, it would kill the user's shell session. Defensive `|| true` guards are used instead.

### 28.2. grazhda-install.sh — First-Run Installer

This script handles the one-time Grazhda installation. Key functions:

**`verify_requirements()`** — checks that `git`, `go`, `just`, and `protoc` are on `$PATH` using `command -v`:

```bash
for cmd in git go just protoc; do
    if ! command -v "$cmd" &>/dev/null; then
        die "Missing required tool: $cmd"
    fi
done
```

**`run_logged()`** — captures command output to a log file while showing a one-line status:

```bash
run_logged() {
    local label="$1"; shift
    echo "  → $label"
    if "$@" >> "$LOG_FILE" 2>&1; then
        echo "    ✓ done"
    else
        echo "    ✗ failed (last 20 lines from log):"
        tail -20 "$LOG_FILE"
        exit 1
    fi
}
```

This pattern keeps the installation output clean while still capturing full logs for debugging.

**`create_config()`** — creates `config.yaml` from the template **only if it doesn't already exist**. This is important for reinstalls: the user's workspace configuration survives.

**`copy_pkgs_registry()`** — copies `.grazhda.pkgs.yaml` from the source tree to `$GRAZHDA_DIR/`. Unlike `create_config`, this **always overwrites** the file because the global registry is maintained by the project and must stay in sync.

**Source-testable guard:**

```bash
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
```

When the script is executed directly (`./grazhda-install.sh`), `BASH_SOURCE[0]` equals `$0`, so `main` runs. When the script is sourced (`source grazhda-install.sh`), they differ, so `main` is skipped. This allows test scripts to source the file and call individual functions in isolation.

---

## 29. Testing Patterns and Strategies

### 29.1. Go Testing Patterns

The project uses the standard `testing` package exclusively — no external test frameworks like testify or gomock.

**Table-driven tests** — the preferred pattern for functions with many input/output scenarios:

```go
func TestUptime(t *testing.T) {
    tests := []struct {
        name string
        in   time.Duration
        want string
    }{
        {"seconds", 17 * time.Second, "17s"},
        {"minutes", 3*time.Minute + 42*time.Second, "3m 42s"},
        {"hours", 2*time.Hour + 15*time.Minute, "2h 15m"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := format.Uptime(tt.in); got != tt.want {
                t.Errorf("Uptime(%v) = %q, want %q", tt.in, got, tt.want)
            }
        })
    }
}
```

**`t.TempDir()`** — Go 1.15+ provides a built-in temp directory that is automatically cleaned up when the test finishes:

```go
func TestLoadLocalRegistry_MissingFileReturnsEmpty(t *testing.T) {
    dir := t.TempDir()
    reg, err := pkgman.LoadLocalRegistry(filepath.Join(dir, "does-not-exist.yaml"))
    // ... assert empty registry, no error
}
```

**`t.Setenv()`** — Go 1.17+ provides a helper that sets an environment variable and restores the original value when the test finishes:

```go
func TestGrazhdaDir_FromEnv(t *testing.T) {
    t.Setenv("GRAZHDA_DIR", "/custom/path")
    if got := grazhdaDir(); got != "/custom/path" {
        t.Errorf("got %q, want /custom/path", got)
    }
}
```

**Mock I/O for interactive commands** — `bufio.Reader` wrapping `strings.NewReader` simulates user input:

```go
func TestPromptDependsOn_SelectsItems(t *testing.T) {
    input := strings.NewReader("1 3\n")
    reader := bufio.NewReader(input)
    var out bytes.Buffer
    deps, _ := promptDependsOn(reader, &out, &pkgman.Registry{...})
    // ... assert deps contains items 1 and 3
}
```

**Cobra command testing** — commands are tested by constructing the command, setting args, and capturing output:

```go
func TestPathCommand_PrintsResolvedPath(t *testing.T) {
    cmd := newPathCmd()
    var out bytes.Buffer
    cmd.SetOut(&out)
    cmd.SetErr(&bytes.Buffer{})
    cmd.Execute()
    // ... assert out.String() contains the expected path
}
```

### 29.2. Bash Script Testing

The project includes a custom test harness in `tests/bash/test_scripts.sh`. It uses no external framework — just Bash functions:

**Assertion helpers:**

```bash
assert_eq() {
    local label="$1" expected="$2" actual="$3"
    if [[ "$expected" != "$actual" ]]; then
        echo "FAIL: $label: expected '$expected', got '$actual'"
        exit 1
    fi
}

assert_file_contains() {
    local label="$1" file="$2" pattern="$3"
    if ! grep -q "$pattern" "$file"; then
        echo "FAIL: $label: '$file' does not contain '$pattern'"
        exit 1
    fi
}
```

**Isolated test environments:** Each test function creates a fresh temporary directory with `mktemp -d`:

```bash
test_installer_copy_pkgs_registry_overwrites() {
    local tmp; tmp=$(mktemp -d)
    TMP_DIRS+=("$tmp")
    export GRAZHDA_DIR="$tmp/grazhda"
    mkdir -p "$GRAZHDA_DIR"
    # ... run function under test, assert results
}
```

**Mock binaries for upgrade tests:** The upgrade test creates fake `git`, `go`, `just`, and `protoc` executables in a temporary bin directory:

```bash
mkdir -p "$tmp/fakebin"
for tool in git go just protoc; do
    printf '#!/bin/bash\ntrue\n' > "$tmp/fakebin/$tool"
    chmod +x "$tmp/fakebin/$tool"
done
export PATH="$tmp/fakebin:$PATH"
```

**Global cleanup:** Instead of per-function cleanup (which can fire at the wrong time when scripts `source` other scripts), the harness uses a global array and a single `trap`:

```bash
TMP_DIRS=()
cleanup() { for d in "${TMP_DIRS[@]}"; do rm -rf "$d"; done; }
trap cleanup EXIT
```

### 29.3. Testing Cobra CLI Commands

A common pattern for testing Cobra commands in Grazhda:

1. **Test command hierarchy** — verify that subcommands exist on the parent:

```go
func TestRootCmd_HasMainSubcommands(t *testing.T) {
    for _, name := range []string{"ws", "config", "pkg"} {
        _, _, err := rootCmd.Find([]string{name})
        if err != nil {
            t.Errorf("expected subcommand %q", name)
        }
    }
}
```

2. **Test flag definitions** — verify that persistent and local flags exist:

```go
func TestRootCmd_HasGlobalFlags(t *testing.T) {
    for _, name := range []string{"no-color", "json", "quiet"} {
        if rootCmd.PersistentFlags().Lookup(name) == nil {
            t.Errorf("missing flag --%s", name)
        }
    }
}
```

3. **Test command execution with args** — set args and assert output:

```go
cmd := newUnregisterCmd()
cmd.SetArgs([]string{"--all"})
cmd.SetIn(nil)
cmd.SetOut(&bytes.Buffer{})
cmd.SetErr(&bytes.Buffer{})
err := cmd.Execute()
```

---

## 30. Complete Project Layout (Phase 4)

```
grazhda/
│
├── go.work                          ← ties three local modules together
├── go.work.sum                      ← checksums for workspace modules
├── Justfile                         ← build, test, fmt, tidy, generate, man
├── .goreleaser.yml                  ← goreleaser cross-platform release config
├── .gitignore
├── LICENSE
│
├── grazhda                          ← management script: upgrade, uninstall, purge, config
├── grazhda-install.sh               ← first-run installer
├── grazhda-init.sh                  ← shell profile init (sourced on every terminal open)
├── config.template.yaml             ← default config template
├── .grazhda.pkgs.yaml               ← global package registry (shipped with project)
├── .grazhda.env                     ← generated env file (env blocks written by pkgman)
│
├── README.md                        ← project landing page
├── QUICK-START.md                   ← 5-minute setup guide
├── STUDY.md                         ← this file
├── feature-ideas.md
├── improvements.md
│
├── bin/                             ← build output (zgard, dukh, scripts)
│
├── docs/
│   ├── CLI.md                       ← complete CLI reference
│   ├── CONFIG.md                    ← configuration file reference
│   ├── DEVELOPMENT.md               ← contributor setup guide
│   ├── prd.md                       ← product requirements document
│   ├── architecture.md              ← system design and decisions
│   ├── ux-design-specification.md   ← output format and symbols
│   ├── epics.md                     ← feature epics (zgard/workspace)
│   └── epics-dukh.md                ← feature epics (dukh)
│
├── proto/
│   └── dukh.proto                   ← gRPC service definition
│
├── internal/                        ← module: github.com/vhula/grazhda/internal
│   ├── go.mod
│   ├── go.sum
│   ├── config/
│   │   ├── config.go                ← Load, Validate, DefaultWorkspace, RenderCloneCmd
│   │   └── config_test.go
│   ├── executor/
│   │   ├── executor.go              ← Executor interface + OsExecutor
│   │   ├── executor_test.go
│   │   └── mock.go                  ← MockExecutor for tests
│   ├── color/
│   │   ├── color.go                 ← Green/Red/Yellow/Blue helpers
│   │   └── color_test.go
│   ├── format/
│   │   ├── uptime.go                ← Uptime(duration) string formatter
│   │   └── uptime_test.go
│   ├── grpcdial/
│   │   ├── dial.go                  ← Addr/DefaultAddr/Dial helpers
│   │   └── dial_test.go
│   ├── ui/
│   │   ├── render.go                ← Render(markdown) for terminal output
│   │   └── render_test.go
│   ├── reporter/
│   │   ├── reporter.go              ← Reporter: ✓/⏭/✗ output + summary
│   │   └── reporter_test.go
│   ├── workspace/
│   │   ├── options.go               ← RunOptions struct
│   │   ├── targeting.go             ← Resolve: picks workspaces from flags
│   │   ├── workspace.go             ← Init, Purge, Pull
│   │   ├── workspace_test.go
│   │   └── targeting_test.go
│   ├── pkgman/
│   │   ├── registry.go              ← Package model, load/save/merge registries
│   │   ├── registry_test.go
│   │   ├── resolver.go              ← Kahn's topological sort + version-aware refs
│   │   ├── resolver_test.go
│   │   ├── env.go                   ← UpsertBlock/RemoveBlock/HasBlock
│   │   ├── env_test.go
│   │   ├── installer.go             ← Installer orchestration
│   │   ├── purger.go                ← Purger orchestration
│   │   ├── runner.go                ← Script execution with env overlays
│   │   ├── runtime_test.go          ← Tests for runner, spinner internals
│   │   └── spinner.go               ← Braille progress indicator
│   └── testdata/
│       ├── valid_single_workspace.yaml
│       ├── valid_multi_workspace.yaml
│       ├── duplicate_workspace_names.yaml
│       ├── missing_required_fields.yaml
│       ├── missing_branch.yaml
│       └── invalid_template.yaml
│
├── zgard/                           ← module: github.com/vhula/grazhda/zgard
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                      ← func main() — program entry point
│   ├── root.go                      ← root Cobra command + Execute()
│   ├── root_test.go
│   ├── cfg/
│   │   ├── cfgcmd.go                ← zgard config {path,validate,list,get}
│   │   └── cfgcmd_test.go
│   ├── ws/
│   │   ├── ws.go                    ← zgard ws parent command
│   │   ├── init.go                  ← zgard ws init
│   │   ├── pull.go                  ← zgard ws pull
│   │   ├── purge.go                 ← zgard ws purge
│   │   └── status.go                ← zgard ws status
│   └── pkg/
│       ├── pkg.go                   ← zgard pkg parent command
│       ├── pkg_test.go
│       ├── install.go               ← zgard pkg install
│       ├── purge.go                 ← zgard pkg purge
│       ├── register.go              ← zgard pkg register (interactive)
│       ├── unregister.go            ← zgard pkg unregister
│       └── registry_load.go         ← loadMergedRegistry() helper
│
├── dukh/                            ← module: github.com/vhula/grazhda/dukh
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── start.go                 ← dukh start (self-daemonizing)
│   │   ├── stop.go                  ← dukh stop
│   │   ├── status.go                ← dukh status
│   │   ├── scan.go                  ← dukh scan
│   │   └── cmd_test.go
│   ├── server/
│   │   ├── server.go                ← gRPC server lifecycle
│   │   ├── monitor.go               ← workspace polling loop
│   │   ├── log.go                   ← structured logging with rotation
│   │   └── server_test.go
│   └── proto/
│       └── dukh.pb.go               ← generated gRPC stubs
│
├── tools/
│   └── gen-manpages/
│       └── main.go                  ← man page generator using Cobra doc
│
├── tests/
│   └── bash/
│       └── test_scripts.sh          ← bash script test harness
│
└── .github/
    └── workflows/
        ├── just.yml                 ← CI: build + test on push/PR
        └── release.yml              ← release: goreleaser on tag push
```

---

## 31. New Go Concepts Introduced in Phases 3–4

| Concept | Where used | Explanation |
|---|---|---|
| **`sync.Mutex`** | `spinner.go` | Mutual exclusion lock; protects shared state between goroutines |
| **Channel close as broadcast** | `spinner.go` | `close(ch)` wakes all listeners of `<-ch`; used for shutdown signals |
| **`exec.CommandContext`** | `runner.go` | Ties child process to a context; cancels the process when context expires |
| **`strings.Cut`** | `runner.go` | Splits string on first separator; cleaner than `SplitN` (Go 1.18+) |
| **Three-index slice** | `env.go` | `s[:n:n]` sets both length and capacity; prevents `append` corruption |
| **`os.IsNotExist`** | `env.go`, `registry.go` | Checks if an error is "file not found"; used for graceful missing-file handling |
| **`bufio.NewReader`** | `register.go` | Buffered reader wrapping stdin; enables `ReadString('\n')` for line input |
| **`io.Discard`** | `installer.go`, `purger.go` | An `io.Writer` that discards all written bytes; used to suppress verbose output |
| **`time.NewTicker`** | `spinner.go` | Creates a repeating timer; delivers ticks on a channel at fixed intervals |
| **`os.Environ` + overlay** | `runner.go` | Inherits full environment then selectively overrides keys; avoids losing PATH etc. |
| **`yaml.Unmarshal` / `yaml.Marshal`** | `registry.go` | Deserialises/serialises Go structs from/to YAML using struct field tags |
| **`filepath.Join`** | throughout | OS-safe path concatenation (uses `/` on Unix, `\` on Windows) |
| **Insertion sort** | `resolver.go` | Simple O(n²) sort for small slices; avoids importing `sort` for deterministic output |

## 32. Effective Go Refactoring — DRY & Helpers

This section documents the structural refactors performed to eliminate code duplication and align with [Effective Go](https://go.dev/doc/effective_go) standards.

### 32.1. `zgard/ws/run.go` — Workspace Operation Runner

**Problem:** Five workspace commands (init, pull, stash, checkout, exec) each repeated ~40 lines of identical boilerplate: load config → resolve workspaces → warn on implicit target → create executor + reporter → iterate → summarise.

**Solution:** A single `runWorkspaceOp()` helper encapsulates the entire ceremony. Each command provides a `wsOpFunc` callback with its domain logic:

```go
type wsOpFunc func(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts workspace.RunOptions) error
```

Commands with extra arguments (checkout's branch, exec's command string) capture them in a closure:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    return runWorkspaceOp(cmd, opts, "checked out", "would check out",
        func(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts workspace.RunOptions) error {
            return workspace.Checkout(ws, args[0], exec, rep, opts)
        })
}
```

### Go concept: function types as callbacks

Go lets you define named function types (`type wsOpFunc func(...)`) and pass closures that match the signature. This is the idiomatic Go alternative to template-method inheritance — compose behaviour through function values rather than class hierarchies.

### 32.2. `internal/workspace/options.go` — `ctxOr()`

**Problem:** Both `RunOptions` and `InspectOptions` had identical `ctx()` methods: "return Context if non-nil, else `context.Background()`".

**Solution:** A standalone `ctxOr(ctx context.Context) context.Context` function replaces both methods. Both option types call `ctxOr(o.Context)`.

### Go concept: standalone functions vs methods

When a method doesn't use any fields of the receiver beyond the argument it's called with, extract it as a standalone function. This improves reusability and makes the intent clearer.

### 32.3. `zgard/pkg/flags.go` — `validateNameOrAll()`

**Problem:** `install.go` and `purge.go` both contained the same "exactly one of `--name` or `--all`" validation.

**Solution:** Extracted to a shared unexported function that both commands call at the top of their `RunE`.

### 32.4. `dukh/cmd/errs.go` — `printErr()`

**Problem:** `stop.go` and `scan.go` repeated `fmt.Fprintln(os.Stderr, icolor.Red("✗ "+msg))`.

**Solution:** A shared `printErr(msg)` helper in the same package.

### 32.5. Test Coverage Improvements

Phase 2 added 1,034 lines of tests across 13 files, covering all three modules:

| Package | Before | After |
|---|---|---|
| `internal/config` | 42.7% | 81.6% |
| `internal/pkgman` | 78.6% | 91.0% |
| `internal/workspace` | 73.2% | 87.0% |
| `dukh/server` | 24.8% | 59.2% |
| `internal/color` | 100% | 100% |
| `internal/format` | 100% | 100% |

Tests follow idiomatic Go patterns: table-driven tests, `t.TempDir()` for filesystem tests, `t.Setenv()` for environment overrides, and the standard `testing` package (no external frameworks).
