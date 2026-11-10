package scope

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalPathsDeriveFromHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root, err := GlobalRoot()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "melon"), root)

	st, err := GlobalStore()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".melon", "store"), st)
}

func TestIsGlobalAndStoreDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".config", "melon")

	assert.True(t, IsGlobal(root))
	assert.True(t, IsGlobal(root+"/"))
	assert.Equal(t, filepath.Join(home, ".melon", "store"), StoreDir(root))

	project := t.TempDir()
	assert.False(t, IsGlobal(project))
	assert.Equal(t, filepath.Join(project, ".melon"), StoreDir(project))
}

func TestExpandHomeAndTilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	for in, want := range map[string]string{
		"~/.claude/skills/": filepath.Join(home, ".claude", "skills"),
		".roo/skills":       filepath.Join(home, ".roo", "skills"),
		"/opt/skills":       "/opt/skills",
	} {
		got, err := ExpandHome(in)
		require.NoError(t, err)
		assert.Equal(t, want, got, in)
	}

	assert.Equal(t, "~/.claude/skills/x", Tilde(filepath.Join(home, ".claude", "skills", "x")))
	assert.Equal(t, "/opt/skills", Tilde("/opt/skills"))
}

func TestDisplay(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".config", "melon")
	assert.Equal(t, "~/.claude/skills/x", Display(root, filepath.Join(home, ".claude", "skills", "x")))

	project := t.TempDir()
	assert.Equal(t, filepath.Join(".claude", "skills", "x"), Display(project, filepath.Join(project, ".claude", "skills", "x")))
}
