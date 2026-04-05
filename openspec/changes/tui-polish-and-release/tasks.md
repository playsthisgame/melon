## 1. Dependencies

- [x] 1.1 Run `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles github.com/charmbracelet/lipgloss` and commit the updated `go.mod` / `go.sum`

## 2. Lipgloss diff colors

- [x] 2.1 Define three lipgloss styles in `cmd/mln/install_cmd.go`: `addStyle` (green bold), `updateStyle` (yellow bold), `removeStyle` (red bold)
- [x] 2.2 Update `printLockDiff` to use these styles for `+`, `~`, and `-` lines

## 3. Install progress bar

- [x] 3.1 Extract the fetch loop from `runInstall` into a `fetchDeps(resolved []resolver.ResolvedDep, dir string, onFetch func(i int, name string, err error)) ([]lockfile.LockedDep, error)` helper
- [x] 3.2 Add TTY detection using `golang.org/x/term`: `isTTY() bool` helper in `cmd/mln/`
- [x] 3.3 Implement `installProgressModel` (bubbletea model) in `cmd/mln/progress.go` — wraps `bubbles/progress`, receives `depFetchedMsg` and `fetchDoneMsg`, renders the bar + current dep label
- [x] 3.4 Wire the progress model into `runInstall`: if TTY, run `tea.Program` with the model driving `fetchDeps` via a goroutine; if non-TTY, fall back to plain `fmt.Fprintf` lines

## 4. Spinners for add and remove

- [x] 4.1 Implement `withSpinner(label string, fn func() error) error` helper in `cmd/mln/spinner.go` using `bubbles/spinner` + `tea.Program`; no-ops when not a TTY
- [x] 4.2 Wrap the `LatestTag` call in `runAdd` with `withSpinner("Resolving latest version of <dep>…", ...)`
- [x] 4.3 Wrap the `runInstall` call in `runRemove` with `withSpinner("Updating…", ...)`

## 5. bubbletea init form

- [x] 5.1 Implement `initModel` tea.Model in `cmd/mln/init_model.go` with steps: `stepName`, `stepType`, `stepDescription`, `stepAgents`, `stepDone`; use `bubbles/textinput` for text steps and `bubbles/list` for type and agent steps
- [x] 5.2 Style the list items with lipgloss: highlight selected items with a checkmark glyph and color, dim unselected ones
- [x] 5.3 Update `runInit` to launch the bubbletea form when not in `--yes` mode; collect the results and pass them to `generateManifestYAML`
- [x] 5.4 Ensure the overwrite-protection check still runs before launching the form

## 6. Release artifacts

- [x] 6.1 Write `README.md` — install instructions (`go install`), quick-start, command reference for all four commands, `agent_compat` table, comparison with npx skill installers
- [x] 6.2 Write `.gitignore` — ignore `.melon/`, do not ignore `melon.lock` or agent skill dirs
- [x] 6.3 Write `.goreleaser.yaml` — builds for macOS arm64/amd64, Linux amd64, Windows amd64
- [x] 6.4 Write `.github/workflows/release.yml` — triggers on `v*` tag push, runs GoReleaser with `GITHUB_TOKEN`
