package handlers

import (
	"context"
	"encoding/json"
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

type UpdateInviteService interface {
	GetInviteByID(ctx context.Context, id int64) (*models.Invite, error)
	UpdateInvite(ctx context.Context, req *invites.UpdateInviteRequest) error
}

type UpdateInviteHandlers struct {
	service   UpdateInviteService
	eventRepo repositories.EventRepository
}

func NewUpdateInviteHandlers(service UpdateInviteService, eventRepo repositories.EventRepository) *UpdateInviteHandlers {
	return &UpdateInviteHandlers{
		service:   service,
		eventRepo: eventRepo,
	}
}

func (h *UpdateInviteHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/invites/{inviteId}", func(r chi.Router) {
		r.Put("/", h.UpdateInvite)
	})
}

type UpdateInviteRequestBody struct {
	Name        *string `json:"name,omitempty"`
	MaxPlusOnes *int    `json:"max_plus_ones,omitempty"`
}

type UpdateInviteResponse struct {
	Message string `json:"message"`
}

func (h *UpdateInviteHandlers) UpdateInvite(w http.ResponseWriter, r *http.Request) {
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

	var reqBody UpdateInviteRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
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

	serviceReq := &invites.UpdateInviteRequest{
		InviteID:    inviteID,
		Name:        reqBody.Name,
		MaxPlusOnes: reqBody.MaxPlusOnes,
	}

	if err := h.service.UpdateInvite(r.Context(), serviceReq); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "cannot update revoked invite") ||
			strings.Contains(errMsg, "cannot update responded invite") {
			HandleError(w, r, NewBadRequestError(err.Error()))
			return
		}
		var validationErr *models.ValidationError
		if errors.As(err, &validationErr) {
			HandleError(w, r, NewBadRequestError(err.Error()))
			return
		}
		HandleError(w, r, &APIError{StatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "failed to update invite"})
		return
	}

	response := UpdateInviteResponse{
		Message: "Invite updated successfully",
	}

	respondJSON(w, http.StatusOK, response)
}
