package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// BinariusHome returns the absolute path to the Binarius home directory.
// Defaults to ~/.binarius, expandable via BINARIUS_HOME environment variable.
func BinariusHome() (string, error) {
	if home := os.Getenv("BINARIUS_HOME"); home != "" {
		return expandPath(home)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".binarius"), nil
}

// BinDir returns the absolute path to the directory containing tool symlinks.
// Defaults to ~/.local/bin, expandable via BINARIUS_BIN_DIR environment variable.
func BinDir() (string, error) {
	if binDir := os.Getenv("BINARIUS_BIN_DIR"); binDir != "" {
		return expandPath(binDir)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "bin"), nil
}

// CacheDir returns the absolute path to the cache directory for downloaded archives.
// Defaults to ~/.binarius/cache, expandable via BINARIUS_CACHE_DIR environment variable.
func CacheDir() (string, error) {
	if cacheDir := os.Getenv("BINARIUS_CACHE_DIR"); cacheDir != "" {
		return expandPath(cacheDir)
	}
	home, err := BinariusHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "cache"), nil
}

// ToolsDir returns the absolute path to the directory containing installed tool binaries.
// Defaults to ~/.binarius/tools.
func ToolsDir() (string, error) {
	home, err := BinariusHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "tools"), nil
}

// expandPath expands tilde (~) prefixes to the user's home directory.
// If the path doesn't start with ~, it's returned as-is.
// Returns an error if the home directory cannot be determined.
func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return homeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:]), nil
	}

	// Path starts with ~ but not ~/ (e.g., ~user/path)
	// We don't support ~user expansion, return as-is
	return path, nil
}
