package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/storage"
)

func TestImageUpload_Integration_CompleteFlow(t *testing.T) {
	storageProvider := storage.NewMockProvider()
	imageService := assets.NewImageService(storageProvider)
	
	var updatedEvent *models.Event
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
			return user.ID == event.CreatedBy
		},
	}

	handlers := NewImageHandlers(imageService, eventService, authz)

	jpegData := createValidJPEG()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test-image.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	
	if _, err := part.Write(jpegData); err != nil {
		t.Fatalf("Failed to write file data: %v", err)
	}
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
		t.Fatalf("Status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var response ImageUploadResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Image == nil {
		t.Fatal("Expected image metadata in response")
	}

	if response.Image.PublicURL == "" {
		t.Error("Expected non-empty public URL")
	}

	if response.Image.ContentType != "image/jpeg" {
		t.Errorf("ContentType = %s, want image/jpeg", response.Image.ContentType)
	}

	if updatedEvent == nil {
		t.Fatal("Expected event to be updated")
	}

	if updatedEvent.CustomThemeImageURL == nil {
		t.Fatal("Expected CustomThemeImageURL to be set")
	}

	if *updatedEvent.CustomThemeImageURL != response.Image.PublicURL {
		t.Errorf("Event CustomThemeImageURL = %s, want %s", *updatedEvent.CustomThemeImageURL, response.Image.PublicURL)
	}

	objects, err := storageProvider.ListObjects(context.Background(), "images/123/")
	if err != nil {
		t.Fatalf("Failed to list objects: %v", err)
	}

	if len(objects) != 1 {
		t.Errorf("Expected 1 stored object, got %d", len(objects))
	}
}

func TestImageUpload_Integration_ReplaceExistingImage(t *testing.T) {
	storageProvider := storage.NewMockProvider()
	imageService := assets.NewImageService(storageProvider)
	
	oldImageURL := "http://localhost:8080/assets/images/123/old-image.jpg"
	ctx := context.Background()
	
	oldImageData := []byte{0xFF, 0xD8, 0xFF, 0xD9}
	if err := storageProvider.PutObject(ctx, "images/123/old-image.jpg", bytes.NewReader(oldImageData), "image/jpeg"); err != nil {
		t.Fatalf("Failed to setup old image: %v", err)
	}
	
	var updatedEvent *models.Event
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
			updatedEvent = event
			return nil
		},
	}
	
	authz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return user.ID == event.CreatedBy
		},
	}

	handlers := NewImageHandlers(imageService, eventService, authz)

	jpegData := createValidJPEG()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "new-image.jpg")
	part.Write(jpegData)
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
		t.Fatalf("Status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	if updatedEvent == nil {
		t.Fatal("Expected event to be updated")
	}

	if updatedEvent.CustomThemeImageURL == nil {
		t.Fatal("Expected CustomThemeImageURL to be set")
	}

	if *updatedEvent.CustomThemeImageURL == oldImageURL {
		t.Error("Expected CustomThemeImageURL to be updated with new image URL")
	}

	_, err := storageProvider.GetObject(ctx, "images/123/old-image.jpg")
	if err == nil {
		t.Error("Expected old image to be deleted")
	}

	objects, err := storageProvider.ListObjects(ctx, "images/123/")
	if err != nil {
		t.Fatalf("Failed to list objects: %v", err)
	}

	if len(objects) != 1 {
		t.Errorf("Expected 1 stored object (new image), got %d", len(objects))
	}
}

func TestImageUpload_Integration_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		fileData    []byte
		filename    string
		wantStatus  int
		wantErrMsg  string
	}{
		{
			name:       "file too large",
			fileData:   make([]byte, 6*1024*1024),
			filename:   "large.jpg",
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "exceeds",
		},
		{
			name:       "invalid image format",
			fileData:   []byte("not an image"),
			filename:   "invalid.jpg",
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "allowed",
		},
		{
			name:       "empty file",
			fileData:   []byte{},
			filename:   "empty.jpg",
			wantStatus: http.StatusBadRequest,
			wantErrMsg: "allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageProvider := storage.NewMockProvider()
			imageService := assets.NewImageService(storageProvider)
			
			eventService := &mockImageEventServiceWithUpdate{
				GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Title:     "Test Event",
					}, nil
				},
				UpdateEventFunc: func(ctx context.Context, event *models.Event) error {
					t.Error("UpdateEvent should not be called for validation errors")
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
			part, _ := writer.CreateFormFile("file", tt.filename)
			part.Write(tt.fileData)
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

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d. Body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			responseBody := w.Body.String()
			if tt.wantErrMsg != "" && !strings.Contains(responseBody, tt.wantErrMsg) {
				t.Errorf("Expected error message to contain %q, got: %s", tt.wantErrMsg, responseBody)
			}
		})
	}
}

func createValidJPEG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		panic(err)
	}
	
	return buf.Bytes()
}
