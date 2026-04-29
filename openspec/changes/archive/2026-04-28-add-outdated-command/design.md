## Context

`melon update` already does a pre-flight version check before running the install pipeline, but it is a mutating command and requires user interaction. There is no read-only equivalent. Users in CI or those who simply want to audit their dependencies before deciding to upgrade have no clean tool for this.

The `outdated` command is deliberately a pure read operation: it never writes to `melon.yaml`, `melon.lock`, or any agent directory.

## Goals / Non-Goals

**Goals:**
- Read `melon.yaml` and `melon.lock`; compare locked versions against latest compatible and absolute latest
- Print a formatted table: dep name | locked | latest compatible | absolute latest (if outside constraint)
- Exit code 1 if any dep is outdated (enables `melon outdated && echo "up to date"` in CI)
- Exit code 0 if everything is current or only branch-pinned deps exist
- Skip branch-pinned deps with a one-line note; they are always "current" from melon's perspective
- Show spinner while resolving (TTY); plain output in non-TTY

**Non-Goals:**
- Modifying any file — this is strictly read-only
- Interactive selection (no TUI list beyond spinner)
- Filtering by dep name as an argument (check all or nothing)

## Decisions

**Read from melon.lock for "current", not melon.yaml**
The locked version is the ground truth for what is actually installed. A dep in `melon.yaml` that hasn't been installed yet has no locked version — these are shown as `(not installed)` in the current column.

**Two version columns: "latest compatible" and "absolute latest"**
- *Latest compatible*: highest version satisfying the existing constraint (`^1.x`, `~1.x`, exact). If this differs from locked, the dep is outdated within its constraint.
- *Absolute latest*: highest tag regardless of constraint. Shown only when it exceeds latest compatible — surfaces available major upgrades without implying they should be auto-applied.
- This mirrors `npm outdated`'s `wanted` vs `latest` columns.

**Exit code 1 when outdated**
Standard convention (npm, cargo outdated, etc.). Allows CI pipelines to fail on stale deps without needing to parse output.

**Reuse `isBranchPin` and `fetcher` directly**
`isBranchPin` is already defined in `update_cmd.go` (same package). `fetcher.LatestMatchingVersion` and `fetcher.LatestTag` are the same calls used by `update`. No new packages or interfaces needed.

**Concurrent resolution**
Each dep requires two network calls (`LatestMatchingVersion` + `LatestTag`). Resolve all deps concurrently (goroutines + errgroup or manual channel) to keep latency proportional to the slowest dep, not the sum. This matters more for `outdated` than for `update` since there is no install pipeline to follow.

## Risks / Trade-offs

**Rate limits on large dependency sets** → Same as `melon update`. `GITHUB_TOKEN` env var mitigates this; print a note if resolution fails due to rate limiting.

**No lock file present** → If `melon.lock` doesn't exist yet (project never installed), treat all locked versions as `(not installed)` and still resolve latest. The command remains useful as a "what would install" preview.
