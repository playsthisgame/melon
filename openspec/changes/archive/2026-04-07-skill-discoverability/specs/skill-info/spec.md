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
