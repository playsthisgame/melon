# Tasks

## 1. Core Command

- [x] 1.1 Create `internal/cli/list_cmd.go` with `runList` function and `listCmd` cobra command
- [x] 1.2 Register `listCmd` in `cli.go` `Run()` function alongside existing commands
- [x] 1.3 Add `--pending` and `--check` boolean flags to `listCmd` in an `init()` block

## 2. Installed Skills Output

- [x] 2.1 Load `melon.lock` (using `lockfile.Load`); if absent or empty, print "No skills installed." and return
- [x] 2.2 Sort locked deps alphabetically by name and print each as `<name>  <version>`

## 3. Pending Skills (`--pending`)

- [x] 3.1 Load `melon.yaml` (using `manifest.Load`); return error if not found
- [x] 3.2 Build a set of names from `melon.lock` and compare against `melon.yaml` dependencies
- [x] 3.3 Print pending skills under "Pending (not installed):" header, or "No pending skills." if none

## 4. Placement Check (`--check`)

- [x] 4.1 Derive target directories using `agents.DeriveTargets(m.ToolCompat)` or `m.Outputs` (same logic as placer)
- [x] 4.2 For each locked dep × each target dir, use `os.Stat` to verify the symlink path resolves
- [x] 4.3 Print each skill with "OK" or "MISSING `<path>`" status; exit with code 1 if any are missing

## 5. Documentation

- [x] 5.1 Add a `### melon list` section to `README.md` under Commands, showing default usage, `--pending`, and `--check` flags with example output

## 6. Tests

- [x] 6.1 Add `internal/cli/list_cmd_test.go` with table-driven tests covering: installed skills listed, no lock file, pending skills shown, no pending skills, missing symlink detected
