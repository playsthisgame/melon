package main

import (
	"fmt"
	"path/filepath"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/spf13/cobra"
)

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(dir, "melon.yml")

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	if _, ok := m.Dependencies[name]; !ok {
		return fmt.Errorf("remove: %q is not a dependency in melon.yml", name)
	}

	delete(m.Dependencies, name)

	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("remove: save melon.yml: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Removed %s from melon.yml\n", name)

	// Run the full install pipeline — this regenerates melon.lock and prunes
	// the removed dep's agent symlink and .melon/ cache entry.
	return withSpinner("Updating…", func() error {
		return runInstall(cmd, args)
	})
}
