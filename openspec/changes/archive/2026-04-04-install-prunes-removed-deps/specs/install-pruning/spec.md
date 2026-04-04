## ADDED Requirements

### Requirement: mln install removes the cache entry for each dependency no longer in melon.yml
After computing the new lock, `mln install` SHALL delete the `.melon/<encoded>@<version>/` directory for every dependency that was present in the previous `melon.lock` but is absent from the newly resolved lock.

#### Scenario: Cache entry is deleted after dependency is removed from melon.yml
- **WHEN** `melon.lock` previously contained `alice/pdf-skill@1.3.0` and `alice/pdf-skill` has been removed from `melon.yml`, and the user runs `mln install`
- **THEN** `.melon/github.com-alice-pdf-skill@1.3.0/` (or the equivalent encoded path) SHALL no longer exist after the command completes

#### Scenario: Cache entries for remaining dependencies are untouched
- **WHEN** `melon.yml` still contains `bob/base-utils` after removing `alice/pdf-skill`, and the user runs `mln install`
- **THEN** `.melon/` SHALL still contain the cache directory for `bob/base-utils`

#### Scenario: Already-absent cache entry does not cause an error
- **WHEN** the `.melon/` directory for a removed dep was already deleted manually before `mln install` is run
- **THEN** `mln install` SHALL complete successfully without error

### Requirement: mln install removes agent symlinks for each dependency no longer in melon.yml
After computing the new lock, `mln install` SHALL remove the agent directory symlink for every dependency that was in the previous lock but is absent from the new lock. This applies to every agent directory derived from `agent_compat` (or `outputs` if declared).

#### Scenario: Agent symlink is removed after dependency is removed from melon.yml
- **WHEN** `.claude/skills/pdf-skill` is a symlink and `alice/pdf-skill` is removed from `melon.yml`, and the user runs `mln install`
- **THEN** `.claude/skills/pdf-skill` SHALL no longer exist after the command completes

#### Scenario: Symlink removal is skipped when --no-place is set
- **WHEN** the user runs `mln install --no-place` after removing a dependency from `melon.yml`
- **THEN** the agent symlink for the removed dependency SHALL NOT be deleted (placement operations are skipped entirely)

#### Scenario: Missing symlink does not cause an error during pruning
- **WHEN** the agent symlink for a removed dep was already deleted manually before `mln install` is run
- **THEN** `mln install` SHALL complete successfully without error
