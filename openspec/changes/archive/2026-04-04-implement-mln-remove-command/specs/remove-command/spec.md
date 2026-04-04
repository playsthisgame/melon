## ADDED Requirements

### Requirement: mln remove removes a named dependency from melon.yml
When the user runs `mln remove <name>`, the command SHALL remove the entry for `<name>` from `melon.yml` and write the updated file to disk.

#### Scenario: Dependency is removed from melon.yml
- **WHEN** `melon.yml` contains `alice/pdf-skill: "^1.3.0"` and the user runs `mln remove alice/pdf-skill`
- **THEN** `melon.yml` SHALL no longer contain an entry for `alice/pdf-skill` after the command completes

### Requirement: mln remove errors if the dependency does not exist
If `<name>` is not present in `melon.yml`, `mln remove` SHALL exit with a non-zero status and print an error message indicating the dependency was not found. It SHALL NOT modify `melon.yml` or `melon.lock`.

#### Scenario: Remove unknown dependency returns error
- **WHEN** `melon.yml` does not contain `alice/unknown-skill` and the user runs `mln remove alice/unknown-skill`
- **THEN** the command SHALL exit non-zero and print an error referencing the dependency name

### Requirement: mln remove unlinks the skill from all agent directories and removes the cache entry
After updating `melon.yml`, `mln remove` SHALL run the full install pipeline, which SHALL remove the agent directory symlink for the named skill from every agent directory derived from `agent_compat` (or `outputs` if declared), and SHALL delete the dependency's `.melon/` cache directory. The skill name used for the symlink is the last path segment of the dependency name.

#### Scenario: Symlink is removed from agent directories
- **WHEN** `alice/pdf-skill` is linked as `pdf-skill` in `.claude/skills/` and the user runs `mln remove alice/pdf-skill`
- **THEN** the symlink at `.claude/skills/pdf-skill` SHALL no longer exist after the command completes

#### Scenario: Cache entry is removed from .melon/

- **WHEN** the user runs `mln remove alice/pdf-skill` successfully
- **THEN** the `.melon/` cache directory for `alice/pdf-skill` SHALL no longer exist after the command completes

#### Scenario: Missing symlink is not an error
- **WHEN** the symlink for the dependency does not exist in one or more agent directories (e.g. it was manually deleted)
- **THEN** `mln remove` SHALL continue without error

### Requirement: mln remove updates melon.lock via a full install
After updating `melon.yml`, `mln remove` SHALL run the full install pipeline so that `melon.lock` is regenerated to reflect the removed dependency.

#### Scenario: Lock file no longer contains the removed dependency
- **WHEN** the user runs `mln remove alice/pdf-skill` successfully
- **THEN** `melon.lock` SHALL NOT contain an entry for `alice/pdf-skill` after the command completes

#### Scenario: Remove last dependency results in empty lock
- **WHEN** `melon.yml` has only one dependency and the user removes it with `mln remove`
- **THEN** `melon.lock` SHALL reflect zero dependencies and the command SHALL exit with status 0
