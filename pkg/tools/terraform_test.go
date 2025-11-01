package tools

import (
	"strings"
	"testing"
)

// TestTerraformGetName verifies the terraform tool name.
func TestTerraformGetName(t *testing.T) {
	tf := &Terraform{Name: "terraform"}
	got := tf.GetName()
	want := "terraform"

	if got != want {
		t.Errorf("GetName() = %q, want %q", got, want)
	}
}

// TestTerraformGetDownloadURL verifies download URL generation.
func TestTerraformGetDownloadURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{
			name:    "terraform v1.6.0 linux amd64",
			version: "v1.6.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip",
		},
		{
			name:    "terraform v1.6.0 linux arm64",
			version: "v1.6.0",
			os:      "linux",
			arch:    "arm64",
			want:    "https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_arm64.zip",
		},
		{
			name:    "terraform v1.5.7 linux amd64",
			version: "v1.5.7",
			os:      "linux",
			arch:    "amd64",
			want:    "https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip",
		},
		{
			name:    "version without v prefix",
			version: "1.6.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := &Terraform{Name: "terraform"}
			got := tf.GetDownloadURL(tt.version, tt.os, tt.arch)

			if got != tt.want {
				t.Errorf("GetDownloadURL(%q, %q, %q) = %q, want %q",
					tt.version, tt.os, tt.arch, got, tt.want)
			}

			// Verify HTTPS
			if !strings.HasPrefix(got, "https://") {
				t.Errorf("GetDownloadURL() must return HTTPS URL, got: %s", got)
			}
		})
	}
}

// TestTerraformGetChecksumURL verifies checksum URL generation.
func TestTerraformGetChecksumURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{
			name:    "terraform v1.6.0 checksum",
			version: "v1.6.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_SHA256SUMS",
		},
		{
			name:    "terraform v1.5.7 checksum",
			version: "v1.5.7",
			os:      "linux",
			arch:    "arm64",
			want:    "https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_SHA256SUMS",
		},
		{
			name:    "version without v prefix",
			version: "1.6.0",
			os:      "linux",
			arch:    "amd64",
			want:    "https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_SHA256SUMS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := &Terraform{Name: "terraform"}
			got := tf.GetChecksumURL(tt.version, tt.os, tt.arch)

			if got != tt.want {
				t.Errorf("GetChecksumURL(%q, %q, %q) = %q, want %q",
					tt.version, tt.os, tt.arch, got, tt.want)
			}

			// Verify HTTPS
			if !strings.HasPrefix(got, "https://") {
				t.Errorf("GetChecksumURL() must return HTTPS URL, got: %s", got)
			}
		})
	}
}

// TestTerraformGetBinaryName verifies the binary name.
func TestTerraformGetBinaryName(t *testing.T) {
	tf := &Terraform{Name: "terraform"}
	got := tf.GetBinaryName()
	want := "terraform"

	if got != want {
		t.Errorf("GetBinaryName() = %q, want %q", got, want)
	}
}

// TestTerraformGetArchiveFormat verifies the archive format.
func TestTerraformGetArchiveFormat(t *testing.T) {
	tf := &Terraform{Name: "terraform"}
	got := tf.GetArchiveFormat()
	want := "zip"

	if got != want {
		t.Errorf("GetArchiveFormat() = %q, want %q", got, want)
	}
}

// TestTerraformSupportedArchs verifies supported architectures.
func TestTerraformSupportedArchs(t *testing.T) {
	tf := &Terraform{Name: "terraform"}
	got := tf.SupportedArchs()

	// Check that it includes amd64 and arm64
	expectedArchs := []string{"amd64", "arm64"}
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

	// Check that there are at least 2 architectures
	if len(got) < 2 {
		t.Errorf("SupportedArchs() should return at least 2 architectures, got %d", len(got))
	}
}

// TestTerraformListVersions verifies version listing functionality.
// This test will be skipped initially as it requires network access.
func TestTerraformListVersions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	tf := &Terraform{Name: "terraform"}
	versions, err := tf.ListVersions()

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

		// Only check first 5 to keep test fast
		if i >= 4 {
			break
		}
	}

	// Verify descending order (newest first)
	// For terraform, v1.6.0 should come before v1.5.0
	// This is a simple check - full semantic version comparison would be better
	if len(versions) >= 2 {
		// Just verify they look like version strings
		first := versions[0]
		if !strings.Contains(first, ".") {
			t.Errorf("ListVersions()[0] = %q, doesn't look like a version", first)
		}
	}
}

// TestTerraformInterface verifies Terraform implements Tool interface.
func TestTerraformInterface(t *testing.T) {
	var _ Tool = (*Terraform)(nil)
}

// TestTerraformRegistration verifies terraform can be registered.
func TestTerraformRegistration(t *testing.T) {
	// Clear registry for clean test
	Clear()
	defer Clear()

	tf := &Terraform{Name: "terraform"}
	err := Register("terraform", tf)

	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}

	// Verify retrieval
	retrieved, err := Get("terraform")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if retrieved.GetName() != "terraform" {
		t.Errorf("retrieved tool name = %q, want %q", retrieved.GetName(), "terraform")
	}
}

// TestTerraformURLFormat verifies URL format correctness.
func TestTerraformURLFormat(t *testing.T) {
	tf := &Terraform{Name: "terraform"}

	// Test that URLs don't have double slashes (except after https://)
	url := tf.GetDownloadURL("v1.6.0", "linux", "amd64")

	// Remove https:// prefix for checking
	urlWithoutScheme := strings.TrimPrefix(url, "https://")
	if strings.Contains(urlWithoutScheme, "//") {
		t.Errorf("GetDownloadURL() contains double slashes: %s", url)
	}

	checksumURL := tf.GetChecksumURL("v1.6.0", "linux", "amd64")
	checksumWithoutScheme := strings.TrimPrefix(checksumURL, "https://")
	if strings.Contains(checksumWithoutScheme, "//") {
		t.Errorf("GetChecksumURL() contains double slashes: %s", checksumURL)
	}
}
