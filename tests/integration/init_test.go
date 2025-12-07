package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nixknight/binarius/pkg/config"
	"github.com/nixknight/binarius/pkg/paths"
)

// TestInitCommand verifies the init command creates the proper directory structure.
func TestInitCommand(t *testing.T) {
	// Use temporary directory for testing
	tmpHome := t.TempDir()

	// Override home directory for this test
	t.Setenv("HOME", tmpHome)

	// Expected directory structure
	binariusHome := filepath.Join(tmpHome, ".binarius")
	toolsDir := filepath.Join(binariusHome, "tools")
	cacheDir := filepath.Join(binariusHome, "cache")
	configPath := filepath.Join(binariusHome, "config.yaml")
	registryPath := filepath.Join(binariusHome, "installation.json")

	// Run init logic
	// This will be implemented in cmd/init.go
	err := runInit()
	if err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	// Verify directories created
	t.Run("directories created", func(t *testing.T) {
		dirs := []string{binariusHome, toolsDir, cacheDir}
		for _, dir := range dirs {
			info, err := os.Stat(dir)
			if os.IsNotExist(err) {
				t.Errorf("directory %s not created", dir)
				continue
			}
			if err != nil {
				t.Errorf("failed to stat %s: %v", dir, err)
				continue
			}

			// Verify it's a directory
			if !info.IsDir() {
				t.Errorf("%s is not a directory", dir)
			}

			// Verify permissions (0755)
			if info.Mode().Perm() != 0755 {
				t.Errorf("%s permissions = %o, want 0755", dir, info.Mode().Perm())
			}
		}
	})

	// Verify config.yaml created
	t.Run("config.yaml created", func(t *testing.T) {
		info, err := os.Stat(configPath)
		if os.IsNotExist(err) {
			t.Errorf("config.yaml not created at %s", configPath)
			return
		}
		if err != nil {
			t.Fatalf("failed to stat config.yaml: %v", err)
		}

		// Verify it's a file
		if info.IsDir() {
			t.Errorf("%s is a directory, expected file", configPath)
		}

		// Verify permissions (0644)
		if info.Mode().Perm() != 0644 {
			t.Errorf("config.yaml permissions = %o, want 0644", info.Mode().Perm())
		}

		// Verify config content
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		// Verify paths are set correctly
		if cfg.Paths.BinariusHome != binariusHome {
			t.Errorf("config BinariusHome = %s, want %s", cfg.Paths.BinariusHome, binariusHome)
		}

		expectedBinDir := filepath.Join(tmpHome, ".local", "bin")
		if cfg.Paths.BinDir != expectedBinDir {
			t.Errorf("config BinDir = %s, want %s", cfg.Paths.BinDir, expectedBinDir)
		}

		if cfg.Paths.CacheDir != cacheDir {
			t.Errorf("config CacheDir = %s, want %s", cfg.Paths.CacheDir, cacheDir)
		}

		// Verify defaults map exists (empty initially)
		if cfg.Defaults == nil {
			t.Error("config Defaults map is nil, expected empty map")
		}
	})

	// Verify installation.json created
	t.Run("installation.json created", func(t *testing.T) {
		info, err := os.Stat(registryPath)
		if os.IsNotExist(err) {
			t.Errorf("installation.json not created at %s", registryPath)
			return
		}
		if err != nil {
			t.Fatalf("failed to stat installation.json: %v", err)
		}

		// Verify it's a file
		if info.IsDir() {
			t.Errorf("%s is a directory, expected file", registryPath)
		}

		// Verify permissions (0644)
		if info.Mode().Perm() != 0644 {
			t.Errorf("installation.json permissions = %o, want 0644", info.Mode().Perm())
		}

		// Verify registry content (should be empty)
		registry, err := config.LoadRegistry(registryPath)
		if err != nil {
			t.Fatalf("failed to load registry: %v", err)
		}

		if len(registry.Tools) != 0 {
			t.Errorf("registry should be empty initially, got %d tools", len(registry.Tools))
		}
	})

	// Verify idempotency - running init again should not fail
	t.Run("init is idempotent", func(t *testing.T) {
		err := runInit()
		if err != nil {
			t.Errorf("second runInit() error = %v, want nil", err)
		}

		// Verify directories still exist
		if _, err := os.Stat(binariusHome); err != nil {
			t.Errorf("binarius home missing after second init: %v", err)
		}
	})
}

// TestInitPathWarning verifies init warns when ~/.local/bin is not in PATH.
func TestInitPathWarning(t *testing.T) {
	tmpHome := t.TempDir()

	t.Setenv("HOME", tmpHome)
	// Set PATH to not include ~/.local/bin
	t.Setenv("PATH", "/usr/bin:/bin")

	// Run init and capture warning
	// This will be implemented in cmd/init.go with warning output
	warning := checkPathWarning()

	binDir, _ := paths.BinDir()
	if warning {
		// Test expects warning to be true when bin dir not in PATH
		t.Logf("Correctly detected that %s is not in PATH", binDir)
	} else {
		t.Errorf("expected PATH warning when %s not in PATH", binDir)
	}
}

// runInit is a helper function that mimics the init command logic.
// This will be replaced with actual command execution once cmd/init.go is implemented.
func runInit() error {
	// Get paths
	binariusHome, err := paths.BinariusHome()
	if err != nil {
		return err
	}

	toolsDir, err := paths.ToolsDir()
	if err != nil {
		return err
	}

	cacheDir, err := paths.CacheDir()
	if err != nil {
		return err
	}

	// Create directories
	dirs := []string{binariusHome, toolsDir, cacheDir}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create config.yaml
	cfg, err := config.DefaultConfig()
	if err != nil {
		return err
	}
	configPath := filepath.Join(binariusHome, "config.yaml")
	if err := config.Save(cfg, configPath); err != nil {
		return err
	}

	// Create installation.json
	registry := &config.Registry{
		Tools: make(map[string]map[string]config.ToolVersion),
	}
	registryPath := filepath.Join(binariusHome, "installation.json")
	if err := config.SaveRegistry(registry, registryPath); err != nil {
		return err
	}

	return nil
}

// checkPathWarning checks if ~/.local/bin is in PATH.
func checkPathWarning() bool {
	binDir, err := paths.BinDir()
	if err != nil {
		return true
	}

	path := os.Getenv("PATH")

	// Check if binDir is in PATH
	pathDirs := strings.Split(path, ":")
	for _, dir := range pathDirs {
		if dir == binDir {
			return false // No warning needed
		}
	}

	return true // Warning: binDir not in PATH
}
