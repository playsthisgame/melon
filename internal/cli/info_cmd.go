package cli

import (
	"fmt"
	"strings"

	gh "github.com/playsthisgame/melon/internal/github"
	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/index"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <github-path>",
	Short: "Show metadata for a skill before installing it",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	path := strings.TrimRight(args[0], "/")

	// Parse the GitHub path into owner, repo, and optional subpath.
	owner, repo, subdir, err := parseGitHubPath(path)
	if err != nil {
		return fmt.Errorf("info: %w", err)
	}

	// Look up the skill across all active indices for description and author.
	var description, author string
	seen := make(map[string]struct{})
	var allEntries []index.Entry
	for _, u := range resolveIndexURLs() {
		entries, err := index.Fetch(u)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if _, dup := seen[e.Name]; !dup {
				seen[e.Name] = struct{}{}
				allEntries = append(allEntries, e)
			}
		}
	}
	if entry := index.Find(allEntries, path); entry != nil {
		description = entry.Description
		author = entry.Author
	}

	client := gh.New()

	// Fall back to GitHub repo about field if not in index.
	if description == "" {
		var metaErr error
		description, metaErr = client.RepoMeta(owner, repo)
		if metaErr != nil {
			return fmt.Errorf("info: %w", metaErr)
		}
	}

	// Fetch tags; fall back to branches if none.
	tags, tagsErr := client.ListTags(owner, repo)
	var versions []string
	versionLabel := "Versions"
	if tagsErr == nil && len(tags) > 0 {
		versions = tags
	} else {
		branches, branchErr := client.ListBranches(owner, repo)
		if branchErr != nil {
			return fmt.Errorf("info: list branches: %w", branchErr)
		}
		versions = branches
		versionLabel = "Branches"
	}

	// Print.
	out := cmd.OutOrStdout()
	fmt.Fprintln(out)
	fmt.Fprintf(out, "  %s\n", titleStyle.Render(path))
	fmt.Fprintln(out)
	if subdir != "" {
		fmt.Fprintf(out, "  %-12s %s\n", "Repository", "github.com/"+owner+"/"+repo)
		fmt.Fprintf(out, "  %-12s %s\n", "Subpath", subdir)
	}
	if author != "" {
		fmt.Fprintf(out, "  %-12s %s\n", "Author", author)
	}
	if description != "" {
		fmt.Fprintf(out, "  %-12s %s\n", "Description", description)
	}
	if len(versions) > 0 {
		fmt.Fprintf(out, "  %-12s %s\n", "Latest", versions[0])
		if len(versions) > 1 {
			shown := versions
			if len(shown) > 8 {
				shown = shown[:8]
			}
			fmt.Fprintf(out, "  %-12s %s\n", versionLabel, strings.Join(shown, ", "))
		}
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "  Run 'mln add %s' to install.\n\n", path)

	return nil
}

// parseGitHubPath extracts owner, repo, and subdir from a GitHub dep path.
// Accepts formats like:
//
//	github.com/owner/repo
//	github.com/owner/repo/sub/path
func parseGitHubPath(path string) (owner, repo, subdir string, err error) {
	repoURL, sub := fetcher.ParseDepName(path)
	// repoURL is "https://github.com/owner/repo"
	trimmed := strings.TrimPrefix(repoURL, "https://github.com/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid GitHub path %q: expected github.com/owner/repo", path)
	}
	return parts[0], parts[1], sub, nil
}
