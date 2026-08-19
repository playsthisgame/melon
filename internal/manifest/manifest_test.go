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
	Tags:       []string{"coding-agent"},
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

func TestIndex_RoundTrip(t *testing.T) {
	urls := []string{"https://example.com/index.yaml", "https://corp.example.com/index.yaml"}
	f := false
	m := manifest.Manifest{
		Name:    "x",
		Version: "0.1.0",
		Index:   &manifest.IndexConfig{URLs: urls, PublicIndex: &f},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, manifest.Save(m, path))

	loaded, err := manifest.Load(path)
	require.NoError(t, err)
	require.NotNil(t, loaded.Index)
	assert.Equal(t, urls, loaded.Index.URLs)
	require.NotNil(t, loaded.Index.PublicIndex)
	assert.False(t, *loaded.Index.PublicIndex)
}

func TestIndex_AbsentFieldRoundTrip(t *testing.T) {
	m := manifest.Manifest{Name: "x", Version: "0.1.0"}
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, manifest.Save(m, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "index:", "absent index field must not be serialized")

	loaded, err := manifest.Load(path)
	require.NoError(t, err)
	assert.Nil(t, loaded.Index)
}

func TestPolicy_RoundTrip(t *testing.T) {
	m := manifest.Manifest{
		Name:    "x",
		Version: "0.1.0",
		Policy:  &manifest.PolicyConfig{AllowedSources: []string{"github.com/my-company/*", "github.com/trusted/*"}},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, manifest.Save(m, path))

	loaded, err := manifest.Load(path)
	require.NoError(t, err)
	require.NotNil(t, loaded.Policy)
	assert.Equal(t, m.Policy.AllowedSources, loaded.Policy.AllowedSources)
}

func TestPolicy_AbsentFieldRoundTrip(t *testing.T) {
	m := manifest.Manifest{Name: "x", Version: "0.1.0"}
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	require.NoError(t, manifest.Save(m, path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "policy:", "absent policy field must not be serialized")

	loaded, err := manifest.Load(path)
	require.NoError(t, err)
	assert.Nil(t, loaded.Policy)
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

func TestHarnessList_PrefersHarnessesOverToolCompat(t *testing.T) {
	m := manifest.Manifest{
		Harnesses:  []string{"claude-code"},
		ToolCompat: []string{"cursor"},
	}
	assert.Equal(t, []string{"claude-code"}, m.HarnessList())
	assert.False(t, m.UsesLegacyToolCompat())
}

func TestHarnessList_FallsBackToToolCompat(t *testing.T) {
	m := manifest.Manifest{ToolCompat: []string{"cursor", "windsurf"}}
	assert.Equal(t, []string{"cursor", "windsurf"}, m.HarnessList())
	assert.True(t, m.UsesLegacyToolCompat())
}

func TestHarnessList_EmptyWhenNeitherSet(t *testing.T) {
	var m manifest.Manifest
	assert.Empty(t, m.HarnessList())
	assert.False(t, m.UsesLegacyToolCompat())
}

// A pre-v0.5.0 manifest must keep working, and melon must NOT silently rewrite
// tool_compat to harnesses — a teammate on an older binary would then read a
// manifest with no tool_compat and place skills in the wrong directory.
func TestToolCompat_LegacyManifestParsesAndIsNotRewritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	legacy := "name: legacy\nversion: 1.0.0\nentrypoint: SKILL.md\ntool_compat:\n  - claude-code\n"
	require.NoError(t, os.WriteFile(path, []byte(legacy), 0644))

	m, err := manifest.Load(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"claude-code"}, m.HarnessList())
	assert.True(t, m.UsesLegacyToolCompat())

	// Re-saving must preserve the original key, not migrate it.
	require.NoError(t, manifest.Save(m, path))
	out, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(out), "tool_compat:")
	assert.NotContains(t, string(out), "harnesses:")
}

func TestHarnesses_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yaml")
	m := manifest.Manifest{
		Name:       "modern",
		Version:    "1.0.0",
		Entrypoint: "SKILL.md",
		Harnesses:  []string{"claude-code", "codex"},
	}
	require.NoError(t, manifest.Save(m, path))

	out, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(out), "harnesses:")
	assert.NotContains(t, string(out), "tool_compat:")

	got, err := manifest.Load(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"claude-code", "codex"}, got.HarnessList())
	assert.False(t, got.UsesLegacyToolCompat())
}
