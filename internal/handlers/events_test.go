package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockEventService struct {
	CreateEventFunc        func(ctx context.Context, event *models.Event) error
	GetEventFunc           func(ctx context.Context, id int64) (*models.Event, error)
	UpdateEventFunc        func(ctx context.Context, event *models.Event) error
	DeleteEventFunc        func(ctx context.Context, id int64) error
	ListEventsFunc         func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error)
	PublishEventFunc       func(ctx context.Context, id int64) error
	CancelEventFunc        func(ctx context.Context, id int64, reason string) error
	ArchiveEventFunc       func(ctx context.Context, id int64) error
	GetEventsToArchiveFunc func(ctx context.Context, daysAfterEvent int) ([]*models.Event, error)
}

func (m *mockEventService) CreateEvent(ctx context.Context, event *models.Event) error {
	if m.CreateEventFunc != nil {
		return m.CreateEventFunc(ctx, event)
	}
	return nil
}

func (m *mockEventService) GetEvent(ctx context.Context, id int64) (*models.Event, error) {
	if m.GetEventFunc != nil {
		return m.GetEventFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockEventService) UpdateEvent(ctx context.Context, event *models.Event) error {
	if m.UpdateEventFunc != nil {
		return m.UpdateEventFunc(ctx, event)
	}
	return nil
}

func (m *mockEventService) DeleteEvent(ctx context.Context, id int64) error {
	if m.DeleteEventFunc != nil {
		return m.DeleteEventFunc(ctx, id)
	}
	return nil
}

func (m *mockEventService) ListEvents(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
	if m.ListEventsFunc != nil {
		return m.ListEventsFunc(ctx, filters)
	}
	return nil, nil
}

func (m *mockEventService) PublishEvent(ctx context.Context, id int64) error {
	if m.PublishEventFunc != nil {
		return m.PublishEventFunc(ctx, id)
	}
	return nil
}

func (m *mockEventService) CancelEvent(ctx context.Context, id int64, reason string) error {
	if m.CancelEventFunc != nil {
		return m.CancelEventFunc(ctx, id, reason)
	}
	return nil
}

func (m *mockEventService) ArchiveEvent(ctx context.Context, id int64) error {
	if m.ArchiveEventFunc != nil {
		return m.ArchiveEventFunc(ctx, id)
	}
	return nil
}

func (m *mockEventService) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	if m.GetEventsToArchiveFunc != nil {
		return m.GetEventsToArchiveFunc(ctx, daysAfterEvent)
	}
	return nil, nil
}

func TestNewEventHandlers(t *testing.T) {
	mockService := &mockEventService{}
	handlers := NewEventHandlers(mockService)

	if handlers == nil {
		t.Fatal("NewEventHandlers returned nil")
	}

	if handlers.service == nil {
		t.Error("service is nil")
	}
}

func TestEventHandlers_CreateEvent(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "valid event creation",
			body: `{
				"title": "Birthday Party",
				"start_time": "2026-06-15T14:00:00-07:00",
				"timezone": "America/Los_Angeles",
				"max_plus_ones": 2
			}`,
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
			wantStatus: http.StatusCreated,
			wantBody:   `"id":1`,
		},
		{
			name: "invalid JSON",
			body: `{invalid json}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid request body",
		},
		{
			name: "missing required field title",
			body: `{
				"start_time": "2026-06-15T14:00:00-07:00",
				"timezone": "America/Los_Angeles"
			}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "title",
		},
		{
			name: "missing required field start_time",
			body: `{
				"title": "Event",
				"timezone": "America/Los_Angeles"
			}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "start_time",
		},
		{
			name: "missing required field timezone",
			body: `{
				"title": "Event",
				"start_time": "2026-06-15T14:00:00-07:00"
			}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "timezone",
		},
		{
			name: "unauthorized user",
			body: `{
				"title": "Event",
				"start_time": "2026-06-15T14:00:00-07:00",
				"timezone": "America/Los_Angeles"
			}`,
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
		{
			name: "service error",
			body: `{
				"title": "Event",
				"start_time": "2026-06-15T14:00:00-07:00",
				"timezone": "America/Los_Angeles"
			}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.CreateEventFunc = func(ctx context.Context, e *models.Event) error {
					return errors.New("database error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "failed to create event",
		},
		{
			name: "validation error",
			body: `{
				"title": "Event",
				"start_time": "2026-06-15T14:00:00-07:00",
				"timezone": "America/Los_Angeles"
			}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.CreateEventFunc = func(ctx context.Context, e *models.Event) error {
					return &models.ValidationError{
						Field:   "start_time",
						Message: "must be in the future",
					}
				}
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "start_time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewEventHandlers(mockService)

			req := httptest.NewRequest("POST", "/api/events", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			ctx := auth.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.CreateEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_GetEvent(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "get existing event",
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
						StartTime: time.Now(),
						Timezone:  "America/Los_Angeles",
						Status:    models.EventStatusDraft,
						CreatedBy: 1,
						Version:   1,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":1`,
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
		{
			name:    "non-existent event",
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
			wantBody:   "event not found",
		},
		{
			name:    "unauthorized access",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewEventHandlers(mockService)

			req := httptest.NewRequest("GET", "/api/events/"+tt.eventID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.GetEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_UpdateEvent(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		body       string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "valid update",
			eventID: "1",
			body: `{
				"title": "Updated Title",
				"version": 1
			}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Original Title",
						StartTime: time.Now().Add(24 * time.Hour),
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
			wantStatus: http.StatusOK,
		},
		{
			name:    "partial update",
			eventID: "1",
			body: `{
				"description": "New description",
				"version": 1
			}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Original Title",
						StartTime: time.Now().Add(24 * time.Hour),
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
			wantStatus: http.StatusOK,
		},
		{
			name:    "version conflict",
			eventID: "1",
			body: `{
				"title": "Updated Title",
				"version": 1
			}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Original Title",
						StartTime: time.Now().Add(24 * time.Hour),
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
			name:    "invalid event ID",
			eventID: "invalid",
			body:    `{"title": "Updated", "version": 1}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid event ID",
		},
		{
			name:    "invalid JSON",
			eventID: "1",
			body:    `{invalid}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid request body",
		},
		{
			name:    "missing version",
			eventID: "1",
			body:    `{"title": "Updated"}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "version",
		},
		{
			name:    "unauthorized update",
			eventID: "1",
			body: `{
				"title": "Updated Title",
				"version": 1
			}`,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewEventHandlers(mockService)

			req := httptest.NewRequest("PUT", "/api/events/"+tt.eventID, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.UpdateEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_DeleteEvent(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "delete own event",
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
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "delete as admin",
			eventID: "1",
			user: &models.User{
				ID:   2,
				Role: models.RoleAdmin,
			},
			setupMock: func(m *mockEventService) {
				m.DeleteEventFunc = func(ctx context.Context, id int64) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "unauthorized delete",
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
			name:    "non-existent event",
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
			wantBody:   "event not found",
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

			handlers := NewEventHandlers(mockService)

			req := httptest.NewRequest("DELETE", "/api/events/"+tt.eventID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.DeleteEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_ListEvents(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:  "list all events",
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
							Title:     "Event 1",
							StartTime: time.Now(),
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
			wantBody:   `"total":1`,
		},
		{
			name:  "list with pagination",
			query: "?limit=10&offset=20",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
					if filters.Limit != 10 || filters.Offset != 20 {
						return nil, fmt.Errorf("unexpected filters")
					}
					return []*models.Event{}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantBody:   `"limit":10`,
		},
		{
			name:  "list with status filter",
			query: "?status=published",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
					if filters.Status == nil || *filters.Status != models.EventStatusPublished {
						return nil, fmt.Errorf("unexpected status filter")
					}
					return []*models.Event{}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "invalid limit parameter",
			query: "?limit=invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid limit",
		},
		{
			name:  "invalid offset parameter",
			query: "?offset=invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid offset",
		},
		{
			name:  "invalid status parameter",
			query: "?status=invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid status",
		},
		{
			name:  "list with creator_id filter",
			query: "?creator_id=1",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
					if filters.CreatorID == nil || *filters.CreatorID != 1 {
						return nil, fmt.Errorf("unexpected creator_id filter")
					}
					return []*models.Event{}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "invalid creator_id parameter",
			query: "?creator_id=invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid creator_id",
		},
		{
			name:  "negative creator_id parameter",
			query: "?creator_id=-1",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
					if filters.CreatorID == nil || *filters.CreatorID != -1 {
						return nil, fmt.Errorf("unexpected creator_id filter")
					}
					return []*models.Event{}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewEventHandlers(mockService)

			req := httptest.NewRequest("GET", "/api/events"+tt.query, nil)

			ctx := auth.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.ListEvents(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_PublishEvent(t *testing.T) {
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
			wantStatus: http.StatusOK,
		},
		{
			name:    "invalid state transition",
			eventID: "1",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.PublishEventFunc = func(ctx context.Context, id int64) error {
					return errors.New("invalid state transition")
				}
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid state transition",
		},
		{
			name:    "unauthorized",
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
			name:    "invalid event ID",
			eventID: "invalid",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid event ID",
		},
		{
			name:    "non-existent event",
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
			wantBody:   "event not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handlers := NewEventHandlers(mockService)

			req := httptest.NewRequest("POST", "/api/events/"+tt.eventID+"/publish", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.PublishEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_CancelEvent(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		body       string
		user       *models.User
		setupMock  func(*mockEventService)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "successful cancel with reason",
			eventID: "1",
			body:    `{"reason": "Event location is no longer available"}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.CancelEventFunc = func(ctx context.Context, id int64, reason string) error {
					if reason != "Event location is no longer available" {
						return fmt.Errorf("unexpected reason")
					}
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "missing reason",
			eventID: "1",
			body:    `{}`,
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
			body:    `{"reason": "short"}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "reason",
		},
		{
			name:    "invalid state transition",
			eventID: "1",
			body:    `{"reason": "Event location is no longer available"}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			setupMock: func(m *mockEventService) {
				m.CancelEventFunc = func(ctx context.Context, id int64, reason string) error {
					return errors.New("invalid state transition")
				}
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid state transition",
		},
		{
			name:    "unauthorized",
			eventID: "1",
			body:    `{"reason": "Event location is no longer available"}`,
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
			name:    "invalid JSON",
			eventID: "1",
			body:    `{invalid}`,
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid request body",
		},
		{
			name:    "invalid event ID",
			eventID: "invalid",
			body:    `{"reason": "Event location is no longer available"}`,
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

			handlers := NewEventHandlers(mockService)

			req := httptest.NewRequest("POST", "/api/events/"+tt.eventID+"/cancel", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), tt.user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.CancelEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_RegisterRoutes(t *testing.T) {
	mockService := &mockEventService{}
	mockService.GetEventFunc = func(ctx context.Context, id int64) (*models.Event, error) {
		return &models.Event{
			ID:        id,
			Title:     "Test Event",
			StartTime: time.Now(),
			Timezone:  "America/Los_Angeles",
			Status:    models.EventStatusDraft,
			CreatedBy: 1,
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}
	mockService.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
		return []*models.Event{}, nil
	}

	handlers := NewEventHandlers(mockService)

	r := chi.NewRouter()
	handlers.RegisterRoutes(r)

	user := &models.User{
		ID:   1,
		Role: models.RoleEventManager,
	}

	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/events"},
		{"GET", "/api/events"},
		{"GET", "/api/events/1"},
		{"PUT", "/api/events/1"},
		{"DELETE", "/api/events/1"},
		{"POST", "/api/events/1/publish"},
		{"POST", "/api/events/1/cancel"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			ctx := auth.WithUser(req.Context(), user)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code == http.StatusMethodNotAllowed {
				t.Errorf("Route %s %s not registered", route.method, route.path)
			}
		})
	}
}

func TestEventHandlers_CreateEvent_TitleBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "title exactly 2 characters",
			title:      "AB",
			wantStatus: http.StatusBadRequest,
			wantBody:   "title must be between 3 and 200 characters",
		},
		{
			name:       "title exactly 3 characters",
			title:      "ABC",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "title exactly 200 characters",
			title:      strings.Repeat("A", 200),
			wantStatus: http.StatusCreated,
		},
		{
			name:       "title exactly 201 characters",
			title:      strings.Repeat("A", 201),
			wantStatus: http.StatusBadRequest,
			wantBody:   "title must be between 3 and 200 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			mockService.CreateEventFunc = func(ctx context.Context, e *models.Event) error {
				e.ID = 1
				e.CreatedAt = time.Now()
				e.UpdatedAt = time.Now()
				return nil
			}

			handlers := NewEventHandlers(mockService)

			body := fmt.Sprintf(`{
				"title": "%s",
				"start_time": "2026-06-15T14:00:00-07:00",
				"timezone": "America/Los_Angeles",
				"max_plus_ones": 0
			}`, tt.title)

			req := httptest.NewRequest("POST", "/api/events", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			ctx := auth.WithUser(req.Context(), &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			})
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.CreateEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_CreateEvent_MaxPlusOnesBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		maxPlusOnes int
		wantStatus  int
		wantBody    string
	}{
		{
			name:        "max_plus_ones exactly -1",
			maxPlusOnes: -1,
			wantStatus:  http.StatusBadRequest,
			wantBody:    "max_plus_ones must be between 0 and 10",
		},
		{
			name:        "max_plus_ones exactly 0",
			maxPlusOnes: 0,
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "max_plus_ones exactly 10",
			maxPlusOnes: 10,
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "max_plus_ones exactly 11",
			maxPlusOnes: 11,
			wantStatus:  http.StatusBadRequest,
			wantBody:    "max_plus_ones must be between 0 and 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			mockService.CreateEventFunc = func(ctx context.Context, e *models.Event) error {
				e.ID = 1
				e.CreatedAt = time.Now()
				e.UpdatedAt = time.Now()
				return nil
			}

			handlers := NewEventHandlers(mockService)

			body := fmt.Sprintf(`{
				"title": "Test Event",
				"start_time": "2026-06-15T14:00:00-07:00",
				"timezone": "America/Los_Angeles",
				"max_plus_ones": %d
			}`, tt.maxPlusOnes)

			req := httptest.NewRequest("POST", "/api/events", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			ctx := auth.WithUser(req.Context(), &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			})
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.CreateEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_ListEvents_LimitOffsetBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "limit exactly 0",
			query:      "?limit=0",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid limit",
		},
		{
			name:       "limit exactly 1",
			query:      "?limit=1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "limit exactly 100",
			query:      "?limit=100",
			wantStatus: http.StatusOK,
		},
		{
			name:       "limit exactly 101",
			query:      "?limit=101",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid limit",
		},
		{
			name:       "limit exactly -1",
			query:      "?limit=-1",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid limit",
		},
		{
			name:       "offset exactly -1",
			query:      "?offset=-1",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid offset",
		},
		{
			name:       "offset exactly 0",
			query:      "?offset=0",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			mockService.ListEventsFunc = func(ctx context.Context, filters events.ListFilters) ([]*models.Event, error) {
				return []*models.Event{}, nil
			}

			handlers := NewEventHandlers(mockService)

			req := httptest.NewRequest("GET", "/api/events"+tt.query, nil)

			ctx := auth.WithUser(req.Context(), &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			})
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.ListEvents(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestEventHandlers_CancelEvent_ReasonBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		reason     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "reason exactly 9 characters",
			reason:     "123456789",
			wantStatus: http.StatusBadRequest,
			wantBody:   "reason must be between 10 and 500 characters",
		},
		{
			name:       "reason exactly 10 characters",
			reason:     "1234567890",
			wantStatus: http.StatusOK,
		},
		{
			name:       "reason exactly 500 characters",
			reason:     strings.Repeat("A", 500),
			wantStatus: http.StatusOK,
		},
		{
			name:       "reason exactly 501 characters",
			reason:     strings.Repeat("A", 501),
			wantStatus: http.StatusBadRequest,
			wantBody:   "reason must be between 10 and 500 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockEventService{}
			mockService.CancelEventFunc = func(ctx context.Context, id int64, reason string) error {
				return nil
			}

			handlers := NewEventHandlers(mockService)

			body := fmt.Sprintf(`{"reason": "%s"}`, tt.reason)

			req := httptest.NewRequest("POST", "/api/events/1/cancel", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			ctx := auth.WithUser(req.Context(), &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			})
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handlers.CancelEvent(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}
