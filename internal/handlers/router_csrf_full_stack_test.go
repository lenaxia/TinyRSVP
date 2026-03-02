package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/middleware"
)

func TestRouter_CSRFWithEventCreation(t *testing.T) {
	t.Run("full stack event creation with CSRF", func(t *testing.T) {
		router := NewRouter(&RouterHandlers{
			EventWebHandlers: &mockEventWebHandlers{},
			AuthMiddleware:   &mockAuthMiddlewareForCSRF{},
		})

		req1 := httptest.NewRequest(http.MethodGet, "/events/new", nil)
		rec1 := httptest.NewRecorder()
		router.ServeHTTP(rec1, req1)

		var csrfCookie *http.Cookie
		for _, c := range rec1.Result().Cookies() {
			if c.Name == middleware.CSRFCookieName {
				csrfCookie = c
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("No CSRF cookie from GET /events/new")
		}

		t.Logf("CSRF cookie from GET: %s", csrfCookie.Value)

		formData := url.Values{}
		formData.Set("csrf_token", csrfCookie.Value)
		formData.Set("title", "Test Event")
		formData.Set("start_time", "2026-01-15T10:00")
		formData.Set("timezone", "America/Los_Angeles")
		formData.Set("max_plus_ones", "0")
		formData.Set("action", "publish")

		req2 := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(formData.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req2.AddCookie(csrfCookie)

		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)

		t.Logf("POST response status: %d", rec2.Code)
		t.Logf("POST response body: %s", rec2.Body.String())

		if rec2.Code == http.StatusForbidden {
			t.Errorf("Got 403 with valid CSRF token through full router stack!")

			for _, c := range rec2.Result().Cookies() {
				t.Logf("Response cookie: %s = %s", c.Name, c.Value)
			}
		}
	})
}

type mockEventWebHandlers struct{}

func (m *mockEventWebHandlers) ListEventsPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockEventWebHandlers) NewEventForm(w http.ResponseWriter, r *http.Request) {
	csrfToken := middleware.GetCSRFToken(r.Context())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("CSRF:" + csrfToken))
}

func (m *mockEventWebHandlers) EditEventForm(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockEventWebHandlers) GetEventPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockEventWebHandlers) CreateEventFromForm(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Event created"))
}

func (m *mockEventWebHandlers) UpdateEventFromForm(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockEventWebHandlers) PublishEventAction(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockEventWebHandlers) CancelEventAction(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockEventWebHandlers) DeleteEventAction(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type mockAuthMiddlewareForCSRF struct{}

func (m *mockAuthMiddlewareForCSRF) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (m *mockAuthMiddlewareForCSRF) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
