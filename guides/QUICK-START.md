# Quickstart for Grazhda

**Get your entire dev environment cloned, configured, and ready in five minutes.**

If you manage multiple repositories across projects, Grazhda automates the repetitive setup. This quickstart shows you how to install Grazhda, configure your workspace, and clone all your repositories with a single command.

## What you'll accomplish

- Install `zgard` (workspace CLI) and `dukh` (health monitor)
- Configure your workspace in a single YAML file
- Clone all repositories at once
- Check workspace health

**Time: ~5 minutes**  
**Prerequisites:** `bash`, `curl`, `git`, `just`, `protoc`, Go `1.26+`

---

## Step 1: Install Grazhda

Grazhda installs itself from source and places binaries in `$GRAZHDA_DIR/bin/`
(default: `$HOME/.grazhda/bin/`).

```bash
curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda-install.sh | bash
```

To install to a custom location, set `GRAZHDA_DIR` first:

```bash
export GRAZHDA_DIR="$HOME/devtools/grazhda" && curl -s https://raw.githubusercontent.com/vhula/grazhda/refs/heads/main/grazhda-install.sh | bash
```

The installer will:
1. Clone the Grazhda sources to `$GRAZHDA_DIR/sources`
2. Build `zgard` and `dukh` from source
3. Copy binaries to `$GRAZHDA_DIR/bin`
4. Show you a snippet to add to your shell profile

**After installation,** add the snippet to your `~/.bashrc` or `~/.zshrc`:

```bash
# Use the same directory you installed with:
# export GRAZHDA_DIR="$HOME/devtools/grazhda"
# Default:
export GRAZHDA_DIR="$HOME/.grazhda"
[[ -s "${GRAZHDA_DIR}/bin/grazhda-init.sh" ]] && source "${GRAZHDA_DIR}/bin/grazhda-init.sh"
```

Then reload your shell:

```bash
source ~/.bashrc
# or: source ~/.zshrc
```

Verify the installation:

```bash
zgard --help
# zgard manages local workspace lifecycle — init, purge, and pull repositories.
```

---

## Step 2: Configure your workspace

Create a configuration file that describes your repositories:

```bash
cp "$GRAZHDA_DIR/sources/config.template.yaml" "$GRAZHDA_DIR/config.yaml"
```

Open it in your editor and replace the example repositories with your own:

```bash
grazhda config --edit
```

### Minimal example

Here's a simple `config.yaml` with one workspace and two projects:

```yaml
editor: vim

dukh:
  host: localhost
  port: 50501
  monitoring:
    period_mins: 5

workspaces:
  - name: default
    default: true
    path: ~/workspace
    clone_command_template: "git clone --branch {{.Branch}} git@github.com:myorg/{{.RepoName}} {{.DestDir}}"
    structure: tree

    projects:
      - name: backend
        branch: main
        repositories:
          - name: api
          - name: auth-service
            branch: dev              # override default branch
          - name: database

      - name: frontend
        branch: main
        repositories:
          - name: web-app
          - name: mobile-app
```

**Key fields:**

| Field | Meaning |
|---|---|
| `default: true` | Makes this the workspace targeted by default (`zgard ws init` with no flags) |
| `path: ~/workspace` | Where repositories will be cloned on your machine |
| `clone_command_template` | How to clone each repo; `{{.RepoName}}` and `{{.Branch}}` are filled in by zgard |
| `projects:` | Groups of related repositories |
| `repositories:` | List of repos to clone in each project |

For full schema details, see [Configuration Reference](CONFIG.md).

---

## Step 3: Clone all repositories

Now clone every repository in your workspace with a single command:

```bash
zgard ws init
```

You'll see a live progress report:

```
Workspace: default
  Project: backend
    ✓ api             — cloned (main)
    ✓ auth-service    — cloned (dev)
    ✓ database        — cloned (main)
  Project: frontend
    ✓ web-app         — cloned (main)
    ✓ mobile-app      — cloned (main)

✓ 5 cloned  ⏭ 0 skipped  ✗ 0 failed
```

All repos are now cloned at `~/workspace/<project>/<repo>`.

---

## Step 4: Start the health monitor

Start `dukh` to continuously watch your workspace for branch drift and missing repositories:

```bash
dukh start
```

You'll see:

```
✓ dukh started (pid 12345)
```

---

## Step 5: Check workspace health

Query the live health snapshot maintained by `dukh`:

```bash
zgard ws status
```

Sample output:

```
Dukh  running  •  uptime: 2m 15s

Workspace: default
  Project: backend
    ✓ api             main → main
    ✓ auth-service    dev → dev
    ✓ database        main → main
  Project: frontend
    ✓ web-app         main → main
    ✓ mobile-app      main → main

✓ 5 aligned  ⚠ 0 drifted  ✗ 0 missing
```

The monitor checks every 5 minutes. For an immediate fresh scan, use:

```bash
zgard ws status --rescan
```

---

## Common next steps

**Pull latest changes on all repos:**

```bash
zgard ws pull
```

**Run a command across all repositories:**

```bash
zgard ws exec "make test"
```

**Check repository state (uncommitted changes, ahead/behind):**

```bash
zgard ws diff
```

**Manage multiple workspaces:**

```bash
# Configure a second workspace called "personal"
# in config.yaml (set default: false or omit it)

# Then target it explicitly:
zgard ws init --name personal
zgard ws pull --name personal
```

**Filter by project or repository:**

```bash
# Clone only the "backend" project
zgard ws init --project-name backend

# Clone repos whose names contain "auth"
zgard ws init --project-name backend --repo-name auth
```

**Use tags to group repositories:**

```yaml
# In config.yaml, tag your repos:
projects:
  - name: backend
    tags: [api, core]         # project-level tags
    repositories:
      - name: auth
        tags: [auth, critical]  # repo-level tags (merged with project)
```

```bash
# Then filter by tag:
zgard ws init -t critical   # Only repos tagged "critical"
zgard ws pull -t api        # Only repos tagged "api"
```

---

## Troubleshooting

**Issue:** `zgard: command not found`

- Ensure the snippet was added to your shell profile and the shell was reloaded.
- Check that `$GRAZHDA_DIR/bin` is on your `$PATH`:

```bash
echo $PATH | grep grazhda
```

**Issue:** Clone fails with "repository not found"

- Check your `clone_command_template` in `config.yaml` — it must be valid for your Git provider.
- Verify `{{.RepoName}}` and `{{.Branch}}` are substituted correctly:

```bash
zgard ws init --dry-run
```

**Issue:** `dukh` isn't running

```bash
dukh status
# ○  dukh: not running
```

Start it again:

```bash
dukh start
```

For detailed logs, check `$GRAZHDA_DIR/logs/dukh.log`.

---

## What to read next

- **[CLI Reference](CLI.md)** — Full command documentation, targeting flags, and examples
- **[Configuration Reference](CONFIG.md)** — Detailed schema, clone templates, structure modes
- **[Architecture](architecture.md)** — How zgard and dukh work together

---

## Tips

- **Dry-run before executing:** Use `--dry-run` to preview any action without modifying your workspace:

```bash
zgard ws init --dry-run
zgard ws pull --dry-run
```

- **Parallel execution:** Speed up operations across many repositories with `--parallel`:

```bash
zgard ws init --parallel      # clone all repos concurrently
zgard ws pull --parallel
zgard ws exec "make test" --parallel
```

- **Upgrade Grazhda:** Keep your installation up-to-date:

```bash
grazhda upgrade
```

---

**Happy coding!** 🚀
