package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type RevokeInviteService interface {
	RevokeInvite(ctx context.Context, req *invites.RevokeInviteRequest) error
	GetInviteByID(ctx context.Context, id int64) (*models.Invite, error)
}

type RevokeInviteHandlers struct {
	service   RevokeInviteService
	eventRepo repositories.EventRepository
}

func NewRevokeInviteHandlers(service RevokeInviteService, eventRepo repositories.EventRepository) *RevokeInviteHandlers {
	return &RevokeInviteHandlers{
		service:   service,
		eventRepo: eventRepo,
	}
}

func (h *RevokeInviteHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/invites/{inviteId}", func(r chi.Router) {
		r.Post("/revoke", h.RevokeInvite)
	})
}

type RevokeInviteRequest struct {
	Reason *string `json:"reason,omitempty"`
}

type RevokeInviteResponse struct {
	Message string `json:"message"`
}

func (h *RevokeInviteHandlers) RevokeInvite(w http.ResponseWriter, r *http.Request) {
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

	var req RevokeInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
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

	serviceReq := &invites.RevokeInviteRequest{
		InviteID: inviteID,
		Reason:   req.Reason,
	}

	if err := h.service.RevokeInvite(r.Context(), serviceReq); err != nil {
		errMsg := err.Error()
		if errMsg == "cannot transition from responded" ||
		   errMsg == "cannot transition from revoked" {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to revoke invite")
		return
	}

	response := RevokeInviteResponse{
		Message: "Invite revoked successfully",
	}

	respondJSON(w, http.StatusOK, response)
}
