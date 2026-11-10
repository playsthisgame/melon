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

	// Resolve the pinned version and read the dep's own manifest before
	// touching melon.yaml, so we know whether it is a plain skill or a pack.
	repoURL, subdir := fetcher.ParseDepName(name)
	version, gitTag, err := resolveVersionFn(repoURL, constraint)
	if err != nil {
		return fmt.Errorf("add: resolve %s: %w", name, err)
	}
	depManifest, err := fetchManifestFn(repoURL, gitTag, subdir)
	if err != nil {
		return fmt.Errorf("add: fetch manifest for %s: %w", name, err)
	}

	existing, hadExisting := m.Dependencies[name]
	m.Dependencies[name] = constraint

	// Packs, and skills with their own dependencies, change the shape of the
	// graph: resolve everything and run the full install pipeline.
	if depManifest.IsPack() || len(depManifest.Dependencies) > 0 {
		proceed, err := previewGraphAdd(cmd, m, name, depManifest)
		if err != nil || !proceed {
			return err
		}
		if hadExisting {
			fmt.Fprintf(cmd.OutOrStdout(), "warning: updating %s from %s to %s\n", name, existing, constraint)
		}
		if err := manifest.Save(m, manifestPath); err != nil {
			return fmt.Errorf("add: save melon.yaml: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Added %s %s to melon.yaml\n", name, constraint)
		return runInstallCore(cmd, nil)
	}

	if hadExisting {
		fmt.Fprintf(cmd.OutOrStdout(), "warning: updating %s from %s to %s\n", name, existing, constraint)
	}
	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("add: save melon.yaml: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Added %s %s to melon.yaml\n", name, constraint)

	// Install only the new dep — existing deps are already fetched, locked, and placed.
	return installSingleDep(cmd, dir, m, name, version, gitTag, depManifest)
}

// previewGraphAdd resolves m (which already contains the new dep, in memory
// only), enforces allowed_sources on the full graph, and — for packs — shows
// the member list and asks for confirmation in a TTY. It returns false with a
// nil error when the user declines. melon.yaml is never written here.
func previewGraphAdd(cmd *cobra.Command, m manifest.Manifest, name string, depManifest manifest.Manifest) (bool, error) {
	var resolved []resolver.ResolvedDep
	if err := withSpinner(fmt.Sprintf("Resolving dependencies of %s…", name), func() error {
		var err error
		resolved, err = resolver.Resolve(m, resolveVersionFn, fetchManifestFn)
		return err
	}); err != nil {
		return false, fmt.Errorf("add: %w", err)
	}
	if err := checkResolvedPolicy(m, resolved); err != nil {
		return false, fmt.Errorf("add: %w", err)
	}
	if !depManifest.IsPack() {
		return true, nil
	}

	// Count the skills the pack really brings in, including those inside
	// nested packs, rather than its direct member entries.
	g := resolvedGraph(resolved)
	total := len(g.skillsUnder(name))
	members := sortedKeys(depManifest.Dependencies)
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s %s\n", name, packTag(total))
	for _, member := range members {
		label := ""
		if g.isPack[member] {
			label = " " + packTag(len(g.skillsUnder(member)))
		}
		fmt.Fprintf(out, "  - %s %s%s\n", member, depManifest.Dependencies[member], label)
	}

	if flagAddYes || !addIsTTY() {
		return true, nil
	}
	fmt.Fprintf(out, "Install %s? [Y/n]: ", selectionSummary(1, total, 0))
	answer := strings.ToLower(strings.TrimSpace(readLine(cmd.InOrStdin())))
	if answer != "" && answer != "y" && answer != "yes" {
		fmt.Fprintln(out, "Aborted; melon.yaml unchanged.")
		return false, nil
	}
	return true, nil
}

// addIsTTY is overridable in tests to exercise the confirmation prompt.
var addIsTTY = isTTY

// installSingleDep fetches, locks, and places a single, already-resolved
// dependency without touching any other already-installed deps.
func installSingleDep(cmd *cobra.Command, dir string, m manifest.Manifest, name, version, gitTag string, depManifest manifest.Manifest) error {
	repoURL, subdir := fetcher.ParseDepName(name)

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
	if err := os.MkdirAll(store.Dir(dir), 0755); err != nil {
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
		Type:       depManifest.EffectiveType(),
		Direct:     true,
	}

	var oldVersion string
	upserted := false
	for i, ld := range lf.Dependencies {
		if ld.Name == name {
			oldVersion = ld.Version
			// Keep pack provenance if this skill was already installed via a pack.
			newEntry.Via = ld.Via
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
	if managesGitignore(m, dir) {
		gitignorePath := filepath.Join(dir, ".gitignore")
		entries := managedGitignoreEntries([]resolver.ResolvedDep{dep}, m, dir)
		if _, err := gitignore.EnsureEntries(gitignorePath, entries); err != nil {
			return fmt.Errorf("add: update .gitignore: %w", err)
		}
	}

	return nil
}
