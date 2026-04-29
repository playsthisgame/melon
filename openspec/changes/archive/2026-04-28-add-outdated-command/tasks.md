## 1. Core Resolution Logic

- [x] 1.1 Create `internal/cli/outdated_cmd.go` with `runOutdated(cmd, args)` entry point
- [x] 1.2 Load `melon.yaml`; return early with "No dependencies declared" message if empty
- [x] 1.3 Load `melon.lock` (ignore error if absent — treat all deps as not installed)
- [x] 1.4 Separate deps into semver-constrained and branch-pinned; print a skip note for branch-pinned deps
- [x] 1.5 For each semver-constrained dep, concurrently resolve: (a) latest compatible version via `fetcher.LatestMatchingVersion`, (b) absolute latest via `fetcher.LatestTag`
- [x] 1.6 Collect results; for each dep where locked version != latest compatible, record as outdated (use `(not installed)` when no locked version exists)

## 2. Output Formatting

- [x] 2.1 Print a formatted table with columns: dep name, current (locked), latest compatible, absolute latest (shown only when outside constraint, with a `*` or `↑` indicator)
- [x] 2.2 When no rows are outdated, print "All skills are up to date." and exit 0
- [x] 2.3 When rows exist, print the table then exit with code 1
- [x] 2.4 Print branch-pinned skip note before the table (or up-to-date message)

## 3. Spinner Integration

- [x] 3.1 Wrap concurrent resolution in `withSpinner("Checking for updates…", ...)` when stdout is a TTY
- [x] 3.2 Verify spinner clears before table or up-to-date message is printed

## 4. CLI Registration

- [x] 4.1 Register `outdated` subcommand in `internal/cli/cli.go` with `Use: "outdated"`, a `Short` description, and `Args: cobra.NoArgs`
- [x] 4.2 Add `melon outdated` to the README command list

## 5. Tests

- [x] 5.1 Test: all deps up to date → prints up-to-date message, no error
- [x] 5.2 Test: dep with newer compatible version → appears in outdated results
- [x] 5.3 Test: dep not in lock file → shown as `(not installed)`, appears as outdated
- [x] 5.4 Test: missing lock file → all deps shown as not installed
- [x] 5.5 Test: branch-pinned dep → excluded from results, skip note printed
- [x] 5.6 Test: empty manifest → prints "No dependencies declared" message
- [x] 5.7 Run `go test ./...` and confirm all tests pass
