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
