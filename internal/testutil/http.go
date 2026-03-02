package testutil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

// HTTPRequest is a builder for constructing *http.Request in tests.
// It sets sensible defaults and supports optional auth, CSRF, chi URL params,
// form body, and JSON body.
//
// Example:
//
//	req := testutil.NewHTTPRequest("GET", "/events").
//	    WithUser(adminUser).
//	    WithCSRF().
//	    Build()
type HTTPRequest struct {
	method        string
	target        string
	body          io.Reader
	urlParams     map[string]string
	user          *models.User
	contextValues []contextEntry
	headers       map[string]string
}

type contextEntry struct {
	key   any
	value any
}

// NewHTTPRequest returns an HTTPRequest builder for the given method and target URL.
func NewHTTPRequest(method, target string) *HTTPRequest {
	return &HTTPRequest{
		method:        method,
		target:        target,
		urlParams:     make(map[string]string),
		contextValues: nil,
		headers:       make(map[string]string),
	}
}

// WithUser attaches the user to the request context (simulates authentication).
func (r *HTTPRequest) WithUser(user *models.User) *HTTPRequest {
	r.user = user
	return r
}

// WithAdminUser attaches a test admin user to the request context.
func (r *HTTPRequest) WithAdminUser() *HTTPRequest {
	r.user = &models.User{
		ID:    1,
		Email: "admin@test.example.com",
		Name:  "Test Admin",
		Role:  models.RoleAdmin,
	}
	return r
}

// WithEventManagerUser attaches a test event manager user to the request context.
func (r *HTTPRequest) WithEventManagerUser() *HTTPRequest {
	r.user = &models.User{
		ID:    2,
		Email: "manager@test.example.com",
		Name:  "Test Manager",
		Role:  models.RoleEventManager,
	}
	return r
}

// WithContextValue adds an arbitrary key-value pair to the request context.
// Use this to inject middleware-managed values (e.g. CSRF tokens) without
// creating an import cycle.
//
// Example:
//
//	req := testutil.NewHTTPRequest("POST", "/events").
//	    WithContextValue(middleware.CSRFTokenKey, "test-csrf-token").
//	    Build()
func (r *HTTPRequest) WithContextValue(key, value any) *HTTPRequest {
	r.contextValues = append(r.contextValues, contextEntry{key: key, value: value})
	return r
}

// WithURLParam adds a chi URL parameter (e.g. "id" -> "42").
func (r *HTTPRequest) WithURLParam(key, value string) *HTTPRequest {
	r.urlParams[key] = value
	return r
}

// WithFormBody sets the request body to URL-encoded form values and
// sets Content-Type to application/x-www-form-urlencoded.
func (r *HTTPRequest) WithFormBody(values url.Values) *HTTPRequest {
	r.body = strings.NewReader(values.Encode())
	r.headers["Content-Type"] = "application/x-www-form-urlencoded"
	return r
}

// WithBody sets the request body to the provided reader.
func (r *HTTPRequest) WithBody(body io.Reader) *HTTPRequest {
	r.body = body
	return r
}

// WithHeader sets a request header.
func (r *HTTPRequest) WithHeader(key, value string) *HTTPRequest {
	r.headers[key] = value
	return r
}

// Build constructs the final *http.Request.
func (r *HTTPRequest) Build() *http.Request {
	req := httptest.NewRequest(r.method, r.target, r.body)

	// Apply chi URL params
	if len(r.urlParams) > 0 {
		rctx := chi.NewRouteContext()
		for k, v := range r.urlParams {
			rctx.URLParams.Add(k, v)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	// Apply user auth context
	ctx := req.Context()
	if r.user != nil {
		ctx = auth.WithUser(ctx, r.user)
	}

	// Apply extra context values
	for _, e := range r.contextValues {
		ctx = context.WithValue(ctx, e.key, e.value)
	}

	req = req.WithContext(ctx)

	// Apply headers
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	return req
}

// NewResponseRecorder returns a new httptest.ResponseRecorder.
// Convenience alias to avoid importing net/http/httptest in every test file.
func NewResponseRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// AssertStatus is a test helper that fails t if the recorder's status code
// does not match want. Pass t.Helper() at the call site for cleaner output.
//
// Example:
//
//	w := testutil.NewResponseRecorder()
//	handler.ServeHTTP(w, req)
//	testutil.AssertStatus(t, w, http.StatusOK)
func AssertStatus(t interface {
	Helper()
	Errorf(format string, args ...any)
}, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d, want %d; body: %s", w.Code, want, w.Body.String())
	}
}

// AssertBodyContains is a test helper that fails t if the response body does
// not contain the expected substring.
//
// Example:
//
//	testutil.AssertBodyContains(t, w, "Welcome")
func AssertBodyContains(t interface {
	Helper()
	Errorf(format string, args ...any)
}, w *httptest.ResponseRecorder, want string) {
	t.Helper()
	if !strings.Contains(w.Body.String(), want) {
		t.Errorf("body does not contain %q; body: %s", want, w.Body.String())
	}
}

// AssertRedirect is a test helper that fails t if the response is not a
// redirect to the expected location.
//
// Example:
//
//	testutil.AssertRedirect(t, w, "/events")
func AssertRedirect(t interface {
	Helper()
	Errorf(format string, args ...any)
}, w *httptest.ResponseRecorder, wantLocation string) {
	t.Helper()
	code := w.Code
	if code < 300 || code >= 400 {
		t.Errorf("status = %d, want 3xx redirect", code)
		return
	}
	got := w.Header().Get("Location")
	if got != wantLocation {
		t.Errorf("Location = %q, want %q", got, wantLocation)
	}
}
