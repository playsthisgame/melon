## Why

Users have no way to check which version of `mln` they have installed, and `mln init` hardcodes `0.1.0` as the initial version in `melon.yml` rather than reflecting the actual CLI version. Both are basic usability gaps for any CLI tool.

## What Changes

- Add `mln --version` (and `-v`) flag that prints the current CLI version to stdout
- Update `mln init` to populate the `version` field in the generated `melon.yml` with the actual CLI version rather than a hardcoded default
- Inject the version at build time via `-ldflags` so the binary always carries the correct version from the git tag

## Capabilities

### New Capabilities

- `version-flag`: `mln --version` and `mln -v` print the CLI version (e.g. `mln v0.1.3`) and exit

### Modified Capabilities

- `multi-agent-init`: The `version` field written to `melon.yml` during `mln init` now uses the CLI's runtime version instead of a hardcoded `0.1.0`

## Impact

- `cmd/mln/main.go` — add version flag handling and a package-level `version` variable
- `.goreleaser.yaml` — add `-ldflags` to inject version at build time
- `cmd/mln/init_cmd.go` (or equivalent) — pass runtime version into manifest scaffolding
- No API or dependency changes; no breaking changes
