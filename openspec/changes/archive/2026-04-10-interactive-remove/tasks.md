## 1. Cobra Command Update

- [x] 1.1 Change `removeCmd.Args` in `internal/cli/cli.go` from `cobra.ExactArgs(1)` to `cobra.MaximumNArgs(1)`
- [x] 1.2 Update `removeCmd.Use` to `"remove [name]"` to reflect the optional argument

## 2. TUI Model

- [x] 2.1 Create `internal/cli/remove_model.go` with a `removeModel` bubbletea model that accepts a list of `(name, version)` pairs
- [x] 2.2 Implement a `removeMultiSelectDelegate` that renders each skill as `[✓] <name>  <version>` (single-line rows)
- [x] 2.3 Implement `Init`, `Update` (arrow keys, space toggle, enter confirm, esc/ctrl+c cancel), and `View` on `removeModel`
- [x] 2.4 Add a `runRemoveTUI(skills []removeSkillItem) ([]string, error)` helper that runs the program and returns selected names

## 3. Confirmation Helper

- [x] 3.1 Add `offerRemoveMany(cmd *cobra.Command, names []string) error` in `remove_cmd.go` that prints selected skills, prompts `Remove N skill(s)? [y/N]:`, and calls `runRemove` for each confirmed name

## 4. Interactive Branch in runRemove

- [x] 4.1 At the top of `runRemove`, check `len(args) == 0`; if true, branch to interactive path
- [x] 4.2 In the interactive path, load `melon.yaml`, extract dependency names+versions, and handle the empty case (print "No skills in melon.yaml." and return nil)
- [x] 4.3 Detect TTY (reuse the `isatty` check from `search_cmd.go`); if non-TTY, return an error "remove: skill name required (non-interactive mode)"
- [x] 4.4 Call `runRemoveTUI` with the skill list, handle cancellation (empty selection → return nil)
- [x] 4.5 Call `offerRemoveMany` with the selected names

## 5. Tests

- [x] 5.1 Add unit tests for `removeModel`: toggle behavior, enter with no selection, esc cancels
- [x] 5.2 Add a test for the non-TTY no-args path in `runRemove` (expects non-zero error)
- [x] 5.3 Add a test for the empty `melon.yaml` no-args path (expects clean exit with message)
