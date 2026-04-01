package fetcher

import (
	"github.com/playsthisgame/melon/internal/resolver"
)

// Fetch clones the GitHub repo at the resolved tag into installDir,
// verifies the SHA256 of the entrypoint markdown file against the
// checksum stored in dep, and skips the fetch if already present and
// checksum matches (idempotent).
//
// Uses os/exec git — no GitHub API.
func Fetch(dep resolver.ResolvedDep, installDir string) error {
	// TODO: implement Fetch
	// 1. Compute the expected install path via store.InstalledPath.
	// 2. If the path exists and checksum matches dep.Checksum, return nil (already installed).
	// 3. git clone dep.RepoURL into a temp dir.
	// 4. git checkout dep.GitTag.
	// 5. Copy dep.Entrypoint to installDir.
	// 6. Compute SHA256 of the copied file and compare to dep.Checksum.
	//    Return an error if they don't match.
	return nil
}
