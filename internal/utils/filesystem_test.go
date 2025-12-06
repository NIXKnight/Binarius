package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		base    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid path within base",
			base:    tempDir,
			target:  filepath.Join(tempDir, "subdir", "file.txt"),
			wantErr: false,
		},
		{
			name:    "valid path - same as base",
			base:    tempDir,
			target:  tempDir,
			wantErr: false,
		},
		{
			name:    "invalid path - parent directory traversal",
			base:    tempDir,
			target:  filepath.Join(tempDir, "..", "outside"),
			wantErr: true,
		},
		{
			name:    "invalid path - absolute path outside base",
			base:    tempDir,
			target:  "/tmp/outside",
			wantErr: true,
		},
		{
			name:    "valid path with dots in name",
			base:    tempDir,
			target:  filepath.Join(tempDir, "file.tar.gz"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.base, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T, string) string // returns path to test
		perm    os.FileMode
		wantErr bool
	}{
		{
			name: "create new directory",
			setup: func(t *testing.T, base string) string {
				return filepath.Join(base, "newdir")
			},
			perm:    0755,
			wantErr: false,
		},
		{
			name: "create nested directories",
			setup: func(t *testing.T, base string) string {
				return filepath.Join(base, "parent", "child", "grandchild")
			},
			perm:    0755,
			wantErr: false,
		},
		{
			name: "existing directory - no error",
			setup: func(t *testing.T, base string) string {
				dir := filepath.Join(base, "existing")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				return dir
			},
			perm:    0755,
			wantErr: false,
		},
		{
			name: "path exists as file - error",
			setup: func(t *testing.T, base string) string {
				file := filepath.Join(base, "file.txt")
				if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
				return file
			},
			perm:    0755,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			path := tt.setup(t, tempDir)

			err := EnsureDir(path, tt.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error expected, verify directory was created
			if !tt.wantErr {
				info, err := os.Stat(path)
				if err != nil {
					t.Errorf("EnsureDir() created directory doesn't exist: %v", err)
				}
				if !info.IsDir() {
					t.Errorf("EnsureDir() created path is not a directory")
				}
			}
		})
	}
}

func TestIsWritable(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, string) string // returns path to test
		want  bool
	}{
		{
			name: "writable directory",
			setup: func(t *testing.T, base string) string {
				dir := filepath.Join(base, "writable")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				return dir
			},
			want: true,
		},
		{
			name: "writable file",
			setup: func(t *testing.T, base string) string {
				file := filepath.Join(base, "writable.txt")
				if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
				return file
			},
			want: true,
		},
		{
			name: "non-existent path in writable directory",
			setup: func(t *testing.T, base string) string {
				return filepath.Join(base, "nonexistent")
			},
			want: true,
		},
		{
			name: "read-only directory",
			setup: func(t *testing.T, base string) string {
				dir := filepath.Join(base, "readonly")
				if err := os.MkdirAll(dir, 0555); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				return dir
			},
			want: false,
		},
		{
			name: "read-only file",
			setup: func(t *testing.T, base string) string {
				file := filepath.Join(base, "readonly.txt")
				if err := os.WriteFile(file, []byte("test"), 0444); err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
				return file
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			path := tt.setup(t, tempDir)

			got := IsWritable(path)
			if got != tt.want {
				t.Errorf("IsWritable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePath_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		base    string
		target  string
		wantErr bool
	}{
		{
			name:    "multiple parent traversals",
			base:    tempDir,
			target:  filepath.Join(tempDir, "a", "b", "..", "..", "..", "outside"),
			wantErr: true,
		},
		{
			name:    "hidden directory traversal",
			base:    filepath.Join(tempDir, "base"),
			target:  filepath.Join(tempDir, "base", "..", "outside"),
			wantErr: true,
		},
		{
			name:    "deeply nested valid path",
			base:    tempDir,
			target:  filepath.Join(tempDir, "a", "b", "c", "d", "e", "f"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.base, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
