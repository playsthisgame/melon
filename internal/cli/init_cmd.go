package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/playsthisgame/melon/internal/agents"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/spf13/cobra"
)

var flagYes bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a new melon.yaml and create the .melon/ store directory",
	Long: `Interactively create a melon.yaml in the current directory and initialize
the .melon/ package store. Does not run install.

Use --yes to accept all defaults without prompts (useful for scripting).`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&flagYes, "yes", false, "accept all defaults without interactive prompts")
}

func runInit(cmd *cobra.Command, args []string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}

	manifestPath := manifest.FindPath(dir)
	mlnDir := filepath.Join(dir, ".melon")

	// Overwrite protection.
	if _, err := os.Stat(manifestPath); err == nil {
		if flagYes {
			fmt.Fprintf(cmd.OutOrStdout(), "melon.yaml already exists — overwriting (--yes)\n")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "melon.yaml already exists. Overwrite? [y/N] ")
			answer := readLine(cmd.InOrStdin())
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}
		}
	}

	// Collect values via bubbletea form (TTY) or plain prompts (non-TTY / --yes).
	defaultName := filepath.Base(dir)
	var name, description string
	var agentNames []string

	var vendor bool
	if flagYes || !isTTY() {
		name = prompt(cmd, "Project name?", defaultName)
		description = prompt(cmd, "Short description?", "")
		agentNames = promptMultiChoice(cmd, "AI tools?", agents.KnownAgents(), []string{})
		vendor = promptVendor(cmd)
	} else {
		model := newInitModel(defaultName)
		p := tea.NewProgram(model)
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("init: %w", err)
		}
		result := finalModel.(initModel).result
		if finalModel.(initModel).quitting && result.name == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}
		name = result.name
		description = result.description
		agentNames = result.agentNames
		vendor = result.vendor
	}

	// Write melon.yaml with inline comments.
	content := generateManifestYAML(name, description, agentNames, vendor)
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("init: write melon.yaml: %w", err)
	}

	// Create .melon/ store directory.
	if err := os.MkdirAll(mlnDir, 0755); err != nil {
		return fmt.Errorf("init: create .melon/: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "melon.yaml created. Add dependencies with: melon add <dep>")
	return nil
}

// resolveProjectDir returns --dir if set, otherwise the current working directory.
func resolveProjectDir() (string, error) {
	if flagDir != "" {
		return filepath.Abs(flagDir)
	}
	return os.Getwd()
}

// prompt prints a question and reads a line from stdin.
// In --yes mode it returns the default immediately.
func prompt(cmd *cobra.Command, question, defaultVal string) string {
	if flagYes {
		return defaultVal
	}
	if defaultVal != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] ", question, defaultVal)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "%s ", question)
	}
	line := strings.TrimSpace(readLine(cmd.InOrStdin()))
	if line == "" {
		return defaultVal
	}
	return line
}

// promptMultiChoice prints a numbered list of options and reads a
// comma-or-space-separated selection. Returns matched options in the order
// they appear in options. In --yes mode returns defaults.
func promptMultiChoice(cmd *cobra.Command, question string, options []string, defaults []string) []string {
	if flagYes {
		return defaults
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s (space or comma-separated numbers, blank=default)\n", question)
	for i, o := range options {
		fmt.Fprintf(cmd.OutOrStdout(), "  %2d) %s\n", i+1, o)
	}
	defaultNums := make([]string, 0, len(defaults))
	for _, d := range defaults {
		for i, o := range options {
			if strings.EqualFold(o, d) {
				defaultNums = append(defaultNums, fmt.Sprintf("%d", i+1))
			}
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Selection [%s]: ", strings.Join(defaultNums, ","))
	line := strings.TrimSpace(readLine(cmd.InOrStdin()))
	if line == "" {
		return defaults
	}
	// Split on commas and spaces.
	raw := strings.FieldsFunc(line, func(r rune) bool { return r == ',' || r == ' ' })
	selected := make([]string, 0, len(raw))
	for _, tok := range raw {
		tok = strings.TrimSpace(tok)
		// Accept number references.
		matched := false
		for i, o := range options {
			if tok == fmt.Sprintf("%d", i+1) || strings.EqualFold(tok, o) {
				selected = append(selected, o)
				matched = true
				break
			}
		}
		_ = matched
	}
	return selected
}

func readLine(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// promptVendor asks whether the user wants to vendor skills in git.
// Default is yes (vendor = true); only "n"/"N" opts out.
// In --yes mode it returns true immediately.
func promptVendor(cmd *cobra.Command) bool {
	if flagYes {
		return true
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Vendor skills in git? Skills will be committed to your repo; disable to auto-manage .gitignore instead. [Y/n] ")
	answer := strings.TrimSpace(readLine(cmd.InOrStdin()))
	return !strings.EqualFold(answer, "n")
}

// generateManifestYAML produces a fully commented melon.yaml string.
// The outputs block is intentionally omitted — mln install derives output paths
// automatically from tool_compat using the agent_directory_conventions table.
// Users can add an explicit outputs block to override the derived paths.
// When vendor is false, a vendor: false line is emitted.
func generateManifestYAML(name, description string, agentNames []string, vendor bool) string {
	escapedDesc := strings.ReplaceAll(description, `"`, `\"`)

	var toolCompatBlock string
	if len(agentNames) == 0 {
		toolCompatBlock = "tool_compat: []"
	} else {
		lines := make([]string, 0, len(agentNames)+1)
		lines = append(lines, "tool_compat:")
		for _, a := range agentNames {
			lines = append(lines, "  - "+a)
		}
		toolCompatBlock = strings.Join(lines, "\n")
	}

	vendorBlock := ""
	if !vendor {
		vendorBlock = "\n# vendor: false tells melon to manage .gitignore for its cache and skill symlinks.\nvendor: false"
	}

	return fmt.Sprintf(`# melon.yaml — melon package manifest
# Edit this file to add dependencies, then run: melon install

name: %s
version: %s

description: "%s"

# dependencies lists skills this project depends on.
# Keys are GitHub shorthand (owner/repo) or monorepo paths (owner/repo/path/to/skill).
# Values are semver constraints (^, ~, or exact) or a branch name (e.g. "main").
# Add with: melon add <dep>
# Example:
#   anthropics/skills/skills/skill-creator: "main"
#   alice/pdf-skill: "^1.2.0"
dependencies: {}

# tool_compat drives where melon install places skill directories.
# When empty, skills are placed in .agents/skills/ by default.
# Melon uses the known directory convention for each tool automatically:
#   claude-code    -> .claude/skills/
#   cursor         -> .agents/skills/
#   windsurf       -> .windsurf/skills/
#   roo            -> .roo/skills/
#   (see README for the full table)
%s

# outputs is optional. Declare it only when you need non-standard placement.
# If omitted, paths are derived from tool_compat (or .agents/skills/ if empty).
# outputs:
#   .claude/skills/: "*"
#   .windsurf/skills/: "alice/pdf-skill"
%s

tags: []
`, name, "0.1.0", escapedDesc, toolCompatBlock, vendorBlock)
}
