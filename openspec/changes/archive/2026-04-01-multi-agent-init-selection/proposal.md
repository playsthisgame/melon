## Why

`mln init` currently prompts for a single agent and writes a single `agent_compat` entry to `mln.yaml`. Users who work across multiple agents (e.g., claude-code and cursor) must manually edit `mln.yaml` after init. The skill placement step in `mln install` already supports multiple targets, so the only gap is surfacing multi-select in the init prompt.

## What Changes

- The `mln init` agent prompt changes from a single-select to a multi-select, matching the UX of `openspec init` (checkbox-style list of all known agents).
- `mln.yaml` is written with `agent_compat` as a list of all selected agents.
- `mln install` placement logic is already multi-target aware — no change needed there.

## Capabilities

### New Capabilities

- `multi-agent-init`: Multi-select agent prompt in `mln init` that writes multiple `agent_compat` entries to `mln.yaml`.

### Modified Capabilities

<!-- No existing spec-level behavior changes. -->

## Impact

- `cmd/mln/init_cmd.go` — replace single-agent survey prompt with a multi-select checkbox prompt; update `generateManifestYAML` to accept `[]string` instead of `string` for agents.
- `internal/agents/agents.go` — no changes needed; `KnownAgents()` already returns the full list.
- `mln.yaml` output — `agent_compat` field now always written as a YAML list (already the declared type `[]string`).
