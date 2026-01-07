package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type ManualInviteHandlers struct {
	service   invites.InviteService
	eventRepo repositories.EventRepository
	baseURL   string
}

func NewManualInviteHandlers(service invites.InviteService, eventRepo repositories.EventRepository, baseURL string) *ManualInviteHandlers {
	return &ManualInviteHandlers{
		service:   service,
		eventRepo: eventRepo,
		baseURL:   baseURL,
	}
}

func (h *ManualInviteHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/events/{eventId}/invites", func(r chi.Router) {
		r.Post("/manual", h.CreateManualInvite)
	})
}

type CreateManualInviteRequest struct {
	Name        *string `json:"name,omitempty"`
	MaxPlusOnes *int    `json:"max_plus_ones,omitempty"`
}

func (h *ManualInviteHandlers) CreateManualInvite(w http.ResponseWriter, r *http.Request) {
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

	if event.Status == models.EventStatusCancelled {
		respondError(w, http.StatusBadRequest, "cannot create invite for cancelled event")
		return
	}

	if event.Status == models.EventStatusArchived {
		respondError(w, http.StatusBadRequest, "cannot create invite for archived event")
		return
	}

	if !user.IsAdmin() && event.CreatedBy != user.ID {
		respondError(w, http.StatusForbidden, "permission denied")
		return
	}

	var req CreateManualInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	expiresAt := event.StartTime.Add(30 * 24 * time.Hour)
	serviceReq := &invites.CreateManualInviteRequest{
		EventID:     eventID,
		Name:        req.Name,
		MaxPlusOnes: req.MaxPlusOnes,
	}

	resp, err := h.service.CreateManualInvite(r.Context(), serviceReq, expiresAt)
	if err != nil {
		handleInviteServiceError(w, err)
		return
	}

	rsvpURL := fmt.Sprintf("%s%s", h.baseURL, resp.RSVPURL)

	response := CreateInviteResponse{
		Invite:  toInviteResponse(resp.Invite),
		Token:   resp.Token,
		RSVPURL: rsvpURL,
	}

	respondJSON(w, http.StatusCreated, response)
}
