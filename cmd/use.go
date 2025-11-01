package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nixknight/binarius/internal/utils"
	"github.com/nixknight/binarius/pkg/config"
	"github.com/nixknight/binarius/pkg/paths"
	"github.com/nixknight/binarius/pkg/symlink"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <tool>@<version>",
	Short: "Activate a tool version",
	Long: `Activate a specific version of a tool by creating/updating a symlink.

This makes the specified version the active version by creating a symlink:
  ~/.local/bin/<tool> -> ~/.binarius/tools/<tool>/<version>/<binary>

Examples:
  binarius use terraform@v1.6.0
  binarius use tofu@v1.5.0
  binarius use terragrunt@v0.54.0`,
	Args: cobra.ExactArgs(1),
	RunE: runUse,
}

func init() {
	rootCmd.AddCommand(useCmd)
}

func runUse(cmd *cobra.Command, args []string) error {
	// Parse tool@version
	parts := strings.Split(args[0], "@")
	if len(parts) != 2 {
		return utils.NewUserError(
			"Invalid argument format",
			fmt.Sprintf("Expected format: <tool>@<version>, got: %s", args[0]),
			"Use format like 'terraform@v1.6.0'",
		)
	}

	toolName := parts[0]
	version := parts[1]

	// Validate tool name
	if err := utils.ValidateToolName(toolName); err != nil {
		return utils.NewUserError(
			"Invalid tool name",
			err.Error(),
			"Tool name must be lowercase alphanumeric with hyphens only",
		)
	}

	// Normalize version (ensure 'v' prefix)
	version, err := utils.NormalizeVersion(version)
	if err != nil {
		return utils.NewUserError(
			"Invalid version format",
			err.Error(),
			"Version must follow semantic versioning (e.g., v1.6.0, 1.6.0-beta1)",
		)
	}

	// Get paths
	binariusHome, err := paths.BinariusHome()
	if err != nil {
		return err
	}

	registryPath := filepath.Join(binariusHome, "installation.json")
	configPath := filepath.Join(binariusHome, "config.yaml")

	binDir, err := paths.BinDir()
	if err != nil {
		return err
	}

	// Load registry
	registry, err := config.LoadRegistry(registryPath)
	if err != nil {
		return utils.NewUserError(
			"Failed to load installation registry",
			err.Error(),
			"Run 'binarius init' to initialize Binarius",
		)
	}

	// Check if version is installed
	if !registry.IsInstalled(toolName, version) {
		return utils.NewUserError(
			fmt.Sprintf("%s@%s is not installed", toolName, version),
			"Version not found in registry",
			fmt.Sprintf("Run 'binarius install %s@%s' to install it", toolName, version),
		)
	}

	// Get tool version metadata
	toolVersion := registry.GetVersion(toolName, version)

	// Create symlink
	symlinkPath := filepath.Join(binDir, toolName)
	sourcePath := toolVersion.BinaryPath

	manager := &symlink.Manager{}
	if err := manager.Update(sourcePath, symlinkPath); err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to create symlink at %s", symlinkPath),
			err.Error(),
			fmt.Sprintf("Ensure %s exists and is writable", binDir),
		)
	}

	fmt.Printf("✓ Activated %s@%s\n", toolName, version)
	fmt.Printf("Symlink: %s -> %s\n", symlinkPath, sourcePath)

	// Update config with default version
	cfg, err := config.Load(configPath)
	if err != nil {
		// If config doesn't exist, create a default one
		cfg, err = config.DefaultConfig()
		if err != nil {
			return err
		}
	}

	cfg.SetDefault(toolName, version)
	if err := config.Save(cfg, configPath); err != nil {
		// Non-fatal: symlink is created, but config update failed
		fmt.Printf("⚠️  Warning: Failed to update config.yaml with default version\n")
	} else {
		fmt.Printf("Updated default version in config.yaml\n")
	}

	fmt.Printf("\nYou can now use '%s' command with version %s\n", toolName, version)

	return nil
}
