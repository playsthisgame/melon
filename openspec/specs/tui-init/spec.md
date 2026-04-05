### Requirement: mln init uses a bubbletea multi-step form in interactive mode
When `mln init` is run interactively (without `--yes`), it SHALL present a bubbletea-driven multi-step form: a text input for project name (with cwd as default), a single-select list for package type, a text input for description, and a scrollable multi-select list for agent_compat. The user navigates with arrow keys, toggles selections with space, and confirms each step with enter.

#### Scenario: Full interactive form collects all fields
- **WHEN** the user runs `mln init` interactively and fills in each step
- **THEN** the generated `melon.yml` SHALL contain the entered name, type, description, and agent_compat values

#### Scenario: Arrow keys navigate the agent list
- **WHEN** the agent_compat multi-select step is displayed
- **THEN** pressing ↑/↓ SHALL move the cursor through the agent list without scrolling past the ends

#### Scenario: Space toggles agent selection
- **WHEN** the cursor is on an agent in the multi-select list
- **THEN** pressing space SHALL toggle that agent's selected state, and the selection SHALL be visually indicated (e.g. a checkmark or color highlight)

#### Scenario: Enter confirms each step and advances the form
- **WHEN** the user presses enter on any step
- **THEN** the form SHALL advance to the next step, preserving the entered value

#### Scenario: --yes flag bypasses the bubbletea form
- **WHEN** `mln init --yes` is run
- **THEN** all defaults SHALL be accepted without launching the bubbletea model

### Requirement: mln init type selection uses a single-select list
The package type step SHALL display all valid types (skill, agent, workflow, persona, memory) as a scrollable list. The user navigates with arrow keys and confirms with enter. The default selection SHALL be `agent`.

#### Scenario: Type list shows all valid options
- **WHEN** the type selection step is displayed
- **THEN** all five types (agent, skill, workflow, persona, memory) SHALL be visible as selectable options

#### Scenario: Default type is pre-selected
- **WHEN** the type selection step is first displayed
- **THEN** `agent` SHALL be the initially highlighted option
