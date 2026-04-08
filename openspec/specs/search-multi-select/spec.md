## ADDED Requirements

### Requirement: melon search uses multi-select in interactive mode
When `mln search <term>` is run in a TTY and results are found, it SHALL present a bubbletea-driven multi-select list. The user navigates with arrow keys, toggles selections with space, and confirms with enter. After confirmation, all selected skills SHALL be installed.

#### Scenario: Space toggles a result's selected state
- **WHEN** the cursor is on a search result and the user presses space
- **THEN** the item SHALL be toggled selected/deselected, visually indicated with a checkbox (`[✓]` or `[ ]`)

#### Scenario: Enter confirms and installs all selected skills
- **WHEN** the user presses enter with one or more skills selected
- **THEN** the TUI SHALL exit and each selected skill SHALL be installed in sequence

#### Scenario: Enter with no selections is a no-op
- **WHEN** the user presses enter with no skills selected
- **THEN** nothing SHALL be installed and the command SHALL exit cleanly

#### Scenario: Escape or Ctrl+C cancels without installing
- **WHEN** the user presses esc or ctrl+c
- **THEN** nothing SHALL be installed and the command SHALL exit cleanly

#### Scenario: Multiple skills can be selected in one pass
- **WHEN** the user toggles multiple results and presses enter
- **THEN** all toggled skills SHALL be installed, one after another

### Requirement: melon search shows a batch install prompt before installing
Before installing selected skills, the command SHALL print the list of selected skills and prompt the user to confirm with `Install N skill(s)? [y/N]`.

#### Scenario: User confirms the batch install prompt
- **WHEN** the user types `y` or `yes` at the install prompt
- **THEN** all selected skills SHALL be installed

#### Scenario: User declines the batch install prompt
- **WHEN** the user types anything other than `y`/`yes` at the install prompt
- **THEN** no skills SHALL be installed and the command SHALL exit cleanly
