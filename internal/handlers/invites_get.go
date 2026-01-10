package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type GetInviteService interface {
	GetInviteByID(ctx context.Context, id int64) (*models.Invite, error)
}

type GetInviteHandlers struct {
	service   GetInviteService
	eventRepo repositories.EventRepository
}

func NewGetInviteHandlers(service GetInviteService, eventRepo repositories.EventRepository) *GetInviteHandlers {
	return &GetInviteHandlers{
		service:   service,
		eventRepo: eventRepo,
	}
}

func (h *GetInviteHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/invites/{inviteId}", func(r chi.Router) {
		r.Get("/", h.GetInvite)
	})
}

func (h *GetInviteHandlers) GetInvite(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	inviteIDStr := chi.URLParam(r, "inviteId")
	inviteID, err := strconv.ParseInt(inviteIDStr, 10, 64)
	if err != nil || inviteID <= 0 {
		HandleError(w, r, NewBadRequestError("invalid invite ID"))
		return
	}

	invite, err := h.service.GetInviteByID(r.Context(), inviteID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			HandleError(w, r, NewNotFoundError("invite not found"))
			return
		}
		HandleError(w, r, &APIError{StatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "failed to retrieve invite"})
		return
	}

	event, err := h.eventRepo.GetByID(r.Context(), invite.EventID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			HandleError(w, r, NewNotFoundError("event not found"))
			return
		}
		HandleError(w, r, &APIError{StatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "failed to retrieve event"})
		return
	}

	if !user.IsAdmin() && event.CreatedBy != user.ID {
		HandleError(w, r, NewPermissionDeniedError("permission denied"))
		return
	}

	response := toInviteResponse(invite)
	respondJSON(w, http.StatusOK, response)
}
