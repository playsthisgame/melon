package compiler

import "errors"

// ErrNotInstalled is returned when mln.lock does not exist.
// The user should run mln install first.
var ErrNotInstalled = errors.New("compiler: mln.lock not found — run mln install first")

// ErrConflict is returned when directive conflicts are detected.
var ErrConflict = errors.New("compiler: directive conflicts detected")

// CompileOptions controls the behavior of Compile.
type CompileOptions struct {
	ProjectDir string // root of the project (where mln.yaml and mln.lock live)
	DryRun     bool   // if true, print what would be written but do not write files
	CheckOnly  bool   // if true, run conflict detection only, do not write files
}

// Compile runs the full compile pipeline:
//  1. Load mln.lock — return ErrNotInstalled if missing.
//  2. Load each installed markdown file from .mln/ into []InstalledFile.
//  3. Run conflict detection via Detect(). If conflicts found, call Print() and
//     return ErrConflict. Do not write any output files on conflict.
//  4. If clean, group InstalledFiles by their declared output target from mln.yaml outputs block.
//  5. Call Assemble per output target group.
//  6. Write each assembled string to its declared target path on disk (unless DryRun or CheckOnly).
//  7. Print a summary of written files.
func Compile(opts CompileOptions) error {
	// TODO: implement Compile
	return nil
}
