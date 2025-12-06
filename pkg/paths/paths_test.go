package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBinariusHome(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		want    func(string) string // function that takes home dir and returns expected path
		wantErr bool
	}{
		{
			name:   "default home directory",
			envVar: "",
			want: func(home string) string {
				return filepath.Join(home, ".binarius")
			},
			wantErr: false,
		},
		{
			name:   "custom environment variable with absolute path",
			envVar: "/custom/binarius",
			want: func(home string) string {
				return "/custom/binarius"
			},
			wantErr: false,
		},
		{
			name:   "custom environment variable with tilde",
			envVar: "~/custom/binarius",
			want: func(home string) string {
				return filepath.Join(home, "custom", "binarius")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set or clear environment variable
			if tt.envVar != "" {
				t.Setenv("BINARIUS_HOME", tt.envVar)
			}

			got, err := BinariusHome()
			if (err != nil) != tt.wantErr {
				t.Errorf("BinariusHome() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				homeDir, _ := os.UserHomeDir()
				want := tt.want(homeDir)
				if got != want {
					t.Errorf("BinariusHome() = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestBinDir(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		want    func(string) string
		wantErr bool
	}{
		{
			name:   "default bin directory",
			envVar: "",
			want: func(home string) string {
				return filepath.Join(home, ".local", "bin")
			},
			wantErr: false,
		},
		{
			name:   "custom environment variable with absolute path",
			envVar: "/usr/local/bin",
			want: func(home string) string {
				return "/usr/local/bin"
			},
			wantErr: false,
		},
		{
			name:   "custom environment variable with tilde",
			envVar: "~/bin",
			want: func(home string) string {
				return filepath.Join(home, "bin")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv("BINARIUS_BIN_DIR", tt.envVar)
			}

			got, err := BinDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("BinDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				homeDir, _ := os.UserHomeDir()
				want := tt.want(homeDir)
				if got != want {
					t.Errorf("BinDir() = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestCacheDir(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		want    func(string) string
		wantErr bool
	}{
		{
			name:   "default cache directory",
			envVar: "",
			want: func(home string) string {
				return filepath.Join(home, ".binarius", "cache")
			},
			wantErr: false,
		},
		{
			name:   "custom environment variable with absolute path",
			envVar: "/tmp/binarius-cache",
			want: func(home string) string {
				return "/tmp/binarius-cache"
			},
			wantErr: false,
		},
		{
			name:   "custom environment variable with tilde",
			envVar: "~/.cache/binarius",
			want: func(home string) string {
				return filepath.Join(home, ".cache", "binarius")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv("BINARIUS_CACHE_DIR", tt.envVar)
			}

			got, err := CacheDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				homeDir, _ := os.UserHomeDir()
				want := tt.want(homeDir)
				if got != want {
					t.Errorf("CacheDir() = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestToolsDir(t *testing.T) {
	tests := []struct {
		name    string
		want    func(string) string
		wantErr bool
	}{
		{
			name: "default tools directory",
			want: func(home string) string {
				return filepath.Join(home, ".binarius", "tools")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToolsDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToolsDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				homeDir, _ := os.UserHomeDir()
				want := tt.want(homeDir)
				if got != want {
					t.Errorf("ToolsDir() = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "absolute path unchanged",
			path:    "/absolute/path",
			want:    "/absolute/path",
			wantErr: false,
		},
		{
			name:    "relative path unchanged",
			path:    "relative/path",
			want:    "relative/path",
			wantErr: false,
		},
		{
			name:    "tilde only",
			path:    "~",
			want:    homeDir,
			wantErr: false,
		},
		{
			name:    "tilde with slash",
			path:    "~/path/to/file",
			want:    filepath.Join(homeDir, "path", "to", "file"),
			wantErr: false,
		},
		{
			name:    "tilde without slash (unsupported ~user)",
			path:    "~user/path",
			want:    "~user/path",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("expandPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
