## Context

All four commands work correctly but produce plain undecorated text. `mln init` uses `bufio.Scanner` prompts that can't handle multi-value input elegantly. `mln install` gives no feedback during the (potentially slow) fetch loop. The bubbletea ecosystem (`bubbletea` + `bubbles` + `lipgloss`) is the idiomatic Go TUI stack and fits well here — it is lightweight, composable, and handles TTY detection naturally.

## Goals / Non-Goals

**Goals:**

- `mln init` interactive mode: replace `promptMultiChoice` and `promptChoice` with a bubbletea multi-step form (text inputs + scrollable multi-select list)
- `mln install` fetch loop: show a live `bubbles/progress` bar advancing per dep, with TTY fallback to plain text for CI
- `mln add`: spinner during `LatestTag` network call
- `mln remove`: spinner during the `runInstall` call
- `printLockDiff`: use lipgloss for green/yellow/red colored diff lines
- Release artifacts: `README.md`, `.gitignore`, `.goreleaser.yaml`, `.github/workflows/release.yml`

**Non-Goals:**

- Full alt-screen TUI — all rendering stays inline in the terminal scroll buffer
- Animating placement or lock-write steps (only the fetch loop is slow enough to warrant feedback)
- Changing any functional behavior — TUI is purely presentation

## Decisions

### bubbletea multi-step model for `mln init`

`init_cmd.go` gains a `initModel` tea.Model with a `step` enum: `stepName → stepType → stepDescription → stepAgents → stepDone`. Each step renders a prompt; `stepAgents` uses `bubbles/list` in multi-select mode. When the model reaches `stepDone`, `tea.Quit` is sent and the collected values are passed back to `runInit` for YAML generation. The existing `--yes` path is unchanged.

_Alternative_: use `charmbracelet/huh` (a form library). Rejected — adds a 4th dependency; `bubbles/list` + `bubbles/textinput` cover everything needed with deps already in scope.

### Goroutine + `tea.Msg` for the install progress bar

The fetch loop in `runInstall` is extracted into `fetchDeps(deps []resolver.ResolvedDep, dir string, onFetch func(i int, dep resolver.ResolvedDep, result fetcher.Result, err error))`. A `tea.Program` runs an `installModel` that tracks progress; a goroutine calls `fetchDeps` and sends `depFetchedMsg` after each dep. The model advances `bubbles/progress` on each message and quits when all deps are done. Result/error state is collected via a shared slice.

_Alternative_: render progress directly with `\r` escape codes without a tea.Program. Rejected — bubbletea handles line clearing and TTY detection more robustly across platforms.

### TTY detection for install progress bar

`os.Stdout.Fd()` is passed to `term.IsTerminal` (`golang.org/x/term`, already an indirect dependency via bubbletea). If not a TTY, the progress model is skipped and plain `fmt.Fprintf` lines are printed instead, preserving CI output.

### Spinner pattern for `add` and `remove`

A small `withSpinner(message string, fn func() error) error` helper launches a `tea.Program` with a `bubbles/spinner` model, runs `fn` in a goroutine, sends `doneMsg` on completion, and returns any error. Called in `runAdd` around `LatestTag` and in `runRemove` around `runInstall`.

### lipgloss diff colors

`printLockDiff` is updated to use three lipgloss styles: `addStyle` (green bold `+`), `updateStyle` (yellow bold `~`), `removeStyle` (red bold `-`). No structural change — only the format strings change.

## Risks / Trade-offs

- **bubbletea + cobra interaction**: bubbletea takes over stdin/stdout for the duration of the model. cobra's `cmd.OutOrStdout()` won't be used inside a tea.Program — output goes through the model's `View()` string. This is fine for init and install but means the spinner helpers must be careful not to mix tea output with cobra output. Mitigation: run tea.Program, return, then resume normal cobra output.
- **Test coverage for TUI**: bubbletea models are testable via `tea.NewProgram` with `WithInput`/`WithOutput` overrides, but it adds complexity to unit tests. Mitigation: keep business logic (YAML generation, fetch loop) fully separate from the model; test the logic, not the rendering.
- **`golang.org/x/term` dependency**: needed for TTY detection. It is already an indirect dep of bubbletea so it won't add to `go.mod` size.
