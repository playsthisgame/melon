package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// writeLockfile writes a minimal melon.lock with the given locked deps to dir.
func writeLockfile(t *testing.T, dir string, deps []lockfile.LockedDep) {
	t.Helper()
	lf := lockfile.LockFile{
		GeneratedAt:  "2026-01-01T00:00:00Z",
		Dependencies: deps,
	}
	data, err := yaml.Marshal(lf)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.lock"), data, 0644))
}

// --- isBranchPin unit tests ---

func TestIsBranchPin(t *testing.T) {
	cases := []struct {
		constraint string
		want       bool
	}{
		{"^1.2.3", false},
		{"~1.2.3", false},
		{"1.2.3", false},
		{"0.1.0", false},
		{"main", true},
		{"HEAD", true},
		{"feature-branch", true},
		{"v1.2.3", false}, // bare version with v prefix
	}
	for _, tc := range cases {
		t.Run(tc.constraint, func(t *testing.T) {
			got := isBranchPin(tc.constraint)
			assert.Equal(t, tc.want, got, "isBranchPin(%q)", tc.constraint)
		})
	}
}

// --- lockedVersionFor unit tests ---

func TestLockedVersionFor(t *testing.T) {
	dir := t.TempDir()
	writeLockfile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/alice/skills/skill-a", Version: "1.3.0"},
		{Name: "github.com/bob/tools/tool-b", Version: "2.0.1"},
	})

	lockPath := filepath.Join(dir, "melon.lock")

	assert.Equal(t, "1.3.0", lockedVersionFor(lockPath, "github.com/alice/skills/skill-a"))
	assert.Equal(t, "2.0.1", lockedVersionFor(lockPath, "github.com/bob/tools/tool-b"))
	assert.Equal(t, "", lockedVersionFor(lockPath, "github.com/nobody/missing"))
}

func TestLockedVersionFor_MissingLockfile(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "melon.lock") // does not exist
	assert.Equal(t, "", lockedVersionFor(lockPath, "any/dep"))
}

// --- runUpdate targeted mode ---

func TestRunUpdateTargeted_DepNotInManifest(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "^1.0.0",
	})

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err := runUpdateTargeted(cmd, "github.com/nobody/unknown-skill")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a dependency in melon.yaml")
}

func TestRunUpdateTargeted_BranchPinnedDep(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "main",
	})

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err := runUpdateTargeted(cmd, "github.com/alice/skills/skill-a")
	require.NoError(t, err, "branch-pinned dep must not error")
	assert.Contains(t, buf.String(), "branch-pinned")
}

// --- runUpdate no-args non-TTY ---

func TestRunUpdate_NoArgNonTTYReturnsError(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "^1.0.0",
	})

	err := runUpdate(updateCmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// --- runUpdateInteractive empty manifest ---

func TestRunUpdateInteractive_EmptyManifestPrintsMessage(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{})

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	// Non-TTY: runUpdate returns the non-interactive error, not the empty message.
	// Call runUpdateInteractive directly to test the empty-manifest branch.
	err := runUpdateInteractive(cmd)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No skills in melon.yaml.")
}

// --- runUpdateInteractive all-branch-pinned ---

func TestRunUpdateInteractive_AllBranchPinnedPrintsMessage(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "main",
		"github.com/bob/tools/tool-b":     "HEAD",
	})

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	err := runUpdateInteractive(cmd)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No updatable skills")
}

// --- updateModel TUI unit tests ---

func TestUpdateModel_SpaceTogglesSelection(t *testing.T) {
	skills := []updateSkillItem{
		{name: "github.com/alice/skills/skill-a", constraint: "^1.0.0"},
	}
	m := newUpdateModel(skills, []string{"github.com/alice/skills/skill-a"})

	// index 0 is the "Update all" sentinel; index 1 is skill-a.
	// Press down to move to skill-a.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(updateModel)

	assert.False(t, m.sel[1], "skill-a not yet selected")

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(updateModel)
	assert.True(t, m.sel[1], "skill-a should be selected after space")

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(updateModel)
	assert.False(t, m.sel[1], "skill-a should be deselected after second space")
}

func TestUpdateModel_UpdateAllSentinelReturnsAllDeps(t *testing.T) {
	allDeps := []string{
		"github.com/alice/skills/skill-a",
		"github.com/bob/tools/tool-b",
	}
	skills := []updateSkillItem{
		{name: allDeps[0], constraint: "^1.0.0"},
		{name: allDeps[1], constraint: "^2.0.0"},
	}
	m := newUpdateModel(skills, allDeps)

	// Select "Update all" sentinel (index 0) and confirm.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(updateModel)
	assert.True(t, m.sel[0], "sentinel must be selected")

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(updateModel)

	assert.ElementsMatch(t, allDeps, m.selected, "Update all must return every dep")
}

func TestUpdateModel_IndividualSelectionReturnsOnlySelected(t *testing.T) {
	allDeps := []string{
		"github.com/alice/skills/skill-a",
		"github.com/bob/tools/tool-b",
	}
	skills := []updateSkillItem{
		{name: allDeps[0], constraint: "^1.0.0"},
		{name: allDeps[1], constraint: "^2.0.0"},
	}
	m := newUpdateModel(skills, allDeps)

	// Move to index 1 (skill-a) and select, leave sentinel unselected.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(updateModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(updateModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(updateModel)

	require.Len(t, m.selected, 1)
	assert.Equal(t, allDeps[0], m.selected[0])
}

func TestUpdateModel_EscCancelsWithoutSelection(t *testing.T) {
	skills := []updateSkillItem{
		{name: "github.com/alice/skills/skill-a", constraint: "^1.0.0"},
	}
	m := newUpdateModel(skills, []string{"github.com/alice/skills/skill-a"})

	// Select sentinel, then cancel.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(updateModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(updateModel)

	assert.Empty(t, m.selected, "esc must not commit selection")
	assert.True(t, m.quitting)
}

// --- newer major version hint ---

func TestPrintNewerMajorHint_PrintsWhenOutsideConstraint(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	// constraint ^1.x, absolute latest is 2.0.0 — outside constraint
	printNewerMajorHint(cmd, "github.com/alice/skills/skill-a", "^1.0.0", "2.0.0")
	assert.Contains(t, buf.String(), "hint:")
	assert.Contains(t, buf.String(), "2.0.0")
	assert.Contains(t, buf.String(), "^2.0.0")
}

func TestPrintNewerMajorHint_SilentWhenInsideConstraint(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	// absolute latest 1.5.0 is within ^1.0.0 — no hint
	printNewerMajorHint(cmd, "github.com/alice/skills/skill-a", "^1.0.0", "1.5.0")
	assert.Empty(t, buf.String())
}

func TestPrintNewerMajorHint_SilentWhenNoAbsoluteLatest(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	printNewerMajorHint(cmd, "github.com/alice/skills/skill-a", "^1.0.0", "")
	assert.Empty(t, buf.String())
}
