package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtractZip verifies ZIP archive extraction functionality.
func TestExtractZip(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (zipPath, destDir string, cleanup func())
		wantErr     bool
		errContains string
		verify      func(t *testing.T, destDir string)
	}{
		{
			name: "extract valid zip with single binary",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				zipPath := filepath.Join(tmpDir, "test.zip")
				destDir := filepath.Join(tmpDir, "extracted")

				// Create a ZIP file with a binary
				createTestZip(t, zipPath, map[string][]byte{
					"terraform": []byte("fake terraform binary"),
				})

				return zipPath, destDir, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, destDir string) {
				// Check binary exists
				binaryPath := filepath.Join(destDir, "terraform")
				if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
					t.Errorf("binary not extracted: %v", err)
				}

				// Check permissions (should be executable)
				info, err := os.Stat(binaryPath)
				if err != nil {
					t.Fatalf("failed to stat binary: %v", err)
				}
				if info.Mode().Perm()&0111 == 0 {
					t.Errorf("binary is not executable: %o", info.Mode().Perm())
				}
			},
		},
		{
			name: "extract zip with multiple files",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				zipPath := filepath.Join(tmpDir, "test.zip")
				destDir := filepath.Join(tmpDir, "extracted")

				// Create ZIP with multiple files
				createTestZip(t, zipPath, map[string][]byte{
					"terraform": []byte("binary"),
					"LICENSE":   []byte("MIT License"),
					"README.md": []byte("# Terraform"),
				})

				return zipPath, destDir, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, destDir string) {
				expectedFiles := []string{"terraform", "LICENSE", "README.md"}
				for _, file := range expectedFiles {
					path := filepath.Join(destDir, file)
					if _, err := os.Stat(path); os.IsNotExist(err) {
						t.Errorf("file %s not extracted", file)
					}
				}
			},
		},
		{
			name: "fail when zip file does not exist",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				zipPath := filepath.Join(tmpDir, "nonexistent.zip")
				destDir := filepath.Join(tmpDir, "extracted")

				return zipPath, destDir, func() {}
			},
			wantErr:     true,
			errContains: "no such file",
		},
		{
			name: "prevent path traversal with .. sequences",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				zipPath := filepath.Join(tmpDir, "malicious.zip")
				destDir := filepath.Join(tmpDir, "extracted")

				// Create ZIP with path traversal attempt
				createTestZip(t, zipPath, map[string][]byte{
					"../../../etc/passwd": []byte("malicious"),
				})

				return zipPath, destDir, func() {}
			},
			wantErr:     true,
			errContains: "Path traversal",
		},
		{
			name: "create destination directory if it does not exist",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				zipPath := filepath.Join(tmpDir, "test.zip")
				destDir := filepath.Join(tmpDir, "new-dir", "extracted")

				createTestZip(t, zipPath, map[string][]byte{
					"terraform": []byte("binary"),
				})

				return zipPath, destDir, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, destDir string) {
				if _, err := os.Stat(destDir); os.IsNotExist(err) {
					t.Errorf("destination directory not created")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zipPath, destDir, cleanup := tt.setup(t)
			defer cleanup()

			err := ExtractZip(zipPath, destDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractZip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, destDir)
			}
		})
	}
}

// TestExtractTarGz verifies tar.gz archive extraction functionality.
func TestExtractTarGz(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (tarGzPath, destDir string, cleanup func())
		wantErr     bool
		errContains string
		verify      func(t *testing.T, destDir string)
	}{
		{
			name: "extract valid tar.gz with single binary",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				tarGzPath := filepath.Join(tmpDir, "test.tar.gz")
				destDir := filepath.Join(tmpDir, "extracted")

				// Create tar.gz with binary
				createTestTarGz(t, tarGzPath, map[string][]byte{
					"tofu": []byte("fake tofu binary"),
				})

				return tarGzPath, destDir, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, destDir string) {
				binaryPath := filepath.Join(destDir, "tofu")
				if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
					t.Errorf("binary not extracted: %v", err)
				}

				// Check executable permissions
				info, err := os.Stat(binaryPath)
				if err != nil {
					t.Fatalf("failed to stat binary: %v", err)
				}
				if info.Mode().Perm()&0111 == 0 {
					t.Errorf("binary is not executable: %o", info.Mode().Perm())
				}
			},
		},
		{
			name: "extract tar.gz with nested directory",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				tarGzPath := filepath.Join(tmpDir, "test.tar.gz")
				destDir := filepath.Join(tmpDir, "extracted")

				// Create tar.gz with nested structure
				createTestTarGz(t, tarGzPath, map[string][]byte{
					"bin/tofu":    []byte("binary"),
					"docs/README": []byte("readme"),
				})

				return tarGzPath, destDir, func() {}
			},
			wantErr: false,
			verify: func(t *testing.T, destDir string) {
				binaryPath := filepath.Join(destDir, "bin", "tofu")
				docsPath := filepath.Join(destDir, "docs", "README")

				if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
					t.Errorf("binary not extracted: %v", err)
				}
				if _, err := os.Stat(docsPath); os.IsNotExist(err) {
					t.Errorf("docs not extracted: %v", err)
				}
			},
		},
		{
			name: "fail when tar.gz file does not exist",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				tarGzPath := filepath.Join(tmpDir, "nonexistent.tar.gz")
				destDir := filepath.Join(tmpDir, "extracted")

				return tarGzPath, destDir, func() {}
			},
			wantErr:     true,
			errContains: "no such file",
		},
		{
			name: "prevent path traversal in tar.gz",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				tarGzPath := filepath.Join(tmpDir, "malicious.tar.gz")
				destDir := filepath.Join(tmpDir, "extracted")

				// Create tar.gz with path traversal
				createTestTarGz(t, tarGzPath, map[string][]byte{
					"../../../etc/passwd": []byte("malicious"),
				})

				return tarGzPath, destDir, func() {}
			},
			wantErr:     true,
			errContains: "Path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tarGzPath, destDir, cleanup := tt.setup(t)
			defer cleanup()

			err := ExtractTarGz(tarGzPath, destDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTarGz() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, destDir)
			}
		})
	}
}

// Helper function to create a test ZIP file
func createTestZip(t *testing.T, zipPath string, files map[string][]byte) {
	t.Helper()

	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for name, content := range files {
		writer, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry %s: %v", name, err)
		}
		if _, err := writer.Write(content); err != nil {
			t.Fatalf("failed to write zip entry %s: %v", name, err)
		}
	}
}

// Helper function to create a test tar.gz file
func createTestTarGz(t *testing.T, tarGzPath string, files map[string][]byte) {
	t.Helper()

	tarGzFile, err := os.Create(tarGzPath)
	if err != nil {
		t.Fatalf("failed to create tar.gz file: %v", err)
	}
	defer tarGzFile.Close()

	gzipWriter := gzip.NewWriter(tarGzFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for name, content := range files {
		// Create parent directories in tar if name contains /
		if strings.Contains(name, "/") {
			dir := filepath.Dir(name)
			header := &tar.Header{
				Name:     dir + "/",
				Mode:     0755,
				Typeflag: tar.TypeDir,
			}
			if err := tarWriter.WriteHeader(header); err != nil {
				t.Fatalf("failed to write tar directory header: %v", err)
			}
		}

		// Write file
		header := &tar.Header{
			Name: name,
			Mode: 0755,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("failed to write tar header for %s: %v", name, err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			t.Fatalf("failed to write tar content for %s: %v", name, err)
		}
	}
}

// TestValidateExtractPath verifies path validation for extraction.
func TestValidateExtractPath(t *testing.T) {
	tests := []struct {
		name        string
		destDir     string
		targetPath  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid path within destDir",
			destDir:    "/home/user/.binarius/tools/terraform/v1.6.0",
			targetPath: "/home/user/.binarius/tools/terraform/v1.6.0/terraform",
			wantErr:    false,
		},
		{
			name:        "reject path traversal with ..",
			destDir:     "/home/user/.binarius/tools/terraform/v1.6.0",
			targetPath:  "/home/user/.binarius/tools/terraform/../../../etc/passwd",
			wantErr:     true,
			errContains: "Path traversal",
		},
		{
			name:        "reject absolute path outside destDir",
			destDir:     "/home/user/.binarius/tools/terraform/v1.6.0",
			targetPath:  "/etc/passwd",
			wantErr:     true,
			errContains: "Path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExtractPath(tt.destDir, tt.targetPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateExtractPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}
