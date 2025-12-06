package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PathConfig holds directory path configuration for Binarius.
type PathConfig struct {
	BinariusHome string `yaml:"binarius_home"` // Binarius home directory (stores tools, cache, config, registry)
	BinDir       string `yaml:"bin_dir"`       // Symlink directory (active tool versions)
	CacheDir     string `yaml:"cache_dir"`     // Downloaded archives cache directory
}

// Config represents the Binarius configuration stored in config.yaml.
type Config struct {
	Defaults map[string]string `yaml:"defaults"` // Map of tool names to default active versions
	Paths    PathConfig        `yaml:"paths"`    // Directory paths configuration
}

// DefaultConfig returns a Config with default values based on the user's home directory.
func DefaultConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	return &Config{
		Defaults: make(map[string]string),
		Paths: PathConfig{
			BinariusHome: filepath.Join(homeDir, ".binarius"),
			BinDir:       filepath.Join(homeDir, ".local", "bin"),
			CacheDir:     filepath.Join(homeDir, ".binarius", "cache"),
		},
	}, nil
}

// Load reads and parses the configuration file from the specified path.
// Returns an error if the file doesn't exist or contains invalid YAML.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Initialize Defaults map if it's nil
	if config.Defaults == nil {
		config.Defaults = make(map[string]string)
	}

	return &config, nil
}

// Save writes the configuration to the specified path using atomic write pattern.
// The configuration is written to a temporary file and then renamed to ensure atomicity.
func Save(config *Config, path string) error {
	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	// Write to temporary file
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary config file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temporary file on failure
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}

// SetDefault sets the default version for a tool.
// If the version is empty, the tool's default is removed.
func (c *Config) SetDefault(tool, version string) {
	if c.Defaults == nil {
		c.Defaults = make(map[string]string)
	}

	if version == "" {
		delete(c.Defaults, tool)
	} else {
		c.Defaults[tool] = version
	}
}

// GetDefault returns the default version for a tool.
// Returns an empty string if no default is set.
func (c *Config) GetDefault(tool string) string {
	if c.Defaults == nil {
		return ""
	}
	return c.Defaults[tool]
}
