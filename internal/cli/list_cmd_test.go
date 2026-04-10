package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func writeLockFile(t *testing.T, dir string, deps []lockfile.LockedDep) {
	t.Helper()
	lf := lockfile.LockFile{
		GeneratedAt:  "2025-01-01T00:00:00Z",
		Dependencies: deps,
	}
	data, err := yaml.Marshal(lf)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.lock"), data, 0644))
}

func runListCmd(t *testing.T, dir string, pending, check bool) (string, error) {
	t.Helper()

	origDir := flagDir
	origPending := flagPending
	origCheck := flagCheck
	t.Cleanup(func() {
		flagDir = origDir
		flagPending = origPending
		flagCheck = origCheck
	})
	flagDir = dir
	flagPending = pending
	flagCheck = check

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runList(cmd, nil)
	return buf.String(), err
}

func TestList_InstalledSkills(t *testing.T) {
	dir := t.TempDir()
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/skill-b", Version: "1.0.0"},
		{Name: "github.com/owner/skill-a", Version: "2.0.0"},
	})

	out, err := runListCmd(t, dir, false, false)
	require.NoError(t, err)
	assert.Contains(t, out, "github.com/owner/skill-a")
	assert.Contains(t, out, "github.com/owner/skill-b")
	// skill-a should appear before skill-b (sorted)
	assert.Less(t, indexOf(out, "skill-a"), indexOf(out, "skill-b"))
}

func TestList_NoLockFile(t *testing.T) {
	dir := t.TempDir()

	out, err := runListCmd(t, dir, false, false)
	require.NoError(t, err)
	assert.Contains(t, out, "No skills installed.")
}

func TestList_EmptyLockFile(t *testing.T) {
	dir := t.TempDir()
	writeLockFile(t, dir, nil)

	out, err := runListCmd(t, dir, false, false)
	require.NoError(t, err)
	assert.Contains(t, out, "No skills installed.")
}

func TestList_PendingSkillsExist(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, map[string]string{
		"github.com/owner/installed-skill": "^1.0.0",
		"github.com/owner/pending-skill":   "^2.0.0",
	})
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/installed-skill", Version: "1.0.0"},
	})

	out, err := runListCmd(t, dir, true, false)
	require.NoError(t, err)
	assert.Contains(t, out, "Pending (not installed):")
	assert.Contains(t, out, "github.com/owner/pending-skill")
	assert.NotContains(t, out, "github.com/owner/installed-skill\n") // installed skill not in pending section
}

func TestList_NoPendingSkills(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, map[string]string{
		"github.com/owner/skill-a": "^1.0.0",
	})
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/skill-a", Version: "1.0.0"},
	})

	out, err := runListCmd(t, dir, true, false)
	require.NoError(t, err)
	assert.Contains(t, out, "No pending skills.")
}

func TestList_CheckMissingSymlink(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, map[string]string{
		"github.com/owner/skill-a": "^1.0.0",
	})
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/skill-a", Version: "1.0.0"},
	})

	// No symlink created — check should report MISSING and return error.
	out, err := runListCmd(t, dir, false, true)
	require.Error(t, err)
	assert.Contains(t, out, "MISSING")
	assert.Contains(t, out, "skill-a")
}

func TestList_CheckSymlinkPresent(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, map[string]string{
		"github.com/owner/skill-a": "^1.0.0",
	})
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/skill-a", Version: "1.0.0"},
	})

	// Create the expected skill directory (simulating a placed skill).
	skillDir := filepath.Join(dir, ".claude", "skills", "skill-a")
	require.NoError(t, os.MkdirAll(skillDir, 0755))

	out, err := runListCmd(t, dir, false, true)
	require.NoError(t, err)
	assert.Contains(t, out, "OK")
	assert.Contains(t, out, "skill-a")
}

func indexOf(s, sub string) int {
	idx := 0
	for i := range s {
		if s[i:] >= sub && len(s[i:]) >= len(sub) && s[i:i+len(sub)] == sub {
			return idx
		}
		idx++
	}
	return -1
}
