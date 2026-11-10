package cli

import (
	"bytes"
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

// useGlobal points HOME at a temp dir, turns on --global and returns home.
func useGlobal(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	origGlobal, origDir := flagGlobal, flagDir
	flagGlobal, flagDir = true, ""
	t.Cleanup(func() { flagGlobal, flagDir = origGlobal, origDir })
	return home
}

func writeGlobalManifest(t *testing.T, home string, m manifest.Manifest) string {
	t.Helper()
	root := filepath.Join(home, ".config", "melon")
	require.NoError(t, os.MkdirAll(root, 0755))
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "melon.yaml"), data, 0644))
	return root
}

// fakeNetwork stubs version resolution, manifest fetching and fetching.
func fakeNetwork(t *testing.T) {
	t.Helper()
	origRV, origFM, origF := resolveVersionFn, fetchManifestFn, fetchFn
	t.Cleanup(func() { resolveVersionFn, fetchManifestFn, fetchFn = origRV, origFM, origF })
	resolveVersionFn = func(string, string) (string, string, error) { return "1.0.0", "v1.0.0", nil }
	fetchManifestFn = func(string, string, string) (manifest.Manifest, error) { return manifest.Manifest{}, nil }
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		if err := os.MkdirAll(installDir, 0755); err != nil {
			return fetcher.FetchResult{}, err
		}
		if err := os.WriteFile(filepath.Join(installDir, "SKILL.md"), []byte("# skill"), 0644); err != nil {
			return fetcher.FetchResult{}, err
		}
		return fetcher.FetchResult{TreeHash: "sha256:abc", Files: []string{"SKILL.md"}}, nil
	}
}

func TestInitGlobal_WritesManifestWithHarnesses(t *testing.T) {
	home := useGlobal(t)
	origYes := flagYes
	t.Cleanup(func() { flagYes = origYes })
	flagYes = false

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	// Not a TTY in tests, so the plain prompt reads the harness selection.
	cmd.SetIn(strings.NewReader("claude-code opencode\n"))
	require.NoError(t, runInit(cmd, nil))

	m, err := manifest.Load(filepath.Join(home, ".config", "melon", "melon.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "global", m.Name)
	assert.Equal(t, []string{"claude-code", "opencode"}, m.HarnessList())
	assert.True(t, m.IsVendored(), "global manifest must not opt into .gitignore management")

	info, err := os.Stat(filepath.Join(home, ".melon", "store"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGlobal_RequiresInit(t *testing.T) {
	useGlobal(t)
	_, err := resolveProjectDir()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "melon init --global")
}

func TestGlobal_RejectsDirFlag(t *testing.T) {
	useGlobal(t)
	flagDir = t.TempDir()
	_, err := resolveRootDir()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--global and --dir")
}

func TestInstallGlobal_PlacesSkipsAndRemoves(t *testing.T) {
	home := useGlobal(t)
	fakeNetwork(t)
	root := writeGlobalManifest(t, home, manifest.Manifest{
		Name: "global", Version: "0.1.0",
		Harnesses: []string{"claude-code"},
		Vendor:    boolPtr(false), // even vendor: false must not write a .gitignore
		Dependencies: map[string]string{
			"github.com/alice/skills/pdf":  "^1.0.0",
			"github.com/alice/skills/xlsx": "^1.0.0",
		},
	})

	// The user already keeps their own xlsx skill in ~/.claude/skills/.
	userSkill := filepath.Join(home, ".claude", "skills", "xlsx")
	require.NoError(t, os.MkdirAll(userSkill, 0755))

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	require.NoError(t, runInstall(cmd, nil))

	// Lock lives next to the global manifest; content in ~/.melon/store/.
	lock, err := lockfile.Load(filepath.Join(root, "melon.lock"))
	require.NoError(t, err)
	assert.Len(t, lock.Dependencies, 2)
	store := filepath.Join(home, ".melon", "store", "github.com-alice-skills-pdf@1.0.0")
	target, err := os.Readlink(filepath.Join(home, ".claude", "skills", "pdf"))
	require.NoError(t, err)
	assert.Equal(t, store, target)

	// Existing unmanaged skill is skipped and reported.
	fi, err := os.Lstat(userSkill)
	require.NoError(t, err)
	assert.True(t, fi.IsDir(), "user's skill must not be replaced")
	assert.Contains(t, out.String(), "Skipped 1 skill already present and not managed by melon")

	_, err = os.Stat(filepath.Join(root, ".gitignore"))
	assert.True(t, os.IsNotExist(err), "global installs never manage .gitignore")

	// Removing both deps unlinks the managed one and leaves the user's alone.
	cmd.SetIn(strings.NewReader("y\n"))
	require.NoError(t, runRemove(cmd, []string{"github.com/alice/skills/pdf"}))
	_, err = os.Lstat(filepath.Join(home, ".claude", "skills", "pdf"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(store)
	assert.True(t, os.IsNotExist(err), "store entry removed")

	require.NoError(t, runRemove(cmd, []string{"github.com/alice/skills/xlsx"}))
	_, err = os.Stat(userSkill)
	assert.NoError(t, err, "user's skill survives remove")
}

func TestInstallGlobal_AllowedSourcesEnforced(t *testing.T) {
	home := useGlobal(t)
	fakeNetwork(t)
	writeGlobalManifest(t, home, manifest.Manifest{
		Name: "global", Version: "0.1.0",
		Policy:       &manifest.PolicyConfig{AllowedSources: []string{"github.com/trusted/*"}},
		Dependencies: map[string]string{"github.com/evil/skills/x": "^1.0.0"},
	})

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	err := runInstall(cmd, nil)
	require.Error(t, err)
	_, statErr := os.Stat(filepath.Join(home, ".melon", "store", "github.com-evil-skills-x@1.0.0"))
	assert.True(t, os.IsNotExist(statErr), "blocked dep must not be fetched")
}

func TestOfferAddMany_AnswerGlobalWritesGlobalManifest(t *testing.T) {
	home := useGlobal(t)
	flagGlobal = false // the prompt, not the flag, chooses global here
	root := writeGlobalManifest(t, home, manifest.Manifest{Name: "global", Version: "0.1.0"})
	project := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(project, "melon.yaml"), []byte("name: p\nversion: 0.0.1\n"), 0644))
	setFlagDir(t, project)

	origRunInstall := runInstallFn
	t.Cleanup(func() { runInstallFn = origRunInstall })
	var installedGlobally bool
	runInstallFn = func(*cobra.Command, []string) error {
		installedGlobally = flagGlobal
		return nil
	}

	// flagDir is set, but answering "y" switches to the global root, which
	// rejects --dir; clear it to model a plain `melon search`.
	flagDir = ""
	out, err := runOfferAddMany(t, []string{"github.com/owner/skill-a@^1.0.0"}, "\ny\n")
	require.NoError(t, err)
	assert.Contains(t, out, "Install globally")
	assert.True(t, installedGlobally)
	assert.False(t, flagGlobal, "flag is restored after the install")

	m, err := manifest.Load(filepath.Join(root, "melon.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "^1.0.0", m.Dependencies["github.com/owner/skill-a"])

	pm, err := manifest.Load(filepath.Join(project, "melon.yaml"))
	require.NoError(t, err)
	assert.Empty(t, pm.Dependencies, "project manifest untouched")
}

func boolPtr(b bool) *bool { return &b }

func TestListGlobalCheck_FlagsUnmanagedSkill(t *testing.T) {
	home := useGlobal(t)
	fakeNetwork(t)
	writeGlobalManifest(t, home, manifest.Manifest{
		Name: "global", Version: "0.1.0", Harnesses: []string{"claude-code"},
		Dependencies: map[string]string{
			"github.com/alice/skills/pdf":  "^1.0.0",
			"github.com/alice/skills/xlsx": "^1.0.0",
		},
	})
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude", "skills", "xlsx"), 0755))

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	require.NoError(t, runInstall(cmd, nil))

	origCheck := flagCheck
	t.Cleanup(func() { flagCheck = origCheck })
	flagCheck = true
	out.Reset()
	err := runList(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, out.String(), "OK       github.com/alice/skills/pdf")
	assert.Contains(t, out.String(), "SKIPPED  github.com/alice/skills/xlsx")
	assert.Contains(t, out.String(), "~/.claude/skills/xlsx")
}
