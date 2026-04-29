## ADDED Requirements

### Requirement: melon outdated shows a table of deps with available updates
`melon outdated` SHALL read `melon.yaml` and `melon.lock`, resolve the latest compatible and absolute latest versions for each semver-constrained dep, and print a formatted table. The table SHALL include columns for: dep name, current locked version, latest compatible version, and absolute latest (when outside the constraint). Deps whose locked version equals the latest compatible version SHALL be omitted from the table.

#### Scenario: Dep has a newer compatible version
- **WHEN** a dep is locked at `1.0.1` and `1.0.2` satisfies its `^1.0.0` constraint
- **THEN** the table SHALL show a row with the dep name, `1.0.1` as current, `1.0.2` as latest compatible

#### Scenario: Dep has a newer major outside the constraint
- **WHEN** a dep is locked at `1.2.0` (latest within `^1.x`) and `2.0.0` exists but is outside the constraint
- **THEN** the table row SHALL additionally show `2.0.0` in the absolute latest column with a visual indicator that it requires a constraint change

#### Scenario: All deps are up to date
- **WHEN** all semver-constrained deps are already at the latest compatible version
- **THEN** no table is printed and the command SHALL print "All skills are up to date." and exit with code 0

#### Scenario: No dependencies declared
- **WHEN** `melon.yaml` has no dependencies
- **THEN** the command SHALL print "No dependencies declared in melon.yaml." and exit with code 0

### Requirement: melon outdated exits with code 1 when any dep is outdated
If the table contains any rows (i.e. at least one dep has a newer compatible version), the command SHALL exit with code 1. This enables CI pipelines to fail on stale dependencies.

#### Scenario: Outdated dep triggers exit code 1
- **WHEN** at least one dep has a newer compatible version available
- **THEN** the command SHALL exit with code 1 after printing the table

#### Scenario: Up-to-date exits with code 0
- **WHEN** no dep has a newer compatible version available
- **THEN** the command SHALL exit with code 0

### Requirement: melon outdated skips branch-pinned deps
Deps whose constraint is a branch name (e.g. `main`, `HEAD`) SHALL be excluded from version resolution and SHALL NOT appear in the outdated table. If branch-pinned deps exist, a single note SHALL be printed listing them.

#### Scenario: Branch-pinned dep excluded from table
- **WHEN** a dep has a branch constraint such as `main`
- **THEN** it SHALL NOT appear in the outdated table

#### Scenario: Branch-pinned deps noted
- **WHEN** one or more branch-pinned deps exist
- **THEN** the command SHALL print a line such as `note: skipped 1 branch-pinned dep(s): github.com/owner/repo/skill`

### Requirement: melon outdated shows spinner during resolution in TTY
When stdout is a TTY, the command SHALL display a `bubbles/spinner` with the message `Checking for updates…` while resolving versions. The spinner SHALL clear before the table or up-to-date message is printed.

#### Scenario: Spinner shown in TTY
- **WHEN** `melon outdated` is run in a TTY environment
- **THEN** a spinning animation with the label `Checking for updates…` SHALL be visible during network resolution

#### Scenario: No spinner in non-TTY
- **WHEN** `melon outdated` is run with stdout not a TTY
- **THEN** no spinner or ANSI codes SHALL be emitted

### Requirement: melon outdated treats unlocked deps as not installed
If a dep appears in `melon.yaml` but has no entry in `melon.lock` (e.g. install has never been run), its current version SHALL be shown as `(not installed)` in the table, and it SHALL always appear as outdated.

#### Scenario: Dep in manifest but not in lock
- **WHEN** a dep is declared in `melon.yaml` but absent from `melon.lock`
- **THEN** the table SHALL show `(not installed)` as the current version for that dep and include it in the outdated output

#### Scenario: Missing lock file treated as all unlocked
- **WHEN** `melon.lock` does not exist
- **THEN** all deps SHALL be shown as `(not installed)` and the table SHALL list all of them
