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

type SendInviteService interface {
	GetInviteByID(ctx context.Context, id int64) (*models.Invite, error)
	SendInvite(ctx context.Context, req *invites.SendInviteRequest, emailRepo repositories.EmailQueueRepository) error
}

type SendInviteHandlers struct {
	service   SendInviteService
	eventRepo repositories.EventRepository
	emailRepo repositories.EmailQueueRepository
	baseURL   string
}

func NewSendInviteHandlers(service SendInviteService, eventRepo repositories.EventRepository, emailRepo repositories.EmailQueueRepository, baseURL string) *SendInviteHandlers {
	return &SendInviteHandlers{
		service:   service,
		eventRepo: eventRepo,
		emailRepo: emailRepo,
		baseURL:   baseURL,
	}
}

func (h *SendInviteHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/invites/{inviteId}", func(r chi.Router) {
		r.Post("/send", h.SendInvite)
	})
}

type SendInviteResponse struct {
	Message string `json:"message"`
}

func (h *SendInviteHandlers) SendInvite(w http.ResponseWriter, r *http.Request) {
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

	serviceReq := &invites.SendInviteRequest{
		InviteID: inviteID,
		BaseURL:  h.baseURL,
		Event:    event,
	}

	if err := h.service.SendInvite(r.Context(), serviceReq, h.emailRepo); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "invite has no email address") ||
			strings.Contains(errMsg, "cannot send revoked invite") {
			HandleError(w, r, NewBadRequestError(err.Error()))
			return
		}
		HandleError(w, r, &APIError{StatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "failed to send invite"})
		return
	}

	response := SendInviteResponse{
		Message: "Invite email queued successfully",
	}

	respondJSON(w, http.StatusOK, response)
}
