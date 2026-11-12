# Custom Index Config Spec

## ADDED Requirements

### Requirement: Index block in melon.yaml
The manifest SHALL support an optional `index` block with two fields: `url` (string) pointing to a custom `index.yaml`, and `exclusive` (bool) that when `true` suppresses the default public melon index. Both fields are optional; omitting the block preserves existing behaviour.

#### Scenario: No index block present
- **WHEN** `melon.yaml` has no `index` block
- **THEN** the CLI uses the default public melon index URL

#### Scenario: Custom URL provided, exclusive false
- **WHEN** `melon.yaml` contains `index.url` and `index.exclusive` is `false` (or omitted)
- **THEN** both the custom index and the default public index are used as sources

#### Scenario: Custom URL provided, exclusive true
- **WHEN** `melon.yaml` contains `index.url` and `index.exclusive: true`
- **THEN** only the custom index URL is fetched; the default public index is not queried

#### Scenario: No melon.yaml in working directory
- **WHEN** `melon search` or `melon info` is run in a directory with no `melon.yaml`
- **THEN** the CLI falls back to the default public index and does not error

### Requirement: Custom index result ordering
When both a custom index and the public index are active, the CLI SHALL present custom index results before public index results. Entries with the same `name` field in the public index SHALL be suppressed if already present in the custom index results.

#### Scenario: Overlapping entries
- **WHEN** both indices return an entry with the same `name`
- **THEN** only the custom index entry appears in the merged results

#### Scenario: No overlap
- **WHEN** the custom index and public index return distinct entries
- **THEN** all custom index entries appear before all public index entries

### Requirement: Private GitHub indices
When an index URL is on a GitHub host (`github.com`, `raw.githubusercontent.com`, `api.github.com`) and `GITHUB_TOKEN` (or, when unset, `GH_TOKEN`) is set, the CLI SHALL send it as a Bearer token so indices in private repositories can be read. The token SHALL NOT be sent to any other host.

#### Scenario: Private index with token
- **WHEN** `index.urls` contains `github.com/my-company/skill-registry/index.yaml` in a private repository and `GITHUB_TOKEN` can read it
- **THEN** `melon search` and `melon info` SHALL include that index's entries

#### Scenario: Private index without token
- **WHEN** the private index returns 404 and no token is set
- **THEN** the CLI SHALL warn on stderr that GitHub refused access to the index (named in its `github.com/...` form), explain that this is an access problem rather than a melon error, and show the fix: `export GITHUB_TOKEN=$(gh auth token)` when `gh` is installed, otherwise a generic `export GITHUB_TOKEN=<token>`

#### Scenario: Token cannot read the private index
- **WHEN** a token is set and GitHub still refuses the index
- **THEN** the warning SHALL name the variable the token came from and say to check the URL and the token's read access to the repository

#### Scenario: Invalid token does not break public indices
- **WHEN** the token is invalid or expired and GitHub refuses a request carrying it
- **THEN** the CLI SHALL retry once without the token, so public indices still load

#### Scenario: No results while an index failed
- **WHEN** no entries match and at least one index failed to load
- **THEN** the result line SHALL say no skills were found in the indices melon could load, with how many failed, rather than a plain "No skills found"

#### Scenario: Token not sent to other hosts
- **WHEN** an index URL is on a non-GitHub host and `GITHUB_TOKEN` is set
- **THEN** the request SHALL NOT carry an Authorization header

### Requirement: Failed indices are reported, not hidden
When any active index fails to load, the CLI SHALL print a warning to stderr naming it and continue with the remaining indices, whether or not the others return results.

#### Scenario: Private index fails, public index succeeds
- **WHEN** the custom index cannot be loaded and the public index returns matches
- **THEN** the matches SHALL be shown and a warning for the custom index SHALL be printed to stderr
