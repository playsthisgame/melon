### Requirement: --global selects the user-global install
Every melon command SHALL accept `--global` / `-g`. With it, melon SHALL read and write `~/.config/melon/melon.yaml` and `~/.config/melon/melon.lock` instead of the project's files, cache fetched skills in `~/.melon/store/`, and place skills into each harness's global skills directory from the agent conventions table. `--global` and `--dir` SHALL NOT be combined.

#### Scenario: Global install places skills into harness global directories
- **WHEN** the global manifest lists `harnesses: [claude-code]` and depends on `alice/skills/pdf`, and the user runs `melon install --global`
- **THEN** `~/.claude/skills/pdf` SHALL be a symlink to `~/.melon/store/github.com-alice-skills-pdf@<version>/` and `~/.config/melon/melon.lock` SHALL record the dep

#### Scenario: Global symlinks are absolute
- **WHEN** a skill is placed globally
- **THEN** the symlink target SHALL be the absolute store path

#### Scenario: Global commands require the global manifest
- **WHEN** the user runs any `--global` command other than `init` and `~/.config/melon/melon.yaml` does not exist
- **THEN** melon SHALL exit with an error telling the user to run `melon init --global`

#### Scenario: --global with --dir is rejected
- **WHEN** the user passes both `--global` and `--dir`
- **THEN** melon SHALL exit with an error

### Requirement: melon init --global creates the global manifest
`melon init --global` SHALL ask which harnesses global skills target and write `~/.config/melon/melon.yaml` with that `harnesses` list, then create `~/.melon/store/`. Every global install SHALL target every harness in that list. When the list is empty, skills SHALL be placed in `~/.agents/skills/`.

#### Scenario: Harnesses chosen at init are written to the global manifest
- **WHEN** the user runs `melon init --global` and selects `claude-code` and `opencode`
- **THEN** the global manifest SHALL contain `harnesses: [claude-code, opencode]`

### Requirement: Unmanaged global skills are never replaced
When a global harness directory already contains an entry with a skill's name that is not a symlink into `~/.melon/store/`, melon SHALL leave it untouched, continue placing the remaining skills, and after placement list every skipped path. Global unplace (`remove`, install pruning, `clean`) SHALL only delete symlinks that point into `~/.melon/store/`.

#### Scenario: Existing user skill is skipped and reported
- **WHEN** `~/.claude/skills/xlsx/` is a directory the user created and `melon install --global` installs `alice/skills/xlsx`
- **THEN** the directory SHALL be unchanged, other skills SHALL still be placed, and the output SHALL list `~/.claude/skills/xlsx` as skipped

#### Scenario: Remove leaves the user's skill alone
- **WHEN** the user runs `melon remove --global alice/skills/xlsx` after it was skipped
- **THEN** `~/.claude/skills/xlsx/` SHALL still exist

### Requirement: Global policy and gitignore
`policy.allowed_sources` in the global manifest SHALL be enforced on the full resolved graph, as for projects. Melon SHALL never create or modify a `.gitignore` for global installs, regardless of `vendor`.

#### Scenario: Blocked source is not fetched globally
- **WHEN** the global manifest allows only `github.com/trusted/*` and depends on `github.com/evil/skills/x`
- **THEN** `melon install --global` SHALL fail before fetching it

### Requirement: Search offers a global install
After the user confirms installing skills from `melon search` without `--global`, melon SHALL ask whether to install them globally (default no). Answering yes SHALL add them to the global manifest and run the global install.

#### Scenario: Answering yes installs globally
- **WHEN** the user selects a skill in `melon search`, confirms, and answers `y` to "Install globally"
- **THEN** the skill SHALL be added to `~/.config/melon/melon.yaml`, not the project's melon.yaml
