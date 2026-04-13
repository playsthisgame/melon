## Why

When the terminal is shorter than the search or remove result list, the TUI list height is set larger than the visible terminal area. Bubbletea renders the list and the viewport scrolls to the bottom, leaving the cursor (selection arrow) above what is visible on screen — the user cannot interact with their selection without resizing the terminal. This affects both `melon search` and `melon remove`, which share the same static-height pattern.

## What Changes

- The search TUI list height is capped to the actual terminal height minus reserved rows (title, hint bar, padding), not just `min(results*2, 20)`
- The remove TUI list height is capped the same way, not just `min(skills, 20)`
- On startup, `tea.WindowSizeMsg` is handled in both models so the list resizes correctly even if the terminal is resized after launch
- The initial selected item (index 0) is always within the visible viewport on render for both TUIs

## Capabilities

### New Capabilities

- `search-tui-viewport-fit`: Both the search and remove result lists SHALL size themselves to fit within the current terminal height, keeping the cursor visible without requiring the user to resize the terminal.

### Modified Capabilities

- `skill-search`: The interactive search TUI requirement gains a constraint that the list viewport must not exceed terminal height and must keep the cursor visible on initial render.
- `remove-interactive`: The interactive remove TUI requirement gains the same viewport-fit constraint.

## Impact

- `internal/cli/search_model.go`: `newSearchModel` height calculation, `Update` to handle `tea.WindowSizeMsg`
- `internal/cli/remove_model.go`: `newRemoveModel` height calculation, `Update` to handle `tea.WindowSizeMsg`
- No manifest, lock file, resolver, or fetcher changes
