## 1. Extract CLI logic into internal package

- [x] 1.1 Create `internal/cli/cli.go` with a `Run(name, version string)` function that builds and executes the cobra root command, accepting the binary name so help text is correct
- [x] 1.2 Move all command definitions (root, install, add, remove, init) and flag wiring from `cmd/mln/main.go` into `internal/cli`
- [x] 1.3 Update `cmd/mln/main.go` to call `cli.Run("mln", version)`
- [x] 1.4 Run existing tests to confirm no regressions

## 2. Add cmd/melon entrypoint

- [x] 2.1 Create `cmd/melon/main.go` that calls `cli.Run("melon", version)` with the same `ldflags` version injection as `cmd/mln`
- [x] 2.2 Verify `go build ./cmd/melon` produces a working binary that responds correctly to `melon --version` and `melon --help`

## 3. Update GoReleaser config

- [x] 3.1 Add a second `builds` entry in `.goreleaser.yaml` with `id: melon`, `main: ./cmd/melon`, `binary: melon`, targeting the same OS/arch matrix as `mln`
- [x] 3.2 Update the `archives` section so both `mln` and `melon` builds are included in the same archive
- [x] 3.3 Verify locally with `goreleaser build --snapshot --clean` that both binaries are produced

## 4. Update npm package

- [x] 4.1 Add `"melon": "./bin/mln.js"` to the `bin` field in `npm/package.json`
- [x] 4.2 Update `npm/postinstall.js` to extract and install both `mln` and `melon` binaries from the release archive
- [x] 4.3 Verify that after a local test install both `mln` and `melon` are on PATH and respond to `--version`
