### Requirement: Multi-select agent prompt in mln init
`mln init` SHALL present an interactive multi-select checkbox list of all known agents (from `agents.KnownAgents()`) so that users can select one or more target agents. The selected agents SHALL be written to `mln.yaml` as a `agent_compat` list.

#### Scenario: User selects multiple agents
- **WHEN** `mln init` is run interactively and the user checks both `claude-code` and `cursor`
- **THEN** the generated `mln.yaml` contains `agent_compat: [claude-code, cursor]`

#### Scenario: User selects a single agent
- **WHEN** `mln init` is run interactively and the user checks only `claude-code`
- **THEN** the generated `mln.yaml` contains `agent_compat: [claude-code]`

#### Scenario: User selects no agents
- **WHEN** `mln init` is run interactively and the user checks no agents
- **THEN** the generated `mln.yaml` contains `agent_compat: []`

#### Scenario: --yes flag skips prompt with default
- **WHEN** `mln init --yes` is run
- **THEN** the generated `mln.yaml` contains `agent_compat: [claude-code]` without prompting

### Requirement: All known agents are shown as options
`mln init` SHALL display every agent returned by `agents.KnownAgents()` as a selectable option. The list SHALL be sorted alphabetically.

#### Scenario: Full agent list is presented
- **WHEN** the multi-select prompt is shown
- **THEN** all 10 known agents (amp, cline, claude-code, codex, cursor, gemini-cli, github-copilot, opencode, roo, windsurf) appear as options
