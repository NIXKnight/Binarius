package tools

import (
	"sync"
	"testing"
)

// mockTool is a test implementation of the Tool interface
type mockTool struct {
	name           string
	downloadURL    string
	checksumURL    string
	versions       []string
	binaryName     string
	archiveFormat  string
	supportedArchs []string
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
	return m.versions, nil
}

func (m *mockTool) GetBinaryName() string {
	return m.binaryName
}

func (m *mockTool) GetArchiveFormat() string {
	return m.archiveFormat
}

func (m *mockTool) SupportedArchs() []string {
	return m.supportedArchs
}

func TestRegister(t *testing.T) {
	// Clear registry before each test
	Clear()

	tests := []struct {
		name      string
		toolName  string
		tool      Tool
		wantErr   bool
		setupFunc func() // Setup function to register tools before the test
	}{
		{
			name:     "register new tool",
			toolName: "terraform",
			tool: &mockTool{
				name:       "terraform",
				binaryName: "terraform",
			},
			wantErr: false,
		},
		{
			name:     "register duplicate tool",
			toolName: "terraform",
			tool: &mockTool{
				name:       "terraform",
				binaryName: "terraform",
			},
			setupFunc: func() {
				Register("terraform", &mockTool{name: "terraform"})
			},
			wantErr: true,
		},
		{
			name:     "tool name mismatch",
			toolName: "terraform",
			tool: &mockTool{
				name:       "tofu", // Different from registration name
				binaryName: "tofu",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Clear() // Clear before each subtest

			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			err := Register(tt.toolName, tt.tool)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	Clear()

	// Setup: Register a test tool
	testTool := &mockTool{
		name:       "terraform",
		binaryName: "terraform",
	}
	Register("terraform", testTool)

	tests := []struct {
		name     string
		toolName string
		wantErr  bool
	}{
		{
			name:     "get existing tool",
			toolName: "terraform",
			wantErr:  false,
		},
		{
			name:     "get non-existent tool",
			toolName: "nonexistent",
			wantErr:  true,
		},
		{
			name:     "case sensitive lookup",
			toolName: "Terraform", // Wrong case
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.toolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Error("Get() returned nil tool when error not expected")
				}
				if got.GetName() != tt.toolName {
					t.Errorf("Get() returned tool with name %q, want %q", got.GetName(), tt.toolName)
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	Clear()

	tests := []struct {
		name     string
		setup    func()
		wantLen  int
		wantTool string // Tool name that should be in the list
	}{
		{
			name: "empty registry",
			setup: func() {
				Clear()
			},
			wantLen: 0,
		},
		{
			name: "single tool registered",
			setup: func() {
				Clear()
				Register("terraform", &mockTool{name: "terraform"})
			},
			wantLen:  1,
			wantTool: "terraform",
		},
		{
			name: "multiple tools registered",
			setup: func() {
				Clear()
				Register("terraform", &mockTool{name: "terraform"})
				Register("tofu", &mockTool{name: "tofu"})
				Register("terragrunt", &mockTool{name: "terragrunt"})
			},
			wantLen:  3,
			wantTool: "terraform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got := List()
			if len(got) != tt.wantLen {
				t.Errorf("List() returned %d tools, want %d", len(got), tt.wantLen)
			}

			if tt.wantTool != "" {
				found := false
				for _, name := range got {
					if name == tt.wantTool {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("List() missing expected tool %q in %v", tt.wantTool, got)
				}
			}
		})
	}
}

func TestClear(t *testing.T) {
	// Clear registry first to ensure clean state
	Clear()

	// Setup: Register some tools
	Register("terraform", &mockTool{name: "terraform"})
	Register("tofu", &mockTool{name: "tofu"})

	// Verify tools are registered
	if len(List()) != 2 {
		t.Fatalf("Setup failed: expected 2 tools registered, got %d", len(List()))
	}

	// Clear the registry
	Clear()

	// Verify registry is empty
	if len(List()) != 0 {
		t.Errorf("Clear() failed: registry still has %d tools", len(List()))
	}

	// Verify we can register again after clear
	err := Register("terraform", &mockTool{name: "terraform"})
	if err != nil {
		t.Errorf("Register() after Clear() failed: %v", err)
	}
}

// TestConcurrentAccess verifies that the registry is safe for concurrent use
func TestConcurrentAccess(t *testing.T) {
	Clear()

	// Register initial tools
	Register("terraform", &mockTool{name: "terraform"})
	Register("tofu", &mockTool{name: "tofu"})

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = Get("terraform")
			_ = List()
		}()
	}

	// Concurrent registrations (should mostly fail due to duplicates)
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			toolName := "tool" + string(rune(index))
			_ = Register(toolName, &mockTool{name: toolName})
		}(i)
	}

	wg.Wait()

	// Verify registry is still functional
	tools := List()
	if len(tools) < 2 {
		t.Errorf("Concurrent access corrupted registry: got %d tools, want at least 2", len(tools))
	}
}

// TestRegisterConcurrent verifies race-free registration
func TestRegisterConcurrent(t *testing.T) {
	Clear()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Try to register the same tool concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Register("concurrent-tool", &mockTool{name: "concurrent-tool"})
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Only one should succeed, the rest should fail with "already registered" error
	errorCount := 0
	for range errors {
		errorCount++
	}

	if errorCount != 9 {
		t.Errorf("Expected 9 registration errors (9 duplicates), got %d", errorCount)
	}

	// Verify the tool was registered
	tool, err := Get("concurrent-tool")
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}
	if tool == nil {
		t.Error("Get() returned nil tool")
	}
}

func TestToolInterface(t *testing.T) {
	// Verify mockTool implements Tool interface at compile time
	var _ Tool = (*mockTool)(nil)

	tool := &mockTool{
		name:           "test-tool",
		downloadURL:    "https://example.com/download",
		checksumURL:    "https://example.com/checksum",
		versions:       []string{"v1.0.0", "v0.9.0"},
		binaryName:     "test-tool",
		archiveFormat:  "zip",
		supportedArchs: []string{"amd64", "arm64"},
	}

	// Test all interface methods
	if tool.GetName() != "test-tool" {
		t.Errorf("GetName() = %q, want %q", tool.GetName(), "test-tool")
	}

	if tool.GetDownloadURL("v1.0.0", "linux", "amd64") != "https://example.com/download" {
		t.Error("GetDownloadURL() returned unexpected value")
	}

	if tool.GetChecksumURL("v1.0.0", "linux", "amd64") != "https://example.com/checksum" {
		t.Error("GetChecksumURL() returned unexpected value")
	}

	versions, err := tool.ListVersions()
	if err != nil {
		t.Errorf("ListVersions() error = %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("ListVersions() returned %d versions, want 2", len(versions))
	}

	if tool.GetBinaryName() != "test-tool" {
		t.Errorf("GetBinaryName() = %q, want %q", tool.GetBinaryName(), "test-tool")
	}

	if tool.GetArchiveFormat() != "zip" {
		t.Errorf("GetArchiveFormat() = %q, want %q", tool.GetArchiveFormat(), "zip")
	}

	archs := tool.SupportedArchs()
	if len(archs) != 2 {
		t.Errorf("SupportedArchs() returned %d architectures, want 2", len(archs))
	}
}
