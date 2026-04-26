# Known Gaps

> [← Back to README](../README.md)

This file records source and documentation gaps found during the current
source/docs alignment pass. It intentionally describes the current behavior,
not the desired future behavior.

## Source gaps

| Area | Current behavior | Impact |
| :--- | :--- | :--- |
| `zgard ws status` targeting | `--name` filters the dukh snapshot by workspace. Inherited `--project-name`, `--repo-name`, and `--tag` flags are accepted but not applied. | Users may expect the same filters as `init`, `pull`, `exec`, `search`, `diff`, `stats`, and `list`. |
| `zgard ws init --no-confirm` | The flag is accepted, but `init` has no confirmation prompt. | The flag is effectively a no-op for `init`; only `purge` uses it today. |
| `zgard ws search --glob --regex` | Both flags can be supplied; glob mode wins because content regex is skipped when `--glob` is true. | The CLI does not reject an ambiguous mode selection. |
| `workspace.structure` validation | Runtime treats any value other than `list` as `tree`; validation does not reject unknown values. | A typo silently falls back to tree-style paths. |
| `zgard.config` and `general.*` config fields | Fields are parsed and available through `zgard config get`, but no workspace/package operation reads them. | They are metadata/inspection fields, not operational settings. |

## Documentation boundaries

The source-aligned reference docs are:

- [README.md](../README.md)
- [QUICK-START.md](QUICK-START.md)
- [CLI.md](CLI.md)
- [CONFIG.md](CONFIG.md)
- [DEVELOPMENT.md](DEVELOPMENT.md)

The PRD, epic, UX, and architecture documents are planning/history documents.
They can contain intended behavior or earlier implementation notes and should
not be treated as command references without checking the current source.
