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
// Points to runInstallCore — policy is already checked for the specific dep being added.
var runInstallFn func(*cobra.Command, []string) error = runInstallCore

var searchCmd = &cobra.Command{
	Use:   "search <term>",
	Short: "Search for skills in the melon index",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	term := args[0]

	// 1. Resolve which index URLs to query based on optional melon.yaml config.
	urls := resolveIndexURLs()

	// 2. Fetch and merge entries across all active index URLs.
	// Custom index results come first; public index entries with duplicate names
	// are suppressed.
	seen := make(map[string]struct{})
	var allEntries []index.Entry
	var lastErr error
	for _, u := range urls {
		entries, err := index.Fetch(u)
		if err != nil {
			lastErr = err
			continue
		}
		for _, e := range entries {
			if _, dup := seen[e.Name]; !dup {
				seen[e.Name] = struct{}{}
				allEntries = append(allEntries, e)
			}
		}
	}

	// 3. Search across merged entries.
	var items []searchResultItem
	if len(allEntries) > 0 {
		matched := index.Search(allEntries, term)
		for _, e := range matched {
			items = append(items, searchResultItem{
				path:        e.Name,
				author:      e.Author,
				description: e.Description,
				featured:    e.Featured,
				pack:        e.IsPack(),
			})
		}
	}

	// 4. No results.
	if len(items) == 0 {
		if lastErr != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "warning: could not load curated index (%v)\n", lastErr)
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
		return offerAddMany(cmd, selected, packIndex(allEntries))
	}

	// 4b. Plain-text non-TTY mode.
	for _, r := range items {
		star := ""
		if r.featured {
			star = "★ "
		}
		path := r.path
		if r.pack {
			path += " [pack]"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s%s\t%s\t%s\n", star, path, r.author, r.description)
	}
	return nil
}

// resolveIndexURLs returns the ordered list of index URLs to query, based on
// the optional index block in melon.yaml. If no manifest is found or no index
// block is configured, it returns just the default public index.
func resolveIndexURLs() []string {
	dir, err := resolveProjectDir()
	if err != nil {
		return []string{index.DefaultIndexURL}
	}
	m, err := manifest.Load(manifest.FindPath(dir))
	if err != nil || m.Index == nil || len(m.Index.URLs) == 0 {
		return []string{index.DefaultIndexURL}
	}
	candidates := m.Index.URLs
	// Include public index by default (PublicIndex == nil or true)
	if m.Index.PublicIndex == nil || *m.Index.PublicIndex {
		candidates = append(candidates, index.DefaultIndexURL)
	}
	// Normalize GitHub paths/URLs to raw content URLs
	normalized := make([]string, len(candidates))
	for i, u := range candidates {
		normalized[i] = normalizeIndexURL(u)
	}
	return uniqueURLs(normalized)
}

// normalizeIndexURL converts a GitHub path or web URL to a raw content URL.
// Supports:
//   - github.com/owner/repo/path/to/index.yaml
//   - github.com/owner/repo/tree/main/path/to/index.yaml
//   - https://github.com/owner/repo/blob/main/path/to/index.yaml
//   - https://raw.githubusercontent.com/owner/repo/main/path/to/index.yaml (returned as-is)
func normalizeIndexURL(u string) string {
	// Already a raw GitHub URL, return as-is
	if strings.Contains(u, "raw.githubusercontent.com") {
		return u
	}

	// If it doesn't look like a GitHub path, return as-is (might be a different URL)
	if !strings.Contains(u, "github.com") {
		return u
	}

	// Strip https:// or http:// prefix
	u = strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")

	parts := strings.Split(u, "/")
	if len(parts) < 3 {
		return u // Invalid format, return as-is
	}

	owner := parts[1]
	repo := parts[2]
	branch := "main" // default branch
	rest := parts[3:]

	// Strip tree/<branch>/ or blob/<branch>/ segments
	if len(rest) >= 2 && (rest[0] == "tree" || rest[0] == "blob") {
		branch = rest[1]
		rest = rest[2:]
	}

	path := strings.Join(rest, "/")
	if path == "" {
		path = "index.yaml" // Default filename
	}

	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, path)
}

// uniqueURLs returns urls with duplicates removed, preserving first-occurrence order.
func uniqueURLs(urls []string) []string {
	seen := make(map[string]struct{}, len(urls))
	out := make([]string, 0, len(urls))
	for _, u := range urls {
		if _, ok := seen[u]; !ok {
			seen[u] = struct{}{}
			out = append(out, u)
		}
	}
	return out
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
// packs reports which index names are packs; nested packs are followed through it.
func offerAddMany(cmd *cobra.Command, paths []string, packs map[string]bool) error {
	out := cmd.OutOrStdout()

	// For packs, look up how many skills they install so the prompt does not
	// count a whole pack as one skill.
	packSkills := make(map[string]int)     // pack path -> skills it installs (-1 if unknown)
	allPackSkills := make(map[string]bool) // union across selected packs
	var plain int
	for _, p := range paths {
		name, _, _ := strings.Cut(p, "@")
		if !packs[name] {
			plain++
			continue
		}
		var skills map[string]bool
		err := withSpinner(fmt.Sprintf("Reading %s…", name), func() error {
			var err error
			skills, err = packSkillSet(name, packs)
			return err
		})
		if err != nil {
			packSkills[p] = -1
			continue
		}
		packSkills[p] = len(skills)
		for s := range skills {
			allPackSkills[s] = true
		}
	}

	fmt.Fprintf(out, "\nSelected:\n")
	for _, p := range paths {
		n, isPack := packSkills[p]
		switch {
		case !isPack:
			fmt.Fprintf(out, "  - %s\n", p)
		case n < 0:
			fmt.Fprintf(out, "  - %s %s\n", p, packLabel.Render("[pack]"))
		default:
			fmt.Fprintf(out, "  - %s %s\n", p, packTag(n))
		}
	}
	fmt.Fprintf(out, "Install %s? [Y/n]: ", installSummary(len(packSkills), len(allPackSkills), plain, packSkills))

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
	if err := checkSourcePolicy(m, paths); err != nil {
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
				version, _, err = latestTagFn(repoURL)
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

// latestTagFn is overridable in tests.
var latestTagFn = fetcher.LatestTag

// packIndex returns the set of index entry names that are packs.
func packIndex(entries []index.Entry) map[string]bool {
	packs := make(map[string]bool)
	for _, e := range entries {
		if e.IsPack() {
			packs[e.Name] = true
		}
	}
	return packs
}

// packSkillSet returns the distinct skills a pack installs, following nested
// packs that appear in packs. Each pack's manifest is read at its latest semver
// tag (or the default branch when it has none). This is a preview for the
// confirmation prompt; the install itself does full resolution.
func packSkillSet(name string, packs map[string]bool) (map[string]bool, error) {
	skills := make(map[string]bool)
	visited := make(map[string]bool)
	var walk func(n string) error
	walk = func(n string) error {
		if visited[n] {
			return nil
		}
		visited[n] = true
		repoURL, subdir := fetcher.ParseDepName(n)
		ref := "HEAD"
		if _, tag, err := latestTagFn(repoURL); err == nil {
			ref = tag
		}
		m, err := fetchManifestFn(repoURL, ref, subdir)
		if err != nil {
			return err
		}
		for dep := range m.Dependencies {
			if packs[dep] {
				if err := walk(dep); err != nil {
					return err
				}
			} else {
				skills[dep] = true
			}
		}
		return nil
	}
	return skills, walk(name)
}

// installSummary renders the confirmation question, e.g.
// "1 pack (7 skills) and 2 skills". unique is the number of distinct skills
// across all selected packs; packs whose contents are unknown are left out of
// the skill count.
func installSummary(nPacks, unique, plain int, packSkills map[string]int) string {
	var parts []string
	if nPacks > 0 {
		known := true
		for _, n := range packSkills {
			if n < 0 {
				known = false
			}
		}
		p := count(nPacks, "pack")
		if known {
			p += " (" + count(unique, "skill") + ")"
		}
		parts = append(parts, p)
	}
	if plain > 0 {
		parts = append(parts, count(plain, "skill"))
	}
	return strings.Join(parts, " and ")
}

// count renders "1 skill" / "3 skills" using the shared plural helper.
func count(n int, word string) string {
	return fmt.Sprintf("%d %s", n, plural(n, word, word+"s"))
}
