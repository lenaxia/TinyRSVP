package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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

type mockImportService struct {
	importCSVFunc            func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error)
	createManualInviteFunc   func(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error)
	cleanupExpiredTokensFunc func(ctx context.Context) (int64, error)
}

func (m *mockImportService) CreateInvite(ctx context.Context, eventID int64, name *string, email *string, maxPlusOnes int, expiresAt time.Time) (*models.Invite, string, error) {
	return nil, "", nil
}

func (m *mockImportService) CreateManualInvite(ctx context.Context, req *invites.CreateManualInviteRequest, expiresAt time.Time) (*invites.CreateManualInviteResponse, error) {
	if m.createManualInviteFunc != nil {
		return m.createManualInviteFunc(ctx, req, expiresAt)
	}
	return nil, nil
}

func (m *mockImportService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	return nil, nil
}

func (m *mockImportService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
	return nil, nil
}

func (m *mockImportService) RevokeInvite(ctx context.Context, req *invites.RevokeInviteRequest) error {
	return nil
}

func (m *mockImportService) RegenerateToken(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error) {
	return nil, nil
}

func (m *mockImportService) ListInvitesByEventID(ctx context.Context, eventID int64, filters repositories.InviteFilters) ([]*models.Invite, error) {
	return nil, nil
}

func (m *mockImportService) ImportCSV(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
	if m.importCSVFunc != nil {
		return m.importCSVFunc(ctx, eventID, csvData, defaultMaxPlusOnes, expiresAt)
	}
	return nil, nil
}

func (m *mockImportService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	if m.cleanupExpiredTokensFunc != nil {
		return m.cleanupExpiredTokensFunc(ctx)
	}
	return 0, nil
}

func (m *mockImportService) MarkInviteSent(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockImportService) MarkInviteViewed(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockImportService) MarkInviteResponded(ctx context.Context, inviteID int64) error {
	return nil
}

func (m *mockImportService) ListInvites(ctx context.Context, req *invites.ListInvitesRequest) (*invites.ListInvitesResponse, error) {
	return &invites.ListInvitesResponse{
		Invites: []*models.Invite{},
		Total:   0,
		Stats:   &repositories.InviteStats{},
	}, nil
}

func createMultipartRequest(csvContent string, filename string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	if _, err := io.WriteString(part, csvContent); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func createMockEventRepo(userID int64) *mockEventRepository {
	return &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:          1,
				Title:       "Test Event",
				StartTime:   time.Now().Add(30 * 24 * time.Hour),
				Status:      models.EventStatusDraft,
				CreatedBy:   userID,
				MaxPlusOnes: 2,
			}, nil
		},
	}
}

func TestImportInvitesHandler_Success(t *testing.T) {
	csvContent := `email,name,max_plus_ones
john@example.com,John Doe,2
jane@example.com,Jane Smith,1`

	mockService := &mockImportService{
		importCSVFunc: func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
			return &invites.ImportResult{
				Total:      2,
				Created:    2,
				Failed:     0,
				Duplicates: 0,
				Errors:     []invites.ImportError{},
			}, nil
		},
	}

	mockEventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:          1,
				Title:       "Test Event",
				StartTime:   time.Now().Add(30 * 24 * time.Hour),
				Status:      models.EventStatusDraft,
				CreatedBy:   100,
				MaxPlusOnes: 2,
			}, nil
		},
	}
	handler := NewImportInviteHandlers(mockService, mockEventRepo, "https://rsvp.example.com")

	req, err := createMultipartRequest(csvContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response invites.ImportResult
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Total != 2 {
		t.Errorf("Expected total 2, got %d", response.Total)
	}

	if response.Created != 2 {
		t.Errorf("Expected created 2, got %d", response.Created)
	}
}

func TestImportInvitesHandler_PartialSuccess(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe
invalid-email,Jane Smith
bob@example.com,Bob Johnson`

	mockService := &mockImportService{
		importCSVFunc: func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
			return &invites.ImportResult{
				Total:      3,
				Created:    2,
				Failed:     1,
				Duplicates: 0,
				Errors: []invites.ImportError{
					{
						Row:     3,
						Email:   "invalid-email",
						Message: "invalid email format",
					},
				},
			}, nil
		},
	}

	handler := NewImportInviteHandlers(mockService, createMockEventRepo(100), "https://rsvp.example.com")

	req, err := createMultipartRequest(csvContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response invites.ImportResult
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Failed != 1 {
		t.Errorf("Expected failed 1, got %d", response.Failed)
	}

	if len(response.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(response.Errors))
	}
}

func TestImportInvitesHandler_NoFile(t *testing.T) {
	mockService := &mockImportService{}
	handler := NewImportInviteHandlers(mockService, createMockEventRepo(100), "https://rsvp.example.com")

	req := httptest.NewRequest(http.MethodPost, "/api/events/1/invites/import", nil)
	req.Header.Set("Content-Type", "multipart/form-data")

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestImportInvitesHandler_InvalidFileExtension(t *testing.T) {
	mockService := &mockImportService{}
	handler := NewImportInviteHandlers(mockService, createMockEventRepo(100), "https://rsvp.example.com")

	req, err := createMultipartRequest("test content", "guests.txt")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestImportInvitesHandler_FileTooLarge(t *testing.T) {
	mockService := &mockImportService{}
	handler := NewImportInviteHandlers(mockService, createMockEventRepo(100), "https://rsvp.example.com")

	largeContent := strings.Repeat("a", 2*1024*1024)
	req, err := createMultipartRequest(largeContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestImportInvitesHandler_Unauthorized(t *testing.T) {
	mockService := &mockImportService{}
	handler := NewImportInviteHandlers(mockService, createMockEventRepo(100), "https://rsvp.example.com")

	req, err := createMultipartRequest("email\ntest@example.com", "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestImportInvitesHandler_InvalidEventID(t *testing.T) {
	mockService := &mockImportService{}
	handler := NewImportInviteHandlers(mockService, createMockEventRepo(100), "https://rsvp.example.com")

	req, err := createMultipartRequest("email\ntest@example.com", "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestImportInvitesHandler_ServiceError(t *testing.T) {
	mockService := &mockImportService{
		importCSVFunc: func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
			return nil, &models.ValidationError{Field: "csv", Message: "CSV exceeds 500 row limit"}
		},
	}

	handler := NewImportInviteHandlers(mockService, createMockEventRepo(100), "https://rsvp.example.com")

	req, err := createMultipartRequest("email\ntest@example.com", "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleEventManager}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
