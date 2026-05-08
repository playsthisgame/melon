package index_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/playsthisgame/melon/internal/index"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleIndexYAML = `
skills:
  - name: github.com/alice/skill-a
    description: Skill A
    author: alice
    tags: [go]
    featured: true
  - name: github.com/bob/skill-b
    description: Skill B
    author: bob
    tags: [python]
    featured: false
`

func TestFetch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleIndexYAML))
	}))
	defer srv.Close()

	entries, err := index.Fetch(srv.URL)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "github.com/alice/skill-a", entries[0].Name)
	assert.Equal(t, "github.com/bob/skill-b", entries[1].Name)
}

func TestFetch_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := index.Fetch(srv.URL)
	assert.Error(t, err)
}

func TestFetch_InvalidYAML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(":\tinvalid: {{{"))
	}))
	defer srv.Close()

	_, err := index.Fetch(srv.URL)
	assert.Error(t, err)
}

func TestSearch_FeaturedFirst(t *testing.T) {
	entries := []index.Entry{
		{Name: "github.com/bob/skill-b", Featured: false},
		{Name: "github.com/alice/skill-a", Featured: true},
	}
	results := index.Search(entries, "skill")
	require.Len(t, results, 2)
	assert.True(t, results[0].Featured)
	assert.Equal(t, "github.com/alice/skill-a", results[0].Name)
}
