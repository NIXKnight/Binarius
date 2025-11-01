package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePath ensures that the target path is within the base directory,
// preventing path traversal attacks. Both paths are resolved to absolute paths
// before comparison.
func ValidatePath(base, target string) error {
	// Resolve to absolute paths
	absBase, err := filepath.Abs(base)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Ensure absBase ends with separator for proper prefix matching
	if !strings.HasSuffix(absBase, string(filepath.Separator)) {
		absBase += string(filepath.Separator)
	}

	// Check if target is within base directory
	if !strings.HasPrefix(absTarget+string(filepath.Separator), absBase) {
		return fmt.Errorf("path traversal attempt: %s is outside %s", target, base)
	}

	return nil
}

// EnsureDir creates a directory with the specified permissions if it doesn't exist.
// If the directory already exists, it does nothing. Parent directories are created
// as needed.
func EnsureDir(path string, perm os.FileMode) error {
	// Check if directory already exists
	info, err := os.Stat(path)
	if err == nil {
		// Directory exists, verify it's actually a directory
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", path)
		}
		return nil
	}

	// If error is not "not exist", return it
	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	// Create directory with parents
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return nil
}

// IsWritable checks if a path is writable by attempting to create a temporary file.
// If the path is a directory, it checks write permission in that directory.
// If the path is a file, it checks write permission on the file.
func IsWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// Path doesn't exist, check parent directory
		parent := filepath.Dir(path)
		return IsWritable(parent)
	}

	if info.IsDir() {
		// Test directory write by creating and removing a temp file
		testFile := filepath.Join(path, ".binarius_write_test")
		f, err := os.Create(testFile)
		if err != nil {
			return false
		}
		f.Close()
		os.Remove(testFile)
		return true
	}

	// Test file write by opening in append mode
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
