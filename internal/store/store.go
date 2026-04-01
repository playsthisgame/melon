package store

import (
	"strings"

	"github.com/playsthisgame/melon/internal/resolver"
)

const StoreDir = ".mln"

// dirName converts a ResolvedDep into the directory name used inside .mln/.
// Format: <name-with-slashes-replaced-by-dashes>@<version>
// e.g. "github.com/alice/xlsx-skill" at "1.2.0" -> "github.com-alice-xlsx-skill@1.2.0"
func dirName(dep resolver.ResolvedDep) string {
	safeName := strings.ReplaceAll(dep.Name, "/", "-")
	return safeName + "@" + dep.Version
}

// InstalledPath returns the absolute path to dep's installed directory
// relative to projectDir.
func InstalledPath(projectDir string, dep resolver.ResolvedDep) string {
	// TODO: implement InstalledPath
	// Return filepath.Join(projectDir, StoreDir, dirName(dep))
	return ""
}

// List returns all ResolvedDeps currently present in the .mln/ directory
// under projectDir by reading directory names and parsing them.
func List(projectDir string) ([]resolver.ResolvedDep, error) {
	// TODO: implement List
	// 1. Read projectDir/.mln/ entries.
	// 2. For each directory entry matching "<name>@<version>", reconstruct a
	//    ResolvedDep with Name and Version populated (other fields will be
	//    populated from mln.lock by the caller).
	return nil, nil
}
