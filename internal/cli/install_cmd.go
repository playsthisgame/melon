package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/playsthisgame/melon/internal/fetcher"
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

func runInstall(cmd *cobra.Command, args []string) error {
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
	resolved, err := resolver.Resolve(m, resolveVersionFn, fetchManifestFn)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

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
		fmt.Fprintf(os.Stdout, "✓ %d packages installed\n", len(locked))
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

	return nil
}

func fetchDeps(resolved []resolver.ResolvedDep, dir string, onFetch func(i int, name string, err error)) ([]lockfile.LockedDep, error) {
	const maxConcurrent = 4
	sem := make(chan struct{}, maxConcurrent)

	locked := make([]lockfile.LockedDep, len(resolved))
	errs := make([]error, len(resolved))
	var wg sync.WaitGroup

	for i, dep := range resolved {
		wg.Add(1)
		go func(i int, dep resolver.ResolvedDep) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			installDir := store.InstalledPath(dir, dep)
			result, err := fetcher.Fetch(dep, installDir)
			onFetch(i, fmt.Sprintf("%s@%s", dep.Name, dep.Version), err)
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
