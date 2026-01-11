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

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockImageEventServiceWithUpdate struct {
	GetEventFunc    func(ctx context.Context, id int64) (*models.Event, error)
	UpdateEventFunc func(ctx context.Context, event *models.Event) error
}

func (m *mockImageEventServiceWithUpdate) GetEvent(ctx context.Context, id int64) (*models.Event, error) {
	if m.GetEventFunc != nil {
		return m.GetEventFunc(ctx, id)
	}
	return &models.Event{
		ID:        id,
		CreatedBy: 1,
	}, nil
}

func (m *mockImageEventServiceWithUpdate) UpdateEvent(ctx context.Context, event *models.Event) error {
	if m.UpdateEventFunc != nil {
		return m.UpdateEventFunc(ctx, event)
	}
	return nil
}

func TestImageHandlers_UploadImage_UpdatesEventCustomThemeImageURL(t *testing.T) {
	var updatedEvent *models.Event
	
	imageService := &mockImageService{
		UploadImageFunc: func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
			return &assets.ImageMetadata{
				Path:        "images/123/test_abc123.jpg",
				PublicURL:   "http://localhost:8080/assets/images/123/test_abc123.jpg",
				Filename:    "test_abc123.jpg",
				ContentType: "image/jpeg",
				Size:        1024,
				Width:       800,
				Height:      600,
			}, nil
		},
	}
	
	eventService := &mockImageEventServiceWithUpdate{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        id,
				CreatedBy: 1,
				Title:     "Test Event",
			}, nil
		},
		UpdateEventFunc: func(ctx context.Context, event *models.Event) error {
			updatedEvent = event
			return nil
		},
	}
	
	authz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return true
		},
	}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.jpg")
	part.Write([]byte{0xFF, 0xD8, 0xFF})
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/events/123/images", body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handlers.UploadImage(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	if updatedEvent == nil {
		t.Fatal("Expected event to be updated")
	}

	if updatedEvent.CustomThemeImageURL == nil {
		t.Fatal("Expected CustomThemeImageURL to be set")
	}

	expectedURL := "http://localhost:8080/assets/images/123/test_abc123.jpg"
	if *updatedEvent.CustomThemeImageURL != expectedURL {
		t.Errorf("CustomThemeImageURL = %s, want %s", *updatedEvent.CustomThemeImageURL, expectedURL)
	}
}

func TestImageHandlers_UploadImage_DeletesOldImage(t *testing.T) {
	var deletedPath string
	oldImageURL := "http://localhost:8080/assets/images/123/old_image.jpg"
	
	imageService := &mockImageService{
		UploadImageFunc: func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
			return &assets.ImageMetadata{
				Path:        "images/123/new_image.jpg",
				PublicURL:   "http://localhost:8080/assets/images/123/new_image.jpg",
				Filename:    "new_image.jpg",
				ContentType: "image/jpeg",
				Size:        1024,
				Width:       800,
				Height:      600,
			}, nil
		},
		DeleteImageFunc: func(ctx context.Context, path string) error {
			deletedPath = path
			return nil
		},
	}
	
	eventService := &mockImageEventServiceWithUpdate{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:                  id,
				CreatedBy:           1,
				Title:               "Test Event",
				CustomThemeImageURL: &oldImageURL,
			}, nil
		},
		UpdateEventFunc: func(ctx context.Context, event *models.Event) error {
			return nil
		},
	}
	
	authz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return true
		},
	}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "new.jpg")
	part.Write([]byte{0xFF, 0xD8, 0xFF})
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/events/123/images", body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handlers.UploadImage(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	expectedDeletedPath := "images/123/old_image.jpg"
	if deletedPath != expectedDeletedPath {
		t.Errorf("Deleted path = %s, want %s", deletedPath, expectedDeletedPath)
	}
}

func TestImageHandlers_UploadImage_NoOldImageToDelete(t *testing.T) {
	deleteImageCalled := false
	
	imageService := &mockImageService{
		UploadImageFunc: func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
			return &assets.ImageMetadata{
				Path:        "images/123/new_image.jpg",
				PublicURL:   "http://localhost:8080/assets/images/123/new_image.jpg",
				Filename:    "new_image.jpg",
				ContentType: "image/jpeg",
				Size:        1024,
				Width:       800,
				Height:      600,
			}, nil
		},
		DeleteImageFunc: func(ctx context.Context, path string) error {
			deleteImageCalled = true
			return nil
		},
	}
	
	eventService := &mockImageEventServiceWithUpdate{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:                  id,
				CreatedBy:           1,
				Title:               "Test Event",
				CustomThemeImageURL: nil,
			}, nil
		},
		UpdateEventFunc: func(ctx context.Context, event *models.Event) error {
			return nil
		},
	}
	
	authz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return true
		},
	}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "new.jpg")
	part.Write([]byte{0xFF, 0xD8, 0xFF})
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/events/123/images", body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handlers.UploadImage(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	if deleteImageCalled {
		t.Error("DeleteImage should not be called when there is no old image")
	}
}

func TestImageHandlers_UploadImage_UpdateEventFails(t *testing.T) {
	imageService := &mockImageService{
		UploadImageFunc: func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
			return &assets.ImageMetadata{
				Path:        "images/123/test.jpg",
				PublicURL:   "http://localhost:8080/assets/images/123/test.jpg",
				Filename:    "test.jpg",
				ContentType: "image/jpeg",
				Size:        1024,
				Width:       800,
				Height:      600,
			}, nil
		},
	}
	
	eventService := &mockImageEventServiceWithUpdate{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        id,
				CreatedBy: 1,
				Title:     "Test Event",
			}, nil
		},
		UpdateEventFunc: func(ctx context.Context, event *models.Event) error {
			return &models.ValidationError{Field: "event", Message: "update failed"}
		},
	}
	
	authz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return true
		},
	}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.jpg")
	part.Write([]byte{0xFF, 0xD8, 0xFF})
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/events/123/images", body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handlers.UploadImage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	
	if response["error"] == nil {
		t.Error("Expected error in response")
	}
}
