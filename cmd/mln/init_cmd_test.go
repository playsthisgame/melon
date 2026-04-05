package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitYes_CreatesValidManifestAndStoreDir(t *testing.T) {
	dir := t.TempDir()

	origDir := flagDir
	origYes := flagYes
	t.Cleanup(func() {
		flagDir = origDir
		flagYes = origYes
	})

	flagDir = dir
	flagYes = true

	// Call RunE directly (cobra.Execute swallows errors via os.Exit).
	err := runInit(initCmd, nil)
	require.NoError(t, err)

	manifestPath := filepath.Join(dir, "melon.yml")
	mlnDir := filepath.Join(dir, ".melon")

	// melon.yml must exist and parse without error.
	_, statErr := os.Stat(manifestPath)
	require.NoError(t, statErr, "melon.yml should exist")

	m, loadErr := manifest.Load(manifestPath)
	require.NoError(t, loadErr, "manifest.Load should parse the generated melon.yml")

	// Basic fields from defaults.
	assert.Equal(t, filepath.Base(dir), m.Name)
	assert.Equal(t, "0.1.0", m.Version)

	// tool_compat defaults to ["claude-code"].
	assert.Equal(t, []string{"claude-code"}, m.ToolCompat)

	// outputs must be absent — paths are derived from tool_compat automatically.
	assert.Nil(t, m.Outputs, "outputs should not be set in generated melon.yml")

	// .melon/ directory must exist.
	info, statErr := os.Stat(mlnDir)
	require.NoError(t, statErr, ".melon/ directory should exist")
	assert.True(t, info.IsDir(), ".melon/ should be a directory")
}

func TestInitYes_NoOverwriteCheck(t *testing.T) {
	dir := t.TempDir()

	origDir := flagDir
	origYes := flagYes
	t.Cleanup(func() {
		flagDir = origDir
		flagYes = origYes
	})

	flagDir = dir
	flagYes = true

	require.NoError(t, runInit(initCmd, nil))
	// Second init with --yes should overwrite without error.
	require.NoError(t, runInit(initCmd, nil))

	_, err := manifest.Load(filepath.Join(dir, "melon.yml"))
	assert.NoError(t, err, "melon.yml should still be valid after second init --yes")
}

func TestGenerateManifestYAML_ParsesCleanly(t *testing.T) {
	cases := []struct {
		name       string
		desc       string
		agentNames []string
	}{
		{"my-agent", "A test agent", []string{"claude-code"}},
		{"my-skill", "", []string{"cursor"}},
		{"has-quotes", `description with "quotes"`, []string{"windsurf"}},
		{"multi-agent", "Multi target", []string{"claude-code", "cursor"}},
		{"no-tools", "No tools yet", []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			yaml := generateManifestYAML(tc.name, tc.desc, tc.agentNames)

			dir := t.TempDir()
			path := filepath.Join(dir, "melon.yml")
			require.NoError(t, os.WriteFile(path, []byte(yaml), 0644))

			m, err := manifest.Load(path)
			require.NoError(t, err)
			assert.Equal(t, tc.name, m.Name)
			if len(tc.agentNames) == 0 {
				assert.Empty(t, m.ToolCompat)
			} else {
				assert.Equal(t, tc.agentNames, m.ToolCompat)
			}
			// outputs should not be present — derived from tool_compat.
			assert.Nil(t, m.Outputs)
		})
	}
}
