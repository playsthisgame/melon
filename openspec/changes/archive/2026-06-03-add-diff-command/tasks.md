## 1. Argument Parsing & Version Selection

- [x] 1.1 Create `internal/cli/diff_cmd.go` with `runDiff(cmd, args)` entry point and a `<dep>[@<target>]` positional argument
- [x] 1.2 Split the argument into dep name and optional `@<target>` (reuse the same parsing convention as `melon add`)
- [x] 1.3 Load `melon.yaml`; error if the named dep is not declared in `dependencies`
- [x] 1.4 Load `melon.lock`; error with a "run melon install first" hint if the dep has no locked entry (this is the "from" version)
- [x] 1.5 Determine the "to" version: use the explicit `@<target>` if given; otherwise resolve `fetcher.LatestMatchingVersion` from the dep's constraint
- [x] 1.6 For branch-pinned constraints with no explicit target, error with guidance to supply a `@<target>`
- [x] 1.7 Error if the requested target version cannot be resolved to a tag/branch

## 2. Materializing Both Trees

- [x] 2.1 Resolve the "from" tree to its cached `store.InstalledPath` (locked version is already in `.melon/`)
- [x] 2.2 Build a `resolver.ResolvedDep` for the target version and call `fetcher.Fetch` into its `store.InstalledPath` (idempotent; warms cache)
- [x] 2.3 Fast path: if the locked tree hash equals the target's fetch-result tree hash, print `No changes` and exit 0

## 3. Diff Computation & Rendering

- [x] 3.1 Union the sorted relative file paths of both trees; classify each as added / removed / changed / unchanged by content comparison
- [x] 3.2 Add a unified-diff helper (small pure-Go renderer or vetted library, `CGO_ENABLED=0`-safe) to format hunks for changed text files
- [x] 3.3 Detect non-UTF8/NUL-containing files and report "binary file changed" instead of hunks
- [x] 3.4 Print added-file and removed-file headers for whole-file additions/removals
- [x] 3.5 Implement `--stat`: per-file `+added/-removed` counts and a totals line, no hunks
- [x] 3.6 Apply ANSI coloring for added/removed lines only when stdout is a TTY (via `cli/tty.go`) and `--no-color` is not set

## 4. CLI Registration

- [x] 4.1 Register `diff` subcommand in `internal/cli/cli.go` with `Use: "diff <dep>[@<target>]"`, a `Short` description, `Args: cobra.ExactArgs(1)`, and the `--stat` / `--no-color` flags
- [x] 4.2 Exit 0 whether or not differences exist; non-zero only on errors
- [x] 4.3 Add a `melon diff` section to the README command list

## 5. Tests

- [x] 5.1 Test: locked vs. latest-compatible with a changed `SKILL.md` → unified hunk printed, exit 0
- [x] 5.2 Test: target adds and removes files → added/removed headers printed
- [x] 5.3 Test: identical tree hashes → prints `No changes`, no hunks, exit 0
- [x] 5.4 Test: explicit `@<version>` target overrides the constraint
- [x] 5.5 Test: explicit `@<branch>` target diffs against branch tree
- [x] 5.6 Test: dep not in lock → error suggesting `melon install`, non-zero exit
- [x] 5.7 Test: dep not in manifest → "not declared" error, non-zero exit
- [x] 5.8 Test: unresolvable target version → error, non-zero exit
- [x] 5.9 Test: branch-pinned dep without target → error requiring a target
- [x] 5.10 Test: `--stat` prints summary only, no hunks
- [x] 5.11 Test: `--no-color` / non-TTY → no ANSI escape codes in output
- [x] 5.12 Test: binary file change → reported as binary, no hunks
- [x] 5.13 Test: `melon.yaml` and `melon.lock` unchanged after running diff
- [x] 5.14 Run `go test ./...` and confirm all tests pass
