## ADDED Requirements

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
