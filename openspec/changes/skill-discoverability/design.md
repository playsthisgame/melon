# Design

## Context

Melon has no mechanism for discovering skills. Users must know a skill's full GitHub path before they can install it. Discoverability is built around two complementary sources: a maintainer-curated index file at `github.com/playsthisgame/melon-index`, and GitHub Topics as a fallback for skills not yet in the index.

The index is a single `index.yml` fetched at search time from `https://raw.githubusercontent.com/playsthisgame/melon-index/main/index.yml`. It contains only skills that have been reviewed and approved. When a search term returns no index results, `mln search` falls back to querying the GitHub Topics API for repos tagged `melon-skill` — giving authors a way to make their skills discoverable before they're formally added to the index. The fetcher already handles GitHub tag listing, which `mln info` reuses for version resolution.

## Goals / Non-Goals

**Goals:**

- `mln search <term>` — search the curated index first; fall back to GitHub Topics (`melon-skill`) when the index returns no matches
- `mln info <github-path>` — display metadata for a specific skill before adding it, including versions from GitHub tags
- Define the `index.yml` schema (name, description, author, tags, featured)
- Document how to submit a skill to the index and the `melon-skill` topic convention in the README

**Non-Goals:**

- A web-based registry or frontend
- Merging index and GitHub Topics results (Topics is a fallback only, not shown alongside index results)
- Caching the index or search results locally between runs
- Supporting non-GitHub hosts

## Decisions

### Curated index as primary source, GitHub Topics as fallback

The index gives quality-controlled results for the most common searches. GitHub Topics (`melon-skill`) acts as a fallback only when the index returns no matches — giving authors a way to be discoverable before their skill is formally reviewed and added. The two sources are never shown together; if the index has any results for a term, Topics is not queried.

This avoids cluttering curated results with unreviewed skills while still giving the broader ecosystem a discovery path.

### Fetch raw index file, filter client-side

`mln search` fetches the full `index.yml` and filters in-process rather than calling a search API. The index will be small (tens to low hundreds of entries for the foreseeable future), so a full fetch is fast and avoids any server-side infrastructure. `raw.githubusercontent.com` is served by a CDN with no meaningful rate limit for a single YAML file.

### `mln info` uses GitHub tags API for versions

The index stores human-curated metadata but not version lists — those change with every release and would make the index stale immediately. `mln info` fetches version information live from the GitHub tags API, reusing the existing fetcher's tag resolution logic. If the skill is in the index, its description and author come from there; if not, they fall back to the GitHub repo's about field.

### Featured entries sort first in results

Entries with `featured: true` are surfaced at the top of `mln search` output. This gives the maintainer a lightweight curation mechanism without building a ranking algorithm. Non-featured results follow in index order.

### Author field is the skill author, not necessarily the repo owner

The `author` field records who created the skill, which may differ from the repo owner (e.g., a fork, a monorepo maintained by an org). This is a free-text GitHub username, not validated at search time.

## Risks / Trade-offs

- **Index staleness**: If a skill is removed from GitHub or its path changes, the index entry becomes stale. → Mitigation: `mln info` verifies the path is reachable and warns if not; periodic index maintenance via PRs.
- **PR bottleneck**: Adding a skill to the curated index requires a review. → The GitHub Topics fallback means authors aren't blocked from being discoverable while waiting for review.
- **Rate limiting on Topics fallback**: Unauthenticated GitHub API allows 60 req/hr. Topics is only queried when the index misses, so this is rarely hit in practice. → Display a clear error when rate-limited and suggest setting `GITHUB_TOKEN`.
- **Single point of failure**: If `raw.githubusercontent.com` is unreachable, index search fails. → `mln search` falls through to GitHub Topics in this case. `mln add` and `mln install` are unaffected.

## Migration Plan

No migration needed. Both commands are additive. The `melon-index` repo is a new external resource with no impact on existing melon.yml or lock files.
