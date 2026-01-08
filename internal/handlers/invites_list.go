package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type ListInviteHandlers struct {
	service   invites.InviteService
	eventRepo repositories.EventRepository
}

func NewListInviteHandlers(service invites.InviteService, eventRepo repositories.EventRepository) *ListInviteHandlers {
	return &ListInviteHandlers{
		service:   service,
		eventRepo: eventRepo,
	}
}

func (h *ListInviteHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/events/{eventId}/invites", func(r chi.Router) {
		r.Get("/", h.ListInvites)
	})
}

func (h *ListInviteHandlers) ListInvites(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	eventIDStr := chi.URLParam(r, "eventId")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		respondError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	event, err := h.eventRepo.GetByID(r.Context(), eventID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			respondError(w, http.StatusNotFound, "event not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to retrieve event")
		return
	}

	if !user.IsAdmin() && event.CreatedBy != user.ID {
		respondError(w, http.StatusForbidden, "permission denied")
		return
	}

	query := r.URL.Query()

	limit := 50
	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		if parsedLimit < 1 || parsedLimit > 100 {
			respondError(w, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid offset parameter")
			return
		}
		if parsedOffset < 0 {
			respondError(w, http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = parsedOffset
	}

	req := &invites.ListInvitesRequest{
		EventID: eventID,
		Limit:   limit,
		Offset:  offset,
	}

	if status := query.Get("status"); status != "" {
		req.Status = &status
	}

	if unsubscribed := query.Get("unsubscribed"); unsubscribed != "" {
		val := unsubscribed == "true"
		req.Unsubscribed = &val
	}

	if emailInvalid := query.Get("email_invalid"); emailInvalid != "" {
		val := emailInvalid == "true"
		req.EmailInvalid = &val
	}

	if search := query.Get("search"); search != "" {
		req.Search = &search
	}

	if sortBy := query.Get("sort_by"); sortBy != "" {
		req.SortBy = &sortBy
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		req.SortOrder = &sortOrder
	}

	resp, err := h.service.ListInvites(r.Context(), req)
	if err != nil {
		var validationErr *models.ValidationError
		if errors.As(err, &validationErr) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to list invites")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}
