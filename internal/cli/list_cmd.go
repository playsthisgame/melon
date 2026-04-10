package cli

import (
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
	flagPending bool
	flagCheck   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&flagPending, "pending", false, "show skills in melon.yaml that are not yet installed")
	listCmd.Flags().BoolVar(&flagCheck, "check", false, "verify that installed skill symlinks exist in all tool directories")
}

func runList(cmd *cobra.Command, args []string) error {
	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}

	lockPath := filepath.Join(dir, "melon.lock")
	lock, _ := lockfile.Load(lockPath) // absent lock → empty, handled below

	anyMissing := false

	// --- Installed skills ---
	deps := lock.Dependencies
	sort.Slice(deps, func(i, j int) bool { return deps[i].Name < deps[j].Name })

	if len(deps) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No skills installed.")
	} else if flagCheck {
		// Derive target directories once.
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

		for _, dep := range deps {
			skillName := dep.Name
			if idx := strings.LastIndex(dep.Name, "/"); idx >= 0 {
				skillName = dep.Name[idx+1:]
			}
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
