## ADDED Requirements

### Requirement: melon add installs a pack's full member graph
When the dependency being added is a pack, or its manifest declares any `dependencies`, `melon add` SHALL resolve the full dependency graph with the new dependency included and run the full install pipeline, so every member and transitive dependency is fetched, locked, and placed.

#### Scenario: Adding a pack installs members
- **WHEN** the user runs `melon add github.com/community/packs/spring-expert-pack` and the pack declares `spring-boot` and `spring-data`
- **THEN** after the command, `melon.yaml` SHALL contain the pack, and `melon.lock` SHALL contain the pack, `spring-boot`, and `spring-data`, with both skills placed

#### Scenario: Adding a skill with transitive deps installs them
- **WHEN** the user adds a `type: skill` dependency whose manifest declares `bob/base-utils`
- **THEN** `bob/base-utils` SHALL also be installed

### Requirement: melon add shows pack members and confirms before writing
When adding a pack, `melon add` SHALL check `allowed_sources` against every resolved dependency and SHALL print the pack's direct members (marking nested packs with `(pack)`) before modifying `melon.yaml`. In a TTY it SHALL prompt for confirmation with a default of yes. The prompt SHALL be skipped when `--yes`/`-y` is passed or when stdout is not a TTY. If the user declines or a member is blocked by policy, `melon.yaml` and `melon.lock` SHALL be left unchanged.

#### Scenario: Members are listed before install
- **WHEN** the user runs `melon add <pack>` in non-TTY mode
- **THEN** the member list SHALL be printed and the install SHALL proceed without prompting

#### Scenario: User declines the pack
- **WHEN** the user runs `melon add <pack>` in a TTY and answers `n` at the prompt
- **THEN** the command SHALL exit 0 and `melon.yaml` SHALL be unchanged

#### Scenario: --yes skips the prompt
- **WHEN** the user runs `melon add <pack> --yes` in a TTY
- **THEN** no prompt SHALL be shown and the install SHALL proceed

#### Scenario: Member blocked by policy
- **WHEN** `allowed_sources` is `[github.com/my-company/*]` and the pack being added declares `github.com/public/spring-boot`
- **THEN** the command SHALL exit non-zero naming `github.com/public/spring-boot`, and `melon.yaml` SHALL be unchanged
