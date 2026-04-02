## Why

The tool is named `melon` and its binary is `mln`, but all of the file artifacts it creates use the `mln` prefix (`.mln/`, `mln.yaml`, `mln.lock`). Using the full `melon` name for artifacts makes the project more discoverable and self-documenting — someone browsing a repo can immediately understand what `.melon/` and `melon.yml` belong to without knowing the shorthand.

## What Changes

- `.mln/` store directory renamed to `.melon/`
- `mln.yaml` manifest file renamed to `melon.yml`
- `mln.lock` lockfile renamed to `melon.lock`
- All code references to these paths updated accordingly
- `.gitignore` patterns updated (`.melon/` instead of `.mln/`)

## Capabilities

### New Capabilities

<!-- None — this is a pure rename with no behavior changes. -->

### Modified Capabilities

- `artifact-naming`: The file and directory names produced and consumed by all mln commands change from `mln.*` / `.mln/` to `melon.*` / `.melon/`.

## Impact

- `internal/store/store.go` — `StoreDir` constant changes from `.mln` to `.melon`
- `cmd/mln/install_cmd.go` — manifest path uses `melon.yml`, lock path uses `melon.lock`
- `cmd/mln/init_cmd.go` — `.mln/` references change to `.melon/`
- Any tests that reference the old names
- `.gitignore` — `.mln/` entry changes to `.melon/`
- Documentation / comments referencing the old names
