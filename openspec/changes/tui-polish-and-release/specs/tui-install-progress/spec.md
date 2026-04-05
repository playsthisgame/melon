## ADDED Requirements

### Requirement: mln install shows a live progress bar while fetching dependencies
When stdout is a TTY and `mln install` has one or more deps to fetch, it SHALL display a `bubbles/progress` bar that advances after each dep is fetched. The bar SHALL be labeled with the name and version of the dep currently being fetched (e.g. `Fetching alice/pdf-skill@1.3.0…`). After all deps are fetched, the bar SHALL show 100% and a summary line (e.g. `✓ 3 packages installed`) SHALL be printed.

#### Scenario: Progress bar advances per dep
- **WHEN** `mln install` fetches three deps sequentially in a TTY environment
- **THEN** the progress bar SHALL visually advance three times (once per dep) ending at 100%

#### Scenario: Current dep name is shown during fetch
- **WHEN** dep `alice/pdf-skill@1.3.0` is being fetched
- **THEN** the label beneath or beside the progress bar SHALL include `alice/pdf-skill@1.3.0`

#### Scenario: Summary line printed on completion
- **WHEN** all deps have been fetched
- **THEN** a summary line indicating the count of installed packages SHALL be printed below the progress bar

### Requirement: mln install falls back to plain text output when stdout is not a TTY
When `mln install` is run in a non-TTY context (e.g. CI pipeline, piped output), it SHALL print plain text progress lines (`fetching <name>@<version>…`) instead of rendering the bubbletea progress bar.

#### Scenario: Non-TTY context uses plain text
- **WHEN** `mln install` is run with stdout redirected to a file or pipe
- **THEN** no ANSI escape codes or progress bar rendering SHALL appear in the output; each dep SHALL produce a plain text line

### Requirement: mln install uses lipgloss to color the lock diff output
After installing, the lock diff lines SHALL be styled: added deps (`+`) in green bold, updated deps (`~`) in yellow bold, removed deps (`-`) in red bold.

#### Scenario: Added dep is shown in green
- **WHEN** a new dep appears in the lock diff
- **THEN** the `+` line SHALL be rendered in green

#### Scenario: Removed dep is shown in red
- **WHEN** a dep is absent from the new lock
- **THEN** the `-` line SHALL be rendered in red
