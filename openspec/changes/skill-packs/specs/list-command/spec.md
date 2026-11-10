## ADDED Requirements

### Requirement: melon list shows pack membership
`melon list` SHALL label pack entries with `(pack)` and SHALL print each pack's members indented beneath it. Dependencies that are not reached through any pack SHALL be printed at the top level as before. A dependency reached through several packs SHALL appear under each of them; a dependency that is both direct and a pack member SHALL also appear at the top level.

#### Scenario: Pack with members
- **WHEN** `melon.lock` contains `spring-expert-pack` (pack, direct), `spring-boot` (via the pack), and `pdf-skill` (direct)
- **THEN** `melon list` SHALL print `spring-expert-pack` labeled `(pack)` with `spring-boot` indented beneath it, and `pdf-skill` at the top level

#### Scenario: No packs installed
- **WHEN** no lock entry has `type: pack`
- **THEN** `melon list` output SHALL be identical to the output before this change

### Requirement: melon list --json includes provenance fields
Each object in `melon list --json` output SHALL include `type`, `direct`, and `via` (an empty array when the dependency has no parents).

#### Scenario: JSON output for a pack member
- **WHEN** `melon list --json` is run and `spring-boot` is locked via `spring-expert-pack`
- **THEN** its object SHALL contain `"type": "skill"`, `"direct": false`, and `"via": ["github.com/community/packs/spring-expert-pack"]`

### Requirement: melon list --check ignores pure packs
`melon list --check` SHALL NOT report a missing symlink for a pure pack.

#### Scenario: Pure pack not flagged
- **WHEN** `melon list --check` is run and a pure pack is installed
- **THEN** the pack SHALL NOT be reported as missing or broken
