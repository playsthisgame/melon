## Why

There is no way to see which installed skills have newer versions available without actually running `melon update` — a read-only check command is missing. `melon outdated` fills this gap: it shows what's stale without modifying anything, letting users make informed decisions before upgrading.

## What Changes

- Adds a new `melon outdated` command that:
  - Reads `melon.yaml` and `melon.lock` (no network calls until needed)
  - For each locked dep, resolves the latest version satisfying its constraint and the absolute latest tag
  - Prints a table showing: dep name, current locked version, latest compatible version, and absolute latest (if outside constraint)
  - Exits with code 1 if any dep is outdated (useful in CI), code 0 if everything is current
  - Skips branch-pinned deps with a note
  - Prints "All skills are up to date." when nothing is outdated

## Capabilities

### New Capabilities

- `outdated-command`: The `melon outdated` command — reads lock file, resolves latest versions, prints a comparison table, exits 1 if outdated

### Modified Capabilities

_(none)_

## Impact

- New file: `internal/cli/outdated_cmd.go`
- Reads from existing `manifest`, `lockfile`, and `fetcher` packages — no new packages needed
- `internal/cli/cli.go` — register new `outdated` subcommand
- Exit code 1 on outdated (CI-friendly); no writes to disk
