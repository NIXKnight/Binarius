package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nixknight/binarius/internal/utils"
	"github.com/nixknight/binarius/pkg/config"
	"github.com/nixknight/binarius/pkg/installer"
	"github.com/nixknight/binarius/pkg/paths"
	"github.com/nixknight/binarius/pkg/tools"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <tool>@<version>",
	Short: "Install a tool version",
	Long: `Install a specific version of a tool.

Examples:
  binarius install terraform@v1.6.0
  binarius install tofu@latest
  binarius install terragrunt@v0.54.0

The tool binary will be downloaded, verified, and installed to ~/.binarius/tools/<tool>/<version>/`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Parse tool@version
	parts := strings.Split(args[0], "@")
	if len(parts) != 2 {
		return utils.NewUserError(
			"Invalid argument format",
			fmt.Sprintf("Expected format: <tool>@<version>, got: %s", args[0]),
			"Use format like 'terraform@v1.6.0' or 'tofu@latest'",
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

	// Get tool from registry
	tool, err := tools.Get(toolName)
	if err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Tool '%s' is not supported", toolName),
			err.Error(),
			fmt.Sprintf("Supported tools: %s", strings.Join(tools.List(), ", ")),
		)
	}

	// Resolve version if it's "latest"
	if version == "latest" {
		fmt.Printf("Resolving latest version for %s...\n", toolName)
		versions, err := tool.ListVersions()
		if err != nil {
			return utils.NewUserError(
				"Failed to fetch available versions",
				err.Error(),
				"Check your internet connection and try again",
			)
		}

		if len(versions) == 0 {
			return utils.NewUserError(
				"No versions found",
				fmt.Sprintf("No versions available for %s", toolName),
				"Contact the tool maintainer or check the official website",
			)
		}

		version = versions[0] // First version is the latest
		fmt.Printf("Latest version: %s\n", version)
	}

	// Normalize version (ensure 'v' prefix)
	version, err = utils.NormalizeVersion(version)
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
	cacheDir, err := paths.CacheDir()
	if err != nil {
		return err
	}

	toolsDir, err := paths.ToolsDir()
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

	// Check if already installed
	if registry.IsInstalled(toolName, version) {
		fmt.Printf("✓ %s@%s is already installed\n", toolName, version)
		return nil
	}

	// Get download URL
	osName := runtime.GOOS
	arch := runtime.GOARCH
	downloadURL := tool.GetDownloadURL(version, osName, arch)

	fmt.Printf("Installing %s@%s for %s/%s...\n", toolName, version, osName, arch)
	fmt.Printf("Download URL: %s\n", downloadURL)

	// Determine archive filename from URL
	urlParts := strings.Split(downloadURL, "/")
	archiveName := urlParts[len(urlParts)-1]
	archivePath := filepath.Join(cacheDir, archiveName)

	// Download archive
	fmt.Println("Downloading...")
	if err := installer.Download(downloadURL, archivePath); err != nil {
		return err
	}
	fmt.Println("✓ Download complete")

	// Download checksum file and verify
	checksumURL := tool.GetChecksumURL(version, osName, arch)
	checksumPath := filepath.Join(cacheDir, fmt.Sprintf("%s-%s.sha256sums", toolName, version))

	fmt.Println("Downloading checksums...")
	if err := installer.Download(checksumURL, checksumPath); err != nil {
		return utils.NewUserError(
			"Failed to download checksum file",
			err.Error(),
			fmt.Sprintf("Could not download checksums from %s. Check your internet connection.", checksumURL),
		)
	}

	// Parse checksum file to find the expected checksum
	expectedChecksum, err := parseChecksumFile(checksumPath, filepath.Base(archivePath))
	if err != nil {
		return utils.NewUserError(
			"Failed to parse checksum file",
			err.Error(),
			"The checksum file format may be invalid. Please report this issue.",
		)
	}

	// Verify checksum
	fmt.Println("Verifying download integrity...")
	if err := installer.VerifyChecksum(archivePath, expectedChecksum); err != nil {
		// Delete corrupted file
		os.Remove(archivePath)
		return utils.NewUserError(
			"Checksum verification failed",
			err.Error(),
			"The downloaded file may be corrupted or tampered with. Please try downloading again.",
		)
	}
	fmt.Println("✓ Checksum verified")

	// Create version directory
	versionDir := filepath.Join(toolsDir, toolName, version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to create version directory: %s", versionDir),
			err.Error(),
			"Ensure you have write permissions for ~/.binarius",
		)
	}

	// Extract archive based on format
	archiveFormat := tool.GetArchiveFormat()
	fmt.Printf("Extracting %s archive...\n", archiveFormat)

	switch archiveFormat {
	case "zip":
		if err := installer.ExtractZip(archivePath, versionDir); err != nil {
			return err
		}
	case "tar.gz":
		if err := installer.ExtractTarGz(archivePath, versionDir); err != nil {
			return err
		}
	case "binary":
		// Direct binary, just copy it
		binaryPath := filepath.Join(versionDir, tool.GetBinaryName())
		if err := copyFile(archivePath, binaryPath); err != nil {
			return utils.NewUserError(
				"Failed to copy binary",
				err.Error(),
				"Ensure you have write permissions for ~/.binarius",
			)
		}
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return err
		}
	default:
		return utils.NewUserError(
			"Unsupported archive format",
			fmt.Sprintf("Archive format '%s' is not supported", archiveFormat),
			"This is a bug. Please report it to the maintainer.",
		)
	}

	fmt.Println("✓ Extraction complete")

	// Verify binary exists
	binaryPath := filepath.Join(versionDir, tool.GetBinaryName())
	binaryInfo, err := os.Stat(binaryPath)
	if err != nil {
		return utils.NewUserError(
			"Binary not found after extraction",
			fmt.Sprintf("Expected binary at %s, but it doesn't exist", binaryPath),
			"The downloaded archive may not contain the expected binary",
		)
	}

	// Update registry
	toolVersion := config.ToolVersion{
		ToolName:     toolName,
		Version:      version,
		BinaryPath:   binaryPath,
		InstalledAt:  time.Now(),
		SizeBytes:    binaryInfo.Size(),
		SourceURL:    downloadURL,
		Checksum:     expectedChecksum,
		Architecture: fmt.Sprintf("%s/%s", osName, arch),
		Status:       "complete",
	}

	registry.AddVersion(toolName, version, toolVersion)
	if err := config.SaveRegistry(registry, registryPath); err != nil {
		return utils.NewUserError(
			"Failed to update installation registry",
			err.Error(),
			"The installation completed but was not recorded. Run 'binarius init' and try again.",
		)
	}

	fmt.Printf("\n✓ Successfully installed %s@%s\n", toolName, version)
	fmt.Printf("Binary: %s\n", binaryPath)
	fmt.Printf("\nTo use this version, run:\n    binarius use %s@%s\n", toolName, version)

	return nil
}

// parseChecksumFile reads a SHA256SUMS file and extracts the checksum for the given filename.
// SHA256SUMS files follow the format: "checksum  filename" (two spaces between).
func parseChecksumFile(checksumPath, targetFilename string) (string, error) {
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return "", fmt.Errorf("failed to read checksum file: %w", err)
	}

	// Parse each line: "checksum  filename"
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split on whitespace
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		checksum := parts[0]
		filename := parts[1]

		if filename == targetFilename {
			return checksum, nil
		}
	}

	return "", fmt.Errorf("checksum not found for %s in checksums file", targetFilename)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}

	return nil
}
