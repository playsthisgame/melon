package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/playsthisgame/melon/internal/agents"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove cached entries in .melon/ not referenced by melon.lock",
	Args:  cobra.NoArgs,
	RunE:  runClean,
}

func runClean(cmd *cobra.Command, args []string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}

	lockPath := filepath.Join(dir, "melon.lock")

	// 3.1 – No lock file: inform and exit cleanly.
	lf, err := lockfile.Load(lockPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || isFileNotFound(err) {
			fmt.Fprintln(cmd.OutOrStdout(), "No melon.lock found. Run 'melon install' first.")
			return nil
		}
		return fmt.Errorf("clean: %w", err)
	}

	// Build a set of dir names that are currently locked.
	locked := make(map[string]lockfile.LockedDep, len(lf.Dependencies))
	for _, dep := range lf.Dependencies {
		locked[store.DirName(dep.Name, dep.Version)] = dep
	}

	// List what is currently in .melon/.
	storeDir := filepath.Join(dir, store.StoreDir)
	entries, err := os.ReadDir(storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(cmd.OutOrStdout(), "Nothing to clean.")
			return nil
		}
		return fmt.Errorf("clean: read store: %w", err)
	}

	// 3.2 – Identify orphaned entries.
	var orphans []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, ok := locked[e.Name()]; !ok {
			orphans = append(orphans, e.Name())
		}
	}

	if len(orphans) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Nothing to clean.")
		return nil
	}

	// 2.1 – Derive agent skill dirs best-effort from manifest.
	var skillDirs []string
	manifestPath := manifest.FindPath(dir)
	m, manifestErr := manifest.Load(manifestPath)
	hasManifest := manifestErr == nil
	if hasManifest {
		if len(m.Outputs) > 0 {
			for base := range m.Outputs {
				skillDirs = append(skillDirs, filepath.Join(dir, base))
			}
		} else {
			targets, _ := agents.DeriveTargets(m.ToolCompat)
			for _, t := range targets {
				skillDirs = append(skillDirs, filepath.Join(dir, t))
			}
		}
	}

	// 3.3 / 2.2 – Remove each orphaned cache entry and any symlinks pointing to it.
	removed := 0
	for _, dirEntry := range orphans {
		entryPath := filepath.Join(storeDir, dirEntry)

		// Remove any symlinks in agent skill dirs that point into this cache entry.
		for _, skillDir := range skillDirs {
			removeSymlinksPointingTo(skillDir, entryPath, cmd)
		}

		// Remove the cache directory itself.
		if err := os.RemoveAll(entryPath); err != nil {
			return fmt.Errorf("clean: remove %s: %w", dirEntry, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", removeStyle.Render("removed"), dirEntry)
		removed++
	}

	if !hasManifest {
		fmt.Fprintln(cmd.OutOrStdout(), "Warning: melon.yaml not found — symlinks were not removed.")
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%d cache entr%s cleaned.\n", removed, plural(removed, "y", "ies"))
	return nil
}

// removeSymlinksPointingTo removes any symlinks inside skillDir whose target
// resolves to (or is a child of) cacheEntryPath. Missing skillDir is silently ignored.
func removeSymlinksPointingTo(skillDir, cacheEntryPath string, cmd *cobra.Command) {
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return // dir doesn't exist or unreadable — skip silently
	}
	for _, e := range entries {
		linkPath := filepath.Join(skillDir, e.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue // not a symlink
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(skillDir, target)
		}
		if target == cacheEntryPath || strings.HasPrefix(target, cacheEntryPath+string(os.PathSeparator)) {
			if removeErr := os.Remove(linkPath); removeErr != nil && !os.IsNotExist(removeErr) {
				fmt.Fprintf(cmd.OutOrStdout(), "  warning: could not remove symlink %s: %v\n", linkPath, removeErr)
			} else {
				rel, _ := filepath.Rel(filepath.Dir(skillDir), linkPath)
				fmt.Fprintf(cmd.OutOrStdout(), "  unlinked %s\n", rel)
			}
		}
	}
}

// isFileNotFound reports whether err contains a "no such file" message, which
// lockfile.Load wraps rather than returning directly.
func isFileNotFound(err error) bool {
	return err != nil && (os.IsNotExist(err) || strings.Contains(err.Error(), "no such file"))
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return singular
	}
	return pluralForm
}
