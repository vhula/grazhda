# 🏔️ Grazhda (Grazhda)

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

Grazhda is a local automation ecosystem focused on workspace lifecycle, filesystem operations, and repeatable developer setup.

## Table of Contents

- [Zgard + Dukh Design](#zgard--dukh-design)
  - [Responsibilities](#responsibilities)
  - [Flow](#flow)
- [Toolkit Status](#toolkit-status)
- [Quick Start](#quick-start)
  - [Prerequisites](#prerequisites)
  - [Install](#install)
  - [Start Dukh](#start-dukh)
  - [Run Zgard Commands](#run-zgard-commands)
- [License](#license)

## Zgard + Dukh Design

The current foundation of Grazhda is built around two Go services:

- **`dukh` (server):** a gRPC worker that manages workspaces from `config.yaml`.
- **`zgard` (client):** a CLI that reads config and sends commands to `dukh`.

### Responsibilities

- **`dukh`**
  - Initializes all configured workspaces.
  - Creates workspace directories and project subdirectories.
  - Builds and executes clone commands from each workspace `clone_command_template`.
  - Purges all configured workspaces.
  - Exposes gRPC operations for init, purge, and get.

- **`zgard`**
  - Reads `dukh.host` and `dukh.port` from config.
  - Calls `dukh` over gRPC.
  - Provides user-facing commands such as:
	- `zgard ws init`
	- `zgard ws purge`

### Flow

1. `zgard` loads `${GRAZHDA_DIR}/config.yaml`.
2. `zgard` connects to `dukh` via gRPC.
3. `dukh` executes filesystem operations using the configured workspace layout.

## Toolkit Status

| Tool | Language | Role | Status |
| :--- | :------- | :--- | :----- |
| **Grazhda** | Bash | Installer | 🚧 In Development |
| **Dukh** | Go | Worker (gRPC server) | 🚧 In Development |
| **Zgard** | Go | Command CLI | 🚧 In Development |
| **Molfar** | Java | Brain | 📅 Planned |
| **Molf** | Java | Interface CLI | 📅 Planned |

## Quick Start

### Prerequisites

- Bash
- Go 1.26+
- Administrative privileges (if required by your environment)

### Install

```bash
curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda.sh | bash
```

### Start Dukh

```bash
dukh start
```

### Run Zgard Commands

```bash
zgard ws init
zgard ws purge
```

## License

This project is licensed under the GNU General Public License v3.0. See `LICENSE`.
