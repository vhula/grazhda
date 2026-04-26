# Grazhda

[![Build](https://github.com/vhula/grazhda/actions/workflows/just.yml/badge.svg)](https://github.com/vhula/grazhda/actions/workflows/just.yml)
[![Release](https://img.shields.io/github/v/release/vhula/grazhda)](https://github.com/vhula/grazhda/releases)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)

Grazhda is a multi-repository workspace lifecycle toolkit. You describe workspaces in YAML, then use:
- `zgard` to clone, pull, inspect, and manage repos at scale
- `dukh` to monitor workspace health in the background
- `grazhda` to install/upgrade/manage the toolchain itself

---

## What the project does

Grazhda automates day-to-day operations for teams working across many repositories:

1. Declarative workspace config (`config.yaml`) with workspaces, projects, repositories, and branch rules.
2. Bulk operations (`zgard ws init`, `pull`, `exec`, `stash`, `checkout`, `search`, `diff`, `stats`).
3. Background health monitoring (`dukh`) for branch drift and missing repositories.
4. Declarative package management (`zgard pkg`) with layered registries:
   - Global registry: `$GRAZHDA_DIR/.grazhda.pkgs.yaml` (managed by install/upgrade)
   - Local registry: `$GRAZHDA_DIR/registry.pkgs.local.yaml` (user-managed)

---

## Why the project is useful

- **Fast environment setup:** clone an entire workspace with one command.
- **Consistent developer experience:** workspace state lives in versioned YAML, not tribal knowledge.
- **Safer large-scale changes:** dry-run, targeting filters, and clear per-repo status output.
- **Health visibility:** `dukh` gives near-real-time workspace health snapshots.
- **Flexible package control:** local package registry can override global definitions and supports version-aware dependencies (`depends_on: name` or `name@version`).

---

## How users can get started

### 1. Prerequisites

- `bash`
- `curl`
- `git`
- `just`
- `protoc`
- Go `1.26+`

### 2. Install

```bash
curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda-install.sh | bash
```

This installs into `$GRAZHDA_DIR` (default: `$HOME/.grazhda`) and builds binaries from source.

To install into a custom directory, set `GRAZHDA_DIR` before running the installer:

```bash
export GRAZHDA_DIR="$HOME/devtools/grazhda" && curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda-install.sh | bash
```

### 3. Configure your workspace

```bash
cp config.template.yaml "$GRAZHDA_DIR/config.yaml"
grazhda config --edit
zgard config validate
```

### 4. Run core workflow commands

```bash
zgard ws init      # clone repositories from the default workspace
zgard ws pull      # pull updates across repositories
dukh start         # start background health monitor
zgard ws status    # show health snapshot
```

### 5. Use package registry layering

Install/purge resolves packages from both registries:

- `$GRAZHDA_DIR/.grazhda.pkgs.yaml` (global)
- `$GRAZHDA_DIR/registry.pkgs.local.yaml` (local, optional)

Local entries override global entries when `name+version` match exactly.

```bash
# Install from merged registries
zgard pkg install --all
zgard pkg install --name jdk@17.0.8-tem

# Add/update package interactively in local registry
zgard pkg register

# Remove from local registry
zgard pkg unregister --name jdk --version 17.0.8-tem
zgard pkg unregister --name jdk
zgard pkg unregister --all
```

For a full guided setup, see [guides/QUICK-START.md](guides/QUICK-START.md).

---

## Where users can get help

- [guides/QUICK-START.md](guides/QUICK-START.md) — fast installation and first-run flow
- [guides/CLI.md](guides/CLI.md) — complete command reference
- [guides/CONFIG.md](guides/CONFIG.md) — configuration schema and examples
- [guides/DEVELOPMENT.md](guides/DEVELOPMENT.md) — local development, build, and test workflow
- [guides/architecture.md](guides/architecture.md) — system design and module boundaries

Issue tracker: [github.com/vhula/grazhda/issues](https://github.com/vhula/grazhda/issues)

---

## Who maintains and contributes

Maintained by the Grazhda project maintainers (repository owner: [@vhula](https://github.com/vhula)).

### Contributing

1. Fork and clone the repository.
2. Create a feature branch.
3. Run local checks:

```bash
just build
just test
```

4. Open a pull request with a clear description and test evidence.

Development conventions and module layout are documented in [guides/DEVELOPMENT.md](guides/DEVELOPMENT.md).

---

## License

GNU GPL v3 — see [LICENSE](LICENSE).
