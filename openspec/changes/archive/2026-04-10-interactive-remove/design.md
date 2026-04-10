## Context

`mln remove <name>` removes a skill from `melon.yaml` and runs the install pipeline to clean up symlinks and the cache. Today it requires an explicit skill name argument — omitting it causes cobra to return an argument error.

The search command already ships a bubbletea multi-select TUI (`searchModel` / `searchMultiSelectDelegate` in `search_model.go`) backed by a confirm prompt (`offerAddMany`). The remove interactive mode should follow the same pattern.

## Goals / Non-Goals

**Goals:**
- When `mln remove` is invoked with no arguments in a TTY, show a multi-select list of all skills in `melon.yaml`
- User toggles skills with space, confirms with enter; selected skills are removed in sequence
- A confirmation prompt (`Remove N skill(s)? [y/N]`) is shown before any destructive action
- Non-TTY (no args) emits a clear error message instead of launching TUI
- All existing `mln remove <name>` behavior is preserved

**Non-Goals:**
- Fuzzy-search/filtering within the selector (the list is bounded by what's in `melon.yaml`)
- Batch remove via `mln remove foo bar` (not requested; still errors with arg-count validation if desired later)
- UI changes to `mln search`

## Decisions

### Reuse `searchModel` vs. a new model

The `searchModel` is tightly coupled to `searchResultItem` (path, author, description, featured star). Installed skills in `melon.yaml` are just names + version strings — a different shape.

**Decision:** Create a new, simpler `removeModel` (and matching delegate) that renders a single-line `[✓] <skill-name>  <version>` row. This avoids coupling remove to search internals and keeps each model independently testable. The bubbletea program plumbing (`tea.NewProgram`, `p.Run()`) is trivial to duplicate.

### Cobra `Args` validation

Currently cobra may be configured to require at least one arg.

**Decision:** Change `Args` for the remove command to `cobra.ArbitraryArgs` (or `cobra.MaximumNArgs(1)`) and branch in `runRemove`: zero args → interactive path, one arg → existing path.

### Confirmation prompt

**Decision:** Mirror `offerAddMany` with an `offerRemoveMany` helper that prints the selected skill names, shows `Remove N skill(s)? [y/N]`, and only proceeds on `y`/`yes`. This keeps the destructive action guarded consistently with the install flow.

### TTY detection

**Decision:** Reuse the same `isatty` / `term.IsTerminal` check already used by `search_cmd.go` to decide whether to launch the TUI.

## Risks / Trade-offs

- [Empty `melon.yaml`] If no skills are declared, the selector has nothing to show → print "No skills in melon.yaml." and exit cleanly.
- [Partial failure] If removing skill A succeeds but skill B fails, A is already gone. Mitigation: print per-skill errors and continue (same behavior as `offerAddMany` for installs).
- [Duplicate TUI code] A small amount of bubbletea boilerplate is repeated between search and remove. Mitigation: acceptable at this scale; a shared selector abstraction can be extracted later if a third consumer appears.
