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
	path := filepath.Join(dir, "mln.yaml")

	err := manifest.Save(exampleManifest, path)
	require.NoError(t, err)

	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, exampleManifest, loaded)
}

func TestRoundTrip_OutputsPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mln.yaml")

	require.NoError(t, manifest.Save(exampleManifest, path))
	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, exampleManifest.Outputs, loaded.Outputs)
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := manifest.Load("/nonexistent/path/mln.yaml")
	assert.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mln.yaml")
	require.NoError(t, os.WriteFile(path, []byte(":\tinvalid: yaml: {{{"), 0644))

	_, err := manifest.Load(path)
	assert.Error(t, err)
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mln.yaml")

	require.NoError(t, manifest.Save(exampleManifest, path))

	_, err := os.Stat(path)
	assert.NoError(t, err)
}

func TestRoundTrip_EmptyOptionalFields(t *testing.T) {
	m := manifest.Manifest{
		Name:       "minimal",
		Version:    "0.1.0",
		Type:       "skill",
		Entrypoint: "SKILL.md",
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "mln.yaml")
	require.NoError(t, manifest.Save(m, path))

	loaded, err := manifest.Load(path)
	require.NoError(t, err)

	assert.Equal(t, m, loaded)
}
