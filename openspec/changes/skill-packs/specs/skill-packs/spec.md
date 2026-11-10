## ADDED Requirements

### Requirement: melon.yaml supports a type field
`melon.yaml` SHALL accept an optional top-level `type` field with the values `skill` or `pack`. When absent, the manifest SHALL be treated as `type: skill`. The legacy values `agent`, `workflow`, `persona`, and `memory` (from the earlier schema) SHALL be accepted and treated as `skill`. Any other value SHALL cause manifest loading to fail with an error naming the invalid value and listing the valid values.

#### Scenario: Type absent defaults to skill
- **WHEN** a dependency's `melon.yaml` has no `type` field
- **THEN** melon SHALL treat the dependency as a skill and behave exactly as before this change

#### Scenario: Pack type is accepted
- **WHEN** a `melon.yaml` contains `type: pack` and a `dependencies` map
- **THEN** the manifest SHALL load successfully and the dependency SHALL be treated as a pack

#### Scenario: Unknown type is rejected
- **WHEN** a `melon.yaml` contains `type: bundle`
- **THEN** loading SHALL fail with an error mentioning `bundle` and the valid values `skill` and `pack`

#### Scenario: Legacy type is accepted as a skill
- **WHEN** a dependency's `melon.yaml` contains `type: agent`
- **THEN** the manifest SHALL load successfully and the dependency SHALL be treated as a skill

### Requirement: Installing a pack installs its members
When a pack is a dependency (direct or transitive), every dependency declared in the pack's `melon.yaml` SHALL be resolved, fetched, locked, and placed as part of the same install.

#### Scenario: Pack members are installed
- **WHEN** the project's `melon.yaml` depends on `github.com/community/packs/spring-expert-pack` and that pack declares `github.com/acme/skills/spring-boot: "^1.0.0"` and `github.com/acme/skills/spring-data: "^2.0.0"`
- **THEN** after `melon install`, `melon.lock` SHALL contain the pack, `spring-boot`, and `spring-data`, and both member skills SHALL be placed in every harness directory

#### Scenario: Pack depends on another pack
- **WHEN** pack A declares pack B as a dependency and pack B declares skill C
- **THEN** `melon install` SHALL install A, B, and C, and SHALL place C

### Requirement: Pure packs are not placed
A pack whose directory does not contain its entrypoint file (default `SKILL.md`) is a pure pack. A pure pack SHALL be fetched into `.melon/` and recorded in `melon.lock`, but SHALL NOT be symlinked into any harness directory.

#### Scenario: Pure pack has no symlink
- **WHEN** `spring-expert-pack` has `type: pack` and contains only `melon.yaml` and the project is installed for `claude-code`
- **THEN** `.claude/skills/spring-expert-pack` SHALL NOT exist after `melon install`

### Requirement: Hybrid packs are placed
A pack whose directory contains its entrypoint file is a hybrid pack. A hybrid pack SHALL be placed exactly like a skill, in addition to its members being installed.

#### Scenario: Hybrid pack has a symlink
- **WHEN** `spring-expert-pack` has `type: pack` and contains `melon.yaml` and `SKILL.md`
- **THEN** after `melon install`, `.claude/skills/spring-expert-pack` SHALL be a symlink to the pack's cache directory, and its members SHALL also be placed

### Requirement: melon.lock records type and provenance
Each `melon.lock` entry SHALL record `type` (`skill` or `pack`), `direct` (true when the dependency is listed in the project's `melon.yaml`), and `via` (the sorted, de-duplicated list of resolved dependencies whose manifests declare it). `direct` SHALL be omitted when false and `via` SHALL be omitted when empty.

#### Scenario: Member reached only through a pack
- **WHEN** `spring-boot` is declared only by `spring-expert-pack`
- **THEN** the lock entry for `spring-boot` SHALL have `type: skill`, no `direct` field, and `via: [github.com/community/packs/spring-expert-pack]`

#### Scenario: Skill is both direct and a pack member
- **WHEN** `spring-boot` is listed in the project's `melon.yaml` and also declared by `spring-expert-pack`
- **THEN** the lock entry SHALL have `direct: true` and `via` SHALL contain `spring-expert-pack`

#### Scenario: Skill required by two packs
- **WHEN** both `spring-expert-pack` and `java-testing-pack` declare `junit-skill`
- **THEN** the lock entry for `junit-skill` SHALL list both packs in `via`, sorted alphabetically

#### Scenario: Old lock file without provenance loads
- **WHEN** `melon.lock` was written by an older melon and has no `type`, `direct`, or `via` fields
- **THEN** it SHALL load without error and the next non-frozen `melon install` SHALL rewrite it with the new fields

### Requirement: Install warns on exact pins in packs
When a resolved dependency is a pack and any of its member constraints is an exact version (no `^` or `~` and not a branch), `melon install` SHALL print a one-line warning naming the pack and the pinned member, recommending a range.

#### Scenario: Exact pin in a pack
- **WHEN** `spring-expert-pack` declares `spring-boot: "1.4.0"`
- **THEN** `melon install` SHALL print a warning naming `spring-expert-pack` and `spring-boot` and SHALL still complete the install

### Requirement: Frozen install detects provenance changes
`melon install --frozen` SHALL fail when the newly computed lock would differ from `melon.lock` in any entry's `version`, `type`, `direct`, or `via`, except that an existing entry with no `type` field SHALL be considered matching on provenance when its version matches.

#### Scenario: Skill becomes direct
- **WHEN** `spring-boot` was locked only `via` a pack and the user adds it directly to `melon.yaml` without re-installing
- **THEN** `melon install --frozen` SHALL exit non-zero reporting the lock is out of date

#### Scenario: Legacy lock passes frozen
- **WHEN** `melon.lock` has no provenance fields and all versions match the resolution
- **THEN** `melon install --frozen` SHALL succeed
