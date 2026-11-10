package placer_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/placer"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makePack creates a pack dep in the cache. withSkill controls whether it is a
// hybrid pack (SKILL.md present) or a pure pack (only melon.yaml).
func makePack(t *testing.T, projectDir, name string, withSkill bool) resolver.ResolvedDep {
	t.Helper()
	dep := resolver.ResolvedDep{Name: name, Version: "1.0.0", Entrypoint: "SKILL.md", Type: manifest.TypePack}
	cacheDir := store.InstalledPath(projectDir, dep)
	require.NoError(t, os.MkdirAll(cacheDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "melon.yaml"), []byte("type: pack\n"), 0644))
	if withSkill {
		require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "SKILL.md"), []byte("# expert"), 0644))
	}
	return dep
}

func TestShouldPlace(t *testing.T) {
	dir := t.TempDir()
	assert.True(t, placer.ShouldPlace(makeDep(t, dir, "github.com/alice/pdf-skill", "1.0.0"), dir))
	assert.False(t, placer.ShouldPlace(makePack(t, dir, "github.com/c/pure-pack", false), dir))
	assert.True(t, placer.ShouldPlace(makePack(t, dir, "github.com/c/hybrid-pack", true), dir))
	assert.False(t, placer.ShouldPlaceLocked(lockfile.LockedDep{
		Name: "github.com/c/pure-pack", Version: "1.0.0", Entrypoint: "SKILL.md", Type: "pack"}, dir))
}

func TestPlace_PurePackNotPlaced(t *testing.T) {
	dir := t.TempDir()
	pack := makePack(t, dir, "github.com/c/spring-expert-pack", false)
	member := makeDep(t, dir, "github.com/acme/spring-boot", "1.0.0")

	m := manifest.Manifest{Harnesses: []string{"claude-code"}}
	require.NoError(t, placer.Place([]resolver.ResolvedDep{pack, member}, m, dir, &bytes.Buffer{}))

	_, err := os.Lstat(filepath.Join(dir, ".claude/skills/spring-expert-pack"))
	assert.True(t, os.IsNotExist(err), "pure pack must not be placed")
	_, err = os.Lstat(filepath.Join(dir, ".claude/skills/spring-boot"))
	assert.NoError(t, err, "member must be placed")
}

func TestPlace_HybridPackPlaced(t *testing.T) {
	dir := t.TempDir()
	pack := makePack(t, dir, "github.com/c/spring-expert-pack", true)

	m := manifest.Manifest{Harnesses: []string{"claude-code"}}
	require.NoError(t, placer.Place([]resolver.ResolvedDep{pack}, m, dir, &bytes.Buffer{}))

	info, err := os.Lstat(filepath.Join(dir, ".claude/skills/spring-expert-pack"))
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode()&os.ModeSymlink)
}

func TestPlace_PurePackClearsStaleLink(t *testing.T) {
	dir := t.TempDir()
	linkDir := filepath.Join(dir, ".claude/skills")
	require.NoError(t, os.MkdirAll(linkDir, 0755))
	require.NoError(t, os.Symlink("../../.melon/gone", filepath.Join(linkDir, "spring-expert-pack")))

	pack := makePack(t, dir, "github.com/c/spring-expert-pack", false)
	m := manifest.Manifest{Harnesses: []string{"claude-code"}}
	require.NoError(t, placer.Place([]resolver.ResolvedDep{pack}, m, dir, &bytes.Buffer{}))

	_, err := os.Lstat(filepath.Join(linkDir, "spring-expert-pack"))
	assert.True(t, os.IsNotExist(err), "stale link from a former hybrid version must be removed")
}

func TestUnplace_PurePackIsQuiet(t *testing.T) {
	dir := t.TempDir()
	makePack(t, dir, "github.com/c/spring-expert-pack", false)
	var out bytes.Buffer
	m := manifest.Manifest{Harnesses: []string{"claude-code"}}
	require.NoError(t, placer.Unplace([]lockfile.LockedDep{{
		Name: "github.com/c/spring-expert-pack", Version: "1.0.0", Entrypoint: "SKILL.md", Type: "pack"}}, m, dir, &out))
	assert.NotContains(t, out.String(), "unlinked")
}
