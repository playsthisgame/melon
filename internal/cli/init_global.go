package cli

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/playsthisgame/melon/internal/agents"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/scope"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/spf13/cobra"
)

// runInitGlobal implements `melon init --global`: it writes the global
// manifest to ~/.config/melon/melon.yaml with the harnesses that global
// installs should target, and creates the ~/.melon/store/ cache.
func runInitGlobal(cmd *cobra.Command, dir string) error {
	out := cmd.OutOrStdout()
	manifestPath := manifest.FindPath(dir)

	if _, err := os.Stat(manifestPath); err == nil {
		if flagYes {
			fmt.Fprintf(out, "%s already exists — overwriting (--yes)\n", scope.Tilde(manifestPath))
		} else {
			fmt.Fprintf(out, "%s already exists. Overwrite? [y/N] ", scope.Tilde(manifestPath))
			answer := readLine(cmd.InOrStdin())
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				fmt.Fprintln(out, "Aborted.")
				return nil
			}
		}
	}

	var harnesses []string
	if flagYes || !isTTY() {
		harnesses = promptMultiChoice(cmd, "Harnesses to install global skills for?", agents.KnownAgents(), []string{})
	} else {
		p := tea.NewProgram(newGlobalInitModel())
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("init: %w", err)
		}
		fm := finalModel.(initModel)
		if fm.step != stepDone {
			fmt.Fprintln(out, "Aborted.")
			return nil
		}
		harnesses = fm.result.agentNames
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("init: create %s: %w", scope.Tilde(dir), err)
	}
	content, err := generateGlobalManifestYAML(harnesses)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("init: write %s: %w", scope.Tilde(manifestPath), err)
	}
	if err := os.MkdirAll(store.Dir(dir), 0755); err != nil {
		return fmt.Errorf("init: create %s: %w", scope.Tilde(store.Dir(dir)), err)
	}

	fmt.Fprintf(out, "%s created. Add global skills with: melon add --global <dep>\n", scope.Tilde(manifestPath))
	if len(harnesses) == 0 {
		fmt.Fprintln(out, "No harnesses selected: global skills will be placed in ~/.agents/skills/.")
	}
	return nil
}

// generateGlobalManifestYAML produces the commented global melon.yaml.
func generateGlobalManifestYAML(harnesses []string) (string, error) {
	harnessBlock := "harnesses: []"
	if len(harnesses) > 0 {
		lines := []string{"harnesses:"}
		for _, h := range harnesses {
			dir, err := agents.GlobalDir(h)
			if err != nil {
				return "", err
			}
			lines = append(lines, fmt.Sprintf("  - %s # %s", h, dir))
		}
		harnessBlock = strings.Join(lines, "\n")
	}

	return fmt.Sprintf(`# Global melon.yaml — skills installed for every session of your harnesses.
# Manage it with the --global (-g) flag, e.g.: melon add -g <dep>
# Skills are cached in ~/.melon/store/ and symlinked into each harness's
# global skills directory.

name: global
version: 0.1.0

description: "Globally installed skills"

# dependencies lists globally installed skills.
# Add with: melon add --global <dep>
dependencies: {}

# harnesses lists the tools that receive every global skill.
# When empty, skills are placed in ~/.agents/skills/.
%s

# policy.allowed_sources optionally restricts where global skills may come from.
# policy:
#   allowed_sources:
#     - github.com/my-org/*

tags: []
`, harnessBlock), nil
}
