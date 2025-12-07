package symlink

import (
	"fmt"
	"os"
	"path/filepath"
)

// Manager handles symlink operations for Binarius.
// It provides atomic symlink creation, updates, removal, and verification.
type Manager struct{}

// Create creates a new symlink from target to source.
// Returns an error if the target already exists or if the operation fails.
//
// Parameters:
//   - source: The path to the actual binary
//   - target: The path where the symlink should be created
func (m *Manager) Create(source, target string) error {
	// Verify source exists
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("failed to create symlink: source %s does not exist: %w", source, err)
	}

	// Check if target already exists
	if _, err := os.Lstat(target); err == nil {
		return fmt.Errorf("failed to create symlink: target %s already exists", target)
	}

	// Create the symlink
	if err := os.Symlink(source, target); err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", target, source, err)
	}

	return nil
}

// Update atomically updates an existing symlink or creates it if it doesn't exist.
// Uses a temporary symlink and atomic rename to ensure zero downtime.
//
// Parameters:
//   - source: The path to the actual binary
//   - target: The path where the symlink should be created/updated
func (m *Manager) Update(source, target string) error {
	// Verify source exists
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("failed to update symlink: source %s does not exist: %w", source, err)
	}

	// Get the directory of the target
	targetDir := filepath.Dir(target)

	// Create a temporary symlink in the same directory
	// This ensures atomic rename works (same filesystem)
	tmpLink, err := os.CreateTemp(targetDir, ".binarius-symlink-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file for atomic update: %w", err)
	}
	tmpLinkPath := tmpLink.Name()
	_ = tmpLink.Close()
	_ = os.Remove(tmpLinkPath) // Remove the file, we just need the name

	// Create temporary symlink
	if err := os.Symlink(source, tmpLinkPath); err != nil {
		return fmt.Errorf("failed to create temporary symlink: %w", err)
	}

	// Atomic rename: replace old symlink with new one
	if err := os.Rename(tmpLinkPath, target); err != nil {
		// Clean up temporary symlink on failure
		_ = os.Remove(tmpLinkPath)
		return fmt.Errorf("failed to atomically update symlink: %w", err)
	}

	return nil
}

// Remove removes a symlink at the specified path.
// This operation is idempotent - it succeeds even if the symlink doesn't exist.
//
// Parameters:
//   - target: The path to the symlink to remove
func (m *Manager) Remove(target string) error {
	// Check if symlink exists
	if _, err := os.Lstat(target); os.IsNotExist(err) {
		// Already removed, idempotent success
		return nil
	}

	// Remove the symlink
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("failed to remove symlink %s: %w", target, err)
	}

	return nil
}

// Verify checks that a symlink at target points to the expected source.
// Returns an error if the symlink doesn't exist or points to a different location.
//
// Parameters:
//   - target: The path to the symlink
//   - expectedSource: The expected path the symlink should point to
func (m *Manager) Verify(target, expectedSource string) error {
	// Read the symlink
	actualSource, err := os.Readlink(target)
	if err != nil {
		return fmt.Errorf("failed to read symlink %s: %w", target, err)
	}

	// Compare with expected source
	if actualSource != expectedSource {
		return fmt.Errorf("symlink verification failed: %s points to %s, expected %s",
			target, actualSource, expectedSource)
	}

	return nil
}
