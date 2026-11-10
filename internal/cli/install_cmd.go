package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/playsthisgame/melon/internal/agents"
	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/gitignore"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/placer"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/spf13/cobra"
)

var (
	addStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	updateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	removeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
)

// resolveVersionFn and fetchManifestFn are the functions used by runInstall to
// resolve versions and fetch dependency manifests. They are package-level
// variables so tests can inject fakes without network access.
var (
	resolveVersionFn = fetcher.LatestMatchingVersion
	fetchManifestFn  = resolver.DefaultFetchManifest
	fetchFn          = fetcher.Fetch
)

// runInstall is the cobra handler for `melon install`. It enforces the source
// policy across all deps before delegating to runInstallCore.
func runInstall(cmd *cobra.Command, args []string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	m, err := manifest.Load(manifest.FindPath(dir))
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if len(m.Dependencies) > 0 {
		depPaths := make([]string, 0, len(m.Dependencies))
		for name := range m.Dependencies {
			depPaths = append(depPaths, name)
		}
		if err := checkSourcePolicy(m, depPaths); err != nil {
			return fmt.Errorf("install: %w", err)
		}
	}
	return runInstallCore(cmd, args)
}

// runInstallCore performs the full install pipeline without a policy check.
// Called by runInstall (after its policy check) and by melon add / melon search
// (which have already validated the specific dep being added).
func runInstallCore(cmd *cobra.Command, args []string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}

	manifestPath := manifest.FindPath(dir)
	lockPath := filepath.Join(dir, "melon.lock")

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

	if m.UsesLegacyToolCompat() {
		fmt.Fprintln(cmd.OutOrStdout(),
			"warning: melon.yaml uses tool_compat, renamed to harnesses in v0.5.0.\n"+
				"         tool_compat still works; rename the key to silence this.")
	}

	// Load existing lock file for diff display (ignore error if absent).
	var oldLock lockfile.LockFile
	oldLock, _ = lockfile.Load(lockPath)

	if len(m.Dependencies) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No dependencies declared in melon.yaml.")
		// Prune any deps that were in the previous lock.
		if len(oldLock.Dependencies) > 0 {
			if !flagNoPlace {
				if err := placer.Unplace(oldLock.Dependencies, m, dir, cmd.OutOrStdout()); err != nil {
					return fmt.Errorf("install: %w", err)
				}
			}
			for _, dep := range oldLock.Dependencies {
				if err := store.Remove(dir, resolver.ResolvedDep{Name: dep.Name, Version: dep.Version}); err != nil {
					return fmt.Errorf("install: remove store entry %s: %w", dep.Name, err)
				}
			}
			emptyLock := lockfile.LockFile{GeneratedAt: time.Now().UTC().Format(time.RFC3339)}
			if err := lockfile.Save(emptyLock, lockPath); err != nil {
				return fmt.Errorf("install: write melon.lock: %w", err)
			}
		}
		return nil
	}

	// Ensure .melon/ exists.
	if err := os.MkdirAll(filepath.Join(dir, store.StoreDir), 0755); err != nil {
		return fmt.Errorf("install: create store: %w", err)
	}

	// Resolve the full transitive dependency graph.
	fmt.Fprintln(cmd.OutOrStdout(), "Resolving dependencies...")
	packManifests := make(map[string]manifest.Manifest)
	resolved, err := resolver.Resolve(m, resolveVersionFn, recordPackManifests(fetchManifestFn, packManifests))
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

	// Enforce allowed_sources on the whole graph (pack members included)
	// before any skill content is fetched.
	if err := checkResolvedPolicy(m, resolved); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	warnExactPackPins(cmd.OutOrStdout(), packManifests)

	// Fetch each resolved dep into the store.
	var locked []lockfile.LockedDep
	if isTTY() {
		model := newInstallProgressModel(len(resolved))
		p := tea.NewProgram(model)
		var fetchErr error
		done := make(chan struct{})
		go func() {
			defer close(done)
			locked, fetchErr = fetchDeps(resolved, dir, func(i int, name string, e error) {
				p.Send(depFetchedMsg{index: i, name: name, total: len(resolved), err: e})
			})
			if fetchErr == nil {
				p.Send(fetchDoneMsg{count: len(locked)})
			} else {
				p.Send(depFetchedMsg{err: fetchErr})
			}
		}()
		if _, runErr := p.Run(); runErr != nil {
			<-done
			return fmt.Errorf("install: %w", runErr)
		}
		<-done // ensure goroutine has written locked and fetchErr before we read them
		if fetchErr != nil {
			return fetchErr
		}
		fmt.Fprintf(os.Stdout, "✓ %s installed\n", installedSummary(locked))
	} else {
		locked, err = fetchDeps(resolved, dir, func(i int, name string, e error) {
			if e == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  fetching %s...\n", name)
			}
		})
		if err != nil {
			return err
		}
	}

	newLock := lockfile.LockFile{
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Dependencies: locked,
	}

	// --frozen: verify melon.lock would not change; exit non-zero if it would.
	if flagFrozen {
		diff := lockfile.Diff(oldLock, newLock)
		if len(diff.Added)+len(diff.Removed)+len(diff.Updated) > 0 {
			printLockDiff(cmd, diff)
			return fmt.Errorf("install --frozen: melon.lock is out of date")
		}
		if lockfile.ProvenanceChanged(oldLock, newLock) {
			return fmt.Errorf("install --frozen: melon.lock is out of date (dependency provenance changed)")
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Lock file is up to date.")
		return nil
	}

	if err := lockfile.Save(newLock, lockPath); err != nil {
		return fmt.Errorf("install: write melon.lock: %w", err)
	}

	diff := lockfile.Diff(oldLock, newLock)

	// Print diff vs the previous lock.
	printLockDiff(cmd, diff)

	// Prune removed deps: unplace symlinks then delete cache entries.
	if len(diff.Removed) > 0 {
		if !flagNoPlace {
			if err := placer.Unplace(diff.Removed, m, dir, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("install: %w", err)
			}
		}
		for _, dep := range diff.Removed {
			if err := store.Remove(dir, resolver.ResolvedDep{Name: dep.Name, Version: dep.Version}); err != nil {
				return fmt.Errorf("install: remove store entry %s: %w", dep.Name, err)
			}
		}
	}

	// Delete stale store entries for updated deps (old version dirs are no longer needed).
	if len(diff.Updated) > 0 {
		oldByName := make(map[string]string, len(oldLock.Dependencies))
		for _, dep := range oldLock.Dependencies {
			oldByName[dep.Name] = dep.Version
		}
		for _, dep := range diff.Updated {
			if oldVersion, ok := oldByName[dep.Name]; ok {
				if err := store.Remove(dir, resolver.ResolvedDep{Name: dep.Name, Version: oldVersion}); err != nil {
					return fmt.Errorf("install: remove old store entry %s@%s: %w", dep.Name, oldVersion, err)
				}
			}
		}
	}

	// Place skills into agent directories unless --no-place is set.
	if !flagNoPlace {
		if err := placer.Place(resolved, m, dir, cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("install: %w", err)
		}
	}

	// Sync .gitignore when vendor: false.
	if !m.IsVendored() {
		gitignorePath := filepath.Join(dir, ".gitignore")
		entries := managedGitignoreEntries(resolved, m, dir)
		added, err := gitignore.EnsureEntries(gitignorePath, entries)
		if err != nil {
			return fmt.Errorf("install: update .gitignore: %w", err)
		}
		if len(added) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  updated .gitignore with melon-managed paths")
			fmt.Fprintln(cmd.OutOrStdout(), "  Tip: run `git rm --cached <path>` to stop tracking previously committed files")
		}
	}

	return nil
}

// gitignoreSymlinkEntries returns the .gitignore paths for a single skill name
// across all target base directories derived from the manifest.
func gitignoreSymlinkEntries(skillName string, m manifest.Manifest) []string {
	var targetBases []string
	if len(m.Outputs) > 0 {
		for base := range m.Outputs {
			targetBases = append(targetBases, base)
		}
	} else {
		bases, err := agents.DeriveTargets(m.HarnessList())
		if err == nil {
			targetBases = bases
		}
		if len(targetBases) == 0 {
			targetBases = []string{".agents/skills/"}
		}
	}
	entries := make([]string, 0, len(targetBases))
	for _, base := range targetBases {
		entries = append(entries, strings.TrimSuffix(base, "/")+"/"+skillName)
	}
	return entries
}

// managedGitignoreEntries returns the list of .gitignore entries melon should
// maintain when vendor: false — the .melon/ cache dir plus every managed symlink path.
func managedGitignoreEntries(resolved []resolver.ResolvedDep, m manifest.Manifest, projectDir string) []string {
	entries := []string{store.StoreDir + "/"}
	for _, dep := range resolved {
		if !placer.ShouldPlace(dep, projectDir) {
			continue // pure packs have no symlink to ignore
		}
		skillName := dep.Name
		if idx := strings.LastIndex(dep.Name, "/"); idx >= 0 {
			skillName = dep.Name[idx+1:]
		}
		entries = append(entries, gitignoreSymlinkEntries(skillName, m)...)
	}
	return entries
}

func fetchDeps(resolved []resolver.ResolvedDep, dir string, onFetch func(i int, name string, err error)) ([]lockfile.LockedDep, error) {
	const maxConcurrent = 4
	sem := make(chan struct{}, maxConcurrent)

	locked := make([]lockfile.LockedDep, len(resolved))
	errs := make([]error, len(resolved))
	var wg sync.WaitGroup
	// onFetch writes progress output; serialize it so parallel fetches never
	// write to the same writer concurrently.
	var onFetchMu sync.Mutex

	for i, dep := range resolved {
		wg.Add(1)
		go func(i int, dep resolver.ResolvedDep) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			installDir := store.InstalledPath(dir, dep)
			result, err := fetchFn(dep, installDir)
			onFetchMu.Lock()
			onFetch(i, fmt.Sprintf("%s@%s", dep.Name, dep.Version), err)
			onFetchMu.Unlock()
			if err != nil {
				errs[i] = fmt.Errorf("install: fetch %s: %w", dep.Name, err)
				return
			}
			locked[i] = lockfile.LockedDep{
				Name:       dep.Name,
				Version:    dep.Version,
				GitTag:     dep.GitTag,
				RepoURL:    dep.RepoURL,
				Subdir:     dep.Subdir,
				Entrypoint: dep.Entrypoint,
				TreeHash:   result.TreeHash,
				Files:      result.Files,
				Type:       lockType(dep.Type),
				Direct:     dep.Direct,
				Via:        dep.Via,
			}
		}(i, dep)
	}

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return locked, nil
}

// recordPackManifests wraps a manifest fetcher so every pack manifest seen
// during resolution is stored in into, keyed by dep name.
func recordPackManifests(
	fetch func(repoURL, gitTag, subdir string) (manifest.Manifest, error),
	into map[string]manifest.Manifest,
) func(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
	return func(repoURL, gitTag, subdir string) (manifest.Manifest, error) {
		mm, err := fetch(repoURL, gitTag, subdir)
		if err == nil && mm.IsPack() {
			name := strings.TrimPrefix(repoURL, "https://")
			if subdir != "" {
				name += "/" + subdir
			}
			into[name] = mm
		}
		return mm, err
	}
}

var exactVersionRe = regexp.MustCompile(`^[=v]?\d+\.\d+\.\d+(?:[-+].*)?$`)

// warnExactPackPins prints a warning for each pack member pinned to an exact
// version. Exact pins in packs conflict with any other constraint on the same
// skill, so pack authors should use ranges.
func warnExactPackPins(out io.Writer, packs map[string]manifest.Manifest) {
	for _, pack := range sortedKeys(packs) {
		deps := packs[pack].Dependencies
		for _, member := range sortedKeys(deps) {
			if exactVersionRe.MatchString(deps[member]) {
				fmt.Fprintf(out, "warning: pack %s pins %s to exact version %s; packs should use a range like ^%s\n",
					pack, member, deps[member], strings.TrimLeft(deps[member], "=v"))
			}
		}
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// installedSummary renders "23 skills and 3 packs" (or "3 skills").
func installedSummary(locked []lockfile.LockedDep) string {
	var skills, packs int
	for _, d := range locked {
		if d.IsPack() {
			packs++
		} else {
			skills++
		}
	}
	s := count(skills, "skill")
	if packs > 0 {
		s += " and " + count(packs, "pack")
	}
	return s
}

// lockType normalises a resolved type for the lock file ("skill" when empty).
func lockType(t string) string {
	if t == "" {
		return manifest.TypeSkill
	}
	return t
}

func printLockDiff(cmd *cobra.Command, d lockfile.LockDiff) {
	if len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Updated) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Nothing changed.")
		return
	}
	for _, dep := range d.Added {
		fmt.Fprintln(cmd.OutOrStdout(), addStyle.Render(fmt.Sprintf("  + %s@%s", dep.Name, dep.Version)))
	}
	for _, dep := range d.Updated {
		fmt.Fprintln(cmd.OutOrStdout(), updateStyle.Render(fmt.Sprintf("  ~ %s@%s", dep.Name, dep.Version)))
	}
	for _, dep := range d.Removed {
		fmt.Fprintln(cmd.OutOrStdout(), removeStyle.Render(fmt.Sprintf("  - %s@%s", dep.Name, dep.Version)))
	}
}
