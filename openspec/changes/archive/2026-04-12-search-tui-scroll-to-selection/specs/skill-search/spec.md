## MODIFIED Requirements

### Requirement: Interactive result selection in TTY mode

When stdout is a TTY, search results SHALL be presented as an interactive multi-select list using the existing bubbletea TUI infrastructure. The user navigates with arrow keys, toggles items with space, and presses Enter to confirm selections. After the interactive list, the CLI SHALL display a confirmation prompt in the form `Install N skill(s)? [Y/n]` where yes is the default — pressing Enter without input proceeds with the install. The list viewport SHALL be sized to fit the current terminal height so the cursor is always visible without the user needing to resize the terminal.

#### Scenario: User selects a skill and confirms with Enter

- **WHEN** results are shown in the interactive list, the user selects one or more items, and presses Enter to confirm
- **THEN** the CLI shows the `Install N skill(s)? [Y/n]` prompt; pressing Enter (or `y`) installs the selected skills

#### Scenario: User exits without selecting

- **WHEN** the user presses Escape or Ctrl+C during the interactive list
- **THEN** the CLI exits cleanly with code 0 and no action is taken

#### Scenario: User declines at the confirmation prompt

- **WHEN** the user has selected skills in the interactive list and types `n` at the `[Y/n]` prompt
- **THEN** no skills are installed and the command exits cleanly

#### Scenario: Non-TTY mode (piped or CI output)

- **WHEN** stdout is not a TTY (e.g. piped to another command or run in CI)
- **THEN** results are printed as plain text, one per line, with no interactive prompt

#### Scenario: Short terminal does not hide the cursor

- **WHEN** the terminal height is smaller than the full result list height
- **THEN** the list is clamped to fit the terminal and the cursor at index 0 is visible on first render
