package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	gh "github.com/playsthisgame/melon/internal/github"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/spf13/cobra"
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

// newInfoJSONCmd builds a cobra.Command wired to runInfo with --json and
// captures stdout/stderr into buffers for inspection.
func newInfoJSONCmd(t *testing.T) (cmd *cobra.Command, outBuf, errBuf *bytes.Buffer) {
	t.Helper()
	cmd = &cobra.Command{RunE: runInfo}
	outBuf = &bytes.Buffer{}
	errBuf = &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	return cmd, outBuf, errBuf
}

// fakeGitHubAndIndexServer returns an httptest.Server that serves:
//   - GET /index.yaml  → the provided index YAML
//   - GET /repos/<owner>/<repo>  → {"description": repoAbout}
//   - GET /repos/<owner>/<repo>/tags  → [{"name":"v1.0.0"},{"name":"v0.9.0"}] or []
//   - GET /repos/<owner>/<repo>/branches  → [{"name":"main"}] (used when noTags)
func fakeGitHubAndIndexServer(t *testing.T, indexYAML, owner, repo, repoAbout string, noTags bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/index.yaml":
			w.Header().Set("Content-Type", "text/yaml")
			fmt.Fprint(w, indexYAML)
		case r.URL.Path == fmt.Sprintf("/repos/%s/%s", owner, repo):
			fmt.Fprintf(w, `{"description":%q}`, repoAbout)
		case r.URL.Path == fmt.Sprintf("/repos/%s/%s/tags", owner, repo):
			if noTags {
				fmt.Fprint(w, `[]`)
			} else {
				fmt.Fprint(w, `[{"name":"v1.0.0"},{"name":"v0.9.0"}]`)
			}
		case r.URL.Path == fmt.Sprintf("/repos/%s/%s/branches", owner, repo):
			fmt.Fprint(w, `[{"name":"main"},{"name":"dev"}]`)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestInfo_JSON_InIndexSkill(t *testing.T) {
	const owner, repo = "acme", "my-skill"
	indexYAML := fmt.Sprintf(`skills:
  - name: github.com/%s/%s
    description: Acme skill
    author: acme-team
    tags: []
    featured: false
`, owner, repo)

	srv := fakeGitHubAndIndexServer(t, indexYAML, owner, repo, "", false)
	defer srv.Close()

	dir := t.TempDir()
	f := false
	m := manifest.Manifest{
		Name:    "test",
		Version: "0.1.0",
		Index:   &manifest.IndexConfig{URLs: []string{srv.URL + "/index.yaml"}, PublicIndex: &f},
	}
	require.NoError(t, manifest.Save(m, filepath.Join(dir, "melon.yaml")))
	setFlagDir(t, dir)

	origJSON := flagInfoJSON
	flagInfoJSON = true
	t.Cleanup(func() { flagInfoJSON = origJSON })

	// Patch the GitHub client base URL to point at the fake server.
	origGHClientFn := newGHClientFn
	newGHClientFn = func() *gh.Client { return gh.NewWithBase(srv.URL) }
	t.Cleanup(func() { newGHClientFn = origGHClientFn })

	cmd, outBuf, _ := newInfoJSONCmd(t)
	err := runInfo(cmd, []string{fmt.Sprintf("github.com/%s/%s", owner, repo)})
	require.NoError(t, err)

	var out infoJSONOutput
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &out))
	assert.Equal(t, fmt.Sprintf("github.com/%s/%s", owner, repo), out.Name)
	assert.Equal(t, "Acme skill", out.Description)
	assert.Equal(t, "acme-team", out.Author)
	assert.Equal(t, "v1.0.0", out.LatestVersion)
	assert.Equal(t, []string{"v1.0.0", "v0.9.0"}, out.Versions)
	assert.Empty(t, out.Branches)
}

func TestInfo_JSON_NotInIndexSkill(t *testing.T) {
	const owner, repo = "bob", "cool-skill"

	srv := fakeGitHubAndIndexServer(t, "skills: []\n", owner, repo, "A cool skill from Bob", false)
	defer srv.Close()

	dir := t.TempDir()
	f := false
	m := manifest.Manifest{
		Name:    "test",
		Version: "0.1.0",
		Index:   &manifest.IndexConfig{URLs: []string{srv.URL + "/index.yaml"}, PublicIndex: &f},
	}
	require.NoError(t, manifest.Save(m, filepath.Join(dir, "melon.yaml")))
	setFlagDir(t, dir)

	origJSON := flagInfoJSON
	flagInfoJSON = true
	t.Cleanup(func() { flagInfoJSON = origJSON })

	origGHClientFn := newGHClientFn
	newGHClientFn = func() *gh.Client { return gh.NewWithBase(srv.URL) }
	t.Cleanup(func() { newGHClientFn = origGHClientFn })

	cmd, outBuf, _ := newInfoJSONCmd(t)
	err := runInfo(cmd, []string{fmt.Sprintf("github.com/%s/%s", owner, repo)})
	require.NoError(t, err)

	var out infoJSONOutput
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &out))
	assert.Equal(t, "A cool skill from Bob", out.Description)
	assert.Equal(t, "", out.Author)
	assert.Equal(t, "v1.0.0", out.LatestVersion)
}

func TestInfo_JSON_NoTags(t *testing.T) {
	const owner, repo = "charlie", "branch-only"

	srv := fakeGitHubAndIndexServer(t, "skills: []\n", owner, repo, "No tags here", true)
	defer srv.Close()

	dir := t.TempDir()
	f := false
	m := manifest.Manifest{
		Name:    "test",
		Version: "0.1.0",
		Index:   &manifest.IndexConfig{URLs: []string{srv.URL + "/index.yaml"}, PublicIndex: &f},
	}
	require.NoError(t, manifest.Save(m, filepath.Join(dir, "melon.yaml")))
	setFlagDir(t, dir)

	origJSON := flagInfoJSON
	flagInfoJSON = true
	t.Cleanup(func() { flagInfoJSON = origJSON })

	origGHClientFn := newGHClientFn
	newGHClientFn = func() *gh.Client { return gh.NewWithBase(srv.URL) }
	t.Cleanup(func() { newGHClientFn = origGHClientFn })

	cmd, outBuf, _ := newInfoJSONCmd(t)
	err := runInfo(cmd, []string{fmt.Sprintf("github.com/%s/%s", owner, repo)})
	require.NoError(t, err)

	var out infoJSONOutput
	require.NoError(t, json.Unmarshal(outBuf.Bytes(), &out))
	assert.Empty(t, out.Versions)
	assert.Equal(t, []string{"main", "dev"}, out.Branches)
	assert.Equal(t, "", out.LatestVersion)
}

func TestInfo_JSON_ErrorInvalidPath(t *testing.T) {
	setFlagDir(t, t.TempDir())

	origJSON := flagInfoJSON
	flagInfoJSON = true
	t.Cleanup(func() { flagInfoJSON = origJSON })

	cmd, outBuf, errBuf := newInfoJSONCmd(t)
	err := runInfo(cmd, []string{"github.com/onlyowner"})
	require.Error(t, err)
	assert.Empty(t, outBuf.String())
	assert.Contains(t, errBuf.String(), `"error"`)
}
