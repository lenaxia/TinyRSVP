package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouter_NotFoundHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		acceptHeader   string
		wantStatusCode int
		wantBodyContains string
	}{
		{
			name:           "API request returns JSON",
			path:           "/api/nonexistent",
			acceptHeader:   "application/json",
			wantStatusCode: http.StatusNotFound,
			wantBodyContains: "not found",
		},
		{
			name:           "Web request returns HTML",
			path:           "/nonexistent",
			acceptHeader:   "text/html",
			wantStatusCode: http.StatusNotFound,
			wantBodyContains: "404",
		},
		{
			name:           "No accept header defaults to HTML",
			path:           "/missing",
			acceptHeader:   "",
			wantStatusCode: http.StatusNotFound,
			wantBodyContains: "404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			w := httptest.NewRecorder()

			NotFoundHandler(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("NotFoundHandler() status = %d, want %d", w.Code, tt.wantStatusCode)
			}

			body := w.Body.String()
			if !strings.Contains(strings.ToLower(body), strings.ToLower(tt.wantBodyContains)) {
				t.Errorf("NotFoundHandler() body = %q, want to contain %q", body, tt.wantBodyContains)
			}
		})
	}
}

func TestRouter_MethodNotAllowedHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		acceptHeader   string
		wantStatusCode int
		wantBodyContains string
	}{
		{
			name:           "API request returns JSON",
			path:           "/api/events",
			acceptHeader:   "application/json",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBodyContains: "method not allowed",
		},
		{
			name:           "Web request returns HTML",
			path:           "/events",
			acceptHeader:   "text/html",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBodyContains: "405",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			w := httptest.NewRecorder()

			MethodNotAllowedHandler(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("MethodNotAllowedHandler() status = %d, want %d", w.Code, tt.wantStatusCode)
			}

			body := w.Body.String()
			if !strings.Contains(strings.ToLower(body), strings.ToLower(tt.wantBodyContains)) {
				t.Errorf("MethodNotAllowedHandler() body = %q, want to contain %q", body, tt.wantBodyContains)
			}
		})
	}
}

func TestRouter_GetInt64Param(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantVal int64
		wantErr bool
	}{
		{
			name:    "valid positive integer",
			value:   "123",
			wantVal: 123,
			wantErr: false,
		},
		{
			name:    "valid large integer",
			value:   "9223372036854775807",
			wantVal: 9223372036854775807,
			wantErr: false,
		},
		{
			name:    "zero",
			value:   "0",
			wantVal: 0,
			wantErr: false,
		},
		{
			name:    "negative integer",
			value:   "-1",
			wantVal: 0,
			wantErr: true,
		},
		{
			name:    "invalid string",
			value:   "abc",
			wantVal: 0,
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantVal: 0,
			wantErr: true,
		},
		{
			name:    "float value",
			value:   "123.45",
			wantVal: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := GetInt64Param(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt64Param() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("GetInt64Param() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestRouter_GetStringParam(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantVal string
		wantErr bool
	}{
		{
			name:    "valid string",
			value:   "test-token",
			wantVal: "test-token",
			wantErr: false,
		},
		{
			name:    "alphanumeric with dashes",
			value:   "abc-123-def",
			wantVal: "abc-123-def",
			wantErr: false,
		},
		{
			name:    "empty string",
			value:   "",
			wantVal: "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantVal: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := GetStringParam(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStringParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("GetStringParam() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestRouter_IsAPIRequest(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		acceptHeader string
		want         bool
	}{
		{
			name:         "API path with /api/ prefix",
			path:         "/api/events",
			acceptHeader: "",
			want:         true,
		},
		{
			name:         "API path with JSON accept",
			path:         "/events",
			acceptHeader: "application/json",
			want:         true,
		},
		{
			name:         "Web path with HTML accept",
			path:         "/events",
			acceptHeader: "text/html",
			want:         false,
		},
		{
			name:         "Web path no accept header",
			path:         "/events",
			acceptHeader: "",
			want:         false,
		},
		{
			name:         "Root path",
			path:         "/",
			acceptHeader: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			if got := IsAPIRequest(req); got != tt.want {
				t.Errorf("IsAPIRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRouter(t *testing.T) {
	router := NewRouter(nil)
	if router == nil {
		t.Fatal("NewRouter() returned nil")
	}
	if router.mux == nil {
		t.Error("NewRouter() mux is nil")
	}
}

func TestRouter_ServeHTTP(t *testing.T) {
	router := NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Router.ServeHTTP() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRouter_HealthEndpoint(t *testing.T) {
	router := NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health endpoint status = %d, want %d", w.Code, http.StatusOK)
	}

	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "ok") {
		t.Errorf("Health endpoint body = %q, want to contain 'ok'", string(body))
	}
}

func TestRouter_RouteGroups(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		method     string
		wantStatus int
	}{
		{
			name:       "auth route group exists",
			path:       "/auth/login",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
		{
			name:       "events route group exists",
			path:       "/api/events",
			method:     http.MethodGet,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invites route group exists",
			path:       "/api/invites/cleanup",
			method:     http.MethodPost,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "rsvp route group exists",
			path:       "/rsvp/test-token",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
		{
			name:       "static assets route exists",
			path:       "/static/test.css",
			method:     http.MethodGet,
			wantStatus: http.StatusNotFound,
		},
	}

	router := NewRouter(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Route %s %s status = %d, want %d", tt.method, tt.path, w.Code, tt.wantStatus)
			}
		})
	}
}
