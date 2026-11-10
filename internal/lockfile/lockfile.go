package lockfile

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LockFile is the in-memory representation of melon.lock.
type LockFile struct {
	GeneratedAt  string      `yaml:"generated_at"`
	Dependencies []LockedDep `yaml:"dependencies"`
}

// LockedDep is one entry in melon.lock.
type LockedDep struct {
	Name       string   `yaml:"name"`
	Version    string   `yaml:"version"`
	GitTag     string   `yaml:"git_tag"`
	RepoURL    string   `yaml:"repo_url"`
	Subdir     string   `yaml:"subdir"`
	Entrypoint string   `yaml:"entrypoint"`
	TreeHash   string   `yaml:"tree_hash"`
	Files      []string `yaml:"files"`
	// Type is "skill" or "pack". Empty in lock files written before packs
	// existed; treat empty as "skill".
	Type string `yaml:"type,omitempty"`
	// Direct is true when the dep is listed in the project's melon.yaml.
	Direct bool `yaml:"direct,omitempty"`
	// Via lists, sorted, the resolved deps whose manifests declare this dep.
	Via []string `yaml:"via,omitempty"`
}

// IsPack reports whether the locked dep is a pack.
func (d LockedDep) IsPack() bool {
	return d.Type == "pack"
}

// LockDiff describes what changed between two LockFile snapshots.
type LockDiff struct {
	Added   []LockedDep
	Removed []LockedDep
	Updated []LockedDep // present in both but version changed
}

// Load reads and parses a melon.lock file at path.
func Load(path string) (LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LockFile{}, fmt.Errorf("lockfile: read %s: %w", path, err)
	}
	var lf LockFile
	if err := yaml.Unmarshal(data, &lf); err != nil {
		return LockFile{}, fmt.Errorf("lockfile: parse %s: %w", path, err)
	}
	return lf, nil
}

// Save serializes lf to YAML and writes it to path.
func Save(lf LockFile, path string) error {
	data, err := yaml.Marshal(lf)
	if err != nil {
		return fmt.Errorf("lockfile: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("lockfile: write %s: %w", path, err)
	}
	return nil
}

// Diff computes the set of added, removed, and updated deps between old and new.
func Diff(old, new LockFile) LockDiff {
	oldMap := make(map[string]LockedDep, len(old.Dependencies))
	for _, d := range old.Dependencies {
		oldMap[d.Name] = d
	}
	newMap := make(map[string]LockedDep, len(new.Dependencies))
	for _, d := range new.Dependencies {
		newMap[d.Name] = d
	}

	var diff LockDiff
	for name, newDep := range newMap {
		if oldDep, exists := oldMap[name]; !exists {
			diff.Added = append(diff.Added, newDep)
		} else if oldDep.Version != newDep.Version {
			diff.Updated = append(diff.Updated, newDep)
		}
	}
	for name, oldDep := range oldMap {
		if _, exists := newMap[name]; !exists {
			diff.Removed = append(diff.Removed, oldDep)
		}
	}
	return diff
}

// ProvenanceChanged reports whether any dep present in both old and new has a
// different type, direct flag, or via list. Entries in old with an empty Type
// were written before provenance existed and are treated as matching, so the
// first --frozen run after upgrading melon does not fail spuriously.
func ProvenanceChanged(old, new LockFile) bool {
	oldMap := make(map[string]LockedDep, len(old.Dependencies))
	for _, d := range old.Dependencies {
		oldMap[d.Name] = d
	}
	for _, n := range new.Dependencies {
		o, ok := oldMap[n.Name]
		if !ok || o.Type == "" {
			continue
		}
		if o.Type != n.Type || o.Direct != n.Direct || !equalStrings(o.Via, n.Via) {
			return true
		}
	}
	return false
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
