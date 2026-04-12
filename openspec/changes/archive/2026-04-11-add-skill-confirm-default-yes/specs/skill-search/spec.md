## MODIFIED Requirements

### Requirement: Interactive result selection in TTY mode
When stdout is a TTY, search results SHALL be presented as an interactive single-select list using the existing bubbletea TUI infrastructure. The user navigates with arrow keys and presses Enter to select a skill.

#### Scenario: User selects a skill from results
- **WHEN** results are shown in the interactive list and the user presses Enter on an item
- **THEN** the CLI presents the selected skill(s) in a confirmation prompt with `[Y/n]` (defaulting to yes) and offers to run `mln add` for each

#### Scenario: User confirms with Enter at the install prompt
- **WHEN** the confirmation prompt is shown after skill selection and the user presses Enter without typing
- **THEN** the CLI SHALL install all selected skills as if the user had typed `y`

#### Scenario: User exits without selecting
- **WHEN** the user presses Escape or Ctrl+C during the interactive list
- **THEN** the CLI exits cleanly with code 0 and no action is taken

#### Scenario: Non-TTY mode (piped or CI output)
- **WHEN** stdout is not a TTY (e.g. piped to another command or run in CI)
- **THEN** results are printed as plain text, one per line, with no interactive prompt
