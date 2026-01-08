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

type mockManualInviteService struct {
	createManualInviteFunc func(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error)
}

func (m *mockManualInviteService) CreateManualInvite(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error) {
	if m.createManualInviteFunc != nil {
		return m.createManualInviteFunc(ctx, req, expiresAt)
	}
	return nil, errors.New("not implemented")
}

func (m *mockManualInviteService) CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error) {
	return nil, "", errors.New("not implemented")
}

func (m *mockManualInviteService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	return nil, errors.New("not implemented")
}

func (m *mockManualInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	return nil, errors.New("not implemented")
}

func (m *mockManualInviteService) RevokeInvite(ctx context.Context, req *invites.RevokeInviteRequest) error {
	return errors.New("not implemented")
}

func (m *mockManualInviteService) RegenerateToken(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockManualInviteService) ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, errors.New("not implemented")
}

func (m *mockManualInviteService) ImportCSV(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockManualInviteService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return 0, errors.New("not implemented")
}

func (m *mockManualInviteService) MarkInviteSent(ctx context.Context, inviteID int64) error {
	return errors.New("not implemented")
}

func (m *mockManualInviteService) MarkInviteViewed(ctx context.Context, inviteID int64) error {
	return errors.New("not implemented")
}

func (m *mockManualInviteService) MarkInviteResponded(ctx context.Context, inviteID int64) error {
	return errors.New("not implemented")
}

func TestManualInviteHandlers_CreateManualInvite(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		requestBody    interface{}
		user           *models.User
		mockService    *mockManualInviteService
		mockEventRepo  *mockEventRepository
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name:    "successful manual invite creation",
			eventID: "1",
			requestBody: map[string]interface{}{
				"name":          "John Doe",
				"max_plus_ones": 2,
			},
			user: &models.User{
				ID:    1,
				Email: "admin@example.com",
				Role:  models.RoleAdmin,
			},
			mockEventRepo: &mockEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:          1,
						Title:       "Test Event",
						CreatedBy:   1,
						Status:      models.EventStatusDraft,
						StartTime:   time.Now().Add(30 * 24 * time.Hour),
						MaxPlusOnes: 2,
					}, nil
				},
			},
			mockService: &mockManualInviteService{
				createManualInviteFunc: func(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error) {
					name := "John Doe"
					return &invites.CreateManualInviteResponse{
						Invite: &models.Invite{
							ID:          1,
							EventID:     1,
							Name:        &name,
							Email:       nil,
							MaxPlusOnes: 2,
							Status:      models.InviteStatusDraft,
							ExpiresAt:   expiresAt,
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						},
						Token:   strings.Repeat("a", 43),
						RSVPURL: "/rsvp/" + strings.Repeat("a", 43),
					}, nil
				},
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				if body["token"] == nil || body["token"].(string) == "" {
					t.Error("expected token in response")
				}
				if body["rsvp_url"] == nil || body["rsvp_url"].(string) == "" {
					t.Error("expected rsvp_url in response")
				}
				invite := body["invite"].(map[string]interface{})
				if invite["email"] != nil {
					t.Error("expected nil email for manual invite")
				}
				if invite["status"] != string(models.InviteStatusDraft) {
					t.Errorf("expected status %s, got %v", models.InviteStatusDraft, invite["status"])
				}
			},
		},
		{
			name:    "successful manual invite without name",
			eventID: "1",
			requestBody: map[string]interface{}{
				"max_plus_ones": 0,
			},
			user: &models.User{
				ID:    1,
				Email: "admin@example.com",
				Role:  models.RoleAdmin,
			},
			mockEventRepo: &mockEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:          1,
						Title:       "Test Event",
						CreatedBy:   1,
						Status:      models.EventStatusDraft,
						StartTime:   time.Now().Add(30 * 24 * time.Hour),
						MaxPlusOnes: 2,
					}, nil
				},
			},
			mockService: &mockManualInviteService{
				createManualInviteFunc: func(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error) {
					return &invites.CreateManualInviteResponse{
						Invite: &models.Invite{
							ID:          2,
							EventID:     1,
							Name:        nil,
							Email:       nil,
							MaxPlusOnes: 0,
							Status:      models.InviteStatusDraft,
							ExpiresAt:   expiresAt,
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						},
						Token:   strings.Repeat("b", 43),
						RSVPURL: "/rsvp/" + strings.Repeat("b", 43),
					}, nil
				},
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				invite := body["invite"].(map[string]interface{})
				if invite["name"] != nil {
					t.Errorf("expected nil name, got %v", invite["name"])
				}
			},
		},
		{
			name:    "unauthorized - no user",
			eventID: "1",
			requestBody: map[string]interface{}{
				"name": "Test",
			},
			user:           nil,
			mockEventRepo:  &mockEventRepository{},
			mockService:    &mockManualInviteService{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:    "invalid event ID",
			eventID: "invalid",
			requestBody: map[string]interface{}{
				"name": "Test",
			},
			user: &models.User{
				ID:    1,
				Email: "admin@example.com",
				Role:  models.RoleAdmin,
			},
			mockEventRepo:  &mockEventRepository{},
			mockService:    &mockManualInviteService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "event not found",
			eventID: "999",
			requestBody: map[string]interface{}{
				"name": "Test",
			},
			user: &models.User{
				ID:    1,
				Email: "admin@example.com",
				Role:  models.RoleAdmin,
			},
			mockEventRepo: &mockEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return nil, &models.NotFoundError{Resource: "event"}
				},
			},
			mockService:    &mockManualInviteService{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "permission denied - not creator or admin",
			eventID: "1",
			requestBody: map[string]interface{}{
				"name": "Test",
			},
			user: &models.User{
				ID:    2,
				Email: "user@example.com",
				Role:  models.RoleEventManager,
			},
			mockEventRepo: &mockEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
						StartTime: time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			mockService:    &mockManualInviteService{},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:    "cannot create invite for cancelled event",
			eventID: "1",
			requestBody: map[string]interface{}{
				"name": "Test",
			},
			user: &models.User{
				ID:    1,
				Email: "admin@example.com",
				Role:  models.RoleAdmin,
			},
			mockEventRepo: &mockEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusCancelled,
						StartTime: time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			mockService:    &mockManualInviteService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid request body",
			eventID:     "1",
			requestBody: "invalid json",
			user:        &models.User{ID: 1, Role: models.RoleAdmin},
			mockEventRepo: &mockEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
						StartTime: time.Now().Add(30 * 24 * time.Hour),
					}, nil
				},
			},
			mockService:    &mockManualInviteService{},
			expectedStatus: http.StatusBadRequest,
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

			req := httptest.NewRequest(http.MethodPost, "/api/events/"+tt.eventID+"/invites/manual", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.user != nil {
				ctx := auth.WithUser(req.Context(), tt.user)
				req = req.WithContext(ctx)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("eventId", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handlers := NewManualInviteHandlers(tt.mockService, tt.mockEventRepo, "https://example.com")
			handlers.CreateManualInvite(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.validateBody != nil && rr.Code == http.StatusCreated {
				var responseBody map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				tt.validateBody(t, responseBody)
			}
		})
	}
}
