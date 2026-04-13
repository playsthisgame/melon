## Context

`newSearchModel` in `internal/cli/search_model.go` sets the bubbletea `list.Model` height to `min(len(results)*2, 20)`. `newRemoveModel` in `internal/cli/remove_model.go` does the same with `min(len(skills), 20)`. Both are static values that ignore actual terminal dimensions. When the terminal is shorter than the computed height, the list renders taller than the screen. Bubbletea's list component scrolls its viewport to show the last item, which places the initial cursor (index 0) above the visible area. The user sees the bottom of the list but cannot navigate from there without first understanding why.

The bubbletea `list.Model` has built-in viewport/pagination support — it will clamp the visible window to the set height. The fix is simply to set that height to something that fits in the terminal.

Bubbletea delivers terminal dimensions via `tea.WindowSizeMsg`, dispatched at startup and again on resize.

## Goals / Non-Goals

**Goals:**
- Cap the list height to fit the terminal on initial render for both search and remove TUIs
- Keep the cursor (index 0) visible when the list first appears
- Adapt the list height if the user resizes the terminal mid-session

**Non-Goals:**
- Changing the visual design of each list row
- Adding a minimum terminal size requirement or resize prompt
- Modifying any TUI other than the search and remove models

## Decisions

**Use `tea.WindowSizeMsg` to set the list height dynamically.**
Bubbletea sends this message before the first render and on every terminal resize. Handling it in `Update` is the idiomatic way to make a TUI responsive — no polling, no `golang.org/x/term` calls at init time. The initial height passed to `list.New` can stay as a reasonable fallback (e.g. 20) for non-TTY or test contexts where `WindowSizeMsg` may never arrive.

**Reserve rows for chrome above/below the list.**
Both `View` methods render a title line, the list itself, and a hint line, each followed by `\n`. Total overhead is ~3 rows for both models. The list height should be `termHeight - 3` (clamped to at least 2). A shared named constant `listReservedRows = 3` in a common file (or duplicated in each model file) makes this obvious.

**Row height differs between models — handle independently.**
Search items are 2 lines each; remove items are 1 line each. The max-items clamp is therefore `len(items)*2` for search and `len(items)` for remove. Each model handles its own `WindowSizeMsg` with the appropriate per-row height.

**Do not query terminal size at model construction time.**
`golang.org/x/term` could give us the size synchronously, but it adds a dependency and breaks in test contexts. The `WindowSizeMsg` approach is zero-dependency and already used by bubbletea apps everywhere.

## Risks / Trade-offs

- [Risk] `WindowSizeMsg` arrives slightly after the first render, causing a single frame at the wrong height → Mitigation: imperceptible in practice; bubbletea redraws immediately on message receipt. The fallback height of 20 means the worst case is one frame that is slightly too tall, not the current broken-cursor state.
- [Risk] Very tall result sets on a short terminal will still show a paginated list → This is correct behavior; the list scrolls as expected once the height is constrained.

## Migration Plan

No migration needed. Pure UX fix, no persisted state changed.
