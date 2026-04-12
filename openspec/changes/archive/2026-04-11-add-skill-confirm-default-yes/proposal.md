## Why

After selecting skills from a search result, the user has already indicated intent by choosing them — the confirmation prompt defaulting to "no" creates unnecessary friction and feels contrary to the user's stated intent. Defaulting to "yes" reduces the number of keystrokes for the happy path.

## What Changes

- The `offerAddMany` confirmation prompt changes from `[y/N]` to `[Y/n]`
- An empty input (pressing Enter) now proceeds with installation instead of cancelling
- The acceptance check is updated to treat empty input as confirmation

## Capabilities

### New Capabilities

- `search-add-confirm-default`: The install confirmation prompt shown after selecting skills from search results defaults to yes, so pressing Enter installs without requiring an explicit `y`.

### Modified Capabilities

- `skill-search`: The interactive post-selection confirmation behavior changes — the default answer flips from no to yes.

## Impact

- `internal/cli/search_cmd.go`: `offerAddMany` — prompt string and input-acceptance logic
- No changes to manifest, lock file, resolver, fetcher, or placer
