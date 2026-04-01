package lockfile

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LockFile is the in-memory representation of mln.lock.
type LockFile struct {
	GeneratedAt  string      `yaml:"generated_at"`
	Dependencies []LockedDep `yaml:"dependencies"`
}

// LockedDep is one entry in mln.lock.
type LockedDep struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	GitTag     string `yaml:"git_tag"`
	RepoURL    string `yaml:"repo_url"`
	Entrypoint string `yaml:"entrypoint"`
	Checksum   string `yaml:"checksum"`
}

// LockDiff describes what changed between two LockFile snapshots.
type LockDiff struct {
	Added   []LockedDep
	Removed []LockedDep
	Updated []LockedDep // present in both but version changed
}

// Load reads and parses an mln.lock file at path.
func Load(path string) (LockFile, error) {
	// TODO: implement Load
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
	// TODO: implement Save
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
	// TODO: implement Diff
	// Build maps by dep name for O(n) comparison.
	// Added: in new but not in old.
	// Removed: in old but not in new.
	// Updated: in both but with different versions.
	return LockDiff{}
}
