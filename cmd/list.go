package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nixknight/binarius/internal/utils"
	"github.com/nixknight/binarius/pkg/config"
	"github.com/nixknight/binarius/pkg/paths"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [tool]",
	Short: "List installed versions",
	Long: `List all installed tool versions.

Without arguments, lists all installed tools and their versions.
With a tool name, lists only versions of that specific tool.

Examples:
  binarius list              # List all tools and versions
  binarius list terraform    # List only terraform versions`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Get active versions by reading symlinks
	activeVersions := make(map[string]string)
	allTools := registry.ListTools()
	for _, tool := range allTools {
		symlinkPath := filepath.Join(binDir, tool)
		if target, err := os.Readlink(symlinkPath); err == nil {
			// Extract version from symlink target path
			// Format: ~/.binarius/tools/<tool>/<version>/<binary>
			parts := strings.Split(target, string(filepath.Separator))
			if len(parts) >= 2 {
				// Find the version part (between tool name and binary name)
				for i, part := range parts {
					if part == tool && i+1 < len(parts) {
						activeVersions[tool] = parts[i+1]
						break
					}
				}
			}
		}
	}

	// If specific tool requested
	if len(args) == 1 {
		toolName := args[0]

		// Validate tool name
		if err := utils.ValidateToolName(toolName); err != nil {
			return utils.NewUserError(
				"Invalid tool name",
				err.Error(),
				"Tool name must be lowercase alphanumeric with hyphens only",
			)
		}

		versions := registry.ListVersions(toolName)
		if len(versions) == 0 {
			fmt.Printf("No versions of %s are installed\n", toolName)
			fmt.Printf("\nTo install %s, run:\n    binarius install %s@<version>\n", toolName, toolName)
			return nil
		}

		// Sort versions
		sort.Strings(versions)

		fmt.Printf("Installed versions of %s:\n", toolName)
		for _, version := range versions {
			marker := "  "
			if activeVersion, ok := activeVersions[toolName]; ok && activeVersion == version {
				marker = "* " // Active version
			}
			fmt.Printf("%s %s\n", marker, version)
		}

		if activeVersion, ok := activeVersions[toolName]; ok {
			fmt.Printf("\n* Active version: %s\n", activeVersion)
		} else {
			fmt.Printf("\nNo active version (no symlink found)\n")
			fmt.Printf("To activate a version, run:\n    binarius use %s@<version>\n", toolName)
		}

		return nil
	}

	// List all tools
	tools := registry.ListTools()
	if len(tools) == 0 {
		fmt.Println("No tools installed")
		fmt.Println("\nTo install a tool, run:")
		fmt.Println("    binarius install <tool>@<version>")
		fmt.Println("\nExample:")
		fmt.Println("    binarius install terraform@v1.6.0")
		return nil
	}

	// Sort tools alphabetically
	sort.Strings(tools)

	fmt.Println("Installed tools and versions:")
	fmt.Println()

	for _, tool := range tools {
		versions := registry.ListVersions(tool)
		sort.Strings(versions)

		activeVersion := ""
		if av, ok := activeVersions[tool]; ok {
			activeVersion = av
		}

		fmt.Printf("%s:\n", tool)
		for _, version := range versions {
			marker := "  "
			if version == activeVersion {
				marker = "* " // Active version
			}
			fmt.Printf("%s %s\n", marker, version)
		}
		fmt.Println()
	}

	fmt.Println("* = Active version")

	return nil
}
