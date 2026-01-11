package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EventServiceForImages interface {
	GetEvent(ctx context.Context, id int64) (*models.Event, error)
	UpdateEvent(ctx context.Context, event *models.Event) error
}

type ImageAuthz interface {
	CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool
}

type ImageHandlers struct {
	imageService assets.ImageService
	eventService EventServiceForImages
	authz        ImageAuthz
}

func NewImageHandlers(imageService assets.ImageService, eventService EventServiceForImages, authz ImageAuthz) *ImageHandlers {
	return &ImageHandlers{
		imageService: imageService,
		eventService: eventService,
		authz:        authz,
	}
}

func (h *ImageHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/events/{event_id}/images", func(r chi.Router) {
		r.Post("/", h.UploadImage)
		r.Delete("/{filename}", h.DeleteImage)
	})
}

type ImageUploadResponse struct {
	Image *assets.ImageMetadata `json:"image"`
}

func (h *ImageHandlers) UploadImage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	eventID, err := parseEventID(chi.URLParam(r, "event_id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	event, err := h.eventService.GetEvent(r.Context(), eventID)
	if err != nil {
		handleImageServiceError(w, r, err)
		return
	}

	if !h.authz.CanEditEvent(r.Context(), user, event) {
		HandleError(w, r, NewPermissionDeniedError("you can only upload images for your own events"))
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		HandleError(w, r, NewBadRequestError("failed to parse multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		HandleError(w, r, NewBadRequestError("file field is required"))
		return
	}
	defer file.Close()

	metadata, err := h.imageService.UploadImage(r.Context(), eventID, header.Filename, file)
	if err != nil {
		handleImageServiceError(w, r, err)
		return
	}

	if event.CustomThemeImageURL != nil && *event.CustomThemeImageURL != "" {
		oldPath := extractStoragePathFromURL(*event.CustomThemeImageURL)
		if oldPath != "" {
			if err := h.imageService.DeleteImage(r.Context(), oldPath); err != nil {
			}
		}
	}

	event.CustomThemeImageURL = &metadata.PublicURL
	if err := h.eventService.UpdateEvent(r.Context(), event); err != nil {
		HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, ImageUploadResponse{
		Image: metadata,
	})
}

func extractStoragePathFromURL(url string) string {
	if url == "" {
		return ""
	}
	
	parts := []string{"/assets/", "/static/"}
	for _, part := range parts {
		if idx := findSubstring(url, part); idx >= 0 {
			return url[idx+len(part):]
		}
	}
	
	return ""
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (h *ImageHandlers) DeleteImage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	eventID, err := parseEventID(chi.URLParam(r, "event_id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	event, err := h.eventService.GetEvent(r.Context(), eventID)
	if err != nil {
		handleImageServiceError(w, r, err)
		return
	}

	if !h.authz.CanEditEvent(r.Context(), user, event) {
		HandleError(w, r, NewPermissionDeniedError("you can only delete images for your own events"))
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		HandleError(w, r, NewBadRequestError("filename is required"))
		return
	}

	path := fmt.Sprintf("images/%d/%s", eventID, filename)

	if err := h.imageService.DeleteImage(r.Context(), path); err != nil {
		handleImageServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleImageServiceError(w http.ResponseWriter, r *http.Request, err error) {
	HandleError(w, r, err)
}
