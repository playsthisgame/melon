package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/playsthisgame/melon/internal/agents"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/spf13/cobra"
)

var (
	flagPending  bool
	flagCheck    bool
	flagListJSON bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&flagPending, "pending", false, "show skills in melon.yaml that are not yet installed")
	listCmd.Flags().BoolVar(&flagCheck, "check", false, "verify that installed skill symlinks exist in all tool directories")
	listCmd.Flags().BoolVar(&flagListJSON, "json", false, "output as JSON")
}

// listJSONOutput is the envelope written to stdout when --json is set.
type listJSONOutput struct {
	Installed []listInstalledEntry `json:"installed"`
	Pending   []string             `json:"pending,omitempty"`
	Check     []listCheckEntry     `json:"check,omitempty"`
}

type listInstalledEntry struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	GitTag     string `json:"git_tag"`
	RepoURL    string `json:"repo_url"`
	Subdir     string `json:"subdir"`
	Entrypoint string `json:"entrypoint"`
	TreeHash   string `json:"tree_hash"`
}

type listCheckEntry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Status string `json:"status"` // "ok" or "missing"
}

func runList(cmd *cobra.Command, args []string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return listErr(cmd, err)
	}

	lockPath := filepath.Join(dir, "melon.lock")
	lock, _ := lockfile.Load(lockPath) // absent lock → empty, handled below

	deps := lock.Dependencies
	sort.Slice(deps, func(i, j int) bool { return deps[i].Name < deps[j].Name })

	if flagListJSON {
		return runListJSON(cmd, dir, deps)
	}

	anyMissing := false

	// --- Installed skills ---
	if len(deps) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No skills installed.")
	} else if flagCheck {
		targetBases := resolveTargetBases(dir)
		for _, dep := range deps {
			skillName := skillNameFromDep(dep.Name)
			for _, base := range targetBases {
				linkPath := filepath.Join(dir, base, skillName)
				rel, _ := filepath.Rel(dir, linkPath)
				if _, serr := os.Stat(linkPath); serr != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  MISSING  %-50s  %s\n", dep.Name+"@"+dep.Version, rel)
					anyMissing = true
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  OK       %-50s  %s\n", dep.Name+"@"+dep.Version, rel)
				}
			}
		}
	} else {
		for _, dep := range deps {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s  %s\n", dep.Name, dep.Version)
		}
	}

	// --- Pending skills ---
	if flagPending {
		manifestPath := manifest.FindPath(dir)
		m, merr := manifest.Load(manifestPath)
		if merr != nil {
			return fmt.Errorf("list --pending: %w", merr)
		}

		installed := make(map[string]struct{}, len(deps))
		for _, dep := range deps {
			installed[dep.Name] = struct{}{}
		}

		var pending []string
		for name := range m.Dependencies {
			if _, ok := installed[name]; !ok {
				pending = append(pending, name)
			}
		}
		sort.Strings(pending)

		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Pending (not installed):")
		if len(pending) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  No pending skills.")
		} else {
			for _, name := range pending {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", name)
			}
		}
	}

	if anyMissing {
		return fmt.Errorf("list --check: one or more skill placements are missing")
	}
	return nil
}

func runListJSON(cmd *cobra.Command, dir string, deps []lockfile.LockedDep) error {
	out := listJSONOutput{Installed: []listInstalledEntry{}}

	for _, dep := range deps {
		out.Installed = append(out.Installed, listInstalledEntry{
			Name:       dep.Name,
			Version:    dep.Version,
			GitTag:     dep.GitTag,
			RepoURL:    dep.RepoURL,
			Subdir:     dep.Subdir,
			Entrypoint: dep.Entrypoint,
			TreeHash:   dep.TreeHash,
		})
	}

	// --pending: include names of deps in manifest but not in lock.
	if flagPending {
		manifestPath := manifest.FindPath(dir)
		m, merr := manifest.Load(manifestPath)
		if merr != nil {
			return listErr(cmd, fmt.Errorf("list --pending: %w", merr))
		}
		installed := make(map[string]struct{}, len(deps))
		for _, dep := range deps {
			installed[dep.Name] = struct{}{}
		}
		var pending []string
		for name := range m.Dependencies {
			if _, ok := installed[name]; !ok {
				pending = append(pending, name)
			}
		}
		sort.Strings(pending)
		out.Pending = pending
	}

	// --check: verify symlinks exist in all tool directories.
	anyMissing := false
	if flagCheck {
		targetBases := resolveTargetBases(dir)
		for _, dep := range deps {
			skillName := skillNameFromDep(dep.Name)
			for _, base := range targetBases {
				linkPath := filepath.Join(dir, base, skillName)
				rel, _ := filepath.Rel(dir, linkPath)
				status := "ok"
				if _, serr := os.Stat(linkPath); serr != nil {
					status = "missing"
					anyMissing = true
				}
				out.Check = append(out.Check, listCheckEntry{
					Name:   dep.Name,
					Path:   rel,
					Status: status,
				})
			}
		}
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return listErr(cmd, err)
	}

	if anyMissing {
		return listErr(cmd, fmt.Errorf("list --check: one or more skill placements are missing"))
	}
	return nil
}

// listErr writes a JSON error to stderr when --json is set, otherwise returns
// the error as-is for cobra to print.
func listErr(cmd *cobra.Command, err error) error {
	if !flagListJSON {
		return err
	}
	fmt.Fprintf(cmd.ErrOrStderr(), `{"error": %q}`+"\n", err.Error())
	cmd.SilenceErrors = true
	return err
}

// resolveTargetBases returns the list of tool directories for the project.
func resolveTargetBases(dir string) []string {
	manifestPath := manifest.FindPath(dir)
	m, merr := manifest.Load(manifestPath)
	var targetBases []string
	if merr == nil && len(m.Outputs) > 0 {
		for base := range m.Outputs {
			targetBases = append(targetBases, base)
		}
		sort.Strings(targetBases)
	} else if merr == nil {
		targetBases, _ = agents.DeriveTargets(m.ToolCompat)
	}
	if len(targetBases) == 0 {
		targetBases = []string{".agents/skills/"}
	}
	return targetBases
}

// skillNameFromDep extracts the skill directory name from a dep path.
func skillNameFromDep(depName string) string {
	if idx := strings.LastIndex(depName, "/"); idx >= 0 {
		return depName[idx+1:]
	}
	return depName
}
