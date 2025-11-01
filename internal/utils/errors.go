package utils

import "fmt"

// UserError represents a user-facing error with context, reason, and actionable guidance.
// It implements the error interface and formats messages for clear communication.
type UserError struct {
	Context string // What operation failed
	Reason  string // Why it failed (in user-friendly terms)
	Action  string // What the user should do to fix it
}

// Error implements the error interface, formatting the error message
// with context, reason, and action in a user-friendly format.
func (e *UserError) Error() string {
	return fmt.Sprintf("Error: %s\nReason: %s\nAction: %s", e.Context, e.Reason, e.Action)
}

// NewUserError creates a new UserError with the provided context, reason, and action.
func NewUserError(context, reason, action string) *UserError {
	return &UserError{
		Context: context,
		Reason:  reason,
		Action:  action,
	}
}
