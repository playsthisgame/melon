## 1. Update init_cmd.go prompt

- [x] 1.1 Add `promptMultiChoice(cmd, question string, options []string, defaults []string) []string` helper to `init_cmd.go` — prints a numbered list, reads comma-separated or space-separated input, returns matched options; in `--yes` mode returns defaults
- [x] 1.2 Replace the `promptChoice` call for `agentName` with `promptMultiChoice` call that returns `[]string`, defaulting to `["claude-code"]`
- [x] 1.3 Update `generateManifestYAML` signature from `(name, pkgType, description, agentName string)` to `(name, pkgType, description string, agentNames []string)`
- [x] 1.4 Update `generateManifestYAML` body: render `agent_compat` as a YAML list (one `- <name>` entry per agent, or `[]` if empty)
- [x] 1.5 Update the call site in `runInit` to pass the `[]string` result

## 2. Fix broken test

- [x] 2.1 Update `TestGenerateManifestYAML_ParsesCleanly` in `cmd/mln` to call `generateManifestYAML` with `[]string{"claude-code"}` instead of `"claude-code"`

## 3. Verify

- [x] 3.1 Run `go test ./...` — all tests pass
- [ ] 3.2 Manually run `mln init` in a temp directory, select multiple agents, verify `mln.yaml` `agent_compat` contains all selected agents
