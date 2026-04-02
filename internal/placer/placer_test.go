package placer_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/placer"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeDep creates a ResolvedDep and populates its .melon/ cache directory with
// a single SKILL.md file so the symlink target actually exists.
func makeDep(t *testing.T, projectDir, name, version string) resolver.ResolvedDep {
	t.Helper()
	dep := resolver.ResolvedDep{
		Name:       name,
		Version:    version,
		GitTag:     "v" + version,
		RepoURL:    "https://github.com/alice/" + filepath.Base(name),
		Entrypoint: "SKILL.md",
	}
	cacheDir := store.InstalledPath(projectDir, dep)
	require.NoError(t, os.MkdirAll(cacheDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "SKILL.md"), []byte("# skill"), 0644))
	return dep
}

func TestPlace_CreatesSymlink(t *testing.T) {
	projectDir := t.TempDir()
	dep := makeDep(t, projectDir, "github.com/alice/pdf-skill", "1.2.0")

	m := manifest.Manifest{AgentCompat: []string{"claude-code"}}
	require.NoError(t, placer.Place([]resolver.ResolvedDep{dep}, m, projectDir, &bytes.Buffer{}))

	linkPath := filepath.Join(projectDir, ".claude/skills/pdf-skill")
	info, err := os.Lstat(linkPath)
	require.NoError(t, err, "link path must exist")
	assert.Equal(t, os.ModeSymlink, info.Mode()&os.ModeSymlink, "entry must be a symlink")
}

func TestPlace_SymlinkTargetResolvesToCache(t *testing.T) {
	projectDir := t.TempDir()
	dep := makeDep(t, projectDir, "github.com/alice/pdf-skill", "1.2.0")

	m := manifest.Manifest{AgentCompat: []string{"claude-code"}}
	require.NoError(t, placer.Place([]resolver.ResolvedDep{dep}, m, projectDir, &bytes.Buffer{}))

	linkPath := filepath.Join(projectDir, ".claude/skills/pdf-skill")
	resolved, err := filepath.EvalSymlinks(linkPath)
	require.NoError(t, err)

	expected, err := filepath.EvalSymlinks(store.InstalledPath(projectDir, dep))
	require.NoError(t, err)

	assert.Equal(t, expected, resolved, "symlink must resolve to .melon/ cache entry")
}

func TestPlace_Idempotent(t *testing.T) {
	projectDir := t.TempDir()
	dep := makeDep(t, projectDir, "github.com/alice/pdf-skill", "1.2.0")

	m := manifest.Manifest{AgentCompat: []string{"claude-code"}}
	require.NoError(t, placer.Place([]resolver.ResolvedDep{dep}, m, projectDir, &bytes.Buffer{}))
	// Second call must not error.
	require.NoError(t, placer.Place([]resolver.ResolvedDep{dep}, m, projectDir, &bytes.Buffer{}))

	linkPath := filepath.Join(projectDir, ".claude/skills/pdf-skill")
	info, err := os.Lstat(linkPath)
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode()&os.ModeSymlink)
}

func TestPlace_ReplacesStaleDirectory(t *testing.T) {
	projectDir := t.TempDir()
	dep := makeDep(t, projectDir, "github.com/alice/pdf-skill", "1.2.0")

	// Pre-populate the link path with a real directory (old copy-based layout).
	staleDir := filepath.Join(projectDir, ".claude/skills/pdf-skill")
	require.NoError(t, os.MkdirAll(staleDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(staleDir, "old.md"), []byte("old"), 0644))

	m := manifest.Manifest{AgentCompat: []string{"claude-code"}}
	require.NoError(t, placer.Place([]resolver.ResolvedDep{dep}, m, projectDir, &bytes.Buffer{}))

	info, err := os.Lstat(staleDir)
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode()&os.ModeSymlink, "stale directory must be replaced by symlink")
}
