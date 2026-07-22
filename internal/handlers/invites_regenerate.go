package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type RegenerateInviteTokenService interface {
	RegenerateToken(ctx context.Context, inviteID int64) (*invites.RegenerateTokenResponse, error)
	GetInviteByID(ctx context.Context, id int64) (*models.Invite, error)
}

type RegenerateInviteTokenHandlers struct {
	service   RegenerateInviteTokenService
	eventRepo repositories.EventRepository
}

func NewRegenerateInviteTokenHandlers(service RegenerateInviteTokenService, eventRepo repositories.EventRepository) *RegenerateInviteTokenHandlers {
	return &RegenerateInviteTokenHandlers{
		service:   service,
		eventRepo: eventRepo,
	}
}

func (h *RegenerateInviteTokenHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/invites/{inviteId}", func(r chi.Router) {
		r.Post("/regenerate", h.RegenerateInviteToken)
	})
}

func (h *RegenerateInviteTokenHandlers) RegenerateInviteToken(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.service.RegenerateToken(r.Context(), inviteID)
	if err != nil {
		var valErr *models.ValidationError
		if errors.As(err, &valErr) {
			HandleError(w, r, NewBadRequestError(err.Error()))
			return
		}
		HandleError(w, r, &APIError{StatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "failed to regenerate token"})
		return
	}

	respondJSON(w, http.StatusOK, result)
}
