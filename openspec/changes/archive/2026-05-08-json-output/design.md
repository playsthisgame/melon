## Context

`melon list` and `melon info` currently produce human-readable output (formatted tables, styled text via lipgloss). There is no way for external tooling — editor extensions, CI scripts, dashboards — to consume their output reliably. A `--json` flag on these two commands is the minimal addition that unlocks machine-readable output without touching the install pipeline or data model.

## Goals / Non-Goals

**Goals:**

- Add `--json` to `melon list`: emit a JSON array of installed dep objects read from `melon.lock`
- Add `--json` to `melon info`: emit a JSON object with dep metadata
- Suppress all TUI/spinner output when `--json` is active
- Write errors to stderr as `{"error": "..."}` in JSON mode so callers can distinguish them from normal output

**Non-Goals:**

- `--json` on `melon install`, `melon add`, `melon remove`, or `melon search`
- A global `--json` persistent flag — only the two commands that make sense get it
- Changing the default (human-readable) output of any command

## Decisions

### Per-command flag, not persistent root flag

Adding `--json` as a persistent flag on the root command would imply every command supports it, which is misleading when most don't. Per-command flags are explicit and honest about scope. The two commands are independent enough that sharing flag state via a package-level var isn't needed — each command reads its own flag directly.

### Single JSON write at the end, not streaming

Both `list` and `info` complete all their work before printing anything. There is no streaming data model here, so emitting one JSON document at the end is the right call. This keeps the output valid JSON even if it is piped or captured.

### JSON shape follows the lock file schema

For `melon list`, each element in the array maps directly to a `lockfile.Entry` field set (name, version, git\_tag, repo\_url, subdir, entrypoint, tree\_hash). No computed or derived fields. This keeps the shape stable and tied to the data melon actually tracks.

For `melon info`, the object includes fields available at info-time: name, description, author (from index if available), latest\_version, all\_versions (or branches if no semver tags). This is a superset of what the lock file stores.

### Errors to stderr as `{"error": "..."}`

When `--json` is active, human-readable error strings on stderr would break callers that capture both streams. A JSON-shaped error on stderr lets callers detect and parse failures without special-casing the exit code alone.

**Alternative considered:** always exit non-zero and let callers rely solely on exit code. Rejected — callers need the error message to surface useful diagnostics.

## Risks / Trade-offs

- **Shape stability**: Once external tools rely on the JSON output, changing field names is a breaking change. The lock file schema is already stable, so `melon list` output is low risk. `melon info` output depends on index data which could evolve. → Mitigate by keeping the shape minimal and documented.
- **`--json` + `--pending` / `--check` on `list`**: The `--pending` and `--check` flags add extra sections to list output. In JSON mode these should be represented as distinct keys in the top-level object (e.g., `{"installed": [...], "pending": [...], "check": [...]}`). → Design the JSON envelope to accommodate this from the start rather than retrofitting it.
