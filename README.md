# Grazhda

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Phase](https://img.shields.io/badge/Phase-2%20%E2%80%94%20dukh-brightgreen)](https://github.com/vhula/grazhda)

**One command to clone your entire dev environment. Exactly how you left it.**

---

## 🎬 Demo

![Grazhda demo](grazhda-demo.gif)

---

## 🔥 The Problem

You change laptops. You onboard a new teammate. You re-provision a machine after a crash. Then comes the ritual: remember which repos go where, which branches are right for each project, which SSH key format each remote uses. Repeat forty times. Miss one. Break something.

Developer environments are configuration. Configuration should be code.

---

## ⚡ The Solution

`zgard` reads a single YAML file that describes your workspaces — where they live on disk, which repos belong to each project, which branch each repo tracks. Then it does the work.

```
zgard ws init
```

```
Workspace: default
  Project: backend
    ✓ api          — cloned (main)
    ✓ auth         — cloned (dev)
    ✓ gateway      — cloned (main)
  Project: infra
    ✓ terraform    — cloned (main)
    ⏭ k8s-configs  — already exists, skipped

✓ 4 cloned  ⏭ 1 skipped  ✗ 0 failed
```

Every failure is reported with the actual git error — no more hunting for `exit status 128`.

---

## 🚀 Quick Start

### Install

```bash
curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda.sh | bash
```

The installer builds `zgard` and `dukh` from source and places them in `$GRAZHDA_DIR/bin/`.

### Configure

```bash
cp config.template.yaml "$GRAZHDA_DIR/config.yaml"
$EDITOR "$GRAZHDA_DIR/config.yaml"
```

### Run

```bash
zgard ws init            # clone everything in the default workspace
zgard ws pull            # pull latest on all repos
zgard ws list            # show clone status for all repos
dukh start               # start background workspace health monitor
zgard ws status          # check workspace health
zgard config validate    # validate configuration
```

> **Prerequisites:** `bash`, `curl`, `git`, `just`, `protoc`, Go `1.26+`

---

## 🗂️ Workspace Concept

A **workspace** is the central organizing unit in Grazhda. It describes a group of related projects that live together on your local machine — their disk locations, their repositories, and the exact branch each repository should track.

```
workspace: default
  └── project: backend
  │     ├── repo: api          (branch: main)
  │     ├── repo: auth         (branch: dev)
  │     └── repo: gateway      (branch: main)
  └── project: infra
        └── repo: terraform    (branch: main)
```

The three Grazhda tools each have a distinct role in the workspace lifecycle:

| Tool | Role |
| :--- | :--- |
| **`grazhda`** | Manages your Grazhda installation — upgrade, configure |
| **`zgard`** | Manages workspace operations on demand — clone, pull, purge, inspect health |
| **`dukh`** | Monitors workspace health continuously in the background — detects branch drift and missing repos |

`dukh` is the *overseer*: it runs silently in the background and keeps a live health snapshot of every workspace. `zgard` is the *operator*: you run it when you want to act — clone repositories, pull changes, or query the health snapshot that `dukh` maintains. Both tools read the same `config.yaml` and share the same workspace model.

---

## 📖 Documentation

| Document | What's inside |
| :--- | :--- |
| **[Quickstart](QUICK-START.md)** | 5-minute setup guide — install, configure, clone all repos |
| **[CLI Reference](docs/CLI.md)** | All `grazhda`, `zgard`, and `dukh` commands — targeting flags, subcommands, sample output |
| **[Configuration](docs/CONFIG.md)** | `config.yaml` schema, clone template variables, structure modes, field reference |
| **[Development](docs/DEVELOPMENT.md)** | Module layout, Just tasks, build/test instructions, protobuf workflow |

### Design Documents

| Document | Description |
| :--- | :--- |
| [Product Requirements](docs/prd.md) | Functional requirements and acceptance criteria |
| [Architecture](docs/architecture.md) | System design, gRPC contracts, and data flow |
| [UX Design Specification](docs/ux-design-specification.md) | CLI output formatting, colors, and layout rules |
| [Epics — Workspace](docs/epics.md) | User stories for zgard workspace features |
| [Epics — Dukh](docs/epics-dukh.md) | User stories for dukh server features |

---

## 🗺️ Roadmap

| Tool | Role | Status |
| :--- | :--- | :---: |
| **zgard** | Workspace lifecycle CLI | ✅ Phase 1 |
| **dukh** | Background gRPC workspace monitor | ✅ Phase 2 |
| **grazhda** | Installer + management script | ✅ |
| **Molfar** | Orchestration server | 📅 Phase 3 |
| **Molf** | Orchestrator CLI | 📅 Phase 3 |

---

## 📄 License

GNU GPL v3 — see [`LICENSE`](LICENSE).
