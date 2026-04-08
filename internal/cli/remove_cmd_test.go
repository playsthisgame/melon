package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
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
