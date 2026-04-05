## Why

Users who install melon should be able to invoke it by its full name (`melon`) and not just the abbreviated form (`mln`). The short alias aids power users, but the full name is more discoverable—especially for new users who find the tool by name and expect `melon --version` to just work.

## What Changes

- A second binary `melon` is built and distributed alongside `mln` in every release archive
- The npm package gains a second `bin` entry so `melon <command>` works after `npm install -g`
- A new `cmd/melon` entrypoint is added that shares the same CLI logic as `cmd/mln`
- Both binaries behave identically in all respects; `melon` is a full alias, not a shim

## Capabilities

### New Capabilities
- `melon-cli-alias`: The `melon` binary is distributed and installed alongside `mln`, supporting all the same subcommands and flags

### Modified Capabilities
- `release-artifacts`: Release archives and the npm postinstall must now deliver and register both `mln` and `melon` binaries

## Impact

- `cmd/melon/` — new Go entrypoint package
- `cmd/mln/` — CLI logic extracted into a shared internal package so both entrypoints stay in sync
- `.goreleaser.yaml` — second build entry producing the `melon` binary
- `npm/package.json` — second `bin` entry (`"melon": "./bin/mln.js"`)
- `npm/postinstall.js` — creates a `melon` binary/symlink alongside `mln` after download
