### Requirement: melon diff shows file-level changes between locked and target versions
`melon diff <dep>` SHALL resolve a "from" version (the dep's version in `melon.lock`) and a "to" version, materialize both version trees as directories under `.melon/`, and print a unified, file-by-file diff of their contents. The diff SHALL classify each file path as added, removed, or changed, and SHALL print unified hunks for changed text files. Added and removed files SHALL be shown with file headers indicating the operation.

#### Scenario: Dep has content changes between locked and latest compatible
- **WHEN** a dep is locked at `1.0.0` and `1.1.0` (the latest version satisfying its constraint) modifies `SKILL.md`
- **THEN** the command SHALL print a unified diff for `SKILL.md` showing the changed lines and exit with code 0

#### Scenario: Target adds and removes files
- **WHEN** the target version adds a new file and deletes an existing file relative to the locked version
- **THEN** the command SHALL print an added-file header for the new file and a removed-file header for the deleted file

#### Scenario: No changes between versions
- **WHEN** the locked version and the resolved target version have identical tree hashes
- **THEN** the command SHALL print `No changes` and exit with code 0 without rendering any hunks

### Requirement: melon diff resolves the target version from the constraint or an explicit override
By default `melon diff <dep>` SHALL use the latest version satisfying the dep's constraint in `melon.yaml` as the target. If the argument includes an explicit `@<target>` suffix, that exact version or branch SHALL be used as the target instead. An explicit `@<target>` SHALL bypass constraint-based resolution.

#### Scenario: Default target is latest compatible version
- **WHEN** a dep with constraint `^1.0.0` is locked at `1.0.0` and `1.2.0` is the newest tag satisfying `^1.0.0`
- **THEN** the command SHALL diff `1.0.0` against `1.2.0`

#### Scenario: Explicit version target
- **WHEN** the user runs `melon diff <dep>@2.0.0`
- **THEN** the command SHALL diff the locked version against `2.0.0` regardless of the dep's constraint

#### Scenario: Explicit branch target
- **WHEN** the user runs `melon diff <dep>@main`
- **THEN** the command SHALL diff the locked version against the current tree of branch `main`

### Requirement: melon diff requires the dep to be installed and resolvable
The command SHALL require the named dep to exist in `melon.lock`; if it does not, the command SHALL exit non-zero with a message directing the user to run `melon install` first. The command SHALL also exit non-zero if the named dep does not appear in `melon.yaml`, or if the requested target version cannot be resolved.

#### Scenario: Dep not in lock file
- **WHEN** the named dep is declared in `melon.yaml` but absent from `melon.lock`
- **THEN** the command SHALL print an error indicating the dep is not installed and SHALL suggest running `melon install`, exiting non-zero

#### Scenario: Unknown dep
- **WHEN** the named dep is not present in `melon.yaml`
- **THEN** the command SHALL print an error that the dependency is not declared and exit non-zero

#### Scenario: Unresolvable target version
- **WHEN** an explicit `@<version>` target does not correspond to any tag in the repository
- **THEN** the command SHALL print an error that the target could not be resolved and exit non-zero

### Requirement: melon diff requires an explicit target for branch-pinned deps
If a dep's constraint in `melon.yaml` is a branch name rather than a semver constraint, "latest compatible" is undefined. In that case `melon diff <dep>` without an explicit `@<target>` SHALL exit non-zero with guidance to supply a target.

#### Scenario: Branch-pinned dep without explicit target
- **WHEN** a dep has a branch constraint such as `main` and the user runs `melon diff <dep>` with no `@<target>`
- **THEN** the command SHALL print an error explaining that a target is required for branch-pinned deps and exit non-zero

#### Scenario: Branch-pinned dep with explicit target
- **WHEN** a dep has a branch constraint and the user runs `melon diff <dep>@1.0.0`
- **THEN** the command SHALL diff the locked version against `1.0.0` and exit with code 0

### Requirement: melon diff supports a --stat summary mode
When invoked with `--stat`, the command SHALL print only a per-file summary — each changed, added, or removed path with its added and removed line counts — followed by a totals line, and SHALL NOT print unified hunks.

#### Scenario: Stat mode prints summary only
- **WHEN** the user runs `melon diff <dep> --stat` and the target changes two files
- **THEN** the output SHALL list each changed file with its `+`/`-` line counts and a totals line, with no unified hunks

### Requirement: melon diff renders colored output only in a TTY
When stdout is a TTY and `--no-color` is not set, the command SHALL color added lines and removed lines distinctly. When stdout is not a TTY, or `--no-color` is set, the command SHALL emit no ANSI escape codes.

#### Scenario: Color in TTY
- **WHEN** `melon diff <dep>` is run with stdout attached to a terminal
- **THEN** added and removed lines SHALL be visually colored

#### Scenario: No color when piped
- **WHEN** `melon diff <dep>` output is piped to a file or `--no-color` is passed
- **THEN** the output SHALL contain no ANSI escape codes

### Requirement: melon diff handles binary files without rendering hunks
For files that are not valid UTF-8 text (e.g. contain NUL bytes), the command SHALL report the change as a binary file change rather than attempting to render a unified diff.

#### Scenario: Binary file changed
- **WHEN** a binary asset differs between the locked and target versions
- **THEN** the command SHALL print a line indicating the binary file changed instead of unified hunks

### Requirement: melon diff is read-only with respect to project manifests
The command SHALL NOT modify `melon.yaml` or `melon.lock`. Its only permitted side effect is populating the `.melon/` cache with the target version via the existing fetch path.

#### Scenario: Manifest and lock unchanged after diff
- **WHEN** `melon diff <dep>` runs to completion
- **THEN** `melon.yaml` and `melon.lock` SHALL be byte-for-byte unchanged
