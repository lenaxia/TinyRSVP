package models

import (
	"errors"
	"testing"
)

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      *NotFoundError
		wantMsg  string
		wantType bool
	}{
		{
			name: "user not found with int ID",
			err: &NotFoundError{
				Resource: "User",
				ID:       123,
			},
			wantMsg:  "User not found: 123",
			wantType: true,
		},
		{
			name: "session not found with string ID",
			err: &NotFoundError{
				Resource: "Session",
				ID:       "session-abc-123",
			},
			wantMsg:  "Session not found: session-abc-123",
			wantType: true,
		},
		{
			name: "event not found with int64 ID",
			err: &NotFoundError{
				Resource: "Event",
				ID:       int64(999),
			},
			wantMsg:  "Event not found: 999",
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("NotFoundError.Error() = %v, want %v", got, tt.wantMsg)
			}

			var target *NotFoundError
			if errors.As(tt.err, &target) != tt.wantType {
				t.Errorf("errors.As() = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestConflictError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConflictError
		wantMsg  string
		wantType bool
	}{
		{
			name: "duplicate email",
			err: &ConflictError{
				Resource: "User",
				Field:    "email",
				Value:    "test@example.com",
			},
			wantMsg:  "User conflict on email: test@example.com",
			wantType: true,
		},
		{
			name: "duplicate token hash",
			err: &ConflictError{
				Resource: "Invite",
				Field:    "token_hash",
				Value:    "abc123hash",
			},
			wantMsg:  "Invite conflict on token_hash: abc123hash",
			wantType: true,
		},
		{
			name: "duplicate OIDC subject",
			err: &ConflictError{
				Resource: "User",
				Field:    "oidc_subject",
				Value:    "google-oauth2|123456",
			},
			wantMsg:  "User conflict on oidc_subject: google-oauth2|123456",
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ConflictError.Error() = %v, want %v", got, tt.wantMsg)
			}

			var target *ConflictError
			if errors.As(tt.err, &target) != tt.wantType {
				t.Errorf("errors.As() = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		wantMsg  string
		wantType bool
	}{
		{
			name: "empty email",
			err: &ValidationError{
				Field:   "email",
				Message: "email is required",
			},
			wantMsg:  "validation error on email: email is required",
			wantType: true,
		},
		{
			name: "invalid email format",
			err: &ValidationError{
				Field:   "email",
				Message: "invalid email format",
			},
			wantMsg:  "validation error on email: invalid email format",
			wantType: true,
		},
		{
			name: "title too short",
			err: &ValidationError{
				Field:   "title",
				Message: "title must be at least 3 characters",
			},
			wantMsg:  "validation error on title: title must be at least 3 characters",
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.wantMsg)
			}

			var target *ValidationError
			if errors.As(tt.err, &target) != tt.wantType {
				t.Errorf("errors.As() = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestOptimisticLockError(t *testing.T) {
	tests := []struct {
		name     string
		err      *OptimisticLockError
		wantMsg  string
		wantType bool
	}{
		{
			name: "event version mismatch",
			err: &OptimisticLockError{
				Resource:        "Event",
				ID:              123,
				ExpectedVersion: 1,
				ActualVersion:   2,
			},
			wantMsg:  "Event 123 was modified (expected version 1, got 2)",
			wantType: true,
		},
		{
			name: "template version mismatch",
			err: &OptimisticLockError{
				Resource:        "Template",
				ID:              456,
				ExpectedVersion: 5,
				ActualVersion:   7,
			},
			wantMsg:  "Template 456 was modified (expected version 5, got 7)",
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("OptimisticLockError.Error() = %v, want %v", got, tt.wantMsg)
			}

			var target *OptimisticLockError
			if errors.As(tt.err, &target) != tt.wantType {
				t.Errorf("errors.As() = %v, want %v", !tt.wantType, tt.wantType)
			}
		})
	}
}

func TestErrorTypeChecking(t *testing.T) {
	t.Run("can distinguish between error types", func(t *testing.T) {
		notFoundErr := &NotFoundError{Resource: "User", ID: 1}
		conflictErr := &ConflictError{Resource: "User", Field: "email", Value: "test@example.com"}
		validationErr := &ValidationError{Field: "email", Message: "required"}
		lockErr := &OptimisticLockError{Resource: "Event", ID: 1, ExpectedVersion: 1, ActualVersion: 2}

		var nf *NotFoundError
		if !errors.As(notFoundErr, &nf) {
			t.Error("Expected NotFoundError to match NotFoundError type")
		}
		if errors.As(conflictErr, &nf) {
			t.Error("Expected ConflictError not to match NotFoundError type")
		}

		var cf *ConflictError
		if !errors.As(conflictErr, &cf) {
			t.Error("Expected ConflictError to match ConflictError type")
		}
		if errors.As(validationErr, &cf) {
			t.Error("Expected ValidationError not to match ConflictError type")
		}

		var vf *ValidationError
		if !errors.As(validationErr, &vf) {
			t.Error("Expected ValidationError to match ValidationError type")
		}
		if errors.As(lockErr, &vf) {
			t.Error("Expected OptimisticLockError not to match ValidationError type")
		}

		var lf *OptimisticLockError
		if !errors.As(lockErr, &lf) {
			t.Error("Expected OptimisticLockError to match OptimisticLockError type")
		}
		if errors.As(notFoundErr, &lf) {
			t.Error("Expected NotFoundError not to match OptimisticLockError type")
		}
	})
}
