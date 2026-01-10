package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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

type mockEventRepository struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockEventRepository) Create(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockEventRepository) Update(ctx context.Context, event *models.Event) error {
	return nil
}

func (m *mockEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return nil
}

func (m *mockEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return nil
}

func (m *mockEventRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, nil
}

func TestImportInvitesHandler_PermissionDenied_NotAdmin_NotCreator(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe`

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:          1,
				Title:       "Test Event",
				StartTime:   time.Now().Add(30 * 24 * time.Hour),
				Status:      models.EventStatusDraft,
				CreatedBy:   999,
				MaxPlusOnes: 2,
			}, nil
		},
	}

	mockService := &mockImportService{
		importCSVFunc: func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
			t.Fatal("ImportCSV should not be called when permission is denied")
			return nil, nil
		},
	}

	handler := NewImportInviteHandlers(mockService, eventRepo, "https://rsvp.example.com")

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

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestImportInvitesHandler_PermissionGranted_Admin(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe`

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:          1,
				Title:       "Test Event",
				StartTime:   time.Now().Add(30 * 24 * time.Hour),
				Status:      models.EventStatusDraft,
				CreatedBy:   999,
				MaxPlusOnes: 2,
			}, nil
		},
	}

	mockService := &mockImportService{
		importCSVFunc: func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
			return &invites.ImportResult{
				Total:      1,
				Created:    1,
				Failed:     0,
				Duplicates: 0,
				Errors:     []invites.ImportError{},
			}, nil
		},
	}

	handler := NewImportInviteHandlers(mockService, eventRepo, "https://rsvp.example.com")

	req, err := createMultipartRequest(csvContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleAdmin}
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
}

func TestImportInvitesHandler_PermissionGranted_Creator(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe`

	eventRepo := &mockEventRepository{
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

	mockService := &mockImportService{
		importCSVFunc: func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
			return &invites.ImportResult{
				Total:      1,
				Created:    1,
				Failed:     0,
				Duplicates: 0,
				Errors:     []invites.ImportError{},
			}, nil
		},
	}

	handler := NewImportInviteHandlers(mockService, eventRepo, "https://rsvp.example.com")

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
}

func TestImportInvitesHandler_EventNotFound(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe`

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return nil, &models.NotFoundError{
				Resource: "Event",
				ID:       id,
			}
		},
	}

	mockService := &mockImportService{}

	handler := NewImportInviteHandlers(mockService, eventRepo, "https://rsvp.example.com")

	req, err := createMultipartRequest(csvContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleAdmin}
	ctx := auth.WithUser(req.Context(), user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("eventId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ImportInvites(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestImportInvitesHandler_CancelledEvent(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe`

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:          1,
				Title:       "Test Event",
				StartTime:   time.Now().Add(30 * 24 * time.Hour),
				Status:      models.EventStatusCancelled,
				CreatedBy:   100,
				MaxPlusOnes: 2,
			}, nil
		},
	}

	mockService := &mockImportService{}

	handler := NewImportInviteHandlers(mockService, eventRepo, "https://rsvp.example.com")

	req, err := createMultipartRequest(csvContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleAdmin}
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

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "cannot create invite for cancelled event" {
		t.Errorf("Expected cancelled event error, got: %s", response["message"])
	}
}

func TestImportInvitesHandler_ArchivedEvent(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe`

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:          1,
				Title:       "Test Event",
				StartTime:   time.Now().Add(30 * 24 * time.Hour),
				Status:      models.EventStatusArchived,
				CreatedBy:   100,
				MaxPlusOnes: 2,
			}, nil
		},
	}

	mockService := &mockImportService{}

	handler := NewImportInviteHandlers(mockService, eventRepo, "https://rsvp.example.com")

	req, err := createMultipartRequest(csvContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleAdmin}
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

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "cannot create invite for archived event" {
		t.Errorf("Expected archived event error, got: %s", response["message"])
	}
}

func TestImportInvitesHandler_CorrectExpirationTime(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe`

	eventStartTime := time.Now().Add(60 * 24 * time.Hour)
	expectedExpiration := eventStartTime.Add(30 * 24 * time.Hour)

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:          1,
				Title:       "Test Event",
				StartTime:   eventStartTime,
				Status:      models.EventStatusDraft,
				CreatedBy:   100,
				MaxPlusOnes: 2,
			}, nil
		},
	}

	var capturedExpiresAt time.Time
	mockService := &mockImportService{
		importCSVFunc: func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
			capturedExpiresAt = expiresAt
			return &invites.ImportResult{
				Total:      1,
				Created:    1,
				Failed:     0,
				Duplicates: 0,
				Errors:     []invites.ImportError{},
			}, nil
		},
	}

	handler := NewImportInviteHandlers(mockService, eventRepo, "https://rsvp.example.com")

	req, err := createMultipartRequest(csvContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleAdmin}
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

	if capturedExpiresAt.Sub(expectedExpiration).Abs() > time.Second {
		t.Errorf("Expected expiration %v, got %v", expectedExpiration, capturedExpiresAt)
	}
}

func TestImportInvitesHandler_CorrectDefaultMaxPlusOnes(t *testing.T) {
	csvContent := `email,name
john@example.com,John Doe`

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:          1,
				Title:       "Test Event",
				StartTime:   time.Now().Add(30 * 24 * time.Hour),
				Status:      models.EventStatusDraft,
				CreatedBy:   100,
				MaxPlusOnes: 3,
			}, nil
		},
	}

	var capturedMaxPlusOnes int
	mockService := &mockImportService{
		importCSVFunc: func(ctx context.Context, eventID int64, csvData []byte, defaultMaxPlusOnes int, expiresAt time.Time) (*invites.ImportResult, error) {
			capturedMaxPlusOnes = defaultMaxPlusOnes
			return &invites.ImportResult{
				Total:      1,
				Created:    1,
				Failed:     0,
				Duplicates: 0,
				Errors:     []invites.ImportError{},
			}, nil
		},
	}

	handler := NewImportInviteHandlers(mockService, eventRepo, "https://rsvp.example.com")

	req, err := createMultipartRequest(csvContent, "guests.csv")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	user := &models.User{ID: 100, Role: models.RoleAdmin}
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

	if capturedMaxPlusOnes != 3 {
		t.Errorf("Expected defaultMaxPlusOnes 3, got %d", capturedMaxPlusOnes)
	}
}

func createMultipartRequestHelper(csvContent string, filename string) (*http.Request, error) {
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
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
