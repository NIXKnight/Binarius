package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nixknight/binarius/internal/utils"
	"github.com/nixknight/binarius/pkg/config"
	"github.com/nixknight/binarius/pkg/paths"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <tool>",
	Short: "Show active version information",
	Long: `Display detailed information about the currently active version of a tool.

Shows:
  - Active version
  - Binary path
  - Installation date
  - Binary size
  - Source URL
  - Architecture

Example:
  binarius info terraform`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	// Validate tool name
	if err := utils.ValidateToolName(toolName); err != nil {
		return utils.NewUserError(
			"Invalid tool name",
			err.Error(),
			"Tool name must be lowercase alphanumeric with hyphens only",
		)
	}

	// Get paths
	binariusHome, err := paths.BinariusHome()
	if err != nil {
		return err
	}

	registryPath := filepath.Join(binariusHome, "installation.json")
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

	// Check if tool has any installed versions
	versions := registry.ListVersions(toolName)
	if len(versions) == 0 {
		return utils.NewUserError(
			fmt.Sprintf("No versions of %s are installed", toolName),
			"Tool not found in registry",
			fmt.Sprintf("Run 'binarius install %s@<version>' to install it", toolName),
		)
	}

	// Resolve symlink to get active version
	symlinkPath := filepath.Join(binDir, toolName)
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return utils.NewUserError(
			fmt.Sprintf("No active version of %s", toolName),
			"Symlink not found or broken",
			fmt.Sprintf("Run 'binarius use %s@<version>' to activate a version", toolName),
		)
	}

	// Extract version from symlink target path
	// Format: ~/.binarius/tools/<tool>/<version>/<binary>
	activeVersion := ""
	parts := strings.Split(target, string(filepath.Separator))
	for i, part := range parts {
		if part == toolName && i+1 < len(parts) {
			activeVersion = parts[i+1]
			break
		}
	}

	if activeVersion == "" {
		return utils.NewUserError(
			"Failed to determine active version",
			fmt.Sprintf("Symlink target has unexpected format: %s", target),
			fmt.Sprintf("Try re-activating: binarius use %s@<version>", toolName),
		)
	}

	// Get version metadata from registry
	toolVersion := registry.GetVersion(toolName, activeVersion)
	if toolVersion.BinaryPath == "" {
		return utils.NewUserError(
			fmt.Sprintf("Version %s@%s not found in registry", toolName, activeVersion),
			"Registry may be corrupted",
			fmt.Sprintf("Try reinstalling: binarius install %s@%s", toolName, activeVersion),
		)
	}

	// Display information
	fmt.Printf("Tool: %s\n", toolName)
	fmt.Printf("Active Version: %s\n", activeVersion)
	fmt.Printf("Binary Path: %s\n", toolVersion.BinaryPath)
	fmt.Printf("Symlink: %s -> %s\n", symlinkPath, target)

	if !toolVersion.InstalledAt.IsZero() {
		fmt.Printf("Installed: %s\n", toolVersion.InstalledAt.Format("2006-01-02 15:04:05"))
	}

	if toolVersion.SizeBytes > 0 {
		fmt.Printf("Binary Size: %s\n", formatBytes(toolVersion.SizeBytes))
	}

	if toolVersion.Architecture != "" {
		fmt.Printf("Architecture: %s\n", toolVersion.Architecture)
	}

	if toolVersion.SourceURL != "" {
		fmt.Printf("Source URL: %s\n", toolVersion.SourceURL)
	}

	if toolVersion.Checksum != "" {
		fmt.Printf("Checksum: %s\n", toolVersion.Checksum)
	}

	// Verify binary still exists
	if _, err := os.Stat(toolVersion.BinaryPath); err != nil {
		fmt.Printf("\n⚠️  WARNING: Binary file not found at %s\n", toolVersion.BinaryPath)
		fmt.Printf("The tool may have been manually deleted. Consider reinstalling.\n")
	}

	return nil
}

// formatBytes formats a byte count as a human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
