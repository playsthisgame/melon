package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/gitignore"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/placer"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/spf13/cobra"
)

func runAdd(cmd *cobra.Command, args []string) error {
	arg := args[0]

	// Parse <dep>[@<constraint>]. Everything before the first "@" is the dep name;
	// everything after is the constraint. If no "@" is present, we resolve the
	// latest semver tag and build a "^<version>" constraint.
	name, constraint, hasConstraint := strings.Cut(arg, "@")

	// Warn if the dep name looks like a file path — users sometimes paste a
	// GitHub blob URL including the filename (e.g. .../SKILL.md).
	// The dep name must be a directory path, not a file.
	if strings.HasSuffix(name, ".md") {
		return fmt.Errorf("add: dep name %q looks like a file path — use the directory path instead\n  e.g.  mln add github.com/owner/repo/path/to/skill", name)
	}
	if !hasConstraint || constraint == "" {
		// No constraint supplied — find the latest semver tag.
		repoURL, _ := fetcher.ParseDepName(name)
		var version string
		if err := withSpinner(fmt.Sprintf("Resolving latest version of %s…", name), func() error {
			var err error
			version, _, err = fetcher.LatestTag(repoURL)
			return err
		}); err != nil {
			if !errors.Is(err, fetcher.ErrNoSemverTags) {
				return fmt.Errorf("add: resolve latest tag for %s: %w", name, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "warning: no semver tags found for %s — using main branch\n", name)
			constraint = "main"
		} else {
			constraint = "^" + version
		}
	}

	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := manifest.FindPath(dir)

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("add: %w", err)
	}

	if err := checkSourcePolicy(m, []string{name}); err != nil {
		return fmt.Errorf("add: %w", err)
	}

	if m.Dependencies == nil {
		m.Dependencies = make(map[string]string)
	}

	if existing, ok := m.Dependencies[name]; ok {
		fmt.Fprintf(cmd.OutOrStdout(), "warning: updating %s from %s to %s\n", name, existing, constraint)
	}
	m.Dependencies[name] = constraint

	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("add: save melon.yaml: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Added %s %s to melon.yaml\n", name, constraint)

	// Install only the new dep — existing deps are already fetched, locked, and placed.
	return installSingleDep(cmd, dir, m, name, constraint)
}

// installSingleDep resolves, fetches, locks, and places a single dependency
// without touching any other already-installed deps.
func installSingleDep(cmd *cobra.Command, dir string, m manifest.Manifest, name, constraint string) error {
	repoURL, subdir := fetcher.ParseDepName(name)

	// Resolve the pinned version for this constraint.
	version, gitTag, err := resolveVersionFn(repoURL, constraint)
	if err != nil {
		return fmt.Errorf("add: resolve %s: %w", name, err)
	}

	// Fetch the dep's own manifest to get its declared entrypoint.
	depManifest, err := fetchManifestFn(repoURL, gitTag, subdir)
	if err != nil {
		return fmt.Errorf("add: fetch manifest for %s: %w", name, err)
	}
	entrypoint := "SKILL.md"
	if depManifest.Entrypoint != "" {
		entrypoint = depManifest.Entrypoint
	}

	dep := resolver.ResolvedDep{
		Name:       name,
		Version:    version,
		GitTag:     gitTag,
		RepoURL:    repoURL,
		Subdir:     subdir,
		Entrypoint: entrypoint,
	}

	// Fetch into the store (idempotent — skips if tree hash matches).
	if err := os.MkdirAll(filepath.Join(dir, store.StoreDir), 0755); err != nil {
		return fmt.Errorf("add: create store: %w", err)
	}
	installDir := store.InstalledPath(dir, dep)
	result, err := fetchFn(dep, installDir)
	if err != nil {
		return fmt.Errorf("add: fetch %s: %w", name, err)
	}
	dep.TreeHash = result.TreeHash
	fmt.Fprintln(cmd.OutOrStdout(), addStyle.Render(fmt.Sprintf("  + %s@%s", name, version)))

	// Upsert this dep in melon.lock, preserving all other entries.
	lockPath := filepath.Join(dir, "melon.lock")
	lf, _ := lockfile.Load(lockPath)

	newEntry := lockfile.LockedDep{
		Name:       dep.Name,
		Version:    dep.Version,
		GitTag:     dep.GitTag,
		RepoURL:    dep.RepoURL,
		Subdir:     dep.Subdir,
		Entrypoint: dep.Entrypoint,
		TreeHash:   result.TreeHash,
		Files:      result.Files,
	}

	var oldVersion string
	upserted := false
	for i, ld := range lf.Dependencies {
		if ld.Name == name {
			oldVersion = ld.Version
			lf.Dependencies[i] = newEntry
			upserted = true
			break
		}
	}
	if !upserted {
		lf.Dependencies = append(lf.Dependencies, newEntry)
	}
	lf.GeneratedAt = time.Now().UTC().Format(time.RFC3339)

	if err := lockfile.Save(lf, lockPath); err != nil {
		return fmt.Errorf("add: write melon.lock: %w", err)
	}

	// Remove old store entry when the version changed.
	if oldVersion != "" && oldVersion != version {
		_ = store.Remove(dir, resolver.ResolvedDep{Name: name, Version: oldVersion})
	}

	// Place the skill into agent directories.
	if !flagNoPlace {
		if err := placer.Place([]resolver.ResolvedDep{dep}, m, dir, cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("add: %w", err)
		}
	}

	// Sync .gitignore when vendor: false.
	if !m.IsVendored() {
		gitignorePath := filepath.Join(dir, ".gitignore")
		entries := managedGitignoreEntries([]resolver.ResolvedDep{dep}, m, dir)
		if _, err := gitignore.EnsureEntries(gitignorePath, entries); err != nil {
			return fmt.Errorf("add: update .gitignore: %w", err)
		}
	}

	return nil
}
