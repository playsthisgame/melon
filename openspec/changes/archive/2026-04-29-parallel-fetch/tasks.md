## 1. Rewrite `fetchDeps` for concurrent execution

- [x] 1.1 Add `"sync"` to the imports in `internal/cli/install_cmd.go`
- [x] 1.2 Replace the sequential loop in `fetchDeps` with goroutines: pre-allocate `locked` and `errs` slices by index, launch one goroutine per dep, use a buffered channel semaphore (`const maxConcurrent = 4`) to bound concurrency, and `sync.WaitGroup` to wait for all goroutines
- [x] 1.3 After `wg.Wait()`, iterate `errs` in index order and return the first non-nil error (preserving fail-fast behaviour)

## 2. Verify correctness of progress reporting

- [x] 2.1 Confirm TTY path: `onFetch` calls `p.Send(depFetchedMsg{...})` from goroutines — verify bubbletea Send is concurrency-safe (no change needed, just validate)
- [x] 2.2 Confirm plain-text path: `fmt.Fprintf` calls from goroutines are acceptable (lines will not tear; ordering is non-deterministic but that is acceptable for progress output)

## 3. Test

- [x] 3.1 Run `go test ./internal/cli/...` and confirm all existing install/add/update tests pass
- [x] 3.2 Run `go test ./...` to confirm no regressions across the full module
- [x] 3.3 Manual smoke test: add 3+ deps to a test `melon.yaml` and run `melon install` — confirm all deps are fetched, lock file is written in deterministic order, and symlinks are placed correctly
