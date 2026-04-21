## ADDED Requirements

### Requirement: melon update with no args shows an interactive multi-select in TTY
When `melon update` is run with no arguments and stdout is a TTY, it SHALL display a multi-select list of all declared dependencies from `melon.yaml`. The first item in the list SHALL be "Update all". The user MAY select any combination of deps. Selecting "Update all" SHALL update every dep regardless of individual selections.

#### Scenario: Interactive list shown with all deps
- **WHEN** `melon update` is run in a TTY with dependencies in melon.yaml
- **THEN** a multi-select list SHALL appear with "Update all" as the first item followed by each dep name and its current constraint

#### Scenario: Selecting Update all updates every dep
- **WHEN** the user selects "Update all" and confirms
- **THEN** all declared dependencies SHALL be resolved against their existing constraints and updated if a newer compatible version exists

#### Scenario: Selecting individual deps updates only those
- **WHEN** the user selects a subset of deps and confirms
- **THEN** only the selected deps SHALL be resolved and updated; unselected deps SHALL remain unchanged

#### Scenario: No deps in melon.yaml
- **WHEN** `melon update` is run with no dependencies declared in melon.yaml
- **THEN** the command SHALL print "No skills in melon.yaml." and exit without error

#### Scenario: Non-TTY with no args errors
- **WHEN** `melon update` is run with no arguments and stdout is not a TTY
- **THEN** the command SHALL print an error asking the user to provide a dep name and exit with a non-zero code

### Requirement: melon update <dep> updates a single named dependency
When `melon update <dep>` is run with a dep name, it SHALL resolve the latest version satisfying the existing semver constraint in `melon.yaml` and run the install pipeline. If `<dep>` is not declared in `melon.yaml`, the command SHALL print a clear error and exit.

#### Scenario: Named dep is updated to latest compatible version
- **WHEN** `melon update github.com/owner/repo/skill` is run and a newer compatible version exists
- **THEN** the dep SHALL be updated to the latest compatible version and the install pipeline SHALL run

#### Scenario: Named dep not in melon.yaml
- **WHEN** `melon update github.com/owner/repo/skill` is run and the dep is not in melon.yaml
- **THEN** the command SHALL print `update: "github.com/owner/repo/skill" is not a dependency in melon.yaml` and exit with a non-zero code

### Requirement: melon update reports when everything is already up to date
When all selected deps are already at the latest version satisfying their constraints, the command SHALL print a message indicating everything is up to date and exit without modifying `melon.yaml`, `melon.lock`, or any symlinks.

#### Scenario: All selected deps already at latest
- **WHEN** `melon update` is run and no dep has a newer compatible version available
- **THEN** the command SHALL print "All selected skills are up to date." and exit with code 0

#### Scenario: Mix of up-to-date and updatable deps
- **WHEN** some selected deps are already at latest and others have updates
- **THEN** only the deps with updates SHALL be fetched and installed; a note SHALL be printed for the ones skipped

### Requirement: Branch-pinned deps are skipped during update
Deps whose constraint is a branch name (not a semver constraint) SHALL be excluded from update resolution. In interactive mode they SHALL not appear in the selection list. In targeted mode a warning SHALL be printed and the command SHALL exit without error.

#### Scenario: Branch-pinned dep excluded from interactive list
- **WHEN** `melon update` is run in TTY mode and a dep has a branch constraint (e.g. "main")
- **THEN** that dep SHALL not appear in the multi-select list

#### Scenario: Targeted update on branch-pinned dep
- **WHEN** `melon update github.com/owner/repo/skill` is run and that dep is pinned to a branch
- **THEN** the command SHALL print `update: "github.com/owner/repo/skill" is branch-pinned — use melon install to fetch latest` and exit with code 0

### Requirement: melon update hints when a newer major version exists outside the constraint
When resolving a dep and the absolute latest tag is outside the current constraint (e.g. `2.0.0` exists but constraint is `^1.x`), the command SHALL print a hint informing the user how to upgrade.

#### Scenario: Newer major version hint
- **WHEN** `melon update github.com/owner/repo/skill` is run, the dep is at the latest `^1.x` version, and `2.0.0` exists
- **THEN** the command SHALL print `hint: github.com/owner/repo/skill 2.0.0 is available — run melon add github.com/owner/repo/skill@^2.0.0 to upgrade`
