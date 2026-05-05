# Grazhda

[![Build](https://github.com/vhula/grazhda/actions/workflows/just.yml/badge.svg)](https://github.com/vhula/grazhda/actions/workflows/just.yml)
[![Release](https://img.shields.io/github/v/release/vhula/grazhda)](https://github.com/vhula/grazhda/releases)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)

Grazhda is a multi-repository workspace lifecycle toolkit powered by GitOps. You describe workspaces in YAML, then use:
- `zgard` to clone, pull, inspect, and manage repos at scale
- `dukh` to monitor and sync workspaces in the background with the config
- `grazhda` to install/upgrade/manage the toolchain itself

