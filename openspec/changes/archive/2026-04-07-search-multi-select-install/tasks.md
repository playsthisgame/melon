## 1. Update search_model.go — multi-select TUI

- [x] 1.1 Add a `searchMultiSelectDelegate` struct with `selected map[int]bool` that renders two lines per item: a checkbox + name line and an indented description line
- [x] 1.2 Update `searchModel` to hold `selected map[int]bool` instead of a single `selected string`
- [x] 1.3 Handle `tea.KeySpace` in `searchModel.Update` to toggle the item at the current index and refresh the delegate
- [x] 1.4 Handle `tea.KeyEnter` in `searchModel.Update` to collect all selected paths into `[]string` and quit
- [x] 1.5 Update `searchModel.View` hint line to read `↑↓ navigate  space to toggle  enter to install  esc to cancel`

## 2. Update search_cmd.go — batch install flow

- [x] 2.1 Change `runSearchTUI` return type from `(string, error)` to `([]string, error)`
- [x] 2.2 Delete `offerAdd`; replace with `offerAddMany(cmd *cobra.Command, paths []string) error` that prints the selected list and prompts `Install N skill(s)? [y/N]`
- [x] 2.3 In `offerAddMany`, on confirmation iterate over `paths` and call `runAdd` for each, continuing on error (log each failure)
- [x] 2.4 Update the `runSearch` TTY branch to call `runSearchTUI`, then pass the result slice to `offerAddMany`

## 3. Verification

- [x] 3.1 Build the binary (`go build ./...`) and confirm no compilation errors
- [ ] 3.2 Run `mln search <term>` against a live index hit, toggle multiple results with space, confirm with enter, and verify all skills are installed
- [ ] 3.3 Verify pressing enter with no selection exits cleanly with no installs
- [ ] 3.4 Verify esc/ctrl+c exits cleanly with no installs
- [x] 3.5 Verify `mln search <term>` in a non-TTY context still prints plain-text output unchanged
