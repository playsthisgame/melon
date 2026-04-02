## ADDED Requirements

### Requirement: Skills are placed via symlinks
When `mln install` places skills into agent directories, it SHALL create a directory symlink at the target location rather than copying files. The symlink SHALL point to the corresponding entry in the `.melon/` package cache.

#### Scenario: Symlink created in agent directory after install
- **WHEN** `mln install` completes for a project with `agent_compat: [claude-code]`
- **THEN** `.claude/skills/<skill-name>` SHALL be a symlink pointing into `.melon/<encoded>@<version>/`

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
- **THEN** each derived agent skills directory SHALL contain a symlink for every installed skill

#### Scenario: --no-place skips symlink creation
- **WHEN** `mln install` is run with `--no-place`
- **THEN** no symlinks SHALL be created in any agent directory
