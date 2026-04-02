package resolver

import (
	"errors"

	"github.com/playsthisgame/melon/internal/manifest"
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

// Resolve fetches each dependency's mln.yaml transitively, builds a directed
// acyclic graph, and returns a flat list of pinned versions.
//
// Uses a greedy highest-compatible-version strategy for MVP.
// Returns ErrConflict if two packages require incompatible versions of a
// shared transitive dep.
func Resolve(m manifest.Manifest) ([]ResolvedDep, error) {
	// TODO: implement transitive resolution
	// 1. For each direct dep in m.Dependencies, fetch its mln.yaml from GitHub.
	// 2. Recursively collect all transitive deps.
	// 3. For each dep that appears multiple times, select the highest version
	//    that satisfies all constraints using pkg/semver.IsCompatible.
	// 4. If no version satisfies all constraints, return ErrConflict.
	// 5. Return the flat resolved list.
	return nil, nil
}
