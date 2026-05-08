### Requirement: mln add resolves and pins the latest compatible version when none is given
When the user runs `mln add <dep>` with no version constraint, the `add` command SHALL call `LatestTag` to find the latest semver tag for the dependency and add it to `mln.yaml` with a caret (`^`) constraint pinned to that version.

#### Scenario: Add dep without version — latest tag is resolved
- **WHEN** the user runs `mln add alice/pdf-skill` and the latest tag for `alice/pdf-skill` is `v1.3.0`
- **THEN** `mln.yaml` SHALL contain `alice/pdf-skill: "^1.3.0"` after the command completes

#### Scenario: Add dep with explicit version constraint
- **WHEN** the user runs `mln add alice/pdf-skill@^1.2.0`
- **THEN** `mln.yaml` SHALL contain `alice/pdf-skill: "^1.2.0"` without fetching the latest tag

### Requirement: mln add runs a full install after updating mln.yaml
After writing the updated `mln.yaml`, `mln add` SHALL run the full install pipeline (resolve → fetch → lock → place) so the dep is immediately available in agent directories. When `vendor: false`, the gitignore sync step SHALL also run so the new symlink path is added to `.gitignore`.

#### Scenario: Dep is placed after add
- **WHEN** `mln add alice/pdf-skill` completes successfully
- **THEN** the skill directory SHALL be present in each agent directory derived from `agent_compat`, and `mln.lock` SHALL contain an entry for `alice/pdf-skill`

#### Scenario: New dep symlink path added to .gitignore when vendor is false

- **WHEN** `vendor: false` and the user runs `mln add alice/pdf-skill`
- **THEN** `.gitignore` SHALL contain an entry for the symlink path of `pdf-skill` after the command completes

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

## ADDED Requirements

### Requirement: melon add validates source against allowed_sources before writing to melon.yaml
When a `policy` block with `allowed_sources` is present, `melon add` SHALL check the dependency path against the allowlist before modifying `melon.yaml` or running install. If the source is not permitted, the command SHALL exit non-zero with a clear error message and leave `melon.yaml` unchanged.

#### Scenario: add blocked by policy — melon.yaml unchanged

- **WHEN** `allowed_sources` is `[github.com/my-company/*]` and the user runs `melon add github.com/public/cool-skill`
- **THEN** the command SHALL exit non-zero, print a message identifying the blocked source and the active policy, and `melon.yaml` SHALL remain unmodified

#### Scenario: add permitted by policy — proceeds normally

- **WHEN** `allowed_sources` is `[github.com/my-company/*]` and the user runs `melon add github.com/my-company/approved-skill`
- **THEN** the command SHALL proceed normally, writing the dep to `melon.yaml` and running install

#### Scenario: add with no policy — no restriction

- **WHEN** no `policy` block is present in `melon.yaml`
- **THEN** `melon add` SHALL not perform any source validation and SHALL accept any dependency path
