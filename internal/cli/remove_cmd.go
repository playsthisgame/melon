package cli

import (
	"bufio"
	"fmt"
	"os"
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

	if _, ok := m.Dependencies[name]; !ok {
		return fmt.Errorf("remove: %q is not a dependency in melon.yaml", name)
	}

	delete(m.Dependencies, name)

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

	skills := make([]removeSkillItem, 0, len(m.Dependencies))
	for name, version := range m.Dependencies {
		skills = append(skills, removeSkillItem{name: name, version: version})
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
	fmt.Fprintf(cmd.OutOrStdout(), "\nSelected skills:\n")
	for _, n := range names {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", n)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Remove %d skill(s)? [y/N]: ", len(names))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		return nil
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

	for _, n := range names {
		delete(m.Dependencies, n)
		fmt.Fprintf(cmd.OutOrStdout(), "Removed %s from melon.yaml\n", n)
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

	return withSpinner("Updating…", func() error {
		for _, n := range names {
			if err := removeSingleDep(cmd, dir, m, n); err != nil {
				return err
			}
		}
		return nil
	})
}
