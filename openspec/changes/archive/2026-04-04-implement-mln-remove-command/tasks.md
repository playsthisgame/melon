## 1. Implement remove_cmd.go

- [x] 1.1 Create `cmd/mln/remove_cmd.go` with a `runRemove` function
- [x] 1.2 Load `melon.yml` and error if the named dependency is not present
- [x] 1.3 Delete the dependency key from `m.Dependencies` and save the updated `melon.yml`
- [x] 1.4 Call `runInstall` to regenerate `melon.lock` and trigger pruning (symlink removal + cache deletion) for the removed dep

## 2. Wire up the command

- [x] 2.1 In `cmd/mln/main.go`, replace the TODO stub in `removeCmd.RunE` with a call to `runRemove`

## 3. Tests

- [x] 3.1 Write a test for the happy path: dep exists in `melon.yml`, it is removed from `melon.yml`, and `runInstall` is invoked
- [x] 3.2 Write a test for the error path: dep not in `melon.yml` returns a non-zero error without modifying any files
