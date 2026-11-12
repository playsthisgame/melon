package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/playsthisgame/melon/internal/index"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestLoadIndexEntries_WarnsOnFailedIndexButKeepsOthers covers a private
// index that fails to load while the public index succeeds: results from the
// public index are still returned, and the failure is reported, not dropped.
func TestLoadIndexEntries_WarnsOnFailedIndexButKeepsOthers(t *testing.T) {
	orig := indexFetchFn
	t.Cleanup(func() { indexFetchFn = orig })
	indexFetchFn = func(u string) ([]index.Entry, error) {
		if u == "private" {
			return nil, errors.New("index: fetch private: unexpected status 404 Not Found (set GITHUB_TOKEN)")
		}
		return []index.Entry{{Name: "github.com/pub/skill"}}, nil
	}

	cmd := &cobra.Command{}
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	entries := loadIndexEntries(cmd, []string{"private", "public"})
	assert.Len(t, entries, 1)
	assert.Contains(t, stderr.String(), "warning: could not load index: index: fetch private")
	assert.Empty(t, stdout.String(), "warnings stay off stdout so piped results are clean")
}

func TestLoadIndexEntries_EarlierIndexWinsOnDuplicates(t *testing.T) {
	orig := indexFetchFn
	t.Cleanup(func() { indexFetchFn = orig })
	indexFetchFn = func(u string) ([]index.Entry, error) {
		return []index.Entry{{Name: "github.com/x/skill", Description: u}}, nil
	}
	cmd := &cobra.Command{}
	cmd.SetErr(&bytes.Buffer{})
	entries := loadIndexEntries(cmd, []string{"custom", "public"})
	assert.Len(t, entries, 1)
	assert.Equal(t, "custom", entries[0].Description)
}
