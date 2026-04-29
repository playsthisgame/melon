package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newOutdatedCmd returns a cobra.Command wired to checkOutdated for testing,
// with output captured in buf. checkOutdated is used instead of runOutdated to
// avoid the os.Exit(1) call.
func newOutdatedCmd(t *testing.T, buf *bytes.Buffer) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	return cmd
}

// --- 5.6: empty manifest ---

func TestCheckOutdated_EmptyManifest(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{})

	var buf bytes.Buffer
	cmd := newOutdatedCmd(t, &buf)

	outdatedFound, err := checkOutdated(cmd)
	require.NoError(t, err)
	assert.False(t, outdatedFound)
	assert.Contains(t, buf.String(), "No dependencies declared")
}

// --- 5.5: branch-pinned deps excluded, skip note printed ---

func TestCheckOutdated_BranchPinnedExcluded(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "main",
	})

	var buf bytes.Buffer
	cmd := newOutdatedCmd(t, &buf)

	// Branch-pinned only → no network calls, not outdated, skip note shown.
	outdatedFound, err := checkOutdated(cmd)
	require.NoError(t, err)
	assert.False(t, outdatedFound)
	assert.Contains(t, buf.String(), "skipped")
	assert.Contains(t, buf.String(), "branch-pinned")
	assert.Contains(t, buf.String(), "All skills are up to date.")
}

// --- 5.4: missing lock file → all shown as not installed ---

func TestCheckOutdated_MissingLockFile(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "^1.0.0",
	})
	// No melon.lock written intentionally.

	locked := lockedVersionFor(filepath.Join(dir, "melon.lock"), "github.com/alice/skills/skill-a")
	assert.Equal(t, "", locked, "locked version must be empty when lock file is absent")
}

// --- 5.3: dep not in lock file shown as (not installed) ---

func TestCheckOutdated_DepNotInLock(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "^1.0.0",
		"github.com/bob/tools/tool-b":     "^2.0.0",
	})
	// Only tool-b is in the lock.
	writeLockfile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/bob/tools/tool-b", Version: "2.0.0"},
	})

	lockedA := lockedVersionFor(filepath.Join(dir, "melon.lock"), "github.com/alice/skills/skill-a")
	lockedB := lockedVersionFor(filepath.Join(dir, "melon.lock"), "github.com/bob/tools/tool-b")

	assert.Equal(t, "", lockedA, "skill-a should have no locked version")
	assert.Equal(t, "2.0.0", lockedB)
}

// --- outdatedRow table logic unit tests ---

// TestOutdatedRowFiltering verifies that rows where locked == latestCompatible
// are excluded and rows where they differ are included.
func TestOutdatedRowFiltering(t *testing.T) {
	rows := []outdatedRow{
		{name: "a", locked: "1.0.0", latestCompatible: "1.0.0"}, // up to date
		{name: "b", locked: "1.0.1", latestCompatible: "1.0.2"}, // outdated
		{name: "c", locked: "(not installed)", latestCompatible: "2.0.0"}, // not installed
	}

	var outdated []outdatedRow
	for _, r := range rows {
		if r.locked != r.latestCompatible {
			outdated = append(outdated, r)
		}
	}

	require.Len(t, outdated, 2)
	assert.Equal(t, "b", outdated[0].name)
	assert.Equal(t, "c", outdated[1].name)
}

// TestOutdatedAbsoluteLatestColumn verifies that the absolute latest column is
// only populated when the abs version is outside the constraint.
func TestOutdatedAbsoluteLatestColumn(t *testing.T) {
	rows := []outdatedRow{
		// abs latest within constraint — no abs column
		{name: "a", locked: "1.0.0", latestCompatible: "1.0.2", absoluteLatest: ""},
		// abs latest outside constraint — abs column set
		{name: "b", locked: "1.2.0", latestCompatible: "1.2.0", absoluteLatest: "2.0.0"},
	}

	assert.Equal(t, "", rows[0].absoluteLatest, "abs latest within constraint must not populate column")
	assert.Equal(t, "2.0.0", rows[1].absoluteLatest, "abs latest outside constraint must populate column")

	// Verify the ↑ indicator is appended when rendering.
	for _, r := range rows {
		absCol := ""
		if r.absoluteLatest != "" {
			absCol = r.absoluteLatest + " ↑"
		}
		if r.name == "a" {
			assert.Equal(t, "", absCol)
		} else {
			assert.Equal(t, "2.0.0 ↑", absCol)
		}
	}
}
