package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// OpenTofu implements the Tool interface for OpenTofu.
// OpenTofu is an open-source fork of Terraform maintained by the Linux Foundation.
type OpenTofu struct {
	Name string
}

// GetName returns the tool name.
func (o *OpenTofu) GetName() string {
	return o.Name
}

// GetDownloadURL returns the download URL for a specific OpenTofu version.
// OpenTofu releases use the pattern:
// https://github.com/opentofu/opentofu/releases/download/v{version}/tofu_{version}_linux_{arch}.zip
//
// Note: GitHub release tags include 'v' prefix, but binary filenames do not.
func (o *OpenTofu) GetDownloadURL(version, os, arch string) string {
	// Ensure version has 'v' prefix for the tag part
	versionTag := version
	if !strings.HasPrefix(versionTag, "v") {
		versionTag = "v" + versionTag
	}

	// Remove 'v' prefix for the filename part
	versionNum := strings.TrimPrefix(version, "v")

	return fmt.Sprintf(
		"https://github.com/opentofu/opentofu/releases/download/%s/tofu_%s_%s_%s.zip",
		versionTag, versionNum, os, arch,
	)
}

// GetChecksumURL returns the URL for the SHA256SUMS file for a specific version.
// OpenTofu provides checksums at:
// https://github.com/opentofu/opentofu/releases/download/v{version}/tofu_{version}_SHA256SUMS
func (o *OpenTofu) GetChecksumURL(version, os, arch string) string {
	// Ensure version has 'v' prefix for the tag part
	versionTag := version
	if !strings.HasPrefix(versionTag, "v") {
		versionTag = "v" + versionTag
	}

	// Remove 'v' prefix for the filename part
	versionNum := strings.TrimPrefix(version, "v")

	return fmt.Sprintf(
		"https://github.com/opentofu/opentofu/releases/download/%s/tofu_%s_SHA256SUMS",
		versionTag, versionNum,
	)
}

// githubRelease represents a GitHub release API response.
type githubRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	PreRelease bool   `json:"prerelease"`
}

// ListVersions fetches all available OpenTofu versions from GitHub releases API.
// Returns versions in descending order (newest first).
func (o *OpenTofu) ListVersions() ([]string, error) {
	// GitHub API endpoint for OpenTofu releases
	apiURL := "https://api.github.com/repos/opentofu/opentofu/releases?per_page=100"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch releases
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for OpenTofu releases: %w", err)
	}

	// Set User-Agent header (GitHub API requires it)
	req.Header.Set("User-Agent", "binarius-version-manager")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenTofu versions from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch OpenTofu versions: HTTP %d", resp.StatusCode)
	}

	// Parse the JSON response
	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse OpenTofu releases: %w", err)
	}

	// Extract version tags, filtering out drafts and pre-releases
	versions := make([]string, 0, len(releases))
	for _, release := range releases {
		// Skip drafts and pre-releases
		if release.Draft || release.PreRelease {
			continue
		}

		// Tag name should already have 'v' prefix from GitHub
		version := release.TagName
		if version != "" {
			// Ensure 'v' prefix is present
			if !strings.HasPrefix(version, "v") {
				version = "v" + version
			}
			versions = append(versions, version)
		}
	}

	// Sort versions in descending order (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return compareTofuVersions(versions[i], versions[j]) > 0
	})

	return versions, nil
}

// GetBinaryName returns the name of the OpenTofu binary.
func (o *OpenTofu) GetBinaryName() string {
	return "tofu"
}

// GetArchiveFormat returns the archive format for OpenTofu downloads.
func (o *OpenTofu) GetArchiveFormat() string {
	return "zip"
}

// SupportedArchs returns the list of architectures supported by OpenTofu.
func (o *OpenTofu) SupportedArchs() []string {
	return []string{"amd64", "arm64"}
}

// compareTofuVersions compares two semantic versions.
// Returns:
//   - positive number if v1 > v2
//   - negative number if v1 < v2
//   - zero if v1 == v2
//
// This is a simplified comparison that works for most cases.
// Adapted from the terraform.go implementation.
func compareTofuVersions(v1, v2 string) int {
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
			// Ignore error - if parsing fails, p1 remains 0
			_, _ = fmt.Sscanf(parts1[i], "%d", &p1)
		}
		if i < len(parts2) {
			// Ignore error - if parsing fails, p2 remains 0
			_, _ = fmt.Sscanf(parts2[i], "%d", &p2)
		}

		if p1 != p2 {
			return p1 - p2
		}
	}

	return 0
}

// init registers the OpenTofu tool in the global registry.
// This function is called automatically when the package is imported.
func init() {
	if err := Register("tofu", &OpenTofu{Name: "tofu"}); err != nil {
		// This should never happen in normal operation, but we need to handle it
		// gracefully. Panic is appropriate here as this is a programming error
		// (duplicate registration or initialization issue).
		panic(fmt.Sprintf("failed to register tofu tool: %v", err))
	}
}
