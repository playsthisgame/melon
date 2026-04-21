package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/pkg/semver"
	"github.com/spf13/cobra"
)

func runUpdate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if !isTTY() {
			return fmt.Errorf("update: skill name required (non-interactive mode)")
		}
		return runUpdateInteractive(cmd)
	}
	return runUpdateTargeted(cmd, args[0])
}

func runUpdateTargeted(cmd *cobra.Command, name string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := manifest.FindPath(dir)
	lockPath := filepath.Join(dir, "melon.lock")

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	constraint, ok := m.Dependencies[name]
	if !ok {
		return fmt.Errorf("update: %q is not a dependency in melon.yaml", name)
	}

	if isBranchPin(constraint) {
		fmt.Fprintf(cmd.OutOrStdout(), "update: %q is branch-pinned — use melon install to fetch latest\n", name)
		return nil
	}

	lockedVersion := lockedVersionFor(lockPath, name)
	repoURL, _ := fetcher.ParseDepName(name)

	var latestCompatible, absoluteLatest string
	if err := withSpinner("Resolving updates…", func() error {
		var err error
		latestCompatible, _, err = fetcher.LatestMatchingVersion(repoURL, constraint)
		if err != nil {
			return err
		}
		abs, _, absErr := fetcher.LatestTag(repoURL)
		if absErr == nil {
			absoluteLatest = abs
		}
		return nil
	}); err != nil {
		return fmt.Errorf("update: resolve %s: %w", name, err)
	}

	printNewerMajorHint(cmd, name, constraint, absoluteLatest)

	if lockedVersion == latestCompatible {
		fmt.Fprintf(cmd.OutOrStdout(), "%s is already up to date (%s)\n", name, latestCompatible)
		return nil
	}

	return runInstall(cmd, nil)
}

func runUpdateInteractive(cmd *cobra.Command) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := manifest.FindPath(dir)
	lockPath := filepath.Join(dir, "melon.lock")

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	if len(m.Dependencies) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No skills in melon.yaml.")
		return nil
	}

	// Load current locked versions for display in the list.
	lockedVersions := map[string]string{}
	if lf, lfErr := lockfile.Load(lockPath); lfErr == nil {
		for _, dep := range lf.Dependencies {
			lockedVersions[dep.Name] = dep.Version
		}
	}

	// Build the selectable list, filtering out branch-pinned deps.
	var skills []updateSkillItem
	var allSemverDeps []string
	for name, constraint := range m.Dependencies {
		if isBranchPin(constraint) {
			continue
		}
		skills = append(skills, updateSkillItem{
			name:       name,
			constraint: constraint,
			locked:     lockedVersions[name],
		})
		allSemverDeps = append(allSemverDeps, name)
	}

	if len(skills) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No updatable skills (all deps are branch-pinned).")
		return nil
	}

	selected, err := runUpdateTUI(skills, allSemverDeps)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	if len(selected) == 0 {
		return nil
	}

	// Pre-flight: resolve latest compatible versions for each selected dep.
	type updateResult struct {
		name    string
		current string
		latest  string
		hint    string
	}
	var results []updateResult

	if err := withSpinner("Resolving updates…", func() error {
		for _, name := range selected {
			constraint := m.Dependencies[name]
			repoURL, _ := fetcher.ParseDepName(name)

			latest, _, err := fetcher.LatestMatchingVersion(repoURL, constraint)
			if err != nil {
				return fmt.Errorf("resolve %s: %w", name, err)
			}

			var hint string
			if abs, _, absErr := fetcher.LatestTag(repoURL); absErr == nil {
				if !semver.IsCompatible(constraint, abs) {
					hint = fmt.Sprintf("hint: %s %s available — run: melon add %s@^%s",
						name, abs, name, abs)
				}
			}

			results = append(results, updateResult{
				name:    name,
				current: lockedVersions[name],
				latest:  latest,
				hint:    hint,
			})
		}
		return nil
	}); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	var hasUpdates bool
	for _, r := range results {
		if r.hint != "" {
			fmt.Fprintln(cmd.OutOrStdout(), r.hint)
		}
		if r.current == r.latest {
			fmt.Fprintf(cmd.OutOrStdout(), "  already up to date: %s (%s)\n", r.name, r.latest)
		} else {
			hasUpdates = true
		}
	}

	if !hasUpdates {
		fmt.Fprintln(cmd.OutOrStdout(), "All selected skills are up to date.")
		return nil
	}

	return runInstall(cmd, nil)
}

// isBranchPin reports whether constraint is a branch/ref name rather than a
// semver constraint (^X.Y.Z, ~X.Y.Z, or exact X.Y.Z).
func isBranchPin(constraint string) bool {
	if strings.HasPrefix(constraint, "^") || strings.HasPrefix(constraint, "~") {
		return false
	}
	// Exact semver (X.Y.Z): IsCompatible with itself returns true for valid versions.
	v := constraint
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return !semver.IsCompatible(v, v)
}

// lockedVersionFor returns the currently locked version for the named dep, or "".
func lockedVersionFor(lockPath, name string) string {
	lf, err := lockfile.Load(lockPath)
	if err != nil {
		return ""
	}
	for _, dep := range lf.Dependencies {
		if dep.Name == name {
			return dep.Version
		}
	}
	return ""
}

// printNewerMajorHint prints a hint when absoluteLatest is outside constraint.
func printNewerMajorHint(cmd *cobra.Command, name, constraint, absoluteLatest string) {
	if absoluteLatest == "" {
		return
	}
	if !semver.IsCompatible(constraint, absoluteLatest) {
		fmt.Fprintf(cmd.OutOrStdout(), "hint: %s %s is available — run: melon add %s@^%s\n",
			name, absoluteLatest, name, absoluteLatest)
	}
}

