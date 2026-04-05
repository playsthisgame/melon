package placer_test

// Integration test: verifies that the pruning flow (Unplace + store.Remove)
// correctly removes a stale dep while leaving the remaining dep intact.
// This exercises the combination used by install_cmd after lockfile.Diff.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/placer"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPruning_InstallRemovesDep simulates the pruning step that install runs
// after a dep is removed from melon.yml.
//
// Setup:
//   - Two deps placed and cached: alice/pdf-skill and bob/base-utils
//   - bob/base-utils is removed from the manifest (simulates user editing melon.yml)
//
// Expectation:
//   - bob/base-utils symlink and cache entry are gone
//   - alice/pdf-skill symlink and cache entry are still present
func TestPruning_InstallRemovesDep(t *testing.T) {
	projectDir := t.TempDir()

	depA := makeDep(t, projectDir, "github.com/alice/pdf-skill", "1.2.0")
	depB := makeDep(t, projectDir, "github.com/bob/base-utils", "0.5.0")

	m := manifest.Manifest{ToolCompat: []string{"claude-code"}}

	// Place both deps.
	require.NoError(t, placer.Place([]resolver.ResolvedDep{depA, depB}, m, projectDir, &bytes.Buffer{}))

	// Verify both symlinks and cache dirs exist before pruning.
	linkA := filepath.Join(projectDir, ".claude/skills/pdf-skill")
	linkB := filepath.Join(projectDir, ".claude/skills/base-utils")
	_, err := os.Lstat(linkA)
	require.NoError(t, err, "pdf-skill symlink must exist before pruning")
	_, err = os.Lstat(linkB)
	require.NoError(t, err, "base-utils symlink must exist before pruning")
	require.DirExists(t, store.InstalledPath(projectDir, depA))
	require.DirExists(t, store.InstalledPath(projectDir, depB))

	// Simulate lockfile.Diff: old lock had both, new lock only has depA.
	oldLock := lockfile.LockFile{
		Dependencies: []lockfile.LockedDep{
			{Name: depA.Name, Version: depA.Version},
			{Name: depB.Name, Version: depB.Version},
		},
	}
	newLock := lockfile.LockFile{
		Dependencies: []lockfile.LockedDep{
			{Name: depA.Name, Version: depA.Version},
		},
	}
	diff := lockfile.Diff(oldLock, newLock)
	require.Len(t, diff.Removed, 1)
	assert.Equal(t, depB.Name, diff.Removed[0].Name)

	// Run pruning: unplace then remove store entry for removed deps.
	require.NoError(t, placer.Unplace(diff.Removed, m, projectDir, &bytes.Buffer{}))
	for _, dep := range diff.Removed {
		require.NoError(t, store.Remove(projectDir, resolver.ResolvedDep{Name: dep.Name, Version: dep.Version}))
	}

	// bob/base-utils: symlink and cache dir must be gone.
	_, err = os.Lstat(linkB)
	assert.True(t, os.IsNotExist(err), "base-utils symlink must be removed after pruning")
	assert.NoDirExists(t, store.InstalledPath(projectDir, depB), "base-utils cache dir must be removed after pruning")

	// alice/pdf-skill: symlink and cache dir must still be present.
	_, err = os.Lstat(linkA)
	assert.NoError(t, err, "pdf-skill symlink must still exist after pruning")
	assert.DirExists(t, store.InstalledPath(projectDir, depA), "pdf-skill cache dir must still exist after pruning")
}
