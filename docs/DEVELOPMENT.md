# Development Guide

> [← Back to README](../README.md) · [CLI Reference](CLI.md)

## Prerequisites

- **Go 1.24+** — [go.dev/dl](https://go.dev/dl/)
- **Just** — task runner ([github.com/casey/just](https://github.com/casey/just))
- **protoc + protoc-gen-go-grpc** — only if modifying `proto/dukh.proto`

### Installing Just

```bash
# Using Cargo (if you have Rust installed)
cargo install just

# Or download a pre-built binary from:
# https://github.com/casey/just/releases
```

---

## Quick Start

```bash
just build    # build all binaries → bin/
just test     # run all tests
just clean    # remove bin/
```

---

## Just Tasks

```bash
just build          # Build zgard + dukh binaries, copy bash scripts → bin/
just build-zgard    # Build only zgard CLI → bin/zgard
just build-dukh     # Build only dukh daemon → bin/dukh
just copy-scripts   # Copy grazhda / grazhda-init.sh → bin/
just generate       # Regenerate protobuf Go code from proto/dukh.proto
just test           # Run go test ./... across all modules
just fmt            # Run gofmt across all modules
just tidy           # Run go mod tidy for each module
just clean          # Remove the bin/ directory
just help           # Show all available tasks
```

---

## Build Output

| File | Description |
| :--- | :--- |
| `bin/zgard` | Workspace lifecycle CLI |
| `bin/dukh` | Background health monitor daemon |
| `bin/grazhda` | Installer bootstrap script |
| `bin/grazhda-init.sh` | Workspace initialization shell script |

---

## Module Layout

Grazhda is a Go workspace with three modules:

```
grazhda/
├── go.work                 # Go workspace (internal, zgard, dukh)
├── Justfile                # Task automation
├── proto/                  # Protobuf sources
│   └── dukh.proto
├── config.template.yaml    # Workspace config template
├── internal/               # Shared Go module
│   ├── color/              # Terminal colour helpers
│   ├── config/             # Load · validate · render templates
│   ├── executor/           # Shell command interface + mock
│   ├── reporter/           # Per-repo progress + run summary
│   └── workspace/          # Init · purge · pull · targeting
├── zgard/                  # CLI module
│   ├── main.go
│   ├── root.go
│   └── ws/                 # ws init · ws purge · ws pull · ws status · …
├── dukh/                   # gRPC server module
│   ├── cmd/                # dukh start · stop · status · scan
│   ├── proto/              # Generated protobuf (do not edit)
│   └── server/             # gRPC server · monitor · logging
└── bin/                    # Built outputs (created by just build)
```

### Module boundaries

| Module | Import path | Role |
| :--- | :--- | :--- |
| `internal` | `github.com/vhula/grazhda/internal` | Shared domain logic — config, workspace ops, colour, reporting |
| `zgard` | `github.com/vhula/grazhda/zgard` | Thin CLI layer — Cobra commands calling into `internal/` |
| `dukh` | `github.com/vhula/grazhda/dukh` | Background gRPC daemon — workspace health polling |

Both `zgard` and `dukh` depend on `internal/`. They do not depend on each other.

---

## Working with Protobuf

The gRPC contract between `dukh` (server) and `zgard ws status` (client) is defined in `proto/dukh.proto`.

```bash
just generate    # regenerates dukh/proto/*.pb.go from proto/dukh.proto
```

You need `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` installed.

---

## Running Tests

```bash
just test           # all modules
cd internal && go test ./...   # internal only
cd zgard && go test ./...      # zgard only
```
