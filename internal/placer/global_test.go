package placer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// globalSetup points HOME at a temp dir and returns (home, globalRoot).
func globalSetup(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".config", "melon")
	require.NoError(t, os.MkdirAll(root, 0755))
	return home, root
}

func cacheSkill(t *testing.T, root string, dep resolver.ResolvedDep) string {
	t.Helper()
	dir := store.InstalledPath(root, dep)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0644))
	return dir
}

func TestTargetDirs_Global(t *testing.T) {
	home, root := globalSetup(t)

	dirs, err := TargetDirs(manifest.Manifest{Harnesses: []string{"opencode", "claude-code"}}, root, true)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".config", "opencode", "skills"),
	}, dirs)

	dirs, err = TargetDirs(manifest.Manifest{}, root, true)
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(home, ".agents", "skills")}, dirs)

	dirs, err = TargetDirs(manifest.Manifest{Outputs: map[string]string{"~/custom/skills/": "*"}}, root, true)
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(home, "custom", "skills")}, dirs)
}

func TestPlace_Global_LinksIntoHarnessGlobalDir(t *testing.T) {
	home, root := globalSetup(t)
	dep := resolver.ResolvedDep{Name: "github.com/alice/skills/pdf", Version: "1.0.0", Entrypoint: "SKILL.md"}
	cache := cacheSkill(t, root, dep)
	assert.Equal(t, filepath.Join(home, ".melon", "store", "github.com-alice-skills-pdf@1.0.0"), cache)

	m := manifest.Manifest{Harnesses: []string{"claude-code"}}
	var out bytes.Buffer
	require.NoError(t, Place([]resolver.ResolvedDep{dep}, m, root, &out))

	link := filepath.Join(home, ".claude", "skills", "pdf")
	target, err := os.Readlink(link)
	require.NoError(t, err)
	assert.Equal(t, cache, target, "global links are absolute")
	assert.Contains(t, out.String(), "~/.claude/skills/pdf")

	// Re-placing a melon-managed link is idempotent.
	require.NoError(t, Place([]resolver.ResolvedDep{dep}, m, root, &out))
	_, err = os.Readlink(link)
	require.NoError(t, err)
}

func TestPlace_Global_SkipsUnmanagedSkillAndReports(t *testing.T) {
	home, root := globalSetup(t)
	mine := resolver.ResolvedDep{Name: "github.com/alice/skills/pdf", Version: "1.0.0", Entrypoint: "SKILL.md"}
	other := resolver.ResolvedDep{Name: "github.com/alice/skills/xlsx", Version: "1.0.0", Entrypoint: "SKILL.md"}
	cacheSkill(t, root, mine)
	cacheSkill(t, root, other)

	// The user already has their own pdf skill in ~/.claude/skills/.
	userSkill := filepath.Join(home, ".claude", "skills", "pdf")
	require.NoError(t, os.MkdirAll(userSkill, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(userSkill, "SKILL.md"), []byte("# mine"), 0644))

	m := manifest.Manifest{Harnesses: []string{"claude-code"}}
	var out bytes.Buffer
	require.NoError(t, Place([]resolver.ResolvedDep{mine, other}, m, root, &out))

	data, err := os.ReadFile(filepath.Join(userSkill, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, "# mine", string(data), "user's skill must be untouched")

	_, err = os.Readlink(filepath.Join(home, ".claude", "skills", "xlsx"))
	assert.NoError(t, err, "other skills are still placed")

	assert.Contains(t, out.String(), "Skipped 1 skill already present and not managed by melon")
	assert.Contains(t, out.String(), "~/.claude/skills/pdf")

	// Unplace must not delete the user's skill either.
	require.NoError(t, Unplace([]lockfile.LockedDep{{Name: mine.Name, Version: mine.Version}}, m, root, &out))
	_, err = os.Stat(filepath.Join(userSkill, "SKILL.md"))
	assert.NoError(t, err)
}

func TestUnplace_Global_RemovesManagedLink(t *testing.T) {
	home, root := globalSetup(t)
	dep := resolver.ResolvedDep{Name: "github.com/alice/skills/pdf", Version: "1.0.0", Entrypoint: "SKILL.md"}
	cacheSkill(t, root, dep)
	m := manifest.Manifest{Harnesses: []string{"claude-code"}}
	var out bytes.Buffer
	require.NoError(t, Place([]resolver.ResolvedDep{dep}, m, root, &out))

	require.NoError(t, Unplace([]lockfile.LockedDep{{Name: dep.Name, Version: dep.Version}}, m, root, &out))
	_, err := os.Lstat(filepath.Join(home, ".claude", "skills", "pdf"))
	assert.True(t, os.IsNotExist(err))
}
