package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/pkg/semver"
	"github.com/spf13/cobra"
)

// outdatedRow holds resolved version info for one dep.
type outdatedRow struct {
	name             string
	constraint       string
	locked           string // "(not installed)" when absent from lock
	latestCompatible string
	absoluteLatest   string // non-empty only when outside constraint
}

func runOutdated(cmd *cobra.Command, _ []string) error {
	outdatedFound, err := checkOutdated(cmd)
	if err != nil {
		return err
	}
	if outdatedFound {
		os.Exit(1)
	}
	return nil
}

// checkOutdated contains the testable core logic. Returns true if any dep has
// a newer compatible version available.
func checkOutdated(cmd *cobra.Command) (bool, error) {
	dir, err := resolveProjectDir()
	if err != nil {
		return false, err
	}

	manifestPath := manifest.FindPath(dir)
	lockPath := filepath.Join(dir, "melon.lock")

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return false, fmt.Errorf("outdated: %w", err)
	}
	if len(m.Dependencies) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No dependencies declared in melon.yaml.")
		return false, nil
	}

	// Load lock file; treat absence as all-unlocked.
	lockedVersions := map[string]string{}
	if lf, lfErr := lockfile.Load(lockPath); lfErr == nil {
		for _, dep := range lf.Dependencies {
			lockedVersions[dep.Name] = dep.Version
		}
	}

	// Separate semver-constrained from branch-pinned deps; stable sort for output.
	type depEntry struct{ name, constraint string }
	var semverDeps []depEntry
	var branchPinned []string

	names := make([]string, 0, len(m.Dependencies))
	for name := range m.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		constraint := m.Dependencies[name]
		if isBranchPin(constraint) {
			branchPinned = append(branchPinned, name)
		} else {
			semverDeps = append(semverDeps, depEntry{name, constraint})
		}
	}

	// Concurrently resolve latest compatible + absolute latest for each dep.
	rows := make([]outdatedRow, len(semverDeps))
	var mu sync.Mutex
	var firstErr error
	var wg sync.WaitGroup

	for i, dep := range semverDeps {
		wg.Add(1)
		go func(i int, name, constraint string) {
			defer wg.Done()
			repoURL, _ := fetcher.ParseDepName(name)

			latest, _, resolveErr := fetcher.LatestMatchingVersion(repoURL, constraint)
			if resolveErr != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("resolve %s: %w", name, resolveErr)
				}
				mu.Unlock()
				return
			}

			var absLatest string
			if abs, _, absErr := fetcher.LatestTag(repoURL); absErr == nil {
				if !semver.IsCompatible(constraint, abs) {
					absLatest = abs
				}
			}

			locked := lockedVersions[name]
			if locked == "" {
				locked = "(not installed)"
			}

			mu.Lock()
			rows[i] = outdatedRow{
				name:             name,
				constraint:       constraint,
				locked:           locked,
				latestCompatible: latest,
				absoluteLatest:   absLatest,
			}
			mu.Unlock()
		}(i, dep.name, dep.constraint)
	}

	if spinErr := withSpinner("Checking for updates…", func() error {
		wg.Wait()
		return firstErr
	}); spinErr != nil {
		return false, fmt.Errorf("outdated: %w", spinErr)
	}

	// Print branch-pinned skip note before table.
	if len(branchPinned) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "note: skipped %d branch-pinned dep(s): %s\n",
			len(branchPinned), strings.Join(branchPinned, ", "))
	}

	// Filter to outdated rows only.
	var outdated []outdatedRow
	for _, r := range rows {
		if r.locked != r.latestCompatible {
			outdated = append(outdated, r)
		}
	}

	if len(outdated) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "All skills are up to date.")
		return false, nil
	}

	// Print formatted table.
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Dep\tCurrent\tLatest\tAbsolute Latest")
	fmt.Fprintln(w, "---\t-------\t------\t---------------")
	for _, r := range outdated {
		absCol := ""
		if r.absoluteLatest != "" {
			absCol = r.absoluteLatest + " ↑"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.name, r.locked, r.latestCompatible, absCol)
	}
	w.Flush()

	return true, nil
}
