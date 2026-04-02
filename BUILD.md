# Grazhda Monorepo - Using Justfile

This project uses Justfile for task automation and building.

## Installation

Install Just first:

```bash
# Using Cargo (if you have Rust installed)
cargo install just

# Or download a pre-built binary from:
# https://github.com/casey/just/releases
```

## Available Tasks

Run any task with `just <task-name>`:

```bash
just build          # Build Dukh and Zgard binaries
just build-dukh     # Build only the Dukh CLI
just build-zgard    # Build only the Zgard CLI
just test           # Run tests for all modules
just zip            # Build everything and create bin/grazhda.zip
just clean          # Remove all built binaries
just help           # Show all available tasks
```

## Quick Start

```bash
# Build all modules to the bin/ directory
just build

# Test everything
just test

# Create a distributable zip file
just zip

# Clean up generated files
just clean
```

## Output

- Built binaries are placed in `bin/dukh` and `bin/zgard`
- Zip archive is created at `bin/grazhda.zip`

## Project Structure

```
grazhda/
├── go.work              # Go workspace definition
├── justfile             # Task automation
├── Dockerfile           # Container build
├── dukh/                # CLI module 1
│   ├── go.mod
│   └── main.go
├── zgard/               # CLI module 2
│   ├── go.mod
│   └── main.go
└── bin/                 # Built binaries (created by build)
```
