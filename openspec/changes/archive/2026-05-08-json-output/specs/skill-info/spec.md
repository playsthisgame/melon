## ADDED Requirements

### Requirement: melon info --json emits skill metadata as a JSON object

When `--json` is set, `melon info <path>` SHALL write a JSON object to stdout containing: `name`, `description`, `author`, `latest_version`, `versions` (array of version strings, or empty if none), and `branches` (array of branch name strings, populated when no semver tags exist).

#### Scenario: JSON output for skill with semver tags in index

- **WHEN** `melon info github.com/owner/repo --json` is run and the skill is in the index and has semver tags
- **THEN** stdout is a JSON object with `name`, `description`, `author` from the index, `latest_version`, and a non-empty `versions` array

#### Scenario: JSON output for skill not in index

- **WHEN** `melon info github.com/owner/repo --json` is run and the skill is not in the index
- **THEN** stdout is a JSON object with `description` from the GitHub repo about field, empty `author`, and version fields populated from tags

#### Scenario: JSON output for skill with no semver tags

- **WHEN** `melon info github.com/owner/repo --json` is run and the repo has no semver tags
- **THEN** stdout is a JSON object with an empty `versions` array and a non-empty `branches` array

#### Scenario: JSON error for unknown path

- **WHEN** `melon info github.com/owner/nonexistent --json` is run
- **THEN** stderr contains `{"error": "..."}` and the command exits non-zero
