## MODIFIED Requirements

### Requirement: mln remove removes a named dependency from melon.yml
When the user runs `mln remove <name>`, the command SHALL remove the entry for `<name>` from `melon.yaml` and write the updated file to disk. When run with no arguments in a TTY, the command SHALL instead launch an interactive selector (see `remove-interactive` capability).

#### Scenario: Dependency is removed from melon.yml
- **WHEN** `melon.yaml` contains `alice/pdf-skill: "^1.3.0"` and the user runs `mln remove alice/pdf-skill`
- **THEN** `melon.yaml` SHALL no longer contain an entry for `alice/pdf-skill` after the command completes

#### Scenario: No argument in TTY launches interactive mode
- **WHEN** `mln remove` is run with no arguments in a TTY
- **THEN** the command SHALL launch the interactive multi-select selector instead of exiting with an argument error

### Requirement: mln remove errors if the dependency does not exist
If `<name>` is not present in `melon.yml`, `mln remove` SHALL exit with a non-zero status and print an error message indicating the dependency was not found. It SHALL NOT modify `melon.yml` or `melon.lock`.

#### Scenario: Remove unknown dependency returns error
- **WHEN** `melon.yml` does not contain `alice/unknown-skill` and the user runs `mln remove alice/unknown-skill`
- **THEN** the command SHALL exit non-zero and print an error referencing the dependency name

### Requirement: mln remove unlinks the skill from all agent directories and removes the cache entry
After updating `melon.yml`, `mln remove` SHALL run the full install pipeline, which SHALL remove the agent directory symlink for the named skill from every agent directory derived from `agent_compat` (or `outputs` if declared), and SHALL delete the dependency's `.melon/` cache directory. The skill name used for the symlink is the last path segment of the dependency name. When `vendor: false`, the removed symlink's path SHALL also be removed from `.gitignore`.

#### Scenario: Symlink is removed from agent directories
- **WHEN** `alice/pdf-skill` is linked as `pdf-skill` in `.claude/skills/` and the user runs `mln remove alice/pdf-skill`
- **THEN** the symlink at `.claude/skills/pdf-skill` SHALL no longer exist after the command completes

#### Scenario: Cache entry is removed from .melon/
- **WHEN** the user runs `mln remove alice/pdf-skill` successfully
- **THEN** the `.melon/` cache directory for `alice/pdf-skill` SHALL no longer exist after the command completes

#### Scenario: Missing symlink is not an error
- **WHEN** the symlink for the dependency does not exist in one or more agent directories (e.g. it was manually deleted)
- **THEN** `mln remove` SHALL continue without error

#### Scenario: Removed dep symlink path deleted from .gitignore when vendor is false
- **WHEN** `vendor: false` and the user runs `mln remove alice/pdf-skill`
- **THEN** `.gitignore` SHALL NOT contain an entry for the `pdf-skill` symlink path after the command completes

### Requirement: mln remove updates melon.lock via a full install
After updating `melon.yml`, `mln remove` SHALL run the full install pipeline so that `melon.lock` is regenerated to reflect the removed dependency.

#### Scenario: Lock file no longer contains the removed dependency
- **WHEN** the user runs `mln remove alice/pdf-skill` successfully
- **THEN** `melon.lock` SHALL NOT contain an entry for `alice/pdf-skill` after the command completes

#### Scenario: Remove last dependency results in empty lock
- **WHEN** `melon.yml` has only one dependency and the user removes it with `mln remove`
- **THEN** `melon.lock` SHALL reflect zero dependencies and the command SHALL exit with status 0
