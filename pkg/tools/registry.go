package tools

import (
	"fmt"
	"sync"
)

// Tool defines the interface that all managed tools must implement.
// Each tool (terraform, tofu, terragrunt) will have a concrete implementation of this interface.
//
// All methods must be safe for concurrent use (read-only operations).
// All URLs returned must use HTTPS scheme only.
// Version strings must follow semantic versioning (vX.Y.Z format).
type Tool interface {
	// GetName returns the unique identifier for this tool (e.g., "terraform", "tofu").
	// Must be lowercase, alphanumeric + hyphens only.
	GetName() string

	// GetDownloadURL returns the HTTPS URL for downloading the tool binary or archive
	// for the specified version, operating system, and architecture.
	GetDownloadURL(version, os, arch string) string

	// GetChecksumURL returns the HTTPS URL for the SHA256 checksum file.
	GetChecksumURL(version, os, arch string) string

	// ListVersions fetches all available versions from the official source.
	// Returns versions in descending order (newest first).
	ListVersions() ([]string, error)

	// GetBinaryName returns the name of the executable binary within the downloaded archive.
	GetBinaryName() string

	// GetArchiveFormat returns the archive format: "zip", "tar.gz", or "binary".
	GetArchiveFormat() string

	// SupportedArchs returns the list of CPU architectures this tool supports.
	SupportedArchs() []string
}

// registry is the global tool registry that stores all registered tools.
type registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// Global registry instance
var globalRegistry = &registry{
	tools: make(map[string]Tool),
}

// Register adds a tool implementation to the global registry.
// Returns an error if a tool with the same name is already registered.
// This function is safe for concurrent use.
func Register(name string, tool Tool) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if _, exists := globalRegistry.tools[name]; exists {
		return fmt.Errorf("tool %q is already registered", name)
	}

	// Validate that the tool's GetName() matches the registration name
	if tool.GetName() != name {
		return fmt.Errorf("tool name mismatch: registration name %q != GetName() %q", name, tool.GetName())
	}

	globalRegistry.tools[name] = tool
	return nil
}

// Get retrieves a tool implementation by name from the global registry.
// Returns an error if the tool is not registered.
// This function is safe for concurrent use.
func Get(name string) (Tool, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	tool, exists := globalRegistry.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %q is not registered", name)
	}

	return tool, nil
}

// List returns all registered tool names.
// The order is undefined; callers should sort if needed.
// This function is safe for concurrent use.
func List() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.tools))
	for name := range globalRegistry.tools {
		names = append(names, name)
	}

	return names
}

// Clear removes all registered tools. This is primarily for testing.
// This function is safe for concurrent use.
func Clear() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	globalRegistry.tools = make(map[string]Tool)
}
