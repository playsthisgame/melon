package cli

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"

	"github.com/playsthisgame/melon/internal/index"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const privateRawURL = "https://raw.githubusercontent.com/acme/registry/main/index.yaml"

func stubIndexFetch(t *testing.T, fn func(string) ([]index.Entry, error)) {
	t.Helper()
	orig := indexFetchFn
	t.Cleanup(func() { indexFetchFn = orig })
	indexFetchFn = fn
}

func stubGH(t *testing.T, installed bool) {
	t.Helper()
	orig := lookPathFn
	t.Cleanup(func() { lookPathFn = orig })
	lookPathFn = func(string) (string, error) {
		if installed {
			return "/usr/local/bin/gh", nil
		}
		return "", exec.ErrNotFound
	}
}

func newSearchTestCmd() (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}

// A private index that fails while the public index succeeds: public results
// are still returned, and the failure is reported, not dropped.
func TestLoadIndexEntries_WarnsOnFailedIndexButKeepsOthers(t *testing.T) {
	stubIndexFetch(t, func(u string) ([]index.Entry, error) {
		if u == privateRawURL {
			return nil, &index.AccessError{URL: u, Status: "404 Not Found"}
		}
		return []index.Entry{{Name: "github.com/pub/skill"}}, nil
	})
	stubGH(t, true)

	cmd, stdout, stderr := newSearchTestCmd()
	entries, failed := loadIndexEntries(cmd, []string{privateRawURL, "public"})
	assert.Len(t, entries, 1)
	assert.Equal(t, 1, failed)
	assert.Contains(t, stderr.String(), "GitHub refused access to index github.com/acme/registry/index.yaml (404 Not Found)")
	assert.Empty(t, stdout.String(), "warnings stay off stdout so piped results are clean")
}

func TestPrintIndexError_NoToken_ExplainsAndSuggestsGH(t *testing.T) {
	stubGH(t, true)
	var buf bytes.Buffer
	printIndexError(&buf, &index.AccessError{URL: privateRawURL, Status: "404 Not Found"})
	out := buf.String()
	assert.Contains(t, out, "not a melon error")
	assert.Contains(t, out, "private repository when no token is sent")
	assert.Contains(t, out, "export GITHUB_TOKEN=$(gh auth token)")
}

func TestPrintIndexError_NoToken_NoGH_GenericFix(t *testing.T) {
	stubGH(t, false)
	var buf bytes.Buffer
	printIndexError(&buf, &index.AccessError{URL: privateRawURL, Status: "404 Not Found"})
	assert.Contains(t, buf.String(), "export GITHUB_TOKEN=<a token that can read the repository>")
	assert.NotContains(t, buf.String(), "gh auth token")
}

func TestPrintIndexError_WithToken_NamesVariableAndChecks(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "x")
	var buf bytes.Buffer
	printIndexError(&buf, &index.AccessError{URL: privateRawURL, Status: "404 Not Found", HadToken: true})
	out := buf.String()
	assert.Contains(t, out, "using the token in GH_TOKEN")
	assert.Contains(t, out, "Contents: read")
	assert.NotContains(t, out, "export GITHUB_TOKEN")
}

func TestPrintIndexError_OtherErrorsUnchanged(t *testing.T) {
	var buf bytes.Buffer
	printIndexError(&buf, errors.New("index: parse x: bad yaml"))
	assert.Equal(t, "warning: could not load index: index: parse x: bad yaml\n", buf.String())
}

func TestDisplayIndexURL(t *testing.T) {
	assert.Equal(t, "github.com/acme/registry/index.yaml", displayIndexURL(privateRawURL))
	assert.Equal(t, "github.com/acme/registry/blob/dev/skills/index.yaml",
		displayIndexURL("https://raw.githubusercontent.com/acme/registry/dev/skills/index.yaml"))
	assert.Equal(t, "https://example.com/index.yaml", displayIndexURL("https://example.com/index.yaml"))
}

// When an index failed and nothing matched, the result line must not read
// like a plain "no matches".
func TestRunSearch_NoResults_SaysIndicesFailed(t *testing.T) {
	setFlagDir(t, t.TempDir()) // no melon.yaml: only the public index
	stubIndexFetch(t, func(u string) ([]index.Entry, error) {
		return nil, &index.AccessError{URL: u, Status: "404 Not Found"}
	})
	stubGH(t, true)

	cmd, stdout, stderr := newSearchTestCmd()
	require.NoError(t, runSearch(cmd, []string{"deploy"}))
	assert.Contains(t, stdout.String(), `No skills found for "deploy" in the indices melon could load (1 of 1 failed, see above).`)
	assert.Contains(t, stderr.String(), "not a melon error")
}

func TestLoadIndexEntries_EarlierIndexWinsOnDuplicates(t *testing.T) {
	stubIndexFetch(t, func(u string) ([]index.Entry, error) {
		return []index.Entry{{Name: "github.com/x/skill", Description: u}}, nil
	})
	cmd, _, _ := newSearchTestCmd()
	entries, failed := loadIndexEntries(cmd, []string{"custom", "public"})
	assert.Len(t, entries, 1)
	assert.Zero(t, failed)
	assert.Equal(t, "custom", entries[0].Description)
}
