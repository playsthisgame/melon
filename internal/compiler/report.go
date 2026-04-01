package compiler

import "io"

// Print formats conflicts into a human-readable terminal report and writes it
// to w. For each conflict it prints:
//   - the directive key in conflict
//   - which packages define it
//   - each package's conflicting value
//   - a suggested resolution: remove one conflicting dep from mln.yaml,
//     or add a manual override directive in the mln.yaml outputs block.
//
// Uses github.com/fatih/color for colored output.
// The caller is responsible for exiting non-zero after calling Print.
func Print(w io.Writer, conflicts []Conflict) {
	// TODO: implement Print
	// For each conflict:
	//   1. Print the directive key in red/bold.
	//   2. For each instance, print "  <source>: <value>".
	//   3. Print the suggested resolution.
}
