## Context

Melon's install pipeline is Resolve → Fetch → Write lock → Place → Prune. The resolver (`internal/resolver/resolver.go`) already does BFS over each dependency's `melon.yaml`, so a manifest whose `dependencies` list a set of skills already pulls those skills in transitively. What's missing:

- Every resolved dep is assumed to be a skill (entrypoint defaults to `SKILL.md`) and the placer symlinks every dep into each harness directory, so a group manifest would show up as a bogus skill.
- The resolver discards provenance: on a repeat encounter it only checks compatibility and `continue`s, so nobody knows which pack pulled a skill in.
- `checkSourcePolicy` runs in `install_cmd.go` only over `m.Dependencies` (direct deps), so transitive deps bypass `allowed_sources`. Packs make transitive deps the normal case, so this gap has to close.
- `melon add` and `melon remove` do **not** run the full pipeline today. `add` calls `installSingleDep` (`internal/cli/add_cmd.go`), which resolves/fetches/locks/places only the named dep and never follows its manifest's `dependencies`. `removeSingleDep` (`internal/cli/remove_cmd.go`) deletes only the named lock entry. For a pack, `add` would install nothing useful and `remove` would orphan every member.
- The full `melon install` pipeline already prunes correctly: removing a pack from `melon.yaml` → re-resolve → members absent from the new lock land in `diff.Removed` → unplaced and deleted, while members still required elsewhere stay in the graph.

## Goals / Non-Goals

**Goals:**
- Community-authored packs installable with `melon add github.com/owner/repo/path`.
- Pure and hybrid packs, with correct placement for each.
- Lock-file provenance that drives `list` tree output and explains why a skill is installed.
- `allowed_sources` enforced on the whole resolved graph.
- Discovery: packs labeled in search and expanded in `info`; members shown before `add` commits.

**Non-Goals:**
- Excluding individual members of a pack.
- Scaffolding (`melon init --pack`) or publishing tooling.
- Changing the resolution algorithm (still greedy BFS, highest-compatible).
- Packs that override placement settings (`harnesses`, `outputs`) of the consuming project — a pack's own `harnesses`/`outputs` are ignored, as today for any dependency.

## Decisions

### 1. Explicit `type: pack` instead of inferring from "no SKILL.md"
Inferring a pack from a missing entrypoint would make a broken skill (typo'd entrypoint) silently install as nothing. An explicit field gives clear error messages, lets the index label packs, and lets `info` know to list members. Default is `skill`, so all existing manifests are unchanged. Unknown values fail `manifest.Load` validation with a message listing the valid values.

*Alternative considered:* a separate `melon-pack.yaml` file. Rejected — it duplicates the dependency schema and the resolver would need a second fetch per dep.

### 2. Pure vs. hybrid is decided at placement time from the fetched files
The resolver knows `type` from the manifest but cannot cheaply know whether the entrypoint file exists (it only fetches `melon.yaml`). The fetcher already sparse-checks-out the pack directory into `.melon/<dep@version>/`, so the placer checks: `type == pack` and `<cacheDir>/<entrypoint>` does not exist → skip placement. Hybrid packs (entrypoint present) are placed like any skill. Pure packs are still fetched and hashed so caching, lock entries, and pruning stay uniform; the cost is one tiny checkout.

The lock entry for a pure pack keeps `entrypoint` populated (default `SKILL.md`) for format consistency; `list --check` skips placement checks for entries the placer would skip, using the same predicate (shared helper `placer.ShouldPlace(dep, projectDir)`).

### 3. Provenance: `direct` + `via` on each lock entry
```yaml
- name: github.com/acme/skills/spring-boot
  version: 1.4.0
  type: skill
  direct: false
  via:
    - github.com/community/packs/spring-expert-pack
```
- `direct: true` when the name is a key of the project's `melon.yaml` dependencies.
- `via` is the sorted, de-duplicated list of resolved deps whose manifest declares this dep. Empty/omitted for deps reached only directly.
- The resolver gains a `parents map[string]map[string]struct{}`; it records the parent on *every* encounter (first and repeat), not just the first. `ResolvedDep` gains `Type`, `Direct`, `Via`; `fetchDeps` copies them into `LockedDep`.
- Fields are `omitempty` except `type`, keeping lock diffs small for plain skills.

*Alternative considered:* storing a nested tree in the lock. Rejected — the lock is a flat, sorted list today, and a flat list with parent pointers reconstructs any tree view while keeping diffs stable.

### 4. Direct deps win if compatible — no algorithm change
BFS already enqueues all direct deps before any transitive ones, so a direct dep is always selected first at the highest version matching the user's constraint. The pack's constraint is then checked against that selection; if incompatible, `ErrConflict` is returned. The only change is the error wording: when one side is the root manifest and the other is a pack, the message says so and suggests loosening the direct constraint or asking the pack to widen its range. Pack authors are told (in docs and in a `melon install` warning) to use ranges: when a resolved dep's manifest is `type: pack` and any member constraint is an exact version, print a one-line warning.

### 5. Policy check moves after resolve, before fetch
`checkSourcePolicy` is called twice in `install`: once on direct deps before resolution (fast fail, no network, current behavior preserved) and once on all resolved names before `fetchDeps`. Resolution only downloads `melon.yaml` files via raw.githubusercontent.com, never skill content, so "no fetching before policy check" still holds for skill files. The error lists each blocked dep with its `via` chain so users can see which pack pulled it in. `add` keeps its pre-write check on the added name and, for packs, also checks the resolved member set *before* writing `melon.yaml` (see Decision 7), so a blocked member never leaves the manifest half-edited.

### 6. Index and discovery
`index.Entry` gains `Type string \`yaml:"type,omitempty"\``. Search renders `[pack]` next to pack entries (TUI and plain text) and includes `type` in JSON. `info` already fetches the dep's metadata; for packs it additionally fetches the manifest at the latest matching tag via `DefaultFetchManifest` and lists members with their constraints. `info` determines pack-ness from the fetched manifest, not the index, so unindexed packs work too.

### 7. `add` and `remove` use the full pipeline when the graph changes shape
`installSingleDep` stays as the fast path for the common case: a plain skill with no `dependencies`. `add` already fetches the target's manifest to read its entrypoint; if that manifest is `type: pack` **or** declares any `dependencies`, `add` instead:

1. Resolves the project manifest *with the new dep added in memory* (`resolver.Resolve`), which yields the full member set.
2. Runs `checkSourcePolicy` over all resolved names; on failure, exits without touching `melon.yaml`.
3. For packs, prints the member list (direct members, one level deep, nested packs marked `(pack)`). In a TTY it prompts `Install pack with N skills? [Y/n]`; a new `--yes`/`-y` flag skips the prompt; non-TTY never prompts. Declining exits 0 without touching `melon.yaml`.
4. Saves `melon.yaml` and runs `runInstallCore` (full resolve/fetch/lock/place/prune).

This incidentally fixes a latent bug: `add` of a plain skill that declares transitive deps currently skips them.

Symmetrically, `remove` uses the full pipeline (save manifest → `runInstallCore`) when the removed dep's lock entry is `type: pack` or when any lock entry lists it in `via`; otherwise it keeps the `removeSingleDep` fast path. Interactive multi-remove (`offerRemoveMany`) already runs a full install and needs no change.

*Alternative considered:* always using the full pipeline in `add`/`remove`. Simpler, but it re-resolves every dependency (one manifest fetch + tag lookup per dep), which noticeably slows down the common single-skill case.

### 8. Lock diff and `--frozen`
`lockfile.Diff` stays version-based for the human-readable install summary. `--frozen` additionally treats changes to `type`, `direct`, or `via` as "lock would change", since the lock bytes would change. To avoid failing every CI pipeline right after upgrading, a lock entry with *no* `type` field (written by an older melon) is treated as matching on provenance when the version matches; it is rewritten on the next non-frozen install.

## Risks / Trade-offs

- [Community packs pull skills from many repos the user never looked at] → Members are shown at `add` time, `info` lists them, `list` shows the tree, and `allowed_sources` now covers transitive deps.
- [Tightening policy to transitive deps breaks existing installs under strict policies] → Intended fix for a real gap; call it out in release notes; the error names the `via` chain so the fix (add the pattern or drop the parent) is obvious.
- [Exact pins in packs cause frequent conflicts] → Install warning for packs with exact member pins; conflict message names the pack.
- [Extra manifest fetches in `add`/`info` add latency and rate-limit usage] → One small raw-content request each; uses `GITHUB_TOKEN` when set.
- [Skill-name collisions: two members with the same last path segment map to the same symlink] → Existing behavior for any two deps; not made worse, but more likely with packs. Leave as-is and track separately.
- [Hybrid pack detection relies on the fetched tree] → `ShouldPlace` is the single predicate used by `Place`, `Unplace`, and `list --check` so they never disagree.

## Migration Plan

Additive formats; no migration step. Old manifests default to `type: skill`. Old lock files load (new fields zero-valued) and are rewritten on the next `melon install`. Rollback: older melon binaries ignore the unknown lock and manifest fields; a pure pack would then be placed as an (empty) skill, which is cosmetic.

## Open Questions

- Should `melon list` hide pure packs from the flat (non-tree) view by default, or always show them? Proposed: show them, labeled `(pack)`.
- Should packs be allowed to depend on other packs? Proposed: yes (it falls out of BFS); `add` shows nested packs one level deep.
