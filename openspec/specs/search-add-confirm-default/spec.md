# Search Add Confirm Default Spec

## ADDED Requirements

### Requirement: Install confirmation prompt defaults to yes

When `mln search` presents a confirmation prompt before installing selected skills, the prompt SHALL default to yes (`[Y/n]`). Pressing Enter without typing a value SHALL be treated as confirmation to proceed with the install.

#### Scenario: User presses Enter without input — install proceeds

- **WHEN** the install confirmation prompt is shown and the user presses Enter without typing anything
- **THEN** the selected skills SHALL be installed, as if the user had typed `y`

#### Scenario: User types `n` — install is cancelled

- **WHEN** the install confirmation prompt is shown and the user types `n` or `no`
- **THEN** no skills SHALL be installed and the command SHALL exit cleanly

#### Scenario: Prompt displays `[Y/n]` to indicate the default

- **WHEN** the install confirmation prompt is rendered
- **THEN** it SHALL display `[Y/n]` (uppercase Y, lowercase n) to signal that yes is the default

#### Scenario: User types `y` explicitly — install proceeds

- **WHEN** the install confirmation prompt is shown and the user types `y` or `yes`
- **THEN** the selected skills SHALL be installed
