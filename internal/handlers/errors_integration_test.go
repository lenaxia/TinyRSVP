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

func TestErrorHandling_Integration_JSON(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantCode       string
		wantMessage    string
	}{
		{
			name:           "not found error",
			err:            &models.NotFoundError{Resource: "Event", ID: 123},
			wantStatusCode: http.StatusNotFound,
			wantCode:       "NOT_FOUND",
			wantMessage:    "Event not found: 123",
		},
		{
			name:           "validation error",
			err:            &models.ValidationError{Field: "email", Message: "Email is required"},
			wantStatusCode: http.StatusBadRequest,
			wantCode:       "VALIDATION_ERROR",
			wantMessage:    "validation error on email: Email is required",
		},
		{
			name:           "conflict error",
			err:            &models.ConflictError{Resource: "User", Field: "email", Value: "test@example.com"},
			wantStatusCode: http.StatusConflict,
			wantCode:       "CONFLICT",
			wantMessage:    "User conflict on email: test@example.com",
		},
		{
			name:           "permission denied error",
			err:            &models.PermissionDeniedError{Action: "delete", Resource: "Event", ID: 456},
			wantStatusCode: http.StatusForbidden,
			wantCode:       "PERMISSION_DENIED",
			wantMessage:    "permission denied: cannot delete Event 456",
		},
		{
			name:           "unauthorized error",
			err:            &models.UnauthorizedError{Message: "Authentication required"},
			wantStatusCode: http.StatusUnauthorized,
			wantCode:       "UNAUTHORIZED",
			wantMessage:    "Authentication required",
		},
		{
			name:           "forbidden error",
			err:            &models.ForbiddenError{Message: "Access forbidden"},
			wantStatusCode: http.StatusForbidden,
			wantCode:       "PERMISSION_DENIED",
			wantMessage:    "Access forbidden",
		},
		{
			name:           "generic error",
			err:            errors.New("unexpected database error"),
			wantStatusCode: http.StatusInternalServerError,
			wantCode:       "INTERNAL_ERROR",
			wantMessage:    "An internal error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Accept", "application/json")
			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			HandleError(w, req, tt.err)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.wantStatusCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", resp.Code, tt.wantCode)
			}

			if resp.Message != tt.wantMessage {
				t.Errorf("Message = %v, want %v", resp.Message, tt.wantMessage)
			}

			if resp.Error != http.StatusText(tt.wantStatusCode) {
				t.Errorf("Error = %v, want %v", resp.Error, http.StatusText(tt.wantStatusCode))
			}
		})
	}
}

func TestErrorHandling_Integration_HTML(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatusCode int
		wantInBody     []string
	}{
		{
			name:           "not found error",
			err:            NewNotFoundError("Page not found"),
			wantStatusCode: http.StatusNotFound,
			wantInBody:     []string{"404", "Not Found", "Page not found", "test-req-id"},
		},
		{
			name:           "internal error",
			err:            NewInternalError(),
			wantStatusCode: http.StatusInternalServerError,
			wantInBody:     []string{"500", "Internal Server Error", "An internal error occurred"},
		},
		{
			name:           "forbidden error",
			err:            NewPermissionDeniedError("Access denied"),
			wantStatusCode: http.StatusForbidden,
			wantInBody:     []string{"403", "Forbidden", "Access denied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Accept", "text/html")
			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			HandleError(w, req, tt.err)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Status code = %v, want %v", w.Code, tt.wantStatusCode)
			}

			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "text/html") {
				t.Errorf("Content-Type = %v, want text/html", contentType)
			}

			body := w.Body.String()
			for _, want := range tt.wantInBody {
				if !strings.Contains(body, want) {
					t.Errorf("Body should contain %q, but doesn't. Body: %s", want, body)
				}
			}
		})
	}
}

func TestErrorHandling_Integration_ContentNegotiation(t *testing.T) {
	tests := []struct {
		name        string
		acceptHeader string
		wantJSON    bool
	}{
		{
			name:         "prefers JSON",
			acceptHeader: "application/json",
			wantJSON:     true,
		},
		{
			name:         "prefers HTML",
			acceptHeader: "text/html",
			wantJSON:     false,
		},
		{
			name:         "JSON with quality",
			acceptHeader: "application/json;q=0.9, text/html;q=0.8",
			wantJSON:     true,
		},
		{
			name:         "no accept header",
			acceptHeader: "",
			wantJSON:     false,
		},
		{
			name:         "wildcard",
			acceptHeader: "*/*",
			wantJSON:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			err := NewNotFoundError("Resource not found")
			HandleError(w, req, err)

			contentType := w.Header().Get("Content-Type")
			isJSON := strings.Contains(contentType, "application/json")

			if isJSON != tt.wantJSON {
				t.Errorf("Got JSON response = %v, want %v (Content-Type: %s)", isJSON, tt.wantJSON, contentType)
			}
		})
	}
}

func TestErrorHandling_Integration_ValidationErrorWithFields(t *testing.T) {
	fields := map[string]string{
		"email":      "Email is required",
		"title":      "Title must be at least 3 characters",
		"start_time": "Start time must be in the future",
	}

	err := NewValidationError(fields)

	req := httptest.NewRequest(http.MethodPost, "/events", nil)
	req.Header.Set("Accept", "application/json")
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	HandleError(w, req, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != "VALIDATION_ERROR" {
		t.Errorf("Code = %v, want VALIDATION_ERROR", resp.Code)
	}

	if len(resp.Fields) != len(fields) {
		t.Errorf("Fields length = %v, want %v", len(resp.Fields), len(fields))
	}

	for field, expectedMsg := range fields {
		if gotMsg, ok := resp.Fields[field]; !ok {
			t.Errorf("Field %q not found in response", field)
		} else if gotMsg != expectedMsg {
			t.Errorf("Field %q message = %v, want %v", field, gotMsg, expectedMsg)
		}
	}
}

func TestErrorHandling_Integration_ErrorWrapping(t *testing.T) {
	underlyingErr := errors.New("database connection timeout")
	wrappedErr := &models.NotFoundError{Resource: "Event", ID: 123}

	apiErr := toAPIError(wrappedErr)

	if apiErr.Err == nil {
		t.Error("Expected wrapped error to be preserved")
	}

	if !errors.Is(apiErr, wrappedErr) {
		t.Error("Expected errors.Is to work with wrapped error")
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/json")
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	apiErr2 := &APIError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    "Database error",
		Err:        underlyingErr,
	}

	HandleError(w, req, apiErr2)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}

func TestErrorHandling_Integration_RequestIDPropagation(t *testing.T) {
	requestID := "unique-request-id-12345"

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "text/html")
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, requestID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	err := NewNotFoundError("Resource not found")
	HandleError(w, req, err)

	body := w.Body.String()
	if !strings.Contains(body, requestID) {
		t.Errorf("HTML response should contain request ID %q", requestID)
	}
}

func TestErrorHandling_Integration_NoRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "text/html")

	w := httptest.NewRecorder()

	err := NewNotFoundError("Resource not found")
	HandleError(w, req, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusNotFound)
	}

	body := w.Body.String()
	if !strings.Contains(body, "404") {
		t.Error("HTML response should still render without request ID")
	}
}
