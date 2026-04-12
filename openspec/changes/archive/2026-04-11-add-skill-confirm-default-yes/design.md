## Context

After a user selects one or more skills from the interactive search TUI, `offerAddMany` presents a final confirmation before running `mln add`. The current prompt is `[y/N]`, making "no" the default — pressing Enter cancels the operation. This is inconsistent with the user's demonstrated intent (they just selected skills) and adds an extra keystroke on the happy path.

The change is isolated to a single function in a single file: `offerAddMany` in `internal/cli/search_cmd.go`.

## Goals / Non-Goals

**Goals:**
- Change the confirmation prompt to `[Y/n]` so pressing Enter proceeds with installation
- Update input-acceptance logic to treat empty input as "yes"

**Non-Goals:**
- Changing any other confirmation prompts in the codebase (e.g., `init` overwrite prompt uses `[y/N]` intentionally — destructive action)
- Adding a `--yes` / `--no-confirm` flag to skip the prompt entirely
- Changing non-TTY / piped behavior

## Decisions

**Treat blank input as yes, not no.**
The user selected skills from the TUI — intent is already established. The confirmation is a safety net against accidental Enter-during-selection, but once the list is shown and reviewed, defaulting to yes matches expectations. This mirrors common CLI conventions (e.g., `npm install` does not re-ask; `brew install` proceeds after selection).

**Keep the confirmation prompt at all.**
Removing it entirely was considered, but one accidental multi-select in the TUI could install several unwanted skills. A single Enter-to-confirm is low friction and still provides a review moment.

## Risks / Trade-offs

- [Risk] Users who muscle-memoried the old `[y/N]` behavior may inadvertently install skills → Mitigation: the prompt text visibly changes to `[Y/n]`; the change is small and the blast radius of an accidental install is low (skills can be removed with `mln remove`)

## Migration Plan

No migration needed. This is a UX-only change with no persisted state involved. Deploy by shipping a new binary release.
