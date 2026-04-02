# Work in Progress: The README for the Grazhda ecosystem is currently being developed. Stay tuned for updates!

# 🏔️ Grazhda (Ґражда)

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

**Grazhda** is a unified automation ecosystem designed to streamline daily workflows. It creates a "local homestead" that bridges the gap between high-level tools and low-level local filesystem management, reducing context-switching and saving developers hours per week.

Instead of jumping between browser tabs and terminal windows, Grazhda allows you to orchestrate your entire environment through a set of specialized, lightweight tools.

## Table of Contents

- [The Toolkit](#-the-toolkit)
- [Quick Start](#-quick-start)
- [Example Usage](#-example-usage)
- [License](#-license)

-----

## 🛠️ The Toolkit

The ecosystem consists of five specialized components that work together:

| Tool | Language | Role | Status | Description |
| :--- | :------- | :--- | :----- | :---------- |
| **Grazhda** | Bash | **The Installer** | 🚧 In Development | A bash-based tool to install, configure, and manage the entire ecosystem. |
| **Molfar** | Java | **The Brain** | 📅 Planned | The central server that handles integrations, schedules tasks, and runs complex workflows. |
| **Molf** | Java | **The Interface** | 📅 Planned | The CLI tool you use to talk to **Molfar**. |
| **Dukh** | Go | **The Worker** | 📅 Planned | A background server that performs native tasks directly on your filesystem. |
| **Zgard** | Go | **The Command** | 📅 Planned | The CLI tool you use to talk to **Dukh** (local file tasks and system commands). |

-----

## 🚀 Quick Start

### Prerequisites

- Bash (for Grazhda)
- Java 25+ (for Molfar and Molf)
- Go 1.26+ (for Dukh and Zgard)
- Administrative privileges for system access

### 1\. Installation

Download and install Grazhda using the following command:

```bash
curl -fsSL https://get.grazhda.dev | bash
```

This will download the installer script and set up the environment automatically.

### 2\. Configuration

Create a YAML config file defining server ports and API keys (e.g., for integrations).

```bash
./grazhda config --path /path/to/config.yaml
```

### 3\. Running the Servers

```bash
# Start the Molfar server (handles enterprise integrations)
./grazhda start molfar
# Start the Dukh server (handles local filesystem tasks)
./grazhda start dukh
```

```bash
# Start both servers in one command
./grazhda start all
```

To verify, check status:

```bash
./grazhda status
# Or: curl http://localhost:8080/health (for Molfar)
```

### 4\. Using the CLI Tools

-----

## 🧿 Example Usage

### Managing the Enterprise (via `molf`)

```bash
# List all active workflows in Molfar
$ molf wf list

# Create a new workflow for daily backups
$ molf wf create backup --schedule 'daily' --action 'zgard fs backup /home/user'
```

### Managing the Local System (via `zgard`)

```bash
# Bootstrap the local workspace
zgard ws init
```

```bash
# Purge the local workspace
zgard ws purge

# Run a custom script
zgard run --script my-script.sh
```

-----

## 📄 License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
