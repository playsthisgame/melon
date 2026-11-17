## Requirements

### Requirement: Search, remove and update TUI list height fits the terminal
The search result list, the remove skill list and the update skill list SHALL each size their visible height to fit within the current terminal height, reserving rows for the title line, hint bar, and any padding. Neither list height SHALL exceed `terminalHeight - listReservedRows`.

#### Scenario: Terminal shorter than the default list height
- **WHEN** the terminal height is less than the number of result rows plus reserved rows
- **THEN** the list height is clamped so all visible rows fit on screen without overflow

#### Scenario: Terminal taller than all results
- **WHEN** the terminal height is larger than all result rows plus reserved rows
- **THEN** the list height equals the total result rows (no unnecessary padding)

#### Scenario: Terminal is resized while the TUI is open
- **WHEN** the user resizes the terminal while the search list is displayed
- **THEN** the list height adapts to the new terminal height, growing back up to its content-capped height when the terminal gets taller, and the cursor remains visible

#### Scenario: Rendered output never exceeds the terminal
- **WHEN** the list is displayed in a terminal at least tall enough for one item, the pagination row and the reserved rows
- **THEN** the rendered view (title, list rows including pagination, and hint bar) SHALL be no taller than the terminal, because bubbletea drops lines from the top of a taller view and the cursor is near the top

#### Scenario: List rows render at their declared height
- **WHEN** a list delegate declares a row height of N lines
- **THEN** each rendered row SHALL be exactly N lines, with no trailing newline, so the list's height math matches what is drawn

### Requirement: Cursor is visible on initial render
On first render the selection cursor SHALL be within the visible viewport — the list SHALL NOT open with the viewport scrolled to the bottom while the cursor is above the visible area.

#### Scenario: Initial render with more results than visible rows
- **WHEN** there are more search results than fit on screen
- **THEN** the first result (index 0) and its cursor are visible when the list first appears
