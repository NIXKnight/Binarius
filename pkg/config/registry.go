package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ToolVersion represents metadata for a single installed tool version.
type ToolVersion struct {
	ToolName     string    `json:"tool_name,omitempty"`
	Version      string    `json:"version,omitempty"`
	BinaryPath   string    `json:"binary_path"`
	InstalledAt  time.Time `json:"installed_at,omitempty"`
	SizeBytes    int64     `json:"size_bytes,omitempty"`
	SourceURL    string    `json:"source_url,omitempty"`
	Checksum     string    `json:"checksum,omitempty"`
	Architecture string    `json:"architecture,omitempty"`
	Status       string    `json:"status,omitempty"` // "complete", "partial", or "broken"
}

// Registry represents the installation registry that tracks all installed tool versions.
// Structure: map[toolName]map[version]ToolVersion
type Registry struct {
	Tools map[string]map[string]ToolVersion `json:"tools,omitempty"`
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		Tools: make(map[string]map[string]ToolVersion),
	}
}

// LoadRegistry reads and parses the installation registry from the specified path.
// Returns an empty registry if the file doesn't exist.
func LoadRegistry(path string) (*Registry, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return empty registry if file doesn't exist
		return NewRegistry(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file %s: %w", path, err)
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry file %s: %w", path, err)
	}

	// Initialize map if nil
	if registry.Tools == nil {
		registry.Tools = make(map[string]map[string]ToolVersion)
	}

	return &registry, nil
}

// SaveRegistry writes the registry to the specified path using atomic write pattern.
func SaveRegistry(registry *Registry, path string) error {
	// Marshal registry to JSON with indentation for readability
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory %s: %w", dir, err)
	}

	// Write to temporary file
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary registry file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temporary file on failure
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to save registry file: %w", err)
	}

	return nil
}

// AddVersion adds or updates a tool version in the registry.
func (r *Registry) AddVersion(toolName, version string, tv ToolVersion) {
	if r.Tools == nil {
		r.Tools = make(map[string]map[string]ToolVersion)
	}

	if r.Tools[toolName] == nil {
		r.Tools[toolName] = make(map[string]ToolVersion)
	}

	r.Tools[toolName][version] = tv
}

// RemoveVersion removes a tool version from the registry.
// If this was the last version of the tool, the tool entry is also removed.
func (r *Registry) RemoveVersion(toolName, version string) {
	if r.Tools == nil || r.Tools[toolName] == nil {
		return
	}

	delete(r.Tools[toolName], version)

	// Remove tool entry if no versions remain
	if len(r.Tools[toolName]) == 0 {
		delete(r.Tools, toolName)
	}
}

// GetVersion retrieves metadata for a specific tool version.
// Returns empty ToolVersion if the version is not found.
func (r *Registry) GetVersion(toolName, version string) ToolVersion {
	if r.Tools == nil || r.Tools[toolName] == nil {
		return ToolVersion{}
	}

	return r.Tools[toolName][version]
}

// ListVersions returns all installed versions for a specific tool.
// Returns an empty slice if the tool is not found.
func (r *Registry) ListVersions(toolName string) []string {
	if r.Tools == nil || r.Tools[toolName] == nil {
		return []string{}
	}

	versions := make([]string, 0, len(r.Tools[toolName]))
	for version := range r.Tools[toolName] {
		versions = append(versions, version)
	}

	return versions
}

// ListTools returns all tool names in the registry.
// Returns an empty slice if the registry is empty.
func (r *Registry) ListTools() []string {
	if r.Tools == nil {
		return []string{}
	}

	tools := make([]string, 0, len(r.Tools))
	for tool := range r.Tools {
		tools = append(tools, tool)
	}

	return tools
}

// HasVersion checks if a specific tool version is installed.
func (r *Registry) HasVersion(toolName, version string) bool {
	if r.Tools == nil || r.Tools[toolName] == nil {
		return false
	}

	_, exists := r.Tools[toolName][version]
	return exists
}

// IsInstalled checks if a specific tool version is installed (alias for HasVersion).
func (r *Registry) IsInstalled(toolName, version string) bool {
	return r.HasVersion(toolName, version)
}
