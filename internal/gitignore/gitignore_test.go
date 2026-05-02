package gitignore_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playsthisgame/melon/internal/gitignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gitignorePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), ".gitignore")
}

// --- EnsureEntries ---

func TestEnsureEntries_CreatesFileWhenAbsent(t *testing.T) {
	path := gitignorePath(t)
	added, err := gitignore.EnsureEntries(path, []string{".melon/", ".claude/skills/foo"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{".melon/", ".claude/skills/foo"}, added)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, ".melon/")
	assert.Contains(t, content, ".claude/skills/foo")
	assert.Contains(t, content, "# melon managed")
}

func TestEnsureEntries_Idempotent(t *testing.T) {
	path := gitignorePath(t)
	_, err := gitignore.EnsureEntries(path, []string{".melon/"})
	require.NoError(t, err)

	added, err := gitignore.EnsureEntries(path, []string{".melon/"})
	require.NoError(t, err)
	assert.Empty(t, added, "no entries should be added on second call")

	data, _ := os.ReadFile(path)
	count := strings.Count(string(data), ".melon/")
	assert.Equal(t, 1, count, ".melon/ should appear exactly once")
}

func TestEnsureEntries_AppendsNewEntries(t *testing.T) {
	path := gitignorePath(t)
	_, err := gitignore.EnsureEntries(path, []string{".melon/"})
	require.NoError(t, err)

	added, err := gitignore.EnsureEntries(path, []string{".melon/", ".claude/skills/bar"})
	require.NoError(t, err)
	assert.Equal(t, []string{".claude/skills/bar"}, added)

	data, _ := os.ReadFile(path)
	assert.Contains(t, string(data), ".claude/skills/bar")
}

func TestEnsureEntries_CommentHeaderWrittenOnce(t *testing.T) {
	path := gitignorePath(t)
	_, err := gitignore.EnsureEntries(path, []string{".melon/"})
	require.NoError(t, err)
	_, err = gitignore.EnsureEntries(path, []string{".claude/skills/foo"})
	require.NoError(t, err)

	data, _ := os.ReadFile(path)
	count := strings.Count(string(data), "# melon managed")
	assert.Equal(t, 1, count, "comment header should appear exactly once")
}

func TestEnsureEntries_PreservesExistingContent(t *testing.T) {
	path := gitignorePath(t)
	require.NoError(t, os.WriteFile(path, []byte("*.log\n*.tmp\n"), 0644))

	_, err := gitignore.EnsureEntries(path, []string{".melon/"})
	require.NoError(t, err)

	data, _ := os.ReadFile(path)
	content := string(data)
	assert.Contains(t, content, "*.log")
	assert.Contains(t, content, "*.tmp")
	assert.Contains(t, content, ".melon/")
}

// --- RemoveEntries ---

func TestRemoveEntries_RemovesMatchingLines(t *testing.T) {
	path := gitignorePath(t)
	require.NoError(t, os.WriteFile(path, []byte("*.log\n.melon/\n.claude/skills/foo\n"), 0644))

	err := gitignore.RemoveEntries(path, []string{".melon/", ".claude/skills/foo"})
	require.NoError(t, err)

	data, _ := os.ReadFile(path)
	content := string(data)
	assert.Contains(t, content, "*.log")
	assert.NotContains(t, content, ".melon/")
	assert.NotContains(t, content, ".claude/skills/foo")
}

func TestRemoveEntries_NoopWhenFileAbsent(t *testing.T) {
	path := gitignorePath(t)
	err := gitignore.RemoveEntries(path, []string{".melon/"})
	assert.NoError(t, err)
}

func TestRemoveEntries_NoopWhenEntryAbsent(t *testing.T) {
	path := gitignorePath(t)
	require.NoError(t, os.WriteFile(path, []byte("*.log\n"), 0644))

	err := gitignore.RemoveEntries(path, []string{".melon/"})
	require.NoError(t, err)

	data, _ := os.ReadFile(path)
	assert.Contains(t, string(data), "*.log")
}

// --- ContainsEntry ---

func TestContainsEntry_ReturnsTrueWhenPresent(t *testing.T) {
	path := gitignorePath(t)
	require.NoError(t, os.WriteFile(path, []byte("*.log\n.melon/\n"), 0644))

	ok, err := gitignore.ContainsEntry(path, ".melon/")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestContainsEntry_ReturnsFalseWhenAbsent(t *testing.T) {
	path := gitignorePath(t)
	require.NoError(t, os.WriteFile(path, []byte("*.log\n"), 0644))

	ok, err := gitignore.ContainsEntry(path, ".melon/")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestContainsEntry_ReturnsFalseWhenFileAbsent(t *testing.T) {
	path := gitignorePath(t)
	ok, err := gitignore.ContainsEntry(path, ".melon/")
	require.NoError(t, err)
	assert.False(t, ok)
}
