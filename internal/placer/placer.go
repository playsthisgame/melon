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
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
)

// Place creates a directory symlink for each dep in every agent directory
// derived from m.AgentCompat (or m.Outputs if declared). Each symlink points
// from <agent-dir>/<skill-name> into the corresponding .melon/ cache entry.
// Existing entries at the link path are removed before the symlink is created
// (idempotent). Returns the first error encountered.
func Place(deps []resolver.ResolvedDep, m manifest.Manifest, projectDir string, out io.Writer) error {
	// Determine target base paths.
	var targetBases []string

	if len(m.Outputs) > 0 {
		// Explicit outputs override agent_compat derivation.
		for base := range m.Outputs {
			targetBases = append(targetBases, base)
		}
	} else {
		var err error
		targetBases, err = agents.DeriveTargets(m.AgentCompat)
		if err != nil {
			return fmt.Errorf("placer: %w", err)
		}
	}

	if len(targetBases) == 0 {
		fmt.Fprintln(out, "placer: no target agent directories — set agent_compat in melon.yml")
		return nil
	}

	for _, dep := range deps {
		// skillName is the last path segment of dep.Name.
		skillName := dep.Name
		if idx := strings.LastIndex(dep.Name, "/"); idx >= 0 {
			skillName = dep.Name[idx+1:]
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
