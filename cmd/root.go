package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information (set from main.go)
var (
	Version   string
	BuildDate string
	GitCommit string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "binarius",
	Short: "Binarius - Universal binary version manager",
	Long: `Binarius is a universal binary version manager for CLI tools.

It provides zero-overhead version management through symlink-based execution,
allowing you to install, switch between, and manage multiple versions of
any single-binary CLI tool.

Currently supports: terraform, opentofu (tofu), and terragrunt.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Set version information after it's been initialized from main
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(fmt.Sprintf(`Binarius %s
Build Date: %s
Git Commit: %s
`, Version, BuildDate, GitCommit))

	return rootCmd.Execute()
}
