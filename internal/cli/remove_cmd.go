package cli

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/playsthisgame/melon/internal/gitignore"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/placer"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/spf13/cobra"
)

func runRemove(cmd *cobra.Command, args []string) error {
	// No argument: launch interactive selector (TTY only).
	if len(args) == 0 {
		return runRemoveInteractive(cmd)
	}

	name := args[0]

	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := manifest.FindPath(dir)

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	lock, _ := lockfile.Load(filepath.Join(dir, "melon.lock"))

	if _, ok := m.Dependencies[name]; !ok {
		for _, ld := range lock.Dependencies {
			if ld.Name == name && len(ld.Via) > 0 {
				return fmt.Errorf("remove: %q is not in melon.yaml; it is required by %s — remove that instead",
					name, strings.Join(ld.Via, ", "))
			}
		}
		return fmt.Errorf("remove: %q is not a dependency in melon.yaml", name)
	}

	delete(m.Dependencies, name)

	// Removing a pack (or anything other deps were installed through) changes
	// the graph: re-resolve so members nothing else needs are pruned.
	if hasDependents(lock, name) {
		return removeViaInstall(cmd, dir, m, manifestPath, lock, name)
	}

	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("remove: save melon.yaml: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Removed %s from melon.yaml\n", name)

	// When vendor: false, remove the stale symlink path(s) from .gitignore now,
	// before running install. This must happen regardless of whether install
	// succeeds, and must not depend on the install pipeline's TTY spinner path.
	if !m.IsVendored() {
		skillName := name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			skillName = name[idx+1:]
		}
		entries := gitignoreSymlinkEntries(skillName, m)
		if removeErr := gitignore.RemoveEntries(filepath.Join(dir, ".gitignore"), entries); removeErr != nil {
			return fmt.Errorf("remove: update .gitignore: %w", removeErr)
		}
	}

	return withSpinner("Updating…", func() error {
		return removeSingleDep(cmd, dir, m, name)
	})
}

// hasDependents reports whether name is a pack in lock or is listed in the via
// of any other locked dep.
func hasDependents(lock lockfile.LockFile, name string) bool {
	for _, ld := range lock.Dependencies {
		if ld.Name == name && ld.IsPack() {
			return true
		}
		for _, v := range ld.Via {
			if v == name {
				return true
			}
		}
	}
	return false
}

// removeViaInstall saves the manifest and runs the full install pipeline, which
// prunes every dep no longer reachable. When vendor is false, the .gitignore
// entries of all pruned deps are removed too.
func removeViaInstall(cmd *cobra.Command, dir string, m manifest.Manifest, manifestPath string, oldLock lockfile.LockFile, name string) error {
	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("remove: save melon.yaml: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Removed %s from melon.yaml\n", name)
	return installAndPrune(cmd, dir, m, oldLock)
}

// installAndPrune runs the full install pipeline against the already-saved
// manifest, which prunes every dep no longer reachable. When vendor is false,
// the .gitignore entries of all pruned deps are removed too.
func installAndPrune(cmd *cobra.Command, dir string, m manifest.Manifest, oldLock lockfile.LockFile) error {
	if err := runInstallCore(cmd, nil); err != nil {
		return err
	}

	if !m.IsVendored() {
		newLock, _ := lockfile.Load(filepath.Join(dir, "melon.lock"))
		var entries []string
		for _, ld := range lockfile.Diff(oldLock, newLock).Removed {
			entries = append(entries, gitignoreSymlinkEntries(skillNameFromDep(ld.Name), m)...)
		}
		if len(entries) > 0 {
			if err := gitignore.RemoveEntries(filepath.Join(dir, ".gitignore"), entries); err != nil {
				return fmt.Errorf("remove: update .gitignore: %w", err)
			}
		}
	}
	return nil
}

// removeSingleDep unplaces, de-stores, and removes a single dep from melon.lock
// without touching any other installed deps.
func removeSingleDep(cmd *cobra.Command, dir string, m manifest.Manifest, name string) error {
	lockPath := filepath.Join(dir, "melon.lock")
	lf, _ := lockfile.Load(lockPath)

	// Find the locked entry so we can unplace and remove the right version.
	var found *lockfile.LockedDep
	remaining := lf.Dependencies[:0]
	for _, ld := range lf.Dependencies {
		if ld.Name == name {
			cp := ld
			found = &cp
		} else {
			remaining = append(remaining, ld)
		}
	}

	if found != nil && !flagNoPlace {
		if err := placer.Unplace([]lockfile.LockedDep{*found}, m, dir, cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("remove: %w", err)
		}
		if err := store.Remove(dir, resolver.ResolvedDep{Name: found.Name, Version: found.Version}); err != nil {
			return fmt.Errorf("remove: %w", err)
		}
	}

	lf.Dependencies = remaining
	lf.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	if err := lockfile.Save(lf, lockPath); err != nil {
		return fmt.Errorf("remove: write melon.lock: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), removeStyle.Render(fmt.Sprintf("  - %s", name)))
	return nil
}

// runRemoveInteractive loads melon.yaml, presents a multi-select TUI, and
// removes the confirmed skills.
func runRemoveInteractive(cmd *cobra.Command) error {
	if !isTTY() {
		return fmt.Errorf("remove: skill name required (non-interactive mode)")
	}

	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := manifest.FindPath(dir)

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	if len(m.Dependencies) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No skills in melon.yaml.")
		return nil
	}

	lock, _ := lockfile.Load(filepath.Join(dir, "melon.lock"))
	packs := lockGraph(lock)

	skills := make([]removeSkillItem, 0, len(m.Dependencies))
	for name, version := range m.Dependencies {
		item := removeSkillItem{name: name, version: version}
		if packs.isPack[name] {
			item.pack = true
			item.skills = len(packs.skillsUnder(name))
		}
		skills = append(skills, item)
	}

	selected, err := runRemoveTUI(skills)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	if len(selected) == 0 {
		return nil
	}

	return offerRemoveMany(cmd, selected)
}

// offerRemoveMany prints the selected skills, prompts for confirmation, then
// removes all of them from melon.yaml in one pass before running a single
// install so that fetch and prune operate on all changes at once.
func offerRemoveMany(cmd *cobra.Command, names []string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := manifest.FindPath(dir)

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	oldLock, _ := lockfile.Load(filepath.Join(dir, "melon.lock"))

	// Work out what a removal really deletes: packs take their members with
	// them, except members something else still needs.
	removing := make(map[string]bool, len(names))
	for _, n := range names {
		removing[n] = true
	}
	remaining := make(map[string]string, len(m.Dependencies))
	for k, v := range m.Dependencies {
		if !removing[k] {
			remaining[k] = v
		}
	}
	pruned := unreachableAfter(oldLock, remaining, removing)
	g := lockGraph(oldLock)

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "\nSelected:\n")
	var nPacks, nPlain int
	packMembers := map[string]bool{}
	for _, n := range names {
		if !g.isPack[n] {
			nPlain++
			fmt.Fprintf(out, "  - %s\n", n)
			continue
		}
		nPacks++
		members := g.skillsUnder(n)
		for s := range members {
			packMembers[s] = true
		}
		fmt.Fprintf(out, "  - %s %s\n", n, packTag(len(members)))
	}

	var removedViaPacks, kept int
	for s := range packMembers {
		if removing[s] {
			continue // counted as a selected skill
		}
		if pruned[s] {
			removedViaPacks++
		} else {
			kept++
		}
	}
	if kept > 0 {
		fmt.Fprintf(out, "  (%s from these packs %s installed because other dependencies still need them)\n",
			count(kept, "skill"), plural(kept, "stays", "stay"))
	}
	fmt.Fprintf(out, "Remove %s? [y/N]: ", selectionSummary(nPacks, removedViaPacks, nPlain))

	reader := bufio.NewReader(cmd.InOrStdin())
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		return nil
	}

	for _, n := range names {
		delete(m.Dependencies, n)
		fmt.Fprintf(out, "Removed %s from melon.yaml\n", n)
	}

	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("remove: save melon.yaml: %w", err)
	}

	// When vendor: false, remove the stale symlink path(s) from .gitignore
	// before running install, for the same reason as in runRemove.
	if !m.IsVendored() {
		var entries []string
		for _, n := range names {
			skillName := n
			if idx := strings.LastIndex(n, "/"); idx >= 0 {
				skillName = n[idx+1:]
			}
			entries = append(entries, gitignoreSymlinkEntries(skillName, m)...)
		}
		if removeErr := gitignore.RemoveEntries(filepath.Join(dir, ".gitignore"), entries); removeErr != nil {
			return fmt.Errorf("remove: update .gitignore: %w", removeErr)
		}
	}

	// Packs (or anything other deps came in through) need a full install so
	// their members are pruned too.
	for _, n := range names {
		if hasDependents(oldLock, n) {
			return installAndPrune(cmd, dir, m, oldLock)
		}
	}

	return withSpinner("Updating…", func() error {
		for _, n := range names {
			if err := removeSingleDep(cmd, dir, m, n); err != nil {
				return err
			}
		}
		return nil
	})
}
