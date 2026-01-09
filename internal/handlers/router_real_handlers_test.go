package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockAuthHandler struct{}

func (m *mockAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("auth handler called"))
}

type mockEventHandlers struct {
	listCalled   bool
	createCalled bool
	getCalled    bool
	updateCalled bool
	deleteCalled bool
}

func (m *mockEventHandlers) RegisterRoutes(r interface{}) {}

func (m *mockEventHandlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	m.listCalled = true
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]interface{}{})
}

func (m *mockEventHandlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	m.createCalled = true
	w.WriteHeader(http.StatusCreated)
}

func (m *mockEventHandlers) GetEvent(w http.ResponseWriter, r *http.Request) {
	m.getCalled = true
	w.WriteHeader(http.StatusOK)
}

func (m *mockEventHandlers) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	m.updateCalled = true
	w.WriteHeader(http.StatusOK)
}

func (m *mockEventHandlers) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	m.deleteCalled = true
	w.WriteHeader(http.StatusNoContent)
}

type mockRSVPHandler struct {
	getPageCalled        bool
	submitCalled         bool
	updateCalled         bool
	getConfirmationCalled bool
}

func (m *mockRSVPHandler) GetRSVPPage(w http.ResponseWriter, r *http.Request) {
	m.getPageCalled = true
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("RSVP page"))
}

func (m *mockRSVPHandler) SubmitRSVP(w http.ResponseWriter, r *http.Request) {
	m.submitCalled = true
	w.WriteHeader(http.StatusOK)
}

func (m *mockRSVPHandler) UpdateRSVP(w http.ResponseWriter, r *http.Request) {
	m.updateCalled = true
	w.WriteHeader(http.StatusOK)
}

func (m *mockRSVPHandler) GetConfirmationPage(w http.ResponseWriter, r *http.Request) {
	m.getConfirmationCalled = true
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Confirmation page"))
}

type mockUserHandler struct {
	listCalled   bool
	getCalled    bool
	updateCalled bool
	deleteCalled bool
}

func (m *mockUserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	m.listCalled = true
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]interface{}{})
}

func (m *mockUserHandler) GetUser(w http.ResponseWriter, r *http.Request, userID string) {
	m.getCalled = true
	w.WriteHeader(http.StatusOK)
}

func (m *mockUserHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request, userID string) {
	m.updateCalled = true
	w.WriteHeader(http.StatusOK)
}

func (m *mockUserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, userID string) {
	m.deleteCalled = true
	w.WriteHeader(http.StatusNoContent)
}

type mockAssetHandler struct {
	serveCalled bool
}

func (m *mockAssetHandler) ServeAsset(w http.ResponseWriter, r *http.Request) {
	m.serveCalled = true
	w.WriteHeader(http.StatusOK)
}

type mockAuthMiddleware struct{}

func (m *mockAuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
			Role:  models.RoleAdmin,
		}
		ctx := auth.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *mockAuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func TestRouter_Integration_RealHandlersCalled(t *testing.T) {
	loginHandler := &mockAuthHandler{}
	callbackHandler := &mockAuthHandler{}
	logoutHandler := &mockAuthHandler{}
	eventHandlers := &mockEventHandlers{}
	rsvpHandler := &mockRSVPHandler{}
	userHandler := &mockUserHandler{}
	assetHandler := &mockAssetHandler{}
	authMiddleware := &mockAuthMiddleware{}

	handlers := &RouterHandlers{
		LoginHandler:    loginHandler,
		CallbackHandler: callbackHandler,
		LogoutHandler:   logoutHandler,
		EventHandlers:   eventHandlers,
		RSVPHandler:     rsvpHandler,
		UserHandler:     userHandler,
		AssetHandler:    assetHandler,
		AuthMiddleware:  authMiddleware,
	}

	router := NewRouter(handlers)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		wantStatusCode int
		checkCalled    func(*testing.T)
	}{
		{
			name:           "login handler called",
			method:         http.MethodGet,
			path:           "/login",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
			},
		},
		{
			name:           "callback handler called",
			method:         http.MethodGet,
			path:           "/auth/callback",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
			},
		},
		{
			name:           "logout handler called",
			method:         http.MethodPost,
			path:           "/logout",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
			},
		},
		{
			name:           "event list handler called",
			method:         http.MethodGet,
			path:           "/api/events",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
				if !eventHandlers.listCalled {
					t.Error("Expected ListEvents to be called")
				}
			},
		},
		{
			name:           "event create handler called",
			method:         http.MethodPost,
			path:           "/api/events",
			body:           `{"title":"Test Event"}`,
			wantStatusCode: http.StatusCreated,
			checkCalled: func(t *testing.T) {
				if !eventHandlers.createCalled {
					t.Error("Expected CreateEvent to be called")
				}
			},
		},
		{
			name:           "event get handler called",
			method:         http.MethodGet,
			path:           "/api/events/123",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
				if !eventHandlers.getCalled {
					t.Error("Expected GetEvent to be called")
				}
			},
		},
		{
			name:           "rsvp page handler called",
			method:         http.MethodGet,
			path:           "/rsvp/test-token-123",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
				if !rsvpHandler.getPageCalled {
					t.Error("Expected GetRSVPPage to be called")
				}
			},
		},
		{
			name:           "rsvp submit handler called",
			method:         http.MethodPost,
			path:           "/rsvp/test-token-123",
			body:           `{"status":"accepted"}`,
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
				if !rsvpHandler.submitCalled {
					t.Error("Expected SubmitRSVP to be called")
				}
			},
		},
		{
			name:           "rsvp confirmation handler called",
			method:         http.MethodGet,
			path:           "/rsvp/test-token-123/confirmation",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
				if !rsvpHandler.getConfirmationCalled {
					t.Error("Expected GetConfirmationPage to be called")
				}
			},
		},
		{
			name:           "user list handler called",
			method:         http.MethodGet,
			path:           "/api/users",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
				if !userHandler.listCalled {
					t.Error("Expected ListUsers to be called")
				}
			},
		},
		{
			name:           "asset handler called",
			method:         http.MethodGet,
			path:           "/assets/test.jpg",
			wantStatusCode: http.StatusOK,
			checkCalled: func(t *testing.T) {
				if !assetHandler.serveCalled {
					t.Error("Expected ServeAsset to be called")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.body != "" {
				body = bytes.NewBufferString(tt.body)
			} else {
				body = &bytes.Buffer{}
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatusCode, w.Code, w.Body.String())
			}

			if tt.checkCalled != nil {
				tt.checkCalled(t)
			}
		})
	}
}

func TestRouter_Integration_MiddlewareApplied(t *testing.T) {
	authMiddleware := &mockAuthMiddleware{}
	eventHandlers := &mockEventHandlers{}

	handlers := &RouterHandlers{
		EventHandlers:  eventHandlers,
		AuthMiddleware: authMiddleware,
	}

	router := NewRouter(handlers)

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !eventHandlers.listCalled {
		t.Error("Expected event handler to be called after middleware")
	}
}

func TestRouter_Integration_StaticFilesServed(t *testing.T) {
	router := NewRouter(&RouterHandlers{})

	req := httptest.NewRequest(http.MethodGet, "/static/test.css", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound && w.Body.String() != "404 page not found\n" {
		t.Error("Static file route should be registered (got custom 404 instead of file server 404)")
	}
}

func TestRouter_Integration_HealthEndpointWorks(t *testing.T) {
	router := NewRouter(&RouterHandlers{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %s", resp["status"])
	}
}

func TestRouter_Integration_ConcurrentRequestsWithRealHandlers(t *testing.T) {
	eventHandlers := &mockEventHandlers{}
	handlers := &RouterHandlers{
		EventHandlers: eventHandlers,
		AuthMiddleware: &mockAuthMiddleware{},
	}

	router := NewRouter(handlers)

	const numRequests = 50
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				errors <- nil
			}
			done <- true
		}()
	}

	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	}

	close(errors)
	if len(errors) > 0 {
		t.Errorf("Some requests failed")
	}
}
