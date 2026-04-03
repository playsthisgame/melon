### Requirement: mln add resolves and pins the latest compatible version when none is given
When the user runs `mln add <dep>` with no version constraint, the `add` command SHALL call `LatestTag` to find the latest semver tag for the dependency and add it to `mln.yaml` with a caret (`^`) constraint pinned to that version.

#### Scenario: Add dep without version — latest tag is resolved
- **WHEN** the user runs `mln add alice/pdf-skill` and the latest tag for `alice/pdf-skill` is `v1.3.0`
- **THEN** `mln.yaml` SHALL contain `alice/pdf-skill: "^1.3.0"` after the command completes

#### Scenario: Add dep with explicit version constraint
- **WHEN** the user runs `mln add alice/pdf-skill@^1.2.0`
- **THEN** `mln.yaml` SHALL contain `alice/pdf-skill: "^1.2.0"` without fetching the latest tag

### Requirement: mln add runs a full install after updating mln.yaml
After writing the updated `mln.yaml`, `mln add` SHALL run the full install pipeline (resolve → fetch → lock → place) so the dep is immediately available in agent directories.

#### Scenario: Dep is placed after add
- **WHEN** `mln add alice/pdf-skill` completes successfully
- **THEN** the skill directory SHALL be present in each agent directory derived from `agent_compat`, and `mln.lock` SHALL contain an entry for `alice/pdf-skill`

### Requirement: mln add warns and updates when the dep already exists
If the dependency is already present in `mln.yaml`, `mln add` SHALL print a warning, update the constraint to the new value, and continue with a full install.

#### Scenario: Existing dep constraint is updated
- **WHEN** `mln.yaml` already contains `alice/pdf-skill: "^1.0.0"` and the user runs `mln add alice/pdf-skill@^1.3.0`
- **THEN** `mln.yaml` SHALL be updated to `alice/pdf-skill: "^1.3.0"`, a warning SHALL be printed, and install SHALL run

### Requirement: mln install --frozen fails if mln.lock would change
When `--frozen` is set, `mln install` SHALL run the full resolve and fetch pipeline, compute the resulting lock, diff it against the lock file on disk, and exit with a non-zero status code if any differences exist. The lock file SHALL NOT be written.

#### Scenario: Frozen install succeeds when lock is up to date
- **WHEN** `mln install --frozen` is run and the resolved lock matches `mln.lock` on disk
- **THEN** the command SHALL exit with status 0 and not modify `mln.lock`

#### Scenario: Frozen install fails when lock would change
- **WHEN** `mln install --frozen` is run after `mln.yaml` was modified without regenerating `mln.lock`
- **THEN** the command SHALL exit with a non-zero status and print which dependencies were added, removed, or updated
