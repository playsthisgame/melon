// Package scope locates melon's user-global install: the global manifest and
// lock file live in ~/.config/melon/, fetched skills are cached in
// ~/.melon/store/, and skills are linked into each harness's global skills
// directory (for example ~/.claude/skills/).
//
// Everything is derived from the user's home directory, so the global root is
// just another "project directory" to the rest of melon: commands load
// melon.yaml and melon.lock from it as usual, and IsGlobal tells the store and
// placer to use the global cache and harness directories instead of
// project-relative ones.
package scope

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// globalRootRel is the global manifest directory, relative to home.
	globalRootRel = ".config/melon"
	// globalStoreRel is the global skill cache, relative to home.
	globalStoreRel = ".melon/store"
	// ProjectStoreDir is the per-project cache directory name.
	ProjectStoreDir = ".melon"
)

// Home returns the user's home directory.
func Home() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("scope: find home directory: %w", err)
	}
	return home, nil
}

// GlobalRoot returns the directory holding the global melon.yaml and melon.lock.
func GlobalRoot() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, globalRootRel), nil
}

// GlobalStore returns the global skill cache directory.
func GlobalStore() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, globalStoreRel), nil
}

// IsGlobal reports whether dir is the global root.
func IsGlobal(dir string) bool {
	root, err := GlobalRoot()
	if err != nil {
		return false
	}
	return filepath.Clean(dir) == filepath.Clean(root)
}

// StoreDir returns the skill cache directory for dir: the global store when
// dir is the global root, otherwise dir/.melon.
func StoreDir(dir string) string {
	if IsGlobal(dir) {
		if s, err := GlobalStore(); err == nil {
			return s
		}
	}
	return filepath.Join(dir, ProjectStoreDir)
}

// ExpandHome turns a "~/..." path into an absolute one. Absolute paths are
// returned unchanged and other relative paths are taken relative to home.
func ExpandHome(p string) (string, error) {
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	p = strings.TrimPrefix(strings.TrimPrefix(p, "~"), "/")
	return filepath.Join(home, p), nil
}

// Tilde shortens an absolute path under home to "~/..." for display.
func Tilde(p string) string {
	home, err := Home()
	if err != nil {
		return p
	}
	if p == home {
		return "~"
	}
	if rel, err := filepath.Rel(home, p); err == nil && !strings.HasPrefix(rel, "..") {
		return "~/" + filepath.ToSlash(rel)
	}
	return p
}

// Display renders path p for output: "~/..." for global installs, otherwise
// relative to the project directory dir.
func Display(dir, p string) string {
	if IsGlobal(dir) {
		return Tilde(p)
	}
	if rel, err := filepath.Rel(dir, p); err == nil {
		return rel
	}
	return p
}
