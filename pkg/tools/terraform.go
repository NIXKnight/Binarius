package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Terraform implements the Tool interface for HashiCorp Terraform.
type Terraform struct {
	Name string
}

// GetName returns the tool name.
func (t *Terraform) GetName() string {
	return t.Name
}

// GetDownloadURL returns the download URL for a specific terraform version.
// HashiCorp uses the pattern: https://releases.hashicorp.com/terraform/{version}/terraform_{version}_{os}_{arch}.zip
func (t *Terraform) GetDownloadURL(version, os, arch string) string {
	// Remove 'v' prefix if present for consistency with HashiCorp URLs
	version = strings.TrimPrefix(version, "v")

	return fmt.Sprintf(
		"https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip",
		version, version, os, arch,
	)
}

// GetChecksumURL returns the URL for the SHA256SUMS file for a specific version.
// HashiCorp provides checksums at: https://releases.hashicorp.com/terraform/{version}/terraform_{version}_SHA256SUMS
func (t *Terraform) GetChecksumURL(version, os, arch string) string {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	return fmt.Sprintf(
		"https://releases.hashicorp.com/terraform/%s/terraform_%s_SHA256SUMS",
		version, version,
	)
}

// ListVersions fetches all available terraform versions from HashiCorp's releases API.
// Returns versions in descending order (newest first).
func (t *Terraform) ListVersions() ([]string, error) {
	// HashiCorp releases API endpoint
	indexURL := "https://releases.hashicorp.com/terraform/index.json"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch the index
	resp, err := client.Get(indexURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch terraform versions from HashiCorp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch terraform versions: HTTP %d", resp.StatusCode)
	}

	// Parse the JSON response
	var index struct {
		Versions map[string]interface{} `json:"versions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to parse terraform versions index: %w", err)
	}

	// Extract version strings
	versions := make([]string, 0, len(index.Versions))
	for version := range index.Versions {
		// Add 'v' prefix if not present
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}
		versions = append(versions, version)
	}

	// Sort versions in descending order (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})

	return versions, nil
}

// GetBinaryName returns the name of the terraform binary.
func (t *Terraform) GetBinaryName() string {
	return "terraform"
}

// GetArchiveFormat returns the archive format for terraform downloads.
func (t *Terraform) GetArchiveFormat() string {
	return "zip"
}

// SupportedArchs returns the list of architectures supported by terraform.
func (t *Terraform) SupportedArchs() []string {
	return []string{"amd64", "arm64", "386", "arm"}
}

// compareVersions compares two semantic versions.
// Returns:
//   - positive number if v1 > v2
//   - negative number if v1 < v2
//   - zero if v1 == v2
//
// This is a simplified comparison that works for most cases.
// For production, consider using a proper semver library.
func compareVersions(v1, v2 string) int {
	// Remove 'v' prefix
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Split into parts
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each part
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int

		if i < len(parts1) {
			// Ignore error - if parsing fails, p1 remains 0 which is acceptable for version comparison
			_, _ = fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) {
			// Ignore error - if parsing fails, p2 remains 0 which is acceptable for version comparison
			_, _ = fmt.Sscanf(parts2[i], "%d", &p2)
		}

		if p1 != p2 {
			return p1 - p2
		}
	}

	return 0
}

// init registers the terraform tool in the global registry.
// This function is called automatically when the package is imported.
func init() {
	if err := Register("terraform", &Terraform{Name: "terraform"}); err != nil {
		// This should never happen in normal operation, but we need to handle it
		// gracefully. Panic is appropriate here as this is a programming error
		// (duplicate registration or initialization issue).
		panic(fmt.Sprintf("failed to register terraform tool: %v", err))
	}
}
