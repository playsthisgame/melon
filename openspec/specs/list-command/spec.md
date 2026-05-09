## Requirements

### Requirement: List installed skills
The system SHALL provide a `melon list` command that reads `melon.lock` and prints the name and version of every installed skill, one per line.

#### Scenario: Skills are installed
- **WHEN** `melon list` is run and `melon.lock` exists with one or more dependencies
- **THEN** each dependency is printed as `<name>  <version>` sorted alphabetically by name

#### Scenario: No skills installed
- **WHEN** `melon list` is run and `melon.lock` is absent or has no dependencies
- **THEN** the command prints "No skills installed." and exits with code 0

### Requirement: Show pending skills
The system SHALL support a `--pending` flag that prints skills declared in `melon.yaml` but absent from `melon.lock`.

#### Scenario: Pending skills exist
- **WHEN** `melon list --pending` is run and one or more `melon.yaml` dependencies have no matching lock entry
- **THEN** each pending skill name is printed under a "Pending (not installed):" header

#### Scenario: No pending skills
- **WHEN** `melon list --pending` is run and every `melon.yaml` dependency appears in `melon.lock`
- **THEN** the command prints "No pending skills." under the pending section

#### Scenario: melon.yaml absent
- **WHEN** `melon list --pending` is run and no `melon.yaml` can be found
- **THEN** the command returns a non-zero exit code with an error message

### Requirement: Check placement of installed skills
The system SHALL support a `--check` flag that verifies each installed skill's symlink exists (and is not broken) in every expected tool directory.

#### Scenario: All symlinks present and valid
- **WHEN** `melon list --check` is run and every expected symlink resolves successfully
- **THEN** each skill is printed with an "OK" status and the command exits with code 0

#### Scenario: Missing or broken symlink detected
- **WHEN** `melon list --check` is run and one or more expected symlinks are absent or point to a non-existent target
- **THEN** each affected skill and the missing path are printed with a "MISSING" status, and the command exits with code 1

#### Scenario: No lock file for check
- **WHEN** `melon list --check` is run and `melon.lock` is absent
- **THEN** the command prints "No skills installed." and exits with code 0

### Requirement: Combine flags
The system SHALL allow `--pending` and `--check` to be used together in a single invocation.

#### Scenario: Both flags provided
- **WHEN** `melon list --pending --check` is run
- **THEN** the installed+check section is shown first, followed by the pending section

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
