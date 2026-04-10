## Why

Users have no way to inspect the current state of their installed skills or identify gaps — whether skills in `melon.yaml` haven't been installed yet or whether expected tool files are missing from disk. A `melon list` command provides a quick audit view so users always know what's actually installed and what needs attention.

## What Changes

- Add a new `melon list` command that shows all installed skills
- Add `--pending` flag to show skills declared in `melon.yaml` but not yet installed
- Add `--check` flag to verify that expected tool directories/files exist on disk and report any missing ones
- The default output lists installed skills with their name, version, and source

## Capabilities

### New Capabilities

- `list-command`: The `melon list` CLI command with `--pending` and `--check` flags for auditing skill installation state

### Modified Capabilities

## Impact

- New file: `internal/cli/list_cmd.go`
- Reads `melon.yaml` to enumerate declared skills
- Reads the installed skill registry (same source used by `melon remove`) to enumerate installed skills
- Inspects tool target directories on disk for `--check` mode
- `README.md` updated with a `### melon list` section under Commands
- No breaking changes; purely additive
