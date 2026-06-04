package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const diffDepName = "github.com/owner/repo/skills/demo"

// diffTestSetup writes a manifest + lock for diffDepName and points flagDir at a
// temp dir. constraint controls the dep's version constraint in melon.yaml.
func diffTestSetup(t *testing.T, constraint string) string {
	t.Helper()
	dir := t.TempDir()

	m := manifest.Manifest{
		Name:         "test",
		Version:      "0.1.0",
		Dependencies: map[string]string{diffDepName: constraint},
	}
	require.NoError(t, manifest.Save(m, filepath.Join(dir, "melon.yaml")))

	lf := lockfile.LockFile{
		Dependencies: []lockfile.LockedDep{{
			Name:       diffDepName,
			Version:    "1.0.0",
			GitTag:     "v1.0.0",
			RepoURL:    "https://github.com/owner/repo",
			Subdir:     "skills/demo",
			Entrypoint: "SKILL.md",
			TreeHash:   "sha256:from",
			Files:      []string{"SKILL.md"},
		}},
	}
	require.NoError(t, lockfile.Save(lf, filepath.Join(dir, "melon.lock")))

	origDir := flagDir
	flagDir = dir
	t.Cleanup(func() { flagDir = origDir })

	return dir
}

// fakeFetch returns a fetchFn that writes the given per-file content for a
// version and reports the given tree hash.
func setFakeFetch(t *testing.T, contentByVersion map[string]map[string]string, hashByVersion map[string]string) {
	t.Helper()
	orig := fetchFn
	t.Cleanup(func() { fetchFn = orig })
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		require.NoError(t, os.MkdirAll(installDir, 0755))
		files := contentByVersion[dep.Version]
		names := make([]string, 0, len(files))
		for name, content := range files {
			p := filepath.Join(installDir, filepath.FromSlash(name))
			require.NoError(t, os.MkdirAll(filepath.Dir(p), 0755))
			require.NoError(t, os.WriteFile(p, []byte(content), 0644))
			names = append(names, name)
		}
		return fetcher.FetchResult{TreeHash: hashByVersion[dep.Version], Files: names}, nil
	}
}

func setFakeResolve(t *testing.T, version, tag string, err error) *string {
	t.Helper()
	orig := resolveVersionFn
	t.Cleanup(func() { resolveVersionFn = orig })
	captured := new(string)
	resolveVersionFn = func(_, arg string) (string, string, error) {
		*captured = arg
		return version, tag, err
	}
	return captured
}

func resetDiffFlags(t *testing.T) {
	t.Helper()
	origStat, origNoColor := flagDiffStat, flagDiffNoColor
	flagDiffStat, flagDiffNoColor = false, false
	t.Cleanup(func() { flagDiffStat, flagDiffNoColor = origStat, origNoColor })
}

func runDiffCapture(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := runDiff(cmd, args)
	return buf.String(), err
}

func TestRunDiff_ChangedFile_ShowsHunk(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	setFakeResolve(t, "1.1.0", "v1.1.0", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0": {"SKILL.md": "# Demo\nold line\n"},
			"1.1.0": {"SKILL.md": "# Demo\nnew line\n"},
		},
		map[string]string{"1.0.0": "sha256:from", "1.1.0": "sha256:to"},
	)

	out, err := runDiffCapture(t, diffDepName)
	require.NoError(t, err)
	assert.Contains(t, out, "diff "+diffDepName)
	assert.Contains(t, out, "-old line")
	assert.Contains(t, out, "+new line")
}

func TestRunDiff_AddedAndRemovedFiles(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	setFakeResolve(t, "1.1.0", "v1.1.0", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0": {"SKILL.md": "# Demo\n", "old.md": "gone\n"},
			"1.1.0": {"SKILL.md": "# Demo\n", "new.md": "fresh\n"},
		},
		map[string]string{"1.0.0": "sha256:from", "1.1.0": "sha256:to"},
	)

	out, err := runDiffCapture(t, diffDepName)
	require.NoError(t, err)
	// new.md added: present on the +++ side as /dev/null on the from side.
	assert.Contains(t, out, "+fresh")
	assert.Contains(t, out, "-gone")
	assert.Contains(t, out, "/dev/null")
}

func TestRunDiff_IdenticalTreeHash_NoChanges(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	setFakeResolve(t, "1.1.0", "v1.1.0", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0": {"SKILL.md": "# Demo\n"},
			"1.1.0": {"SKILL.md": "# Demo\n"},
		},
		// Same tree hash on both sides triggers the fast path.
		map[string]string{"1.0.0": "sha256:from", "1.1.0": "sha256:from"},
	)

	out, err := runDiffCapture(t, diffDepName)
	require.NoError(t, err)
	assert.Contains(t, out, "No changes")
}

func TestRunDiff_ExplicitVersionTarget_OverridesConstraint(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	captured := setFakeResolve(t, "2.0.0", "v2.0.0", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0": {"SKILL.md": "v1\n"},
			"2.0.0": {"SKILL.md": "v2\n"},
		},
		map[string]string{"1.0.0": "sha256:from", "2.0.0": "sha256:to"},
	)

	out, err := runDiffCapture(t, diffDepName+"@2.0.0")
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", *captured, "explicit target should be passed to resolver, not the constraint")
	assert.Contains(t, out, "1.0.0 → 2.0.0")
}

func TestRunDiff_ExplicitBranchTarget(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	captured := setFakeResolve(t, "abc123", "main", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0":  {"SKILL.md": "v1\n"},
			"abc123": {"SKILL.md": "branch\n"},
		},
		map[string]string{"1.0.0": "sha256:from", "abc123": "sha256:to"},
	)

	out, err := runDiffCapture(t, diffDepName+"@main")
	require.NoError(t, err)
	assert.Equal(t, "main", *captured)
	assert.Contains(t, out, "+branch")
}

func TestRunDiff_DepNotInLock_Errors(t *testing.T) {
	resetDiffFlags(t)
	dir := t.TempDir()
	m := manifest.Manifest{
		Name:         "test",
		Version:      "0.1.0",
		Dependencies: map[string]string{diffDepName: "^1.0.0"},
	}
	require.NoError(t, manifest.Save(m, filepath.Join(dir, "melon.yaml")))
	origDir := flagDir
	flagDir = dir
	t.Cleanup(func() { flagDir = origDir })

	_, err := runDiffCapture(t, diffDepName)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "melon install")
}

func TestRunDiff_DepNotInManifest_Errors(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")

	_, err := runDiffCapture(t, "github.com/owner/repo/skills/other")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a dependency")
}

func TestRunDiff_UnresolvableTarget_Errors(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	setFakeResolve(t, "", "", assert.AnError)

	_, err := runDiffCapture(t, diffDepName+"@9.9.9")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve target")
}

func TestRunDiff_BranchPinnedNoTarget_Errors(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "main")

	_, err := runDiffCapture(t, diffDepName)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch-pinned")
}

func TestRunDiff_StatMode_SummaryOnly(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	flagDiffStat = true
	setFakeResolve(t, "1.1.0", "v1.1.0", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0": {"SKILL.md": "a\nb\n"},
			"1.1.0": {"SKILL.md": "a\nc\nd\n"},
		},
		map[string]string{"1.0.0": "sha256:from", "1.1.0": "sha256:to"},
	)

	out, err := runDiffCapture(t, diffDepName)
	require.NoError(t, err)
	assert.Contains(t, out, "SKILL.md  +")
	assert.Contains(t, out, "file(s) changed")
	assert.NotContains(t, out, "@@", "stat mode must not print hunks")
}

func TestRunDiff_NoColor_NoANSI(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	flagDiffNoColor = true
	setFakeResolve(t, "1.1.0", "v1.1.0", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0": {"SKILL.md": "old\n"},
			"1.1.0": {"SKILL.md": "new\n"},
		},
		map[string]string{"1.0.0": "sha256:from", "1.1.0": "sha256:to"},
	)

	out, err := runDiffCapture(t, diffDepName)
	require.NoError(t, err)
	assert.NotContains(t, out, "\x1b[", "output must contain no ANSI escape codes")
}

func TestRunDiff_BinaryFile_Reported(t *testing.T) {
	resetDiffFlags(t)
	diffTestSetup(t, "^1.0.0")
	setFakeResolve(t, "1.1.0", "v1.1.0", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0": {"logo.png": "\x00\x01binary-old"},
			"1.1.0": {"logo.png": "\x00\x01binary-new"},
		},
		map[string]string{"1.0.0": "sha256:from", "1.1.0": "sha256:to"},
	)

	out, err := runDiffCapture(t, diffDepName)
	require.NoError(t, err)
	assert.Contains(t, out, "binary file changed")
	assert.NotContains(t, out, "@@")
}

func TestRunDiff_ManifestAndLockUnchanged(t *testing.T) {
	resetDiffFlags(t)
	dir := diffTestSetup(t, "^1.0.0")
	setFakeResolve(t, "1.1.0", "v1.1.0", nil)
	setFakeFetch(t,
		map[string]map[string]string{
			"1.0.0": {"SKILL.md": "old\n"},
			"1.1.0": {"SKILL.md": "new\n"},
		},
		map[string]string{"1.0.0": "sha256:from", "1.1.0": "sha256:to"},
	)

	manifestBefore, _ := os.ReadFile(filepath.Join(dir, "melon.yaml"))
	lockBefore, _ := os.ReadFile(filepath.Join(dir, "melon.lock"))

	_, err := runDiffCapture(t, diffDepName)
	require.NoError(t, err)

	manifestAfter, _ := os.ReadFile(filepath.Join(dir, "melon.yaml"))
	lockAfter, _ := os.ReadFile(filepath.Join(dir, "melon.lock"))
	assert.Equal(t, manifestBefore, manifestAfter)
	assert.Equal(t, lockBefore, lockAfter)
}

func TestColorizeDiff(t *testing.T) {
	body := "--- a/x\n+++ b/x\n@@ -1 +1 @@\n-old\n+new\n unchanged\n"
	plain := colorizeDiff(body, false)
	assert.Equal(t, body, plain, "color=false must be a no-op")

	colored := colorizeDiff(body, true)
	assert.Contains(t, colored, "\x1b[", "color=true must emit ANSI codes")
	assert.Contains(t, colored, "unchanged")
}

func TestIsBinary(t *testing.T) {
	assert.False(t, isBinary([]byte("plain text\n")))
	assert.False(t, isBinary(nil))
	assert.True(t, isBinary([]byte{0x00, 0x01, 0x02}))
	assert.True(t, isBinary([]byte{0xff, 0xfe, 0xfd}))
}
