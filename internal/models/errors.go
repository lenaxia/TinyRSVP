package models

import "fmt"

type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %v", e.Resource, e.ID)
}

type ConflictError struct {
	Resource string
	Field    string
	Value    interface{}
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s conflict on %s: %v", e.Resource, e.Field, e.Value)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

type OptimisticLockError struct {
	Resource        string
	ID              int64
	ExpectedVersion int
	ActualVersion   int
}

func (e *OptimisticLockError) Error() string {
	return fmt.Sprintf("%s %d was modified (expected version %d, got %d)",
		e.Resource, e.ID, e.ExpectedVersion, e.ActualVersion)
}

type PermissionDeniedError struct {
	Action   string
	Resource string
	ID       interface{}
}

func (e *PermissionDeniedError) Error() string {
	if e.ID != nil {
		return fmt.Sprintf("permission denied: cannot %s %s %v", e.Action, e.Resource, e.ID)
	}
	return fmt.Sprintf("permission denied: cannot %s %s", e.Action, e.Resource)
}
