package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FindPath returns the path to the manifest file in dir. It prefers
// melon.yaml but falls back to melon.yml for backwards compatibility.
// When neither file exists it returns the canonical melon.yaml path so
// callers receive a clear "file not found" error on Load.
func FindPath(dir string) string {
	yaml := filepath.Join(dir, "melon.yaml")
	if _, err := os.Stat(yaml); err == nil {
		return yaml
	}
	yml := filepath.Join(dir, "melon.yml")
	if _, err := os.Stat(yml); err == nil {
		return yml
	}
	return yaml
}

// Load reads and parses a melon.yaml file at the given path.
func Load(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("manifest: read %s: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("manifest: parse %s: %w", path, err)
	}
	return m, nil
}

// Save serializes m to YAML and writes it to path.
func Save(m Manifest, path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("manifest: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("manifest: write %s: %w", path, err)
	}
	return nil
}
