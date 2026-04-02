package manifest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a melon.yml file at the given path.
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
