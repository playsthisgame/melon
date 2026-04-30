package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playsthisgame/melon/internal/manifest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runOfferAddMany(t *testing.T, paths []string, input string) (string, error) {
	t.Helper()
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader(input))
	err := offerAddMany(cmd, paths)
	return out.String(), err
}

func TestOfferAddMany_EmptyInputProceedsWithInstall(t *testing.T) {
	// Set up a temp project dir with a minimal melon.yaml.
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "melon.yaml"), []byte("name: test\nversion: 0.0.1\n"), 0644))
	origFlagDir := flagDir
	flagDir = tmpDir
	t.Cleanup(func() { flagDir = origFlagDir })

	// Stub runInstallFn to avoid real network/install work.
	origRunInstall := runInstallFn
	var installCalled bool
	runInstallFn = func(cmd *cobra.Command, args []string) error {
		installCalled = true
		return nil
	}
	t.Cleanup(func() { runInstallFn = origRunInstall })

	// Use paths with explicit constraints so LatestTag is not called over the network.
	_, err := runOfferAddMany(t, []string{"github.com/owner/skill-a@^1.0.0"}, "\n")
	require.NoError(t, err)
	assert.True(t, installCalled, "runInstall should have been called")

	// Verify the manifest was updated with the new dep.
	m, err := manifest.Load(filepath.Join(tmpDir, "melon.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "^1.0.0", m.Dependencies["github.com/owner/skill-a"])
}

func TestOfferAddMany_NInputCancels(t *testing.T) {
	origRunInstall := runInstallFn
	var installCalled bool
	runInstallFn = func(cmd *cobra.Command, args []string) error {
		installCalled = true
		return nil
	}
	t.Cleanup(func() { runInstallFn = origRunInstall })

	_, err := runOfferAddMany(t, []string{"github.com/owner/skill-a"}, "n\n")
	require.NoError(t, err)
	assert.False(t, installCalled, "runInstall should not have been called")
}
