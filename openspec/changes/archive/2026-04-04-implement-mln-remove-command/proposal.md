## Why

The `mln remove` command stub exists in `main.go` but its handler is a no-op. Users who add a dependency with `mln add` have no way to remove it from `melon.yml`, clean up its symlinks from agent directories, or update `melon.lock` — they must edit files manually.

## What Changes

- Implement `runRemove` in a new `remove_cmd.go` file
- Remove the named dependency from `melon.yml`
- Remove the corresponding symlink(s) from all agent directories (derived from `agent_compat` or `outputs`)
- Re-run `mln install` so `melon.lock` is regenerated to reflect the removal
- Print an error if the dependency is not found in `melon.yml`

## Capabilities

### New Capabilities

- `remove-command`: `mln remove <name>` removes a dependency from `melon.yml`, unlinks it from agent directories, and updates `melon.lock` via a full install

### Modified Capabilities

_(none — no existing spec requirements are changing)_

## Impact

- New file: `cmd/mln/remove_cmd.go`
- `cmd/mln/main.go`: replace the TODO stub with a call to `runRemove`
- Uses existing `manifest`, `placer`, `agents`, `store`, and `lockfile` packages — no new internal packages needed
