## Context

`fetchDeps` in `internal/cli/install_cmd.go` iterates over resolved dependencies sequentially, calling `fetcher.Fetch` (which does a `git clone` + sparse checkout) for each one in turn. With N deps, total install time is roughly N × single-fetch-time — dominated by network round-trips to GitHub.

Each `fetcher.Fetch` call is fully self-contained: it clones into a unique temp dir, copies files to a unique `installDir` derived from dep name+version (`store.InstalledPath`), and returns a `FetchResult`. There is no shared mutable state between fetches.

## Goals / Non-Goals

**Goals:**
- Fetch all resolved deps concurrently to reduce total install time
- Preserve the ordering of `[]lockfile.LockedDep` (lock file must be deterministic)
- Keep progress reporting working correctly in both TTY and plain-text modes
- Use only the standard library — no new dependencies

**Non-Goals:**
- Parallelising the resolver (it does sequential GitHub API calls with interdependencies)
- Parallelising placement (`placer.Place`) — it's fast and has symlink ordering concerns
- Making concurrency limit configurable via a flag (YAGNI; can be added later)
- Capping global goroutine count beyond the fetch semaphore

## Decisions

### 1. Goroutines + index-keyed pre-allocated slice (no mutex for results)

Pre-allocate `locked := make([]lockfile.LockedDep, len(resolved))` and `errs := make([]error, len(resolved))`. Each goroutine writes to its own index `i`. In Go, writes to different slice elements are not a data race, so no mutex is needed for result collection.

**Alternatives considered:**
- `chan lockfile.LockedDep` with a collector goroutine: requires sorting results by original index afterward to maintain lock-file order — more complex for no benefit.
- `sync.Mutex` around an append: simpler but still requires sorting.

### 2. Semaphore via buffered channel, limit = 4

```go
const maxConcurrent = 4
sem := make(chan struct{}, maxConcurrent)
// goroutine: sem <- struct{}{}; defer func() { <-sem }()
```

Limits simultaneous `git clone` processes. Without a limit, cloning 20 deps at once would saturate the network, exhaust file descriptors, and risk GitHub secondary rate-limits (which kick in on many concurrent unauthenticated requests to the same host).

4 is chosen as a conservative default that still provides substantial speedup for typical use (2–10 deps) while being well under GitHub's limits.

**Alternatives considered:**
- `runtime.GOMAXPROCS(0)`: CPU-oriented, wrong signal for I/O-bound work.
- `golang.org/x/sync/errgroup`: Clean API but adds an external dependency for ~10 lines of standard-library code.
- No limit: Risky on large dep trees.

### 3. Error handling: collect all, return first non-nil

Each goroutine sets `errs[i]`. After `wg.Wait()`, iterate `errs` in index order and return the first non-nil error. This matches the current sequential behavior (fail-fast on first error) while being deterministic.

**Alternative:** Return all errors joined — more informative but changes the existing error surface. Not worth it here.

### 4. `onFetch` callback called from goroutines — no additional synchronization

- **TTY path**: `p.Send(depFetchedMsg{...})` — bubbletea's Send is documented concurrency-safe. ✓
- **Plain-text path**: `fmt.Fprintf(cmd.OutOrStdout(), ...)` — individual `Fprintf` calls are effectively atomic at the OS write-syscall level for short strings to a terminal; lines will not be torn, though ordering across deps is non-deterministic. This is acceptable.

## Risks / Trade-offs

- **GitHub rate limiting** → Semaphore cap of 4 keeps concurrent git requests conservative. `GITHUB_TOKEN` is already supported for higher limits.
- **Temp dir exhaustion** → Each `fetcher.Fetch` creates one temp dir, cleans it up with `defer os.RemoveAll`. With limit=4, at most 4 temp dirs exist simultaneously. Not a concern.
- **Non-deterministic plain-text progress output** → Lines for concurrent fetches may interleave. The final lock diff summary remains deterministic. Acceptable trade-off.
- **Fetch failure leaves partial store entries** → Same as before; `fetcher.Fetch` removes stale installDir before re-fetching. Parallel failures don't interfere with each other's directories.

## Migration Plan

No data migration needed. The change is internal to `fetchDeps`; all external interfaces (CLI flags, lock file schema, symlink placement) are unchanged. Existing lock files remain valid.

Rollback: revert `fetchDeps` to the sequential loop — one-commit revert.

## Open Questions

None. The approach is straightforward and all decisions are resolved above.
