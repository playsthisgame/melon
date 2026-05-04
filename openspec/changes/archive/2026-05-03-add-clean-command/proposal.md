## Why

Over time, `melon install` accumulates stale entries in `.melon/` as dependencies are removed or upgraded, but nothing ever prunes them. A `clean` command gives users a safe, explicit way to reclaim disk space and remove orphaned symlinks without wiping the entire cache.

## What Changes

- Add a new `melon clean` (and `mln clean`) subcommand.
- The command reads `melon.lock` to determine the set of currently required dependencies.
- Any directory under `.melon/` not referenced by the lock file is deleted.
- For each removed cache entry, any corresponding symlinks in agent skill directories (`.claude/skills/`, `.windsurf/skills/`, etc.) are also removed.
- Update README with documentation for the new command.

## Capabilities

### New Capabilities

- `clean-command`: The `melon clean` subcommand — reads the lock file, identifies orphaned `.melon/` cache entries, removes them, and removes any orphaned symlinks in agent skill directories.

### Modified Capabilities

<!-- None -->

## Impact

- **New file**: `internal/cli/clean.go` — cobra command implementation
- **Modified**: `internal/cli/root.go` — register the new subcommand
- **Reads**: `melon.lock` (lockfile package), `.melon/` directory (store package), agent skill dirs (placer/agents packages)
- **Writes/deletes**: directories under `.melon/`, symlinks under agent skill dirs
- **README.md**: add `clean` to the commands reference
- No breaking changes; no new dependencies
