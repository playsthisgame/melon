package index

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authIndexYAML = "skills:\n  - name: github.com/acme/private/skills/deploy\n    description: Deploy\n"

// privateServer serves the index only to requests carrying the token, and
// answers 404 otherwise, like GitHub does for private repositories.
func privateServer(t *testing.T, token string, gotAuth *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotAuth = r.Header.Get("Authorization")
		if *gotAuth != "Bearer "+token {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(authIndexYAML))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// treatAsGitHub makes srv's host count as a GitHub host for this test.
func treatAsGitHub(t *testing.T, srv *httptest.Server) {
	t.Helper()
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	githubHosts[u.Hostname()] = true
	t.Cleanup(func() { delete(githubHosts, u.Hostname()) })
}

func TestFetch_PrivateGitHubIndex_SendsToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "secret")
	t.Setenv("GH_TOKEN", "")
	var auth string
	srv := privateServer(t, "secret", &auth)
	treatAsGitHub(t, srv)

	entries, err := Fetch(srv.URL + "/acme/private/main/index.yaml")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "github.com/acme/private/skills/deploy", entries[0].Name)
	assert.Equal(t, "Bearer secret", auth)
}

func TestFetch_FallsBackToGHToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "from-gh")
	var auth string
	srv := privateServer(t, "from-gh", &auth)
	treatAsGitHub(t, srv)

	_, err := Fetch(srv.URL + "/index.yaml")
	require.NoError(t, err)
	assert.Equal(t, "Bearer from-gh", auth)
}

func TestFetch_NonGitHubHost_NeverGetsToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "secret")
	var auth string
	srv := privateServer(t, "secret", &auth) // not registered as a GitHub host

	_, err := Fetch(srv.URL + "/index.yaml")
	require.Error(t, err)
	assert.Empty(t, auth, "token must not leak to non-GitHub hosts")
	assert.NotContains(t, err.Error(), "GITHUB_TOKEN", "no GitHub hint for other hosts")
}

func TestFetch_PrivateGitHubIndex_NoToken_Hints(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	var auth string
	srv := privateServer(t, "secret", &auth)
	treatAsGitHub(t, srv)

	_, err := Fetch(srv.URL + "/index.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "private repository, set GITHUB_TOKEN")
	assert.Contains(t, err.Error(), srv.URL+"/index.yaml", "error names the index")
}

func TestFetch_PrivateGitHubIndex_WrongToken_Hints(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "wrong")
	var auth string
	srv := privateServer(t, "secret", &auth)
	treatAsGitHub(t, srv)

	_, err := Fetch(srv.URL + "/index.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GITHUB_TOKEN can read this repository")
}
