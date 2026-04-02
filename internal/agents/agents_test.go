package agents_test

import (
	"errors"
	"testing"

	"github.com/playsthisgame/melon/internal/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveOutputs_KnownAgents(t *testing.T) {
	tests := []struct {
		agent       string
		wantPathKey string
	}{
		{"claude-code", ".claude/skills/"},
		{"cursor", ".agents/skills/"},
		{"codex", ".agents/skills/"},
		{"opencode", ".agents/skills/"},
		{"windsurf", ".windsurf/skills/"},
		{"gemini-cli", ".agents/skills/"},
		{"github-copilot", ".agents/skills/"},
		{"roo", ".roo/skills/"},
		{"cline", ".agents/skills/"},
		{"amp", ".agents/skills/"},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			outputs, err := agents.DeriveOutputs([]string{tt.agent}, "/project")
			require.NoError(t, err)
			val, ok := outputs[tt.wantPathKey]
			assert.True(t, ok, "expected key %q in outputs map", tt.wantPathKey)
			assert.Equal(t, "*", val)
		})
	}
}

func TestDeriveOutputs_MultipleAgents(t *testing.T) {
	outputs, err := agents.DeriveOutputs([]string{"claude-code", "windsurf"}, "/project")
	require.NoError(t, err)
	assert.Equal(t, "*", outputs[".claude/skills/"])
	assert.Equal(t, "*", outputs[".windsurf/skills/"])
}

func TestDeriveOutputs_UnknownAgent(t *testing.T) {
	_, err := agents.DeriveOutputs([]string{"unknown-agent"}, "/project")
	require.Error(t, err)
	assert.True(t, errors.Is(err, agents.ErrUnknownAgent),
		"expected ErrUnknownAgent, got %v", err)
}

func TestDeriveOutputs_UnknownAmongKnown(t *testing.T) {
	_, err := agents.DeriveOutputs([]string{"claude-code", "not-a-real-agent"}, "/project")
	require.Error(t, err)
	assert.True(t, errors.Is(err, agents.ErrUnknownAgent))
}

func TestDeriveOutputs_Empty(t *testing.T) {
	outputs, err := agents.DeriveOutputs([]string{}, "/project")
	require.NoError(t, err)
	assert.Empty(t, outputs)
}

func TestKnownAgents_ContainsExpected(t *testing.T) {
	known := agents.KnownAgents()
	expected := []string{"amp", "claude-code", "cline", "codex", "cursor",
		"gemini-cli", "github-copilot", "opencode", "roo", "windsurf"}
	assert.Equal(t, expected, known, "KnownAgents should return all 10 agents sorted")
}
