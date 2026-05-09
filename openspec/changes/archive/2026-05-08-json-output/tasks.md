## 1. melon list --json

- [x] 1.1 Add `--json` flag to the `list` command in `internal/cli/list_cmd.go`
- [x] 1.2 When `--json` is set, skip all TUI/lipgloss output and read lock file entries
- [x] 1.3 Emit `{"installed": [...]}` where each element maps lock file fields (`name`, `version`, `git_tag`, `repo_url`, `subdir`, `entrypoint`, `tree_hash`)
- [x] 1.4 When lock file is absent or empty, emit `{"installed": []}`
- [x] 1.5 When `--json` + `--pending`, include `"pending": [...]` key with dep name strings
- [x] 1.6 When `--json` + `--check`, include `"check": [...]` key with `{name, path, status}` objects
- [x] 1.7 On error with `--json` set, write `{"error": "..."}` to stderr and exit non-zero
- [x] 1.8 Write tests for JSON output: installed, empty, pending, check, error cases

## 2. melon info --json

- [x] 2.1 Add `--json` flag to the `info` command in `internal/cli/info_cmd.go`
- [x] 2.2 When `--json` is set, suppress all formatted/styled output
- [x] 2.3 Emit a JSON object with `name`, `description`, `author`, `latest_version`, `versions`, `branches`
- [x] 2.4 Populate `author` and `description` from index if available, fall back to GitHub repo about field
- [x] 2.5 When no semver tags exist, emit empty `versions` and populate `branches`
- [x] 2.6 On error (path not found, API failure) with `--json`, write `{"error": "..."}` to stderr and exit non-zero
- [x] 2.7 Write tests for JSON output: in-index skill, not-in-index skill, no-tags skill, error case
