package manifest

// Manifest is the parsed representation of a project's melon.yml.
type Manifest struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description,omitempty"`
	Entrypoint   string            `yaml:"entrypoint"`
	Dependencies map[string]string `yaml:"dependencies,omitempty"`
	// Outputs maps target filenames to glob patterns of dep names to include.
	// Example: "CLAUDE.md": "*"  or  ".claude/SKILL.md": "github.com/alice/*"
	Outputs     map[string]string `yaml:"outputs,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	ToolCompat  []string          `yaml:"tool_compat,omitempty"`
}
