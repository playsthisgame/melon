package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/playsthisgame/melon/internal/resolver"
)

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	dep := resolver.ResolvedDep{Name: "alice/pdf-skill", Version: "1.2.0"}
	cacheDir := InstalledPath(dir, dep)

	// Create the cache directory.
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Write a file inside to confirm full removal.
	if err := os.WriteFile(filepath.Join(cacheDir, "SKILL.md"), []byte("hi"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := Remove(dir, dep); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Fatalf("expected cache dir to be gone, got: %v", err)
	}
}

func TestRemoveNonExistent(t *testing.T) {
	dir := t.TempDir()
	dep := resolver.ResolvedDep{Name: "alice/pdf-skill", Version: "1.2.0"}
	// Cache directory was never created — Remove should succeed silently.
	if err := Remove(dir, dep); err != nil {
		t.Fatalf("Remove on non-existent path returned error: %v", err)
	}
}
