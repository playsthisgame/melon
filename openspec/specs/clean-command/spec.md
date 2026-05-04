### Requirement: Clean removes orphaned cache entries
The `melon clean` command SHALL delete every directory under `.melon/` whose `name@version` does not match any entry in `melon.lock`. It SHALL print one line per removed entry and a summary count at the end.

#### Scenario: Orphaned cache entry is removed
- **WHEN** `.melon/` contains a directory not referenced by `melon.lock`
- **THEN** that directory is deleted and a removal message is printed for it

#### Scenario: Nothing to clean
- **WHEN** every directory in `.melon/` is referenced by `melon.lock`
- **THEN** no directories are deleted and "Nothing to clean." is printed

#### Scenario: Empty store
- **WHEN** `.melon/` does not exist or is empty
- **THEN** the command exits successfully with "Nothing to clean."

### Requirement: Clean removes orphaned symlinks for removed cache entries
For each cache entry removed by `clean`, the command SHALL also remove any corresponding symlinks in agent skill directories (e.g. `.claude/skills/`, `.windsurf/skills/`). The agent directories are derived from the project manifest's `tool_compat` or `outputs` fields.

#### Scenario: Symlink removed alongside cache entry
- **WHEN** an orphaned `.melon/` entry also has a symlink in an agent skill directory
- **THEN** the symlink is removed when the cache entry is cleaned

#### Scenario: Missing symlink silently skipped
- **WHEN** an orphaned `.melon/` entry has no corresponding symlink in the agent skill directory
- **THEN** the command continues without error

#### Scenario: No manifest present
- **WHEN** `melon.yaml` does not exist in the project directory
- **THEN** `.melon/` entries are still cleaned and a warning is printed that symlink removal was skipped

### Requirement: Clean is a no-op without a lock file
If `melon.lock` does not exist, `melon clean` SHALL print an informational message and exit successfully without modifying any files.

#### Scenario: Lock file absent
- **WHEN** `melon.lock` does not exist
- **THEN** the command prints "No melon.lock found. Run 'melon install' first." and exits with code 0

### Requirement: Clean does not modify melon.yaml or melon.lock
The `clean` command SHALL NOT write to `melon.yaml` or `melon.lock`.

#### Scenario: Lock and manifest are unchanged after clean
- **WHEN** `melon clean` runs successfully
- **THEN** `melon.yaml` and `melon.lock` have identical contents to before the command ran
