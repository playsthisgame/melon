package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/playsthisgame/melon/internal/fetcher"
	"github.com/playsthisgame/melon/internal/lockfile"
	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/internal/store"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

var (
	flagDiffStat    bool
	flagDiffNoColor bool
)

// Raw ANSI color codes for diff lines. Coloring is gated by an explicit bool
// (derived from isTTY() and --no-color) rather than terminal auto-detection, so
// output is deterministic.
const (
	ansiReset = "\x1b[0m"
	ansiAdd   = "\x1b[32m" // green
	ansiDel   = "\x1b[31m" // red
	ansiHunk  = "\x1b[36m" // cyan
)

func runDiff(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// Parse <dep>[@<target>]. Everything before the first "@" is the dep name;
	// everything after is the explicit target version or branch.
	name, target, hasTarget := strings.Cut(args[0], "@")
	target = strings.TrimSpace(target)

	dir, err := resolveProjectDir()
	if err != nil {
		return err
	}

	m, err := manifest.Load(manifest.FindPath(dir))
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}
	constraint, declared := m.Dependencies[name]
	if !declared {
		return fmt.Errorf("diff: %q is not a dependency declared in melon.yaml", name)
	}

	// The "from" side is the currently locked version.
	lockPath := filepath.Join(dir, "melon.lock")
	lf, lfErr := lockfile.Load(lockPath)
	if lfErr != nil {
		return fmt.Errorf("diff: %q is not installed — run `melon install` first", name)
	}
	var locked *lockfile.LockedDep
	for i := range lf.Dependencies {
		if lf.Dependencies[i].Name == name {
			locked = &lf.Dependencies[i]
			break
		}
	}
	if locked == nil {
		return fmt.Errorf("diff: %q is not installed — run `melon install` first", name)
	}

	repoURL, subdir := fetcher.ParseDepName(name)

	// The "to" side: explicit target if given, otherwise latest compatible.
	var resolveArg string
	switch {
	case hasTarget && target != "":
		resolveArg = target
	case isBranchPin(constraint):
		return fmt.Errorf("diff: %q is branch-pinned (%s) — specify an explicit target, e.g. `melon diff %s@1.0.0`", name, constraint, name)
	default:
		resolveArg = constraint
	}

	var toVersion, toTag string
	if err := withSpinner(fmt.Sprintf("Resolving %s…", name), func() error {
		var rErr error
		toVersion, toTag, rErr = resolveVersionFn(repoURL, resolveArg)
		return rErr
	}); err != nil {
		return fmt.Errorf("diff: resolve target for %s: %w", name, err)
	}

	// Same version on both sides — nothing to diff.
	if toVersion == locked.Version {
		fmt.Fprintln(out, "No changes")
		return nil
	}

	fromDep := resolver.ResolvedDep{
		Name:       locked.Name,
		Version:    locked.Version,
		GitTag:     locked.GitTag,
		RepoURL:    locked.RepoURL,
		Subdir:     locked.Subdir,
		Entrypoint: locked.Entrypoint,
		TreeHash:   locked.TreeHash,
	}
	toDep := resolver.ResolvedDep{
		Name:       name,
		Version:    toVersion,
		GitTag:     toTag,
		RepoURL:    repoURL,
		Subdir:     subdir,
		Entrypoint: locked.Entrypoint,
	}

	// Materialize both trees in the .melon/ cache. fetchFn is idempotent — the
	// locked version is usually already present and skipped.
	if err := os.MkdirAll(filepath.Join(dir, store.StoreDir), 0755); err != nil {
		return fmt.Errorf("diff: create store: %w", err)
	}
	fromDir := store.InstalledPath(dir, fromDep)
	toDir := store.InstalledPath(dir, toDep)

	var fromFiles, toFiles []string
	var toHash string
	if err := withSpinner(fmt.Sprintf("Fetching %s@%s…", name, toVersion), func() error {
		fromRes, fErr := fetchFn(fromDep, fromDir)
		if fErr != nil {
			return fmt.Errorf("fetch locked %s: %w", locked.Version, fErr)
		}
		toRes, tErr := fetchFn(toDep, toDir)
		if tErr != nil {
			return fmt.Errorf("fetch %s: %w", toVersion, tErr)
		}
		fromFiles, toFiles, toHash = fromRes.Files, toRes.Files, toRes.TreeHash
		return nil
	}); err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	// Fast path: identical content trees.
	if locked.TreeHash != "" && locked.TreeHash == toHash {
		fmt.Fprintln(out, "No changes")
		return nil
	}

	color := !flagDiffNoColor && isTTY()
	fmt.Fprintf(out, "diff %s  %s → %s\n", name, locked.Version, toVersion)
	return renderTreeDiff(out, fromDir, toDir, fromFiles, toFiles, flagDiffStat, color)
}

// renderTreeDiff prints a file-by-file diff between two version trees. In stat
// mode it prints only per-file +/- counts and a totals line; otherwise it prints
// a unified diff per changed file.
func renderTreeDiff(out io.Writer, fromDir, toDir string, fromFiles, toFiles []string, stat, color bool) error {
	fromSet := make(map[string]struct{}, len(fromFiles))
	for _, f := range fromFiles {
		fromSet[f] = struct{}{}
	}
	toSet := make(map[string]struct{}, len(toFiles))
	for _, f := range toFiles {
		toSet[f] = struct{}{}
	}

	paths := make([]string, 0, len(fromSet)+len(toSet))
	seen := make(map[string]struct{}, len(fromSet)+len(toSet))
	for _, f := range append(append([]string{}, fromFiles...), toFiles...) {
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		paths = append(paths, f)
	}
	sort.Strings(paths)

	var totalAdded, totalRemoved, changedFiles int
	for _, path := range paths {
		_, inFrom := fromSet[path]
		_, inTo := toSet[path]

		var fromContent, toContent []byte
		if inFrom {
			b, err := os.ReadFile(filepath.Join(fromDir, filepath.FromSlash(path)))
			if err != nil {
				return fmt.Errorf("read %s: %w", path, err)
			}
			fromContent = b
		}
		if inTo {
			b, err := os.ReadFile(filepath.Join(toDir, filepath.FromSlash(path)))
			if err != nil {
				return fmt.Errorf("read %s: %w", path, err)
			}
			toContent = b
		}

		if inFrom && inTo && bytes.Equal(fromContent, toContent) {
			continue // unchanged
		}

		changedFiles++

		// Binary files: report the change without rendering hunks.
		if isBinary(fromContent) || isBinary(toContent) {
			if stat {
				fmt.Fprintf(out, "  %s  (binary)\n", path)
			} else {
				fmt.Fprintf(out, "%s: binary file changed\n", path)
			}
			continue
		}

		added, removed := countDiffLines(fromContent, toContent)
		totalAdded += added
		totalRemoved += removed

		if stat {
			fmt.Fprintf(out, "  %s  +%d -%d\n", path, added, removed)
			continue
		}

		body, err := unifiedDiff(fromContent, toContent, path, inFrom, inTo)
		if err != nil {
			return err
		}
		fmt.Fprint(out, colorizeDiff(body, color))
	}

	if stat {
		fmt.Fprintf(out, "%d file(s) changed, %d insertion(s)(+), %d deletion(s)(-)\n",
			changedFiles, totalAdded, totalRemoved)
	}
	return nil
}

// unifiedDiff produces a unified diff for one file. When the file is added
// (inFrom false) or removed (inTo false), the corresponding side is treated as
// /dev/null.
func unifiedDiff(fromContent, toContent []byte, path string, inFrom, inTo bool) (string, error) {
	fromName := "a/" + path
	toName := "b/" + path
	if !inFrom {
		fromName = "/dev/null"
	}
	if !inTo {
		toName = "/dev/null"
	}
	ud := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(fromContent)),
		B:        difflib.SplitLines(string(toContent)),
		FromFile: fromName,
		ToFile:   toName,
		Context:  3,
	}
	return difflib.GetUnifiedDiffString(ud)
}

// countDiffLines returns the number of added and removed lines between two files.
func countDiffLines(fromContent, toContent []byte) (added, removed int) {
	body, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:       difflib.SplitLines(string(fromContent)),
		B:       difflib.SplitLines(string(toContent)),
		Context: 0,
	})
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(body, "\n") {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			continue
		case strings.HasPrefix(line, "+"):
			added++
		case strings.HasPrefix(line, "-"):
			removed++
		}
	}
	return added, removed
}

// colorizeDiff applies ANSI coloring to a unified diff body when color is true.
func colorizeDiff(body string, color bool) string {
	if !color {
		return body
	}
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			// file headers — leave uncolored
		case strings.HasPrefix(line, "@@"):
			lines[i] = ansiHunk + line + ansiReset
		case strings.HasPrefix(line, "+"):
			lines[i] = ansiAdd + line + ansiReset
		case strings.HasPrefix(line, "-"):
			lines[i] = ansiDel + line + ansiReset
		}
	}
	return strings.Join(lines, "\n")
}

// isBinary reports whether b is likely binary content (contains a NUL byte or
// is not valid UTF-8).
func isBinary(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	if bytes.IndexByte(b, 0) >= 0 {
		return true
	}
	return !utf8.Valid(b)
}
