package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/playsthisgame/melon/internal/github"
	"github.com/playsthisgame/melon/internal/index"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <term>",
	Short: "Search for skills in the melon index",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	term := args[0]

	// 1. Try the curated index first.
	var items []searchResultItem
	entries, indexErr := index.Fetch()
	if indexErr == nil {
		matched := index.Search(entries, term)
		for _, e := range matched {
			items = append(items, searchResultItem{
				path:        e.Name,
				author:      e.Author,
				description: e.Description,
				featured:    e.Featured,
			})
		}
	}

	// 2. Fall back to GitHub Topics if the index missed or was unreachable.
	fromTopics := false
	if len(items) == 0 {
		if indexErr != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "warning: could not load curated index (%v)\n", indexErr)
			fmt.Fprintf(cmd.OutOrStdout(), "         (expected: %s)\n", index.IndexURL)
		}
		client := gh.New()
		results, ghErr := client.SearchByTopic(term)
		if ghErr != nil {
			if indexErr != nil {
				// Both sources failed — report both errors.
				return fmt.Errorf("search: index unavailable (%v) and GitHub Topics failed: %w", indexErr, ghErr)
			}
			return fmt.Errorf("search: %w", ghErr)
		}
		for _, r := range results {
			items = append(items, searchResultItem{
				path:        r.Name,
				author:      r.Owner,
				description: r.Description,
			})
		}
		fromTopics = true
	}

	// 3. No results from either source.
	if len(items) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No skills found for %q.\n", term)
		return nil
	}

	if fromTopics {
		fmt.Fprintln(cmd.OutOrStdout(), "No curated results — showing community-tagged skills from GitHub Topics:")
	}

	// 4a. Interactive TTY mode — show the bubbletea list.
	if isTTY() {
		selected, err := runSearchTUI(items)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}
		if selected == "" {
			return nil // user cancelled
		}
		return offerAdd(cmd, selected)
	}

	// 4b. Plain-text non-TTY mode.
	for _, r := range items {
		star := ""
		if r.featured {
			star = "★ "
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s%s\t%s\t%s\n", star, r.path, r.author, r.description)
	}
	return nil
}

// runSearchTUI runs the bubbletea search list and returns the selected path.
func runSearchTUI(items []searchResultItem) (string, error) {
	m := newSearchModel(items)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	return final.(searchModel).selected, nil
}

// offerAdd prints the mln add command for the selected path and prompts to run it.
func offerAdd(cmd *cobra.Command, path string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "\nSelected: %s\n", path)
	fmt.Fprintf(cmd.OutOrStdout(), "Run 'mln add %s'? [y/N]: ", path)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		return runAdd(cmd, []string{path})
	}
	return nil
}
