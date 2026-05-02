package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestRunAdd_VendorFalse_AppendsGitignore verifies that when vendor: false,
// mln add appends the new skill's symlink path to .gitignore via runInstall.
func TestRunAdd_VendorFalse_AppendsGitignore(t *testing.T) {
	dir := t.TempDir()

	vendorFalse := false
	m := manifest.Manifest{
		Name:         "test-project",
		Version:      "0.1.0",
		ToolCompat:   []string{"claude-code"},
		Dependencies: map[string]string{},
		Vendor:       &vendorFalse,
	}
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.yaml"), data, 0644))

	orig := struct {
		resolveVersion func(string, string) (string, string, error)
		fetchManifest  func(string, string, string) (manifest.Manifest, error)
		fetch          func(resolver.ResolvedDep, string) (fetcher.FetchResult, error)
	}{resolveVersionFn, fetchManifestFn, fetchFn}
	t.Cleanup(func() {
		resolveVersionFn = orig.resolveVersion
		fetchManifestFn = orig.fetchManifest
		fetchFn = orig.fetch
	})

	resolveVersionFn = func(repoURL, constraint string) (string, string, error) { return "1.0.0", "v1.0.0", nil }
	fetchManifestFn = func(repoURL, gitTag, subdir string) (manifest.Manifest, error) { return manifest.Manifest{}, nil }
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		require.NoError(t, os.MkdirAll(installDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(installDir, "SKILL.md"), []byte("# skill"), 0644))
		return fetcher.FetchResult{TreeHash: "sha256:abc", Files: []string{"SKILL.md"}}, nil
	}

	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetIn(strings.NewReader(""))

	// Inject a pre-resolved constraint to bypass LatestTag network call.
	err = runAdd(cmd, []string{"github.com/alice/skills/skill-a@^1.0.0"})
	require.NoError(t, err)

	data, err = os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err, ".gitignore should have been created")
	content := string(data)
	assert.Contains(t, content, ".melon/")
	assert.Contains(t, content, ".claude/skills/skill-a")
}
