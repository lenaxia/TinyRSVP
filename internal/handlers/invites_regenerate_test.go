package handlers

import (
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

type mockRegenerateInviteService struct {
	regenerateTokenFunc func(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error)
	getInviteByIDFunc   func(ctx context.Context, id int64) (*models.Invite, error)
}

func (m *mockRegenerateInviteService) CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error) {
	return nil, "", nil
}

func (m *mockRegenerateInviteService) CreateManualInvite(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error) {
	return nil, nil
}

func (m *mockRegenerateInviteService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	return nil, nil
}

func (m *mockRegenerateInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	if m.getInviteByIDFunc != nil {
		return m.getInviteByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "invite"}
}

func (m *mockRegenerateInviteService) RevokeInvite(ctx context.Context, req *invites.RevokeInviteRequest) error {
	return nil
}

func (m *mockRegenerateInviteService) RegenerateToken(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
	if m.regenerateTokenFunc != nil {
		return m.regenerateTokenFunc(ctx, inviteID)
	}
	return nil, nil
}

func (m *mockRegenerateInviteService) ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockRegenerateInviteService) ImportCSV(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
	return nil, nil
}

func (m *mockRegenerateInviteService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockRegenerateInviteService) MarkInviteSent(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockRegenerateInviteService) MarkInviteViewed(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockRegenerateInviteService) MarkInviteResponded(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockRegenerateInviteService) UnsubscribeFromReminders(ctx context.Context, token string) error {
	return nil
}

func (m *mockRegenerateInviteService) ListInvites(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
	return &invites.ListInvitesResponse{
		Invites: []*models.Invite{},
		Total:   0,
		Stats:   &repositories.InviteStats{},
	}, nil
}

func (m *mockRegenerateInviteService) UpdateInvite(ctx context.Context, req *invites.UpdateInviteRequest) error {
	return nil
}

func (m *mockRegenerateInviteService) DeleteInvite(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockRegenerateInviteService) SendInvite(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error {
	return nil
}

type mockRegenerateEventRepository struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockRegenerateEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "event"}
}

func (m *mockRegenerateEventRepository) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRegenerateEventRepository) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockRegenerateEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockRegenerateEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockRegenerateEventRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockRegenerateEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockRegenerateEventRepository) ListWithStats(ctx context.Context, filters repositories.ListFilters) ([]*models.EventWithStats, error) {
	return nil, nil
}

func (m *mockRegenerateEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockRegenerateEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockRegenerateEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func TestRegenerateInviteTokenHandlers_RegenerateInviteToken(t *testing.T) {
	futureTime := time.Now().Add(30 * 24 * time.Hour)
	email := "test@example.com"

	tests := []struct {
		name           string
		inviteID       string
		user           *models.User
		mockService    *mockRegenerateInviteService
		mockEventRepo  *mockRegenerateEventRepository
		wantStatus     int
		wantErrMessage string
	}{
		{
			name:     "successful regeneration by event creator",
			inviteID: "1",
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRegenerateInviteService{
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
				regenerateTokenFunc: func(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
					if inviteID != 1 {
						t.Errorf("inviteID = %d, want 1", inviteID)
					}
					return &invites.RegenerateTokenResponse{
						Token:   "new-token-abc123",
						RSVPURL: "/rsvp/new-token-abc123",
					}, nil
				},
			},
			mockEventRepo: &mockRegenerateEventRepository{
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
			name:     "successful regeneration by admin",
			inviteID: "1",
			user: &models.User{
				ID:    2,
				Email: "admin@example.com",
				Role:  models.RoleAdmin,
			},
			mockService: &mockRegenerateInviteService{
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
				regenerateTokenFunc: func(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
					return &invites.RegenerateTokenResponse{
						Token:   "new-token-abc123",
						RSVPURL: "/rsvp/new-token-abc123",
					}, nil
				},
			},
			mockEventRepo: &mockRegenerateEventRepository{
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
			name:           "missing authentication",
			inviteID:       "1",
			user:           nil,
			mockService:    &mockRegenerateInviteService{},
			mockEventRepo:  &mockRegenerateEventRepository{},
			wantStatus:     http.StatusUnauthorized,
			wantErrMessage: "authentication required",
		},
		{
			name:     "invalid invite ID",
			inviteID: "invalid",
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService:    &mockRegenerateInviteService{},
			mockEventRepo:  &mockRegenerateEventRepository{},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "invalid invite ID",
		},
		{
			name:     "invite not found",
			inviteID: "999",
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRegenerateInviteService{
				getInviteByIDFunc: func(ctx context.Context, id int64) (*models.Invite, error) {
					return nil, &models.NotFoundError{Resource: "invite"}
				},
			},
			mockEventRepo:  &mockRegenerateEventRepository{},
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "invite not found",
		},
		{
			name:     "event not found",
			inviteID: "1",
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRegenerateInviteService{
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
			mockEventRepo: &mockRegenerateEventRepository{
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
			user: &models.User{
				ID:    2,
				Email: "other@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRegenerateInviteService{
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
			mockEventRepo: &mockRegenerateEventRepository{
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
			name:     "cannot regenerate revoked invite",
			inviteID: "1",
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRegenerateInviteService{
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
				regenerateTokenFunc: func(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
					return nil, errors.New("cannot regenerate token for revoked invite")
				},
			},
			mockEventRepo: &mockRegenerateEventRepository{
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
			wantErrMessage: "cannot regenerate token for revoked invite",
		},
		{
			name:     "cannot regenerate responded invite",
			inviteID: "1",
			user: &models.User{
				ID:    1,
				Email: "creator@example.com",
				Role:  models.RoleEventManager,
			},
			mockService: &mockRegenerateInviteService{
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
				regenerateTokenFunc: func(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
					return nil, errors.New("cannot regenerate token for responded invite")
				},
			},
			mockEventRepo: &mockRegenerateEventRepository{
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
			wantErrMessage: "cannot regenerate token for responded invite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/invites/"+tt.inviteID+"/regenerate", nil)
			req.Header.Set("Accept", "application/json")

			if tt.user != nil {
				ctx := auth.WithUser(req.Context(), tt.user)
				req = req.WithContext(ctx)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("inviteId", tt.inviteID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler := NewRegenerateInviteTokenHandlers(tt.mockService, tt.mockEventRepo)
			handler.RegenerateInviteToken(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("RegenerateInviteToken() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantErrMessage != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if errMsg, ok := response["message"].(string); !ok || !strings.Contains(errMsg, tt.wantErrMessage) {
					t.Errorf("RegenerateInviteToken() error = %v, want error containing %q", errMsg, tt.wantErrMessage)
				}
			}

			if tt.wantStatus == http.StatusOK {
				var response invites.RegenerateTokenResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if response.Token == "" {
					t.Error("RegenerateInviteToken() expected non-empty token")
				}
				if response.RSVPURL == "" {
					t.Error("RegenerateInviteToken() expected non-empty RSVP URL")
				}
			}
		})
	}
}

func (m *mockRegenerateEventRepository) CountEvents(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockRegenerateEventRepository) GetComponentOverrides(ctx context.Context, eventID int64) (*models.ComponentOverrides, error) {
	return nil, nil
}

func (m *mockRegenerateEventRepository) UpdateComponentOverrides(ctx context.Context, eventID int64, overrides *models.ComponentOverrides) error {
	return nil
}

func (m *mockRegenerateEventRepository) DeleteComponentOverrides(ctx context.Context, eventID int64) error {
	return nil
}
func (m *mockRegenerateEventRepository) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRegenerateEventRepository) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	return nil, errors.New("not implemented")
}
