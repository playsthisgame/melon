## Context

The melon index URL is hardcoded as a constant in `internal/index/index.go`. All search and info commands fetch from this single public URL. There is no mechanism for projects to point at a different registry or suppress the public one.

The change is narrow: a new optional block in `melon.yaml`, a signature change on `Fetch()`, and updated callers in the two CLI commands that touch the index.

## Goals / Non-Goals

**Goals:**
- Allow a project to configure a custom index URL in `melon.yaml`
- Allow a project to suppress the default public index entirely (`exclusive: true`)
- When both indices are active, merge results with the custom index results first

**Non-Goals:**
- User-level or system-level config files (`~/.melon/config.yaml`) — out of scope for this change
- Environment variable override (`MELON_INDEX_URL`) — out of scope for this change
- Authentication headers for private index URLs — the existing `GITHUB_TOKEN` pattern covers GitHub-hosted indices; arbitrary auth is out of scope
- Caching of index responses — separate concern

## Decisions

**Nested `index` block rather than top-level fields**

A nested struct (`index.url`, `index.exclusive`) groups related fields and leaves room to add future index options (e.g. auth, ttl) without polluting the top-level manifest namespace. Alternatives considered: two top-level fields (`index_url`, `index_exclusive`) — rejected as less extensible and inconsistent with how similar tools (Helm, Cargo) handle registry config.

**`Fetch(url string)` — caller supplies the URL**

Changing `Fetch` to accept a URL rather than reading a package-level constant keeps the `index` package stateless and easy to test. Alternative: a package-level setter — rejected as global mutable state.

**Merge strategy: custom first, then public**

When `exclusive` is false, results from the custom index are prepended to results from the public index. This surfaces company-specific skills at the top of search without hiding public options. Duplicate names (same `name` field) from the public index are suppressed if already present in the custom results.

**`exclusive` defaults to `false`**

Omitting the field keeps existing behaviour — both indices are searched. Opting in to exclusivity requires an explicit `exclusive: true`. This is the safe default.

## Risks / Trade-offs

- [URL is committed to the repo] A private index URL is visible to anyone with repo access. → Acceptable: the URL itself is not a secret; access control lives on the index server.
- [No auth for arbitrary URLs] A non-GitHub index URL with HTTP auth would not be supported. → Mitigation: document that GitHub-hosted private index repos work via `GITHUB_TOKEN`; arbitrary auth deferred to a future change.
- [Manifest load failure in search/info] If no `melon.yaml` exists in the current directory, the command should fall back to the public index rather than erroring. → Handle gracefully: treat a missing manifest as no `index` config.

## Migration Plan

No migration required. The `index` block is optional; existing `melon.yaml` files without it continue to use the public index unchanged.
