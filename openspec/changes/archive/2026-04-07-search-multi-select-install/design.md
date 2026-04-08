## Context

`melon search` presents a bubbletea list of matching skills. The current model (`searchModel` in `search_model.go`) uses a standard `list.Model` with a custom `searchDelegate` that renders each item as two lines and returns a single selected path on Enter. The `multiSelectDelegate` pattern in `init_model.go` already solves multi-select with space-to-toggle and a `selected map[int]bool` — the approach is proven and consistent.

## Goals / Non-Goals

**Goals:**
- Allow users to toggle multiple search results with space and confirm all with Enter
- Install each selected skill sequentially after confirmation
- Reuse the `multiSelectDelegate` visual pattern for consistency with `mln init`

**Non-Goals:**
- Changing non-TTY (pipe/script) output — still prints plain text
- Parallel installation of skills
- Filtering or sorting search results

## Decisions

**Reuse `multiSelectDelegate` vs. build a new delegate**
Reuse it. The delegate is already in the same package and handles the `[✓]` / `[ ]` rendering correctly. The only difference is that `searchResultItem` renders as two lines (name + description) vs. one — so we need a new delegate that keeps both lines while adding checkbox rendering.

**Return type of `runSearchTUI`**
Change from `string` to `[]string`. The function returns all selected paths. An empty slice means the user cancelled or made no selection.

**Install loop placement**
Move the install loop into `search_cmd.go`. After `runSearchTUI` returns the slice, iterate and call `runAdd` for each path. This avoids the TUI layer knowing about installation, keeping concerns separated.

**Prompt before installing**
Replace the current `offerAdd` single-prompt with a summary prompt: print the list of selected skills and ask `Install N skill(s)? [y/N]` once. This is less noisy than prompting per-skill and matches user expectations from similar CLI tools.

## Risks / Trade-offs

- [Two-line item height with checkbox] The existing `searchDelegate` uses 2-line height. Adding a checkbox to line 1 means it must render selected state while keeping the description on line 2. → Trivial: copy the delegate pattern from `init_model.go` and adjust for the two-line layout.
- [Install failure mid-batch] If installing skill N fails, skills 1..N-1 are already installed. → Acceptable for now; each skill install is idempotent (`mln add` succeeds if already present). Log the error and continue.
