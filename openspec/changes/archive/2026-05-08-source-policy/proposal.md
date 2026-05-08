## Why

Companies using melon internally need a way to restrict which GitHub repositories developers can install skills from — limiting installs to approved or internally-hosted sources and preventing accidental or unauthorized use of public skills.

## What Changes

- Add a `policy` block to `melon.yaml` with an `allowed_sources` field — a list of glob patterns defining permitted dependency sources (e.g. `github.com/my-company/*`)
- `melon add` validates the dependency path against `allowed_sources` before writing to `melon.yaml`, failing fast with a clear error if the source is not permitted
- `melon install` validates every dependency in `melon.yaml` against `allowed_sources` before fetching, so manually-edited manifests are also enforced
- When `allowed_sources` is absent, all sources are permitted (fully backwards-compatible)

## Capabilities

### New Capabilities
- `source-policy`: Configuration and enforcement of an allowlist of permitted skill dependency sources

### Modified Capabilities
- `add-command`: `melon add` must reject dependencies whose source path does not match `allowed_sources` when a policy is configured
- `install-pipeline`: `melon install` must validate all dependencies against `allowed_sources` before fetching

## Impact

- `internal/manifest/schema.go` — new `Policy` struct and `Policy *PolicyConfig` field on `Manifest`
- `internal/cli/add_cmd.go` — source validation before writing to manifest
- `internal/cli/install_cmd.go` — source validation before fetching each dep
- `README.md` — document the `policy` block in the manifest reference
