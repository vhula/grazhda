# Grazhda Monorepo — Build Reference

This project uses [Just](https://github.com/casey/just) for task automation.

## Install Just

```bash
# Using Cargo (if you have Rust installed)
cargo install just

# Or download a pre-built binary from:
# https://github.com/casey/just/releases
```

## Available Tasks

```bash
just build          # Build zgard binary + copy bash scripts to bin/
just build-zgard    # Build only the zgard CLI → bin/zgard
just copy-scripts   # Copy grazhda / grazhda-init.sh to bin/
just test           # Run go test ./... across all modules
just fmt            # Run gofmt across all modules
just tidy           # Run go mod tidy for each module
just clean          # Remove the bin/ directory
just help           # Show all available tasks
```

## Quick Start

```bash
# Build everything
just build

# Test everything
just test

# Clean up
just clean
```

## Output

Built artifacts are placed in `bin/`:

| File | Description |
| :--- | :--- |
| `bin/zgard` | zgard CLI binary |
| `bin/grazhda` | Installer bootstrap script |
| `bin/grazhda-init.sh` | Workspace init shell script |

## Project Structure

```
grazhda/
├── go.work              # Go workspace (internal, zgard, dukh)
├── Justfile             # Task automation
├── config.template.yaml # Workspace config template
├── internal/            # Shared Go module
│   ├── color/           # Terminal color helpers
│   ├── config/          # Config loading + validation
│   ├── executor/        # Shell command executor interface
│   ├── reporter/        # Progress output + run summary
│   └── workspace/       # Init/Purge/Pull domain logic
├── zgard/               # zgard CLI module
│   ├── main.go
│   ├── root.go
│   └── ws/              # ws init / purge / pull commands
├── dukh/                # gRPC server for workspace monitoring
└── bin/                 # Built outputs (created by just build)
```
