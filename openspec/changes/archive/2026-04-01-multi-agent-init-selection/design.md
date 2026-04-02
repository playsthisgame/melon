## Context

`mln init` is implemented in `cmd/mln/init_cmd.go` using the `survey` library (or equivalent interactive prompts). The current flow asks the user to pick a single agent from a list using a `Select` prompt, then passes that single string to `generateManifestYAML`. The `agent_compat` field in `manifest.Manifest` is already `[]string`, so the serialization layer requires no changes — only the prompt and the manifest generator need updating.

`internal/agents.KnownAgents()` already returns the full sorted list of known agent names, so no new data needs to be added.

## Goals / Non-Goals

**Goals:**
- Replace the single-agent `Select` prompt in `mln init` with a multi-select `MultiSelect` prompt.
- Allow zero or more agents to be selected (zero is valid — the user may not know yet).
- Update `generateManifestYAML` signature to accept `[]string` for agents.
- Write `agent_compat` as a YAML list in the generated `mln.yaml`.

**Non-Goals:**
- No changes to `mln install` placement logic — it already handles multiple targets.
- No changes to `internal/agents`, `internal/placer`, or any other package.
- No changes to the `--yes` fast-path behavior beyond selecting a sensible default (empty list or `["claude-code"]`).

## Decisions

**Use `survey.MultiSelect` over a custom checkbox loop.**
The `survey` library (already used by `mln init`) provides `MultiSelect` out of the box. It renders exactly like `openspec init`'s tool selector. No new dependencies needed.

**Default selection for `--yes` mode: `["claude-code"]`**
The `--yes` flag must produce a runnable `mln.yaml` without prompts. Defaulting to `["claude-code"]` preserves the existing non-interactive behavior and is the most common case.

**`generateManifestYAML(name, pkgType, description string, agents []string) string`**
Changing the `agents` parameter from `string` to `[]string` is a clean internal refactor. The only call site is `init_cmd.go`. The manifest YAML template writes `agent_compat` as a YAML inline list.

## Risks / Trade-offs

- **Zero agents selected** → `agent_compat: []` in `mln.yaml`. `mln install` will log "no target agent directories" and skip placement, which is expected. Not a bug.
- **Test breakage** → `TestGenerateManifestYAML_ParsesCleanly` in `cmd/mln` currently passes a single string. It must be updated to pass `[]string{"claude-code"}` — a trivial one-line fix.
