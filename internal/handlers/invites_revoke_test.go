package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockRevokeInviteService struct {
	revokeInviteFunc func(ctx context.Context, req *invites.RevokeInviteRequest) error
	getInviteByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
}

func (m *mockRevokeInviteService) RevokeInvite(ctx context.Context, req *invites.RevokeInviteRequest) error {
	if m.revokeInviteFunc != nil {
		return m.revokeInviteFunc(ctx, req)
	}
	return nil
}

func (m *mockRevokeInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getInviteByIDFunc != nil {
		return m.getInviteByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "invite"}
}

type mockRevokeEventRepository struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockRevokeEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "event"}
}

func (m *mockRevokeEventRepository) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRevokeEventRepository) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRevokeEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockRevokeEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockRevokeEventRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRevokeEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockRevokeEventRepository) GetByCreator(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockRevokeEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockRevokeEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func TestRevokeInviteHandlers_RevokeInvite(t *testing.T) {
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	email := "test@example.com"
	reason := "Wrong email address"

	tests := []struct {
		name           string
		inviteID       string
		requestBody    interface{}
		user           *models.User
		mockService    *mockRevokeInviteService
		mockEventRepo  *mockRevokeEventRepository
		wantStatus     int
		wantErrMessage string
	}{
		{
			name:     "successful revocation by event creator",
			inviteID: "1",
			requestBody: map[string]interface{}{
				"reason": reason,
			},
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRevokeInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("a", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   futureTime,
					}, nil
				},
				revokeInviteFunc: func(ctx context.Context, req *invites.RevokeInviteRequest) error {
					if req.InviteID != 1 {
						t.Errorf("InviteID = %d, want 1", req.InviteID)
					}
					if req.Reason == nil || *req.Reason != reason {
						t.Errorf("Reason = %v, want %s", req.Reason, reason)
					}
					return nil
				},
			},
			mockEventRepo: &mockRevokeEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
						StartTime: futureTime,
					}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "successful revocation by admin",
			inviteID: "1",
			requestBody: map[string]interface{}{
				"reason": reason,
			},
			user: &models.User{
				ID:    2,
				Email: "admin@example.com",
				Role:  models.RoleAdmin,
			},
			mockService: &mockRevokeInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("a", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   futureTime,
					}, nil
				},
				revokeInviteFunc: func(ctx context.Context, req *invites.RevokeInviteRequest) error {
					return nil
				},
			},
			mockEventRepo: &mockRevokeEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
						StartTime: futureTime,
					}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "successful revocation without reason",
			inviteID: "1",
			requestBody: map[string]interface{}{},
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRevokeInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("a", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   futureTime,
					}, nil
				},
				revokeInviteFunc: func(ctx context.Context, req *invites.RevokeInviteRequest) error {
					if req.Reason != nil {
						t.Errorf("Reason = %v, want nil", req.Reason)
					}
					return nil
				},
			},
			mockEventRepo: &mockRevokeEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
						StartTime: futureTime,
					}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "missing authentication",
			inviteID:    "1",
			requestBody: map[string]interface{}{},
			user:        nil,
			mockService: &mockRevokeInviteService{},
			mockEventRepo: &mockRevokeEventRepository{},
			wantStatus:     http.StatusUnauthorized,
			wantErrMessage: "authentication required",
		},
		{
			name:     "invalid invite ID",
			inviteID: "invalid",
			requestBody: map[string]interface{}{},
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService:    &mockRevokeInviteService{},
			mockEventRepo:  &mockRevokeEventRepository{},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "invalid invite ID",
		},
		{
			name:     "invite not found",
			inviteID: "999",
			requestBody: map[string]interface{}{},
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRevokeInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return nil, &models.NotFoundError{Resource: "invite"}
				},
			},
			mockEventRepo:  &mockRevokeEventRepository{},
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "invite not found",
		},
		{
			name:     "event not found",
			inviteID: "1",
			requestBody: map[string]interface{}{},
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRevokeInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("a", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   futureTime,
					}, nil
				},
			},
			mockEventRepo: &mockRevokeEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.NotFoundError{Resource: "event"}
				},
			},
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "event not found",
		},
		{
			name:     "permission denied - not event creator or admin",
			inviteID: "1",
			requestBody: map[string]interface{}{},
			user: &models.User{
				ID:    2,
				Email: "other@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRevokeInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("a", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   futureTime,
					}, nil
				},
			},
			mockEventRepo: &mockRevokeEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
						StartTime: futureTime,
					}, nil
				},
			},
			wantStatus:     http.StatusForbidden,
			wantErrMessage: "permission denied",
		},
		{
			name:     "cannot revoke responded invite",
			inviteID: "1",
			requestBody: map[string]interface{}{},
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRevokeInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("a", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusResponded,
						ExpiresAt:   futureTime,
					}, nil
				},
				revokeInviteFunc: func(ctx context.Context, req *invites.RevokeInviteRequest) error {
					return errors.New("cannot transition from responded to revoked")
				},
			},
			mockEventRepo: &mockRevokeEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
						StartTime: futureTime,
					}, nil
				},
			},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "cannot transition from responded to revoked",
		},
		{
			name:     "cannot revoke already revoked invite",
			inviteID: "1",
			requestBody: map[string]interface{}{},
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRevokeInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return &models.Invite{
						ID:          1,
						EventID:     1,
						Email:       &email,
						TokenHash:   strings.Repeat("a", 43),
						MaxPlusOnes: 2,
						Status:      models.InviteStatusRevoked,
						ExpiresAt:   futureTime,
					}, nil
				},
				revokeInviteFunc: func(ctx context.Context, req *invites.RevokeInviteRequest) error {
					return errors.New("cannot transition from revoked to revoked")
				},
			},
			mockEventRepo: &mockRevokeEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
						StartTime: futureTime,
					}, nil
				},
			},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "cannot transition from revoked to revoked",
		},
		{
			name:     "invalid JSON body",
			inviteID: "1",
			requestBody: "invalid json",
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService:    &mockRevokeInviteService{},
			mockEventRepo:  &mockRevokeEventRepository{},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/invites/"+tt.inviteID+"/revoke", bytes.NewReader(body))
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			if tt.user != nil {
				ctx := auth.WithUser(req.Context(), tt.user)
				req = req.WithContext(ctx)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("inviteId", tt.inviteID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler := NewRevokeInviteHandlers(tt.mockService, tt.mockEventRepo)
			handler.RevokeInvite(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("RevokeInvite() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErrMessage != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if errMsg, ok := response["message"].(string); !ok || !strings.Contains(errMsg, tt.wantErrMessage) {
					t.Errorf("RevokeInvite() error = %v, want error containing %q", errMsg, tt.wantErrMessage)
				}
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if msg, ok := response["message"].(string); !ok || msg == "" {
					t.Error("RevokeInvite() expected success message")
				}
			}
		})
	}
}
