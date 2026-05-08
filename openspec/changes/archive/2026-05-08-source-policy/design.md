## Context

Companies adopting melon internally need guardrails over which GitHub repositories developers can pull skills from. Without a policy mechanism, nothing prevents a developer from adding `github.com/any-random/repo` to `melon.yaml` — intentionally or not. The `index` block (already implemented) restricts what appears in search, but does not restrict what can actually be installed.

The `policy` block addresses the installation side: an allowlist of glob patterns checked by both `melon add` (fail-fast) and `melon install` (catch hand-edited manifests).

## Goals / Non-Goals

**Goals:**
- Allow a project to declare an allowlist of permitted dependency source patterns in `melon.yaml`
- Enforce the allowlist in `melon add` before writing to the manifest
- Enforce the allowlist in `melon install` before any network requests are made
- Backwards-compatible: absent `policy` block = no restrictions

**Non-Goals:**
- Cryptographic signing or tamper-proof enforcement — this is a guardrail, not a hard security boundary
- Per-dependency overrides or exceptions within the allowlist
- Integration with external policy services or OPA

## Decisions

### Schema: top-level `policy` block, not nested under `index`

`index` controls discoverability (search). `policy` controls installation. They are separate concerns and can be configured independently — a team might want a custom search index without restricting sources, or restrict sources without a custom index.

**Alternative considered:** Nesting `allowed_sources` inside `index`. Rejected because `index` is a convenience feature (better search), while `policy` implies enforcement. Merging them conflates intent.

### Pattern matching: `filepath.Match`-style glob on the full dep path

Each `allowed_sources` entry is matched against the full dependency path (e.g. `github.com/owner/repo/sub/dir`) using Go's `path.Match`. A trailing `*` matches any suffix. This is simple, predictable, and consistent with how `.gitignore` patterns work in the ecosystem.

**Alternative considered:** Regex patterns. Rejected — regexes are error-prone for non-technical users. Globs are the expected UX for path filtering.

**Alternative considered:** Prefix-only matching without wildcards. Rejected — too coarse. An exact entry like `github.com/my-company/approved-skill` should not implicitly permit `github.com/my-company/other-skill`.

### Enforcement in both `melon add` and `melon install`

`melon add` enforces at the earliest possible point (fail fast, don't touch `melon.yaml`). `melon install` enforces before any network activity as a second layer, catching deps added by hand-editing the manifest. Both are needed; neither alone is sufficient.

**Alternative considered:** Only enforce in `melon add`. Rejected — a developer can bypass by editing `melon.yaml` directly.

**Alternative considered:** Only enforce in `melon install`. Rejected — worse UX; errors should surface as early as possible.

### PolicyConfig uses `*bool`-style pointer for future extensibility

The `PolicyConfig` struct starts with just `AllowedSources []string`. Using a pointer (`*PolicyConfig`) on the `Manifest` means an absent `policy` block deserializes to `nil` and is cleanly omitted on save — consistent with how `IndexConfig` and `Vendor` are handled.

## Risks / Trade-offs

- **Not a hard security boundary** → A determined developer can edit `melon.yaml` on a machine without the policy (e.g. a different project), add deps, and commit. Mitigation: enforce in CI using `melon install --frozen` plus a policy check in the pipeline.
- **Glob semantics may surprise users** → `github.com/my-company/*` does NOT match `github.com/my-company` (no trailing slash). Document clearly with examples. Mitigation: clear error messages that show what pattern was checked.
- **No migration needed** → `policy` block is optional and additive. Existing `melon.yaml` files are unaffected.

## Open Questions

- Should `melon install` list ALL blocked dependencies in a single error, or stop at the first? (Recommendation: list all — better UX for auditing a newly-imported manifest.)
