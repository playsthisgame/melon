package compiler

import "github.com/playsthisgame/melon/internal/injector"

// Conflict represents two or more deps that define the same directive key
// with different values.
type Conflict struct {
	Key       string               // the conflicting directive key
	Instances []injector.Directive // all definitions of this key across installed files
}

// Detect parses @directive lines from each InstalledFile and returns a
// []Conflict for every directive key defined with more than one distinct value.
//
// Only lines matching "@directive:<key>: <value>" are checked.
// Plain prose is NOT conflict-checked in MVP.
// An empty return slice means the install is clean and safe to assemble.
func Detect(files []injector.InstalledFile) []Conflict {
	// TODO: implement Detect
	// 1. For each file, scan Content line by line.
	// 2. For lines beginning with "@directive:", parse key and value.
	//    Format: "@directive:<key>: <value>"
	// 3. Collect all (key, value, source) tuples into a map[key][]Directive.
	// 4. For each key, if there are two or more distinct values, add a Conflict.
	return nil
}
