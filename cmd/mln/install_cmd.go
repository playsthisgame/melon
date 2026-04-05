package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
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

func runInstall(cmd *cobra.Command, args []string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(dir, "melon.yml")
	lockPath := filepath.Join(dir, "melon.lock")

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

	// Load existing lock file for diff display (ignore error if absent).
	var oldLock lockfile.LockFile
	oldLock, _ = lockfile.Load(lockPath)

	if len(m.Dependencies) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No dependencies declared in melon.yml.")
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
	resolved, err := resolver.Resolve(m, fetcher.LatestMatchingVersion, resolver.DefaultFetchManifest)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}

	// Fetch each resolved dep into the store.
	var locked []lockfile.LockedDep
	if isTTY() {
		model := newInstallProgressModel(len(resolved))
		p := tea.NewProgram(model)
		var fetchErr error
		go func() {
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
			return fmt.Errorf("install: %w", runErr)
		}
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

	// Place skills into agent directories unless --no-place is set.
	if !flagNoPlace {
		if err := placer.Place(resolved, m, dir, cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("install: %w", err)
		}
	}

	return nil
}

func fetchDeps(resolved []resolver.ResolvedDep, dir string, onFetch func(i int, name string, err error)) ([]lockfile.LockedDep, error) {
	var locked []lockfile.LockedDep
	for i, dep := range resolved {
		installDir := store.InstalledPath(dir, dep)
		result, err := fetcher.Fetch(dep, installDir)
		onFetch(i, fmt.Sprintf("%s@%s", dep.Name, dep.Version), err)
		if err != nil {
			return nil, fmt.Errorf("install: fetch %s: %w", dep.Name, err)
		}
		locked = append(locked, lockfile.LockedDep{
			Name:       dep.Name,
			Version:    dep.Version,
			GitTag:     dep.GitTag,
			RepoURL:    dep.RepoURL,
			Subdir:     dep.Subdir,
			Entrypoint: dep.Entrypoint,
			TreeHash:   result.TreeHash,
			Files:      result.Files,
		})
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
