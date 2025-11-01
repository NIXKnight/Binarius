package utils

import (
	"strings"
	"testing"
)

func TestUserError_Error(t *testing.T) {
	tests := []struct {
		name    string
		context string
		reason  string
		action  string
		want    []string // Expected substrings in error message
	}{
		{
			name:    "complete error message",
			context: "terraform v1.6.0 is not installed",
			reason:  "Version not found in registry",
			action:  "Run 'binarius install terraform@v1.6.0' to install it",
			want: []string{
				"Error: terraform v1.6.0 is not installed",
				"Reason: Version not found in registry",
				"Action: Run 'binarius install terraform@v1.6.0' to install it",
			},
		},
		{
			name:    "symlink creation failure",
			context: "Failed to create symlink at ~/.local/bin/terraform",
			reason:  "Permission denied",
			action:  "Ensure ~/.local/bin exists and is writable, or run with sudo",
			want: []string{
				"Error: Failed to create symlink at ~/.local/bin/terraform",
				"Reason: Permission denied",
				"Action: Ensure ~/.local/bin exists and is writable, or run with sudo",
			},
		},
		{
			name:    "empty fields",
			context: "",
			reason:  "",
			action:  "",
			want: []string{
				"Error: ",
				"Reason: ",
				"Action: ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &UserError{
				Context: tt.context,
				Reason:  tt.reason,
				Action:  tt.action,
			}
			got := e.Error()

			for _, substr := range tt.want {
				if !strings.Contains(got, substr) {
					t.Errorf("UserError.Error() missing expected substring\ngot: %q\nwant substring: %q", got, substr)
				}
			}
		})
	}
}

func TestNewUserError(t *testing.T) {
	tests := []struct {
		name    string
		context string
		reason  string
		action  string
	}{
		{
			name:    "creates error with all fields",
			context: "test context",
			reason:  "test reason",
			action:  "test action",
		},
		{
			name:    "creates error with empty fields",
			context: "",
			reason:  "",
			action:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUserError(tt.context, tt.reason, tt.action)

			if got.Context != tt.context {
				t.Errorf("NewUserError().Context = %v, want %v", got.Context, tt.context)
			}
			if got.Reason != tt.reason {
				t.Errorf("NewUserError().Reason = %v, want %v", got.Reason, tt.reason)
			}
			if got.Action != tt.action {
				t.Errorf("NewUserError().Action = %v, want %v", got.Action, tt.action)
			}
		})
	}
}

func TestUserError_ImplementsError(t *testing.T) {
	var _ error = (*UserError)(nil)

	err := NewUserError("test", "test", "test")
	if err.Error() == "" {
		t.Error("UserError.Error() returned empty string")
	}
}

func TestUserError_Formatting(t *testing.T) {
	err := &UserError{
		Context: "operation failed",
		Reason:  "invalid input",
		Action:  "fix the input",
	}

	got := err.Error()
	lines := strings.Split(got, "\n")

	if len(lines) != 3 {
		t.Errorf("UserError.Error() should have 3 lines, got %d", len(lines))
	}

	if !strings.HasPrefix(lines[0], "Error: ") {
		t.Errorf("First line should start with 'Error: ', got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "Reason: ") {
		t.Errorf("Second line should start with 'Reason: ', got %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "Action: ") {
		t.Errorf("Third line should start with 'Action: ', got %q", lines[2])
	}
}
