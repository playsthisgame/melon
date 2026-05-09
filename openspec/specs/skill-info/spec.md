## ADDED Requirements

### Requirement: Show metadata for a specific skill

The CLI SHALL provide a `mln info <github-path>` command that fetches and displays metadata for a skill at the given GitHub path before the user installs it. Description and author are sourced from the melon-index if the skill is listed there, falling back to the GitHub repo's about field.

#### Scenario: Skill found in the index with semver tags

- **WHEN** the user runs `mln info github.com/owner/repo` and the skill is in the melon-index and the repo has semver tags
- **THEN** the CLI displays the skill name, author (from index), description (from index), latest version, and all available versions

#### Scenario: Skill not in the index with semver tags

- **WHEN** the user runs `mln info github.com/owner/repo` and the skill is not in the melon-index but the repo exists and has semver tags
- **THEN** the CLI displays the skill name, description from the GitHub repo about field, latest version, and all available versions

#### Scenario: Skill with no semver tags

- **WHEN** the user runs `mln info github.com/owner/repo` and the repo has no semver tags
- **THEN** the CLI displays available branches in place of versions

#### Scenario: Skill path not found

- **WHEN** the user runs `mln info` with a GitHub path that does not exist or is not accessible
- **THEN** the CLI prints a clear error and exits with a non-zero code

#### Scenario: No argument provided

- **WHEN** the user runs `mln info` with no arguments
- **THEN** the CLI prints a usage error and exits with a non-zero code

### Requirement: Info command accepts monorepo subpaths

The `mln info` command SHALL accept subpath GitHub paths in the same format as `mln add`.

#### Scenario: Monorepo subpath

- **WHEN** the user runs `mln info github.com/owner/repo/path/to/skill`
- **THEN** the CLI fetches metadata for the repo at `github.com/owner/repo`, displays it, and notes the subpath

## ADDED Requirements (json-output)

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
