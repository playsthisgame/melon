package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/playsthisgame/melon/internal/manifest"
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

	// Run the full install pipeline — this regenerates melon.lock and prunes
	// the removed dep's agent symlink and .melon/ cache entry.
	return withSpinner("Updating…", func() error {
		return runInstall(cmd, args)
	})
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

	return withSpinner("Updating…", func() error {
		return runInstall(cmd, nil)
	})
}
