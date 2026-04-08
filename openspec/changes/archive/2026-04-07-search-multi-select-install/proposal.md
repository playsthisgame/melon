## Why

`melon search` currently allows selecting only a single result, requiring users to re-run the command for each skill they want to install. Users should be able to select multiple skills in one pass, matching the multi-select UX already present in `melon init`.

## What Changes

- The search TUI result list changes from single-select (enter to pick one) to multi-select (space to toggle, enter to confirm all selected)
- After confirmation, all selected skills are installed in sequence
- The non-TTY plain-text output path is unchanged
- The `offerAdd` single-install prompt is replaced by a loop that installs each selected skill

## Capabilities

### New Capabilities

- `search-multi-select`: Multi-select behavior for the `melon search` TUI — space toggles selection, enter confirms and installs all selected skills

### Modified Capabilities

- none

## Impact

- `internal/cli/search_model.go`: Replace single-select list model with multi-select model (reusing `multiSelectDelegate` pattern from `init_model.go`)
- `internal/cli/search_cmd.go`: Update `runSearchTUI` return type from `string` to `[]string`; replace `offerAdd` with batch install loop
- No API or dependency changes required
