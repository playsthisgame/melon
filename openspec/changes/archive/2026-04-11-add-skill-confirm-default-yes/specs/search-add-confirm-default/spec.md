## ADDED Requirements

### Requirement: Install confirmation after search selection defaults to yes
After the user selects one or more skills from interactive search results, the CLI SHALL present a confirmation prompt with `[Y/n]` where pressing Enter (empty input) proceeds with installation.

#### Scenario: User presses Enter to confirm
- **WHEN** the user selects skills from the search TUI and presses Enter at the confirmation prompt without typing anything
- **THEN** the selected skills SHALL be installed as if the user had typed `y`

#### Scenario: User types y to confirm
- **WHEN** the user selects skills from the search TUI and types `y` or `Y` at the confirmation prompt
- **THEN** the selected skills SHALL be installed

#### Scenario: User types n to cancel
- **WHEN** the user selects skills from the search TUI and types `n` or `N` at the confirmation prompt
- **THEN** no skills SHALL be installed and the CLI exits with code 0

#### Scenario: Prompt text displays [Y/n]
- **WHEN** the confirmation prompt is displayed
- **THEN** the prompt SHALL read `Install <N> skill(s)? [Y/n]: ` with uppercase Y indicating the default
