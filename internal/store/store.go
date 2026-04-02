package store

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/playsthisgame/melon/internal/resolver"
)

const StoreDir = ".melon"

// dirName converts a ResolvedDep into the directory name used inside .melon/.
// Slashes in the dep name are replaced with dashes.
// e.g. "github.com/alice/xlsx-skill" at "1.2.0" -> "github.com-alice-xlsx-skill@1.2.0"
func dirName(dep resolver.ResolvedDep) string {
	safeName := strings.ReplaceAll(dep.Name, "/", "-")
	return safeName + "@" + dep.Version
}

// InstalledPath returns the absolute path to dep's installed directory inside .melon/.
func InstalledPath(projectDir string, dep resolver.ResolvedDep) string {
	return filepath.Join(projectDir, StoreDir, dirName(dep))
}

// EntrypointPath returns the absolute path to the dep's entrypoint markdown file.
func EntrypointPath(projectDir string, dep resolver.ResolvedDep) string {
	return filepath.Join(InstalledPath(projectDir, dep), dep.Entrypoint)
}

// List returns all ResolvedDeps currently present in the .melon/ directory.
// Only Name and Version are populated; the caller cross-references melon.lock for
// full dep info.
func List(projectDir string) ([]resolver.ResolvedDep, error) {
	storeDir := filepath.Join(projectDir, StoreDir)
	entries, err := os.ReadDir(storeDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var deps []resolver.ResolvedDep
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		at := strings.LastIndex(name, "@")
		if at < 0 {
			continue
		}
		safeName := name[:at]
		version := name[at+1:]
		// Reverse the safe name back to the dep name (best-effort; "/" were replaced with "-").
		// The caller is expected to cross-reference mln.lock for the real name.
		deps = append(deps, resolver.ResolvedDep{
			Name:    strings.ReplaceAll(safeName, "-", "/"),
			Version: version,
		})
	}
	return deps, nil
}
