package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var exampleManifest = manifest.Manifest{
	Name:        "my-agent",
	Version:     "1.0.0",
	Description: "My coding agent",
	Entrypoint:  "CLAUDE.md",
	Dependencies: map[string]string{
		"github.com/alice/pdf-skill":  "^1.2.0",
		"github.com/alice/xlsx-skill": "^2.0.0",
	},
	Outputs: map[string]string{
		"CLAUDE.md":        "*",
		".claude/SKILL.md": "github.com/alice/*",
	},
	Tags:        []string{"coding-agent"},
	ToolCompat: []string{"claude"},
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")

	err := manifest.Save(exampleManifest, path)
	require.NoError(t, err)

	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, exampleManifest, loaded)
}

func TestRoundTrip_OutputsPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")

	require.NoError(t, manifest.Save(exampleManifest, path))
	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, exampleManifest.Outputs, loaded.Outputs)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := manifest.Load("/nonexistent/path/melon.yaml")
	assert.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, os.WriteFile(path, []byte(":\tinvalid: yaml: {{{"), 0644))

	_, err := manifest.Load(path)
	assert.Error(t, err)
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")

	require.NoError(t, manifest.Save(exampleManifest, path))

	_, err := os.Stat(path)
	assert.NoError(t, err)
}

// TestRoundTrip_NoOutputsBlock verifies that loading a melon.yaml with no outputs
// block and saving it back does NOT produce an empty "outputs: {}" entry.
func TestRoundTrip_NoOutputsBlock(t *testing.T) {
	src := `name: my-agent
version: 1.0.0
type: agent
tool_compat:
  - claude-code
`
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, os.WriteFile(path, []byte(src), 0644))

	m, err := manifest.Load(path)
	require.NoError(t, err)
	assert.Nil(t, m.Outputs, "Outputs should be nil when not declared")

	// Save and reload — outputs block must not appear.
	require.NoError(t, manifest.Save(m, path))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "outputs:", "saved YAML must not contain an outputs key")

	m2, err := manifest.Load(path)
	require.NoError(t, err)
	assert.Nil(t, m2.Outputs)
}

func TestVendor_DefaultIsVendored(t *testing.T) {
	// Absent vendor field → IsVendored() == true
	m := manifest.Manifest{Name: "x", Version: "0.1.0"}
	assert.True(t, m.IsVendored(), "nil Vendor should be vendored")

	// Explicit true → IsVendored() == true
	v := true
	m.Vendor = &v
	assert.True(t, m.IsVendored())

	// Explicit false → IsVendored() == false
	f := false
	m.Vendor = &f
	assert.False(t, m.IsVendored())
}

func TestVendor_RoundTrip(t *testing.T) {
	f := false
	m := manifest.Manifest{Name: "x", Version: "0.1.0", Vendor: &f}

	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, manifest.Save(m, path))

	loaded, err := manifest.Load(path)
	require.NoError(t, err)
	require.NotNil(t, loaded.Vendor)
	assert.False(t, *loaded.Vendor)
	assert.False(t, loaded.IsVendored())
}

func TestVendor_AbsentFieldRoundTrip(t *testing.T) {
	// When vendor is nil, saving should not emit a vendor key.
	m := manifest.Manifest{Name: "x", Version: "0.1.0"}
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, manifest.Save(m, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "vendor:", "absent vendor field must not be serialized")

	loaded, err := manifest.Load(path)
	require.NoError(t, err)
	assert.Nil(t, loaded.Vendor)
	assert.True(t, loaded.IsVendored())
}

func TestRoundTrip_EmptyOptionalFields(t *testing.T) {
	m := manifest.Manifest{
		Name:       "minimal",
		Version:    "0.1.0",
		Entrypoint: "SKILL.md",
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, manifest.Save(m, path))

	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, m, loaded)
}
