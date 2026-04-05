package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var (
	flagDir     string
	flagVerbose bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "mln",
	Short:   "A dependency manager for agentic markdown files",
	Version: version,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Resolve and fetch all dependencies into .melon/, write melon.lock",
	RunE:  runInstall,
}

var addCmd = &cobra.Command{
	Use:   "add <github-path>[@<constraint>]",
	Short: "Add a new dependency to melon.yml and run install",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a dependency from melon.yml and update .melon/ and melon.lock",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

var (
	flagFrozen  bool
	flagNoPlace bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&flagDir, "dir", "", "project root directory (default: current working directory)")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "enable verbose output")

	installCmd.Flags().BoolVar(&flagFrozen, "frozen", false, "fail if melon.lock would change (useful in CI)")
	installCmd.Flags().BoolVar(&flagNoPlace, "no-place", false, "fetch and lock only — skip placement into agent directories")

	rootCmd.AddCommand(initCmd, installCmd, addCmd, removeCmd)
}
