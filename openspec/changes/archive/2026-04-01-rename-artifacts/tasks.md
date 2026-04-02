## 1. Rename store directory

- [x] 1.1 In `internal/store/store.go`, change `StoreDir = ".mln"` to `StoreDir = ".melon"`

## 2. Rename manifest file

- [x] 2.1 In `cmd/mln/init_cmd.go`, change `"mln.yaml"` path to `"melon.yml"` (manifest path construction and overwrite check messages)
- [x] 2.2 In `cmd/mln/init_cmd.go`, update `generateManifestYAML` header comment from `# mln.yaml` to `# melon.yml`
- [x] 2.3 In `cmd/mln/init_cmd.go`, update all user-facing messages and comments referencing `mln.yaml` or `.mln/`
- [x] 2.4 In `cmd/mln/install_cmd.go`, change `"mln.yaml"` manifest path to `"melon.yml"`
- [x] 2.5 In `internal/manifest/manifest.go` and `schema.go`, update doc comments referencing `mln.yaml`

## 3. Rename lock file

- [x] 3.1 In `cmd/mln/install_cmd.go`, change `"mln.lock"` lock path to `"melon.lock"` and update related messages/comments
- [x] 3.2 In `internal/lockfile/lockfile.go`, update doc comment referencing `mln.lock`

## 4. Update CLI help strings

- [x] 4.1 In `cmd/mln/main.go`, update Short descriptions for `installCmd`, `addCmd`, `removeCmd` to reference `.melon/`, `melon.yml`, `melon.lock`
- [x] 4.2 In `internal/placer/placer.go`, update comment referencing `mln.yaml`

## 5. Update .gitignore

- [x] 5.1 In `.gitignore`, change `.mln/` entry to `.melon/`

## 6. Update tests

- [x] 6.1 In `cmd/mln/init_cmd_test.go`, replace all `"mln.yaml"` with `"melon.yml"` and `".mln"` with `".melon"`
- [x] 6.2 In `internal/manifest/manifest_test.go`, replace all `"mln.yaml"` with `"melon.yml"`
- [x] 6.3 In `internal/lockfile/lockfile_test.go`, replace all `"mln.lock"` with `"melon.lock"`

## 7. Verify

- [x] 7.1 Run `go test ./...` — all tests pass
