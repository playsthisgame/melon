## Why

Companies with private skill registries need to use melon without exposing public skills to their developers. The current index URL is hardcoded, making it impossible to point melon at an internal registry or restrict search results to vetted, company-approved skills only.

## What Changes

- Add an `index` block to `melon.yaml` with two fields:
  - `url` (string): URL pointing to a custom `index.yaml` file
  - `exclusive` (bool): when `true`, the default melon public index is excluded and only `url` is searched
- `melon search` and `melon info` resolve the active index set from the manifest before fetching
- When `exclusive` is `false` (or omitted), both the custom index and the default index are searched; custom index results appear first

## Capabilities

### New Capabilities
- `custom-index-config`: Configuration of a per-project custom skill index via the `index` block in `melon.yaml`, with optional exclusion of the default public index

### Modified Capabilities
- `skill-search`: Search now resolves one or two index sources depending on manifest config, merging or replacing results accordingly

## Impact

- `internal/manifest/schema.go` — new `IndexConfig` struct and `Index` field on `Manifest`
- `internal/index/index.go` — `Fetch()` accepts a URL parameter instead of using the hardcoded constant
- `internal/cli/search_cmd.go` — reads manifest, resolves index URLs, merges results
- `internal/cli/info_cmd.go` — same index resolution as search
- `README.md` — document the `index` block in the `melon.yaml` schema section and update the `melon search` command description to mention custom index support
