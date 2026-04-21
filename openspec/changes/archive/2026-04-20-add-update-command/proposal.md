## Why

Melon has `add` and `remove` but no way to bump existing dependencies to newer compatible versions. Users who want to stay current must manually edit `melon.yaml` and re-run `install`, with no visibility into what updates are actually available.

## What Changes

- Adds a new `melon update` command with two modes:
  - **Interactive mode** (`melon update` with no args, TTY): shows a multi-select list of all declared dependencies; first option is "Update all"; runs install pipeline on the selected set
  - **Targeted mode** (`melon update <dep>`): updates a single named dependency; errors clearly if the dep is not in `melon.yaml`
- In both modes, if no updates are available (all deps already at latest compatible version), a clear message is shown and the command exits without modifying any files
- The update resolves the latest version satisfying the existing semver constraint in `melon.yaml` — it does not widen constraints

## Capabilities

### New Capabilities

- `update-command`: The `melon update [dep]` command — interactive multi-select or targeted single-dep update, version resolution, install pipeline integration, and up-to-date messaging

### Modified Capabilities

- `tui-spinners`: Update command adds a new spinner context (resolving updates) that follows the same TTY-detection pattern as add/remove

## Impact

- New file: `internal/cli/update_cmd.go`
- Calls into existing `resolver`, `fetcher`, `placer`, `lockfile`, and `manifest` packages — no new packages needed
- Bubbletea multi-select reuses the same pattern as `remove-interactive`
- `internal/cli/cli.go` — register new `update` subcommand
