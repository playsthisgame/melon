// Package agents encodes the agent_directory_conventions table and derives
// output paths for mln compile based on a project's agent_compat list.
package agents

import (
	"errors"
	"fmt"
	"sort"
)

// ErrUnknownAgent is returned when an agent_compat value has no known convention.
var ErrUnknownAgent = errors.New("agents: unknown agent")

// Convention holds the known install paths for a single agent.
type Convention struct {
	// Project is the project-scoped skills directory, relative to the project root.
	Project string
	// Global is the user-global skills directory (unexpanded ~ path).
	Global string
}

// conventions is the authoritative agent_directory_conventions table.
// Sourced from the vercel-labs/skills agent table.
var conventions = map[string]Convention{
	"claude-code":    {Project: ".claude/skills/", Global: "~/.claude/skills/"},
	"cursor":         {Project: ".agents/skills/", Global: "~/.cursor/skills/"},
	"codex":          {Project: ".agents/skills/", Global: "~/.codex/skills/"},
	"opencode":       {Project: ".agents/skills/", Global: "~/.config/opencode/skills/"},
	"windsurf":       {Project: ".windsurf/skills/", Global: "~/.codeium/windsurf/skills/"},
	"gemini-cli":     {Project: ".agents/skills/", Global: "~/.gemini/skills/"},
	"github-copilot": {Project: ".agents/skills/", Global: "~/.copilot/skills/"},
	"roo":            {Project: ".roo/skills/", Global: "~/.roo/skills/"},
	"cline":          {Project: ".agents/skills/", Global: "~/.agents/skills/"},
	"amp":            {Project: ".agents/skills/", Global: "~/.config/agents/skills/"},
}

// KnownAgents returns the sorted list of all known agent names.
func KnownAgents() []string {
	names := make([]string, 0, len(conventions))
	for name := range conventions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// DeriveOutputs maps each agent in agentCompat to its project-scoped skills
// directory. Returns a map of target path (relative to project root) -> dep name
// glob ("*"), suitable for use as the outputs block in melon.yaml.
//
// Returns ErrUnknownAgent (wrapped) if any value in agentCompat is not in the
// convention table.
//
// Example: agentCompat=["claude-code"] → {".claude/skills/": "*"}
func DeriveOutputs(agentCompat []string, projectDir string) (map[string]string, error) {
	outputs := make(map[string]string, len(agentCompat))
	for _, agent := range agentCompat {
		conv, ok := conventions[agent]
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownAgent, agent)
		}
		outputs[conv.Project] = "*"
	}
	return outputs, nil
}

// DeriveTargets returns the deduplicated, sorted list of project-scoped skill
// directories for each agent in m.AgentCompat. Returns ErrUnknownAgent if any
// agent is not in the convention table.
func DeriveTargets(agentCompat []string) ([]string, error) {
	seen := make(map[string]struct{}, len(agentCompat))
	for _, agent := range agentCompat {
		conv, ok := conventions[agent]
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownAgent, agent)
		}
		seen[conv.Project] = struct{}{}
	}
	targets := make([]string, 0, len(seen))
	for t := range seen {
		targets = append(targets, t)
	}
	sort.Strings(targets)
	return targets, nil
}
