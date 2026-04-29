## Why

Skills are fetched sequentially, so installing N dependencies takes N × (clone time). Since each fetch is an independent `git clone` hitting a different remote, they have no shared state and are purely I/O-bound — parallel execution should yield near-linear speedup with no correctness risk.

## What Changes

- `fetchDeps` in `internal/cli/install_cmd.go` is rewritten to launch each fetch in its own goroutine, bounded by a concurrency semaphore.
- A fixed concurrency limit (e.g. 4 simultaneous fetches) is applied to avoid GitHub rate-limiting and excessive resource use.
- Progress callbacks (`onFetch`) continue to work correctly: the bubbletea `p.Send` path is already concurrency-safe; the plain-text path may interleave lines across deps (acceptable).
- The result slice is pre-allocated by index so no mutex is needed for result collection.
- All other pipeline stages (resolve, write lock, place, prune) remain sequential — they are already fast or have ordering requirements.

## Capabilities

### New Capabilities

- `parallel-dep-fetch`: Concurrent fetching of resolved dependencies during `melon install` (and transitively `melon add`, `melon update`, and `melon search` install flows), bounded by a configurable concurrency semaphore.

### Modified Capabilities

<!-- No existing spec-level behavior changes — the install pipeline's observable outputs (lock file, symlinks, error messages) are unchanged. Only the internal execution order of fetches changes. -->

## Impact

- **`internal/cli/install_cmd.go`** — `fetchDeps` function (~20 lines rewritten)
- **No API or lock file format changes**
- **No new dependencies** — uses standard library `sync` package only
- **All commands using `runInstall`** benefit automatically: `install`, `add`, `update`, `search`
