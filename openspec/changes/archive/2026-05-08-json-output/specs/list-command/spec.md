## ADDED Requirements

### Requirement: melon list --json emits installed deps as a JSON array

When `--json` is set, `melon list` SHALL write a JSON object to stdout with an `installed` key containing an array of objects, one per lock file entry. Each object SHALL include: `name`, `version`, `git_tag`, `repo_url`, `subdir`, `entrypoint`, and `tree_hash`.

#### Scenario: JSON output with installed skills

- **WHEN** `melon list --json` is run and `melon.lock` has one or more entries
- **THEN** stdout is a JSON object `{"installed": [...]}` where each element contains the lock file fields for that dep

#### Scenario: JSON output with no skills installed

- **WHEN** `melon list --json` is run and `melon.lock` is absent or empty
- **THEN** stdout is `{"installed": []}`

### Requirement: melon list --json --pending includes pending deps

When `--json` and `--pending` are both set, the output object SHALL include a `pending` key with an array of dep name strings for skills declared in `melon.yaml` but absent from `melon.lock`.

#### Scenario: JSON output with pending skills

- **WHEN** `melon list --json --pending` is run and one or more deps are pending
- **THEN** stdout is `{"installed": [...], "pending": ["github.com/owner/repo/..."]}` 

### Requirement: melon list --json --check includes placement status

When `--json` and `--check` are both set, the output object SHALL include a `check` key with an array of objects, each containing `name`, `path`, and `status` (`"ok"` or `"missing"`).

#### Scenario: JSON output with check results

- **WHEN** `melon list --json --check` is run
- **THEN** stdout is `{"installed": [...], "check": [{"name": "...", "path": "...", "status": "ok|missing"}]}`
