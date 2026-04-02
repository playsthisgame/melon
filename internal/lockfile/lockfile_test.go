package lockfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dep1 = lockfile.LockedDep{
	Name: "github.com/alice/pdf-skill", Version: "1.2.0",
	GitTag: "v1.2.0", RepoURL: "https://github.com/alice/pdf-skill",
	Subdir: "", Entrypoint: "SKILL.md",
	TreeHash: "sha256:aaa", Files: []string{"SKILL.md"},
}
var dep2 = lockfile.LockedDep{
	Name: "github.com/alice/xlsx-skill", Version: "2.0.0",
	GitTag: "v2.0.0", RepoURL: "https://github.com/alice/xlsx-skill",
	Subdir: "", Entrypoint: "SKILL.md",
	TreeHash: "sha256:bbb", Files: []string{"SKILL.md"},
}
var dep2v2 = lockfile.LockedDep{
	Name: "github.com/alice/xlsx-skill", Version: "2.1.0",
	GitTag: "v2.1.0", RepoURL: "https://github.com/alice/xlsx-skill",
	Subdir: "", Entrypoint: "SKILL.md",
	TreeHash: "sha256:ccc", Files: []string{"SKILL.md"},
}

func TestDiff_Added(t *testing.T) {
	old := lockfile.LockFile{}
	new := lockfile.LockFile{Dependencies: []lockfile.LockedDep{dep1}}

	d := lockfile.Diff(old, new)
	assert.Equal(t, []lockfile.LockedDep{dep1}, d.Added)
	assert.Empty(t, d.Removed)
	assert.Empty(t, d.Updated)
}

func TestDiff_Removed(t *testing.T) {
	old := lockfile.LockFile{Dependencies: []lockfile.LockedDep{dep1}}
	new := lockfile.LockFile{}

	d := lockfile.Diff(old, new)
	assert.Empty(t, d.Added)
	assert.Equal(t, []lockfile.LockedDep{dep1}, d.Removed)
	assert.Empty(t, d.Updated)
}

func TestDiff_Updated(t *testing.T) {
	old := lockfile.LockFile{Dependencies: []lockfile.LockedDep{dep2}}
	new := lockfile.LockFile{Dependencies: []lockfile.LockedDep{dep2v2}}

	d := lockfile.Diff(old, new)
	assert.Empty(t, d.Added)
	assert.Empty(t, d.Removed)
	assert.Equal(t, []lockfile.LockedDep{dep2v2}, d.Updated)
}

func TestDiff_NoChange(t *testing.T) {
	lf := lockfile.LockFile{Dependencies: []lockfile.LockedDep{dep1, dep2}}

	d := lockfile.Diff(lf, lf)
	assert.Empty(t, d.Added)
	assert.Empty(t, d.Removed)
	assert.Empty(t, d.Updated)
}

func TestDiff_Mixed(t *testing.T) {
	old := lockfile.LockFile{Dependencies: []lockfile.LockedDep{dep1, dep2}}
	new := lockfile.LockFile{Dependencies: []lockfile.LockedDep{dep2v2}}

	d := lockfile.Diff(old, new)
	assert.Equal(t, []lockfile.LockedDep{dep1}, d.Removed)
	assert.Equal(t, []lockfile.LockedDep{dep2v2}, d.Updated)
	assert.Empty(t, d.Added)
}

func TestRoundTrip(t *testing.T) {
	lf := lockfile.LockFile{
		GeneratedAt:  "2025-03-31T12:00:00Z",
		Dependencies: []lockfile.LockedDep{dep1, dep2},
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.lock")

	require.NoError(t, lockfile.Save(lf, path))
	loaded, err := lockfile.Load(path)
	require.NoError(t, err)

	assert.Equal(t, lf.GeneratedAt, loaded.GeneratedAt)
	assert.Equal(t, lf.Dependencies, loaded.Dependencies)
}

func TestLoad_NewFields(t *testing.T) {
	// Verify subdir, tree_hash, and files round-trip correctly.
	dep := lockfile.LockedDep{
		Name:       "github.com/anthropics/skills/skills/skill-creator",
		Version:    "1.0.0",
		GitTag:     "v1.0.0",
		RepoURL:    "https://github.com/anthropics/skills",
		Subdir:     "skills/skill-creator",
		Entrypoint: "SKILL.md",
		TreeHash:   "sha256:abc123",
		Files:      []string{"SKILL.md", "references/api.md"},
	}
	lf := lockfile.LockFile{GeneratedAt: "2025-03-31T12:00:00Z", Dependencies: []lockfile.LockedDep{dep}}

	dir := t.TempDir()
	path := filepath.Join(dir, "melon.lock")
	require.NoError(t, lockfile.Save(lf, path))

	loaded, err := lockfile.Load(path)
	require.NoError(t, err)
	require.Len(t, loaded.Dependencies, 1)

	got := loaded.Dependencies[0]
	assert.Equal(t, dep.Subdir, got.Subdir)
	assert.Equal(t, dep.TreeHash, got.TreeHash)
	assert.Equal(t, dep.Files, got.Files)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := lockfile.Load("/nonexistent/melon.lock")
	assert.Error(t, err)
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.lock")
	require.NoError(t, lockfile.Save(lockfile.LockFile{}, path))
	_, err := os.Stat(path)
	assert.NoError(t, err)
}
