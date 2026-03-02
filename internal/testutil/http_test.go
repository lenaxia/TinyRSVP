package testutil_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/testutil"
)

func TestNewHTTPRequest_Defaults(t *testing.T) {
	req := testutil.NewHTTPRequest("GET", "/test").Build()
	if req.Method != "GET" {
		t.Errorf("Method = %q, want GET", req.Method)
	}
	if req.URL.Path != "/test" {
		t.Errorf("URL.Path = %q, want /test", req.URL.Path)
	}
}

func TestNewHTTPRequest_WithUser(t *testing.T) {
	user := &models.User{ID: 10, Email: "x@example.com", Role: models.RoleGuest}
	req := testutil.NewHTTPRequest("GET", "/test").WithUser(user).Build()

	got, ok := auth.UserFromContext(req.Context())
	if !ok || got == nil {
		t.Fatal("expected user in context, got nil")
	}
	if got.ID != 10 {
		t.Errorf("user ID = %d, want 10", got.ID)
	}
}

func TestNewHTTPRequest_WithAdminUser(t *testing.T) {
	req := testutil.NewHTTPRequest("GET", "/admin").WithAdminUser().Build()
	user, ok := auth.UserFromContext(req.Context())
	if !ok || user == nil {
		t.Fatal("expected user in context, got nil")
	}
	if user.Role != models.RoleAdmin {
		t.Errorf("role = %q, want admin", user.Role)
	}
}

func TestNewHTTPRequest_WithEventManagerUser(t *testing.T) {
	req := testutil.NewHTTPRequest("GET", "/events").WithEventManagerUser().Build()
	user, ok := auth.UserFromContext(req.Context())
	if !ok || user == nil {
		t.Fatal("expected user in context, got nil")
	}
	if user.Role != models.RoleEventManager {
		t.Errorf("role = %q, want event_manager", user.Role)
	}
}

func TestNewHTTPRequest_WithContextValue(t *testing.T) {
	req := testutil.NewHTTPRequest("POST", "/form").
		WithContextValue(middleware.CSRFTokenKey, "test-csrf-token").
		Build()
	token, ok := req.Context().Value(middleware.CSRFTokenKey).(string)
	if !ok || token != "test-csrf-token" {
		t.Errorf("CSRF token = %q, want test-csrf-token", token)
	}
}

func TestNewHTTPRequest_WithURLParam(t *testing.T) {
	req := testutil.NewHTTPRequest("GET", "/events/42").
		WithURLParam("id", "42").
		Build()

	rctx := chi.RouteContext(req.Context())
	if rctx == nil {
		t.Fatal("chi route context is nil")
	}
	if got := rctx.URLParam("id"); got != "42" {
		t.Errorf("URLParam id = %q, want 42", got)
	}
}

func TestNewHTTPRequest_WithFormBody(t *testing.T) {
	values := url.Values{"title": {"My Event"}, "status": {"draft"}}
	req := testutil.NewHTTPRequest("POST", "/events").WithFormBody(values).Build()

	if ct := req.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", ct)
	}

	if err := req.ParseForm(); err != nil {
		t.Fatalf("ParseForm: %v", err)
	}
	if req.FormValue("title") != "My Event" {
		t.Errorf("form title = %q, want My Event", req.FormValue("title"))
	}
}

func TestNewHTTPRequest_WithBody(t *testing.T) {
	req := testutil.NewHTTPRequest("POST", "/api").
		WithBody(strings.NewReader(`{"key":"value"}`)).
		WithHeader("Content-Type", "application/json").
		Build()

	if ct := req.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestAssertStatus_Pass(t *testing.T) {
	w := testutil.NewResponseRecorder()
	w.WriteHeader(http.StatusOK)
	// Should not fail.
	testutil.AssertStatus(t, w, http.StatusOK)
}

func TestAssertStatus_Fail(t *testing.T) {
	// Use a sub-test with a fake T to verify the helper triggers failure.
	failed := false
	fakeT := &fakeTestingT{onError: func() { failed = true }}
	w := testutil.NewResponseRecorder()
	w.WriteHeader(http.StatusForbidden)
	testutil.AssertStatus(fakeT, w, http.StatusOK)
	if !failed {
		t.Error("AssertStatus should have failed when code mismatches")
	}
}

func TestAssertBodyContains_Pass(t *testing.T) {
	w := testutil.NewResponseRecorder()
	w.WriteString("Hello World") //nolint:errcheck
	testutil.AssertBodyContains(t, w, "World")
}

func TestAssertBodyContains_Fail(t *testing.T) {
	failed := false
	fakeT := &fakeTestingT{onError: func() { failed = true }}
	w := testutil.NewResponseRecorder()
	w.WriteString("Hello") //nolint:errcheck
	testutil.AssertBodyContains(fakeT, w, "World")
	if !failed {
		t.Error("AssertBodyContains should have failed when substring missing")
	}
}

func TestAssertRedirect_Pass(t *testing.T) {
	w := testutil.NewResponseRecorder()
	w.Header().Set("Location", "/events")
	w.WriteHeader(http.StatusFound)
	testutil.AssertRedirect(t, w, "/events")
}

func TestAssertRedirect_WrongStatus(t *testing.T) {
	failed := false
	fakeT := &fakeTestingT{onError: func() { failed = true }}
	w := testutil.NewResponseRecorder()
	w.WriteHeader(http.StatusOK)
	testutil.AssertRedirect(fakeT, w, "/events")
	if !failed {
		t.Error("AssertRedirect should have failed for non-redirect status")
	}
}

func TestAssertRedirect_WrongLocation(t *testing.T) {
	failed := false
	fakeT := &fakeTestingT{onError: func() { failed = true }}
	w := testutil.NewResponseRecorder()
	w.Header().Set("Location", "/other")
	w.WriteHeader(http.StatusFound)
	testutil.AssertRedirect(fakeT, w, "/events")
	if !failed {
		t.Error("AssertRedirect should have failed for wrong Location")
	}
}

// fakeTestingT implements the minimal interface used by testutil helpers.
type fakeTestingT struct {
	onError func()
}

func (f *fakeTestingT) Helper() {}
func (f *fakeTestingT) Errorf(_ string, _ ...any) {
	if f.onError != nil {
		f.onError()
	}
}
