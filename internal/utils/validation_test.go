package utils

import "testing"

func TestValidateToolName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid tool names
		{name: "simple name", input: "terraform", wantErr: false},
		{name: "name with hyphen", input: "my-tool", wantErr: false},
		{name: "name with numbers", input: "tool123", wantErr: false},
		{name: "name with multiple hyphens", input: "my-cool-tool", wantErr: false},
		{name: "single character", input: "t", wantErr: false},
		{name: "all numbers", input: "123", wantErr: false},
		{name: "mixed alphanumeric", input: "tool1-beta2", wantErr: false},

		// Invalid tool names
		{name: "empty string", input: "", wantErr: true},
		{name: "uppercase letters", input: "Terraform", wantErr: true},
		{name: "contains space", input: "my tool", wantErr: true},
		{name: "contains underscore", input: "my_tool", wantErr: true},
		{name: "contains dot", input: "my.tool", wantErr: true},
		{name: "starts with hyphen", input: "-tool", wantErr: true},
		{name: "ends with hyphen", input: "tool-", wantErr: true},
		{name: "consecutive hyphens", input: "my--tool", wantErr: true},
		{name: "special characters", input: "tool@123", wantErr: true},
		{name: "slash", input: "my/tool", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid versions
		{name: "version with v prefix", input: "v1.6.0", wantErr: false},
		{name: "version without v prefix", input: "1.6.0", wantErr: false},
		{name: "version with beta suffix", input: "v1.6.0-beta1", wantErr: false},
		{name: "version with rc suffix", input: "v1.6.0-rc.1", wantErr: false},
		{name: "version with alpha suffix", input: "1.6.0-alpha", wantErr: false},
		{name: "version with complex suffix", input: "v1.6.0-beta.1.2", wantErr: false},
		{name: "large version numbers", input: "v100.200.300", wantErr: false},
		{name: "version with hyphenated suffix", input: "v1.6.0-pre-release", wantErr: false},

		// Invalid versions
		{name: "empty string", input: "", wantErr: true},
		{name: "missing patch version", input: "v1.6", wantErr: true},
		{name: "missing minor version", input: "v1", wantErr: true},
		{name: "only v prefix", input: "v", wantErr: true},
		{name: "non-numeric version", input: "vx.y.z", wantErr: true},
		{name: "four parts", input: "v1.6.0.0", wantErr: true},
		{name: "with spaces", input: "v1.6.0 beta", wantErr: true},
		{name: "invalid suffix format", input: "v1.6.0-", wantErr: true},
		{name: "double v prefix", input: "vv1.6.0", wantErr: true},
		{name: "version without dots", input: "v160", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// Valid normalizations
		{name: "add v prefix", input: "1.6.0", want: "v1.6.0", wantErr: false},
		{name: "keep existing v prefix", input: "v1.6.0", want: "v1.6.0", wantErr: false},
		{name: "add v to version with suffix", input: "1.6.0-beta1", want: "v1.6.0-beta1", wantErr: false},
		{name: "keep v on version with suffix", input: "v1.6.0-rc.1", want: "v1.6.0-rc.1", wantErr: false},

		// Invalid versions
		{name: "empty string", input: "", want: "", wantErr: true},
		{name: "invalid version format", input: "1.6", want: "", wantErr: true},
		{name: "non-numeric version", input: "x.y.z", want: "", wantErr: true},
		{name: "invalid characters", input: "v1.6.0@beta", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("NormalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateToolName_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "very long name", input: "this-is-a-very-long-tool-name-with-many-hyphens", wantErr: false},
		{name: "single hyphen between chars", input: "a-b", wantErr: false},
		{name: "triple hyphen", input: "my---tool", wantErr: true},
		{name: "leading and trailing hyphens", input: "-tool-", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateVersion_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "zero version", input: "v0.0.0", wantErr: false},
		{name: "version with dots in suffix", input: "v1.6.0-beta.1.2.3", wantErr: false},
		{name: "version with mixed case suffix", input: "v1.6.0-BETA", wantErr: false},
		{name: "version with numbers in suffix", input: "v1.6.0-123", wantErr: false},
		{name: "very large numbers", input: "v999.999.999", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
