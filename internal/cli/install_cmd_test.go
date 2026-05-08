package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestRunInstall_UpdatedDep_OldVersionRemovedFromStore verifies that when a dep
// is updated to a new version, the old version's store directory is deleted.
func TestRunInstall_UpdatedDep_OldVersionRemovedFromStore(t *testing.T) {
	dir := t.TempDir()

	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "^1.0.0",
	})

	// Existing lock pins skill-a at 1.0.0.
	writeLockfile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/alice/skills/skill-a", Version: "1.0.0", GitTag: "v1.0.0",
			RepoURL: "https://github.com/alice/skills", Subdir: "skill-a", Entrypoint: "SKILL.md"},
	})

	// Existing store entry for the old version.
	oldStore := filepath.Join(dir, ".melon", "github.com-alice-skills-skill-a@1.0.0")
	require.NoError(t, os.MkdirAll(oldStore, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(oldStore, "SKILL.md"), []byte("# old"), 0644))

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

	// Resolver now returns 1.1.0 — simulating an available update.
	resolveVersionFn = func(repoURL, constraint string) (string, string, error) {
		return "1.1.0", "v1.1.0", nil
	}
	fetchManifestFn = func(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
		return manifest.Manifest{}, nil
	}
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		// Create a minimal file so TreeHash works.
		if err := os.MkdirAll(installDir, 0755); err != nil {
			return fetcher.FetchResult{}, err
		}
		if err := os.WriteFile(filepath.Join(installDir, "SKILL.md"), []byte("# new"), 0644); err != nil {
			return fetcher.FetchResult{}, err
		}
		return fetcher.FetchResult{TreeHash: "sha256:abc", Files: []string{"SKILL.md"}}, nil
	}

	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	require.NoError(t, runInstall(cmd, nil))

	// Old version store entry must be gone.
	_, err := os.Stat(oldStore)
	assert.True(t, os.IsNotExist(err), "old store entry must be deleted after update")

	// New version store entry must exist.
	newStore := filepath.Join(dir, ".melon", "github.com-alice-skills-skill-a@1.1.0")
	_, err = os.Stat(newStore)
	assert.NoError(t, err, "new store entry must exist after update")
}

// TestRunInstall_FetchError_PreservesLockAndStore verifies that when a dep
// fetch fails, runInstall returns an error and does NOT modify the existing
// lock file or delete existing store entries.
//
// This is a regression test for a race condition in the TTY install path: the
// fetch goroutine sent the error to the TUI (causing it to quit) before
// returning from fetchDeps and writing fetchErr. p.Run() could return before
// fetchErr was set, so the main goroutine saw fetchErr==nil, wrote an empty
// lock, and deleted all existing .melon/ entries. The fix (<-done channel)
// ensures the goroutine always finishes writing before the main goroutine reads.
// The non-TTY path exercised here shares the same "don't corrupt state on
// error" invariant.
func TestRunInstall_FetchError_PreservesLockAndStore(t *testing.T) {
	dir := t.TempDir()

	// Set up manifest with two deps.
	writeManifest(t, dir, map[string]string{
		"github.com/alice/skills/skill-a": "^1.0.0",
		"github.com/bob/tools/skill-b":    "^2.0.0",
	})

	// Pre-existing lock file with both deps installed.
	existingDeps := []lockfile.LockedDep{
		{Name: "github.com/alice/skills/skill-a", Version: "1.0.0", GitTag: "v1.0.0",
			RepoURL: "https://github.com/alice/skills", Subdir: "skill-a", Entrypoint: "SKILL.md"},
		{Name: "github.com/bob/tools/skill-b", Version: "2.0.0", GitTag: "v2.0.0",
			RepoURL: "https://github.com/bob/tools", Subdir: "skill-b", Entrypoint: "SKILL.md"},
	}
	writeLockfile(t, dir, existingDeps)

	// Pre-existing store entry for skill-a.
	skillAStore := filepath.Join(dir, ".melon", "github.com-alice-skills-skill-a@1.0.0")
	require.NoError(t, os.MkdirAll(skillAStore, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillAStore, "SKILL.md"), []byte("# skill-a"), 0644))

	// Inject fakes to avoid network calls.
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

	resolveVersionFn = func(repoURL, constraint string) (string, string, error) {
		switch repoURL {
		case "https://github.com/alice/skills":
			return "1.0.0", "v1.0.0", nil
		case "https://github.com/bob/tools":
			return "2.0.0", "v2.0.0", nil
		}
		return "", "", errors.New("unexpected repo: " + repoURL)
	}
	fetchManifestFn = func(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
		return manifest.Manifest{}, nil // no transitive deps
	}
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		return fetcher.FetchResult{}, errors.New("simulated network failure")
	}

	origDir := flagDir
	t.Cleanup(func() { flagDir = origDir })
	flagDir = dir

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runInstall(cmd, nil)
	require.Error(t, err, "runInstall must return an error when fetch fails")

	// Lock file must be unchanged — both deps still present.
	lock, loadErr := lockfile.Load(filepath.Join(dir, "melon.lock"))
	require.NoError(t, loadErr)
	assert.Len(t, lock.Dependencies, 2, "lock file must not be emptied on fetch failure")

	// Existing store entry must not have been deleted.
	_, statErr := os.Stat(skillAStore)
	assert.NoError(t, statErr, "store entry must survive a failed install")
}

// TestRunInstall_VendorFalse_SyncsGitignore verifies that when vendor: false,
// install writes .melon/ and managed symlink paths to .gitignore.
func TestRunInstall_VendorFalse_SyncsGitignore(t *testing.T) {
	dir := t.TempDir()

	vendorFalse := false
	m := manifest.Manifest{
		Name:         "test-project",
		Version:      "0.1.0",
		ToolCompat:   []string{"claude-code"},
		Dependencies: map[string]string{"github.com/alice/skills/skill-a": "^1.0.0"},
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

	resolveVersionFn = func(repoURL, constraint string) (string, string, error) {
		return "1.0.0", "v1.0.0", nil
	}
	fetchManifestFn = func(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
		return manifest.Manifest{}, nil
	}
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

	require.NoError(t, runInstall(cmd, nil))

	gitignorePath := filepath.Join(dir, ".gitignore")
	data, err = os.ReadFile(gitignorePath)
	require.NoError(t, err, ".gitignore should have been created")
	content := string(data)

	assert.True(t, strings.Contains(content, ".melon/"), ".gitignore should contain .melon/")
	assert.True(t, strings.Contains(content, ".claude/skills/skill-a"), ".gitignore should contain the skill symlink path")
	assert.True(t, strings.Contains(content, "melon managed"), ".gitignore should contain the melon comment header")
}

// TestRunInstall_VendorTrue_DoesNotTouchGitignore verifies that when vendor: true
// (or absent), install never writes to .gitignore.
func TestRunInstall_VendorTrue_DoesNotTouchGitignore(t *testing.T) {
	dir := t.TempDir()

	writeManifest(t, dir, map[string]string{"github.com/alice/skills/skill-a": "^1.0.0"})

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

	require.NoError(t, runInstall(cmd, nil))

	_, err := os.Stat(filepath.Join(dir, ".gitignore"))
	assert.True(t, os.IsNotExist(err), ".gitignore must not be created when vendor: true")
}

func TestRunInstall_PolicyBlocked_ErrorsBeforeFetch(t *testing.T) {
	dir := t.TempDir()

	m := manifest.Manifest{
		Name:    "test",
		Version: "0.1.0",
		Dependencies: map[string]string{
			"github.com/my-company/approved-skill": "^1.0.0",
			"github.com/public/blocked-skill":      "^1.0.0",
		},
		Policy: &manifest.PolicyConfig{AllowedSources: []string{"github.com/my-company/*"}},
	}
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.yaml"), data, 0644))

	var fetchCalled bool
	origFetch := fetchFn
	fetchFn = func(_ resolver.ResolvedDep, _ string) (fetcher.FetchResult, error) {
		fetchCalled = true
		return fetcher.FetchResult{}, nil
	}
	t.Cleanup(func() { fetchFn = origFetch })

	origDir := flagDir
	flagDir = dir
	t.Cleanup(func() { flagDir = origDir })

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	err = runInstall(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source policy violation")
	assert.Contains(t, err.Error(), "github.com/public/blocked-skill")
	assert.False(t, fetchCalled, "fetch must not be called when policy blocks")
}

func TestRunInstall_PolicyPermitted_Proceeds(t *testing.T) {
	dir := t.TempDir()

	m := manifest.Manifest{
		Name:       "test",
		Version:    "0.1.0",
		ToolCompat: []string{"claude-code"},
		Dependencies: map[string]string{
			"github.com/my-company/approved-skill": "^1.0.0",
		},
		Policy: &manifest.PolicyConfig{AllowedSources: []string{"github.com/my-company/*"}},
	}
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.yaml"), data, 0644))

	origFns := struct {
		resolveVersion func(string, string) (string, string, error)
		fetchManifest  func(string, string, string) (manifest.Manifest, error)
		fetch          func(resolver.ResolvedDep, string) (fetcher.FetchResult, error)
	}{resolveVersionFn, fetchManifestFn, fetchFn}
	t.Cleanup(func() {
		resolveVersionFn = origFns.resolveVersion
		fetchManifestFn = origFns.fetchManifest
		fetchFn = origFns.fetch
	})
	resolveVersionFn = func(_, _ string) (string, string, error) { return "1.0.0", "v1.0.0", nil }
	fetchManifestFn = func(_, _, _ string) (manifest.Manifest, error) { return manifest.Manifest{}, nil }
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		require.NoError(t, os.MkdirAll(installDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(installDir, "SKILL.md"), []byte("# skill"), 0644))
		return fetcher.FetchResult{TreeHash: "sha256:abc", Files: []string{"SKILL.md"}}, nil
	}

	origDir := flagDir
	flagDir = dir
	t.Cleanup(func() { flagDir = origDir })

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	err = runInstall(cmd, nil)
	require.NoError(t, err)

	_, lockErr := lockfile.Load(filepath.Join(dir, "melon.lock"))
	assert.NoError(t, lockErr, "melon.lock should have been written")
}
