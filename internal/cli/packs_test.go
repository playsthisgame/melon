package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/playsthisgame/melon/internal/fetcher"
	gh "github.com/playsthisgame/melon/internal/github"
	"github.com/playsthisgame/melon/internal/index"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// packWorld fakes the network: remote manifests keyed by dep name, and the set
// of deps whose fetched directory has no SKILL.md (pure packs).
type packWorld struct {
	manifests map[string]manifest.Manifest
	pure      map[string]bool

	mu      sync.Mutex // fetchDeps fetches in parallel
	fetched []string
}

const (
	springPack = "github.com/community/packs/spring-expert-pack"
	springBoot = "github.com/acme/skills/spring-boot"
	springData = "github.com/acme/skills/spring-data"
	junitSkill = "github.com/acme/skills/junit"
	javaPack   = "github.com/community/packs/java-testing-pack"
)

func depNameFrom(repoURL, subdir string) string {
	name := strings.TrimPrefix(repoURL, "https://")
	if subdir != "" {
		name += "/" + subdir
	}
	return name
}

// install wires the fakes into the package-level injectables and points the
// CLI at dir. Everything is restored on test cleanup.
func (w *packWorld) install(t *testing.T, dir string) {
	t.Helper()
	origResolve, origManifest, origFetch, origDir := resolveVersionFn, fetchManifestFn, fetchFn, flagDir
	origNoPlace, origFrozen, origYes := flagNoPlace, flagFrozen, flagAddYes
	t.Cleanup(func() {
		resolveVersionFn, fetchManifestFn, fetchFn, flagDir = origResolve, origManifest, origFetch, origDir
		flagNoPlace, flagFrozen, flagAddYes = origNoPlace, origFrozen, origYes
	})
	flagDir = dir
	flagNoPlace, flagFrozen = false, false

	resolveVersionFn = func(repoURL, constraint string) (string, string, error) { return "1.0.0", "v1.0.0", nil }
	fetchManifestFn = func(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
		return w.manifests[depNameFrom(repoURL, subdir)], nil
	}
	fetchFn = func(dep resolver.ResolvedDep, installDir string) (fetcher.FetchResult, error) {
		w.mu.Lock()
		w.fetched = append(w.fetched, dep.Name)
		w.mu.Unlock()
		if err := os.MkdirAll(installDir, 0755); err != nil {
			return fetcher.FetchResult{}, err
		}
		files := []string{}
		if mm, ok := w.manifests[dep.Name]; ok {
			data, _ := yaml.Marshal(mm)
			_ = os.WriteFile(filepath.Join(installDir, "melon.yaml"), data, 0644)
			files = append(files, "melon.yaml")
		}
		if !w.pure[dep.Name] {
			_ = os.WriteFile(filepath.Join(installDir, "SKILL.md"), []byte("# "+dep.Name), 0644)
			files = append(files, "SKILL.md")
		}
		return fetcher.FetchResult{TreeHash: "sha256:" + dep.Name, Files: files}, nil
	}
}

func springWorld() *packWorld {
	return &packWorld{
		manifests: map[string]manifest.Manifest{
			springPack: {Name: "spring-expert-pack", Type: manifest.TypePack, Dependencies: map[string]string{
				springBoot: "^1.0.0",
				springData: "^1.0.0",
				junitSkill: "^1.0.0",
			}},
			javaPack: {Name: "java-testing-pack", Type: manifest.TypePack, Dependencies: map[string]string{
				junitSkill: "^1.0.0",
			}},
		},
		pure: map[string]bool{springPack: true, javaPack: true},
	}
}

func writeProject(t *testing.T, dir string, m manifest.Manifest) {
	t.Helper()
	if m.Name == "" {
		m.Name, m.Version = "test-project", "0.1.0"
	}
	if len(m.Harnesses) == 0 {
		m.Harnesses = []string{"claude-code"}
	}
	data, err := yaml.Marshal(m)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "melon.yaml"), data, 0644))
}

func loadLock(t *testing.T, dir string) map[string]lockfile.LockedDep {
	t.Helper()
	lf, err := lockfile.Load(filepath.Join(dir, "melon.lock"))
	require.NoError(t, err)
	out := map[string]lockfile.LockedDep{}
	for _, d := range lf.Dependencies {
		out[d.Name] = d
	}
	return out
}

func linkExists(dir, skill string) bool {
	_, err := os.Lstat(filepath.Join(dir, ".claude/skills", skill))
	return err == nil
}

func newTestCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(""))
	return cmd, &buf
}

// --- install ---

func TestInstall_PackMembersInstalledAndPlaced(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	writeProject(t, dir, manifest.Manifest{Dependencies: map[string]string{springPack: "^1.0.0"}})

	cmd, _ := newTestCmd()
	require.NoError(t, runInstall(cmd, nil))

	lock := loadLock(t, dir)
	require.Len(t, lock, 4)
	assert.Equal(t, "pack", lock[springPack].Type)
	assert.True(t, lock[springPack].Direct)
	assert.Equal(t, "skill", lock[springBoot].Type)
	assert.False(t, lock[springBoot].Direct)
	assert.Equal(t, []string{springPack}, lock[springBoot].Via)

	assert.False(t, linkExists(dir, "spring-expert-pack"), "pure pack must not be placed")
	assert.True(t, linkExists(dir, "spring-boot"))
	assert.True(t, linkExists(dir, "spring-data"))
	assert.True(t, linkExists(dir, "junit"))
}

func TestInstall_HybridPackPlaced(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	delete(w.pure, springPack)
	w.install(t, dir)
	writeProject(t, dir, manifest.Manifest{Dependencies: map[string]string{springPack: "^1.0.0"}})

	cmd, _ := newTestCmd()
	require.NoError(t, runInstall(cmd, nil))
	assert.True(t, linkExists(dir, "spring-expert-pack"))
	assert.True(t, linkExists(dir, "spring-boot"))
}

func TestInstall_PolicyBlocksPackMemberBeforeFetch(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	writeProject(t, dir, manifest.Manifest{
		Dependencies: map[string]string{springPack: "^1.0.0"},
		Policy:       &manifest.PolicyConfig{AllowedSources: []string{"github.com/community/*"}},
	})
	writeLockfile(t, dir, nil)
	before, _ := os.ReadFile(filepath.Join(dir, "melon.lock"))

	cmd, _ := newTestCmd()
	err := runInstall(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source policy violation")
	assert.Contains(t, err.Error(), springBoot+" (via "+springPack+")")
	assert.Empty(t, w.fetched, "no skill may be fetched when a member is blocked")

	after, _ := os.ReadFile(filepath.Join(dir, "melon.lock"))
	assert.Equal(t, before, after, "lock must be unchanged")
}

func TestInstall_WarnsOnExactPinInPack(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	pm := w.manifests[springPack]
	pm.Dependencies[springBoot] = "1.4.0"
	w.install(t, dir)
	writeProject(t, dir, manifest.Manifest{Dependencies: map[string]string{springPack: "^1.0.0"}})

	cmd, out := newTestCmd()
	require.NoError(t, runInstall(cmd, nil))
	assert.Contains(t, out.String(), "warning: pack "+springPack+" pins "+springBoot+" to exact version 1.4.0")
	assert.NotContains(t, out.String(), "pins "+springData)
}

// --- frozen ---

func TestInstallFrozen_FailsWhenSkillBecomesDirect(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	writeProject(t, dir, manifest.Manifest{Dependencies: map[string]string{springPack: "^1.0.0"}})
	cmd, _ := newTestCmd()
	require.NoError(t, runInstall(cmd, nil))

	writeProject(t, dir, manifest.Manifest{Dependencies: map[string]string{springPack: "^1.0.0", springBoot: "^1.0.0"}})
	flagFrozen = true
	err := runInstall(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "out of date")
}

func TestInstallFrozen_LegacyLockPasses(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	writeProject(t, dir, manifest.Manifest{Dependencies: map[string]string{springBoot: "^1.0.0"}})
	writeLockfile(t, dir, []lockfile.LockedDep{{
		Name: springBoot, Version: "1.0.0", GitTag: "v1.0.0",
		RepoURL: "https://github.com/acme/skills", Subdir: "skills/spring-boot", Entrypoint: "SKILL.md",
	}})

	flagFrozen = true
	cmd, _ := newTestCmd()
	require.NoError(t, runInstall(cmd, nil))
}

// --- add ---

func setAddTTY(t *testing.T, tty bool) {
	t.Helper()
	orig := addIsTTY
	addIsTTY = func() bool { return tty }
	t.Cleanup(func() { addIsTTY = orig })
}

func TestAdd_PackInstallsMembers(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	setAddTTY(t, false)
	writeProject(t, dir, manifest.Manifest{})

	cmd, out := newTestCmd()
	require.NoError(t, runAdd(cmd, []string{springPack + "@^1.0.0"}))

	assert.Contains(t, out.String(), springPack+" [pack · 3 skills]")
	assert.Contains(t, out.String(), "  - "+springBoot+" ^1.0.0")
	assert.NotContains(t, out.String(), "[Y/n]", "non-TTY must not prompt")

	m, err := manifest.Load(filepath.Join(dir, "melon.yaml"))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{springPack: "^1.0.0"}, m.Dependencies)

	lock := loadLock(t, dir)
	assert.Len(t, lock, 4)
	assert.True(t, linkExists(dir, "spring-boot"))
	assert.False(t, linkExists(dir, "spring-expert-pack"))
}

func TestAdd_PackDeclinedLeavesManifest(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	setAddTTY(t, true)
	writeProject(t, dir, manifest.Manifest{})
	before, _ := os.ReadFile(filepath.Join(dir, "melon.yaml"))

	cmd, out := newTestCmd()
	cmd.SetIn(strings.NewReader("n\n"))
	require.NoError(t, runAdd(cmd, []string{springPack + "@^1.0.0"}))

	assert.Contains(t, out.String(), "Install 1 pack (3 skills)? [Y/n]")
	after, _ := os.ReadFile(filepath.Join(dir, "melon.yaml"))
	assert.Equal(t, before, after)
	assert.Empty(t, w.fetched)
}

func TestAdd_PackYesSkipsPrompt(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	setAddTTY(t, true)
	flagAddYes = true
	writeProject(t, dir, manifest.Manifest{})

	cmd, out := newTestCmd()
	require.NoError(t, runAdd(cmd, []string{springPack + "@^1.0.0"}))
	assert.NotContains(t, out.String(), "[Y/n]")
	assert.Len(t, loadLock(t, dir), 4)
}

func TestAdd_PackMemberBlockedByPolicy(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	setAddTTY(t, false)
	writeProject(t, dir, manifest.Manifest{
		Policy: &manifest.PolicyConfig{AllowedSources: []string{"github.com/community/*"}},
	})
	before, _ := os.ReadFile(filepath.Join(dir, "melon.yaml"))

	cmd, _ := newTestCmd()
	err := runAdd(cmd, []string{springPack + "@^1.0.0"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), springBoot)

	after, _ := os.ReadFile(filepath.Join(dir, "melon.yaml"))
	assert.Equal(t, before, after)
	assert.Empty(t, w.fetched)
}

func TestAdd_SkillWithTransitiveDepInstallsIt(t *testing.T) {
	dir := t.TempDir()
	w := &packWorld{manifests: map[string]manifest.Manifest{
		"github.com/alice/pdf-skill": {Name: "pdf-skill", Dependencies: map[string]string{"github.com/bob/base-utils": "^2.0.0"}},
	}}
	w.install(t, dir)
	setAddTTY(t, false)
	writeProject(t, dir, manifest.Manifest{})

	cmd, out := newTestCmd()
	require.NoError(t, runAdd(cmd, []string{"github.com/alice/pdf-skill@^1.0.0"}))
	assert.NotContains(t, out.String(), "is a pack")

	lock := loadLock(t, dir)
	require.Contains(t, lock, "github.com/bob/base-utils")
	assert.Equal(t, []string{"github.com/alice/pdf-skill"}, lock["github.com/bob/base-utils"].Via)
	assert.True(t, lock["github.com/alice/pdf-skill"].Direct)
}

func TestAdd_PlainSkillUsesFastPath(t *testing.T) {
	dir := t.TempDir()
	w := &packWorld{manifests: map[string]manifest.Manifest{}}
	w.install(t, dir)
	setAddTTY(t, false)
	writeProject(t, dir, manifest.Manifest{})

	cmd, _ := newTestCmd()
	require.NoError(t, runAdd(cmd, []string{"github.com/alice/pdf-skill@^1.0.0"}))
	lock := loadLock(t, dir)
	require.Len(t, lock, 1)
	assert.Equal(t, "skill", lock["github.com/alice/pdf-skill"].Type)
	assert.True(t, lock["github.com/alice/pdf-skill"].Direct)
}

// --- remove ---

func installProject(t *testing.T, dir string, w *packWorld, deps map[string]string) {
	t.Helper()
	w.install(t, dir)
	writeProject(t, dir, manifest.Manifest{Dependencies: deps})
	cmd, _ := newTestCmd()
	require.NoError(t, runInstall(cmd, nil))
}

func TestRemove_PackPrunesMembers(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0", "github.com/alice/pdf-skill": "^1.0.0"})
	require.True(t, linkExists(dir, "spring-boot"))

	cmd, _ := newTestCmd()
	require.NoError(t, runRemove(cmd, []string{springPack}))

	lock := loadLock(t, dir)
	assert.Equal(t, []string{"github.com/alice/pdf-skill"}, sortedKeys(lock))
	for _, s := range []string{"spring-boot", "spring-data", "junit"} {
		assert.False(t, linkExists(dir, s), s+" must be unlinked")
	}
	entries, _ := os.ReadDir(filepath.Join(dir, ".melon"))
	for _, e := range entries {
		assert.NotContains(t, e.Name(), "spring", "pack and member caches must be removed")
	}
}

func TestRemove_SharedMemberKept(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0", javaPack: "^1.0.0"})

	cmd, _ := newTestCmd()
	require.NoError(t, runRemove(cmd, []string{springPack}))

	lock := loadLock(t, dir)
	require.Contains(t, lock, junitSkill)
	assert.Equal(t, []string{javaPack}, lock[junitSkill].Via)
	assert.NotContains(t, lock, springBoot)
	assert.True(t, linkExists(dir, "junit"))
}

func TestRemove_DirectMemberKept(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0", springBoot: "^1.0.0"})

	cmd, _ := newTestCmd()
	require.NoError(t, runRemove(cmd, []string{springPack}))

	lock := loadLock(t, dir)
	require.Contains(t, lock, springBoot)
	assert.True(t, lock[springBoot].Direct)
	assert.Nil(t, lock[springBoot].Via)
	assert.True(t, linkExists(dir, "spring-boot"))
}

func TestRemove_TransitiveMemberRejected(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0"})
	manBefore, _ := os.ReadFile(filepath.Join(dir, "melon.yaml"))
	lockBefore, _ := os.ReadFile(filepath.Join(dir, "melon.lock"))

	cmd, _ := newTestCmd()
	err := runRemove(cmd, []string{springBoot})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required by "+springPack)

	manAfter, _ := os.ReadFile(filepath.Join(dir, "melon.yaml"))
	lockAfter, _ := os.ReadFile(filepath.Join(dir, "melon.lock"))
	assert.Equal(t, manBefore, manAfter)
	assert.Equal(t, lockBefore, lockAfter)
}

// --- list ---

func TestList_TreeShowsPackMembers(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0", "github.com/alice/pdf-skill": "^1.0.0"})

	out, err := runListCmd(t, dir, false, false)
	require.NoError(t, err)
	want := "" +
		"  github.com/alice/pdf-skill  1.0.0\n" +
		"  " + springPack + "  1.0.0  (pack)\n" +
		"    " + junitSkill + "  1.0.0\n" +
		"    " + springBoot + "  1.0.0\n" +
		"    " + springData + "  1.0.0\n"
	assert.Equal(t, want, out)
}

func TestList_DirectMemberAppearsTwice(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0", springBoot: "^1.0.0"})

	out, err := runListCmd(t, dir, false, false)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(out, "  "+springBoot+"  1.0.0\n"), "direct dep at top level")
	assert.Contains(t, out, "    "+springBoot+"  1.0.0\n", "also nested under pack")
}

func TestList_NoPacksUnchanged(t *testing.T) {
	dir := t.TempDir()
	writeLockfile(t, dir, []lockfile.LockedDep{
		{Name: "github.com/alice/xlsx-skill", Version: "2.0.0"},
		{Name: "github.com/alice/pdf-skill", Version: "1.2.0"},
	})
	out, err := runListCmd(t, dir, false, false)
	require.NoError(t, err)
	assert.Equal(t, "  github.com/alice/pdf-skill  1.2.0\n  github.com/alice/xlsx-skill  2.0.0\n", out)
}

func TestList_JSONProvenance(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0"})

	stdout, _, err := runListCmdJSON(t, dir, false, false)
	require.NoError(t, err)
	var parsed struct {
		Installed []listInstalledEntry `json:"installed"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(stdout), &parsed)) // JSON is valid YAML
	byName := map[string]listInstalledEntry{}
	for _, e := range parsed.Installed {
		byName[e.Name] = e
	}
	assert.Equal(t, "skill", byName[springBoot].Type)
	assert.False(t, byName[springBoot].Direct)
	assert.Equal(t, []string{springPack}, byName[springBoot].Via)
	assert.Equal(t, "pack", byName[springPack].Type)
	assert.True(t, byName[springPack].Direct)
	assert.Contains(t, stdout, `"via": []`, "empty via must serialise as an empty array")
}

func TestListCheck_IgnoresPurePacks(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0"})

	out, err := runListCmd(t, dir, false, true)
	require.NoError(t, err, out)
	assert.NotContains(t, out, "MISSING")
	assert.Contains(t, out, "PACK")

	stdout, _, err := runListCmdJSON(t, dir, false, true)
	require.NoError(t, err)
	assert.NotContains(t, stdout, `"missing"`)
}

// --- index / search / info ---

func TestIndexEntry_PackType(t *testing.T) {
	var parsed struct {
		Skills []index.Entry `yaml:"skills"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(`skills:
  - name: github.com/c/spring-expert-pack
    type: pack
  - name: github.com/a/skill
  - name: github.com/a/odd
    type: bundle
`), &parsed))
	assert.True(t, parsed.Skills[0].IsPack())
	assert.False(t, parsed.Skills[1].IsPack())
	assert.False(t, parsed.Skills[2].IsPack(), "unknown index types are treated as skills")
}

func TestSearch_PlainTextLabelsPacks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `skills:
  - name: github.com/community/packs/spring-expert-pack
    description: Everything for Spring
    type: pack
  - name: github.com/acme/skills/spring-boot
    description: Spring Boot skill
`)
	}))
	defer srv.Close()

	dir := t.TempDir()
	f := false
	require.NoError(t, manifest.Save(manifest.Manifest{Name: "t", Version: "0.1.0",
		Index: &manifest.IndexConfig{URLs: []string{srv.URL}, PublicIndex: &f}}, filepath.Join(dir, "melon.yaml")))
	setFlagDir(t, dir)

	cmd, out := newTestCmd()
	require.NoError(t, runSearch(cmd, []string{"spring"}))
	assert.Contains(t, out.String(), springPack+" [pack]\t")
	assert.NotContains(t, out.String(), springBoot+" [pack]")
}

func runInfoForPack(t *testing.T, jsonOut bool) string {
	t.Helper()
	const owner, repo = "community", "packs"
	srv := fakeGitHubAndIndexServer(t, "skills: []\n", owner, repo, "Community packs", false)
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	f := false
	require.NoError(t, manifest.Save(manifest.Manifest{Name: "t", Version: "0.1.0",
		Index: &manifest.IndexConfig{URLs: []string{srv.URL + "/index.yaml"}, PublicIndex: &f}}, filepath.Join(dir, "melon.yaml")))
	setFlagDir(t, dir)

	origGH := newGHClientFn
	newGHClientFn = func() *gh.Client { return gh.NewWithBase(srv.URL) }
	t.Cleanup(func() { newGHClientFn = origGH })

	var gotRef, gotSubdir string
	fetchManifestFn = func(repoURL, ref, subdir string) (manifest.Manifest, error) {
		gotRef, gotSubdir = ref, subdir
		return springWorld().manifests[springPack], nil
	}

	origJSON := flagInfoJSON
	flagInfoJSON = jsonOut
	t.Cleanup(func() { flagInfoJSON = origJSON })

	cmd, out := newTestCmd()
	require.NoError(t, runInfo(cmd, []string{springPack}))
	assert.Equal(t, "v1.0.0", gotRef, "manifest read at the latest tag")
	assert.Equal(t, "spring-expert-pack", gotSubdir)
	return out.String()
}

func TestInfo_PackListsMembers(t *testing.T) {
	out := runInfoForPack(t, false)
	assert.Contains(t, out, "Type         pack")
	assert.Contains(t, out, "Members (3):")
	assert.Contains(t, out, "    "+junitSkill+"  ^1.0.0\n    "+springBoot+"  ^1.0.0\n    "+springData+"  ^1.0.0\n")
}

func TestInfo_PackJSON(t *testing.T) {
	var parsed infoJSONOutput
	require.NoError(t, json.Unmarshal([]byte(runInfoForPack(t, true)), &parsed))
	assert.Equal(t, "pack", parsed.Type)
	require.Len(t, parsed.Members, 3)
	assert.Equal(t, infoMember{Name: junitSkill, Constraint: "^1.0.0"}, parsed.Members[0])
}

// --- search confirmation prompt ---

func stubLatestTag(t *testing.T) {
	t.Helper()
	orig := latestTagFn
	latestTagFn = func(string) (string, string, error) { return "1.0.0", "v1.0.0", nil }
	t.Cleanup(func() { latestTagFn = orig })
}

// offerWithPacks runs offerAddMany, answering "n" so nothing is installed.
func offerWithPacks(t *testing.T, w *packWorld, packs map[string]bool, paths ...string) string {
	t.Helper()
	dir := t.TempDir()
	w.install(t, dir)
	stubLatestTag(t)
	writeProject(t, dir, manifest.Manifest{})
	cmd, out := newTestCmd()
	cmd.SetIn(strings.NewReader("n\n"))
	require.NoError(t, offerAddMany(cmd, paths, packs))
	return out.String()
}

func TestOfferAddMany_CountsPackSkills(t *testing.T) {
	out := offerWithPacks(t, springWorld(), map[string]bool{springPack: true}, springPack)
	assert.Contains(t, out, "  - "+springPack+" [pack · 3 skills]")
	assert.Contains(t, out, "Install 1 pack (3 skills)? [Y/n]")
	assert.NotContains(t, out, "1 skill(s)")
}

func TestOfferAddMany_MixedPacksAndSkillsDedupesSharedMembers(t *testing.T) {
	packs := map[string]bool{springPack: true, javaPack: true}
	out := offerWithPacks(t, springWorld(), packs, springPack, javaPack, "github.com/alice/pdf-skill@^1.0.0")
	assert.Contains(t, out, "  - "+javaPack+" [pack · 1 skill]")
	// junit is in both packs: 3 + 1 members, 3 distinct.
	assert.Contains(t, out, "Install 2 packs (3 skills) and 1 skill? [Y/n]")
	assert.Contains(t, out, "  - github.com/alice/pdf-skill@^1.0.0\n")
}

func TestOfferAddMany_NestedPackCountsLeafSkills(t *testing.T) {
	w := springWorld()
	outer := "github.com/community/packs/outer-pack"
	w.manifests[outer] = manifest.Manifest{Type: manifest.TypePack, Dependencies: map[string]string{
		springPack: "^1.0.0", "github.com/x/extra": "^1.0.0"}}
	out := offerWithPacks(t, w, map[string]bool{outer: true, springPack: true}, outer)
	assert.Contains(t, out, "Install 1 pack (4 skills)? [Y/n]")
}

func TestOfferAddMany_UnknownPackContentsStillPrompts(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	w.install(t, dir)
	stubLatestTag(t)
	fetchManifestFn = func(string, string, string) (manifest.Manifest, error) {
		return manifest.Manifest{}, fmt.Errorf("network down")
	}
	writeProject(t, dir, manifest.Manifest{})
	cmd, out := newTestCmd()
	cmd.SetIn(strings.NewReader("n\n"))
	require.NoError(t, offerAddMany(cmd, []string{springPack}, map[string]bool{springPack: true}))
	assert.Contains(t, out.String(), "  - "+springPack+" [pack]\n")
	assert.Contains(t, out.String(), "Install 1 pack? [Y/n]")
}

func TestSearchRender_PackLabel(t *testing.T) {
	items := []list.Item{
		searchResultItem{path: springPack, pack: true},
		searchResultItem{path: springBoot},
	}
	l := list.New(items, searchMultiSelectDelegate{selected: map[int]bool{}}, 80, 10)
	var buf bytes.Buffer
	d := searchMultiSelectDelegate{selected: map[int]bool{}}
	d.Render(&buf, l, 0, items[0])
	assert.Contains(t, buf.String(), springPack)
	assert.Contains(t, buf.String(), "[pack]")
	buf.Reset()
	d.Render(&buf, l, 1, items[1])
	assert.NotContains(t, buf.String(), "[pack]")
}

// --- pack-aware counts elsewhere ---

func TestAdd_NestedPackCountsAllSkills(t *testing.T) {
	dir := t.TempDir()
	w := springWorld()
	outer := "github.com/community/packs/outer-pack"
	w.manifests[outer] = manifest.Manifest{Type: manifest.TypePack, Dependencies: map[string]string{
		springPack: "^1.0.0", "github.com/x/extra": "^1.0.0"}}
	w.pure[outer] = true
	w.install(t, dir)
	setAddTTY(t, true)
	writeProject(t, dir, manifest.Manifest{})

	cmd, out := newTestCmd()
	cmd.SetIn(strings.NewReader("n\n"))
	require.NoError(t, runAdd(cmd, []string{outer + "@^1.0.0"}))
	assert.Contains(t, out.String(), outer+" [pack · 4 skills]")
	assert.Contains(t, out.String(), "  - "+springPack+" ^1.0.0 [pack · 3 skills]")
	assert.Contains(t, out.String(), "Install 1 pack (4 skills)? [Y/n]")
}

func runOfferRemove(t *testing.T, dir string, answer string, names ...string) string {
	t.Helper()
	cmd, out := newTestCmd()
	cmd.SetIn(strings.NewReader(answer))
	require.NoError(t, offerRemoveMany(cmd, names))
	return out.String()
}

func TestOfferRemoveMany_CountsPackSkills(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0", "github.com/alice/pdf-skill": "^1.0.0"})

	out := runOfferRemove(t, dir, "n\n", springPack, "github.com/alice/pdf-skill")
	assert.Contains(t, out, "  - "+springPack+" [pack · 3 skills]")
	assert.Contains(t, out, "Remove 1 pack (3 skills) and 1 skill? [y/N]")
	assert.NotContains(t, out, "skill(s)")
	assert.Len(t, loadLock(t, dir), 5, "declining must not change anything")
}

func TestOfferRemoveMany_ReportsKeptSharedMembers(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0", javaPack: "^1.0.0"})

	out := runOfferRemove(t, dir, "n\n", springPack)
	// junit is still needed by java-testing-pack, so only 2 of the 3 go.
	assert.Contains(t, out, "(1 skill from these packs stays installed because other dependencies still need them)")
	assert.Contains(t, out, "Remove 1 pack (2 skills)? [y/N]")
}

func TestOfferRemoveMany_PrunesPackMembers(t *testing.T) {
	dir := t.TempDir()
	installProject(t, dir, springWorld(), map[string]string{springPack: "^1.0.0", javaPack: "^1.0.0"})

	runOfferRemove(t, dir, "y\n", springPack)
	lock := loadLock(t, dir)
	assert.NotContains(t, lock, springPack)
	assert.NotContains(t, lock, springBoot, "pack members must be pruned, not left behind")
	assert.NotContains(t, lock, springData)
	assert.Contains(t, lock, junitSkill, "shared member kept")
	assert.Equal(t, []string{javaPack}, lock[junitSkill].Via)
	assert.False(t, linkExists(dir, "spring-boot"))
}

func TestInstalledSummary(t *testing.T) {
	assert.Equal(t, "3 skills", installedSummary([]lockfile.LockedDep{{Type: "skill"}, {Type: "skill"}, {}}))
	assert.Equal(t, "1 skill and 1 pack", installedSummary([]lockfile.LockedDep{{Type: "skill"}, {Type: "pack"}}))
}

func TestUnreachableAfter_LegacyEntriesKept(t *testing.T) {
	lf := lockfile.LockFile{Dependencies: []lockfile.LockedDep{
		{Name: "legacy-transitive"}, // pre-provenance entry: parents unknown
		{Name: "pack", Type: "pack", Direct: true},
		{Name: "member", Type: "skill", Via: []string{"pack"}},
	}}
	got := unreachableAfter(lf, map[string]string{}, map[string]bool{"pack": true})
	assert.Equal(t, map[string]bool{"pack": true, "member": true}, got)
}

func TestRemovePicker_LabelsPacks(t *testing.T) {
	items := []list.Item{removeSkillItem{name: springPack, version: "^1.0.0", pack: true, skills: 3}}
	l := list.New(items, removeMultiSelectDelegate{selected: map[int]bool{}}, 80, 5)
	var buf bytes.Buffer
	removeMultiSelectDelegate{selected: map[int]bool{}}.Render(&buf, l, 0, items[0])
	assert.Contains(t, buf.String(), springPack+" [pack · 3 skills]  ^1.0.0")
}

func TestUpdatePicker_LabelsPacks(t *testing.T) {
	items := []list.Item{updateSkillItem{name: springPack, constraint: "^1.0.0", locked: "1.0.0", pack: true}}
	l := list.New(items, updateMultiSelectDelegate{selected: map[int]bool{}}, 80, 5)
	var buf bytes.Buffer
	updateMultiSelectDelegate{selected: map[int]bool{}}.Render(&buf, l, 0, items[0])
	assert.Contains(t, buf.String(), springPack+" [pack]  ^1.0.0  (locked: 1.0.0)")
}
