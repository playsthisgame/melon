package cli

import (
	"bytes"
	"strings"
	"testing"

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
	// Stub runAdd to track calls without doing real work.
	origRunAdd := runAddFn
	t.Cleanup(func() { runAddFn = origRunAdd })
	var installed []string
	runAddFn = func(cmd *cobra.Command, args []string) error {
		installed = append(installed, args[0])
		return nil
	}

	_, err := runOfferAddMany(t, []string{"github.com/owner/skill-a"}, "\n")
	require.NoError(t, err)
	assert.Equal(t, []string{"github.com/owner/skill-a"}, installed)
}

func TestOfferAddMany_NInputCancels(t *testing.T) {
	origRunAdd := runAddFn
	t.Cleanup(func() { runAddFn = origRunAdd })
	var installed []string
	runAddFn = func(cmd *cobra.Command, args []string) error {
		installed = append(installed, args[0])
		return nil
	}

	_, err := runOfferAddMany(t, []string{"github.com/owner/skill-a"}, "n\n")
	require.NoError(t, err)
	assert.Empty(t, installed)
}
