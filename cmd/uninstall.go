package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nixknight/binarius/internal/utils"
	"github.com/nixknight/binarius/pkg/config"
	"github.com/nixknight/binarius/pkg/paths"
	"github.com/nixknight/binarius/pkg/symlink"
	"github.com/spf13/cobra"
)

var (
	forceUninstall bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <tool>@<version>",
	Short: "Uninstall a tool version",
	Long: `Uninstall a specific version of a tool.

This will:
  - Remove the tool binary files
  - Update the installation registry
  - Warn if attempting to uninstall the active version

Examples:
  binarius uninstall terraform@v1.5.0
  binarius uninstall tofu@v1.6.0 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVarP(&forceUninstall, "force", "f", false, "Force uninstall without confirmation")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
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
	toolsDir, err := paths.ToolsDir()
	if err != nil {
		return err
	}

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
			fmt.Sprintf("Run 'binarius list %s' to see installed versions", toolName),
		)
	}

	// Check if this is the active version
	isActive := false
	symlinkPath := filepath.Join(binDir, toolName)
	if target, err := os.Readlink(symlinkPath); err == nil {
		// Extract version from symlink target
		targetParts := strings.Split(target, string(filepath.Separator))
		for i, part := range targetParts {
			if part == toolName && i+1 < len(targetParts) {
				if targetParts[i+1] == version {
					isActive = true
				}
				break
			}
		}
	}

	// Warn if active version
	if isActive {
		fmt.Printf("⚠️  WARNING: %s@%s is currently the active version\n", toolName, version)
		fmt.Println("Uninstalling it will remove the symlink.")
		fmt.Println()
	}

	// Get version metadata for display
	toolVersion := registry.GetVersion(toolName, version)

	// Confirmation prompt unless --force
	if !forceUninstall {
		fmt.Printf("You are about to uninstall:\n")
		fmt.Printf("  Tool: %s\n", toolName)
		fmt.Printf("  Version: %s\n", version)
		fmt.Printf("  Binary: %s\n", toolVersion.BinaryPath)
		fmt.Println()

		fmt.Print("Are you sure you want to continue? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return utils.NewUserError(
				"Failed to read user input",
				err.Error(),
				"Use --force flag to skip confirmation",
			)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Uninstall cancelled")
			return nil
		}
	}

	// Remove version directory
	versionDir := filepath.Join(toolsDir, toolName, version)
	if err := os.RemoveAll(versionDir); err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to remove version directory: %s", versionDir),
			err.Error(),
			"Ensure you have write permissions for ~/.binarius",
		)
	}

	fmt.Printf("✓ Removed files from %s\n", versionDir)

	// Update registry
	registry.RemoveVersion(toolName, version)
	if err := config.SaveRegistry(registry, registryPath); err != nil {
		return utils.NewUserError(
			"Failed to update installation registry",
			err.Error(),
			"The files were removed but the registry was not updated",
		)
	}

	fmt.Printf("✓ Updated installation registry\n")

	// Check if tool directory is now empty and remove it
	toolDir := filepath.Join(toolsDir, toolName)
	entries, err := os.ReadDir(toolDir)
	if err == nil && len(entries) == 0 {
		if err := os.Remove(toolDir); err == nil {
			fmt.Printf("✓ Removed empty tool directory: %s\n", toolDir)
		}
	}

	// If this was the active version, remove the broken symlink
	if isActive {
		manager := &symlink.Manager{}
		if err := manager.Remove(symlinkPath); err != nil {
			// Non-fatal - warn but don't fail uninstall
			fmt.Printf("⚠️  Warning: Could not remove symlink at %s: %v\n", symlinkPath, err)
		} else {
			fmt.Printf("✓ Removed symlink at %s\n", symlinkPath)
		}
	}

	fmt.Printf("\n✓ Successfully uninstalled %s@%s\n", toolName, version)

	// If this was the active version, provide guidance
	if isActive {
		remainingVersions := registry.ListVersions(toolName)
		if len(remainingVersions) > 0 {
			fmt.Printf("\nTo set a new active version, run:\n")
			fmt.Printf("    binarius use %s@<version>\n", toolName)
			fmt.Printf("\nAvailable versions:\n")
			for _, v := range remainingVersions {
				fmt.Printf("    %s\n", v)
			}
		} else {
			fmt.Printf("\nNo other versions of %s are installed.\n", toolName)
		}
	}

	return nil
}
