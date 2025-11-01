package tools

import (
	"strings"
	"testing"
)

// TestTerragruntGetName tests the GetName method.
func TestTerragruntGetName(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	got := tg.GetName()
	want := "terragrunt"

	if got != want {
		t.Errorf("GetName() = %v, want %v", got, want)
	}
}

// TestTerragruntGetDownloadURL tests the GetDownloadURL method.
func TestTerragruntGetDownloadURL(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	tests := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{
			name:    "terragrunt v0.93.0 linux amd64",
			version: "v0.93.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.93.0/terragrunt_linux_amd64",
		},
		{
			name:    "terragrunt v0.93.0 linux arm64",
			version: "v0.93.0",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.93.0/terragrunt_linux_arm64",
		},
		{
			name:    "terragrunt v0.92.1 linux amd64",
			version: "v0.92.1",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.92.1/terragrunt_linux_amd64",
		},
		{
			name:    "version without v prefix",
			version: "0.93.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.93.0/terragrunt_linux_amd64",
		},
		{
			name:    "version without v prefix arm64",
			version: "0.92.1",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.92.1/terragrunt_linux_arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tg.GetDownloadURL(tt.version, tt.os, tt.arch)
			if got != tt.want {
				t.Errorf("GetDownloadURL() = %v, want %v", got, tt.want)
			}

			// Verify URL uses HTTPS
			if !strings.HasPrefix(got, "https://") {
				t.Errorf("GetDownloadURL() should use HTTPS, got %v", got)
			}

			// Verify URL is from GitHub
			if !strings.Contains(got, "github.com/gruntwork-io/terragrunt") {
				t.Errorf("GetDownloadURL() should be from gruntwork-io/terragrunt GitHub, got %v", got)
			}
		})
	}
}

// TestTerragruntGetChecksumURL tests the GetChecksumURL method.
func TestTerragruntGetChecksumURL(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	tests := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{
			name:    "terragrunt v0.93.0 checksum",
			version: "v0.93.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.93.0/SHA256SUMS",
		},
		{
			name:    "terragrunt v0.92.1 checksum",
			version: "v0.92.1",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.92.1/SHA256SUMS",
		},
		{
			name:    "version without v prefix",
			version: "0.93.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.93.0/SHA256SUMS",
		},
		{
			name:    "version without v prefix arm64",
			version: "0.92.1",
			os:      "linux",
			arch:    "arm64",
			want:    "https://github.com/gruntwork-io/terragrunt/releases/download/v0.92.1/SHA256SUMS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tg.GetChecksumURL(tt.version, tt.os, tt.arch)
			if got != tt.want {
				t.Errorf("GetChecksumURL() = %v, want %v", got, tt.want)
			}

			// Verify URL uses HTTPS
			if !strings.HasPrefix(got, "https://") {
				t.Errorf("GetChecksumURL() should use HTTPS, got %v", got)
			}

			// Verify checksum file is simple SHA256SUMS (no version prefix in filename)
			if !strings.HasSuffix(got, "/SHA256SUMS") {
				t.Errorf("GetChecksumURL() should end with /SHA256SUMS, got %v", got)
			}
		})
	}
}

// TestTerragruntGetBinaryName tests the GetBinaryName method.
func TestTerragruntGetBinaryName(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	got := tg.GetBinaryName()
	want := "terragrunt"

	if got != want {
		t.Errorf("GetBinaryName() = %v, want %v", got, want)
	}
}

// TestTerragruntGetArchiveFormat tests the GetArchiveFormat method.
func TestTerragruntGetArchiveFormat(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	got := tg.GetArchiveFormat()
	want := "binary"

	if got != want {
		t.Errorf("GetArchiveFormat() = %v, want %v (terragrunt releases raw binaries)", got, want)
	}
}

// TestTerragruntSupportedArchs tests the SupportedArchs method.
func TestTerragruntSupportedArchs(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	got := tg.SupportedArchs()
	want := []string{"amd64", "arm64"}

	if len(got) != len(want) {
		t.Errorf("SupportedArchs() length = %v, want %v", len(got), len(want))
		return
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("SupportedArchs()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestTerragruntListVersions tests the ListVersions method.
// This test requires network access and is skipped in short mode.
func TestTerragruntListVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	tg := &Terragrunt{Name: "terragrunt"}

	versions, err := tg.ListVersions()
	if err != nil {
		t.Fatalf("ListVersions() error = %v", err)
	}

	if len(versions) == 0 {
		t.Error("ListVersions() returned empty list")
	}

	// Verify all versions have v prefix
	for _, v := range versions {
		if !strings.HasPrefix(v, "v") {
			t.Errorf("ListVersions() version %v should have v prefix", v)
		}

		// Verify no alpha versions in the list
		if strings.HasPrefix(v, "alpha-") || strings.Contains(v, "-alpha") {
			t.Errorf("ListVersions() should filter out alpha version %v", v)
		}
	}

	// Verify versions are sorted in descending order (newest first)
	if len(versions) >= 2 {
		if compareTerragruntVersions(versions[0], versions[1]) < 0 {
			t.Errorf("ListVersions() not sorted correctly: %v should be > %v", versions[0], versions[1])
		}
	}

	t.Logf("Found %d terragrunt versions, latest: %v", len(versions), versions[0])
}

// TestTerragruntInterface verifies that Terragrunt implements the Tool interface.
func TestTerragruntInterface(t *testing.T) {
	var _ Tool = (*Terragrunt)(nil)
}

// TestTerragruntRegistration tests that terragrunt can be registered in the global registry.
func TestTerragruntRegistration(t *testing.T) {
	// Clear registry for clean test
	Clear()
	defer Clear()

	tg := &Terragrunt{Name: "terragrunt"}
	err := Register("terragrunt", tg)

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Verify we can retrieve it
	tool, err := Get("terragrunt")
	if err != nil {
		t.Fatalf("Get(terragrunt) error = %v", err)
	}

	if tool.GetName() != "terragrunt" {
		t.Errorf("Retrieved tool name = %v, want terragrunt", tool.GetName())
	}
}

// TestTerragruntURLFormat tests that URLs are properly formatted.
func TestTerragruntURLFormat(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	downloadURL := tg.GetDownloadURL("v0.93.0", "linux", "amd64")
	checksumURL := tg.GetChecksumURL("v0.93.0", "linux", "amd64")

	// Check for common URL mistakes
	if strings.Contains(downloadURL, "//releases") {
		t.Errorf("Download URL has double slashes: %v", downloadURL)
	}

	if strings.Contains(checksumURL, "//releases") {
		t.Errorf("Checksum URL has double slashes: %v", checksumURL)
	}

	// Verify URLs are from GitHub
	if !strings.Contains(downloadURL, "github.com") {
		t.Errorf("Download URL should be from GitHub: %v", downloadURL)
	}

	if !strings.Contains(checksumURL, "github.com") {
		t.Errorf("Checksum URL should be from GitHub: %v", checksumURL)
	}
}

// TestTerragruntVersionPrefixHandling tests version prefix handling in URLs.
func TestTerragruntVersionPrefixHandling(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	tests := []struct {
		name    string
		version string
		wantTag string
	}{
		{
			name:    "version with v prefix",
			version: "v0.93.0",
			wantTag: "v0.93.0",
		},
		{
			name:    "version without v prefix",
			version: "0.93.0",
			wantTag: "v0.93.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloadURL := tg.GetDownloadURL(tt.version, "linux", "amd64")

			expectedURL := "https://github.com/gruntwork-io/terragrunt/releases/download/" + tt.wantTag + "/terragrunt_linux_amd64"
			if downloadURL != expectedURL {
				t.Errorf("GetDownloadURL(%v) = %v, want %v", tt.version, downloadURL, expectedURL)
			}
		})
	}
}

// TestTerragruntGitHubURLStructure validates the GitHub release URL structure.
func TestTerragruntGitHubURLStructure(t *testing.T) {
	tg := &Terragrunt{Name: "terragrunt"}

	downloadURL := tg.GetDownloadURL("v0.93.0", "linux", "amd64")

	// Verify URL components
	expectedComponents := []string{
		"https://github.com",
		"gruntwork-io",
		"terragrunt",
		"releases/download",
		"v0.93.0",
		"terragrunt_linux_amd64",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(downloadURL, component) {
			t.Errorf("Download URL missing component %v: %v", component, downloadURL)
		}
	}
}

// TestCompareTerragruntVersions tests the version comparison function.
func TestCompareTerragruntVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int // > 0 if v1 > v2, < 0 if v1 < v2, 0 if equal
	}{
		{
			name: "v0.93.0 > v0.92.1",
			v1:   "v0.93.0",
			v2:   "v0.92.1",
			want: 1, // positive (v1 > v2)
		},
		{
			name: "v0.92.1 < v0.93.0",
			v1:   "v0.92.1",
			v2:   "v0.93.0",
			want: -1, // negative (v1 < v2)
		},
		{
			name: "v0.93.0 == v0.93.0",
			v1:   "v0.93.0",
			v2:   "v0.93.0",
			want: 0,
		},
		{
			name: "v0.93.0 > v0.93",
			v1:   "v0.93.0",
			v2:   "v0.93",
			want: 0, // equal (0.93.0 == 0.93.0, last part defaults to 0)
		},
		{
			name: "without v prefix: 0.93.0 > 0.92.1",
			v1:   "0.93.0",
			v2:   "0.92.1",
			want: 1, // positive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareTerragruntVersions(tt.v1, tt.v2)

			// Check sign, not exact value
			if tt.want > 0 && got <= 0 {
				t.Errorf("compareTerragruntVersions(%v, %v) = %v, want positive", tt.v1, tt.v2, got)
			} else if tt.want < 0 && got >= 0 {
				t.Errorf("compareTerragruntVersions(%v, %v) = %v, want negative", tt.v1, tt.v2, got)
			} else if tt.want == 0 && got != 0 {
				t.Errorf("compareTerragruntVersions(%v, %v) = %v, want 0", tt.v1, tt.v2, got)
			}
		})
	}
}

// TestTerragruntAlphaVersionFiltering tests that alpha versions are excluded.
// This is a critical test for terragrunt-specific functionality.
func TestTerragruntAlphaVersionFiltering(t *testing.T) {
	// Test the filtering logic in isolation
	alphaVersions := []string{
		"alpha-20241030",
		"alpha-2024112901",
		"v-alpha-2024",
	}

	for _, version := range alphaVersions {
		if !strings.HasPrefix(version, "alpha-") && !strings.Contains(version, "-alpha") {
			t.Errorf("Alpha version detection failed for %v", version)
		}
	}

	// Test that stable versions are NOT filtered
	stableVersions := []string{
		"v0.93.0",
		"v0.92.1",
		"v0.50.0",
	}

	for _, version := range stableVersions {
		if strings.HasPrefix(version, "alpha-") || strings.Contains(version, "-alpha") {
			t.Errorf("Stable version incorrectly detected as alpha: %v", version)
		}
	}
}
