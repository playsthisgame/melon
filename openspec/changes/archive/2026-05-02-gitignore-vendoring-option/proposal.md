## Why

When users run `melon init`, the `.melon/` cache directory and any melon-managed symlinks in agent tool directories (e.g. `.claude/skills/`) are not gitignored, so they silently end up committed if the user runs `git add .`. Melon should give users a clear, configurable choice between vendoring (committing deps) and non-vendoring (gitignoring deps), and automate the gitignore maintenance when they choose not to vendor.

## What Changes

- Add a `vendor` boolean field to `melon.yaml` (default: `true` to preserve current behavior)
- `melon init` prompts the user for their vendoring preference and writes `vendor: false` when opted out
- When `vendor: false`, `melon install` ensures `.melon/` and each managed symlink path are present in `.gitignore`, creating the file if it does not exist
- When `vendor: false`, `melon add` appends the new skill's symlink path(s) to `.gitignore`
- When `vendor: false`, `melon remove` removes the skill's symlink path(s) from `.gitignore`
- When `vendor: true` (default), melon never touches `.gitignore` — existing behavior is unchanged

## Capabilities

### New Capabilities

- `gitignore-management`: Melon reads the `vendor` flag and keeps `.gitignore` in sync with managed symlink paths and the `.melon/` cache directory across `init`, `install`, `add`, and `remove` commands.

### Modified Capabilities

- `skill-placement`: The placer pipeline now has a post-placement gitignore sync step when `vendor: false`.
- `add-command`: After placing a new dep, the add command must update `.gitignore` when `vendor: false`.
- `remove-command`: After removing a dep, the remove command must clean up `.gitignore` entries when `vendor: false`.

## Impact

- `internal/manifest/` — new `Vendor bool` field (yaml: `vendor`, default `true`)
- `internal/cli/init_cmd.go` — new prompt/flag for vendoring preference
- `internal/cli/install_cmd.go` — calls gitignore sync after place step
- `internal/cli/add_cmd.go` — calls gitignore sync after place step
- `internal/cli/remove_cmd.go` — calls gitignore cleanup after unplace step
- New `internal/gitignore/` package — pure functions for reading, appending, and removing entries in a `.gitignore` file
- No external dependencies added
