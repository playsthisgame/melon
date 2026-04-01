package semver_test

import (
	"testing"

	"github.com/playsthisgame/melon/pkg/semver"
	"github.com/stretchr/testify/assert"
)

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		// ── Exact match ──────────────────────────────────────────────────────
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "1.2.4", false},
		{"1.2.3", "1.2.2", false},
		// Leading "v" is optional in both fields.
		{"v1.2.3", "1.2.3", true},
		{"1.2.3", "v1.2.3", true},

		// ── Caret (^) — same major, >= base ──────────────────────────────────
		{"^1.2.3", "1.2.3", true},  // exact match
		{"^1.2.3", "1.2.4", true},  // patch bump
		{"^1.2.3", "1.3.0", true},  // minor bump
		{"^1.2.3", "2.0.0", false}, // major bump — incompatible
		{"^1.2.3", "1.2.2", false}, // below base
		{"^1.0.0", "1.9.9", true},
		{"^1.0.0", "2.0.0", false},

		// Caret with major 0 — only patch bumps allowed.
		{"^0.2.3", "0.2.3", true},
		{"^0.2.3", "0.2.9", true},
		{"^0.2.3", "0.3.0", false}, // minor bump when major is 0
		{"^0.2.3", "1.0.0", false},

		// ── Tilde (~) — patch bumps only ─────────────────────────────────────
		{"~1.2.3", "1.2.3", true},
		{"~1.2.3", "1.2.9", true},
		{"~1.2.3", "1.3.0", false}, // minor bump
		{"~1.2.3", "1.2.2", false}, // below base
		{"~1.2.3", "2.0.0", false},

		// ── Invalid inputs ────────────────────────────────────────────────────
		{"^1.2.3", "not-a-version", false},
	}

	for _, tt := range tests {
		t.Run(tt.constraint+"/"+tt.version, func(t *testing.T) {
			got := semver.IsCompatible(tt.constraint, tt.version)
			assert.Equal(t, tt.want, got,
				"IsCompatible(%q, %q)", tt.constraint, tt.version)
		})
	}
}
