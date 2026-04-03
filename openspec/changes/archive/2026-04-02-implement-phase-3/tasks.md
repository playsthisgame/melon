## 1. Resolver — Core Implementation

- [x] 1.1 Create `internal/resolver/resolver.go` with the `Resolve(manifest Manifest) ([]ResolvedDep, error)` signature and package skeleton
- [x] 1.2 Implement GitHub raw content URL construction for both root-level and monorepo-subdir `mln.yaml` paths (e.g. `https://raw.githubusercontent.com/<owner>/<repo>/<tag>/<subdir>/mln.yaml`)
- [x] 1.3 Implement BFS traversal: start from `manifest.Dependencies`, fetch each dep's manifest, enqueue its deps, track visited `name@version` pairs to avoid cycles
- [x] 1.4 Implement greedy highest-compatible-version selection: for each dep name, keep the highest resolved version satisfying all seen constraints; return `ErrVersionConflict` with named packages and constraints when incompatible
- [x] 1.5 Handle missing remote `mln.yaml` gracefully — treat as dep with no transitive dependencies (do not error)
- [x] 1.6 Populate all `ResolvedDep` fields from the fetched manifests and resolved tag (use `LatestTag` from `internal/fetcher` to resolve constraint → pinned version)

## 2. Resolver — Tests

- [x] 2.1 Add `testdata/fixture-direct-only.yaml` — manifest with one direct dep and no transitive deps
- [x] 2.2 Add `testdata/fixture-transitive.yaml` — manifest where a direct dep itself has a dependency
- [x] 2.3 Add `testdata/fixture-conflict.yaml` — two direct deps that require incompatible versions of a shared transitive dep
- [x] 2.4 Write `internal/resolver/resolver_test.go` with table-driven tests covering: happy path (direct only), transitive inclusion, diamond resolution, and version conflict error message content

## 3. Wire Resolver into Install

- [x] 3.1 Update `cmd/mln/main.go` install command to call `resolver.Resolve(manifest)` instead of iterating `manifest.Dependencies` directly
- [x] 3.2 Pass the full `[]ResolvedDep` slice from the resolver to `fetcher.Fetch` and `lockfile.Save` so transitive deps are fetched and locked
- [x] 3.3 Pass the full resolved dep list to `placer.Place` so transitive deps are symlinked in agent directories

## 4. --frozen Flag

- [x] 4.1 Add `--frozen` flag to the install cobra command
- [x] 4.2 When `--frozen` is set: run the full resolve+fetch pipeline, compute the new lock, call `lockfile.Diff` against the on-disk lock, print added/removed/updated entries, and exit non-zero if any diff exists — do NOT write the new lock file

## 5. mln add Command

- [x] 5.1 Implement the `add` cobra subcommand skeleton in `cmd/mln/main.go` (parse `<dep>[@constraint]` argument)
- [x] 5.2 If no constraint given, call `fetcher.LatestTag` to resolve the latest semver tag and construct a `^<version>` constraint
- [x] 5.3 Read existing `mln.yaml`, add or update the dep entry (print a warning if updating an existing entry), write the updated manifest back to disk
- [x] 5.4 Run the full install pipeline (resolve → fetch → lock → place) after updating `mln.yaml`

## 6. Integration Verification

- [ ] 6.1 Manually verify `mln add` on a dep with a real transitive dependency installs and places both direct and transitive deps, and writes a correct `mln.lock`
- [ ] 6.2 Manually verify `mln install --frozen` exits non-zero after editing `mln.yaml` without regenerating the lock
- [x] 6.3 Run `go test ./...` and confirm all unit and integration tests pass
