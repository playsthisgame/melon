## Context

`internal/placer` currently copies the entire skill directory tree from `.melon/<encoded>@<version>/` into each agent target (e.g. `.claude/skills/<skill-name>/`). This duplicates files on disk and means the agent directory content can drift from the cache if something is changed manually.

The change replaces the copy with a single directory symlink per skill slot. The symlink points from the agent target location into the `.melon/` cache entry, keeping one authoritative copy of each skill.

## Goals / Non-Goals

**Goals:**
- Replace file copying with a directory symlink in `Place`
- Symlink is relative (portable if the whole project dir is moved)
- Idempotent: existing symlink or directory at the target is removed and recreated
- All existing behaviour (multi-agent targets, `--no-place`, outputs override) is preserved
- Printed output message updated to reflect symlinking rather than copying

**Non-Goals:**
- Changes to `.melon/` fetch layout or `store.InstalledPath` encoding
- Changes to lock file, manifest, or any package outside `internal/placer`
- Handling Windows junction points or non-symlink platforms (out of scope for now)
- Lazy or deferred symlink creation

## Decisions

### Relative vs absolute symlinks

**Decision**: Use relative symlinks.

The symlink target is computed as the relative path from the symlink's parent directory to the `.melon/` cache entry:

```
linkDir  = filepath.Join(projectDir, base)          // e.g. .claude/skills/
linkPath = filepath.Join(linkDir, skillName)        // e.g. .claude/skills/pdf-skill
target   = filepath.Rel(linkDir, store.InstalledPath(projectDir, dep))
           // e.g. ../../.melon/github.com%2Falice%2Fpdf-skill@1.2.0
```

**Why relative**: Moving or renaming the project directory keeps symlinks valid without needing to re-run `mln install`.

**Alternative**: Absolute symlinks — simpler to compute but break on directory moves/renames.

### Idempotency strategy

**Decision**: Remove then recreate.

Before calling `os.Symlink`, check if a file/dir/symlink already exists at `linkPath`. If so, call `os.Remove` (works for files, symlinks, and empty dirs) then recreate. If the existing entry is a non-empty directory (old copy-based placement), use `os.RemoveAll`.

**Why**: Simplest way to handle all cases — stale symlink, stale copy, version upgrade.

### Removal of `copyDir` / `copyFile`

Both helpers become dead code once `Place` switches to `os.Symlink`. They will be deleted.

## Risks / Trade-offs

- **Symlinks not supported on all filesystems** → Mitigation: surface the `os.Symlink` error directly; user can use `--no-place` as a workaround. Document limitation.
- **Agent tool follows symlinks** → This is the desired behaviour — the agent reads the skill content from `.melon/` transparently.
- **`mln install` deletes `.melon/` then reinstalls** → Symlinks temporarily dangle between remove and re-fetch; since install is a single process this is not observable in normal use.
- **`git clean -fdx` removes `.melon/`** → Symlinks dangle until `mln install` is re-run. Same behaviour as today with copied files (they'd be gone too).

## Migration Plan

1. Replace `Place` implementation in `internal/placer/placer.go`
2. Delete `copyDir` and `copyFile` helpers
3. Update package doc comment
4. Update/add tests in `internal/placer/placer_test.go`

No lockfile or manifest format changes; no migration of existing `.melon/` caches required. Users with existing copy-based agent directories will have them replaced by symlinks on the next `mln install`.
