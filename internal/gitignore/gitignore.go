// Package gitignore provides helpers for maintaining .gitignore entries
// managed by melon. All functions are idempotent and safe to call repeatedly.
package gitignore

import (
	"os"
	"strings"
)

const managedHeader = "# melon managed — do not edit this block"

// EnsureEntries reads the .gitignore at path (creating it if absent), appends
// any entries from the provided list that are not already present, and writes
// the result back. It returns the slice of entries that were newly added.
func EnsureEntries(path string, entries []string) (added []string, err error) {
	existing, err := readLines(path)
	if err != nil {
		return nil, err
	}

	present := make(map[string]bool, len(existing))
	for _, line := range existing {
		present[strings.TrimSpace(line)] = true
	}

	var toAdd []string
	for _, e := range entries {
		if !present[strings.TrimSpace(e)] {
			toAdd = append(toAdd, e)
		}
	}
	if len(toAdd) == 0 {
		return nil, nil
	}

	// If the managed header is not already present, prepend it to the new block.
	lines := existing
	if !present[managedHeader] {
		lines = append(lines, "")
		lines = append(lines, managedHeader)
	}
	lines = append(lines, toAdd...)

	return toAdd, writeLines(path, lines)
}

// RemoveEntries removes lines matching any entry in entries from the .gitignore
// at path. Lines are matched by trimmed value. Missing file is silently ignored.
func RemoveEntries(path string, entries []string) error {
	existing, err := readLines(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	remove := make(map[string]bool, len(entries))
	for _, e := range entries {
		remove[strings.TrimSpace(e)] = true
	}

	var kept []string
	for _, line := range existing {
		if !remove[strings.TrimSpace(line)] {
			kept = append(kept, line)
		}
	}

	if len(kept) == len(existing) {
		return nil // nothing changed
	}

	return writeLines(path, kept)
}

// ContainsEntry reports whether path contains a line matching entry (trimmed).
// Returns false if the file does not exist.
func ContainsEntry(path, entry string) (bool, error) {
	lines, err := readLines(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	target := strings.TrimSpace(entry)
	for _, line := range lines {
		if strings.TrimSpace(line) == target {
			return true, nil
		}
	}
	return false, nil
}

// readLines reads a file and splits it into lines. Returns an empty slice if
// the file does not exist.
func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	content := string(data)
	// Normalise line endings.
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	// Trim trailing empty line introduced by Split on a newline-terminated file.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

// writeLines writes lines to path, joining with newlines and adding a trailing newline.
func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}
