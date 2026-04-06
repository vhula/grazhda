---
phase: 2
component: dukh
status: planned
---

# Epics & Stories — dukh gRPC Monitor (Phase 2)

This document defines the implementation epics and stories for the `dukh` background monitor server and its integration into `zgard`. Stories are ordered for sequential delivery; each story is independently deployable within its epic.

---

## Epic D1: Proto & Build Foundation

**Goal:** Establish the protobuf contract and code-generation pipeline so that all subsequent epics have a stable, generated API to build against.

---

### Story D1.1: Create `proto/dukh.proto`

**Goal:** Define the canonical gRPC service contract as a protobuf source file.

**Acceptance Criteria:**
- `proto/dukh.proto` exists at repo root with `syntax = "proto3"`.
- Defines `DukhService` with `Stop` and `Status` RPCs.
- Defines all messages: `StopRequest`, `StopResponse`, `StatusRequest`, `StatusResponse`, `WorkspaceStatus`, `ProjectStatus`, `RepoStatus` — with field names and numbers matching the architecture spec.
- `option go_package` is set to `"github.com/vhula/grazhda/dukh/proto"`.
- File is syntactically valid (`protoc --version` exits 0 against the file).

**Implementation Notes:**
- Keep the proto file minimal and precise — no extra fields, no comments beyond field descriptions.
- Field numbers 1–N are assigned in order; once assigned, never change them.

---

### Story D1.2: Add `just generate` recipe

**Goal:** Automate protobuf Go code generation via the project's `Justfile`.

**Acceptance Criteria:**
- `Justfile` has a `generate` recipe that invokes `protoc` with `--go_out` and `--go-grpc_out` targeting `dukh/proto/`.
- `just generate` succeeds and produces `dukh/proto/dukh.pb.go` and `dukh/proto/dukh_grpc.pb.go`.
- Generated files have the correct `package proto` declaration.
- `dukh/proto/` is listed in `.gitignore` with a comment explaining it is generated.

**Implementation Notes:**
- Use `--go_opt=paths=source_relative` and `--go-grpc_opt=paths=source_relative`.
- Requires `protoc-gen-go` and `protoc-gen-go-grpc` in `$PATH`; `grazhda-init.sh` should document this prerequisite.

---

### Story D1.3: Update `grazhda-init.sh` to verify protoc and run `just generate`

**Goal:** Ensure the developer environment check covers the protoc toolchain.

**Acceptance Criteria:**
- `grazhda-init.sh` checks for `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` in `$PATH`; prints a clear error and exits if any are missing.
- Script runs `just generate` after dependency checks pass.
- Existing `grazhda-init.sh` checks (Go version, `$GRAZHDA_DIR`, etc.) are unmodified.

**Implementation Notes:**
- Use the same `check_tool` pattern already in the script (if present) or add a simple `command -v <tool> || { echo "missing: <tool>"; exit 1; }` guard.

---

## Epic D2: dukh gRPC Server

**Goal:** Implement the full `dukh` background server: CLI entry point, gRPC lifecycle, workspace monitor, and structured logging.

---

### Story D2.1: Implement `dukh start` command with Cobra CLI

**Goal:** Create the `dukh` binary entry point that wires the Cobra CLI to the server.

**Acceptance Criteria:**
- `dukh/cmd/main.go` defines a root Cobra command with a `start` subcommand.
- `dukh start` calls `server.Start(cfg)` and blocks until shutdown.
- `dukh --help` prints usage with a description of the start command.
- `dukh/go.mod` declares module `github.com/vhula/grazhda/dukh` and lists all required dependencies.
- `go.work` includes `dukh/` as a workspace member.
- `just build-dukh` (new Justfile recipe) produces a `dukh` binary.

**Implementation Notes:**
- Keep `main.go` thin: parse flags, load config via `internal/config.Load()`, pass to `server.Start()`.
- Config loading resolves `$GRAZHDA_DIR` in `main.go` — not inside `internal/`.

---

### Story D2.2: Implement gRPC server lifecycle (start, graceful stop)

**Goal:** Implement the gRPC listener, DukhService server struct, and graceful shutdown.

**Acceptance Criteria:**
- `server/server.go` implements `DukhService` (generated interface).
- Server listens on the address from config (`dukh.host:dukh.port`), defaulting to `localhost:50501`.
- `Stop` RPC calls the internal shutdown function: `GracefulStop()`, remove PID file, flush log.
- `SIGTERM` and `SIGINT` trigger the same shutdown path via `signal.NotifyContext`.
- PID is written to `$GRAZHDA_DIR/run/dukh.pid` on start; removed on stop.
- If a PID file exists and the process is alive, startup aborts with `"dukh already running (pid N)"`.
- Integration smoke test: `dukh start &` then `zgard dukh stop` exits 0 and PID file is gone.

**Implementation Notes:**
- Use `google.golang.org/grpc.NewServer()` with no interceptors in Phase 2.
- Shutdown function must be idempotent (safe to call twice).

---

### Story D2.3: Implement workspace monitor (polling, branch detection)

**Goal:** Implement the background goroutine that polls repo health every 30 seconds.

**Acceptance Criteria:**
- `server/monitor.go` exposes a `Monitor` struct with `Start(ctx context.Context)` and `Snapshot() []WorkspaceStatus` methods.
- On each poll cycle, all workspaces/projects/repos from config are evaluated.
- For each repo: `exists` is determined via `os.Stat`; `actual_branch` via `git -C <path> symbolic-ref --short HEAD`.
- `branch_aligned = exists && actual_branch == configuredBranch`.
- Snapshot is stored under `sync.RWMutex`; `Status` RPC reads under read lock.
- First poll fires immediately on start (not after 30-second delay).
- Poll interval is a constant `30 * time.Second` (not configurable in Phase 2).
- `StatusRequest.workspace_name` filters the snapshot; empty string returns all workspaces.
- Unit test: mock executor returns a known branch; verify snapshot reflects it correctly.

**Implementation Notes:**
- Use `os/exec.CommandContext` for the git invocation so polling respects context cancellation.
- `actual_branch` output is `strings.TrimSpace`d before comparison.
- Repo path: `workspace.Path + "/" + project.Name + "/" + repo.LocalDirName()` — use `filepath.Join`.

---

### Story D2.4: Implement logging with charmbracelet/log + lumberjack rotation

**Goal:** Wire structured logging to a rotating file in `$GRAZHDA_DIR/logs/dukh.log`.

**Acceptance Criteria:**
- `server/log.go` creates a `charmbracelet/log.Logger` writing to a `lumberjack.Logger`.
- Lumberjack config: `MaxSize: 5` (MiB), `MaxBackups: 3`, `Compress: true`.
- Log file path: `$GRAZHDA_DIR/logs/dukh.log`; directory created if absent.
- Logger is passed into `Server` struct (not a global).
- Structured log entries written for: startup (with version, listen address), each poll cycle (duration, repo count), Stop RPC received, any polling error (with repo path and error).
- Log format is `charmbracelet/log` default (timestamp + level + key=value pairs).

**Implementation Notes:**
- `charmbracelet/log.SetOutput(lj)` or construct via `log.New(lj)` — confirm API from the library's README.
- No log entries are written to stdout/stderr during normal operation; only errors that prevent startup go to stderr.

---

## Epic D3: zgard dukh CLI

**Goal:** Add `zgard dukh stop` and `zgard dukh status` commands that connect to the running `dukh` server and render results.

---

### Story D3.1: Implement `zgard dukh stop` command

**Goal:** Let users stop `dukh` from the `zgard` CLI.

**Acceptance Criteria:**
- `zgard dukh stop` creates a gRPC client, calls `Stop`, prints the response message, and exits 0.
- If connection fails (refused / timeout), prints `✗ dukh not reachable — connection refused (<addr>)` to stderr and exits 1.
- gRPC dial uses a 3-second connection timeout.
- Target address resolved from config (`dukh.host:dukh.port`), defaulting to `localhost:50501`.

**Implementation Notes:**
- Use `google.golang.org/grpc.NewClient` (or `Dial` with `WithBlock` + timeout context).
- `WithInsecure` / `grpc.WithTransportCredentials(insecure.NewCredentials())` — no TLS in Phase 2.
- Colour output via `internal/color` package (same as `zgard ws` commands).

---

### Story D3.2: Implement `zgard dukh status` command with colored output

**Goal:** Render the full workspace health report from dukh's Status RPC.

**Acceptance Criteria:**
- `zgard dukh status` calls `Status(StatusRequest{})` and renders output matching the UX spec exactly.
- Header line: `Dukh  running  •  uptime: <formatted>` — `running` in green bold.
- Workspace headers in blue bold; project headers in default colour with 2-space indent.
- Aligned repos: `✓` in green, branch arrow in green; mismatched repos: `✗` in red, actual branch in red with `(branch mismatch)` annotation.
- Missing repos: `✗` in red with `(missing)` annotation; no branch arrows.
- Summary line: `✓ N aligned  ✗ N drifted  ✗ N missing`.
- Repo name column is padded per-project block (not globally).
- Uptime formatted as per UX spec (`Xs` / `Xm Ys` / `Xh Ym`).
- If dukh is not reachable: `✗ dukh not running — start with: dukh start` to stderr, exit 1.
- Colour suppressed when stdout is not a TTY.
- Unit test: given a `StatusResponse` fixture, verify rendered output string matches expected.

**Implementation Notes:**
- Build the renderer as a pure function `RenderStatus(w io.Writer, resp *proto.StatusResponse, colorEnabled bool)` so it is testable without a real gRPC connection.
- Reuse `internal/color` for consistent colour handling.

---

### Story D3.3: Wire dukh commands into zgard root

**Goal:** Register the `dukh` command group in the `zgard` Cobra command tree.

**Acceptance Criteria:**
- `zgard dukh --help` lists `stop` and `status` subcommands.
- `zgard --help` lists `dukh` as a top-level command alongside `ws`.
- All existing `zgard ws` commands continue to pass their tests unmodified.
- `zgard dukh` without a subcommand prints help and exits 0.

---

### Story D3.4: Add `--rescan` flag to `zgard dukh status`

**Goal:** Allow users to request a synchronous workspace rescan before the health report is rendered.

**Acceptance Criteria:**
- `zgard dukh status --rescan` prints `⟳ rescanning workspaces…` in blue before calling the RPC.
- The RPC is sent with `StatusRequest{Rescan: true}` and a 60-second context timeout.
- dukh's `Status` handler calls `monitor.TriggerScanAndWait(ctx)` and waits for the scan to complete.
- The health report that follows reflects the current state of the filesystem (not a cached snapshot).
- If the scan exceeds 60 seconds the RPC returns a deadline error; zgard prints `✗ dukh status failed: <err>` to stderr and exits 1.
- When `--rescan` is not set, behaviour is identical to pre-flag: a plain `StatusRequest{Rescan: false}` is sent with a default context.
- `zgard dukh status --help` documents the `--rescan` flag.

**Implementation Notes:**
- `triggerScan` channel type upgrades from `chan struct{}` to `chan chan struct{}` to carry an optional reply channel.
- `TriggerScan()` sends `nil` (fire-and-forget); `TriggerScanAndWait(ctx)` sends a `make(chan struct{})` and waits for it to be closed.
- Loop must check: `if reply != nil { close(reply) }` after each scan.

