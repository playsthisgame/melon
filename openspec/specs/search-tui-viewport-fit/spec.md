## Requirements

### Requirement: Search and remove TUI list height fits the terminal
The search result list and the remove skill list SHALL each size their visible height to fit within the current terminal height, reserving rows for the title line, hint bar, and any padding. Neither list height SHALL exceed `terminalHeight - listReservedRows`.

#### Scenario: Terminal shorter than the default list height
- **WHEN** the terminal height is less than the number of result rows plus reserved rows
- **THEN** the list height is clamped so all visible rows fit on screen without overflow

#### Scenario: Terminal taller than all results
- **WHEN** the terminal height is larger than all result rows plus reserved rows
- **THEN** the list height equals the total result rows (no unnecessary padding)

#### Scenario: Terminal is resized while the TUI is open
- **WHEN** the user resizes the terminal while the search list is displayed
- **THEN** the list height adapts to the new terminal height and the cursor remains visible

### Requirement: Cursor is visible on initial render
On first render the selection cursor SHALL be within the visible viewport — the list SHALL NOT open with the viewport scrolled to the bottom while the cursor is above the visible area.

#### Scenario: Initial render with more results than visible rows
- **WHEN** there are more search results than fit on screen
- **THEN** the first result (index 0) and its cursor are visible when the list first appears
