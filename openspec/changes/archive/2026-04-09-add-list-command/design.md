## Context

Melon tracks installed skills in two places: `melon.yaml` (desired state) and `melon.lock` (resolved/installed state). Skills are then placed as directory symlinks into one or more agent tool directories derived from `m.ToolCompat` or `m.Outputs`. Currently there is no command to inspect these states, so users can't tell if their lock file is stale or if symlinks are missing without manually inspecting the filesystem.

## Goals / Non-Goals

**Goals:**
- Add `melon list` to display skills recorded in `melon.lock` (installed state)
- Add `--pending` flag to show skills in `melon.yaml` with no corresponding lock entry
- Add `--check` flag to verify symlinks exist in all expected tool directories and report missing ones

**Non-Goals:**
- Re-running install or fixing missing symlinks automatically
- Displaying transitive dependencies (only direct/declared skills)
- Network access or version resolution

## Decisions

### Source of truth for "installed" = melon.lock

`melon.lock` is the canonical record of what has been fetched and placed. Reading it is cheap (local file, already-parsed YAML) and consistent with how `runInstall` and `runRemove` operate. Alternative: walk `.melon/` store directories — rejected because store entries can exist for pruned deps and aren't guaranteed to reflect placement state.

### Source of truth for "pending" = melon.yaml minus melon.lock

A skill is pending if it appears in `m.Dependencies` (manifest) but has no matching name in `lock.Dependencies`. This mirrors the semantic used by `resolver.Resolve`: the manifest declares intent, the lock records fulfillment.

### --check inspects symlinks in target directories

The same target-directory logic used by `placer.Place`/`placer.Unplace` is reused: `agents.DeriveTargets(m.ToolCompat)` or `m.Outputs`. For each locked dep × each target dir, `os.Lstat` is used to test whether the symlink path exists. A missing or broken symlink is reported as an error row. `--check` implies listing installed skills and annotating each with a status column.

### New file: internal/cli/list_cmd.go

Follows the pattern of `remove_cmd.go`: loads manifest and lockfile, applies logic, prints to `cmd.OutOrStdout()`. No new packages needed.

## README Update

The `### melon list` section goes between `melon remove` and `melon search` in `README.md` (alphabetical order is broken there anyway — insert after `melon remove` to match the logical workflow: add → install → list → remove). It should show:

- Default invocation and output format
- `--pending` flag with example
- `--check` flag with example showing OK/MISSING statuses

## Risks / Trade-offs

- **Lock file absent** → Report "no skills installed" gracefully rather than error, since a fresh project may not have run `melon install` yet.
- **`--pending` with no manifest** → Return an error (same as all other commands that require `melon.yaml`).
- **Combining flags** → `--pending` and `--check` are independent sections; both can be shown together. Combining them is additive, not conflicting.
- **Broken symlinks** → `os.Lstat` detects the link exists even if target is missing; to detect a broken link, use `os.Stat` (follows symlink). `--check` will use `os.Stat` so broken symlinks are flagged.
