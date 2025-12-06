package integration

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nixknight/binarius/pkg/config"
	"github.com/nixknight/binarius/pkg/installer"
	"github.com/nixknight/binarius/pkg/paths"
	"github.com/nixknight/binarius/pkg/tools"
)

// TestInstallWorkflow verifies the complete installation workflow.
func TestInstallWorkflow(t *testing.T) {
	// Setup temporary home directory
	tmpHome := t.TempDir()

	t.Setenv("HOME", tmpHome)

	// Initialize binarius structure
	if err := runInit(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// Get paths
	cacheDir, err := paths.CacheDir()
	if err != nil {
		t.Fatalf("failed to get cache dir: %v", err)
	}
	toolsDir, err := paths.ToolsDir()
	if err != nil {
		t.Fatalf("failed to get tools dir: %v", err)
	}
	binariusHome, err := paths.BinariusHome()
	if err != nil {
		t.Fatalf("failed to get binarius home: %v", err)
	}

	// Create mock HTTP server for terraform download
	mockBinary := []byte("fake terraform binary content for testing")
	mockZip := createMockZip(t, "terraform", mockBinary)
	mockChecksum := computeSHA256(mockZip)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip":
			w.Header().Set("Content-Type", "application/zip")
			if _, err := w.Write(mockZip); err != nil {
				t.Errorf("failed to write mock response: %v", err)
			}
		case "/terraform/1.6.0/terraform_1.6.0_SHA256SUMS":
			checksumContent := fmt.Sprintf("%s  terraform_1.6.0_linux_amd64.zip\n", mockChecksum)
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(checksumContent)); err != nil {
				t.Errorf("failed to write mock response: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create mock terraform tool that uses test server
	mockTerraform := &mockTool{
		name:          "terraform",
		downloadURL:   server.URL + "/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip",
		checksumURL:   server.URL + "/terraform/1.6.0/terraform_1.6.0_SHA256SUMS",
		binaryName:    "terraform",
		archiveFormat: "zip",
	}

	// Run installation workflow
	t.Run("download phase", func(t *testing.T) {
		cachePath := filepath.Join(cacheDir, "terraform-v1.6.0.zip")

		err := installer.Download(mockTerraform.downloadURL, cachePath)
		if err != nil {
			t.Fatalf("Download() error = %v", err)
		}

		// Verify file downloaded
		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			t.Errorf("downloaded file does not exist at %s", cachePath)
		}
	})

	t.Run("checksum verification phase", func(t *testing.T) {
		cachePath := filepath.Join(cacheDir, "terraform-v1.6.0.zip")

		err := installer.VerifyChecksum(cachePath, mockChecksum)
		if err != nil {
			t.Fatalf("VerifyChecksum() error = %v", err)
		}
	})

	t.Run("extraction phase", func(t *testing.T) {
		cachePath := filepath.Join(cacheDir, "terraform-v1.6.0.zip")
		destDir := filepath.Join(toolsDir, "terraform", "v1.6.0")

		err := installer.ExtractZip(cachePath, destDir)
		if err != nil {
			t.Fatalf("ExtractZip() error = %v", err)
		}

		// Verify binary extracted
		binaryPath := filepath.Join(destDir, "terraform")
		info, err := os.Stat(binaryPath)
		if os.IsNotExist(err) {
			t.Errorf("binary not extracted at %s", binaryPath)
			return
		}
		if err != nil {
			t.Fatalf("failed to stat binary: %v", err)
		}

		// Verify executable permissions
		if info.Mode().Perm()&0111 == 0 {
			t.Errorf("binary is not executable: permissions = %o", info.Mode().Perm())
		}
	})

	t.Run("registry update phase", func(t *testing.T) {
		registryPath := filepath.Join(binariusHome, "installation.json")
		registry, err := config.LoadRegistry(registryPath)
		if err != nil {
			t.Fatalf("LoadRegistry() error = %v", err)
		}

		// Add terraform version to registry
		binaryPath := filepath.Join(toolsDir, "terraform", "v1.6.0", "terraform")
		info, _ := os.Stat(binaryPath)

		registry.AddVersion("terraform", "v1.6.0", config.ToolVersion{
			BinaryPath: binaryPath,
			SourceURL:  mockTerraform.downloadURL,
			SizeBytes:  info.Size(),
		})

		if err := config.SaveRegistry(registry, registryPath); err != nil {
			t.Fatalf("SaveRegistry() error = %v", err)
		}

		// Verify registry updated
		reloaded, err := config.LoadRegistry(registryPath)
		if err != nil {
			t.Fatalf("failed to reload registry: %v", err)
		}

		if _, exists := reloaded.Tools["terraform"]; !exists {
			t.Error("terraform not found in registry")
		}

		if _, exists := reloaded.Tools["terraform"]["v1.6.0"]; !exists {
			t.Error("terraform v1.6.0 not found in registry")
		}
	})

	t.Run("full workflow integration", func(t *testing.T) {
		// This will test the full Install() function once implemented
		// For now, we verify the pieces work together
		t.Skip("full workflow will be tested with actual installer.Install() function")
	})
}

// TestInstallChecksumMismatch verifies installation fails on checksum mismatch.
func TestInstallChecksumMismatch(t *testing.T) {
	tmpHome := t.TempDir()

	t.Setenv("HOME", tmpHome)

	if err := runInit(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	cacheDir, err := paths.CacheDir()
	if err != nil {
		t.Fatalf("failed to get cache dir: %v", err)
	}

	// Create mock binary with known content
	mockBinary := []byte("fake terraform binary")
	mockZip := createMockZip(t, "terraform", mockBinary)

	// Create server that returns ZIP
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/terraform.zip":
			if _, err := w.Write(mockZip); err != nil {
				t.Errorf("failed to write mock response: %v", err)
			}
		case "/terraform_SHA256SUMS":
			// Return WRONG checksum
			wrongChecksum := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			checksumContent := fmt.Sprintf("%s  terraform_1.6.0_linux_amd64.zip\n", wrongChecksum)
			if _, err := w.Write([]byte(checksumContent)); err != nil {
				t.Errorf("failed to write mock response: %v", err)
			}
		}
	}))
	defer server.Close()

	// Download
	cachePath := filepath.Join(cacheDir, "terraform-test.zip")
	err = installer.Download(server.URL+"/terraform.zip", cachePath)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	// Verify with wrong checksum
	wrongChecksum := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	err = installer.VerifyChecksum(cachePath, wrongChecksum)
	if err == nil {
		t.Error("VerifyChecksum() expected error on mismatch, got nil")
	}
}

// TestInstallNetworkError verifies installation handles network errors gracefully.
func TestInstallNetworkError(t *testing.T) {
	tmpHome := t.TempDir()

	t.Setenv("HOME", tmpHome)

	if err := runInit(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	cacheDir, err := paths.CacheDir()
	if err != nil {
		t.Fatalf("failed to get cache dir: %v", err)
	}

	// Create server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	// Attempt download
	cachePath := filepath.Join(cacheDir, "terraform-fail.zip")
	err = installer.Download(server.URL+"/terraform.zip", cachePath)
	if err == nil {
		t.Error("Download() expected error on server error, got nil")
	}
}

// TestInstallAlreadyInstalled verifies handling of already installed versions.
func TestInstallAlreadyInstalled(t *testing.T) {
	tmpHome := t.TempDir()

	t.Setenv("HOME", tmpHome)

	if err := runInit(); err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	binariusHome, err := paths.BinariusHome()
	if err != nil {
		t.Fatalf("failed to get binarius home: %v", err)
	}

	toolsDir, err := paths.ToolsDir()
	if err != nil {
		t.Fatalf("failed to get tools dir: %v", err)
	}

	// Add terraform v1.6.0 to registry
	registryPath := filepath.Join(binariusHome, "installation.json")
	registry, err := config.LoadRegistry(registryPath)
	if err != nil {
		t.Fatalf("LoadRegistry() error = %v", err)
	}

	binaryPath := filepath.Join(toolsDir, "terraform", "v1.6.0", "terraform")
	registry.AddVersion("terraform", "v1.6.0", config.ToolVersion{
		BinaryPath: binaryPath,
		SourceURL:  "https://example.com/terraform.zip",
		SizeBytes:  12345,
	})

	if err := config.SaveRegistry(registry, registryPath); err != nil {
		t.Fatalf("SaveRegistry() error = %v", err)
	}

	// Check if version is installed
	reloaded, err := config.LoadRegistry(registryPath)
	if err != nil {
		t.Fatalf("failed to reload registry: %v", err)
	}

	isInstalled := reloaded.IsInstalled("terraform", "v1.6.0")
	if !isInstalled {
		t.Error("expected terraform v1.6.0 to be marked as installed")
	}

	// Installing again should be idempotent or return error
	// This behavior will be defined in installer.Install()
}

// Helper functions

// createMockZip creates a mock ZIP file containing a binary.
func createMockZip(t *testing.T, binaryName string, content []byte) []byte {
	t.Helper()

	// Create in-memory ZIP
	tmpFile, err := os.CreateTemp("", "mock-*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()

	zipWriter := zip.NewWriter(tmpFile)
	writer, err := zipWriter.Create(binaryName)
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := writer.Write(content); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	// Read ZIP into memory
	zipData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read zip file: %v", err)
	}

	return zipData
}

// computeSHA256 computes SHA256 hash of data.
func computeSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// mockTool is a mock implementation of the Tool interface for testing.
type mockTool struct {
	name          string
	downloadURL   string
	checksumURL   string
	binaryName    string
	archiveFormat string
}

func (m *mockTool) GetName() string {
	return m.name
}

func (m *mockTool) GetDownloadURL(version, os, arch string) string {
	return m.downloadURL
}

func (m *mockTool) GetChecksumURL(version, os, arch string) string {
	return m.checksumURL
}

func (m *mockTool) ListVersions() ([]string, error) {
	return []string{"v1.6.0", "v1.5.7"}, nil
}

func (m *mockTool) GetBinaryName() string {
	return m.binaryName
}

func (m *mockTool) GetArchiveFormat() string {
	return m.archiveFormat
}

func (m *mockTool) SupportedArchs() []string {
	return []string{"amd64", "arm64"}
}

// Ensure mockTool implements tools.Tool
var _ tools.Tool = (*mockTool)(nil)
