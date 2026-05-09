## Why

Tooling built on top of melon — wrappers, dashboards, editor extensions — needs structured output from read-only commands. `melon list` and `melon info` are pure queries with well-defined data shapes, making them the right starting point for machine-readable output.

## What Changes

- Add a `--json` flag to `melon list` that prints installed dependencies as a JSON array instead of the human-readable table
- Add a `--json` flag to `melon info` that prints a single dependency's details as a JSON object instead of formatted text
- When `--json` is set, suppress all TUI output (spinners, progress bars) and write a single JSON document to stdout
- Errors are written to stderr as `{"error": "..."}` when `--json` is set

## Capabilities

### New Capabilities

- `json-output`: Structured JSON output mode for `melon list` and `melon info` via a `--json` flag

### Modified Capabilities

- `list-command`: `melon list` gains a `--json` flag that changes output format
- `skill-info`: `melon info` gains a `--json` flag that changes output format

## Impact

- `internal/cli/list_cmd.go` — add `--json` flag, branch on output format
- `internal/cli/info_cmd.go` — add `--json` flag, branch on output format
- No changes to manifest, lock file, or install pipeline
- No breaking changes — flag is opt-in, default behavior unchanged
