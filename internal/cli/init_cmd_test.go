package cli

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

	manifestPath := filepath.Join(dir, "melon.yaml")
	mlnDir := filepath.Join(dir, ".melon")

	// melon.yaml must exist and parse without error.
	_, statErr := os.Stat(manifestPath)
	require.NoError(t, statErr, "melon.yaml should exist")

	m, loadErr := manifest.Load(manifestPath)
	require.NoError(t, loadErr, "manifest.Load should parse the generated melon.yaml")

	// Basic fields from defaults.
	assert.Equal(t, filepath.Base(dir), m.Name)
	assert.Equal(t, "0.1.0", m.Version)

	// tool_compat defaults to empty — placement falls back to .agents/skills/.
	assert.Empty(t, m.ToolCompat)

	// outputs must be absent — paths are derived from tool_compat automatically.
	assert.Nil(t, m.Outputs, "outputs should not be set in generated melon.yaml")

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

	_, err := manifest.Load(filepath.Join(dir, "melon.yaml"))
	assert.NoError(t, err, "melon.yaml should still be valid after second init --yes")
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
			yaml := generateManifestYAML(tc.name, tc.desc, tc.agentNames, true)

			dir := t.TempDir()
			path := filepath.Join(dir, "melon.yaml")
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

func TestGenerateManifestYAML_VendorFalse(t *testing.T) {
	yaml := generateManifestYAML("my-agent", "", []string{"claude-code"}, false)

	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0644))

	m, err := manifest.Load(path)
	require.NoError(t, err)
	require.NotNil(t, m.Vendor, "vendor field should be set when vendor=false")
	assert.False(t, *m.Vendor)
	assert.False(t, m.IsVendored())
	assert.Contains(t, yaml, "vendor: false")
}

func TestGenerateManifestYAML_VendorTrue_OmitsField(t *testing.T) {
	yaml := generateManifestYAML("my-agent", "", []string{"claude-code"}, true)
	assert.NotContains(t, yaml, "vendor:", "vendor field should be absent when vendor=true")
}
