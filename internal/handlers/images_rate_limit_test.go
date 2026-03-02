package handlers

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestImageUpload_RateLimiting(t *testing.T) {
	mockService := &mockImageService{
		UploadImageFunc: func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
			return &assets.ImageMetadata{
				Path:        "images/1/test.jpg",
				PublicURL:   "http://example.com/images/1/test.jpg",
				Filename:    "test.jpg",
				ContentType: "image/jpeg",
				Size:        1024,
				Width:       100,
				Height:      100,
			}, nil
		},
	}

	mockEventService := &mockImageEventService{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				CreatedBy: 1,
			}, nil
		},
		UpdateEventFunc: func(ctx context.Context, event *models.Event) error {
			return nil
		},
	}

	mockAuthz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return true
		},
	}

	handlers := NewImageHandlers(mockService, mockEventService, mockAuthz)

	rateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		RequestsPerMinute: 5,
		BurstSize:         5,
	})
	defer rateLimiter.Stop()

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RateLimit(rateLimiter, middleware.RateLimitConfig{
		AnonymousLimit:     0,
		AuthenticatedLimit: 5,
		AdminLimit:         100,
	}))
	handlers.RegisterRoutes(r)

	user := &models.User{
		ID:   1,
		Role: models.RoleEventManager,
	}

	successCount := 0
	rateLimitCount := 0

	for i := 0; i < 10; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", "test.jpg")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		img := createTestImage(100, 100)
		if err := jpeg.Encode(part, img, nil); err != nil {
			t.Fatalf("Failed to encode image: %v", err)
		}

		writer.Close()

		req := httptest.NewRequest("POST", "/api/events/1/images", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-Real-IP", "192.168.1.1")

		req = req.WithContext(auth.WithUser(req.Context(), user))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code == http.StatusCreated {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			rateLimitCount++

			if w.Header().Get("Retry-After") == "" {
				t.Error("Expected Retry-After header on rate limit response")
			}
			if w.Header().Get("X-RateLimit-Limit") == "" {
				t.Error("Expected X-RateLimit-Limit header")
			}
		}
	}

	if successCount != 5 {
		t.Errorf("Expected 5 successful uploads, got %d", successCount)
	}

	if rateLimitCount != 5 {
		t.Errorf("Expected 5 rate limited requests, got %d", rateLimitCount)
	}
}

func TestImageUpload_RateLimitByIP(t *testing.T) {
	mockService := &mockImageService{
		UploadImageFunc: func(ctx context.Context, eventID int64, filename string, data io.Reader) (*assets.ImageMetadata, error) {
			return &assets.ImageMetadata{
				Path:        "images/1/test.jpg",
				PublicURL:   "http://example.com/images/1/test.jpg",
				Filename:    "test.jpg",
				ContentType: "image/jpeg",
				Size:        1024,
				Width:       100,
				Height:      100,
			}, nil
		},
	}

	mockEventService := &mockImageEventService{
		GetEventFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        1,
				CreatedBy: 1,
			}, nil
		},
		UpdateEventFunc: func(ctx context.Context, event *models.Event) error {
			return nil
		},
	}

	mockAuthz := &mockImageAuthz{
		CanEditEventFunc: func(ctx context.Context, user *models.User, event *models.Event) bool {
			return true
		},
	}

	handlers := NewImageHandlers(mockService, mockEventService, mockAuthz)

	rateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		RequestsPerMinute: 3,
		BurstSize:         3,
	})
	defer rateLimiter.Stop()

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RateLimit(rateLimiter, middleware.RateLimitConfig{
		AnonymousLimit:     0,
		AuthenticatedLimit: 3,
		AdminLimit:         100,
	}))
	handlers.RegisterRoutes(r)

	user := &models.User{
		ID:   1,
		Role: models.RoleEventManager,
	}

	ips := []string{"192.168.1.1", "192.168.1.2"}

	for _, ip := range ips {
		successCount := 0

		for i := 0; i < 5; i++ {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", "test.jpg")
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}

			img := createTestImage(100, 100)
			if err := jpeg.Encode(part, img, nil); err != nil {
				t.Fatalf("Failed to encode image: %v", err)
			}

			writer.Close()

			req := httptest.NewRequest("POST", "/api/events/1/images", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("X-Real-IP", ip)

			req = req.WithContext(auth.WithUser(req.Context(), user))

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code == http.StatusCreated {
				successCount++
			}
		}

		if successCount != 3 {
			t.Errorf("IP %s: Expected 3 successful uploads, got %d", ip, successCount)
		}
	}
}

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	return img
}
