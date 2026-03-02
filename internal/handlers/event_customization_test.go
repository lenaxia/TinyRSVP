package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestEventCustomizationHandlers_GetCustomization(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		user           *models.User
		setupMock      func(*mockCustomizationService)
		wantStatus     int
		wantErrMessage string
	}{
		{
			name:    "successful get customization",
			eventID: "1",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			setupMock: func(m *mockCustomizationService) {
				m.GetEventCustomizationFunc = func(ctx context.Context, eventID int64) (*events.EventCustomizationData, error) {
					templateID := int64(1)
					return &events.EventCustomizationData{
						Event: &models.Event{
							ID:         1,
							CreatedBy:  1,
							TemplateID: &templateID,
						},
						Template: &models.Template{
							ID: 1,
						},
						TemplateConfig: &models.ComponentConfiguration{
							Version:    "1.0",
							Components: []models.Component{},
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid event ID",
			eventID:    "invalid",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			setupMock:  func(m *mockCustomizationService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "event not found",
			eventID: "999",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			setupMock: func(m *mockCustomizationService) {
				m.GetEventCustomizationFunc = func(ctx context.Context, eventID int64) (*events.EventCustomizationData, error) {
					return nil, &models.NotFoundError{Resource: "Event", ID: eventID}
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "permission denied",
			eventID: "1",
			user:    &models.User{ID: 2, Role: models.RoleEventManager},
			setupMock: func(m *mockCustomizationService) {
				m.GetEventCustomizationFunc = func(ctx context.Context, eventID int64) (*events.EventCustomizationData, error) {
					return nil, &models.PermissionDeniedError{
						Action:   "get event customization",
						Resource: "Event",
						ID:       eventID,
					}
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockCustomizationService{}
			tt.setupMock(mockService)

			handler := NewEventCustomizationHandlers(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/events/"+tt.eventID+"/template/customization", nil)
			req.Header.Set("Accept", "application/json")

			if tt.user != nil {
				req = req.WithContext(auth.WithUser(req.Context(), tt.user))
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.GetCustomization(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetCustomization() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestEventCustomizationHandlers_UpdateCustomization(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		body       interface{}
		setupMock  func(*mockCustomizationService)
		wantStatus int
	}{
		{
			name:    "successful update",
			eventID: "1",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			body: &models.ComponentOverrides{
				Version: "1.0",
				Overrides: []models.ComponentOverride{
					{
						ID: "title-text",
						Updates: map[string]interface{}{
							"content": map[string]interface{}{
								"color": "#ff0000",
							},
						},
					},
				},
			},
			setupMock: func(m *mockCustomizationService) {
				m.UpdateEventCustomizationFunc = func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid event ID",
			eventID:    "invalid",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			body:       &models.ComponentOverrides{Version: "1.0"},
			setupMock:  func(m *mockCustomizationService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON body",
			eventID:    "1",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			body:       "invalid json",
			setupMock:  func(m *mockCustomizationService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "permission denied",
			eventID: "1",
			user:    &models.User{ID: 2, Role: models.RoleEventManager},
			body:    &models.ComponentOverrides{Version: "1.0"},
			setupMock: func(m *mockCustomizationService) {
				m.UpdateEventCustomizationFunc = func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
					return &models.PermissionDeniedError{
						Action:   "update event customization",
						Resource: "Event",
						ID:       eventID,
					}
				}
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "validation error",
			eventID: "1",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			body:    &models.ComponentOverrides{Version: ""},
			setupMock: func(m *mockCustomizationService) {
				m.UpdateEventCustomizationFunc = func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
					return &models.ValidationError{
						Field:   "version",
						Message: "version is required",
					}
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockCustomizationService{}
			tt.setupMock(mockService)

			handler := NewEventCustomizationHandlers(mockService)

			var bodyBytes []byte
			if str, ok := tt.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/events/"+tt.eventID+"/template/customization", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			if tt.user != nil {
				req = req.WithContext(auth.WithUser(req.Context(), tt.user))
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.UpdateCustomization(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("UpdateCustomization() status = %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestEventCustomizationHandlers_PreviewCustomization(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		body       interface{}
		setupMock  func(*mockCustomizationService)
		wantStatus int
	}{
		{
			name:    "successful preview",
			eventID: "1",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			body: &models.ComponentOverrides{
				Version: "1.0",
				Overrides: []models.ComponentOverride{
					{
						ID: "title-text",
						Updates: map[string]interface{}{
							"content": map[string]interface{}{
								"color": "#0000ff",
							},
						},
					},
				},
			},
			setupMock: func(m *mockCustomizationService) {
				m.PreviewEventCustomizationFunc = func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) (*models.ComponentConfiguration, error) {
					return &models.ComponentConfiguration{
						Version:    "1.0",
						Components: []models.Component{},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid event ID",
			eventID:    "invalid",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			body:       &models.ComponentOverrides{Version: "1.0"},
			setupMock:  func(m *mockCustomizationService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "permission denied",
			eventID: "1",
			user:    &models.User{ID: 2, Role: models.RoleGuest},
			body:    &models.ComponentOverrides{Version: "1.0"},
			setupMock: func(m *mockCustomizationService) {
				m.PreviewEventCustomizationFunc = func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) (*models.ComponentConfiguration, error) {
					return nil, &models.PermissionDeniedError{
						Action:   "preview event customization",
						Resource: "Event",
						ID:       eventID,
					}
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockCustomizationService{}
			tt.setupMock(mockService)

			handler := NewEventCustomizationHandlers(mockService)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/events/"+tt.eventID+"/template/customization/preview", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			if tt.user != nil {
				req = req.WithContext(auth.WithUser(req.Context(), tt.user))
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.PreviewCustomization(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("PreviewCustomization() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestEventCustomizationHandlers_ResetCustomization(t *testing.T) {
	tests := []struct {
		name       string
		eventID    string
		user       *models.User
		setupMock  func(*mockCustomizationService)
		wantStatus int
	}{
		{
			name:    "successful reset",
			eventID: "1",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			setupMock: func(m *mockCustomizationService) {
				m.ResetEventCustomizationFunc = func(ctx context.Context, eventID int64) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid event ID",
			eventID:    "invalid",
			user:       &models.User{ID: 1, Role: models.RoleEventManager},
			setupMock:  func(m *mockCustomizationService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "event not found",
			eventID: "999",
			user:    &models.User{ID: 1, Role: models.RoleEventManager},
			setupMock: func(m *mockCustomizationService) {
				m.ResetEventCustomizationFunc = func(ctx context.Context, eventID int64) error {
					return &models.NotFoundError{Resource: "Event", ID: eventID}
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "permission denied",
			eventID: "1",
			user:    &models.User{ID: 2, Role: models.RoleEventManager},
			setupMock: func(m *mockCustomizationService) {
				m.ResetEventCustomizationFunc = func(ctx context.Context, eventID int64) error {
					return &models.PermissionDeniedError{
						Action:   "reset event customization",
						Resource: "Event",
						ID:       eventID,
					}
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockCustomizationService{}
			tt.setupMock(mockService)

			handler := NewEventCustomizationHandlers(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/events/"+tt.eventID+"/template/customization", nil)
			req.Header.Set("Accept", "application/json")

			if tt.user != nil {
				req = req.WithContext(auth.WithUser(req.Context(), tt.user))
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.ResetCustomization(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ResetCustomization() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

type mockCustomizationService struct {
	GetEventCustomizationFunc      func(ctx context.Context, eventID int64) (*events.EventCustomizationData, error)
	UpdateEventCustomizationFunc   func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error
	PreviewEventCustomizationFunc  func(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) (*models.ComponentConfiguration, error)
	ResetEventCustomizationFunc    func(ctx context.Context, eventID int64) error
	ValidateEventCustomizationFunc func(overrides *models.ComponentOverrides) error
}

func (m *mockCustomizationService) GetEventCustomization(ctx context.Context, eventID int64) (*events.EventCustomizationData, error) {
	if m.GetEventCustomizationFunc != nil {
		return m.GetEventCustomizationFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *mockCustomizationService) UpdateEventCustomization(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
	if m.UpdateEventCustomizationFunc != nil {
		return m.UpdateEventCustomizationFunc(ctx, eventID, overrides)
	}
	return nil
}

func (m *mockCustomizationService) PreviewEventCustomization(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) (*models.ComponentConfiguration, error) {
	if m.PreviewEventCustomizationFunc != nil {
		return m.PreviewEventCustomizationFunc(ctx, eventID, overrides)
	}
	return nil, nil
}

func (m *mockCustomizationService) ResetEventCustomization(ctx context.Context, eventID int64) error {
	if m.ResetEventCustomizationFunc != nil {
		return m.ResetEventCustomizationFunc(ctx, eventID)
	}
	return nil
}

func (m *mockCustomizationService) ValidateEventCustomization(overrides *models.ComponentOverrides) error {
	if m.ValidateEventCustomizationFunc != nil {
		return m.ValidateEventCustomizationFunc(overrides)
	}
	return nil
}
