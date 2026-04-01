// Package semver provides thin helpers for semver constraint checking,
// wrapping golang.org/x/mod/semver and implementing caret (^) and tilde (~)
// range semantics as used in mln.yaml dependency constraints.
package semver

import (
	"strings"

	"golang.org/x/mod/semver"
)

// canonical ensures v has the "v" prefix required by golang.org/x/mod/semver.
func canonical(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// IsCompatible reports whether version satisfies constraint.
//
// Supported constraint prefixes:
//   - "^X.Y.Z"  caret — compatible with X.Y.Z; allows patch and minor bumps
//     within the same major version (^1.2.3 matches >=1.2.3, <2.0.0).
//     Special case: when major is 0, only patch bumps are allowed
//     (^0.2.3 matches >=0.2.3, <0.3.0).
//   - "~X.Y.Z"  tilde — allows patch bumps only (~1.2.3 matches >=1.2.3, <1.3.0).
//   - "X.Y.Z"   exact match (no prefix).
//
// Both constraint and version may omit the leading "v".
func IsCompatible(constraint, version string) bool {
	v := canonical(version)
	if !semver.IsValid(v) {
		return false
	}

	switch {
	case strings.HasPrefix(constraint, "^"):
		return caretCompatible(strings.TrimPrefix(constraint, "^"), v)
	case strings.HasPrefix(constraint, "~"):
		return tildeCompatible(strings.TrimPrefix(constraint, "~"), v)
	default:
		// Exact match.
		return semver.Compare(v, canonical(constraint)) == 0
	}
}

// caretCompatible implements ^ semantics.
func caretCompatible(base, v string) bool {
	b := canonical(base)
	if !semver.IsValid(b) {
		return false
	}
	// v must be >= base
	if semver.Compare(v, b) < 0 {
		return false
	}
	major := semver.Major(b) // e.g. "v1"
	if major == "v0" {
		// ^0.Y.Z — only patch bumps allowed; next minor is incompatible.
		// Upper bound: next minor version.
		minor := minorVersion(b) // e.g. "v0.2"
		upper := bumpMinor(b)
		_ = minor
		return semver.Compare(v, canonical(upper)) < 0
	}
	// ^X.Y.Z where X>0 — upper bound is next major.
	upper := bumpMajor(b)
	return semver.Compare(v, canonical(upper)) < 0
}

// tildeCompatible implements ~ semantics (patch bumps only).
func tildeCompatible(base, v string) bool {
	b := canonical(base)
	if !semver.IsValid(b) {
		return false
	}
	if semver.Compare(v, b) < 0 {
		return false
	}
	upper := bumpMinor(b)
	return semver.Compare(v, canonical(upper)) < 0
}

// minorVersion returns the "vMAJOR.MINOR" prefix of a semver string.
func minorVersion(v string) string {
	// v is already canonical e.g. "v1.2.3"
	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(parts) < 2 {
		return v
	}
	return "v" + parts[0] + "." + parts[1]
}

// bumpMinor returns the version string with the minor component incremented
// and patch reset to 0, e.g. "v1.2.3" -> "1.3.0".
func bumpMinor(v string) string {
	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(parts) < 3 {
		return v
	}
	minor := atoi(parts[1]) + 1
	return itoa(atoi(parts[0])) + "." + itoa(minor) + ".0"
}

// bumpMajor returns the version string with the major component incremented
// and minor/patch reset to 0, e.g. "v1.2.3" -> "2.0.0".
func bumpMajor(v string) string {
	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(parts) < 1 {
		return v
	}
	major := atoi(parts[0]) + 1
	return itoa(major) + ".0.0"
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
