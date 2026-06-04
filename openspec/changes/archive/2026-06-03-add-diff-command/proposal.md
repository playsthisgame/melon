## Why

`melon outdated` tells you _that_ a dependency has a newer version, but not _what changed_. Before running `melon update`, users have no way to inspect the actual content differences in a skill — the markdown, frontmatter, or instructions an AI agent will start consuming. The README's pitch ("you'll see the diff in the lock file and upgrade intentionally") only surfaces a hash change, not the human-readable content. `melon diff <dep>` closes this gap by showing the file-level changes between the locked version and a target version, so upgrades are an informed decision rather than a leap of faith.

## What Changes

- Adds a new `melon diff <dep>` command that:
  - Reads `melon.yaml` and `melon.lock` to find the dep's currently locked version (the "from" side)
  - Resolves the "to" version: the latest version satisfying the dep's constraint by default, or an explicit `@<version>`/`@<branch>` target if given (e.g. `melon diff <dep>@2.0.0`)
  - Fetches the target version into the `.melon/` cache if not already present (the locked version is already cached)
  - Computes and prints a unified, file-by-file diff between the two version trees: added files, removed files, and changed file contents
  - Reports "No changes" when the locked and target trees are identical (matching tree hashes)
- Supports `--stat` to print only a per-file summary (files changed, added/removed line counts) without full hunks
- Supports `--no-color` to suppress ANSI coloring (auto-disabled when stdout is not a TTY)
- Skips network resolution when the target equals the locked version; skips branch-pinned deps' "latest" resolution by requiring an explicit target for branch constraints

## Capabilities

### New Capabilities

- `diff-command`: The `melon diff <dep>` command — resolves a from/to version pair, fetches the target tree, and prints a file-level unified diff (or `--stat` summary) of the skill's contents.

### Modified Capabilities

_(none)_

## Impact

- New file: `internal/cli/diff_cmd.go`
- Reads from existing `manifest`, `lockfile`, `fetcher`, `resolver`, and `store` packages — version resolution (`fetcher.LatestMatchingVersion`), fetch (`fetcher.Fetch`), and cache paths (`store.InstalledPath`) already exist
- New small diffing helper (file tree comparison + unified diff rendering); may add one dependency for unified-diff formatting or implement a minimal renderer in-tree
- `internal/cli/cli.go` — register new `diff` subcommand
- Read-only: no writes to `melon.yaml` or `melon.lock`; the only side effect is populating `.melon/` cache with the target version (same as any fetch)
