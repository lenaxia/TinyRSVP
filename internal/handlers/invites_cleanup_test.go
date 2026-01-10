package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockInviteServiceWithCleanup struct {
	cleanupExpiredTokensFunc func(ctx context.Context) (int64, error)
}

func (m *mockInviteServiceWithCleanup) CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error) {
	return nil, "", nil
}

func (m *mockInviteServiceWithCleanup) CreateManualInvite(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error) {
	return nil, nil
}

func (m *mockInviteServiceWithCleanup) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	return nil, nil
}

func (m *mockInviteServiceWithCleanup) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	return nil, nil
}

func (m *mockInviteServiceWithCleanup) RevokeInvite(ctx context.Context, req *invites.RevokeInviteRequest) error {
	return nil
}

func (m *mockInviteServiceWithCleanup) RegenerateToken(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
	return nil, nil
}

func (m *mockInviteServiceWithCleanup) ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockInviteServiceWithCleanup) ImportCSV(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
	return nil, nil
}

func (m *mockInviteServiceWithCleanup) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	if m.cleanupExpiredTokensFunc != nil {
		return m.cleanupExpiredTokensFunc(ctx)
	}
	return 0, nil
}

func (m *mockInviteServiceWithCleanup) MarkInviteSent(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockInviteServiceWithCleanup) MarkInviteViewed(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockInviteServiceWithCleanup) MarkInviteResponded(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockInviteServiceWithCleanup) ListInvites(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
	return &invites.ListInvitesResponse{
		Invites: []*models.Invite{},
		Total:   0,
		Stats:   &repositories.InviteStats{},
	}, nil
}

func (m *mockInviteServiceWithCleanup) UpdateInvite(ctx context.Context, req *invites.UpdateInviteRequest) error {
	return nil
}

func (m *mockInviteServiceWithCleanup) DeleteInvite(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockInviteServiceWithCleanup) SendInvite(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error {
	return nil
}

func TestCleanupExpiredTokensHandler_Success(t *testing.T) {
	mockService := &mockInviteServiceWithCleanup{
		cleanupExpiredTokensFunc: func(ctx context.Context) (int64, error) {
			return 5, nil
		},
	}

	handler := NewCleanupHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/invites/cleanup", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if deleted, ok := response["deleted"].(float64); !ok || int64(deleted) != 5 {
		t.Errorf("expected deleted=5, got %v", response["deleted"])
	}

	if message, ok := response["message"].(string); !ok || message == "" {
		t.Errorf("expected non-empty message, got %v", response["message"])
	}
}

func TestCleanupExpiredTokensHandler_NoExpiredTokens(t *testing.T) {
	mockService := &mockInviteServiceWithCleanup{
		cleanupExpiredTokensFunc: func(ctx context.Context) (int64, error) {
			return 0, nil
		},
	}

	handler := NewCleanupHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/invites/cleanup", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if deleted, ok := response["deleted"].(float64); !ok || int64(deleted) != 0 {
		t.Errorf("expected deleted=0, got %v", response["deleted"])
	}
}

func TestCleanupExpiredTokensHandler_ServiceError(t *testing.T) {
	mockService := &mockInviteServiceWithCleanup{
		cleanupExpiredTokensFunc: func(ctx context.Context) (int64, error) {
			return 0, errors.New("database error")
		},
	}

	handler := NewCleanupHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/invites/cleanup", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if errorMsg, ok := response["error"].(string); !ok || errorMsg == "" {
		t.Errorf("expected non-empty error message, got %v", response["error"])
	}
}

func TestCleanupExpiredTokensHandler_InvalidMethod(t *testing.T) {
	mockService := &mockInviteServiceWithCleanup{}
	handler := NewCleanupHandler(mockService)

	tests := []struct {
		name   string
		method string
	}{
		{"GET not allowed", http.MethodGet},
		{"PUT not allowed", http.MethodPut},
		{"DELETE not allowed", http.MethodDelete},
		{"PATCH not allowed", http.MethodPatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/invites/cleanup", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
			}
		})
	}
}
