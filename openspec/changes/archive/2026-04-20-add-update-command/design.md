## Context

Melon has `add` (which resolves latest tag and writes a `^version` constraint) and `remove`, but no `update`. Users who want to pull in a newer version of an installed skill must either re-run `melon add <dep>` (which overwrites the constraint) or edit `melon.yaml` manually.

The update command must respect the *existing* semver constraint in `melon.yaml` — it should not widen it. If a user has `^1.2.0` and a `1.5.0` release exists, updating resolves `1.5.0`. If `2.0.0` exists but is outside the constraint, it is ignored.

The interactive path closely mirrors `remove-interactive` (multi-select TUI) and the spinner path mirrors `add` (spinner while resolving). No new packages are needed.

## Goals / Non-Goals

**Goals:**
- `melon update` (no args, TTY): multi-select list of all declared deps; "Update all" as first option; resolves latest compatible version for each selected dep; runs install pipeline
- `melon update <dep>`: resolves latest compatible version for one dep; errors if dep not in `melon.yaml`; runs install pipeline
- If selected dep(s) are already at the latest compatible version, print a clear message and exit without touching `melon.yaml` or `melon.lock`
- Non-TTY `melon update` with no args: error asking for an explicit dep name (same pattern as `remove`)

**Non-Goals:**
- Widening semver constraints (e.g. `^1.x` → `^2.x`) — that's a job for `melon add`
- Updating branch-pinned deps (e.g. `main`) — branch pins always fetch HEAD; skip or no-op these
- Dry-run / `--check` flag — out of scope for this change

## Decisions

**Reuse existing install pipeline**
After determining which deps have updates and writing new pinned versions to `melon.lock`, call `runInstall` exactly as `add` and `remove` do. This gives fetch + symlink + prune for free and keeps the update path consistent with the rest of the CLI.

**Do not modify `melon.yaml` constraints**
`melon update` only changes which pinned version is recorded in `melon.lock`. The constraint in `melon.yaml` stays the same (e.g. `^1.2.0` remains `^1.2.0`). This matches standard package manager semantics (npm `update` vs `install <pkg>@latest`).

**Version resolution via `fetcher.LatestMatchingVersion`**
The resolver already calls `fetcher.LatestMatchingVersion(repoURL, constraint)`. The update command calls this directly for each selected dep (outside the resolver) to check whether a newer version is available before deciding to run install.

**"Update all" as first item in multi-select**
Implemented as a sentinel item at index 0. If selected, all other items are treated as selected regardless of individual checkbox state. This is simpler than a "select all" toggle and is immediately visible.

**Skip branch-pinned deps silently**
Deps with a non-semver constraint (e.g. `main`, `feature-branch`) are excluded from the selectable list in interactive mode and produce a `"skipping <dep>: branch pin"` warning in targeted mode. Branch pins are effectively always "up to date" from melon's perspective.

## Risks / Trade-offs

**Rate limits during bulk update** → Each dep requires a GitHub API call to resolve the latest tag. For projects with many deps, this can hit unauthenticated rate limits. Mitigation: existing `GITHUB_TOKEN` env var support applies automatically; same risk exists with `melon install`.

**No constraint widening** → A user on `^1.x` who wants `2.0.0` must use `melon add <dep>@^2.0.0`. This is the right behavior but may confuse users expecting a full "upgrade". Mitigation: print a hint when the latest version is outside the constraint.
