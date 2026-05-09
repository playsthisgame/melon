package cli

import (
	"bytes"
	"encoding/json"
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

func runListCmdJSON(t *testing.T, dir string, pending, check bool) (stdout, stderr string, err error) {
	t.Helper()

	origDir := flagDir
	origPending := flagPending
	origCheck := flagCheck
	origJSON := flagListJSON
	t.Cleanup(func() {
		flagDir = origDir
		flagPending = origPending
		flagCheck = origCheck
		flagListJSON = origJSON
	})
	flagDir = dir
	flagPending = pending
	flagCheck = check
	flagListJSON = true

	cmd := &cobra.Command{}
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	err = runList(cmd, nil)
	return outBuf.String(), errBuf.String(), err
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

func TestList_JSON_InstalledSkills(t *testing.T) {
	dir := t.TempDir()
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/skill-a", Version: "1.2.3", GitTag: "v1.2.3", RepoURL: "https://github.com/owner/skill-a", TreeHash: "abc"},
		{Name: "github.com/owner/skill-b", Version: "0.1.0", GitTag: "v0.1.0"},
	})

	stdout, _, err := runListCmdJSON(t, dir, false, false)
	require.NoError(t, err)

	var out listJSONOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out.Installed, 2)
	assert.Equal(t, "github.com/owner/skill-a", out.Installed[0].Name)
	assert.Equal(t, "1.2.3", out.Installed[0].Version)
	assert.Equal(t, "v1.2.3", out.Installed[0].GitTag)
	assert.Equal(t, "abc", out.Installed[0].TreeHash)
	assert.Nil(t, out.Pending)
	assert.Nil(t, out.Check)
}

func TestList_JSON_Empty(t *testing.T) {
	dir := t.TempDir()

	stdout, _, err := runListCmdJSON(t, dir, false, false)
	require.NoError(t, err)

	var out listJSONOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.NotNil(t, out.Installed)
	assert.Empty(t, out.Installed)
}

func TestList_JSON_Pending(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, map[string]string{
		"github.com/owner/installed": "^1.0.0",
		"github.com/owner/pending":   "^2.0.0",
	})
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/installed", Version: "1.0.0"},
	})

	stdout, _, err := runListCmdJSON(t, dir, true, false)
	require.NoError(t, err)

	var out listJSONOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out.Installed, 1)
	require.Len(t, out.Pending, 1)
	assert.Equal(t, "github.com/owner/pending", out.Pending[0])
}

func TestList_JSON_Check_Missing(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, map[string]string{
		"github.com/owner/skill-a": "^1.0.0",
	})
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/skill-a", Version: "1.0.0"},
	})

	stdout, stderr, err := runListCmdJSON(t, dir, false, true)
	require.Error(t, err)
	assert.Contains(t, stderr, `"error"`)

	var out listJSONOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out.Check, 1)
	assert.Equal(t, "missing", out.Check[0].Status)
	assert.Equal(t, "github.com/owner/skill-a", out.Check[0].Name)
}

func TestList_JSON_Check_Present(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, map[string]string{
		"github.com/owner/skill-a": "^1.0.0",
	})
	writeLockFile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/owner/skill-a", Version: "1.0.0"},
	})
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude", "skills", "skill-a"), 0755))

	stdout, _, err := runListCmdJSON(t, dir, false, true)
	require.NoError(t, err)

	var out listJSONOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out.Check, 1)
	assert.Equal(t, "ok", out.Check[0].Status)
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
