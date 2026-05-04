package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runCleanCmd(t *testing.T, dir string) (string, error) {
	t.Helper()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := runClean(cmd, nil)
	return buf.String(), err
}

// TestClean_NoLockFile verifies that clean exits cleanly when melon.lock is absent.
func TestClean_NoLockFile(t *testing.T) {
	dir := t.TempDir()

	out, err := runCleanCmd(t, dir)

	require.NoError(t, err)
	assert.Contains(t, out, "No melon.lock found")
}

// TestClean_NothingToClean verifies the no-op path when .melon/ matches the lock.
func TestClean_NothingToClean(t *testing.T) {
	dir := t.TempDir()

	dep := lockfile.LockedDep{
		Name:       "github.com/alice/skills/skill-a",
		Version:    "1.0.0",
		GitTag:     "v1.0.0",
		RepoURL:    "https://github.com/alice/skills",
		Subdir:     "skill-a",
		Entrypoint: "SKILL.md",
	}
	writeLockfile(t, dir, []lockfile.LockedDep{dep})

	// Create a matching cache entry.
	cacheDir := filepath.Join(dir, store.StoreDir, store.DirName(dep.Name, dep.Version))
	require.NoError(t, os.MkdirAll(cacheDir, 0755))

	out, err := runCleanCmd(t, dir)

	require.NoError(t, err)
	assert.Contains(t, out, "Nothing to clean")
}

// TestClean_OrphanedEntryRemoved verifies that a cache dir not in the lock is deleted.
func TestClean_OrphanedEntryRemoved(t *testing.T) {
	dir := t.TempDir()

	// Lock contains only skill-b.
	lockedDep := lockfile.LockedDep{
		Name:       "github.com/alice/skills/skill-b",
		Version:    "2.0.0",
		Entrypoint: "SKILL.md",
	}
	writeLockfile(t, dir, []lockfile.LockedDep{lockedDep})

	// Cache has both skill-a (orphan) and skill-b (locked).
	orphanDir := filepath.Join(dir, store.StoreDir, "github.com-alice-skills-skill-a@1.0.0")
	require.NoError(t, os.MkdirAll(orphanDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(orphanDir, "SKILL.md"), []byte("orphan"), 0644))

	lockedDir := filepath.Join(dir, store.StoreDir, store.DirName(lockedDep.Name, lockedDep.Version))
	require.NoError(t, os.MkdirAll(lockedDir, 0755))

	out, err := runCleanCmd(t, dir)

	require.NoError(t, err)
	assert.Contains(t, out, "removed")
	assert.Contains(t, out, "1 cache entr")

	// Orphan must be gone.
	_, statErr := os.Stat(orphanDir)
	assert.True(t, os.IsNotExist(statErr), "orphaned cache dir should be deleted")

	// Locked entry must still exist.
	_, statErr = os.Stat(lockedDir)
	assert.NoError(t, statErr, "locked cache dir must not be touched")
}

// TestClean_SymlinkRemovedWithCacheEntry verifies that orphaned symlinks in agent
// skill dirs are removed alongside the orphaned cache entry.
func TestClean_SymlinkRemovedWithCacheEntry(t *testing.T) {
	dir := t.TempDir()

	// Empty lock — everything in .melon/ is orphaned.
	writeLockfile(t, dir, []lockfile.LockedDep{})

	// Manifest declares claude-code so placer knows the skill dir.
	writeManifest(t, dir, map[string]string{})

	// Create an orphaned cache entry.
	orphanDir := filepath.Join(dir, store.StoreDir, "github.com-alice-skills-skill-a@1.0.0")
	require.NoError(t, os.MkdirAll(orphanDir, 0755))

	// Create the corresponding symlink in .claude/skills/.
	skillsDir := filepath.Join(dir, ".claude", "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))
	linkPath := filepath.Join(skillsDir, "skill-a")
	require.NoError(t, os.Symlink(orphanDir, linkPath))

	out, err := runCleanCmd(t, dir)

	require.NoError(t, err)
	assert.Contains(t, out, "removed")

	// Cache entry must be gone.
	_, cacheErr := os.Stat(orphanDir)
	assert.True(t, os.IsNotExist(cacheErr), "orphaned cache dir should be deleted")

	// Symlink must be gone.
	_, linkErr := os.Lstat(linkPath)
	assert.True(t, os.IsNotExist(linkErr), "orphaned symlink should be removed")

	// Warning about symlinks should NOT appear (manifest was present).
	assert.False(t, strings.Contains(out, "Warning"), "no warning expected when manifest is present")
}
