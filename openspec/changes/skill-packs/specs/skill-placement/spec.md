## ADDED Requirements

### Requirement: Placement skips pure packs
The placer SHALL skip any dependency whose lock or resolved type is `pack` and whose cached directory does not contain its entrypoint file. The same predicate SHALL be used by placement, unplacement, `.gitignore` symlink management, and `melon list --check`, so they never disagree about which dependencies have symlinks.

#### Scenario: Pure pack skipped during place
- **WHEN** a resolved dependency has `type: pack` and no `SKILL.md` in `.melon/<dep@version>/`
- **THEN** no symlink SHALL be created for it in any harness directory

#### Scenario: Pure pack has no gitignore entry
- **WHEN** `vendor: false` and the project depends on a pure pack
- **THEN** `.gitignore` SHALL NOT contain a symlink entry for the pack

#### Scenario: Hybrid pack placed
- **WHEN** a resolved dependency has `type: pack` and its cache directory contains `SKILL.md`
- **THEN** it SHALL be symlinked into each harness directory like any skill
