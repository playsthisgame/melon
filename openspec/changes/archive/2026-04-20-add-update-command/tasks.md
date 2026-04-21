## 1. Core Update Logic

- [x] 1.1 Create `internal/cli/update_cmd.go` with `runUpdate(cmd, args)` entry point
- [x] 1.2 Implement targeted mode: load `melon.yaml`, error if dep not found, resolve latest compatible version via `fetcher.LatestMatchingVersion`, skip if already up to date
- [x] 1.3 Implement branch-pin detection: check if constraint is a semver constraint or a branch string; print warning and exit for branch-pinned deps in targeted mode
- [x] 1.4 Implement "newer major version" hint: after resolving latest compatible, call `fetcher.LatestTag` and compare; print hint if a newer major exists outside the constraint
- [x] 1.5 After resolving updates, call `runInstall` to fetch, write lock, place symlinks, and prune

## 2. Interactive Multi-Select TUI

- [x] 2.1 Create `runUpdateTUI` function (modelled on `runRemoveTUI`) that renders a multi-select list with "Update all" as the first sentinel item
- [x] 2.2 Handle "Update all" selection: if sentinel is selected, treat all deps as selected regardless of individual state
- [x] 2.3 Filter out branch-pinned deps from the interactive list before rendering
- [x] 2.4 Wire interactive path: if no args and TTY, load deps, call `runUpdateTUI`, pass selected names to update logic
- [x] 2.5 Handle non-TTY no-args case: print error asking for explicit dep name and exit

## 3. Spinner Integration

- [x] 3.1 Wrap version resolution calls in `withSpinner("Resolving updates…", ...)` when in TTY
- [x] 3.2 Verify spinner clears before result output (up-to-date message, hint, or install output)

## 4. Up-to-Date Messaging

- [x] 4.1 After resolving all selected deps, if none have updates, print "All selected skills are up to date." and return without running install
- [x] 4.2 When some deps are up to date and others have updates, print a per-dep skip note for the up-to-date ones before running install on the rest

## 5. CLI Registration

- [x] 5.1 Register `update` subcommand in `internal/cli/cli.go` with `Use: "update [dep]"`, `Short` description, and `Args: cobra.MaximumNArgs(1)`
- [x] 5.2 Add `update` to the command list in README usage section

## 6. Tests

- [x] 6.1 Write table-driven unit tests for version-resolution logic (up to date, has update, branch-pinned, dep not found)
- [x] 6.2 Write test for "newer major version" hint detection
- [x] 6.3 Run `go test ./...` and confirm all tests pass
