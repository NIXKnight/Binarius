package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Verify defaults map is initialized
	if config.Defaults == nil {
		t.Error("DefaultConfig() Defaults map is nil")
	}

	// Verify paths are set
	if config.Paths.BinariusHome == "" {
		t.Error("DefaultConfig() BinariusHome is empty")
	}
	if config.Paths.BinDir == "" {
		t.Error("DefaultConfig() BinDir is empty")
	}
	if config.Paths.CacheDir == "" {
		t.Error("DefaultConfig() CacheDir is empty")
	}

	// Verify paths contain expected components
	homeDir, _ := os.UserHomeDir()
	expectedHome := filepath.Join(homeDir, ".binarius")
	if config.Paths.BinariusHome != expectedHome {
		t.Errorf("DefaultConfig() BinariusHome = %q, want %q", config.Paths.BinariusHome, expectedHome)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		setupFile func(string) // Function to create test file
		wantErr   bool
		validate  func(*testing.T, *Config) // Validation function
	}{
		{
			name: "valid config file",
			setupFile: func(path string) {
				content := `defaults:
  terraform: v1.6.0
  tofu: v1.6.0
paths:
  binarius_home: ~/.binarius
  bin_dir: ~/.local/bin
  cache_dir: ~/.binarius/cache
`
				os.WriteFile(path, []byte(content), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				if c.Defaults["terraform"] != "v1.6.0" {
					t.Errorf("Load() terraform default = %q, want v1.6.0", c.Defaults["terraform"])
				}
				if c.Defaults["tofu"] != "v1.6.0" {
					t.Errorf("Load() tofu default = %q, want v1.6.0", c.Defaults["tofu"])
				}
				if c.Paths.BinariusHome != "~/.binarius" {
					t.Errorf("Load() BinariusHome = %q, want ~/.binarius", c.Paths.BinariusHome)
				}
			},
		},
		{
			name: "empty config file",
			setupFile: func(path string) {
				os.WriteFile(path, []byte(""), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				if c.Defaults == nil {
					t.Error("Load() with empty file should initialize Defaults map")
				}
			},
		},
		{
			name: "config with empty defaults",
			setupFile: func(path string) {
				content := `defaults: {}
paths:
  binarius_home: ~/.binarius
  bin_dir: ~/.local/bin
  cache_dir: ~/.binarius/cache
`
				os.WriteFile(path, []byte(content), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				if c.Defaults == nil {
					t.Error("Load() should initialize Defaults map")
				}
				if len(c.Defaults) != 0 {
					t.Errorf("Load() Defaults should be empty, got %d items", len(c.Defaults))
				}
			},
		},
		{
			name: "invalid yaml",
			setupFile: func(path string) {
				os.WriteFile(path, []byte("invalid: yaml: content:"), 0644)
			},
			wantErr: true,
		},
		{
			name: "non-existent file",
			setupFile: func(path string) {
				// Don't create the file
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")

			tt.setupFile(configPath)

			config, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "save valid config",
			config: &Config{
				Defaults: map[string]string{
					"terraform": "v1.6.0",
					"tofu":      "v1.6.0",
				},
				Paths: PathConfig{
					BinariusHome: "~/.binarius",
					BinDir:       "~/.local/bin",
					CacheDir:     "~/.binarius/cache",
				},
			},
			wantErr: false,
		},
		{
			name: "save config with empty defaults",
			config: &Config{
				Defaults: map[string]string{},
				Paths: PathConfig{
					BinariusHome: "~/.binarius",
					BinDir:       "~/.local/bin",
					CacheDir:     "~/.binarius/cache",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")

			err := Save(tt.config, configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created
				if _, err := os.Stat(configPath); err != nil {
					t.Errorf("Save() file not created: %v", err)
				}

				// Verify no temporary file left behind
				tmpPath := configPath + ".tmp"
				if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
					t.Error("Save() temporary file not cleaned up")
				}

				// Verify content can be loaded back
				loaded, err := Load(configPath)
				if err != nil {
					t.Errorf("Load() after Save() failed: %v", err)
				}

				// Verify defaults match
				if len(loaded.Defaults) != len(tt.config.Defaults) {
					t.Errorf("Save/Load defaults count mismatch: got %d, want %d",
						len(loaded.Defaults), len(tt.config.Defaults))
				}
				for tool, version := range tt.config.Defaults {
					if loaded.Defaults[tool] != version {
						t.Errorf("Save/Load default mismatch for %s: got %q, want %q",
							tool, loaded.Defaults[tool], version)
					}
				}
			}
		})
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create initial config
	config1 := &Config{
		Defaults: map[string]string{"terraform": "v1.5.0"},
		Paths: PathConfig{
			BinariusHome: "~/.binarius",
			BinDir:       "~/.local/bin",
			CacheDir:     "~/.binarius/cache",
		},
	}

	if err := Save(config1, configPath); err != nil {
		t.Fatalf("Save() first write failed: %v", err)
	}

	// Update config
	config2 := &Config{
		Defaults: map[string]string{"terraform": "v1.6.0"},
		Paths: PathConfig{
			BinariusHome: "~/.binarius",
			BinDir:       "~/.local/bin",
			CacheDir:     "~/.binarius/cache",
		},
	}

	if err := Save(config2, configPath); err != nil {
		t.Fatalf("Save() second write failed: %v", err)
	}

	// Verify no temporary file exists
	tmpPath := configPath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("Save() atomic write left temporary file")
	}

	// Verify final content
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() after atomic write failed: %v", err)
	}

	if loaded.Defaults["terraform"] != "v1.6.0" {
		t.Errorf("Atomic write failed: got %q, want v1.6.0", loaded.Defaults["terraform"])
	}
}

func TestSetDefault(t *testing.T) {
	tests := []struct {
		name      string
		initial   *Config
		tool      string
		version   string
		wantValue string // Expected value after SetDefault
		wantInMap bool   // Whether the key should exist in the map
	}{
		{
			name: "set new default",
			initial: &Config{
				Defaults: map[string]string{},
			},
			tool:      "terraform",
			version:   "v1.6.0",
			wantValue: "v1.6.0",
			wantInMap: true,
		},
		{
			name: "update existing default",
			initial: &Config{
				Defaults: map[string]string{"terraform": "v1.5.0"},
			},
			tool:      "terraform",
			version:   "v1.6.0",
			wantValue: "v1.6.0",
			wantInMap: true,
		},
		{
			name: "remove default with empty version",
			initial: &Config{
				Defaults: map[string]string{"terraform": "v1.6.0"},
			},
			tool:      "terraform",
			version:   "",
			wantValue: "",
			wantInMap: false,
		},
		{
			name:      "set default with nil map",
			initial:   &Config{Defaults: nil},
			tool:      "terraform",
			version:   "v1.6.0",
			wantValue: "v1.6.0",
			wantInMap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.initial
			config.SetDefault(tt.tool, tt.version)

			value, exists := config.Defaults[tt.tool]
			if exists != tt.wantInMap {
				t.Errorf("SetDefault() key existence = %v, want %v", exists, tt.wantInMap)
			}

			if tt.wantInMap && value != tt.wantValue {
				t.Errorf("SetDefault() value = %q, want %q", value, tt.wantValue)
			}
		})
	}
}

func TestGetDefault(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		tool   string
		want   string
	}{
		{
			name: "get existing default",
			config: &Config{
				Defaults: map[string]string{"terraform": "v1.6.0"},
			},
			tool: "terraform",
			want: "v1.6.0",
		},
		{
			name: "get non-existent default",
			config: &Config{
				Defaults: map[string]string{"terraform": "v1.6.0"},
			},
			tool: "tofu",
			want: "",
		},
		{
			name: "get from empty defaults",
			config: &Config{
				Defaults: map[string]string{},
			},
			tool: "terraform",
			want: "",
		},
		{
			name: "get from nil defaults",
			config: &Config{
				Defaults: nil,
			},
			tool: "terraform",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetDefault(tt.tool)
			if got != tt.want {
				t.Errorf("GetDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSave_CreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", "nested", "config.yaml")

	config := &Config{
		Defaults: map[string]string{},
		Paths: PathConfig{
			BinariusHome: "~/.binarius",
			BinDir:       "~/.local/bin",
			CacheDir:     "~/.binarius/cache",
		},
	}

	// Save should create parent directories
	if err := Save(config, configPath); err != nil {
		t.Errorf("Save() with nested path failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Save() did not create nested directories: %v", err)
	}
}
