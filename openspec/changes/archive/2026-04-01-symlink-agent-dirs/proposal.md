## Why

Currently `mln install` copies skill files from `.melon/` into each agent's target directory (e.g. `.claude/skills/<skill-name>/`). Copying means skill files are duplicated on disk and changes to the cache are not reflected in agent directories without re-running install. Symlinks eliminate the duplication, keep the single source of truth in `.melon/`, and make the relationship between cache and placement explicit and inspectable.

## What Changes

- `internal/placer` switches from copying files to creating a single directory symlink per skill per target agent
- Each agent target directory gets a symlink: `<agent-dir>/skills/<skill-name>` → `<project-root>/.melon/<encoded-name>@<version>/`
- If a symlink (or directory) for the skill already exists at the target location it is removed and recreated (idempotent behaviour preserved)
- The `--no-place` flag behaviour is unchanged
- No changes to `.melon/` fetch layout or lock file format

## Capabilities

### New Capabilities

- `skill-placement`: How installed skills are placed into agent directories — symlinks from agent skill dirs into the `.melon/` cache

### Modified Capabilities

<!-- No existing spec-level requirements are changing — placement was previously undocumented behaviour -->

## Impact

- `internal/placer/placer.go`: replace `copyDir` calls with `os.Symlink`
- No changes to `internal/fetcher`, `internal/store`, `internal/agents`, manifest, or lockfile packages
- Agent directories must exist before symlinking; placer already creates them via `os.MkdirAll`
- Symlinks are relative paths from the skill slot location back to the `.melon/` entry, keeping the project portable if the whole directory is moved
