## ADDED Requirements

### Requirement: Index entries can be marked as packs
An index entry in `melon-index.yaml` (public or custom) SHALL accept an optional `type` field with the values `skill` (default) or `pack`. Entries with any other value SHALL be treated as `skill`.

#### Scenario: Pack entry parsed
- **WHEN** an index contains an entry with `type: pack`
- **THEN** the parsed entry SHALL report its type as pack

### Requirement: Search results label packs
`melon search` SHALL display a `[pack]` label next to any result whose index entry is a pack, in both TUI and plain-text output.

#### Scenario: Pack result in plain text
- **WHEN** `melon search spring` is run in non-TTY mode and `spring-expert-pack` is a pack entry
- **THEN** its result line SHALL include `[pack]`

#### Scenario: Pack result in TUI
- **WHEN** `melon search spring` is run in a TTY and `spring-expert-pack` is a pack entry
- **THEN** its result row SHALL include `[pack]` after the path
