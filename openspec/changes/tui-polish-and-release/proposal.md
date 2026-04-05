## Why

The CLI works correctly but outputs plain text with no visual feedback — long installs give no progress indication and `mln init` uses raw `fmt.Scan` prompts that are awkward for multi-value inputs like `agent_compat`. Adding a bubbletea TUI layer makes the tool genuinely pleasant to use and brings it to MVP release quality.

## What Changes

- Replace all plain `fmt.Fprintf` status output with lipgloss-styled output (green/yellow/red for add/update/remove, bold dep names)
- Rewrite `mln init` interactive prompts using bubbletea: scrollable multi-select list for `agent_compat` (↑↓ to navigate, space to toggle, enter to confirm), text inputs for name/version/description/type
- Add a bubbles progress bar to `mln install` that advances per-dep as each one is fetched, with a TTY fallback for non-interactive contexts (CI)
- Add a bubbles spinner to `mln add` (during `LatestTag` resolve) and `mln remove` (during install)
- Add three release artifacts: `README.md`, `.gitignore`, `.goreleaser.yaml` + GitHub Actions workflow

## Capabilities

### New Capabilities

- `tui-init`: `mln init` uses a bubbletea model with a scrollable multi-select list for `agent_compat` and text inputs for other fields
- `tui-install-progress`: `mln install` shows a live progress bar per dep fetched, with TTY detection for CI fallback
- `tui-spinners`: `mln add` and `mln remove` show a spinner during network/install waits
- `release-artifacts`: README, .gitignore, and GoReleaser config are present and correct

### Modified Capabilities

- `multi-agent-init`: the init command's interactive flow changes from raw stdio prompts to a bubbletea model — requirement-level behavior (what it collects, what it writes) is unchanged, but the interaction mechanism changes enough to warrant a delta

## Impact

- New Go dependencies: `charmbracelet/bubbletea`, `charmbracelet/bubbles`, `charmbracelet/lipgloss`
- `cmd/mln/init_cmd.go`: full rewrite of the interactive prompt section
- `cmd/mln/install_cmd.go`: fetch loop gains progress bar; `printLockDiff` gains lipgloss colors
- `cmd/mln/add_cmd.go`: spinner wraps the `LatestTag` call
- `cmd/mln/remove_cmd.go`: spinner wraps the `runInstall` call
- New files: `README.md`, `.gitignore`, `.goreleaser.yaml`, `.github/workflows/release.yml`
- `go.mod` / `go.sum`: three new dependencies
