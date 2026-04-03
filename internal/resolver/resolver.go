package resolver

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/pkg/semver"
)

// ErrConflict is returned when two packages require incompatible versions of a
// shared transitive dependency.
var ErrConflict = errors.New("resolver: version conflict")

// ResolvedDep is a single dependency with its pinned version.
type ResolvedDep struct {
	Name       string // full dep name, e.g. "github.com/anthropics/skills/skills/skill-creator"
	Version    string // exact pinned semver, e.g. "1.3.1"
	RepoURL    string // https://github.com/anthropics/skills  (repo root, derived from name)
	Subdir     string // subdirectory within the repo, e.g. "skills/skill-creator" (empty if repo root)
	GitTag     string // e.g. "v1.3.1"
	Entrypoint string // path to SKILL.md relative to subdir root, e.g. "SKILL.md"
	TreeHash   string // SHA256 of the full directory tree at this tag (sorted file paths)
}

// Resolve builds a flat, sorted list of all transitive dependencies from m.
//
// resolveVersion is called to convert a (repoURL, constraint) pair into a
// pinned (version, gitTag). Typically set to fetcher.LatestMatchingVersion.
//
// fetchManifest fetches the mln.yaml for a dep from its source. It should
// return an empty Manifest and nil error when the file is absent (404).
// Typically set to DefaultFetchManifest.
//
// Uses a greedy highest-compatible-version strategy. Returns ErrConflict if
// two packages require incompatible versions of a shared transitive dep.
func Resolve(
	m manifest.Manifest,
	resolveVersion func(repoURL, constraint string) (string, string, error),
	fetchManifest func(repoURL, gitTag, subdir string) (manifest.Manifest, error),
) ([]ResolvedDep, error) {
	type selectedEntry struct {
		version    string
		gitTag     string
		constraint string
		source     string
		entrypoint string
	}

	type queueItem struct {
		name       string
		constraint string
		source     string
	}

	selected := make(map[string]selectedEntry)

	// Sort direct deps for a deterministic BFS order.
	directNames := sortedStringKeys(m.Dependencies)
	queue := make([]queueItem, 0, len(directNames))
	for _, name := range directNames {
		queue = append(queue, queueItem{
			name:       name,
			constraint: m.Dependencies[name],
			source:     "root manifest",
		})
	}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if entry, exists := selected[item.name]; exists {
			// Already have a version selected — check that it satisfies the new constraint.
			if !semver.IsCompatible(item.constraint, entry.version) {
				return nil, fmt.Errorf("%w: %s requires %s %q but %s already selected %s %q",
					ErrConflict,
					item.source, item.name, item.constraint,
					entry.source, item.name, entry.constraint)
			}
			continue
		}

		// First encounter — resolve to a pinned version.
		repoURL, subdir := splitDepName(item.name)
		version, gitTag, err := resolveVersion(repoURL, item.constraint)
		if err != nil {
			return nil, fmt.Errorf("resolver: resolve %s: %w", item.name, err)
		}

		// Fetch the dep's manifest to discover transitive deps and entrypoint.
		depManifest, err := fetchManifest(repoURL, gitTag, subdir)
		if err != nil {
			return nil, fmt.Errorf("resolver: fetch manifest for %s@%s: %w", item.name, version, err)
		}

		entrypoint := "SKILL.md"
		if depManifest.Entrypoint != "" {
			entrypoint = depManifest.Entrypoint
		}

		selected[item.name] = selectedEntry{
			version:    version,
			gitTag:     gitTag,
			constraint: item.constraint,
			source:     item.source,
			entrypoint: entrypoint,
		}

		// Enqueue transitive deps in sorted order for determinism.
		for _, depName := range sortedStringKeys(depManifest.Dependencies) {
			queue = append(queue, queueItem{
				name:       depName,
				constraint: depManifest.Dependencies[depName],
				source:     item.name,
			})
		}
	}

	// Build a sorted result slice.
	allNames := make([]string, 0, len(selected))
	for name := range selected {
		allNames = append(allNames, name)
	}
	sort.Strings(allNames)

	result := make([]ResolvedDep, 0, len(selected))
	for _, name := range allNames {
		entry := selected[name]
		repoURL, subdir := splitDepName(name)
		result = append(result, ResolvedDep{
			Name:       name,
			Version:    entry.version,
			RepoURL:    repoURL,
			Subdir:     subdir,
			GitTag:     entry.gitTag,
			Entrypoint: entry.entrypoint,
		})
	}
	return result, nil
}

// DefaultFetchManifest fetches a dependency's mln.yaml from the GitHub raw
// content API. Returns an empty Manifest and nil error when the file is absent
// (HTTP 404) — missing manifests are treated as deps with no transitive deps.
// Respects the GITHUB_TOKEN environment variable if set.
func DefaultFetchManifest(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
	// repoURL is "https://github.com/owner/repo"
	ownerRepo := strings.TrimPrefix(repoURL, "https://github.com/")
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s", ownerRepo, gitTag)
	if subdir != "" {
		rawURL += "/" + subdir
	}
	rawURL += "/mln.yaml"

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("resolver: build request for %s: %w", rawURL, err)
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("resolver: fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Missing manifest is not an error — treat as dep with no transitive deps.
		return manifest.Manifest{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return manifest.Manifest{}, fmt.Errorf("resolver: fetch %s: status %d", rawURL, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("resolver: read %s: %w", rawURL, err)
	}

	var m manifest.Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return manifest.Manifest{}, fmt.Errorf("resolver: parse %s: %w", rawURL, err)
	}
	return m, nil
}

// splitDepName converts a dep name into a repo root URL and optional subdir.
// This is a local variant of fetcher.ParseDepName to avoid import cycles
// (fetcher imports resolver for the ResolvedDep type).
//
//	"github.com/owner/repo"             → ("https://github.com/owner/repo", "")
//	"github.com/owner/repo/sub/dir"     → ("https://github.com/owner/repo", "sub/dir")
func splitDepName(name string) (repoURL, subdir string) {
	// Strip https:// or http:// prefix so users can paste full GitHub URLs.
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")

	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return "https://" + name, ""
	}
	repoURL = "https://" + strings.Join(parts[:3], "/")
	rest := parts[3:]
	// Strip GitHub web UI path segments: tree/<branch>/ or blob/<branch>/
	if len(rest) >= 2 && (rest[0] == "tree" || rest[0] == "blob") {
		rest = rest[2:]
	}
	if len(rest) > 0 {
		subdir = strings.Join(rest, "/")
	}
	return repoURL, subdir
}

// sortedStringKeys returns the keys of a map[string]string in sorted order.
func sortedStringKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
