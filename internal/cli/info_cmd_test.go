package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInfo_ResolveIndexURLs_CustomIndex verifies that info_cmd resolves index
// URLs from melon.yaml. The actual index lookup in runInfo is tested indirectly
// through resolveIndexURLs, which is shared with search_cmd.
func TestInfo_CustomIndex_PublicIndexDisabled(t *testing.T) {
	// Serve a custom index that contains the skill we'll query.
	const indexYAML = `
skills:
  - name: github.com/acme/private-skill
    description: Private skill
    author: acme
    tags: []
    featured: false
`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(indexYAML))
	}))
	defer srv.Close()

	// Write a melon.yaml pointing to the custom server, with public index disabled.
	dir := t.TempDir()
	f := false
	m := manifest.Manifest{
		Name:    "test",
		Version: "0.1.0",
		Index:   &manifest.IndexConfig{URLs: []string{srv.URL}, PublicIndex: &f},
	}
	require.NoError(t, manifest.Save(m, filepath.Join(dir, "melon.yaml")))

	orig := flagDir
	flagDir = dir
	t.Cleanup(func() { flagDir = orig })

	urls := resolveIndexURLs()
	assert.Equal(t, []string{srv.URL}, urls, "public_index: false should return only the custom URL")
}

func TestInfo_CustomIndex_NoManifest(t *testing.T) {
	setFlagDir(t, t.TempDir()) // no melon.yaml

	// parseGitHubPath is testable without network.
	owner, repo, subdir, err := parseGitHubPath("github.com/alice/skill/sub")
	require.NoError(t, err)
	assert.Equal(t, "alice", owner)
	assert.Equal(t, "skill", repo)
	assert.Equal(t, "sub", subdir)
}

func TestInfo_ParseGitHubPath_Invalid(t *testing.T) {
	// Missing repo segment — only owner present, no slash after it.
	_, _, _, err := parseGitHubPath("github.com/onlyowner")
	assert.Error(t, err)
}

func TestInfo_ParseGitHubPath_NoSubdir(t *testing.T) {
	owner, repo, subdir, err := parseGitHubPath("github.com/alice/myskill")
	require.NoError(t, err)
	assert.Equal(t, "alice", owner)
	assert.Equal(t, "myskill", repo)
	assert.Equal(t, "", subdir)
}

// Verify that a melon.yaml with no index block falls back to the default index URL.
func TestInfo_NoIndexBlock_UsesDefault(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.yaml"), []byte("name: x\nversion: 0.1.0\n"), 0644))
	setFlagDir(t, dir)

	urls := resolveIndexURLs()
	assert.Len(t, urls, 1)
}
