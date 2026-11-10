// Package placer installs skills into agent directories by creating directory
// symlinks that point back into the .melon/ package cache (or ~/.melon/store/
// for global installs).
package placer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/playsthisgame/melon/internal/agents"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/scope"
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

// TargetDirs returns the absolute skill directories m places into, sorted.
// For a project they are m.Outputs keys (or the harnesses' project
// directories) joined to projectDir. For the global root they are the
// harnesses' global directories, with m.Outputs keys read as "~/..." or
// home-relative paths. When nothing is configured and withDefault is set, the
// cross-tool .agents/skills/ directory (~/.agents/skills/ globally) is used.
func TargetDirs(m manifest.Manifest, projectDir string, withDefault bool) ([]string, error) {
	global := scope.IsGlobal(projectDir)

	var bases []string
	if len(m.Outputs) > 0 {
		for base := range m.Outputs {
			bases = append(bases, base)
		}
	} else {
		var err error
		if global {
			bases, err = agents.DeriveGlobalTargets(m.HarnessList())
		} else {
			bases, err = agents.DeriveTargets(m.HarnessList())
		}
		if err != nil {
			return nil, err
		}
	}
	if len(bases) == 0 && withDefault {
		if global {
			bases = []string{"~/.agents/skills/"}
		} else {
			bases = []string{".agents/skills/"}
		}
	}

	seen := make(map[string]struct{}, len(bases))
	dirs := make([]string, 0, len(bases))
	for _, base := range bases {
		var dir string
		if global {
			var err error
			if dir, err = scope.ExpandHome(base); err != nil {
				return nil, err
			}
		} else {
			dir = filepath.Join(projectDir, base)
		}
		if _, dup := seen[dir]; dup {
			continue
		}
		seen[dir] = struct{}{}
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs, nil
}

// skillName is the last path segment of a dep name.
func skillName(depName string) string {
	if idx := strings.LastIndex(depName, "/"); idx >= 0 {
		return depName[idx+1:]
	}
	return depName
}

// Unplace removes the agent directory symlink for each dep in every agent
// directory derived from m.HarnessList() (or m.Outputs if declared). Missing
// symlinks are silently ignored. For the global root, only symlinks that point
// into melon's store are removed, so skills the user put in a global harness
// directory themselves are never touched. Returns the first error encountered.
func Unplace(deps []lockfile.LockedDep, m manifest.Manifest, projectDir string, out io.Writer) error {
	targetDirs, err := TargetDirs(m, projectDir, false)
	if err != nil {
		return fmt.Errorf("placer: %w", err)
	}
	global := scope.IsGlobal(projectDir)

	for _, dep := range deps {
		name := skillName(dep.Name)

		if !ShouldPlaceLocked(dep, projectDir) {
			// Pure pack: nothing was linked. Quietly clear any stale link left
			// by an older version that was a hybrid pack.
			for _, dir := range targetDirs {
				linkPath := filepath.Join(dir, name)
				if !global || isManagedLink(linkPath, projectDir) {
					removeIfSymlink(linkPath)
				}
			}
			continue
		}

		for _, dir := range targetDirs {
			linkPath := filepath.Join(dir, name)
			if global && !isManagedLink(linkPath, projectDir) {
				continue
			}
			if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("placer: remove %s: %w", linkPath, err)
			}
			fmt.Fprintf(out, "  unlinked %s (%s)\n", dep.Name, scope.Display(projectDir, linkPath))
		}
	}
	return nil
}

// Place creates a directory symlink for each dep in every agent directory
// derived from m.HarnessList() (or m.Outputs if declared). Each symlink points
// from <agent-dir>/<skill-name> into the corresponding store entry.
//
// For a project, existing entries at the link path are removed before the
// symlink is created (idempotent) and links are relative so the project stays
// portable. For the global root, links are absolute, and an existing entry that
// is not a melon-managed symlink is left alone: the skill is skipped there and
// every skip is reported once all skills are placed. Returns the first error
// encountered.
func Place(deps []resolver.ResolvedDep, m manifest.Manifest, projectDir string, out io.Writer) error {
	targetDirs, err := TargetDirs(m, projectDir, true)
	if err != nil {
		return fmt.Errorf("placer: %w", err)
	}
	global := scope.IsGlobal(projectDir)
	var skipped []string

	for _, dep := range deps {
		name := skillName(dep.Name)

		if !ShouldPlace(dep, projectDir) {
			// Pure pack: its members are placed, the pack itself is not. Clear a
			// stale link in case a previous version was a hybrid pack.
			for _, dir := range targetDirs {
				linkPath := filepath.Join(dir, name)
				if !global || isManagedLink(linkPath, projectDir) {
					removeIfSymlink(linkPath)
				}
			}
			continue
		}

		cacheDir := store.InstalledPath(projectDir, dep)

		for _, linkDir := range targetDirs {
			linkPath := filepath.Join(linkDir, name)

			if global {
				if _, err := os.Lstat(linkPath); err == nil && !isManagedLink(linkPath, projectDir) {
					skipped = append(skipped, scope.Display(projectDir, linkPath))
					continue
				}
			}

			// Ensure the parent skills directory exists.
			if err := os.MkdirAll(linkDir, 0755); err != nil {
				return fmt.Errorf("placer: mkdir %s: %w", linkDir, err)
			}

			// Remove any existing file, symlink, or directory at the link path.
			if err := os.RemoveAll(linkPath); err != nil {
				return fmt.Errorf("placer: remove existing %s: %w", linkPath, err)
			}

			// Project links are relative so the project is portable. Global
			// links are absolute so they survive a symlinked dotfiles directory.
			target := cacheDir
			if !global {
				if target, err = filepath.Rel(linkDir, cacheDir); err != nil {
					return fmt.Errorf("placer: rel path for %s: %w", dep.Name, err)
				}
			}

			if err := os.Symlink(target, linkPath); err != nil {
				return fmt.Errorf("placer: symlink %s -> %s: %w", linkPath, target, err)
			}

			fmt.Fprintf(out, "  linked %s -> %s\n", dep.Name, scope.Display(projectDir, linkPath))
		}
	}

	if len(skipped) > 0 {
		fmt.Fprintf(out, "\nSkipped %d %s already present and not managed by melon:\n",
			len(skipped), plural(len(skipped), "skill", "skills"))
		for _, p := range skipped {
			fmt.Fprintf(out, "  %s\n", p)
		}
		fmt.Fprintln(out, "Remove or rename them, then run: melon install --global")
	}
	return nil
}

// IsManagedLink reports whether path is a symlink pointing into projectDir's
// store, meaning melon created it.
func IsManagedLink(path, projectDir string) bool { return isManagedLink(path, projectDir) }

func isManagedLink(path, projectDir string) bool {
	target, err := os.Readlink(path)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	storeDir := filepath.Clean(store.Dir(projectDir))
	target = filepath.Clean(target)
	return strings.HasPrefix(target, storeDir+string(os.PathSeparator))
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// removeIfSymlink deletes path only when it is a symlink, leaving real files
// and directories a user may have created untouched.
func removeIfSymlink(path string) {
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		_ = os.Remove(path)
	}
}
