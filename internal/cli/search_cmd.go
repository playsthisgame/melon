package cli

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/index"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/spf13/cobra"
)

// runInstallFn is the function used to install after manifest updates. Overridable in tests.
var runInstallFn func(*cobra.Command, []string) error = runInstall

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

// offerAddMany prints the selected skills, prompts for confirmation, then adds
// all of them to melon.yaml in one pass before running a single install so that
// parallel fetching can operate on all new deps at once.
func offerAddMany(cmd *cobra.Command, paths []string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "\nSelected skills:\n")
	for _, p := range paths {
		fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", p)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Install %d skill(s)? [Y/n]: ", len(paths))

	reader := bufio.NewReader(cmd.InOrStdin())
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "" && input != "y" && input != "yes" {
		return nil
	}

	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}
	manifestPath := manifest.FindPath(dir)

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	if m.Dependencies == nil {
		m.Dependencies = make(map[string]string)
	}

	// Resolve the latest tag for each selected skill and write them all to
	// melon.yaml before running install, so the parallel fetcher can work on
	// all new deps in a single pass.
	for _, p := range paths {
		name, constraint, hasConstraint := strings.Cut(p, "@")
		if !hasConstraint || constraint == "" {
			repoURL, _ := fetcher.ParseDepName(name)
			var version string
			if err := withSpinner(fmt.Sprintf("Resolving %s…", name), func() error {
				var err error
				version, _, err = fetcher.LatestTag(repoURL)
				return err
			}); err != nil {
				if !errors.Is(err, fetcher.ErrNoSemverTags) {
					fmt.Fprintf(cmd.OutOrStdout(), "warning: could not resolve %s: %v — skipping\n", name, err)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "warning: no semver tags found for %s — using main branch\n", name)
				constraint = "main"
			} else {
				constraint = "^" + version
			}
		}
		if existing, ok := m.Dependencies[name]; ok {
			fmt.Fprintf(cmd.OutOrStdout(), "warning: updating %s from %s to %s\n", name, existing, constraint)
		}
		m.Dependencies[name] = constraint
		fmt.Fprintf(cmd.OutOrStdout(), "Added %s %s to melon.yaml\n", name, constraint)
	}

	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("search: save melon.yaml: %w", err)
	}

	return runInstallFn(cmd, nil)
}
