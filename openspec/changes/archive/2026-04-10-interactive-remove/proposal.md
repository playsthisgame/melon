## Why

Running `mln remove` without a skill name produces an error, forcing users to remember exact skill names from `melon.yaml`. An interactive selector — consistent with the multi-select UX already in `mln search` — lets users pick which skills to remove without needing to type names precisely.

## What Changes

- When `mln remove` is invoked with no arguments in a TTY, display a bubbletea-driven multi-select list of all skills declared in `melon.yaml`
- Users navigate with arrow keys, toggle selections with space, and confirm with enter to remove selected skills
- Non-interactive (non-TTY) behavior is unchanged: `mln remove <name>` continues to work as before
- Running `mln remove` with no args in a non-TTY (piped/CI) exits with a helpful error message

## Capabilities

### New Capabilities

- `remove-interactive`: Interactive multi-select mode for `mln remove` when no skill argument is provided

### Modified Capabilities

- `remove-command`: Adding a new requirement for no-argument interactive mode

## Impact

- `internal/cli/remove_cmd.go` — add no-arg branch that reads `melon.yaml` and launches TUI selector
- Reuses or extends the bubbletea multi-select component from the search flow
- No changes to the remove pipeline itself; selected skills are removed in sequence using existing logic
