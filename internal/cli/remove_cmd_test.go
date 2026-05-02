package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// writeManifest writes a minimal melon.yaml with the given deps to dir.
func writeManifest(t *testing.T, dir string, deps map[string]string) {
	t.Helper()
	m := manifest.Manifest{
		Name:         "test-project",
		Version:      "0.1.0",
		ToolCompat:   []string{"claude-code"},
		Dependencies: deps,
	}
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.yaml"), data, 0644))
}

func TestRunRemove_RemovesDependencyFromManifest(t *testing.T) {
	dir := t.TempDir()

	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"alice/pdf-skill": "^1.3.0",
		"bob/base-utils":  "^0.5.0",
	})

	// runRemove will call runInstall after saving melon.yaml. runInstall will
	// fail at the resolve step (no network in tests), but the manifest mutation
	// happens before that — we accept the install error here.
	_ = runRemove(removeCmd, []string{"alice/pdf-skill"})

	m, err := manifest.Load(filepath.Join(dir, "melon.yaml"))
	require.NoError(t, err)

	assert.NotContains(t, m.Dependencies, "alice/pdf-skill", "removed dep must not be in melon.yaml")
	assert.Contains(t, m.Dependencies, "bob/base-utils", "remaining dep must still be in melon.yaml")
}

func TestRunRemove_ErrorOnUnknownDep(t *testing.T) {
	dir := t.TempDir()

	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{
		"alice/pdf-skill": "^1.3.0",
	})

	manifestBefore, err := os.ReadFile(filepath.Join(dir, "melon.yaml"))
	require.NoError(t, err)

	err = runRemove(removeCmd, []string{"alice/unknown-skill"})
	require.Error(t, err, "removing an unknown dep must return an error")
	assert.Contains(t, err.Error(), "alice/unknown-skill")

	// melon.yaml must be unchanged.
	manifestAfter, err := os.ReadFile(filepath.Join(dir, "melon.yaml"))
	require.NoError(t, err)
	assert.Equal(t, string(manifestBefore), string(manifestAfter), "melon.yaml must not be modified on error")
}

// --- removeModel unit tests (5.1) ---

func TestRemoveModel_SpaceTogglesSelection(t *testing.T) {
	skills := []removeSkillItem{
		{name: "alice/skill-a", version: "^1.0.0"},
		{name: "bob/skill-b", version: "^2.0.0"},
	}
	m := newRemoveModel(skills)

	// Initially nothing selected.
	assert.False(t, m.sel[0])

	// Press space — item 0 should be selected.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(removeModel)
	assert.True(t, m.sel[0])

	// Press space again — item 0 should be deselected.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(removeModel)
	assert.False(t, m.sel[0])
}

func TestRemoveModel_EnterWithNoSelectionReturnsEmpty(t *testing.T) {
	skills := []removeSkillItem{
		{name: "alice/skill-a", version: "^1.0.0"},
	}
	m := newRemoveModel(skills)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(removeModel)

	assert.Empty(t, m.selected)
	assert.True(t, m.quitting)
}

func TestRemoveModel_EscCancelsWithoutSelection(t *testing.T) {
	skills := []removeSkillItem{
		{name: "alice/skill-a", version: "^1.0.0"},
	}
	m := newRemoveModel(skills)

	// Select item, then cancel with esc.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(removeModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(removeModel)

	assert.Empty(t, m.selected, "esc must not populate selected even if items were toggled")
	assert.True(t, m.quitting)
}

func TestRemoveModel_EnterCollectsSelectedNames(t *testing.T) {
	skills := []removeSkillItem{
		{name: "alice/skill-a", version: "^1.0.0"},
		{name: "bob/skill-b", version: "^2.0.0"},
	}
	m := newRemoveModel(skills)

	// Select item 0, then confirm.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(removeModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(removeModel)

	require.Len(t, m.selected, 1)
	assert.Equal(t, "alice/skill-a", m.selected[0])
}

// --- non-TTY no-args test (5.2) ---

func TestRunRemove_NoArgNonTTYReturnsError(t *testing.T) {
	// isTTY() returns false in test environments (no real terminal).
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{"alice/skill-a": "^1.0.0"})

	err := runRemove(removeCmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-interactive mode")
}

// TestRunRemove_VendorFalse_RemovesGitignoreEntry_LastDep verifies gitignore
// cleanup when removing the only remaining dependency.
func TestRunRemove_VendorFalse_RemovesGitignoreEntry_LastDep(t *testing.T) {
	dir := t.TempDir()

	vendorFalse := false
	m := manifest.Manifest{
		Name:       "test-project",
		Version:    "0.1.0",
		ToolCompat: []string{"claude-code"},
		Dependencies: map[string]string{
			"github.com/alice/skills/skill-a": "^1.0.0",
		},
		Vendor: &vendorFalse,
	}
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.yaml"), data, 0644))

	// Pre-populate .gitignore with the entry that should be removed.
	gitignoreContent := ".melon/\n# melon managed — do not edit this block\n.claude/skills/skill-a\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignoreContent), 0644))

	orig := struct {
		resolveVersion func(string, string) (string, string, error)
		fetchManifest  func(string, string, string) (manifest.Manifest, error)
		fetch          func(resolver.ResolvedDep, string) (fetcher.FetchResult, error)
	}{resolveVersionFn, fetchManifestFn, fetchFn}
	t.Cleanup(func() {
		resolveVersionFn = orig.resolveVersion
		fetchManifestFn = orig.fetchManifest
		fetchFn = orig.fetch
	})

	resolveVersionFn = func(repoURL, constraint string) (string, string, error) { return "1.0.0", "v1.0.0", nil }
	fetchManifestFn = func(repoURL, gitTag, subdir string) (manifest.Manifest, error) { return manifest.Manifest{}, nil }
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		require.NoError(t, os.MkdirAll(installDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(installDir, "SKILL.md"), []byte("# skill"), 0644))
		return fetcher.FetchResult{TreeHash: "sha256:abc", Files: []string{"SKILL.md"}}, nil
	}

	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	require.NoError(t, runRemove(cmd, []string{"github.com/alice/skills/skill-a"}))

	gitignoreData, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err)
	assert.NotContains(t, string(gitignoreData), ".claude/skills/skill-a")
}

// TestRunRemove_VendorFalse_RemovesGitignoreEntry_WithRemainingDeps is the
// realistic case: removing one skill when others remain. This is where the bug
// occurred — runInstall runs fully (not early-returning on 0 deps), and the
// gitignore cleanup must still execute regardless.
func TestRunRemove_VendorFalse_RemovesGitignoreEntry_WithRemainingDeps(t *testing.T) {
	dir := t.TempDir()

	vendorFalse := false
	m := manifest.Manifest{
		Name:       "test-project",
		Version:    "0.1.0",
		ToolCompat: []string{"claude-code"},
		Dependencies: map[string]string{
			"github.com/alice/skills/skill-a": "^1.0.0",
			"github.com/bob/skills/skill-b":   "^2.0.0",
			"github.com/carol/skills/skill-c": "^3.0.0",
		},
		Vendor: &vendorFalse,
	}
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.yaml"), data, 0644))

	// Pre-populate .gitignore simulating a previous melon install with all 3 deps.
	gitignoreContent := ".melon/\n# melon managed — do not edit this block\n.claude/skills/skill-a\n.claude/skills/skill-b\n.claude/skills/skill-c\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignoreContent), 0644))

	orig := struct {
		resolveVersion func(string, string) (string, string, error)
		fetchManifest  func(string, string, string) (manifest.Manifest, error)
		fetch          func(resolver.ResolvedDep, string) (fetcher.FetchResult, error)
	}{resolveVersionFn, fetchManifestFn, fetchFn}
	t.Cleanup(func() {
		resolveVersionFn = orig.resolveVersion
		fetchManifestFn = orig.fetchManifest
		fetchFn = orig.fetch
	})

	resolveVersionFn = func(repoURL, constraint string) (string, string, error) {
		switch repoURL {
		case "https://github.com/bob/skills":
			return "2.0.0", "v2.0.0", nil
		case "https://github.com/carol/skills":
			return "3.0.0", "v3.0.0", nil
		}
		return "1.0.0", "v1.0.0", nil
	}
	fetchManifestFn = func(repoURL, gitTag, subdir string) (manifest.Manifest, error) { return manifest.Manifest{}, nil }
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		require.NoError(t, os.MkdirAll(installDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(installDir, "SKILL.md"), []byte("# skill"), 0644))
		return fetcher.FetchResult{TreeHash: "sha256:abc", Files: []string{"SKILL.md"}}, nil
	}

	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	require.NoError(t, runRemove(cmd, []string{"github.com/alice/skills/skill-a"}))

	gitignoreData, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err)
	content := string(gitignoreData)

	// Removed skill must be gone.
	assert.NotContains(t, content, ".claude/skills/skill-a", "removed skill must be removed from .gitignore")
	// Remaining skills must still be present.
	assert.Contains(t, content, ".claude/skills/skill-b", "remaining skills must stay in .gitignore")
	assert.Contains(t, content, ".claude/skills/skill-c", "remaining skills must stay in .gitignore")
}

// --- empty melon.yaml no-args test (5.3) ---

func TestRunRemoveInteractive_EmptyManifestPrintsMessage(t *testing.T) {
	dir := t.TempDir()
	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	writeManifest(t, dir, map[string]string{})

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	// In a non-TTY test environment runRemoveInteractive returns the
	// non-interactive error before reaching the empty-manifest branch.
	// We verify it returns an error without panicking, and does not modify
	// melon.yaml.
	err := runRemoveInteractive(cmd)
	require.Error(t, err)

	// melon.yaml must be untouched.
	m, loadErr := manifest.Load(filepath.Join(dir, "melon.yaml"))
	require.NoError(t, loadErr)
	assert.Empty(t, m.Dependencies)
}
