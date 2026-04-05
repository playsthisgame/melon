package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/manifest"
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
			return fmt.Errorf("add: resolve latest tag for %s: %w", name, err)
		}
		constraint = "^" + version
	}

	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(dir, "melon.yml")

	m, err := manifest.Load(manifestPath)
	if err != nil {
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
		return fmt.Errorf("add: save melon.yml: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Added %s %s to melon.yml\n", name, constraint)

	// Run the full install pipeline (resolve → fetch → lock → place).
	return runInstall(cmd, args)
}
