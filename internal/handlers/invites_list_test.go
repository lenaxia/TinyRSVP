package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockListInviteService struct {
	listInvitesFunc func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error)
}

func (m *mockListInviteService) CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error) {
	return nil, "", nil
}

func (m *mockListInviteService) CreateManualInvite(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error) {
	return nil, nil
}

func (m *mockListInviteService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	return nil, nil
}

func (m *mockListInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	return nil, nil
}

func (m *mockListInviteService) RevokeInvite(ctx context.Context, req *invites.RevokeInviteRequest) error {
	return nil
}

func (m *mockListInviteService) RegenerateToken(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
	return nil, nil
}

func (m *mockListInviteService) ListInvites(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
	if m.listInvitesFunc != nil {
		return m.listInvitesFunc(ctx, req)
	}
	return &invites.ListInvitesResponse{
		Invites: []*models.Invite{},
		Total:   0,
		Stats:   &repositories.InviteStats{},
	}, nil
}

func (m *mockListInviteService) ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockListInviteService) ImportCSV(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
	return nil, nil
}

func (m *mockListInviteService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockListInviteService) MarkInviteSent(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockListInviteService) MarkInviteViewed(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockListInviteService) MarkInviteResponded(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockListInviteService) UnsubscribeFromReminders(ctx context.Context, token string) error {
	return nil
}

func (m *mockListInviteService) UpdateInvite(ctx context.Context, req *invites.UpdateInviteRequest) error {
	return nil
}

func (m *mockListInviteService) DeleteInvite(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockListInviteService) SendInvite(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error {
	return nil
}

type mockListEventRepository struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockListEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "event"}
}

func (m *mockListEventRepository) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockListEventRepository) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockListEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockListEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockListEventRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockListEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockListEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockListEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEnd int) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func (m *mockListEventRepository) GetByCreatorID(ctx context.Context, creatorID int64) ([]*models.Event, error) {
	return []*models.Event{}, nil
}

func TestListInvitesHandler_Success(t *testing.T) {
	mockService := &mockListInviteService{
		listInvitesFunc: func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
			email1 := "user1@example.com"
			email2 := "user2@example.com"
			name1 := "User One"
			name2 := "User Two"

			return &invites.ListInvitesResponse{
				Invites: []*models.Invite{
					{
						ID:          1,
						EventID:     1,
						Email:       &email1,
						Name:        &name1,
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
					{
						ID:          2,
						EventID:     1,
						Email:       &email2,
						Name:        &name2,
						MaxPlusOnes: 1,
						Status:      models.InviteStatusDraft,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
				Total: 2,
				Stats: &repositories.InviteStats{
					Total: 2,
					Draft: 1,
					Sent:  1,
				},
			}, nil
		},
	}

	mockEventRepo := &mockListEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 1,
			}, nil
		},
	}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites?limit=50&offset=0", nil)
			req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response invites.ListInvitesResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Invites) != 2 {
		t.Errorf("expected 2 invites, got %d", len(response.Invites))
	}

	if response.Total != 2 {
		t.Errorf("expected total 2, got %d", response.Total)
	}

	if response.Stats == nil {
		t.Fatal("expected stats, got nil")
	}
}

func TestListInvitesHandler_WithFilters(t *testing.T) {
	mockService := &mockListInviteService{
		listInvitesFunc: func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
			if req.Status != nil && *req.Status != "sent" {
				t.Errorf("expected status filter 'sent', got %s", *req.Status)
			}
			if req.Search != nil && *req.Search != "john" {
				t.Errorf("expected search 'john', got %s", *req.Search)
			}
			if req.SortBy != nil && *req.SortBy != "email" {
				t.Errorf("expected sort_by 'email', got %s", *req.SortBy)
			}
			if req.SortOrder != nil && *req.SortOrder != "asc" {
				t.Errorf("expected sort_order 'asc', got %s", *req.SortOrder)
			}

			return &invites.ListInvitesResponse{
				Invites: []*models.Invite{},
				Total:   0,
				Stats:   &repositories.InviteStats{},
			}, nil
		},
	}

	mockEventRepo := &mockListEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 1,
			}, nil
		},
	}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites?status=sent&search=john&sort_by=email&sort_order=asc&limit=25&offset=0", nil)
			req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListInvitesHandler_Unauthorized(t *testing.T) {
	mockService := &mockListInviteService{}
	mockEventRepo := &mockListEventRepository{}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites", nil)
			req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestListInvitesHandler_InvalidEventID(t *testing.T) {
	mockService := &mockListInviteService{}
	mockEventRepo := &mockListEventRepository{}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/invalid/invites", nil)
			req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestListInvitesHandler_EventNotFound(t *testing.T) {
	mockService := &mockListInviteService{}
	mockEventRepo := &mockListEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return nil, &models.NotFoundError{Resource: "event"}
		},
	}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/999/invites", nil)
			req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestListInvitesHandler_PermissionDenied(t *testing.T) {
	mockService := &mockListInviteService{}
	mockEventRepo := &mockListEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 2,
			}, nil
		},
	}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites", nil)
			req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestListInvitesHandler_InvalidQueryParams(t *testing.T) {
	tests := []struct {
		name       string
		queryParam string
		wantStatus int
	}{
		{
			name:       "invalid limit",
			queryParam: "?limit=invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid offset",
			queryParam: "?offset=invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "limit too large",
			queryParam: "?limit=200",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "negative offset",
			queryParam: "?offset=-1",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockListInviteService{}
			mockEventRepo := &mockListEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						Status:    models.EventStatusPublished,
						CreatedBy: 1,
					}, nil
				},
			}

			handler := NewListInviteHandlers(mockService, mockEventRepo)

			r := chi.NewRouter()
			handler.RegisterRoutes(r)

			req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites"+tt.queryParam, nil)
			req.Header.Set("Accept", "application/json")
			user := &models.User{ID: 1, Role: models.RoleAdmin}
			ctx := auth.WithUser(req.Context(), user)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestListInvitesHandler_DefaultValues(t *testing.T) {
	mockService := &mockListInviteService{
		listInvitesFunc: func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
			if req.Limit != 50 {
				t.Errorf("expected default limit 50, got %d", req.Limit)
			}
			if req.Offset != 0 {
				t.Errorf("expected default offset 0, got %d", req.Offset)
			}

			return &invites.ListInvitesResponse{
				Invites: []*models.Invite{},
				Total:   0,
				Stats:   &repositories.InviteStats{},
			}, nil
		},
	}

	mockEventRepo := &mockListEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 1,
			}, nil
		},
	}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites", nil)
			req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListInvitesHandler_ServiceValidationError(t *testing.T) {
	mockService := &mockListInviteService{
		listInvitesFunc: func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
			return nil, &models.ValidationError{
				Field:   "status",
				Message: "invalid status value",
			}
		},
	}

	mockEventRepo := &mockListEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 1,
			}, nil
		},
	}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites?status=invalid", nil)
			req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListInvitesHandler_ServiceInternalError(t *testing.T) {
	mockService := &mockListInviteService{
		listInvitesFunc: func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
			return nil, fmt.Errorf("database connection failed")
		},
	}

	mockEventRepo := &mockListEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 1,
			}, nil
		},
	}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites", nil)
	req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListInvitesHandler_ResponseJSONStructure(t *testing.T) {
	email := "test@example.com"
	name := "Test User"
	now := time.Now()
	sentAt := now.Add(-1 * time.Hour)

	mockService := &mockListInviteService{
		listInvitesFunc: func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
			return &invites.ListInvitesResponse{
				Invites: []*models.Invite{
					{
						ID:          1,
						EventID:     1,
						Email:       &email,
						Name:        &name,
						TokenHash:   "hash123",
						MaxPlusOnes: 2,
						Status:      models.InviteStatusSent,
						SentAt:      &sentAt,
						ExpiresAt:   now.Add(30 * 24 * time.Hour),
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				},
				Total: 1,
				Stats: &repositories.InviteStats{
					Total: 1,
					Sent:  1,
				},
			}, nil
		},
	}

	mockEventRepo := &mockListEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				Title:     "Test Event",
				Status:    models.EventStatusPublished,
				CreatedBy: 1,
			}, nil
		},
	}

	handler := NewListInviteHandlers(mockService, mockEventRepo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites", nil)
			req.Header.Set("Accept", "application/json")
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response invites.ListInvitesResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Invites) != 1 {
		t.Errorf("expected 1 invite, got %d", len(response.Invites))
	}

	if response.Total != 1 {
		t.Errorf("expected total 1, got %d", response.Total)
	}

	if response.Stats == nil {
		t.Fatal("expected stats, got nil")
	}

	if response.Stats.Total != 1 {
		t.Errorf("expected stats total 1, got %d", response.Stats.Total)
	}

	if response.Stats.Sent != 1 {
		t.Errorf("expected stats sent 1, got %d", response.Stats.Sent)
	}

	invite := response.Invites[0]
	if invite.ID != 1 {
		t.Errorf("expected invite ID 1, got %d", invite.ID)
	}
	if invite.Email == nil || *invite.Email != email {
		t.Errorf("expected email %s, got %v", email, invite.Email)
	}
	if invite.Name == nil || *invite.Name != name {
		t.Errorf("expected name %s, got %v", name, invite.Name)
	}
	if invite.Status != models.InviteStatusSent {
		t.Errorf("expected status %s, got %s", models.InviteStatusSent, invite.Status)
	}
}

func TestListInvitesHandler_LimitBoundaryValues(t *testing.T) {
	tests := []struct {
		name       string
		limit      string
		wantStatus int
		wantLimit  int
	}{
		{
			name:       "limit 1",
			limit:      "1",
			wantStatus: http.StatusOK,
			wantLimit:  1,
		},
		{
			name:       "limit 100",
			limit:      "100",
			wantStatus: http.StatusOK,
			wantLimit:  100,
		},
		{
			name:       "limit 0",
			limit:      "0",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "limit 101",
			limit:      "101",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockListInviteService{
				listInvitesFunc: func(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
					if tt.wantStatus == http.StatusOK && req.Limit != tt.wantLimit {
						t.Errorf("expected limit %d, got %d", tt.wantLimit, req.Limit)
					}
					return &invites.ListInvitesResponse{
						Invites: []*models.Invite{},
						Total:   0,
						Stats:   &repositories.InviteStats{},
					}, nil
				},
			}

			mockEventRepo := &mockListEventRepository{
				getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        1,
						Title:     "Test Event",
						Status:    models.EventStatusPublished,
						CreatedBy: 1,
					}, nil
				},
			}

			handler := NewListInviteHandlers(mockService, mockEventRepo)

			r := chi.NewRouter()
			handler.RegisterRoutes(r)

			req := httptest.NewRequest(http.MethodGet, "/api/events/1/invites?limit="+tt.limit, nil)
			req.Header.Set("Accept", "application/json")
			user := &models.User{ID: 1, Role: models.RoleAdmin}
			ctx := auth.WithUser(req.Context(), user)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}


func (m *mockListEventRepository) CountEvents(ctx context.Context) (int, error) {
	return 0, nil
}
