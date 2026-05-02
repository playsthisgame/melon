### Requirement: vendor field controls gitignore management
The `melon.yaml` manifest SHALL support an optional `vendor` boolean field. When `vendor` is absent or `true`, melon SHALL NOT modify `.gitignore`. When `vendor` is `false`, melon SHALL automatically maintain `.gitignore` entries for `.melon/` and all managed symlink paths across `install`, `add`, and `remove` commands.

#### Scenario: vendor defaults to true when field is absent
- **WHEN** `melon.yaml` does not contain a `vendor` field
- **THEN** melon SHALL behave as if `vendor: true` and SHALL NOT touch `.gitignore`

#### Scenario: vendor: true suppresses all gitignore writes
- **WHEN** `melon.yaml` contains `vendor: true` and the user runs `mln install`
- **THEN** `.gitignore` SHALL NOT be created or modified

#### Scenario: vendor: false triggers gitignore sync on install
- **WHEN** `melon.yaml` contains `vendor: false` and the user runs `mln install`
- **THEN** `.gitignore` SHALL contain entries for `.melon/` and every managed symlink path

### Requirement: melon creates .gitignore when vendor is false and file does not exist
When `vendor: false` and no `.gitignore` exists in the project root, melon SHALL create `.gitignore` with the required entries rather than erroring.

#### Scenario: .gitignore is created when missing
- **WHEN** `vendor: false`, no `.gitignore` exists, and the user runs `mln install`
- **THEN** a `.gitignore` file SHALL be created containing `.melon/` and the managed symlink paths

### Requirement: gitignore entries are idempotent
Melon SHALL NOT add duplicate entries to `.gitignore`. If an entry already exists (added by melon or the user manually), melon SHALL skip it.

#### Scenario: Re-running install does not duplicate entries
- **WHEN** `vendor: false` and `mln install` has already been run once
- **THEN** running `mln install` again SHALL NOT add duplicate lines to `.gitignore`

### Requirement: melon labels its managed section in .gitignore
When melon first writes entries to `.gitignore`, it SHALL prepend a comment `# melon managed — do not edit this block` before the entries so users can identify them.

#### Scenario: Comment header appears before melon entries
- **WHEN** `vendor: false` and melon writes entries to `.gitignore` for the first time
- **THEN** the entries SHALL be preceded by the comment `# melon managed — do not edit this block`

### Requirement: melon init prompts for vendoring preference with a clear default and description
During `melon init`, the user SHALL be asked whether they want to vendor dependencies. The prompt SHALL display a one-sentence description of each option and make the default (yes / vendor: true) visually obvious. The prompt text SHALL be:

```text
Vendor skills in git? Skills will be committed to your repo; disable to auto-manage .gitignore instead. [Y/n]
```

The uppercase `Y` and lowercase `n` in `[Y/n]` SHALL signal that yes is the default. When the user answers no, `vendor: false` SHALL be written to `melon.yaml`.

#### Scenario: User accepts default — vendor: true omitted from yaml
- **WHEN** the user runs `mln init` and accepts the default vendoring prompt
- **THEN** `melon.yaml` SHALL NOT contain a `vendor` field (implicit true)

#### Scenario: User opts out — vendor: false written to yaml
- **WHEN** the user runs `mln init` and answers no to the vendoring prompt
- **THEN** `melon.yaml` SHALL contain `vendor: false`

#### Scenario: --yes flag defaults to vendor: true
- **WHEN** `mln init --yes` is run
- **THEN** `melon.yaml` SHALL NOT contain a `vendor` field

#### Scenario: Prompt displays uppercase Y to signal default

- **WHEN** the vendoring prompt is shown in TTY or non-TTY mode
- **THEN** the prompt SHALL include `[Y/n]` with uppercase Y indicating the default is yes

### Requirement: melon hints about git rm --cached after switching to vendor: false
After running `mln install` with `vendor: false` for the first time on a project that was previously vendored, melon SHALL print a hint advising the user to run `git rm --cached` for the newly-ignored paths.

#### Scenario: Hint printed when entries are newly added to gitignore
- **WHEN** `vendor: false` and melon adds entries to `.gitignore` that were not present before
- **THEN** melon SHALL print a message like: `Tip: run \`git rm --cached <path>\` to stop tracking previously committed files`
