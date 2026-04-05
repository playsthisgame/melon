## 1. Version Variable

- [x] 1.1 Add `var version = "dev"` package-level variable to `cmd/mln/main.go`
- [x] 1.2 Set `rootCmd.Version = version` in `main.go` to wire Cobra's built-in `--version` / `-v` flag

## 2. Build-time Injection

- [x] 2.1 Add `-ldflags "-X main.version={{ .Version }}"` to the `builds` section in `.goreleaser.yaml`

## 3. Init Command

- [x] 3.1 Replace the hardcoded `version: 0.1.0` string in `cmd/mln/init_cmd.go` with the runtime `version` variable
