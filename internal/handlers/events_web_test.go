package handlers

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestNewEventWebHandlers(t *testing.T) {
	mockService := &mockEventService{}
	tmpl := template.New("test")
	handlers := NewEventWebHandlers(mockService, nil, tmpl)

	if handlers == nil {
		t.Fatal("NewEventWebHandlers returned nil")
	}

	if handlers.service == nil {
		t.Error("service is nil")
	}

	if handlers.templates == nil {
		t.Error("templates is nil")
	}
}

func TestEventWebHandlers_ListEventsPage(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:  "list events page with events",
			query: "",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
					return []*models.Event{
						{
							ID:        1,
							Title:     "Test Event",
							StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
							Timezone:  "America/Los_Angeles",
							Status:    models.EventStatusDraft,
							CreatedBy: 1,
							Version:   1,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantBody:   "Test Event",
		},
		{
			name:  "list events page empty",
			query: "",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
					return []*models.Event{}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantBody:   "No Events Found",
		},
		{
			name:  "list events with status filter",
			query: "?status=published",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
					if filters.Status == nil || *filters.Status != models.EventStatusPublished {
						t.Error("Expected published status filter")
					}
					return []*models.Event{}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "unauthorized user",
			query: "",
			user: &models.User{
				ID:   1,
				Role: models.RoleGuest,
			},
			setupMock: func(m *mockEventService) {
				m.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
					return nil, &models.PermissionDeniedError{
						Action:   "list events",
						Resource: "Event",
					}
				}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			tmpl := template.Must(template.New("event_list.html").Parse(`
				<!DOCTYPE html>
				<html>
				<body>
					{{if .Error}}
						<div>Error: {{.Error}}</div>
					{{else if eq (len .Events) 0}}
						<div>No Events Found</div>
					{{else}}
						{{range .Events}}
							<div>{{.Title}}</div>
						{{end}}
					{{end}}
				</body>
				</html>
			`))

			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			req := httptest.NewRequest("GET", "/events"+tt.query, nil)
			ctx := auth.WithUser(req.Context(), tt.user)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.ListEventsPage(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventWebHandlers_NewEventForm(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		wantStatus int
		wantBody   string
	}{
		{
			name: "show new event form",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusOK,
			wantBody:   "Create Event",
		},
		{
			name: "unauthorized user",
			user: &models.User{
				ID:   1,
				Role: models.RoleGuest,
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			tmpl := template.Must(template.New("event_form.html").Parse(`
				<!DOCTYPE html>
				<html>
				<body>
					<h1>Create Event</h1>
					<form>
						<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
					</form>
				</body>
				</html>
			`))

			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			req := httptest.NewRequest("GET", "/events/new", nil)
			ctx := auth.WithUser(req.Context(), tt.user)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.NewEventForm(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}

			if tt.wantStatus == http.StatusOK && !strings.Contains(w.Body.String(), "test-csrf-token") {
				t.Error("CSRF token not injected into form")
			}
		})
	}
}

func TestEventWebHandlers_EditEventForm(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "show edit form for existing event",
			eventID: "1",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
						Timezone:  "America/Los_Angeles",
						Status:    models.EventStatusDraft,
						CreatedBy: 1,
						Version:   1,
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantBody:   "Edit Event",
		},
		{
			name:    "event not found",
			eventID: "999",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.NotFoundError{
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Event not found",
		},
		{
			name:    "unauthorized user",
			eventID: "1",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.PermissionDeniedError{
						Action:   "view event",
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:    "invalid event ID",
			eventID: "invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid event ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			tmpl := template.Must(template.New("event_form.html").Parse(`
				<!DOCTYPE html>
				<html>
				<body>
					<h1>Edit Event</h1>
					<form>
						<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
						<input type="text" name="title" value="{{.Event.Title}}">
					</form>
				</body>
				</html>
			`))

			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			req := httptest.NewRequest("GET", "/events/"+tt.eventID+"/edit", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.EditEventForm(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}

			if tt.wantStatus == http.StatusOK && !strings.Contains(w.Body.String(), "test-csrf-token") {
				t.Error("CSRF token not injected into form")
			}
		})
	}
}

func TestEventWebHandlers_GetEventPage(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "view event details",
			eventID: "1",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
						Timezone:  "America/Los_Angeles",
						Status:    models.EventStatusDraft,
						CreatedBy: 1,
						Version:   1,
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantBody:   "Test Event",
		},
		{
			name:    "event not found",
			eventID: "999",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.NotFoundError{
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Event not found",
		},
		{
			name:    "unauthorized user",
			eventID: "1",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.PermissionDeniedError{
						Action:   "view event",
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:    "invalid event ID",
			eventID: "invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid event ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			tmpl := template.Must(template.New("event_detail.html").Parse(`
				<!DOCTYPE html>
				<html>
				<body>
					<h1>{{.Event.Title}}</h1>
				</body>
				</html>
			`))

			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			req := httptest.NewRequest("GET", "/events/"+tt.eventID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.GetEventPage(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventWebHandlers_CreateEventFromForm(t *testing.T) {
	tests := []struct {
		name       string
		formData   url.Values
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "valid form submission",
			formData: url.Values{
				"title":         []string{"Birthday Party"},
				"description":   []string{"A fun celebration"},
				"start_time":    []string{"2026-06-15T14:00"},
				"timezone":      []string{"America/Los_Angeles"},
				"location":      []string{"123 Main St"},
				"max_plus_ones": []string{"2"},
				"csrf_token":    []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.CreateEventFunc = func(ctx context.Context, e *models.Event) error {
					e.ID = 1
					e.CreatedAt = time.Now()
					e.UpdatedAt = time.Now()
					return nil
				}
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name: "missing required field",
			formData: url.Values{
				"description": []string{"Missing title"},
				"start_time":  []string{"2026-06-15T14:00"},
				"timezone":    []string{"America/Los_Angeles"},
				"csrf_token":  []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "title",
		},
		{
			name: "invalid datetime format",
			formData: url.Values{
				"title":        []string{"Test Event"},
				"start_time":   []string{"invalid-date"},
				"timezone":     []string{"America/Los_Angeles"},
				"csrf_token":   []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "start_time",
		},
		{
			name: "unauthorized user",
			formData: url.Values{
				"title":      []string{"Test Event"},
				"start_time": []string{"2026-06-15T14:00"},
				"timezone":   []string{"America/Los_Angeles"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleGuest,
			},
			setupMock: func(m *mockEventService) {
				m.CreateEventFunc = func(ctx context.Context, e *models.Event) error {
					return &models.PermissionDeniedError{
						Action:   "create event",
						Resource: "Event",
					}
				}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			tmpl := template.New("test")
			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			req := httptest.NewRequest("POST", "/events", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			ctx := auth.WithUser(req.Context(), tt.user)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.CreateEventFromForm(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}

			if tt.wantStatus == http.StatusSeeOther {
				location := w.Header().Get("Location")
				if !strings.HasPrefix(location, "/events/") {
					t.Errorf("Expected redirect to /events/{id}, got %s", location)
				}
			}
		})
	}
}

func TestEventWebHandlers_UpdateEventFromForm(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		formData   url.Values
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "valid update",
			eventID: "1",
			formData: url.Values{
				"title":      []string{"Updated Title"},
				"version":    []string{"1"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Original Title",
						StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
						Timezone:  "America/Los_Angeles",
						Status:    models.EventStatusDraft,
						CreatedBy: 1,
						Version:   1,
					}, nil
				}
				m.UpdateEventFunc = func(ctx context.Context, e *models.Event) error {
					return nil
				}
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:    "version conflict",
			eventID: "1",
			formData: url.Values{
				"title":      []string{"Updated Title"},
				"version":    []string{"1"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Original Title",
						StartTime: time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
						Timezone:  "America/Los_Angeles",
						Status:    models.EventStatusDraft,
						CreatedBy: 1,
						Version:   1,
					}, nil
				}
				m.UpdateEventFunc = func(ctx context.Context, e *models.Event) error {
					return &models.VersionConflictError{
						ResourceType: "event",
						ResourceID:   1,
						Expected:     1,
						Actual:       2,
					}
				}
			},
			wantStatus: http.StatusConflict,
			wantBody:   "version conflict",
		},
		{
			name:    "unauthorized user",
			eventID: "1",
			formData: url.Values{
				"title":      []string{"Updated Title"},
				"version":    []string{"1"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.PermissionDeniedError{
						Action:   "view event",
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:    "invalid event ID",
			eventID: "invalid",
			formData: url.Values{
				"title":      []string{"Updated Title"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid event ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			tmpl := template.New("test")
			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			req := httptest.NewRequest("POST", "/events/"+tt.eventID, strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.UpdateEventFromForm(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}

			if tt.wantStatus == http.StatusSeeOther {
				location := w.Header().Get("Location")
				if !strings.HasPrefix(location, "/events/") {
					t.Errorf("Expected redirect to /events/{id}, got %s", location)
				}
			}
		})
	}
}

func TestEventWebHandlers_PublishEventAction(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "successful publish",
			eventID: "1",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.PublishEventFunc = func(ctx context.Context, id int64) error {
					return nil
				}
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:    "unauthorized user",
			eventID: "1",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.PublishEventFunc = func(ctx context.Context, id int64) error {
					return &models.PermissionDeniedError{
						Action:   "publish event",
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:    "event not found",
			eventID: "999",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.PublishEventFunc = func(ctx context.Context, id int64) error {
					return &models.NotFoundError{
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Event not found",
		},
		{
			name:    "invalid event ID",
			eventID: "invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid event ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			tmpl := template.New("test")
			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			formData := url.Values{
				"csrf_token": []string{"test-csrf-token"},
			}

			req := httptest.NewRequest("POST", "/events/"+tt.eventID+"/publish", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.PublishEventAction(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}

			if tt.wantStatus == http.StatusSeeOther {
				location := w.Header().Get("Location")
				if !strings.HasPrefix(location, "/events/") {
					t.Errorf("Expected redirect to /events/{id}, got %s", location)
				}
			}
		})
	}
}

func TestEventWebHandlers_CancelEventAction(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		formData   url.Values
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "successful cancel",
			eventID: "1",
			formData: url.Values{
				"reason":     []string{"Event location is no longer available"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.CancelEventFunc = func(ctx context.Context, id int64, reason string) error {
					return nil
				}
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:    "missing reason",
			eventID: "1",
			formData: url.Values{
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "reason",
		},
		{
			name:    "reason too short",
			eventID: "1",
			formData: url.Values{
				"reason":     []string{"short"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "reason",
		},
		{
			name:    "unauthorized user",
			eventID: "1",
			formData: url.Values{
				"reason":     []string{"Event location is no longer available"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.CancelEventFunc = func(ctx context.Context, id int64, reason string) error {
					return &models.PermissionDeniedError{
						Action:   "cancel event",
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:    "invalid event ID",
			eventID: "invalid",
			formData: url.Values{
				"reason":     []string{"Event location is no longer available"},
				"csrf_token": []string{"test-csrf-token"},
			},
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid event ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			tmpl := template.New("test")
			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			req := httptest.NewRequest("POST", "/events/"+tt.eventID+"/cancel", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.CancelEventAction(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}

			if tt.wantStatus == http.StatusSeeOther {
				location := w.Header().Get("Location")
				if !strings.HasPrefix(location, "/events/") {
					t.Errorf("Expected redirect to /events/{id}, got %s", location)
				}
			}
		})
	}
}

func TestEventWebHandlers_DeleteEventAction(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "successful delete",
			eventID: "1",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.DeleteEventFunc = func(ctx context.Context, id int64) error {
					return nil
				}
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:    "unauthorized user",
			eventID: "1",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.DeleteEventFunc = func(ctx context.Context, id int64) error {
					return &models.PermissionDeniedError{
						Action:   "delete event",
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:    "event not found",
			eventID: "999",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.DeleteEventFunc = func(ctx context.Context, id int64) error {
					return &models.NotFoundError{
						Resource: "Event",
						ID:       id,
					}
				}
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Event not found",
		},
		{
			name:    "invalid event ID",
			eventID: "invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid event ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			tmpl := template.New("test")
			handlers := NewEventWebHandlers(mockService, nil, tmpl)

			formData := url.Values{
				"csrf_token": []string{"test-csrf-token"},
			}

			req := httptest.NewRequest("POST", "/events/"+tt.eventID+"/delete", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			ctx = context.WithValue(ctx, middleware.CSRFTokenKey, "test-csrf-token")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.DeleteEventAction(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}

			if tt.wantStatus == http.StatusSeeOther {
				location := w.Header().Get("Location")
				if location != "/events" {
					t.Errorf("Expected redirect to /events, got %s", location)
				}
			}
		})
	}
}

func TestEventWebHandlers_FormDataParsing(t *testing.T) {
	tests := []struct {
		name       string
		formData   url.Values
		wantTitle  string
		wantDesc   *string
		wantLoc    *string
		wantMaxPO  int
		wantErr    bool
	}{
		{
			name: "all fields provided",
			formData: url.Values{
				"title":         []string{"Test Event"},
				"description":   []string{"Test Description"},
				"location":      []string{"Test Location"},
				"start_time":    []string{"2026-06-15T14:00"},
				"timezone":      []string{"America/Los_Angeles"},
				"max_plus_ones": []string{"2"},
			},
			wantTitle: "Test Event",
			wantDesc:  stringPtr("Test Description"),
			wantLoc:   stringPtr("Test Location"),
			wantMaxPO: 2,
			wantErr:   false,
		},
		{
			name: "optional fields empty",
			formData: url.Values{
				"title":         []string{"Test Event"},
				"description":   []string{""},
				"location":      []string{""},
				"start_time":    []string{"2026-06-15T14:00"},
				"timezone":      []string{"America/Los_Angeles"},
				"max_plus_ones": []string{"0"},
			},
			wantTitle: "Test Event",
			wantDesc:  nil,
			wantLoc:   nil,
			wantMaxPO: 0,
			wantErr:   false,
		},
		{
			name: "missing required field",
			formData: url.Values{
				"description": []string{"Test Description"},
				"start_time":  []string{"2026-06-15T14:00"},
				"timezone":    []string{"America/Los_Angeles"},
			},
			wantErr: true,
		},
		{
			name: "invalid max_plus_ones",
			formData: url.Values{
				"title":         []string{"Test Event"},
				"start_time":    []string{"2026-06-15T14:00"},
				"timezone":      []string{"America/Los_Angeles"},
				"max_plus_ones": []string{"invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parseEventFormData(tt.formData)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if event.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", event.Title, tt.wantTitle)
			}

			if tt.wantDesc == nil && event.Description != nil {
				t.Errorf("Description = %v, want nil", event.Description)
			} else if tt.wantDesc != nil && (event.Description == nil || *event.Description != *tt.wantDesc) {
				t.Errorf("Description = %v, want %v", event.Description, tt.wantDesc)
			}

			if tt.wantLoc == nil && event.Location != nil {
				t.Errorf("Location = %v, want nil", event.Location)
			} else if tt.wantLoc != nil && (event.Location == nil || *event.Location != *tt.wantLoc) {
				t.Errorf("Location = %v, want %v", event.Location, tt.wantLoc)
			}

			if event.MaxPlusOnes != tt.wantMaxPO {
				t.Errorf("MaxPlusOnes = %d, want %d", event.MaxPlusOnes, tt.wantMaxPO)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}