package fetcher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	modver "golang.org/x/mod/semver"

	"github.com/playsthisgame/melon/internal/resolver"
	"github.com/playsthisgame/melon/pkg/semver"
)

// FetchResult contains the computed tree hash and file list produced by Fetch.
type FetchResult struct {
	TreeHash string   // "sha256:<hex>"
	Files    []string // sorted relative file paths within the installed directory
}

// ParseDepName splits a GitHub dep name into the repo root URL and subdirectory.
//
//	"github.com/owner/repo"                                    → ("https://github.com/owner/repo", "")
//	"github.com/owner/repo/sub/dir"                            → ("https://github.com/owner/repo", "sub/dir")
//	"github.com/owner/repo/tree/<branch>/sub/dir"              → ("https://github.com/owner/repo", "sub/dir")
func ParseDepName(name string) (repoURL, subdir string) {
	// Strip https:// or http:// prefix so users can paste full GitHub URLs.
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")

	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return "https://" + name, ""
	}
	repoURL = "https://" + strings.Join(parts[:3], "/")
	rest := parts[3:]
	// Strip GitHub web UI path segments: tree/<branch>/ or blob/<branch>/
	if len(rest) >= 2 && (rest[0] == "tree" || rest[0] == "blob") {
		rest = rest[2:]
	}
	if len(rest) > 0 {
		subdir = strings.Join(rest, "/")
	}
	return repoURL, subdir
}

// LatestTag returns the highest semver tag in repoURL, regardless of constraints.
// Returns the bare version (e.g. "1.2.3") and the git tag (e.g. "v1.2.3").
func LatestTag(repoURL string) (version, gitTag string, err error) {
	out, err := exec.Command("git", "ls-remote", "--tags", repoURL).Output()
	if err != nil {
		return "", "", fmt.Errorf("fetcher: ls-remote %s: %w", repoURL, err)
	}

	var candidates []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ref := fields[1]
		if !strings.HasPrefix(ref, "refs/tags/") || strings.HasSuffix(ref, "^{}") {
			continue
		}
		tag := strings.TrimPrefix(ref, "refs/tags/")
		v := tag
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		if !modver.IsValid(v) {
			continue
		}
		candidates = append(candidates, strings.TrimPrefix(tag, "v"))
	}

	if len(candidates) == 0 {
		return "", "", fmt.Errorf("fetcher: no semver tags found in %s", repoURL)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return modver.Compare("v"+candidates[i], "v"+candidates[j]) < 0
	})

	best := candidates[len(candidates)-1]
	return best, "v" + best, nil
}

// LatestMatchingVersion queries the remote repo's git tags via git ls-remote
// and returns the highest version that satisfies constraint.
// Returns the bare version (e.g. "1.2.3") and the git tag/ref (e.g. "v1.2.3").
//
// If constraint is a branch name (e.g. "main", "HEAD") rather than a semver
// constraint, it is resolved to the current HEAD SHA of that branch and used
// directly as the git ref. The returned version is the short SHA.
func LatestMatchingVersion(repoURL, constraint string) (version, gitTag string, err error) {
	// If the constraint looks like a branch/ref (not a semver constraint), resolve it directly.
	if isBranchConstraint(constraint) {
		return resolveBranchRef(repoURL, constraint)
	}

	out, err := exec.Command("git", "ls-remote", "--tags", repoURL).Output()
	if err != nil {
		return "", "", fmt.Errorf("fetcher: ls-remote %s: %w", repoURL, err)
	}

	var candidates []string // bare versions without "v" prefix
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ref := fields[1]
		// Skip peeled tag objects (e.g. "refs/tags/v1.0.0^{}")
		if !strings.HasPrefix(ref, "refs/tags/") || strings.HasSuffix(ref, "^{}") {
			continue
		}
		tag := strings.TrimPrefix(ref, "refs/tags/")
		// Normalise: ensure the version has a "v" prefix for x/mod/semver validation.
		v := tag
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		if !modver.IsValid(v) {
			continue
		}
		candidates = append(candidates, strings.TrimPrefix(tag, "v"))
	}

	if len(candidates) == 0 {
		return "", "", fmt.Errorf("fetcher: no semver tags found in %s — if the repo uses branches instead of tags, use a branch name as the version (e.g. \"main\")", repoURL)
	}

	// Sort ascending by semver so we can walk from highest down.
	sort.Slice(candidates, func(i, j int) bool {
		return modver.Compare("v"+candidates[i], "v"+candidates[j]) < 0
	})

	for i := len(candidates) - 1; i >= 0; i-- {
		v := candidates[i]
		if semver.IsCompatible(constraint, v) {
			return v, "v" + v, nil
		}
	}

	return "", "", fmt.Errorf("fetcher: no version of %s satisfies %q", repoURL, constraint)
}

// isBranchConstraint reports whether constraint is a branch/ref name rather
// than a semver constraint. A constraint is treated as a branch if it contains
// no semver operators (^, ~) and is not purely numeric (X.Y.Z).
func isBranchConstraint(constraint string) bool {
	if strings.HasPrefix(constraint, "^") || strings.HasPrefix(constraint, "~") {
		return false
	}
	// If it looks like a semver version X.Y.Z, it's not a branch.
	v := constraint
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return !modver.IsValid(v)
}

// resolveBranchRef resolves a branch/ref name to its current HEAD SHA.
// Returns the short SHA as version and the full ref as gitTag.
func resolveBranchRef(repoURL, branch string) (version, gitTag string, err error) {
	out, err := exec.Command("git", "ls-remote", repoURL, branch).Output()
	if err != nil {
		return "", "", fmt.Errorf("fetcher: ls-remote %s %s: %w", repoURL, branch, err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			sha := fields[0]
			short := sha
			if len(short) > 12 {
				short = short[:12]
			}
			return short, branch, nil
		}
	}
	return "", "", fmt.Errorf("fetcher: ref %q not found in %s", branch, repoURL)
}

// TreeHash computes a deterministic SHA256 over all files in dir, sorted by
// relative path. Returns "sha256:<hex>" and the sorted relative file list.
func TreeHash(dir string) (hash string, files []string, err error) {
	var paths []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return "", nil, fmt.Errorf("fetcher: walk %s: %w", dir, err)
	}
	sort.Strings(paths)

	h := sha256.New()
	for _, rel := range paths {
		// Include the path in the hash so renaming a file changes the tree hash.
		fmt.Fprintf(h, "%s\n", rel)
		f, err := os.Open(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", nil, fmt.Errorf("fetcher: open %s: %w", rel, err)
		}
		_, copyErr := io.Copy(h, f)
		f.Close()
		if copyErr != nil {
			return "", nil, fmt.Errorf("fetcher: hash %s: %w", rel, copyErr)
		}
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), paths, nil
}

// Fetch installs the skill directory for dep into installDir using a shallow
// git sparse-checkout. If installDir already exists and its tree hash matches
// dep.TreeHash the install is skipped (idempotent).
//
// Returns a FetchResult with the computed tree hash and sorted file list.
func Fetch(dep resolver.ResolvedDep, installDir string) (FetchResult, error) {
	// Idempotency: skip if already installed and hash matches.
	if dep.TreeHash != "" {
		if _, err := os.Stat(installDir); err == nil {
			hash, files, err := TreeHash(installDir)
			if err == nil && hash == dep.TreeHash {
				return FetchResult{TreeHash: hash, Files: files}, nil
			}
		}
	}

	// Remove any stale install before re-fetching.
	if err := os.RemoveAll(installDir); err != nil {
		return FetchResult{}, fmt.Errorf("fetcher: remove stale %s: %w", installDir, err)
	}

	// Create a unique temp dir for the git clone.
	tmpDir, err := os.MkdirTemp("", "mln-fetch-*")
	if err != nil {
		return FetchResult{}, fmt.Errorf("fetcher: mktemp: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Step 2: shallow clone — no checkout so we only fetch the tree, not working files.
	// GitTag may be a semver tag (e.g. "v1.2.3") or a branch name (e.g. "main").
	if err := runGit("", "clone", "--depth=1", "--branch", dep.GitTag,
		"--no-checkout", dep.RepoURL, tmpDir); err != nil {
		return FetchResult{}, fmt.Errorf("fetcher: clone %s@%s: %w", dep.RepoURL, dep.GitTag, err)
	}

	// Step 3: sparse-checkout the subdirectory (or full root).
	if dep.Subdir != "" {
		if err := runGit(tmpDir, "sparse-checkout", "set", dep.Subdir); err != nil {
			return FetchResult{}, fmt.Errorf("fetcher: sparse-checkout %s: %w", dep.Subdir, err)
		}
	}
	if err := runGit(tmpDir, "checkout"); err != nil {
		return FetchResult{}, fmt.Errorf("fetcher: checkout: %w", err)
	}

	// Step 4: determine the source tree root to copy from.
	srcDir := tmpDir
	if dep.Subdir != "" {
		srcDir = filepath.Join(tmpDir, filepath.FromSlash(dep.Subdir))
		if _, err := os.Stat(srcDir); err != nil {
			return FetchResult{}, fmt.Errorf("fetcher: subdir %q not found in %s@%s",
				dep.Subdir, dep.RepoURL, dep.GitTag)
		}
	}

	// Copy the skill directory tree to installDir.
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return FetchResult{}, fmt.Errorf("fetcher: mkdir %s: %w", installDir, err)
	}
	if err := copyDir(srcDir, installDir); err != nil {
		return FetchResult{}, fmt.Errorf("fetcher: copy: %w", err)
	}

	// Step 5: compute and (optionally) verify the tree hash.
	hash, files, err := TreeHash(installDir)
	if err != nil {
		return FetchResult{}, err
	}
	if dep.TreeHash != "" && hash != dep.TreeHash {
		return FetchResult{}, fmt.Errorf("fetcher: tree hash mismatch for %s: got %s, expected %s",
			dep.Name, hash, dep.TreeHash)
	}

	// tmpDir cleaned up by defer.
	return FetchResult{TreeHash: hash, Files: files}, nil
}

// runGit runs a git subcommand in dir (empty string = inherit cwd).
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}

// copyDir recursively copies the contents of src into dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dstPath := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
