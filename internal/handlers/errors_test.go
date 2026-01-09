package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name    string
		apiErr  *APIError
		wantMsg string
	}{
		{
			name: "with underlying error",
			apiErr: &APIError{
				StatusCode: http.StatusBadRequest,
				Code:       "VALIDATION_ERROR",
				Message:    "Validation failed",
				Err:        errors.New("field error"),
			},
			wantMsg: "Validation failed: field error",
		},
		{
			name: "without underlying error",
			apiErr: &APIError{
				StatusCode: http.StatusNotFound,
				Code:       "NOT_FOUND",
				Message:    "Resource not found",
			},
			wantMsg: "Resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.apiErr.Error(); got != tt.wantMsg {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	apiErr := &APIError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    "Internal error",
		Err:        underlyingErr,
	}

	if unwrapped := apiErr.Unwrap(); unwrapped != underlyingErr {
		t.Errorf("APIError.Unwrap() = %v, want %v", unwrapped, underlyingErr)
	}
}

func TestNewValidationError(t *testing.T) {
	fields := map[string]string{
		"email": "Email is required",
		"title": "Title must be at least 3 characters",
	}

	err := NewValidationError(fields)

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusBadRequest)
	}
	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("Code = %v, want VALIDATION_ERROR", err.Code)
	}
	if err.Message != "Validation failed" {
		t.Errorf("Message = %v, want 'Validation failed'", err.Message)
	}
	if len(err.Fields) != 2 {
		t.Errorf("Fields length = %v, want 2", len(err.Fields))
	}
}

func TestNewNotFoundError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "event not found",
			message: "Event not found",
		},
		{
			name:    "user not found",
			message: "User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewNotFoundError(tt.message)

			if err.StatusCode != http.StatusNotFound {
				t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusNotFound)
			}
			if err.Code != "NOT_FOUND" {
				t.Errorf("Code = %v, want NOT_FOUND", err.Code)
			}
			if err.Message != tt.message {
				t.Errorf("Message = %v, want %v", err.Message, tt.message)
			}
		})
	}
}

func TestNewPermissionDeniedError(t *testing.T) {
	message := "You do not have permission to access this resource"
	err := NewPermissionDeniedError(message)

	if err.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusForbidden)
	}
	if err.Code != "PERMISSION_DENIED" {
		t.Errorf("Code = %v, want PERMISSION_DENIED", err.Code)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	message := "Authentication required"
	err := NewUnauthorizedError(message)

	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusUnauthorized)
	}
	if err.Code != "UNAUTHORIZED" {
		t.Errorf("Code = %v, want UNAUTHORIZED", err.Code)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
}

func TestNewConflictError(t *testing.T) {
	message := "Resource already exists"
	err := NewConflictError(message)

	if err.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusConflict)
	}
	if err.Code != "CONFLICT" {
		t.Errorf("Code = %v, want CONFLICT", err.Code)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError()

	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusInternalServerError)
	}
	if err.Code != "INTERNAL_ERROR" {
		t.Errorf("Code = %v, want INTERNAL_ERROR", err.Code)
	}
	if err.Message != "An internal error occurred" {
		t.Errorf("Message = %v, want 'An internal error occurred'", err.Message)
	}
}

func TestNewBadRequestError(t *testing.T) {
	message := "Invalid request format"
	err := NewBadRequestError(message)

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusBadRequest)
	}
	if err.Code != "BAD_REQUEST" {
		t.Errorf("Code = %v, want BAD_REQUEST", err.Code)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
}

func TestToAPIError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantCode       string
	}{
		{
			name:           "already APIError",
			err:            NewNotFoundError("test"),
			wantStatusCode: http.StatusNotFound,
			wantCode:       "NOT_FOUND",
		},
		{
			name:           "models.NotFoundError",
			err:            &models.NotFoundError{Resource: "Event", ID: 123},
			wantStatusCode: http.StatusNotFound,
			wantCode:       "NOT_FOUND",
		},
		{
			name:           "models.ValidationError",
			err:            &models.ValidationError{Field: "email", Message: "required"},
			wantStatusCode: http.StatusBadRequest,
			wantCode:       "VALIDATION_ERROR",
		},
		{
			name:           "models.ConflictError",
			err:            &models.ConflictError{Resource: "User", Field: "email", Value: "test@example.com"},
			wantStatusCode: http.StatusConflict,
			wantCode:       "CONFLICT",
		},
		{
			name:           "models.PermissionDeniedError",
			err:            &models.PermissionDeniedError{Action: "delete", Resource: "Event"},
			wantStatusCode: http.StatusForbidden,
			wantCode:       "PERMISSION_DENIED",
		},
		{
			name:           "models.UnauthorizedError",
			err:            &models.UnauthorizedError{Message: "auth required"},
			wantStatusCode: http.StatusUnauthorized,
			wantCode:       "UNAUTHORIZED",
		},
		{
			name:           "models.ForbiddenError",
			err:            &models.ForbiddenError{Message: "forbidden"},
			wantStatusCode: http.StatusForbidden,
			wantCode:       "PERMISSION_DENIED",
		},
		{
			name:           "generic error",
			err:            errors.New("something went wrong"),
			wantStatusCode: http.StatusInternalServerError,
			wantCode:       "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := toAPIError(tt.err)

			if apiErr.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, tt.wantStatusCode)
			}
			if apiErr.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", apiErr.Code, tt.wantCode)
			}
		})
	}
}

func TestWantsJSON(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   bool
	}{
		{
			name:   "application/json",
			accept: "application/json",
			want:   true,
		},
		{
			name:   "application/json with charset",
			accept: "application/json; charset=utf-8",
			want:   true,
		},
		{
			name:   "multiple with json first",
			accept: "application/json, text/html",
			want:   true,
		},
		{
			name:   "text/html",
			accept: "text/html",
			want:   false,
		},
		{
			name:   "empty",
			accept: "",
			want:   false,
		},
		{
			name:   "wildcard",
			accept: "*/*",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", tt.accept)

			if got := wantsJSON(req); got != tt.want {
				t.Errorf("wantsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteJSONError(t *testing.T) {
	tests := []struct {
		name   string
		apiErr *APIError
	}{
		{
			name: "simple error",
			apiErr: &APIError{
				StatusCode: http.StatusNotFound,
				Code:       "NOT_FOUND",
				Message:    "Resource not found",
			},
		},
		{
			name: "validation error with fields",
			apiErr: &APIError{
				StatusCode: http.StatusBadRequest,
				Code:       "VALIDATION_ERROR",
				Message:    "Validation failed",
				Fields: map[string]string{
					"email": "Email is required",
					"title": "Title is too short",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSONError(w, tt.apiErr)

			if w.Code != tt.apiErr.StatusCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.apiErr.StatusCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if resp.Error != http.StatusText(tt.apiErr.StatusCode) {
				t.Errorf("Error = %v, want %v", resp.Error, http.StatusText(tt.apiErr.StatusCode))
			}
			if resp.Message != tt.apiErr.Message {
				t.Errorf("Message = %v, want %v", resp.Message, tt.apiErr.Message)
			}
			if resp.Code != tt.apiErr.Code {
				t.Errorf("Code = %v, want %v", resp.Code, tt.apiErr.Code)
			}
			if tt.apiErr.Fields != nil && len(resp.Fields) != len(tt.apiErr.Fields) {
				t.Errorf("Fields length = %v, want %v", len(resp.Fields), len(tt.apiErr.Fields))
			}
		})
	}
}

func TestHandleError_JSON(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "not found error",
			err:        NewNotFoundError("Event not found"),
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name: "validation error",
			err: NewValidationError(map[string]string{
				"email": "Email is required",
			}),
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "internal error",
			err:        errors.New("database connection failed"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", "application/json")
			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-request-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			HandleError(w, req, tt.err)

			if w.Code != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", w.Code, tt.wantStatus)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", resp.Code, tt.wantCode)
			}
		})
	}
}

func TestHandleError_HTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html")
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-request-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	err := NewNotFoundError("Event not found")
	HandleError(w, req, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusNotFound)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Type = %v, want text/html", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "404") {
		t.Error("HTML response should contain status code")
	}
}
