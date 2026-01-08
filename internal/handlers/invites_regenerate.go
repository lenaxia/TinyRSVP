package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	inviteIDStr := chi.URLParam(r, "inviteId")
	inviteID, err := strconv.ParseInt(inviteIDStr, 10, 64)
	if err != nil || inviteID <= 0 {
		respondError(w, http.StatusBadRequest, "invalid invite ID")
		return
	}

	invite, err := h.service.GetInviteByID(r.Context(), inviteID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			respondError(w, http.StatusNotFound, "invite not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to retrieve invite")
		return
	}

	event, err := h.eventRepo.GetByID(r.Context(), invite.EventID)
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

	result, err := h.service.RegenerateToken(r.Context(), inviteID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "cannot regenerate token for revoked invite") ||
			strings.Contains(errMsg, "cannot regenerate token for responded invite") {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to regenerate token")
		return
	}

	respondJSON(w, http.StatusOK, result)
}
