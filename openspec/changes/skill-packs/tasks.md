## 1. Manifest type

- [x] 1.1 Add `Type string \`yaml:"type,omitempty"\`` to `manifest.Manifest` with `TypeSkill`/`TypePack` constants and an `IsPack()` helper (empty → skill)
- [x] 1.2 Validate `type` in `manifest.Load`; unknown values return an error naming the value and listing `skill`, `pack`; legacy `agent`/`workflow`/`persona`/`memory` are accepted as skills
- [x] 1.3 Validate `type` in `resolver.DefaultFetchManifest` parsing the same way (shared validation func)
- [x] 1.4 Tests: default type, pack accepted, unknown type rejected, legacy type accepted, round-trip save/load

## 2. Resolver provenance

- [x] 2.1 Add `Type`, `Direct bool`, `Via []string` to `resolver.ResolvedDep`
- [x] 2.2 Track parents for every encounter (first and repeat) and mark root-manifest deps as direct; emit sorted, de-duplicated `Via`
- [x] 2.3 Carry `Type` from each dep's fetched manifest into the result
- [x] 2.4 Improve `ErrConflict` message when one side is the root manifest and the other a pack (name both, suggest loosening)
- [x] 2.5 Tests: diamond lists both parents, direct+transitive dep, nested packs, compatible direct override picks direct's highest, incompatible override error text, determinism

## 3. Lock file

- [x] 3.1 Add `Type` (always written), `Direct` (omitempty), `Via` (omitempty) to `lockfile.LockedDep`
- [x] 3.2 Copy the new fields from `ResolvedDep` in `fetchDeps` (`install_cmd.go`) and in `installSingleDep` (`add_cmd.go`)
- [x] 3.3 Add a `lockfile.ProvenanceChanged(old, new)` (or equivalent) that compares `type`/`direct`/`via`, treating entries with empty `Type` as matching when versions match
- [x] 3.4 Make `install --frozen` fail when versions differ (existing) or provenance changed
- [x] 3.5 Tests: round-trip of new fields, legacy lock loads, frozen fails on direct flip, frozen passes on legacy lock

## 4. Placement

- [x] 4.1 Add `placer.ShouldPlace(dep, projectDir)`: false when type is pack and `<cacheDir>/<entrypoint>` is missing
- [x] 4.2 Use `ShouldPlace` in `Place`, `Unplace`, `managedGitignoreEntries`/`gitignoreSymlinkEntries`, and `list --check`
- [x] 4.3 Tests: pure pack not placed, hybrid pack placed, pure pack absent from managed gitignore entries

## 5. Install pipeline

- [x] 5.1 After `resolver.Resolve` and before `fetchDeps`, run `checkSourcePolicy` on all resolved names; error lists each blocked dep with its `via` chain
- [x] 5.2 Warn once per pack/member pair when a pack's member constraint is an exact version
- [x] 5.3 Tests (with fake resolveVersion/fetchManifest): pack members installed and placed, policy blocks a member before any fetch and lock unchanged, exact-pin warning printed

## 6. add command

- [x] 6.1 Add `--yes`/`-y` flag to `add`
- [x] 6.2 In `runAdd`, fetch the target manifest before saving; if it is a pack or has dependencies, take the full-graph path: resolve with the new dep in memory → policy check over all resolved names → print members (packs only) → TTY confirm unless `--yes` → save `melon.yaml` → `runInstallCore`
- [x] 6.3 Keep `installSingleDep` fast path for plain skills with no dependencies
- [x] 6.4 Tests: pack add installs members, decline leaves manifest untouched, policy-blocked member leaves manifest untouched, non-TTY prints members without prompt, plain skill with transitive dep installs it

## 7. remove command

- [x] 7.1 If the target's lock entry is `type: pack` or appears in any entry's `via`, save `melon.yaml` and run `runInstallCore` instead of `removeSingleDep`
- [x] 7.2 If the target is not in `melon.yaml` but is in the lock via a pack, error naming the requiring pack(s) without changes
- [x] 7.3 Tests: pack members pruned, shared member kept with updated `via`, direct member kept, removing a transitive member errors

## 8. list command

- [x] 8.1 Render `(pack)` label and indented members under each pack; keep output unchanged when no packs exist
- [x] 8.2 Add `type`, `direct`, `via` (empty array when none) to `list --json`
- [x] 8.3 Tests: tree output, unchanged output without packs, JSON fields, `--check` ignores pure packs

## 9. Index, search, info

- [x] 9.1 Add `Type` to `index.Entry`; normalize unknown values to `skill`
- [x] 9.2 Show `[pack]` label in search TUI and plain-text output (search has no `--json` mode, so no JSON change)
- [x] 9.3 In `info`, fetch the manifest at the latest tag (or default branch); for packs print a pack indicator and sorted members with constraints; add `type` and `members` to JSON
- [x] 9.4 Tests: index parsing of `type`, search labels, info for pack/skill in text and JSON

## 10. Docs and release

- [x] 10.1 Document packs in README (authoring a pack, pure vs hybrid, ranges vs pins, `--yes`)
- [x] 10.2 Update `CLAUDE.md` melon.yaml and melon.lock schema sections with `type`, `direct`, `via`
- [x] 10.3 Add a release-notes entry calling out that `allowed_sources` now applies to transitive deps
- [x] 10.4 Run `go test ./...` (incl. -race) and verify end to end against the real melon-index packs (pure, hybrid, nested): add, info, list, list --check, remove pruning, install --frozen
