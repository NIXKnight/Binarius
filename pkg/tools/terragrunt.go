package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Terragrunt implements the Tool interface for Terragrunt.
type Terragrunt struct {
	Name string
}

// GetName returns the tool name.
func (t *Terragrunt) GetName() string {
	return t.Name
}

// GetDownloadURL returns the download URL for a specific terragrunt version.
// Terragrunt releases are raw binaries, not archives.
// Pattern: https://github.com/gruntwork-io/terragrunt/releases/download/v{version}/terragrunt_linux_{arch}
func (t *Terragrunt) GetDownloadURL(version, os, arch string) string {
	// Ensure version has v prefix for GitHub release tag
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return fmt.Sprintf(
		"https://github.com/gruntwork-io/terragrunt/releases/download/%s/terragrunt_%s_%s",
		version, os, arch,
	)
}

// GetChecksumURL returns the URL for the SHA256SUMS file for a specific version.
// Terragrunt uses simple filename: SHA256SUMS (no version prefix)
func (t *Terragrunt) GetChecksumURL(version, os, arch string) string {
	// Ensure version has v prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return fmt.Sprintf(
		"https://github.com/gruntwork-io/terragrunt/releases/download/%s/SHA256SUMS",
		version,
	)
}

// ListVersions fetches all available terragrunt versions from GitHub releases.
// Filters out alpha versions (alpha-*, v-alpha-*).
// Returns versions in descending order (newest first).
func (t *Terragrunt) ListVersions() ([]string, error) {
	// GitHub releases API endpoint
	apiURL := "https://api.github.com/repos/gruntwork-io/terragrunt/releases?per_page=100"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request with User-Agent (required by GitHub)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "binarius-version-manager")

	// Fetch the releases
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch terragrunt versions from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch terragrunt versions: HTTP %d", resp.StatusCode)
	}

	// Parse the JSON response
	var releases []struct {
		TagName    string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse terragrunt versions: %w", err)
	}

	// Extract version strings, filtering out drafts, prereleases, and alpha versions
	versions := make([]string, 0, len(releases))
	for _, release := range releases {
		// Skip drafts and prereleases
		if release.Draft || release.Prerelease {
			continue
		}

		tag := release.TagName

		// Filter out alpha versions (alpha-YYYYMMDD, v-alpha-*, etc.)
		if strings.HasPrefix(tag, "alpha-") || strings.Contains(tag, "-alpha") {
			continue
		}

		// Add v prefix if not present
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + tag
		}

		versions = append(versions, tag)
	}

	// Sort versions in descending order (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return compareTerragruntVersions(versions[i], versions[j]) > 0
	})

	return versions, nil
}

// GetBinaryName returns the name of the terragrunt binary.
func (t *Terragrunt) GetBinaryName() string {
	return "terragrunt"
}

// GetArchiveFormat returns the archive format for terragrunt downloads.
// Terragrunt releases raw binaries, so return "binary" to indicate no extraction needed.
func (t *Terragrunt) GetArchiveFormat() string {
	return "binary"
}

// SupportedArchs returns the list of architectures supported by terragrunt.
func (t *Terragrunt) SupportedArchs() []string {
	return []string{"amd64", "arm64"}
}

// compareTerragruntVersions compares two semantic versions.
// Returns:
//   - positive number if v1 > v2
//   - negative number if v1 < v2
//   - zero if v1 == v2
func compareTerragruntVersions(v1, v2 string) int {
	// Remove v prefix
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
			fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &p2)
		}

		if p1 != p2 {
			return p1 - p2
		}
	}

	return 0
}

// init registers the terragrunt tool in the global registry.
func init() {
	if err := Register("terragrunt", &Terragrunt{Name: "terragrunt"}); err != nil {
		panic(fmt.Sprintf("failed to register terragrunt tool: %v", err))
	}
}
