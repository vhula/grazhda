# Grazhda — Feature Ideas

---

## 🔍 Workspace Inspection

- **`zgard ws search <pattern>`** *(already planned)* — grep/glob across all repos in a workspace
- **`zgard ws diff`** — show repos with uncommitted changes, unpushed commits, or ahead/behind upstream counts
- **`zgard ws stats`** — aggregate view: last commit date, commit count per repo, contributor counts

---

## ⚡ Cross-Repo Operations

- **`zgard ws run <command>`** *(already planned)* — fan out any shell command across all repos (`zgard ws run "make test"`)
- **`zgard ws stash`** / **`zgard ws checkout <branch>`** — coordinated git ops across all repos in a workspace

---

## 📸 Workspace Snapshots

- **`zgard ws snapshot [--name <name>]`** — capture the current branch of every repo to a `.snapshot` file
- **`zgard ws restore <snapshot>`** — checkout every repo to the branch it had at snapshot time
- Great for context-switching between feature work across multiple repos

---

## 🔔 dukh Enhancements

- **Desktop notifications** — dukh emits a native OS notification when branch drift is detected
- **Webhook events** — dukh POSTs drift/recovery events to a configured URL (for Slack, PagerDuty, etc.)
- **Configurable poll interval** — `monitoring.period_mins` already exists in config but the implementation uses a hardcoded 30s constant

---

## ⚙️ Config & Lifecycle Hooks

- **`zgard config validate`** — validate `config.yaml` and exit; useful as a pre-commit hook
- **`zgard config init`** *(already planned)* — scaffold a starter `config.yaml`
- **Lifecycle hooks** in config — run a script after init/pull succeeds:
  ```yaml
  on_init_success: "make install-deps"
  on_pull_success: "make build"
  ```

---

## 🏷️ Targeting Enhancements

- **Tag-based targeting** — tag repos or projects in config, then target by tag:
  ```yaml
  repositories:
    - name: api
      tags: [backend, critical]
  ```
  ```
  zgard ws pull --tag critical
  ```
- **`zgard ws open --ide <vscode|idea>`** — open all project directories in an IDE

---

## 📦 Distribution

- **GitHub Releases with pre-built binaries** — remove the Go build requirement for end users
- **Homebrew formula** — `brew install grazhda`
- **Shell completions** *(already planned)* — bash/zsh/fish

---

## 🤖 Phase 3+ (Molfar vision)

- **`zgard ws agent`** — run an LLM agent scoped to a workspace for cross-repo analysis (already in PRD vision)
- **Remote config source** — pull `config.yaml` from a URL or git repo, enabling team-wide workspace config sharing without manual file distribution

---

## Recommended Near-Term Priorities

The highest-value near-term picks that extend the existing architecture cleanly without new infrastructure:

1. **`zgard ws diff`** — immediate daily utility
2. **Workspace snapshots** — enables safe context-switching across branch sets
3. **Lifecycle hooks** — bridges workspace setup and build tooling
4. **Tag-based targeting** — unlocks fine-grained operations on large workspaces
