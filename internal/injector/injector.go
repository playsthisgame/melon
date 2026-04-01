package injector

import "strings"

// InstalledFile is an in-memory representation of a fetched markdown file,
// used by the compiler.
type InstalledFile struct {
	DepName    string      // e.g. "github.com/user/xlsx-skill"
	Version    string      // e.g. "1.3.1"
	Content    string      // raw markdown content of the entrypoint file
	Directives []Directive // parsed from Content by the compiler
}

// Directive is a structured instruction line parsed from a markdown file.
type Directive struct {
	Key    string // e.g. "response-format"
	Value  string // e.g. "bullet-points"
	Source string // dep_name that defined this directive — used in conflict reports
}

// Assemble concatenates files into a single ordered markdown string.
// Each file's content is prepended with a header comment:
//
//	# skill: <name>@<version>
//
// The order of files in the output matches the order of the input slice,
// which must match the install order from mln.lock.
func Assemble(files []InstalledFile) string {
	// TODO: implement Assemble
	// For each file in files:
	//   1. Write "# skill: <DepName>@<Version>\n\n"
	//   2. Write file.Content
	//   3. Separate entries with a blank line.
	var sb strings.Builder
	_ = sb
	return ""
}
