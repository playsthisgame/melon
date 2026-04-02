package fetcher_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── ParseDepName ────────────────────────────────────────────────────────────

func TestParseDepName_RepoRoot(t *testing.T) {
	repoURL, subdir := fetcher.ParseDepName("github.com/alice/pdf-skill")
	assert.Equal(t, "https://github.com/alice/pdf-skill", repoURL)
	assert.Equal(t, "", subdir)
}

func TestParseDepName_Monorepo(t *testing.T) {
	repoURL, subdir := fetcher.ParseDepName("github.com/anthropics/skills/skills/skill-creator")
	assert.Equal(t, "https://github.com/anthropics/skills", repoURL)
	assert.Equal(t, "skills/skill-creator", subdir)
}

func TestParseDepName_SingleLevel(t *testing.T) {
	repoURL, subdir := fetcher.ParseDepName("github.com/bob/agents/agents/base-agent")
	assert.Equal(t, "https://github.com/bob/agents", repoURL)
	assert.Equal(t, "agents/base-agent", subdir)
}

func TestParseDepName_GitHubWebUIURL(t *testing.T) {
	// GitHub web UI URLs include tree/<branch>/ — these should be stripped.
	repoURL, subdir := fetcher.ParseDepName("github.com/anthropics/skills/tree/main/skills/skill-creator")
	assert.Equal(t, "https://github.com/anthropics/skills", repoURL)
	assert.Equal(t, "skills/skill-creator", subdir)
}

func TestParseDepName_GitHubWebUIURL_BranchOnly(t *testing.T) {
	// tree/<branch> with no further subdir → root of repo.
	repoURL, subdir := fetcher.ParseDepName("github.com/alice/myrepo/tree/develop")
	assert.Equal(t, "https://github.com/alice/myrepo", repoURL)
	assert.Equal(t, "", subdir)
}

// ── TreeHash ─────────────────────────────────────────────────────────────────

func TestTreeHash_Deterministic(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "refs"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "refs", "api.md"), []byte("api"), 0644))

	h1, f1, err := fetcher.TreeHash(dir)
	require.NoError(t, err)
	h2, f2, err := fetcher.TreeHash(dir)
	require.NoError(t, err)

	assert.Equal(t, h1, h2, "hash must be deterministic")
	assert.Equal(t, f1, f2)
	assert.Equal(t, []string{"SKILL.md", "refs/api.md"}, f1)
	assert.True(t, strings.HasPrefix(h1, "sha256:"), "hash should have sha256: prefix")
}

func TestTreeHash_ChangesOnFileEdit(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "SKILL.md")
	require.NoError(t, os.WriteFile(file, []byte("v1"), 0644))

	h1, _, err := fetcher.TreeHash(dir)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(file, []byte("v2"), 0644))
	h2, _, err := fetcher.TreeHash(dir)
	require.NoError(t, err)

	assert.NotEqual(t, h1, h2, "hash must change when file content changes")
}

func TestTreeHash_ChangesOnNewFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0644))

	h1, _, err := fetcher.TreeHash(dir)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "extra.md"), []byte("extra"), 0644))
	h2, _, err := fetcher.TreeHash(dir)
	require.NoError(t, err)

	assert.NotEqual(t, h1, h2)
}

// ── Integration test — requires network + git ────────────────────────────────

// TestIntegration_FetchSkillCreator installs github.com/anthropics/skills/skills/skill-creator
// into a temp directory and verifies the full directory tree and mln.lock fields.
//
// Gated on MLN_INTEGRATION=1 to avoid running in normal CI.
func TestIntegration_FetchSkillCreator(t *testing.T) {
	if os.Getenv("MLN_INTEGRATION") == "" {
		t.Skip("set MLN_INTEGRATION=1 to run network integration tests")
	}

	const depName = "github.com/anthropics/skills/skills/skill-creator"
	repoURL, subdir := fetcher.ParseDepName(depName)
	assert.Equal(t, "https://github.com/anthropics/skills", repoURL)
	assert.Equal(t, "skills/skill-creator", subdir)

	// Find the latest available tag.
	version, gitTag, err := fetcher.LatestMatchingVersion(repoURL, "^0.0.0")
	if err != nil {
		// Fallback: try any v* tag by using a very permissive approach.
		t.Logf("^0.0.0 failed (%v), attempting ^1.0.0", err)
		version, gitTag, err = fetcher.LatestMatchingVersion(repoURL, "^1.0.0")
	}
	require.NoError(t, err, "should find at least one semver tag in github.com/anthropics/skills")
	t.Logf("resolved %s → %s (%s)", depName, version, gitTag)

	dep := resolver.ResolvedDep{
		Name:       depName,
		Version:    version,
		RepoURL:    repoURL,
		Subdir:     subdir,
		GitTag:     gitTag,
		Entrypoint: "SKILL.md",
	}

	installDir := filepath.Join(t.TempDir(), "skill-creator@"+version)

	result, err := fetcher.Fetch(dep, installDir)
	require.NoError(t, err)

	// Verify tree hash format.
	assert.True(t, strings.HasPrefix(result.TreeHash, "sha256:"), "tree_hash should start with sha256:")
	assert.NotEmpty(t, result.Files, "files list must not be empty")

	// Verify the entrypoint file exists on disk.
	entrypointPath := filepath.Join(installDir, dep.Entrypoint)
	_, err = os.Stat(entrypointPath)
	require.NoError(t, err, "SKILL.md must exist in the installed directory")

	// Verify every file in result.Files exists on disk.
	for _, f := range result.Files {
		fPath := filepath.Join(installDir, filepath.FromSlash(f))
		_, statErr := os.Stat(fPath)
		assert.NoError(t, statErr, "declared file %s must exist on disk", f)
	}

	// Verify tree hash is reproducible.
	recomputedHash, recomputedFiles, err := fetcher.TreeHash(installDir)
	require.NoError(t, err)
	assert.Equal(t, result.TreeHash, recomputedHash, "tree hash must be reproducible")
	assert.Equal(t, result.Files, recomputedFiles)

	// Verify idempotency: second Fetch with known tree hash must skip the download.
	dep.TreeHash = result.TreeHash
	result2, err := fetcher.Fetch(dep, installDir)
	require.NoError(t, err)
	assert.Equal(t, result.TreeHash, result2.TreeHash, "idempotent fetch must return same hash")
}
