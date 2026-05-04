package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagDir     string
	flagVerbose bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Resolve and fetch all dependencies into .melon/, write melon.lock",
	RunE:  runInstall,
}

var addCmd = &cobra.Command{
	Use:   "add <github-path>[@<constraint>]",
	Short: "Add a new dependency to melon.yaml and run install",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a dependency from melon.yaml and update .melon/ and melon.lock",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRemove,
}

var updateCmd = &cobra.Command{
	Use:   "update [dep]",
	Short: "Update dependencies to the latest compatible version",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUpdate,
}

var outdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "Show which dependencies have newer versions available",
	Args:  cobra.NoArgs,
	RunE:  runOutdated,
}

var (
	flagFrozen  bool
	flagNoPlace bool
)

func init() {
	installCmd.Flags().BoolVar(&flagFrozen, "frozen", false, "fail if melon.lock would change (useful in CI)")
	installCmd.Flags().BoolVar(&flagNoPlace, "no-place", false, "fetch and lock only — skip placement into agent directories")
}

// Run builds and executes the cobra command tree under the given binary name.
func Run(name, version string) {
	rootCmd := &cobra.Command{
		Use:     name,
		Short:   "A dependency manager for agentic markdown files",
		Version: version,
	}
	rootCmd.PersistentFlags().StringVar(&flagDir, "dir", "", "project root directory (default: current working directory)")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "enable verbose output")

	rootCmd.AddCommand(initCmd, installCmd, addCmd, removeCmd, updateCmd, outdatedCmd, listCmd, searchCmd, infoCmd, cleanCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
