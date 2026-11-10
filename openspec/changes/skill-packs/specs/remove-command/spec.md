## ADDED Requirements

### Requirement: Removing a pack prunes members no longer required
When `melon remove <name>` targets a dependency that is locked as `type: pack`, or that any lock entry lists in `via`, the command SHALL run the full install pipeline after updating `melon.yaml`. Every dependency that is no longer reachable from the remaining manifest SHALL be unplaced and removed from `.melon/` and `melon.lock`. Dependencies still reachable — directly or through another pack — SHALL be kept.

#### Scenario: Pack members are pruned
- **WHEN** `spring-expert-pack` is the only thing requiring `spring-boot` and `spring-data`, and the user runs `melon remove github.com/community/packs/spring-expert-pack`
- **THEN** the pack, `spring-boot`, and `spring-data` SHALL be removed from `melon.lock`, `.melon/`, and every harness directory

#### Scenario: Shared member is kept
- **WHEN** `junit-skill` is required by both `spring-expert-pack` and `java-testing-pack`, and the user removes `spring-expert-pack`
- **THEN** `junit-skill` SHALL remain installed and its lock entry's `via` SHALL contain only `java-testing-pack`

#### Scenario: Member that is also direct is kept
- **WHEN** `spring-boot` is a direct dependency and a member of `spring-expert-pack`, and the user removes the pack
- **THEN** `spring-boot` SHALL remain installed with `direct: true` and no `via`

### Requirement: Removing a pack member directly is rejected with guidance
When `melon remove <name>` is run for a dependency that is not in `melon.yaml` but is present in `melon.lock` via a pack, the command SHALL exit non-zero with an error that names the pack(s) requiring it, and SHALL NOT modify `melon.yaml` or `melon.lock`.

#### Scenario: Remove a transitive member
- **WHEN** `spring-boot` is installed only via `spring-expert-pack` and the user runs `melon remove github.com/acme/skills/spring-boot`
- **THEN** the command SHALL exit non-zero with an error stating that `spring-boot` is required by `spring-expert-pack`
