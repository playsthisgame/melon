package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

	// TODO: Fall back to GitHub Topics when the community grows. Disabled for now
	// because topic search returns repos rather than skill subdirectories — a
	// multi-skill repo like melon-index would surface as the repo root instead of
	// its individual skills. Revisit once a skill-manifest convention is in place.
	//
	// if len(items) == 0 {
	// 	if indexErr != nil {
	// 		fmt.Fprintf(cmd.OutOrStdout(), "warning: could not load curated index (%v)\n", indexErr)
	// 		fmt.Fprintf(cmd.OutOrStdout(), "         (expected: %s)\n", index.IndexURL)
	// 	}
	// 	client := gh.New()
	// 	results, ghErr := client.SearchByTopic(term)
	// 	if ghErr != nil {
	// 		if indexErr != nil {
	// 			return fmt.Errorf("search: index unavailable (%v) and GitHub Topics failed: %w", indexErr, ghErr)
	// 		}
	// 		return fmt.Errorf("search: %w", ghErr)
	// 	}
	// 	for _, r := range results {
	// 		items = append(items, searchResultItem{
	// 			path:        r.Name,
	// 			author:      r.Owner,
	// 			description: r.Description,
	// 		})
	// 	}
	// }

	// 2. No results from the curated index.
	if len(items) == 0 {
		if indexErr != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "warning: could not load curated index (%v)\n", indexErr)
			fmt.Fprintf(cmd.OutOrStdout(), "         (expected: %s)\n", index.IndexURL)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "No skills found for %q.\n", term)
		return nil
	}

	// 4a. Interactive TTY mode — show the bubbletea list.
	if isTTY() {
		selected, err := runSearchTUI(items)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}
		if len(selected) == 0 {
			return nil // user cancelled or made no selection
		}
		return offerAddMany(cmd, selected)
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

// runSearchTUI runs the bubbletea search list and returns the selected paths.
func runSearchTUI(items []searchResultItem) ([]string, error) {
	m := newSearchModel(items)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	return final.(searchModel).selected, nil
}

// offerAddMany prints the selected skills and prompts the user to install them all.
func offerAddMany(cmd *cobra.Command, paths []string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "\nSelected skills:\n")
	for _, p := range paths {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", p)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Install %d skill(s)? [y/N]: ", len(paths))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		return nil
	}

	for _, p := range paths {
		if err := runAdd(cmd, []string{p}); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "error installing %s: %v\n", p, err)
		}
	}
	return nil
}
