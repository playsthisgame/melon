## Context

After Phase 2, `mln install` fetches only the direct dependencies listed in `mln.yaml`. There is no transitive resolution — if `alice/pdf-skill` itself depends on `bob/base-utils`, that dep is silently ignored. The resolver in Phase 3 closes this gap. It also wires up `mln add` (which needs `LatestTag`, already implemented in the fetcher) and adds `--frozen` for CI pipelines.

The existing packages that Phase 3 builds on:
- `internal/fetcher` — `LatestTag(owner, repo)` and `Fetch(dep, storeDir)`
- `internal/lockfile` — `Load`, `Save`, `Diff`
- `internal/store` — `InstalledPath`, `EntrypointPath`
- `internal/manifest` — `Load`, `Save`, `Manifest` struct

## Goals / Non-Goals

**Goals:**
- Implement `internal/resolver.Resolve(manifest)` returning a flat `[]ResolvedDep` for the full transitive graph
- Wire `mln install` to use the resolver instead of iterating `manifest.Dependencies` directly
- Implement `mln add` — resolve latest tag, write `mln.yaml`, run install
- Add `--frozen` to `mln install`
- Unit tests for resolver with `testdata/` fixture manifests

**Non-Goals:**
- PubGrub or SAT-based resolution (greedy MVP only)
- Central registry or name aliasing
- `mln remove` (already scoped to a prior phase)
- `--global` flag or user-level installs

## Decisions

### 1. Resolver fetches transitive `mln.yaml` via GitHub raw content API

**Decision:** Fetch each dependency's manifest at `https://raw.githubusercontent.com/<owner>/<repo>/<tag>/mln.yaml` (or `<subdir>/mln.yaml` for monorepo deps).

**Rationale:** No authentication required for public repos. Fast, cacheable. Consistent with how the fetcher already talks to GitHub. The alternative — cloning the repo first and reading the file — is wasteful for resolution, which only needs the manifest.

**Alternative considered:** Use the GitHub tree API to find the manifest. More complex, no benefit for this use case.

---

### 2. Greedy highest-compatible-version resolution

**Decision:** For each dep encountered in the graph, track the highest version that satisfies all constraints seen so far. Fail with `ErrVersionConflict` if a new constraint is incompatible with the already-selected version.

**Rationale:** Sufficient for MVP. The skill ecosystem is small and version conflicts will be rare. The error message names the conflicting packages, which is actionable for users. PubGrub would provide better backtracking and diagnostics but is a significant implementation investment with little near-term value.

**Alternative considered:** Pick lowest compatible version (more conservative). Rejected — users generally want the latest compatible version, matching Go module behavior.

---

### 3. Resolution visits the graph breadth-first with a visited set

**Decision:** Start from `manifest.Dependencies`, enqueue each dep, fetch its manifest, enqueue its deps, and so on. Track visited `name@resolved-version` pairs to avoid cycles and redundant fetches.

**Rationale:** BFS naturally handles diamond dependencies (two paths to the same dep). The visited set prevents infinite loops in malformed graphs and avoids fetching the same manifest twice.

---

### 4. `mln add` updates `mln.yaml` then runs full install

**Decision:** `mln add <dep>[@constraint]` calls `fetcher.LatestTag` if no version given, writes the dep into `mln.yaml`, then runs the same install logic (resolve → fetch → lock → place).

**Rationale:** Keeps add as a thin wrapper — no separate code path for the fetch/lock/place pipeline. Consistent with how `go get` works.

---

### 5. `--frozen` compares resolved lock to existing lock before writing

**Decision:** When `--frozen` is set, run the full resolve+fetch pipeline, compute what the new `mln.lock` would be, diff it against the file on disk, and exit non-zero if there are any differences. Do not write the new lock file.

**Rationale:** This is the standard CI pattern (npm ci, go mod verify). Catches cases where `mln.yaml` was edited but `mln.lock` was not regenerated.

## Risks / Trade-offs

- **GitHub API rate limits for unauthenticated requests** → For MVP, resolution makes one raw content fetch per transitive dep. Deep graphs on shared CI hosts could hit the 60 req/hr unauthenticated limit. Mitigation: document the `GITHUB_TOKEN` env var support as a follow-up; keep the resolver's HTTP client configurable from the start.

- **Greedy resolution produces surprising results for complex constraint graphs** → A dep that appears first wins the version slot; later constraints can only tighten it. A conflict error points at the two packages but not a resolution path. Mitigation: clear error message naming both constraints and their source packages. PubGrub is the v2 upgrade path.

- **Monorepo `mln.yaml` path** → A dep like `anthropics/skills/skills/skill-creator` has subdir `skills/skill-creator`; its manifest should be at `<subdir>/mln.yaml`. If a monorepo skill doesn't have its own `mln.yaml`, resolution returns no transitive deps for it (not an error). This is intentional for MVP.

## Open Questions

- Should `mln add` fail loudly if the dep already exists in `mln.yaml`, or silently update the constraint? **Proposed:** warn and update.
- Should `--frozen` skip the fetch step entirely and only compare manifests, or run the full pipeline? **Proposed:** run full pipeline to validate tree hashes too, not just versions.
