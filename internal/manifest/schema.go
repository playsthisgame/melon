package manifest

// PolicyConfig holds optional source restriction policy for a project.
type PolicyConfig struct {
	// AllowedSources is a list of glob patterns. Only dependency paths matching
	// at least one pattern are permitted by melon add and melon install.
	// When empty or absent, all sources are permitted.
	AllowedSources []string `yaml:"allowed_sources,omitempty"`
}

// IndexConfig holds optional custom index configuration for a project.
type IndexConfig struct {
	// URLs is a list of custom index.yaml URLs to search.
	URLs []string `yaml:"urls,omitempty"`
	// PublicIndex, when false, suppresses the default public melon index so only
	// the custom URLs are searched. Defaults to true (public index included).
	PublicIndex *bool `yaml:"public_index,omitempty"`
}

// Manifest is the parsed representation of a project's melon.yaml.
type Manifest struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description,omitempty"`
	Entrypoint   string            `yaml:"entrypoint"`
	Dependencies map[string]string `yaml:"dependencies,omitempty"`
	// Outputs maps target filenames to glob patterns of dep names to include.
	// Example: "CLAUDE.md": "*"  or  ".claude/SKILL.md": "github.com/alice/*"
	Outputs map[string]string `yaml:"outputs,omitempty"`
	Tags    []string          `yaml:"tags,omitempty"`
	// Harnesses lists the agent harnesses this project targets (claude-code,
	// cursor, codex, …). It drives where melon install places skills.
	// Introduced in v0.5.0.
	Harnesses []string `yaml:"harnesses,omitempty"`
	// ToolCompat is the pre-v0.5.0 name for Harnesses. It is still honored when
	// harnesses is absent so older manifests keep working unchanged. Never read
	// this field directly — call HarnessList instead.
	ToolCompat []string `yaml:"tool_compat,omitempty"`
	// Vendor controls whether melon manages .gitignore for its cache and symlinks.
	// When nil or true, melon never touches .gitignore (default: vendor everything).
	// When false, melon keeps .gitignore in sync across install/add/remove.
	Vendor *bool `yaml:"vendor,omitempty"`
	// Index configures a custom skill registry. When absent, the default public
	// melon index is used.
	Index *IndexConfig `yaml:"index,omitempty"`
	// Policy configures source restrictions. When absent, all sources are permitted.
	Policy *PolicyConfig `yaml:"policy,omitempty"`
}

// IsVendored reports whether the manifest is in vendor mode.
// Returns true when Vendor is nil (field absent) or explicitly true.
func (m Manifest) IsVendored() bool {
	return m.Vendor == nil || *m.Vendor
}

// HarnessList returns the harnesses this project targets. harnesses wins when
// both keys are present; tool_compat is honored for manifests written before
// v0.5.0. Callers must use this rather than reading either field directly.
func (m Manifest) HarnessList() []string {
	if len(m.Harnesses) > 0 {
		return m.Harnesses
	}
	return m.ToolCompat
}

// UsesLegacyToolCompat reports whether the manifest depends on the deprecated
// tool_compat key — that is, tool_compat carries the value because harnesses is
// absent. Drives the one-line deprecation notice printed on install.
func (m Manifest) UsesLegacyToolCompat() bool {
	return len(m.Harnesses) == 0 && len(m.ToolCompat) > 0
}
