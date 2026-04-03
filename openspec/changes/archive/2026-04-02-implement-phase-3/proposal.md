## Why

Phase 2 left `mln install` fetching only direct dependencies with no transitive resolution. Phase 3 completes the core dependency management loop by adding the resolver, wiring up `mln add`, and adding `--frozen` for CI — making melon a functional, reproducible package manager.

## What Changes

- Implement `internal/resolver` — fetch transitive `melon.yaml` files via GitHub raw API, build a DAG, apply greedy highest-compatible-version resolution, and return `ErrVersionConflict` on incompatible transitive constraints
- Update `mln install` to run the full resolver rather than operating on direct deps only
- Implement `mln add <dep>[@version]` — resolve latest tag, update `melon.yaml`, run full install
- Add `--frozen` flag to `mln install` — fail if `melon.lock` would change (CI guard)
- Add fixture manifests in `testdata/` and unit tests for the resolver (happy path, transitive deps, version conflict)

## Capabilities

### New Capabilities

- `dependency-resolution`: Transitive dependency graph resolution with greedy semver strategy and `ErrVersionConflict` on conflicts
- `add-command`: `mln add` command that looks up the latest tag, updates `melon.yaml`, and runs a full install

### Modified Capabilities

- `skill-placement`: `mln install` now operates on the full resolved graph (including transitive deps) rather than only direct dependencies

## Impact

- `internal/resolver/resolver.go` — new package, network calls to GitHub raw content API
- `cmd/mln/main.go` / install command — swaps direct-dep loop for resolver output
- `cmd/mln/main.go` / add command — new subcommand implementation
- `internal/fetcher/fetcher.go` — `LatestTag` already exists; wired up by add command
- `testdata/` — new fixture YAML files for resolver unit tests
- No breaking changes to `melon.yaml` or `melon.lock` formats
