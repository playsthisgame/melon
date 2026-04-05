### Requirement: mln add shows a spinner while resolving the latest version
When `mln add` is run without an explicit version constraint and stdout is a TTY, it SHALL display a `bubbles/spinner` with the message `Resolving latest version of <dep>…` while the `LatestTag` network call is in progress. The spinner SHALL stop and clear before any subsequent output is printed.

#### Scenario: Spinner shown during LatestTag call
- **WHEN** `mln add alice/pdf-skill` is run in a TTY and the LatestTag call takes more than a moment
- **THEN** a spinning animation with the dep name SHALL be visible until the version is resolved

#### Scenario: Spinner clears before install output
- **WHEN** the LatestTag call completes
- **THEN** the spinner line SHALL be cleared and normal install output SHALL follow without overlapping

#### Scenario: No spinner in non-TTY context
- **WHEN** `mln add` is run with stdout not a TTY
- **THEN** no spinner or ANSI codes SHALL be emitted; plain text output only

### Requirement: mln remove shows a spinner during the install pipeline
When `mln remove` is run in a TTY, it SHALL display a `bubbles/spinner` with the message `Updating…` while the `runInstall` pipeline executes. The spinner SHALL stop before the lock diff output is printed.

#### Scenario: Spinner shown during remove install step
- **WHEN** `mln remove alice/pdf-skill` is run in a TTY
- **THEN** a spinner SHALL be visible while install runs

#### Scenario: Spinner clears before diff output
- **WHEN** the install pipeline completes during remove
- **THEN** the spinner SHALL be cleared before any diff or summary lines are printed
