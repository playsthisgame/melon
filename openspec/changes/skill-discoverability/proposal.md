## Why

Users have no way to discover melon skills without already knowing the GitHub path. This creates a chicken-and-egg problem for ecosystem growth: authors publish skills that nobody finds, and users add skills they already know about rather than discovering new ones.

## What Changes

- Create `github.com/playsthisgame/melon-index` — a maintainer-curated `index.yml` listing approved skills with author, description, tags, and an optional featured flag
- Add `mln search <term>` command that first searches the curated index, then falls back to querying GitHub Topics (`melon-skill`) if no index results are found, displaying all results as GitHub paths ready to paste into `mln add`
- Add `mln info <github-path>` command that shows details about a specific skill (description, author, versions) before adding it
- Document the index, the `melon-skill` GitHub topic convention, and how authors can get their skill added in the README

## index.yml structure

```yaml
skills:
  - name: github.com/playsthisgame/skills/agentic-spec-dev
    description: "Spec-driven development workflow for AI coding agents"
    author: playsthisgame
    tags: [workflow, spec]
    featured: true

  - name: github.com/someauthor/cool-skill
    description: "Does something useful"
    author: someauthor
    tags: [productivity]
```

Fields:

- `name` — full GitHub path; exactly what is passed to `mln add`
- `description` — short description of what the skill does
- `author` — GitHub username of the skill author (not necessarily the repo owner)
- `tags` — keywords used for search filtering
- `featured` — optional; curated highlight surfaced first in results

## Capabilities

### New Capabilities

- `skill-search`: Fetch and filter the curated melon-index `index.yml` by keyword; fall back to GitHub Topics API (`melon-skill`) when the index returns no matches; display results as installable paths with author, description, and latest version
- `skill-info`: Fetch and display metadata for a specific skill by GitHub path — description, author, available semver versions, and entrypoint file

### Modified Capabilities

<!-- none -->

## Impact

- New CLI commands: `mln search`, `mln info`
- HTTP fetch of `https://raw.githubusercontent.com/playsthisgame/melon-index/main/index.yml` (no auth, no rate limit concerns)
- GitHub Topics API (`GET /search/repositories?q=topic:melon-skill+<term>`) used as fallback search (unauthenticated, 60 req/hr limit; `GITHUB_TOKEN` env var raises this to 5000/hr)
- GitHub tags API used by `mln info` to resolve versions (reuses existing fetcher logic)
- New repo `github.com/playsthisgame/melon-index` to be created with `index.yml` and contribution guidelines
- README gains a "Discovering skills" section and a "Publishing a skill" section covering both the index submission process and the `melon-skill` topic convention
