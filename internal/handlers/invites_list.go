package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
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
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	eventIDStr := chi.URLParam(r, "eventId")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	event, err := h.eventRepo.GetByID(r.Context(), eventID)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	if !user.IsAdmin() && event.CreatedBy != user.ID {
		HandleError(w, r, NewPermissionDeniedError("permission denied"))
		return
	}

	query := r.URL.Query()

	limit := 50
	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			HandleError(w, r, NewBadRequestError("invalid limit parameter"))
			return
		}
		if parsedLimit < 1 || parsedLimit > 100 {
			HandleError(w, r, NewBadRequestError("limit must be between 1 and 100"))
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			HandleError(w, r, NewBadRequestError("invalid offset parameter"))
			return
		}
		if parsedOffset < 0 {
			HandleError(w, r, NewBadRequestError("offset must be non-negative"))
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
		HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}
