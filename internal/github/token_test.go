package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "a")
	t.Setenv("GH_TOKEN", "b")
	assert.Equal(t, "a", Token(), "GITHUB_TOKEN wins")

	t.Setenv("GITHUB_TOKEN", "")
	assert.Equal(t, "b", Token(), "falls back to GH_TOKEN")

	t.Setenv("GH_TOKEN", "")
	assert.Empty(t, Token())
}
