// Package placer installs skills into agent directories by creating directory
// symlinks that point back into the .melon/ package cache.
package placer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/playsthisgame/melon/internal/agents"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
)

// ShouldPlace reports whether dep gets a symlink in harness directories.
// Skills always do. A pack does only when its cached directory contains its
// entrypoint file (a hybrid pack); a pure pack, which only lists members, does
// not. Place, Unplace, gitignore management and list --check all use this so
// they agree on which deps have symlinks.
func ShouldPlace(dep resolver.ResolvedDep, projectDir string) bool {
	if !dep.IsPack() {
		return true
	}
	entrypoint := dep.Entrypoint
	if entrypoint == "" {
		entrypoint = "SKILL.md"
	}
	_, err := os.Stat(filepath.Join(store.InstalledPath(projectDir, dep), entrypoint))
	return err == nil
}

// ShouldPlaceLocked is ShouldPlace for a lock file entry.
func ShouldPlaceLocked(dep lockfile.LockedDep, projectDir string) bool {
	return ShouldPlace(resolver.ResolvedDep{
		Name: dep.Name, Version: dep.Version, Entrypoint: dep.Entrypoint, Type: dep.Type,
	}, projectDir)
}

// Unplace removes the agent directory symlink for each dep in every agent
// directory derived from m.HarnessList() (or m.Outputs if declared). Missing
// symlinks are silently ignored. Returns the first error encountered.
func Unplace(deps []lockfile.LockedDep, m manifest.Manifest, projectDir string, out io.Writer) error {
	var targetBases []string

	if len(m.Outputs) > 0 {
		for base := range m.Outputs {
			targetBases = append(targetBases, base)
		}
	} else {
		var err error
		targetBases, err = agents.DeriveTargets(m.HarnessList())
		if err != nil {
			return fmt.Errorf("placer: %w", err)
		}
	}

	for _, dep := range deps {
		skillName := dep.Name
		if idx := strings.LastIndex(dep.Name, "/"); idx >= 0 {
			skillName = dep.Name[idx+1:]
		}

		if !ShouldPlaceLocked(dep, projectDir) {
			// Pure pack: nothing was linked. Quietly clear any stale link left
			// by an older version that was a hybrid pack.
			for _, base := range targetBases {
				removeIfSymlink(filepath.Join(projectDir, base, skillName))
			}
			continue
		}

		for _, base := range targetBases {
			linkPath := filepath.Join(projectDir, base, skillName)
			if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("placer: remove %s: %w", linkPath, err)
			}
			rel, _ := filepath.Rel(projectDir, linkPath)
			fmt.Fprintf(out, "  unlinked %s (%s)\n", dep.Name, rel)
		}
	}
	return nil
}

// Place creates a directory symlink for each dep in every agent directory
// derived from m.HarnessList() (or m.Outputs if declared). Each symlink points
// from <agent-dir>/<skill-name> into the corresponding .melon/ cache entry.
// Existing entries at the link path are removed before the symlink is created
// (idempotent). Returns the first error encountered.
func Place(deps []resolver.ResolvedDep, m manifest.Manifest, projectDir string, out io.Writer) error {
	// Determine target base paths.
	var targetBases []string

	if len(m.Outputs) > 0 {
		// Explicit outputs override harnesses derivation.
		for base := range m.Outputs {
			targetBases = append(targetBases, base)
		}
	} else {
		var err error
		targetBases, err = agents.DeriveTargets(m.HarnessList())
		if err != nil {
			return fmt.Errorf("placer: %w", err)
		}
	}

	if len(targetBases) == 0 {
		targetBases = []string{".agents/skills/"}
	}

	for _, dep := range deps {
		// skillName is the last path segment of dep.Name.
		skillName := dep.Name
		if idx := strings.LastIndex(dep.Name, "/"); idx >= 0 {
			skillName = dep.Name[idx+1:]
		}

		if !ShouldPlace(dep, projectDir) {
			// Pure pack: its members are placed, the pack itself is not. Clear a
			// stale link in case a previous version was a hybrid pack.
			for _, base := range targetBases {
				removeIfSymlink(filepath.Join(projectDir, base, skillName))
			}
			continue
		}

		cacheDir := store.InstalledPath(projectDir, dep)

		for _, base := range targetBases {
			linkDir := filepath.Join(projectDir, base)
			linkPath := filepath.Join(linkDir, skillName)

			// Ensure the parent skills directory exists.
			if err := os.MkdirAll(linkDir, 0755); err != nil {
				return fmt.Errorf("placer: mkdir %s: %w", linkDir, err)
			}

			// Remove any existing file, symlink, or directory at the link path.
			if err := os.RemoveAll(linkPath); err != nil {
				return fmt.Errorf("placer: remove existing %s: %w", linkPath, err)
			}

			// Compute a relative symlink target so the project is portable.
			relTarget, err := filepath.Rel(linkDir, cacheDir)
			if err != nil {
				return fmt.Errorf("placer: rel path for %s: %w", dep.Name, err)
			}

			if err := os.Symlink(relTarget, linkPath); err != nil {
				return fmt.Errorf("placer: symlink %s -> %s: %w", linkPath, relTarget, err)
			}

			rel, _ := filepath.Rel(projectDir, linkPath)
			fmt.Fprintf(out, "  linked %s -> %s\n", dep.Name, rel)
		}
	}
	return nil
}

// removeIfSymlink deletes path only when it is a symlink, leaving real files
// and directories a user may have created untouched.
func removeIfSymlink(path string) {
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		_ = os.Remove(path)
	}
}
