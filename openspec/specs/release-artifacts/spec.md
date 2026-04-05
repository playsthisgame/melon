### Requirement: A README.md is present with install and usage instructions
The repository SHALL contain a `README.md` that includes: `go install` instructions, a quick-start example, the full command reference for `init`/`install`/`add`/`remove`, the `agent_compat` convention table (all 10 known agents and their placement paths), and a brief explanation of how melon differs from npx-based skill installers.

#### Scenario: README covers all four commands
- **WHEN** a user reads the README
- **THEN** they SHALL find documented usage examples for `mln init`, `mln install`, `mln add`, and `mln remove`

#### Scenario: README includes agent_compat table
- **WHEN** a user reads the README
- **THEN** they SHALL find a table mapping each known agent to its project-scoped skill directory path

### Requirement: A .gitignore excludes .melon/ and includes agent dirs and melon.lock
The repository SHALL contain a `.gitignore` that ignores `.melon/` and does NOT ignore agent skill directories (`.claude/skills/`, `.agents/skills/`, etc.) or `melon.lock`.

#### Scenario: .melon/ is git-ignored
- **WHEN** `.gitignore` is present and `git status` is run after `mln install`
- **THEN** files under `.melon/` SHALL NOT appear as untracked

#### Scenario: melon.lock is not git-ignored
- **WHEN** `.gitignore` is present
- **THEN** `melon.lock` SHALL be tracked by git

### Requirement: A .goreleaser.yaml and GitHub Actions release workflow are present
The repository SHALL contain a `.goreleaser.yaml` that produces prebuilt binaries for macOS arm64, macOS amd64, Linux amd64, and Windows amd64, and a `.github/workflows/release.yml` that triggers GoReleaser on git tag push (tags matching `v*`).

#### Scenario: Release workflow triggers on tag push
- **WHEN** a git tag matching `v*` is pushed
- **THEN** the GitHub Actions workflow SHALL trigger and run GoReleaser
