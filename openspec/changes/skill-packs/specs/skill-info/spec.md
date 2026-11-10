## ADDED Requirements

### Requirement: melon info lists pack members
When the manifest of the requested path (at its latest semver tag, or the default branch if none) has `type: pack`, `melon info` SHALL indicate that the path is a pack and SHALL list each declared member with its constraint, sorted by name. Pack detection SHALL use the fetched manifest, so packs not present in any index are also detected.

#### Scenario: Info on a pack
- **WHEN** `melon info github.com/community/packs/spring-expert-pack` is run and the pack declares `spring-boot: "^1.0.0"` and `spring-data: "^2.0.0"`
- **THEN** the output SHALL identify the path as a pack and list `spring-boot ^1.0.0` and `spring-data ^2.0.0`

#### Scenario: Info on a pack in JSON
- **WHEN** `melon info <pack> --json` is run
- **THEN** the JSON object SHALL contain `"type": "pack"` and a `members` array of objects with `name` and `constraint`

#### Scenario: Info on a skill is unchanged
- **WHEN** `melon info` is run on a path whose manifest is absent or `type: skill`
- **THEN** the output SHALL be unchanged except that `--json` output SHALL include `"type": "skill"`
