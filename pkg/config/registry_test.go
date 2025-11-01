package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if registry.Tools == nil {
		t.Error("NewRegistry() Tools map is nil")
	}

	if len(registry.Tools) != 0 {
		t.Errorf("NewRegistry() should have empty Tools map, got %d items", len(registry.Tools))
	}
}

func TestLoadRegistry(t *testing.T) {
	tests := []struct {
		name      string
		setupFile func(string) // Function to create test file
		wantErr   bool
		validate  func(*testing.T, *Registry)
	}{
		{
			name: "valid registry file",
			setupFile: func(path string) {
				registry := Registry{
					Tools: map[string]map[string]ToolVersion{
						"terraform": {
							"v1.6.0": {
								ToolName:     "terraform",
								Version:      "v1.6.0",
								BinaryPath:   "/home/user/.binarius/tools/terraform/v1.6.0/terraform",
								InstalledAt:  time.Now(),
								SizeBytes:    25678901,
								SourceURL:    "https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip",
								Checksum:     "abc123def456",
								Architecture: "amd64",
								Status:       "complete",
							},
						},
					},
				}
				jsonData, _ := json.MarshalIndent(registry, "", "  ")
				os.WriteFile(path, jsonData, 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, r *Registry) {
				if r.Tools["terraform"] == nil {
					t.Error("LoadRegistry() missing terraform tool")
					return
				}
				if r.Tools["terraform"]["v1.6.0"].BinaryPath == "" {
					t.Error("LoadRegistry() missing terraform v1.6.0")
					return
				}
				tv := r.Tools["terraform"]["v1.6.0"]
				if tv.ToolName != "terraform" {
					t.Errorf("LoadRegistry() ToolName = %q, want terraform", tv.ToolName)
				}
				if tv.Status != "complete" {
					t.Errorf("LoadRegistry() Status = %q, want complete", tv.Status)
				}
			},
		},
		{
			name: "non-existent file returns empty registry",
			setupFile: func(path string) {
				// Don't create file
			},
			wantErr: false,
			validate: func(t *testing.T, r *Registry) {
				if len(r.Tools) != 0 {
					t.Errorf("LoadRegistry() with non-existent file should return empty registry, got %d tools", len(r.Tools))
				}
			},
		},
		{
			name: "empty json object",
			setupFile: func(path string) {
				os.WriteFile(path, []byte("{}"), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, r *Registry) {
				if r.Tools == nil {
					t.Error("LoadRegistry() should initialize Tools map")
				}
			},
		},
		{
			name: "invalid json",
			setupFile: func(path string) {
				os.WriteFile(path, []byte("invalid json"), 0644)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			registryPath := filepath.Join(tempDir, "installation.json")

			tt.setupFile(registryPath)

			registry, err := LoadRegistry(registryPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, registry)
			}
		})
	}
}

func TestSaveRegistry(t *testing.T) {
	tests := []struct {
		name     string
		registry *Registry
		wantErr  bool
	}{
		{
			name: "save valid registry",
			registry: &Registry{
				Tools: map[string]map[string]ToolVersion{
					"terraform": {
						"v1.6.0": {
							ToolName:     "terraform",
							Version:      "v1.6.0",
							BinaryPath:   "/test/path",
							InstalledAt:  time.Now(),
							SizeBytes:    1000,
							SourceURL:    "https://example.com",
							Checksum:     "abc123",
							Architecture: "amd64",
							Status:       "complete",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "save empty registry",
			registry: NewRegistry(),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			registryPath := filepath.Join(tempDir, "installation.json")

			err := SaveRegistry(tt.registry, registryPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created
				if _, err := os.Stat(registryPath); err != nil {
					t.Errorf("SaveRegistry() file not created: %v", err)
				}

				// Verify no temporary file left behind
				tmpPath := registryPath + ".tmp"
				if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
					t.Error("SaveRegistry() temporary file not cleaned up")
				}

				// Verify content can be loaded back
				loaded, err := LoadRegistry(registryPath)
				if err != nil {
					t.Errorf("LoadRegistry() after SaveRegistry() failed: %v", err)
				}

				if len(loaded.Tools) != len(tt.registry.Tools) {
					t.Errorf("SaveRegistry/LoadRegistry tools count mismatch: got %d, want %d",
						len(loaded.Tools), len(tt.registry.Tools))
				}
			}
		})
	}
}

func TestAddVersion(t *testing.T) {
	tests := []struct {
		name     string
		initial  *Registry
		toolName string
		version  string
		tv       ToolVersion
		validate func(*testing.T, *Registry)
	}{
		{
			name:     "add to empty registry",
			initial:  NewRegistry(),
			toolName: "terraform",
			version:  "v1.6.0",
			tv: ToolVersion{
				ToolName:   "terraform",
				Version:    "v1.6.0",
				BinaryPath: "/home/user/.binarius/tools/terraform/v1.6.0/terraform",
				Status:     "complete",
			},
			validate: func(t *testing.T, r *Registry) {
				if r.Tools["terraform"] == nil {
					t.Error("AddVersion() did not create tool entry")
					return
				}
				if r.Tools["terraform"]["v1.6.0"].BinaryPath == "" {
					t.Error("AddVersion() did not add version")
				}
			},
		},
		{
			name: "add to existing tool",
			initial: &Registry{
				Tools: map[string]map[string]ToolVersion{
					"terraform": {
						"v1.5.0": {ToolName: "terraform", Version: "v1.5.0"},
					},
				},
			},
			toolName: "terraform",
			version:  "v1.6.0",
			tv: ToolVersion{
				ToolName: "terraform",
				Version:  "v1.6.0",
				Status:   "complete",
			},
			validate: func(t *testing.T, r *Registry) {
				if len(r.Tools["terraform"]) != 2 {
					t.Errorf("AddVersion() terraform should have 2 versions, got %d", len(r.Tools["terraform"]))
				}
			},
		},
		{
			name:     "add with nil Tools map",
			initial:  &Registry{Tools: nil},
			toolName: "terraform",
			version:  "v1.6.0",
			tv: ToolVersion{
				ToolName:   "terraform",
				Version:    "v1.6.0",
				BinaryPath: "/home/user/.binarius/tools/terraform/v1.6.0/terraform",
			},
			validate: func(t *testing.T, r *Registry) {
				if r.Tools == nil {
					t.Error("AddVersion() should initialize Tools map")
					return
				}
				if r.Tools["terraform"]["v1.6.0"].BinaryPath == "" {
					t.Error("AddVersion() did not add version")
				}
			},
		},
		{
			name: "update existing version",
			initial: &Registry{
				Tools: map[string]map[string]ToolVersion{
					"terraform": {
						"v1.6.0": {Status: "partial"},
					},
				},
			},
			toolName: "terraform",
			version:  "v1.6.0",
			tv: ToolVersion{
				ToolName: "terraform",
				Version:  "v1.6.0",
				Status:   "complete",
			},
			validate: func(t *testing.T, r *Registry) {
				if r.Tools["terraform"]["v1.6.0"].Status != "complete" {
					t.Errorf("AddVersion() should update status, got %q", r.Tools["terraform"]["v1.6.0"].Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := tt.initial
			registry.AddVersion(tt.toolName, tt.version, tt.tv)
			tt.validate(t, registry)
		})
	}
}

func TestRemoveVersion(t *testing.T) {
	tests := []struct {
		name     string
		initial  *Registry
		toolName string
		version  string
		validate func(*testing.T, *Registry)
	}{
		{
			name: "remove existing version",
			initial: &Registry{
				Tools: map[string]map[string]ToolVersion{
					"terraform": {
						"v1.6.0": {ToolName: "terraform", Version: "v1.6.0"},
					},
				},
			},
			toolName: "terraform",
			version:  "v1.6.0",
			validate: func(t *testing.T, r *Registry) {
				if r.Tools["terraform"] != nil {
					t.Error("RemoveVersion() should remove tool entry when last version removed")
				}
			},
		},
		{
			name: "remove one of multiple versions",
			initial: &Registry{
				Tools: map[string]map[string]ToolVersion{
					"terraform": {
						"v1.5.0": {ToolName: "terraform", Version: "v1.5.0"},
						"v1.6.0": {ToolName: "terraform", Version: "v1.6.0"},
					},
				},
			},
			toolName: "terraform",
			version:  "v1.6.0",
			validate: func(t *testing.T, r *Registry) {
				if r.Tools["terraform"] == nil {
					t.Error("RemoveVersion() should not remove tool entry when other versions exist")
					return
				}
				if len(r.Tools["terraform"]) != 1 {
					t.Errorf("RemoveVersion() should leave 1 version, got %d", len(r.Tools["terraform"]))
				}
				if r.Tools["terraform"]["v1.6.0"].BinaryPath != "" {
					t.Error("RemoveVersion() did not remove the specified version")
				}
			},
		},
		{
			name:     "remove from empty registry",
			initial:  NewRegistry(),
			toolName: "terraform",
			version:  "v1.6.0",
			validate: func(t *testing.T, r *Registry) {
				// Should not panic
			},
		},
		{
			name: "remove non-existent version",
			initial: &Registry{
				Tools: map[string]map[string]ToolVersion{
					"terraform": {
						"v1.5.0": {ToolName: "terraform", Version: "v1.5.0"},
					},
				},
			},
			toolName: "terraform",
			version:  "v1.6.0",
			validate: func(t *testing.T, r *Registry) {
				if len(r.Tools["terraform"]) != 1 {
					t.Error("RemoveVersion() should not affect other versions")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := tt.initial
			registry.RemoveVersion(tt.toolName, tt.version)
			tt.validate(t, registry)
		})
	}
}

func TestGetVersion(t *testing.T) {
	registry := &Registry{
		Tools: map[string]map[string]ToolVersion{
			"terraform": {
				"v1.6.0": {
					ToolName:   "terraform",
					Version:    "v1.6.0",
					BinaryPath: "/home/user/.binarius/tools/terraform/v1.6.0/terraform",
					Status:     "complete",
				},
			},
		},
	}

	tests := []struct {
		name     string
		toolName string
		version  string
		wantNil  bool
	}{
		{
			name:     "get existing version",
			toolName: "terraform",
			version:  "v1.6.0",
			wantNil:  false,
		},
		{
			name:     "get non-existent version",
			toolName: "terraform",
			version:  "v1.5.0",
			wantNil:  true,
		},
		{
			name:     "get from non-existent tool",
			toolName: "tofu",
			version:  "v1.6.0",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.GetVersion(tt.toolName, tt.version)
			if (got.BinaryPath == "") != tt.wantNil {
				t.Errorf("GetVersion() nil = %v, wantNil %v", got.BinaryPath == "", tt.wantNil)
			}

			if !tt.wantNil && got.Version != tt.version {
				t.Errorf("GetVersion() version = %q, want %q", got.Version, tt.version)
			}
		})
	}
}

func TestListVersions(t *testing.T) {
	registry := &Registry{
		Tools: map[string]map[string]ToolVersion{
			"terraform": {
				"v1.5.0": {Version: "v1.5.0"},
				"v1.6.0": {Version: "v1.6.0"},
			},
			"tofu": {
				"v1.6.0": {Version: "v1.6.0"},
			},
		},
	}

	tests := []struct {
		name     string
		toolName string
		wantLen  int
	}{
		{
			name:     "list multiple versions",
			toolName: "terraform",
			wantLen:  2,
		},
		{
			name:     "list single version",
			toolName: "tofu",
			wantLen:  1,
		},
		{
			name:     "list non-existent tool",
			toolName: "terragrunt",
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.ListVersions(tt.toolName)
			if len(got) != tt.wantLen {
				t.Errorf("ListVersions() returned %d versions, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestListTools(t *testing.T) {
	tests := []struct {
		name     string
		registry *Registry
		wantLen  int
		wantTool string
	}{
		{
			name: "list multiple tools",
			registry: &Registry{
				Tools: map[string]map[string]ToolVersion{
					"terraform":  {"v1.6.0": {}},
					"tofu":       {"v1.6.0": {}},
					"terragrunt": {"v0.54.0": {}},
				},
			},
			wantLen:  3,
			wantTool: "terraform",
		},
		{
			name: "list single tool",
			registry: &Registry{
				Tools: map[string]map[string]ToolVersion{
					"terraform": {"v1.6.0": {}},
				},
			},
			wantLen:  1,
			wantTool: "terraform",
		},
		{
			name:     "list empty registry",
			registry: NewRegistry(),
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.registry.ListTools()
			if len(got) != tt.wantLen {
				t.Errorf("ListTools() returned %d tools, want %d", len(got), tt.wantLen)
			}

			if tt.wantTool != "" {
				found := false
				for _, tool := range got {
					if tool == tt.wantTool {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ListTools() missing expected tool %q in %v", tt.wantTool, got)
				}
			}
		})
	}
}

func TestHasVersion(t *testing.T) {
	registry := &Registry{
		Tools: map[string]map[string]ToolVersion{
			"terraform": {
				"v1.6.0": {Version: "v1.6.0"},
			},
		},
	}

	tests := []struct {
		name     string
		toolName string
		version  string
		want     bool
	}{
		{
			name:     "version exists",
			toolName: "terraform",
			version:  "v1.6.0",
			want:     true,
		},
		{
			name:     "version does not exist",
			toolName: "terraform",
			version:  "v1.5.0",
			want:     false,
		},
		{
			name:     "tool does not exist",
			toolName: "tofu",
			version:  "v1.6.0",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.HasVersion(tt.toolName, tt.version)
			if got != tt.want {
				t.Errorf("HasVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveRegistry_AtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	registryPath := filepath.Join(tempDir, "installation.json")

	// Create initial registry
	registry1 := NewRegistry()
	registry1.AddVersion("terraform", "v1.5.0", ToolVersion{
		ToolName: "terraform",
		Version:  "v1.5.0",
		Status:   "complete",
	})

	if err := SaveRegistry(registry1, registryPath); err != nil {
		t.Fatalf("SaveRegistry() first write failed: %v", err)
	}

	// Update registry
	registry2 := NewRegistry()
	registry2.AddVersion("terraform", "v1.6.0", ToolVersion{
		ToolName: "terraform",
		Version:  "v1.6.0",
		Status:   "complete",
	})

	if err := SaveRegistry(registry2, registryPath); err != nil {
		t.Fatalf("SaveRegistry() second write failed: %v", err)
	}

	// Verify no temporary file exists
	tmpPath := registryPath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("SaveRegistry() atomic write left temporary file")
	}

	// Verify final content
	loaded, err := LoadRegistry(registryPath)
	if err != nil {
		t.Fatalf("LoadRegistry() after atomic write failed: %v", err)
	}

	if !loaded.HasVersion("terraform", "v1.6.0") {
		t.Error("Atomic write failed: v1.6.0 not found")
	}

	if loaded.HasVersion("terraform", "v1.5.0") {
		t.Error("Atomic write failed: v1.5.0 should be overwritten")
	}
}
