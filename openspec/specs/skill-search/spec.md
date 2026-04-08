# Skill Search Spec

## ADDED Requirements

### Requirement: Search the curated index first

The CLI SHALL provide a `mln search <term>` command that fetches the melon-index `index.yml` from `https://raw.githubusercontent.com/playsthisgame/melon-index/main/index.yml`, filters entries matching the term against name, description, author, and tags fields, and displays results.

#### Scenario: Successful search with index results

- **WHEN** the user runs `mln search <term>` and matching entries exist in the index
- **THEN** featured entries are printed first, followed by non-featured entries, each showing the GitHub path, author, and description
- **THEN** the GitHub Topics fallback is NOT queried

#### Scenario: Index returns no results — fall back to GitHub Topics

- **WHEN** the user runs `mln search <term>` and no index entries match
- **THEN** the CLI queries the GitHub Topics API for repos tagged `melon-skill` matching the term
- **THEN** results from GitHub Topics are displayed with a note indicating they are community-tagged and not in the curated index

#### Scenario: Index unreachable — fall back to GitHub Topics

- **WHEN** the melon-index URL cannot be fetched (network error or non-200 response)
- **THEN** the CLI falls back to querying the GitHub Topics API and displays those results

#### Scenario: No results from either source

- **WHEN** the user runs `mln search <term>` and neither the index nor GitHub Topics return results
- **THEN** the CLI prints a message indicating no skills were found and exits with code 0

#### Scenario: Search with no term provided

- **WHEN** the user runs `mln search` with no arguments
- **THEN** the CLI prints a usage error and exits with a non-zero code

#### Scenario: GitHub Topics rate limit exceeded

- **WHEN** the GitHub Topics API returns a 403 or 429 response during fallback
- **THEN** the CLI prints a clear error explaining the rate limit and suggests setting a `GITHUB_TOKEN` environment variable

#### Scenario: GITHUB_TOKEN is set

- **WHEN** the `GITHUB_TOKEN` environment variable is set and the Topics fallback is triggered
- **THEN** the CLI includes the token as a Bearer authorization header in the GitHub API request

### Requirement: Interactive result selection in TTY mode

When stdout is a TTY, search results SHALL be presented as an interactive single-select list using the existing bubbletea TUI infrastructure. The user navigates with arrow keys and presses Enter to select a skill.

#### Scenario: User selects a skill from results

- **WHEN** results are shown in the interactive list and the user presses Enter on an item
- **THEN** the CLI prints the selected skill's full `mln add <github-path>` command and offers to run it

#### Scenario: User exits without selecting

- **WHEN** the user presses Escape or Ctrl+C during the interactive list
- **THEN** the CLI exits cleanly with code 0 and no action is taken

#### Scenario: Non-TTY mode (piped or CI output)

- **WHEN** stdout is not a TTY (e.g. piped to another command or run in CI)
- **THEN** results are printed as plain text, one per line, with no interactive prompt

### Requirement: Search results are formatted as installable paths

Each search result SHALL display the dependency path in the exact format accepted by `mln add` so users can copy-paste it directly.

#### Scenario: Result path format

- **WHEN** a result has name `github.com/acme/skills`
- **THEN** the displayed path is `github.com/acme/skills` with no `https://` prefix and no trailing slash

### Requirement: Featured skills are surfaced first

The CLI SHALL display entries with `featured: true` before non-featured entries in search results, with a visual indicator distinguishing them from non-featured results.

#### Scenario: Mixed featured and non-featured results

- **WHEN** `mln search <term>` matches both featured and non-featured entries
- **THEN** all featured matches appear before any non-featured matches in the list with a visible marker
