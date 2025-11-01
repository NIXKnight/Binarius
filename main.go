package main

import (
	"os"

	"github.com/nixknight/binarius/cmd"
	// Import tools package to trigger init() function for tool registration
	_ "github.com/nixknight/binarius/pkg/tools"
)

// Version information (will be updated for releases)
var (
	Version   = "v0.1.0-dev"
	BuildDate = "development"
	GitCommit = "unknown"
)

func main() {
	// Set version information for cmd package to access
	cmd.Version = Version
	cmd.BuildDate = BuildDate
	cmd.GitCommit = GitCommit

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
