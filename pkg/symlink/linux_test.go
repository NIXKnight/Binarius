package symlink

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreate verifies symlink creation functionality.
func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (source, target string, cleanup func())
		wantErr     bool
		errContains string
	}{
		{
			name: "create new symlink successfully",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				source := filepath.Join(tmpDir, "source-binary")
				target := filepath.Join(tmpDir, "target-link")

				// Create source file
				if err := os.WriteFile(source, []byte("binary"), 0755); err != nil {
					t.Fatalf("failed to create source file: %v", err)
				}

				return source, target, func() {}
			},
			wantErr: false,
		},
		{
			name: "fail when target already exists",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				source := filepath.Join(tmpDir, "source-binary")
				target := filepath.Join(tmpDir, "target-link")

				// Create source and existing target
				if err := os.WriteFile(source, []byte("binary"), 0755); err != nil {
					t.Fatalf("failed to create source file: %v", err)
				}
				if err := os.Symlink(source, target); err != nil {
					t.Fatalf("failed to create existing symlink: %v", err)
				}

				return source, target, func() {}
			},
			wantErr:     true,
			errContains: "file exists",
		},
		{
			name: "fail when target directory does not exist",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				source := filepath.Join(tmpDir, "source-binary")
				target := filepath.Join(tmpDir, "nonexistent", "target-link")

				// Create source only
				if err := os.WriteFile(source, []byte("binary"), 0755); err != nil {
					t.Fatalf("failed to create source file: %v", err)
				}

				return source, target, func() {}
			},
			wantErr:     true,
			errContains: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, target, cleanup := tt.setup(t)
			defer cleanup()

			manager := &Manager{}
			err := manager.Create(source, target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if err.Error() == "" {
					t.Errorf("Create() error message is empty, want to contain %q", tt.errContains)
				}
				// Note: actual error checking will happen after implementation
				return
			}

			// If success, verify symlink points to source
			if !tt.wantErr {
				link, err := os.Readlink(target)
				if err != nil {
					t.Errorf("failed to read symlink: %v", err)
					return
				}
				if link != source {
					t.Errorf("symlink points to %q, want %q", link, source)
				}
			}
		})
	}
}

// TestUpdate verifies atomic symlink update functionality.
func TestUpdate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (source, target string, cleanup func())
		wantErr     bool
		errContains string
	}{
		{
			name: "update existing symlink atomically",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				oldSource := filepath.Join(tmpDir, "old-binary")
				newSource := filepath.Join(tmpDir, "new-binary")
				target := filepath.Join(tmpDir, "target-link")

				// Create both sources
				if err := os.WriteFile(oldSource, []byte("old"), 0755); err != nil {
					t.Fatalf("failed to create old source: %v", err)
				}
				if err := os.WriteFile(newSource, []byte("new"), 0755); err != nil {
					t.Fatalf("failed to create new source: %v", err)
				}

				// Create initial symlink
				if err := os.Symlink(oldSource, target); err != nil {
					t.Fatalf("failed to create initial symlink: %v", err)
				}

				return newSource, target, func() {}
			},
			wantErr: false,
		},
		{
			name: "create new symlink if target does not exist",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				source := filepath.Join(tmpDir, "source-binary")
				target := filepath.Join(tmpDir, "target-link")

				// Create source only
				if err := os.WriteFile(source, []byte("binary"), 0755); err != nil {
					t.Fatalf("failed to create source: %v", err)
				}

				return source, target, func() {}
			},
			wantErr: false,
		},
		{
			name: "fail when target directory does not exist",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				source := filepath.Join(tmpDir, "source-binary")
				target := filepath.Join(tmpDir, "nonexistent", "target-link")

				// Create source only
				if err := os.WriteFile(source, []byte("binary"), 0755); err != nil {
					t.Fatalf("failed to create source: %v", err)
				}

				return source, target, func() {}
			},
			wantErr:     true,
			errContains: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, target, cleanup := tt.setup(t)
			defer cleanup()

			manager := &Manager{}
			err := manager.Update(source, target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			// If success, verify symlink points to new source
			if !tt.wantErr {
				link, err := os.Readlink(target)
				if err != nil {
					t.Errorf("failed to read symlink: %v", err)
					return
				}
				if link != source {
					t.Errorf("symlink points to %q, want %q", link, source)
				}
			}
		})
	}
}

// TestRemove verifies symlink removal functionality.
func TestRemove(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (target string, cleanup func())
		wantErr     bool
		errContains string
	}{
		{
			name: "remove existing symlink successfully",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				source := filepath.Join(tmpDir, "source-binary")
				target := filepath.Join(tmpDir, "target-link")

				// Create source and symlink
				if err := os.WriteFile(source, []byte("binary"), 0755); err != nil {
					t.Fatalf("failed to create source: %v", err)
				}
				if err := os.Symlink(source, target); err != nil {
					t.Fatalf("failed to create symlink: %v", err)
				}

				return target, func() {}
			},
			wantErr: false,
		},
		{
			name: "succeed when symlink does not exist (idempotent)",
			setup: func(t *testing.T) (string, func()) {
				tmpDir := t.TempDir()
				target := filepath.Join(tmpDir, "nonexistent-link")
				return target, func() {}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, cleanup := tt.setup(t)
			defer cleanup()

			manager := &Manager{}
			err := manager.Remove(target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify symlink is removed
			if !tt.wantErr {
				if _, err := os.Lstat(target); !os.IsNotExist(err) {
					t.Errorf("symlink still exists after Remove()")
				}
			}
		})
	}
}

// TestVerify verifies symlink verification functionality.
func TestVerify(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (target, expectedSource string, cleanup func())
		wantErr     bool
		errContains string
	}{
		{
			name: "verify correct symlink",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				source := filepath.Join(tmpDir, "source-binary")
				target := filepath.Join(tmpDir, "target-link")

				// Create source and symlink
				if err := os.WriteFile(source, []byte("binary"), 0755); err != nil {
					t.Fatalf("failed to create source: %v", err)
				}
				if err := os.Symlink(source, target); err != nil {
					t.Fatalf("failed to create symlink: %v", err)
				}

				return target, source, func() {}
			},
			wantErr: false,
		},
		{
			name: "fail when symlink points to different source",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				actualSource := filepath.Join(tmpDir, "actual-source")
				expectedSource := filepath.Join(tmpDir, "expected-source")
				target := filepath.Join(tmpDir, "target-link")

				// Create sources and symlink to actual
				if err := os.WriteFile(actualSource, []byte("actual"), 0755); err != nil {
					t.Fatalf("failed to create actual source: %v", err)
				}
				if err := os.WriteFile(expectedSource, []byte("expected"), 0755); err != nil {
					t.Fatalf("failed to create expected source: %v", err)
				}
				if err := os.Symlink(actualSource, target); err != nil {
					t.Fatalf("failed to create symlink: %v", err)
				}

				return target, expectedSource, func() {}
			},
			wantErr:     true,
			errContains: "points to",
		},
		{
			name: "fail when target does not exist",
			setup: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				target := filepath.Join(tmpDir, "nonexistent-link")
				source := filepath.Join(tmpDir, "source")

				return target, source, func() {}
			},
			wantErr:     true,
			errContains: "no such file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, expectedSource, cleanup := tt.setup(t)
			defer cleanup()

			manager := &Manager{}
			err := manager.Verify(target, expectedSource)

			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
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
