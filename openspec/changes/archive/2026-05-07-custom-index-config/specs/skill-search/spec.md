## MODIFIED Requirements

### Requirement: Search the curated index first

The CLI SHALL provide a `mln search <term>` command that resolves the active index set from the project's `melon.yaml` (if present), fetches each index `index.yaml`, filters entries matching the term against name, description, author, and tags fields, and displays results. When no `melon.yaml` is present or no `index` block is configured, the default melon public index is used. When both a custom and public index are active, custom index results are merged first with public index results deduplicated against them.

#### Scenario: Successful search with index results
- **WHEN** the user runs `mln search <term>` and matching entries exist in the index
- **THEN** featured entries are printed first, followed by non-featured entries, each showing the GitHub path, author, and description
- **THEN** the GitHub Topics fallback is NOT queried

#### Scenario: Successful search with custom index only (exclusive: true)
- **WHEN** `melon.yaml` has `index.exclusive: true` and matching entries exist in the custom index
- **THEN** only custom index results are shown; the public index is not queried

#### Scenario: Successful search with both indices active
- **WHEN** `melon.yaml` has `index.url` set and `exclusive` is false or absent
- **THEN** results from the custom index appear before results from the public index, with duplicates from the public index suppressed

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
