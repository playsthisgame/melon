## Requirements

### Requirement: mln remove with no arguments launches an interactive multi-select in a TTY
When `mln remove` is invoked with no arguments in a TTY, the system SHALL read all skills from `melon.yaml` and present a bubbletea-driven multi-select list. The user navigates with arrow keys, toggles selections with space, and confirms with enter. After confirmation, a prompt SHALL ask the user to confirm the destructive action before any skills are removed.

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

### Requirement: mln remove with no arguments in a non-TTY exits with an error
When `mln remove` is invoked with no arguments outside of a TTY (e.g., in CI or piped input), the system SHALL exit with a non-zero status and print a message instructing the user to provide a skill name.

#### Scenario: Non-TTY invocation without arguments
- **WHEN** `mln remove` is run with no arguments and stdout is not a TTY
- **THEN** the command SHALL exit non-zero and print an error such as "remove: skill name required (non-interactive mode)"
