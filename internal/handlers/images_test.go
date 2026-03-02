package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockImageService struct {
	UploadImageFunc func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error)
	DeleteImageFunc func(ctx context.Context, path string) error
	GetImageURLFunc func(ctx context.Context, path string) (string, error)
}

func (m *mockImageService) UploadImage(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
	if m.UploadImageFunc != nil {
		return m.UploadImageFunc(ctx, eventID, filename, data)
	}
	return &assets.ImageMetadata{
		Path:        "images/123/test.jpg",
		PublicURL:   "http://localhost:8080/assets/images/123/test.jpg",
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
		Size:        1024,
		Width:       800,
		Height:      600,
	}, nil
}

func (m *mockImageService) DeleteImage(ctx context.Context, path string) error {
	if m.DeleteImageFunc != nil {
		return m.DeleteImageFunc(ctx, path)
	}
	return nil
}

func (m *mockImageService) GetImageURL(ctx context.Context, path string) (string, error) {
	if m.GetImageURLFunc != nil {
		return m.GetImageURLFunc(ctx, path)
	}
	return "http://localhost:8080/assets/" + path, nil
}

type mockImageEventService struct {
	GetEventFunc    func(ctx context.Context, id int64) (*models.Event, error)
	UpdateEventFunc func(ctx context.Context, event *models.Event) error
}

func (m *mockImageEventService) GetEvent(ctx context.Context, id int64) (*models.Event, error) {
	if m.GetEventFunc != nil {
		return m.GetEventFunc(ctx, id)
	}
	return &models.Event{
		ID:        id,
		CreatedBy: 1,
	}, nil
}

func (m *mockImageEventService) UpdateEvent(ctx context.Context, event *models.Event) error {
	if m.UpdateEventFunc != nil {
		return m.UpdateEventFunc(ctx, event)
	}
	return nil
}

type mockImageAuthz struct {
	CanEditEventFunc func(ctx context.Context, user *models.User, event *models.Event) bool
}

func (m *mockImageAuthz) CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool {
	if m.CanEditEventFunc != nil {
		return m.CanEditEventFunc(ctx, user, event)
	}
	return true
}

func TestImageHandlers_UploadImage_Success(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{}
	authz := &mockImageAuthz{}

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

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	image, ok := response["image"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected image in response")
	}

	if image["path"] == nil {
		t.Error("Expected path in image response")
	}
	if image["public_url"] == nil {
		t.Error("Expected public_url in image response")
	}
}

func TestImageHandlers_UploadImage_Unauthorized(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{}
	authz := &mockImageAuthz{}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.CreateFormFile("file", "test.jpg")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/events/123/images", body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handlers.UploadImage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestImageHandlers_UploadImage_InvalidEventID(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{}
	authz := &mockImageAuthz{}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.CreateFormFile("file", "test.jpg")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/events/invalid/images", body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handlers.UploadImage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestImageHandlers_UploadImage_EventNotFound(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return nil, &models.NotFoundError{Resource: "Event", ID: id}
		},
	}
	authz := &mockImageAuthz{}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.CreateFormFile("file", "test.jpg")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/events/999/images", body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handlers.UploadImage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestImageHandlers_UploadImage_Forbidden(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        id,
				CreatedBy: 999,
			}, nil
		},
	}
	authz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return false
		},
	}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.CreateFormFile("file", "test.jpg")
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

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestImageHandlers_UploadImage_NoFile(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{}
	authz := &mockImageAuthz{}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
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

	if !strings.Contains(w.Body.String(), "file") {
		t.Errorf("Error message should mention 'file', got: %s", w.Body.String())
	}
}

func TestImageHandlers_UploadImage_ValidationError(t *testing.T) {
	imageService := &mockImageService{
		UploadImageFunc: func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
			return nil, &assets.ValidationError{
				Field:   "file",
				Message: "Image size exceeds 5MB",
			}
		},
	}
	eventService := &mockImageEventService{}
	authz := &mockImageAuthz{}

	handlers := NewImageHandlers(imageService, eventService, authz)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.jpg")
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

	if !strings.Contains(w.Body.String(), "exceeds") {
		t.Errorf("Error message should mention size limit, got: %s", w.Body.String())
	}
}

func TestImageHandlers_UploadImage_ServiceError(t *testing.T) {
	imageService := &mockImageService{
		UploadImageFunc: func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
			return nil, errors.New("storage failure")
		},
	}
	eventService := &mockImageEventService{}
	authz := &mockImageAuthz{}

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

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestImageHandlers_DeleteImage_Success(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{}
	authz := &mockImageAuthz{}

	handlers := NewImageHandlers(imageService, eventService, authz)

	req := httptest.NewRequest(http.MethodDelete, "/api/events/123/images/test.jpg", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "123")
	rctx.URLParams.Add("filename", "test.jpg")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handlers.DeleteImage(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestImageHandlers_DeleteImage_Unauthorized(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{}
	authz := &mockImageAuthz{}

	handlers := NewImageHandlers(imageService, eventService, authz)

	req := httptest.NewRequest(http.MethodDelete, "/api/events/123/images/test.jpg", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "123")
	rctx.URLParams.Add("filename", "test.jpg")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handlers.DeleteImage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestImageHandlers_DeleteImage_Forbidden(t *testing.T) {
	imageService := &mockImageService{}
	eventService := &mockImageEventService{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        id,
				CreatedBy: 999,
			}, nil
		},
	}
	authz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return false
		},
	}

	handlers := NewImageHandlers(imageService, eventService, authz)

	req := httptest.NewRequest(http.MethodDelete, "/api/events/123/images/test.jpg", nil)
	req.Header.Set("Accept", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("event_id", "123")
	rctx.URLParams.Add("filename", "test.jpg")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleEventManager}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	handlers.DeleteImage(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}
