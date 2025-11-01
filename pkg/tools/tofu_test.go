package tools

import (
	"strings"
	"testing"
)

// TestOpenTofuGetName verifies the OpenTofu tool name.
func TestOpenTofuGetName(t *testing.T) {
	tofu := &OpenTofu{Name: "tofu"}
	got := tofu.GetName()
	want := "tofu"

	if got != want {
		t.Errorf("GetName() = %q, want %q", got, want)
	}
}

// TestOpenTofuGetDownloadURL verifies download URL generation.
func TestOpenTofuGetDownloadURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{
			name:    "tofu v1.10.6 linux amd64",
			version: "v1.10.6",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.10.6/tofu_1.10.6_linux_amd64.zip",
		},
		{
			name:    "tofu v1.10.6 linux arm64",
			version: "v1.10.6",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.10.6/tofu_1.10.6_linux_arm64.zip",
		},
		{
			name:    "tofu v1.8.0 linux amd64",
			version: "v1.8.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.8.0/tofu_1.8.0_linux_amd64.zip",
		},
		{
			name:    "version without v prefix",
			version: "1.10.6",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.10.6/tofu_1.10.6_linux_amd64.zip",
		},
		{
			name:    "version without v prefix arm64",
			version: "1.9.0",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.9.0/tofu_1.9.0_linux_arm64.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tofu := &OpenTofu{Name: "tofu"}
			got := tofu.GetDownloadURL(tt.version, tt.os, tt.arch)

			if got != tt.want {
				t.Errorf("GetDownloadURL(%q, %q, %q) = %q, want %q",
					tt.version, tt.os, tt.arch, got, tt.want)
			}

			// Verify HTTPS
			if !strings.HasPrefix(got, "https://") {
				t.Errorf("GetDownloadURL() must return HTTPS URL, got: %s", got)
			}

			// Verify URL contains both tag version (with v) and filename version (without v)
			if !strings.Contains(got, "/v"+strings.TrimPrefix(tt.version, "v")+"/") {
				t.Errorf("GetDownloadURL() should contain tag with v prefix in path: %s", got)
			}
		})
	}
}

// TestOpenTofuGetChecksumURL verifies checksum URL generation.
func TestOpenTofuGetChecksumURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{
			name:    "tofu v1.10.6 checksum",
			version: "v1.10.6",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.10.6/tofu_1.10.6_SHA256SUMS",
		},
		{
			name:    "tofu v1.8.0 checksum",
			version: "v1.8.0",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.8.0/tofu_1.8.0_SHA256SUMS",
		},
		{
			name:    "version without v prefix",
			version: "1.10.6",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.10.6/tofu_1.10.6_SHA256SUMS",
		},
		{
			name:    "version without v prefix arm64",
			version: "1.9.0",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/opentofu/opentofu/releases/download/v1.9.0/tofu_1.9.0_SHA256SUMS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tofu := &OpenTofu{Name: "tofu"}
			got := tofu.GetChecksumURL(tt.version, tt.os, tt.arch)

			if got != tt.want {
				t.Errorf("GetChecksumURL(%q, %q, %q) = %q, want %q",
					tt.version, tt.os, tt.arch, got, tt.want)
			}

			// Verify HTTPS
			if !strings.HasPrefix(got, "https://") {
				t.Errorf("GetChecksumURL() must return HTTPS URL, got: %s", got)
			}

			// Verify URL contains tag version (with v) in path
			if !strings.Contains(got, "/v"+strings.TrimPrefix(tt.version, "v")+"/") {
				t.Errorf("GetChecksumURL() should contain tag with v prefix in path: %s", got)
			}
		})
	}
}

// TestOpenTofuGetBinaryName verifies the binary name.
func TestOpenTofuGetBinaryName(t *testing.T) {
	tofu := &OpenTofu{Name: "tofu"}
	got := tofu.GetBinaryName()
	want := "tofu"

	if got != want {
		t.Errorf("GetBinaryName() = %q, want %q", got, want)
	}
}

// TestOpenTofuGetArchiveFormat verifies the archive format.
func TestOpenTofuGetArchiveFormat(t *testing.T) {
	tofu := &OpenTofu{Name: "tofu"}
	got := tofu.GetArchiveFormat()
	want := "zip"

	if got != want {
		t.Errorf("GetArchiveFormat() = %q, want %q", got, want)
	}
}

// TestOpenTofuSupportedArchs verifies supported architectures.
func TestOpenTofuSupportedArchs(t *testing.T) {
	tofu := &OpenTofu{Name: "tofu"}
	got := tofu.SupportedArchs()

	// Expected architectures for OpenTofu
	expectedArchs := []string{"amd64", "arm64"}

	// Verify length matches
	if len(got) != len(expectedArchs) {
		t.Errorf("SupportedArchs() returned %d architectures, want %d", len(got), len(expectedArchs))
	}

	// Verify each expected architecture is present
	for _, expectedArch := range expectedArchs {
		found := false
		for _, arch := range got {
			if arch == expectedArch {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SupportedArchs() missing %q, got: %v", expectedArch, got)
		}
	}
}

// TestOpenTofuListVersions verifies version listing functionality.
// This test requires network access and will be skipped in short mode.
func TestOpenTofuListVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	tofu := &OpenTofu{Name: "tofu"}
	versions, err := tofu.ListVersions()

	if err != nil {
		t.Fatalf("ListVersions() error = %v", err)
	}

	// Verify we got versions
	if len(versions) == 0 {
		t.Error("ListVersions() returned no versions")
	}

	// Verify versions have v prefix
	for i, version := range versions {
		if !strings.HasPrefix(version, "v") {
			t.Errorf("ListVersions()[%d] = %q, versions should have 'v' prefix", i, version)
		}

		// Verify version format (should contain dots)
		if !strings.Contains(version, ".") {
			t.Errorf("ListVersions()[%d] = %q, doesn't look like a semantic version", i, version)
		}

		// Only check first 5 versions to keep test fast
		if i >= 4 {
			break
		}
	}

	// Verify descending order by checking first two versions if available
	if len(versions) >= 2 {
		// Both should be valid version strings
		first := versions[0]
		second := versions[1]

		// Compare versions - first should be >= second
		cmp := compareTofuVersions(first, second)
		if cmp < 0 {
			t.Errorf("ListVersions() not in descending order: %s comes before %s but is older", first, second)
		}
	}
}

// TestOpenTofuInterface verifies OpenTofu implements Tool interface.
func TestOpenTofuInterface(t *testing.T) {
	var _ Tool = (*OpenTofu)(nil)
}

// TestOpenTofuRegistration verifies OpenTofu can be registered.
func TestOpenTofuRegistration(t *testing.T) {
	// Clear registry for clean test
	Clear()
	defer Clear()

	tofu := &OpenTofu{Name: "tofu"}
	err := Register("tofu", tofu)

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Verify retrieval
	retrieved, err := Get("tofu")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if retrieved.GetName() != "tofu" {
		t.Errorf("retrieved tool name = %q, want %q", retrieved.GetName(), "tofu")
	}
}

// TestOpenTofuURLFormat verifies URL format correctness.
func TestOpenTofuURLFormat(t *testing.T) {
	tofu := &OpenTofu{Name: "tofu"}

	// Test that URLs don't have double slashes (except after https://)
	url := tofu.GetDownloadURL("v1.10.6", "linux", "amd64")

	// Remove https:// prefix for checking
	urlWithoutScheme := strings.TrimPrefix(url, "https://")
	if strings.Contains(urlWithoutScheme, "//") {
		t.Errorf("GetDownloadURL() contains double slashes: %s", url)
	}

	checksumURL := tofu.GetChecksumURL("v1.10.6", "linux", "amd64")
	checksumWithoutScheme := strings.TrimPrefix(checksumURL, "https://")
	if strings.Contains(checksumWithoutScheme, "//") {
		t.Errorf("GetChecksumURL() contains double slashes: %s", checksumURL)
	}
}

// TestOpenTofuVersionPrefixHandling verifies correct handling of v prefix.
func TestOpenTofuVersionPrefixHandling(t *testing.T) {
	tofu := &OpenTofu{Name: "tofu"}

	tests := []struct {
		name         string
		version      string
		wantTagPart  string // Expected in the GitHub tag path
		wantFilePart string // Expected in the filename
	}{
		{
			name:         "version with v prefix",
			version:      "v1.10.6",
			wantTagPart:  "v1.10.6",
			wantFilePart: "1.10.6",
		},
		{
			name:         "version without v prefix",
			version:      "1.10.6",
			wantTagPart:  "v1.10.6",
			wantFilePart: "1.10.6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test download URL
			downloadURL := tofu.GetDownloadURL(tt.version, "linux", "amd64")

			// Should contain tag part with v prefix
			expectedTagPath := "/download/" + tt.wantTagPart + "/"
			if !strings.Contains(downloadURL, expectedTagPath) {
				t.Errorf("GetDownloadURL() should contain %q, got: %s", expectedTagPath, downloadURL)
			}

			// Should contain filename part without v prefix
			expectedFilename := "tofu_" + tt.wantFilePart + "_"
			if !strings.Contains(downloadURL, expectedFilename) {
				t.Errorf("GetDownloadURL() should contain %q, got: %s", expectedFilename, downloadURL)
			}

			// Test checksum URL
			checksumURL := tofu.GetChecksumURL(tt.version, "linux", "amd64")

			// Should contain tag part with v prefix
			if !strings.Contains(checksumURL, expectedTagPath) {
				t.Errorf("GetChecksumURL() should contain %q, got: %s", expectedTagPath, checksumURL)
			}

			// Should contain filename part without v prefix
			expectedChecksumFile := "tofu_" + tt.wantFilePart + "_SHA256SUMS"
			if !strings.Contains(checksumURL, expectedChecksumFile) {
				t.Errorf("GetChecksumURL() should contain %q, got: %s", expectedChecksumFile, checksumURL)
			}
		})
	}
}

// TestOpenTofuGitHubURLStructure verifies the URL follows GitHub release pattern.
func TestOpenTofuGitHubURLStructure(t *testing.T) {
	tofu := &OpenTofu{Name: "tofu"}

	downloadURL := tofu.GetDownloadURL("v1.10.6", "linux", "amd64")

	// Verify GitHub releases URL structure
	expectedParts := []string{
		"https://github.com/",
		"opentofu/opentofu",
		"/releases/download/",
		"v1.10.6",
		"tofu_1.10.6_linux_amd64.zip",
	}

	for _, part := range expectedParts {
		if !strings.Contains(downloadURL, part) {
			t.Errorf("GetDownloadURL() should contain %q, got: %s", part, downloadURL)
		}
	}

	checksumURL := tofu.GetChecksumURL("v1.10.6", "linux", "amd64")

	// Verify checksum URL structure
	expectedChecksumParts := []string{
		"https://github.com/",
		"opentofu/opentofu",
		"/releases/download/",
		"v1.10.6",
		"tofu_1.10.6_SHA256SUMS",
	}

	for _, part := range expectedChecksumParts {
		if !strings.Contains(checksumURL, part) {
			t.Errorf("GetChecksumURL() should contain %q, got: %s", part, checksumURL)
		}
	}
}
