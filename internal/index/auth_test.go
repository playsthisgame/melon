package index

import (
	"errors"
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

func TestFetch_PrivateGitHubIndex_ReturnsAccessError(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	var auth string
	srv := privateServer(t, "secret", &auth)
	treatAsGitHub(t, srv)

	_, err := Fetch(srv.URL + "/index.yaml")
	var accessErr *AccessError
	require.ErrorAs(t, err, &accessErr)
	assert.False(t, accessErr.HadToken)
	assert.Equal(t, "404 Not Found", accessErr.Status)
}

func TestFetch_OtherHost404_IsNotAccessError(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	_, err := Fetch(srv.URL + "/index.yaml")
	var accessErr *AccessError
	assert.False(t, errors.As(err, &accessErr))
}

// A bad or expired token makes GitHub answer 404 even for public content;
// Fetch must fall back to an anonymous request so the public index still loads.
func TestFetch_BadToken_PublicIndexStillLoads(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "expired")
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		calls = append(calls, auth)
		if auth != "" { // any token is rejected; anonymous is fine (public)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(authIndexYAML))
	}))
	t.Cleanup(srv.Close)
	treatAsGitHub(t, srv)

	entries, err := Fetch(srv.URL + "/index.yaml")
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, []string{"Bearer expired", ""}, calls, "one retry without the token")
}

func TestFetch_GoodToken_NoRetry(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "secret")
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(authIndexYAML))
	}))
	t.Cleanup(srv.Close)
	treatAsGitHub(t, srv)

	_, err := Fetch(srv.URL + "/index.yaml")
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}
