## ADDED Requirements

### Requirement: melon update shows a spinner while resolving versions
When `melon update` is run in a TTY, it SHALL display a `bubbles/spinner` with the message `Resolving updates…` while version resolution network calls are in progress. The spinner SHALL stop and clear before any subsequent output is printed.

#### Scenario: Spinner shown during version resolution
- **WHEN** `melon update` is run in a TTY and version resolution calls are in progress
- **THEN** a spinning animation with the message `Resolving updates…` SHALL be visible until all selected deps are resolved

#### Scenario: Spinner clears before result output
- **WHEN** version resolution completes
- **THEN** the spinner line SHALL be cleared and update result output SHALL follow without overlapping

#### Scenario: No spinner in non-TTY context
- **WHEN** `melon update <dep>` is run with stdout not a TTY
- **THEN** no spinner or ANSI codes SHALL be emitted; plain text output only
