package manifest

// Manifest is the parsed representation of a project's melon.yaml.
type Manifest struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description,omitempty"`
	Entrypoint   string            `yaml:"entrypoint"`
	Dependencies map[string]string `yaml:"dependencies,omitempty"`
	// Outputs maps target filenames to glob patterns of dep names to include.
	// Example: "CLAUDE.md": "*"  or  ".claude/SKILL.md": "github.com/alice/*"
	Outputs    map[string]string `yaml:"outputs,omitempty"`
	Tags       []string          `yaml:"tags,omitempty"`
	ToolCompat []string          `yaml:"tool_compat,omitempty"`
	// Vendor controls whether melon manages .gitignore for its cache and symlinks.
	// When nil or true, melon never touches .gitignore (default: vendor everything).
	// When false, melon keeps .gitignore in sync across install/add/remove.
	Vendor *bool `yaml:"vendor,omitempty"`
}

// IsVendored reports whether the manifest is in vendor mode.
// Returns true when Vendor is nil (field absent) or explicitly true.
func (m Manifest) IsVendored() bool {
	return m.Vendor == nil || *m.Vendor
}
