## MODIFIED Requirements

### Requirement: Skills are placed via symlinks
When `mln install` places skills into agent directories, it SHALL create a directory symlink at the target location rather than copying files. The symlink SHALL point to the corresponding entry in the `.melon/` package cache. The set of skills placed SHALL include all transitive dependencies, not only direct dependencies declared in `mln.yaml`. Symlinks for dependencies that were in the previous lock but are absent from the new lock SHALL be removed from all agent directories (unless `--no-place` is set). When `vendor: false`, after all symlinks are placed, melon SHALL sync `.gitignore` to include `.melon/` and all managed symlink paths.

#### Scenario: Symlink created in agent directory after install
- **WHEN** `mln install` completes for a project with `agent_compat: [claude-code]`
- **THEN** `.claude/skills/<skill-name>` SHALL be a symlink pointing into `.melon/<encoded>@<version>/` for every dep in the resolved graph (including transitive deps)

#### Scenario: Symlink target resolves to the cached skill files
- **WHEN** a symlink is created by `mln install`
- **THEN** reading files through the symlink path SHALL return the same content as reading directly from `.melon/<encoded>@<version>/`

#### Scenario: Symlink uses a relative path
- **WHEN** a symlink is created by `mln install`
- **THEN** the symlink target SHALL be a relative path so the project remains portable if moved

#### Scenario: Existing entry is replaced on re-install
- **WHEN** `mln install` is run a second time for the same dependency
- **THEN** the existing symlink or directory at the skill slot SHALL be removed and a fresh symlink created (idempotent)

#### Scenario: Multi-agent install creates symlinks for each agent
- **WHEN** `mln install` completes for a project with multiple entries in `agent_compat`
- **THEN** each derived agent skills directory SHALL contain a symlink for every installed skill in the full resolved graph

#### Scenario: --no-place skips symlink creation
- **WHEN** `mln install` is run with `--no-place`
- **THEN** no symlinks SHALL be created or removed in any agent directory

#### Scenario: Transitive dep is placed alongside direct dep
- **WHEN** a direct dep `alice/pdf-skill` has a transitive dep `bob/base-utils`
- **THEN** both `pdf-skill` and `base-utils` skill directories SHALL be symlinked in every agent directory after `mln install`

#### Scenario: Stale symlink is removed when dep is no longer in melon.yml
- **WHEN** `.claude/skills/pdf-skill` exists as a symlink from a previous install and `alice/pdf-skill` has been removed from `melon.yml`, and the user runs `mln install`
- **THEN** `.claude/skills/pdf-skill` SHALL no longer exist after the command completes

#### Scenario: gitignore is synced after placement when vendor is false
- **WHEN** `vendor: false` and `mln install` completes successfully
- **THEN** `.gitignore` SHALL contain `.melon/` and every managed symlink path that was placed
