package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EventServiceForImages interface {
	GetEvent(ctx context.Context, id int64) (*models.Event, error)
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
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	eventID, err := parseEventID(chi.URLParam(r, "event_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	event, err := h.eventService.GetEvent(r.Context(), eventID)
	if err != nil {
		handleImageServiceError(w, err)
		return
	}

	if !h.authz.CanEditEvent(r.Context(), user, event) {
		respondError(w, http.StatusForbidden, "you can only upload images for your own events")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	metadata, err := h.imageService.UploadImage(r.Context(), eventID, header.Filename, file)
	if err != nil {
		handleImageServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, ImageUploadResponse{
		Image: metadata,
	})
}

func (h *ImageHandlers) DeleteImage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	eventID, err := parseEventID(chi.URLParam(r, "event_id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	event, err := h.eventService.GetEvent(r.Context(), eventID)
	if err != nil {
		handleImageServiceError(w, err)
		return
	}

	if !h.authz.CanEditEvent(r.Context(), user, event) {
		respondError(w, http.StatusForbidden, "you can only delete images for your own events")
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		respondError(w, http.StatusBadRequest, "filename is required")
		return
	}

	path := fmt.Sprintf("images/%d/%s", eventID, filename)

	if err := h.imageService.DeleteImage(r.Context(), path); err != nil {
		handleImageServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleImageServiceError(w http.ResponseWriter, err error) {
	var notFoundErr *models.NotFoundError
	var validationErr *assets.ValidationError

	switch {
	case errors.As(err, &notFoundErr):
		respondError(w, http.StatusNotFound, "event not found")
	case errors.As(err, &validationErr):
		respondError(w, http.StatusBadRequest, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
