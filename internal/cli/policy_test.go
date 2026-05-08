package cli

import (
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchesAllowedSources(t *testing.T) {
	tests := []struct {
		name     string
		depPath  string
		patterns []string
		want     bool
	}{
		{
			name:     "empty allowlist permits all",
			depPath:  "github.com/anyone/anything",
			patterns: []string{},
			want:     true,
		},
		{
			name:     "wildcard org matches repo",
			depPath:  "github.com/my-company/some-repo",
			patterns: []string{"github.com/my-company/*"},
			want:     true,
		},
		{
			name:     "wildcard org matches deep subpath",
			depPath:  "github.com/my-company/some-repo/skills/my-skill",
			patterns: []string{"github.com/my-company/*"},
			want:     true,
		},
		{
			name:     "wildcard org does not match different org",
			depPath:  "github.com/other-org/repo",
			patterns: []string{"github.com/my-company/*"},
			want:     false,
		},
		{
			name:     "exact pattern matches exact path",
			depPath:  "github.com/my-company/approved-skill",
			patterns: []string{"github.com/my-company/approved-skill"},
			want:     true,
		},
		{
			name:     "exact pattern does not match different repo",
			depPath:  "github.com/my-company/other-skill",
			patterns: []string{"github.com/my-company/approved-skill"},
			want:     false,
		},
		{
			name:     "multiple patterns — first matches",
			depPath:  "github.com/my-company/skill",
			patterns: []string{"github.com/my-company/*", "github.com/trusted/*"},
			want:     true,
		},
		{
			name:     "multiple patterns — second matches",
			depPath:  "github.com/trusted/skill",
			patterns: []string{"github.com/my-company/*", "github.com/trusted/*"},
			want:     true,
		},
		{
			name:     "multiple patterns — none match",
			depPath:  "github.com/stranger/skill",
			patterns: []string{"github.com/my-company/*", "github.com/trusted/*"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, matchesAllowedSources(tt.depPath, tt.patterns))
		})
	}
}

func TestCheckSourcePolicy_NoPolicy(t *testing.T) {
	m := manifest.Manifest{Name: "x", Version: "0.1.0"}
	err := checkSourcePolicy(m, []string{"github.com/anyone/anything"})
	assert.NoError(t, err)
}

func TestCheckSourcePolicy_AllPermitted(t *testing.T) {
	m := manifest.Manifest{
		Name:    "x",
		Version: "0.1.0",
		Policy:  &manifest.PolicyConfig{AllowedSources: []string{"github.com/my-company/*"}},
	}
	err := checkSourcePolicy(m, []string{"github.com/my-company/skill-a", "github.com/my-company/skill-b"})
	assert.NoError(t, err)
}

func TestCheckSourcePolicy_OneBlocked(t *testing.T) {
	m := manifest.Manifest{
		Name:    "x",
		Version: "0.1.0",
		Policy:  &manifest.PolicyConfig{AllowedSources: []string{"github.com/my-company/*"}},
	}
	err := checkSourcePolicy(m, []string{"github.com/my-company/skill-a", "github.com/public/skill"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "github.com/public/skill")
}

func TestCheckSourcePolicy_AllBlockedListedTogether(t *testing.T) {
	m := manifest.Manifest{
		Name:    "x",
		Version: "0.1.0",
		Policy:  &manifest.PolicyConfig{AllowedSources: []string{"github.com/my-company/*"}},
	}
	err := checkSourcePolicy(m, []string{"github.com/public/skill-a", "github.com/stranger/skill-b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "github.com/public/skill-a")
	assert.Contains(t, err.Error(), "github.com/stranger/skill-b")
}
