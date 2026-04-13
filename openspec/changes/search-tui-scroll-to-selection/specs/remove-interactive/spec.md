## MODIFIED Requirements

### Requirement: mln remove with no arguments launches an interactive multi-select in a TTY
When `mln remove` is invoked with no arguments in a TTY, the system SHALL read all skills from `melon.yaml` and present a bubbletea-driven multi-select list. The user navigates with arrow keys, toggles selections with space, and confirms with enter. After confirmation, a prompt SHALL ask the user to confirm the destructive action before any skills are removed. The list viewport SHALL be sized to fit the current terminal height so the cursor is always visible without the user needing to resize the terminal.

#### Scenario: Skills are listed for selection
- **WHEN** `mln remove` is run with no arguments in a TTY and `melon.yaml` contains one or more dependencies
- **THEN** a multi-select list SHALL be displayed showing each skill name and version, one per row

#### Scenario: Space toggles a skill's selected state
- **WHEN** the cursor is on a skill entry and the user presses space
- **THEN** the item SHALL toggle between selected (`[✓]`) and deselected (`[ ]`)

#### Scenario: Enter with selections triggers a confirmation prompt
- **WHEN** the user presses enter with one or more skills selected
- **THEN** the TUI SHALL exit and display `Remove N skill(s)? [y/N]:` listing the selected skill names

#### Scenario: User confirms the removal prompt
- **WHEN** the user types `y` or `yes` at the confirmation prompt
- **THEN** each selected skill SHALL be removed in sequence using the standard remove pipeline

#### Scenario: User declines the removal prompt
- **WHEN** the user types anything other than `y`/`yes` at the confirmation prompt
- **THEN** no skills SHALL be removed and the command SHALL exit cleanly with status 0

#### Scenario: Enter with no selections is a no-op
- **WHEN** the user presses enter with no skills selected
- **THEN** nothing SHALL be removed and the command SHALL exit cleanly with status 0

#### Scenario: Escape or Ctrl+C cancels without removing
- **WHEN** the user presses esc or ctrl+c during selection
- **THEN** nothing SHALL be removed and the command SHALL exit cleanly with status 0

#### Scenario: melon.yaml has no dependencies
- **WHEN** `mln remove` is run with no arguments and `melon.yaml` contains zero dependencies
- **THEN** the command SHALL print "No skills in melon.yaml." and exit with status 0 without launching the TUI

#### Scenario: Short terminal does not hide the cursor
- **WHEN** the terminal height is smaller than the full skill list height
- **THEN** the list is clamped to fit the terminal and the cursor at index 0 is visible on first render
