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
	Type:        "agent",
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
	AgentCompat: []string{"claude"},
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yml")

	err := manifest.Save(exampleManifest, path)
	require.NoError(t, err)

	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, exampleManifest, loaded)
}

func TestRoundTrip_OutputsPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yml")

	require.NoError(t, manifest.Save(exampleManifest, path))
	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, exampleManifest.Outputs, loaded.Outputs)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := manifest.Load("/nonexistent/path/melon.yml")
	assert.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yml")
	require.NoError(t, os.WriteFile(path, []byte(":\tinvalid: yaml: {{{"), 0644))

	_, err := manifest.Load(path)
	assert.Error(t, err)
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yml")

	require.NoError(t, manifest.Save(exampleManifest, path))

	_, err := os.Stat(path)
	assert.NoError(t, err)
}

// TestRoundTrip_NoOutputsBlock verifies that loading a melon.yml with no outputs
// block and saving it back does NOT produce an empty "outputs: {}" entry.
func TestRoundTrip_NoOutputsBlock(t *testing.T) {
	src := `name: my-agent
version: 1.0.0
type: agent
agent_compat:
  - claude-code
`
	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yml")
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

func TestRoundTrip_EmptyOptionalFields(t *testing.T) {
	m := manifest.Manifest{
		Name:       "minimal",
		Version:    "0.1.0",
		Type:       "skill",
		Entrypoint: "SKILL.md",
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "melon.yml")
	require.NoError(t, manifest.Save(m, path))

	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, m, loaded)
}
