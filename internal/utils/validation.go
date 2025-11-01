package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// toolNameRegex matches valid tool names: lowercase alphanumeric with hyphens
	// Examples: terraform, tofu, terragrunt, kubectl, my-tool
	toolNameRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

	// versionRegex matches semantic version strings with optional 'v' prefix
	// Examples: v1.6.0, 1.6.0, v1.6.0-beta1, 1.6.0-rc.1
	versionRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$`)
)

// ValidateToolName validates that a tool name conforms to the expected format:
// lowercase alphanumeric characters with hyphens only.
func ValidateToolName(name string) error {
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if !toolNameRegex.MatchString(name) {
		return fmt.Errorf("invalid tool name %q: must contain only lowercase letters, numbers, and hyphens", name)
	}

	// Additional validation: name cannot start or end with hyphen
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("invalid tool name %q: cannot start or end with hyphen", name)
	}

	// Additional validation: no consecutive hyphens
	if strings.Contains(name, "--") {
		return fmt.Errorf("invalid tool name %q: cannot contain consecutive hyphens", name)
	}

	return nil
}

// ValidateVersion validates that a version string follows semantic versioning format.
// Accepts versions with or without 'v' prefix and optional pre-release suffixes.
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if !versionRegex.MatchString(version) {
		return fmt.Errorf("invalid version %q: must follow semantic versioning (e.g., v1.6.0, 1.6.0-beta1)", version)
	}

	return nil
}

// NormalizeVersion ensures a version string has the 'v' prefix.
// If the version already has a 'v' prefix, it's returned unchanged.
// If not, 'v' is prepended.
func NormalizeVersion(version string) (string, error) {
	// First validate the version format
	if err := ValidateVersion(version); err != nil {
		return "", err
	}

	// Add 'v' prefix if not present
	if !strings.HasPrefix(version, "v") {
		return "v" + version, nil
	}

	return version, nil
}
