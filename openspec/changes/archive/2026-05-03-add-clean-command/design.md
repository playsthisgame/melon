## Context

The install pipeline (Resolve → Fetch → Write lock → Place → Prune) prunes symlinks for removed deps, but it never deletes stale entries from `.melon/`. After removing or upgrading dependencies, old cache directories accumulate silently. The `clean` command fills this gap.

Existing primitives that `clean` can compose:
- `store.List(projectDir)` — returns all `(name, version)` pairs present in `.melon/`
- `store.InstalledPath(projectDir, dep)` — canonical directory name for a dep
- `store.Remove(projectDir, dep)` — deletes a dep's cache directory
- `lockfile.Load(path)` — parses `melon.lock` into a `LockFile` with `Dependencies []LockedDep`
- `placer.Unplace(deps, manifest, projectDir, out)` — removes symlinks for a slice of deps

## Goals / Non-Goals

**Goals:**
- Delete any `.melon/<entry>` directory whose `(name, version)` does not appear in `melon.lock`
- For each removed cache entry, also remove its corresponding symlinks from agent skill directories
- Print a clear per-entry summary of what was removed (or a "nothing to clean" message)
- Be a no-op (and say so) if `melon.lock` does not exist

**Non-Goals:**
- Wiping the entire `.melon/` cache (`--all` flag is out of scope for now)
- Removing placed symlinks for deps that *are* in the lock (i.e., not a full uninstall)
- Modifying `melon.yaml` or `melon.lock`

## Decisions

**1. Lock-file as source of truth, not manifest**

Use `melon.lock` (not `melon.yaml`) to determine what is "in use". The lock file records the exact installed version; the manifest records constraints. Using the lock avoids false-positive removals when a manifest constraint matches a cached version that differs from the pinned one.

Alternatives considered: diffing `melon.yaml` deps — rejected because it requires running resolution, which needs network access.

**2. Reuse `placer.Unplace` for symlink removal**

`placer.Unplace` already handles iterating agent directories and silently ignoring missing links. Rather than reimplementing symlink removal, construct a `[]lockfile.LockedDep` slice from the orphaned entries and call `Unplace`.

The manifest is needed to know *which* agent dirs to look in. Load it the same way other commands do (via `manifest.FindPath`). If the manifest is absent, skip symlink removal (the project may not have one yet).

**3. Match cache dirs to lock entries by directory name**

`store.dirName` encodes `name@version` into a filesystem-safe string. The inverse — reading back from the dir name — is already implemented in `store.List`. Cross-reference `store.List` output against the lock's `Dependencies` slice to find orphans.

**4. Output format**

Follow the style of existing commands: one line per action, indented with two spaces, using `removeStyle` for removed entries. Print a count summary at the end (`N entries cleaned.` / `Nothing to clean.`).

## Risks / Trade-offs

- **`store.List` name reconstruction is lossy** — slashes become dashes, so reconstructed names may be ambiguous for paths with dashes in segments. Mitigation: match on the raw directory name string (before reconstruction) against lock entries using the same `dirName` function, rather than comparing reconstructed names.

- **Manifest not required** — if `melon.yaml` is absent, symlinks cannot be removed (we don't know which agent dirs to check). The command will still clean `.melon/` and warn the user that symlinks were skipped.

- **Orphaned symlinks without orphaned cache** — if a symlink points to a dep that is in the lock but the cache dir was manually deleted, `clean` will not remove that dangling symlink (it only acts on orphaned cache entries). This is a known limitation; `install` is the right fix for that case.
