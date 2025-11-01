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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Binarius directory structure",
	Long: `Initialize Binarius by creating the necessary directory structure and configuration files.

This command:
  - Creates ~/.binarius directory and subdirectories
  - Creates default config.yaml
  - Creates empty installation.json registry
  - Verifies that ~/.local/bin is in your PATH`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Get paths
	binariusHome, err := paths.BinariusHome()
	if err != nil {
		return utils.NewUserError(
			"Failed to determine Binarius home directory",
			err.Error(),
			"Ensure your HOME environment variable is set correctly",
		)
	}

	toolsDir, err := paths.ToolsDir()
	if err != nil {
		return err
	}

	cacheDir, err := paths.CacheDir()
	if err != nil {
		return err
	}

	binDir, err := paths.BinDir()
	if err != nil {
		return err
	}

	// Create directories
	dirs := []string{binariusHome, toolsDir, cacheDir, binDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return utils.NewUserError(
				fmt.Sprintf("Failed to create directory: %s", dir),
				err.Error(),
				"Ensure you have write permissions for your home directory",
			)
		}
	}

	// Create default config.yaml
	configPath := filepath.Join(binariusHome, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig, err := config.DefaultConfig()
		if err != nil {
			return err
		}

		if err := config.Save(defaultConfig, configPath); err != nil {
			return utils.NewUserError(
				"Failed to create config.yaml",
				err.Error(),
				"Ensure you have write permissions for ~/.binarius",
			)
		}
		fmt.Printf("Created config.yaml at %s\n", configPath)
	} else {
		fmt.Printf("Config already exists at %s\n", configPath)
	}

	// Create empty installation.json registry
	registryPath := filepath.Join(binariusHome, "installation.json")
	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		registry := config.NewRegistry()
		if err := config.SaveRegistry(registry, registryPath); err != nil {
			return utils.NewUserError(
				"Failed to create installation.json",
				err.Error(),
				"Ensure you have write permissions for ~/.binarius",
			)
		}
		fmt.Printf("Created installation.json at %s\n", registryPath)
	} else {
		fmt.Printf("Registry already exists at %s\n", registryPath)
	}

	// Check if binDir is in PATH
	pathEnv := os.Getenv("PATH")
	if !strings.Contains(pathEnv, binDir) {
		fmt.Printf("\n⚠️  WARNING: %s is not in your PATH\n", binDir)
		fmt.Println("\nAdd the following to your shell configuration (~/.bashrc or ~/.zshrc):")
		fmt.Printf("    export PATH=\"%s:$PATH\"\n", binDir)
		fmt.Println("\nThen reload your shell configuration:")
		fmt.Println("    source ~/.bashrc  # or source ~/.zshrc")
	} else {
		fmt.Printf("\n✓ %s is in your PATH\n", binDir)
	}

	fmt.Printf("\n✓ Binarius initialized successfully at %s\n", binariusHome)
	return nil
}
