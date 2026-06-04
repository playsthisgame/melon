## Context

Melon already pins each dependency by exact version and a SHA-256 `tree_hash` in `melon.lock`, and caches the full file tree of every fetched version under `.melon/<dep@version>/`. `melon outdated` resolves newer versions but only reports version strings; `melon update` applies an upgrade. There is no command between those two that shows the actual file-level content changes a skill will undergo.

The building blocks needed for a diff command already exist:
- `fetcher.LatestMatchingVersion(repoURL, constraint)` — resolves the newest tag satisfying a semver constraint.
- `fetcher.LatestTag(repoURL)` — absolute latest tag.
- `fetcher.Fetch(dep, installDir)` — sparse-checks out a version into a directory and returns the file list + tree hash; idempotent when the tree hash matches.
- `store.InstalledPath(projectDir, dep)` / `store.DirName(name, version)` — cache directory layout (`.melon/<dep@version>/`).
- `lockfile` / `manifest` parsing — to find the current locked version and the dep's constraint.

What is missing is (a) a from/to version selector, (b) a way to materialize both trees, and (c) a renderer that walks both file trees and prints a unified diff.

## Goals / Non-Goals

**Goals:**
- Give users a read-only, file-level view of what changes between the locked version of a dep and a target version, before they commit to `melon update`.
- Reuse existing version-resolution, fetch, and cache machinery — no new fetch path.
- Sensible default target (latest compatible) with an explicit override (`<dep>@<version-or-branch>`).
- A `--stat` summary mode and TTY-aware coloring, consistent with the rest of the CLI.

**Non-Goals:**
- Modifying `melon.yaml` or `melon.lock` (that remains `melon update`'s job).
- Diffing across arbitrary local working directories or git refs unrelated to a melon dep.
- A three-way / interactive merge UI.
- Diffing binary assets meaningfully — binary changes are reported as "binary file changed" without hunks.

## Decisions

### Decision 1: From/to version selection
- **From** = the dep's version in `melon.lock`. If the dep is not in the lock (never installed), error out with a hint to run `melon install` first — there is no meaningful "before" to diff against.
- **To** = the version after the optional `@<target>` in the argument:
  - No target given → resolve via `fetcher.LatestMatchingVersion` using the constraint from `melon.yaml`. For branch-pinned constraints (e.g. `main`), "latest" is ambiguous (a branch is a moving ref), so require an explicit target and error with guidance otherwise.
  - `@<semver>` → use that exact version (validated to exist as a tag).
  - `@<branch>` → resolve the branch's current tree.
- **Alternative considered:** always diff locked-vs-absolute-latest. Rejected — it would ignore the user's constraint and show changes they can't adopt without a constraint bump; the explicit `@` override covers that case.

### Decision 2: Materializing both trees via the existing cache
Both sides are realized as directories under `.melon/`. The locked version is almost always already cached. For the target, construct a `resolver.ResolvedDep` for the target version and call `fetcher.Fetch` into its `store.InstalledPath`; `Fetch` is idempotent (skips when the tree hash already matches), so a previously fetched target is reused. The diff then operates purely on two local directories — no diff-specific network logic.
- **Alternative considered:** diff git trees directly via `git diff <tagA> <tagB> -- <subdir>`. Rejected — melon's tree hash is computed over the sparse subdir contents, not git tree objects, and shelling into git for diff would duplicate the version-resolution and subdir-scoping logic the fetcher already owns. Operating on the cached directories keeps one source of truth.

### Decision 3: Diff rendering in-tree
Implement a small file-tree differ: union the sorted relative file paths of both trees (the `Files` lists are already produced by `Fetch`/`TreeHash`), classify each as added / removed / changed / unchanged by content comparison, and for changed text files emit a unified diff. Use a minimal, well-scoped unified-diff library (or a tiny in-tree LCS renderer) to format hunks. Coloring is applied only when stdout is a TTY and `--no-color` is not set, consistent with existing `cli/tty.go` detection.
- **Alternative considered:** shell out to the system `diff`/`git diff --no-index`. Rejected for portability (Windows builds, `CGO_ENABLED=0`, deterministic output) and to keep formatting under melon's control. A vetted pure-Go diff helper fits the existing "pure Go, no C deps" constraint.

### Decision 4: Fast path via tree hash
Before rendering, compare the two tree hashes (locked hash from the lock file vs. the target's `Fetch` result). If equal, print "No changes" and exit 0 without walking files. This makes the common "already up to date" case cheap and avoids spurious output.

### Decision 5: Output modes and exit codes
- Default: full unified diff per changed file, with add/remove file headers.
- `--stat`: per-file summary (path, +added/-removed line counts) and a totals line; no hunks.
- `--no-color`: disable ANSI; auto-disabled when not a TTY.
- Exit 0 whether or not differences exist (diff is informational, unlike `outdated` which signals staleness with exit 1). Errors (dep not found, not locked, unresolvable target) exit non-zero.

## Risks / Trade-offs

- **[Network fetch as a side effect of a "read" command]** → `melon diff` may fetch the target version into `.melon/`, populating the cache. Mitigation: this matches existing fetch semantics, is idempotent, and is the same artifact `melon update` would fetch; document it as a benign cache warm.
- **[Binary / non-text skill assets]** → unified diffs are meaningless for binaries. Mitigation: detect non-UTF8/NUL-containing files and report "binary file changed" with size deltas instead of hunks.
- **[New diff dependency]** → adds a third-party package. Mitigation: prefer a single small, widely-used pure-Go unified-diff library, or a minimal in-tree renderer if the footprint is small; either keeps `CGO_ENABLED=0`.
- **[Branch-pinned deps]** → "latest" is undefined for a moving branch. Mitigation: require an explicit `@<target>` for branch-constrained deps and emit a clear error otherwise.
- **[Large skills]** → very large trees could produce huge diffs. Mitigation: `--stat` offers a compact view; full diff is opt-in by default but bounded by the skill's own size (skills are markdown-centric and small in practice).
