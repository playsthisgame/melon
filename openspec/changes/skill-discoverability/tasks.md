# Tasks

## 1. Index Client

- [x] 1.1 Create `internal/index` package with a `Fetch() ([]Entry, error)` function that fetches and parses `https://raw.githubusercontent.com/playsthisgame/melon-index/main/index.yml`
- [x] 1.2 Define the `Entry` struct with fields: `Name`, `Description`, `Author`, `Tags []string`, `Featured bool`
- [x] 1.3 Implement `Search(entries []Entry, term string) []Entry` that filters by term against name, description, author, and tags — featured entries first

## 2. GitHub API Client

- [x] 2.1 Create `internal/github` package with an HTTP client struct that reads `GITHUB_TOKEN` from env and sets the `Authorization` header when present
- [x] 2.2 Implement `SearchByTopic(term string) ([]SearchResult, error)` that calls `GET /search/repositories?q=topic:melon-skill+<term>` and returns results
- [x] 2.3 Handle rate limit responses (403/429) with a descriptive error suggesting `GITHUB_TOKEN`
- [x] 2.4 Implement `RepoMeta(owner, repo string) (description string, err error)` that calls `GET /repos/<owner>/<repo>` and returns the about field
- [x] 2.5 Implement `ListTags(owner, repo string) ([]string, error)` that calls `GET /repos/<owner>/<repo>/tags` and returns tag names sorted by semver descending
- [x] 2.6 Implement `ListBranches(owner, repo string) ([]string, error)` as a fallback when no semver tags exist

## 3. Search Results TUI Model

- [x] 3.1 Create `internal/cli/search_model.go` with a bubbletea single-select list model for search results, reusing the existing `list.Model` and lipgloss styles from `init_model.go`
- [x] 3.2 Each list item displays the GitHub path, author, and description; featured items have a visual marker (e.g. `★`)
- [x] 3.3 On Enter, the model returns the selected item's GitHub path; on Escape/Ctrl+C it returns empty
- [x] 3.4 Use the existing `isTTY()` helper to decide between interactive and plain-text output modes

## 4. `mln search` Command

- [x] 4.1 Create `internal/cli/search_cmd.go` with a `newSearchCmd()` function wired into the root command
- [x] 4.2 Validate that at least one argument (search term) is provided; print usage error and exit non-zero otherwise
- [x] 4.3 Fetch and filter the index; if results are found, sort featured first and proceed to display
- [x] 4.4 If the index returns no results or is unreachable, fall back to `github.SearchByTopic(term)` and note in the output that these are community-tagged results
- [x] 4.5 Print a "no results" message when both sources return nothing
- [x] 4.6 In TTY mode, run the search results TUI model; on selection print `mln add <path>` and prompt to run it
- [x] 4.7 In non-TTY mode, print results as plain text one per line: `github.com/<owner>/<repo>   <author>   <description>`

## 5. `mln info` Command

- [x] 5.1 Create `internal/cli/info_cmd.go` with a `newInfoCmd()` function wired into the root command
- [x] 5.2 Validate that exactly one argument (GitHub path) is provided; print usage error and exit non-zero otherwise
- [x] 5.3 Parse the GitHub path to extract owner and repo (strip subpath for API calls, retain subpath for display)
- [x] 5.4 Look up the skill in the fetched index; use index description and author if found, fall back to GitHub repo about field
- [x] 5.5 Fetch tags via the GitHub client; fall back to branches if no semver tags exist
- [x] 5.6 Print skill name, author, description, latest version, all available versions, and subpath (if any)

## 6. README Updates

- [x] 6.1 Add `mln search` and `mln info` to the Commands section with usage examples
- [x] 6.2 Add a "Discovering skills" section explaining `mln search`, the curated index, and the GitHub Topics fallback
- [x] 6.3 Add a "Publishing a skill" section covering both how to submit a PR to `github.com/playsthisgame/melon-index` and how to tag a repo with `melon-skill` for immediate fallback discoverability
