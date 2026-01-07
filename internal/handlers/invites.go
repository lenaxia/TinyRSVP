package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type InviteHandlers struct {
	service invites.IndividualInviteService
	baseURL string
}

func NewInviteHandlers(service invites.IndividualInviteService, baseURL string) *InviteHandlers {
	return &InviteHandlers{
		service: service,
		baseURL: baseURL,
	}
}

type ImportInviteHandlers struct {
	service invites.InviteService
	baseURL string
}

func NewImportInviteHandlers(service invites.InviteService, baseURL string) *ImportInviteHandlers {
	return &ImportInviteHandlers{
		service: service,
		baseURL: baseURL,
	}
}

func (h *InviteHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/events/{eventId}/invites", func(r chi.Router) {
		r.Post("/", h.CreateInvite)
	})
}

func (h *ImportInviteHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/events/{eventId}/invites", func(r chi.Router) {
		r.Post("/import", h.ImportInvites)
	})
}

type CreateInviteRequest struct {
	Email       string  `json:"email"`
	Name        *string `json:"name,omitempty"`
	MaxPlusOnes *int    `json:"max_plus_ones,omitempty"`
}

type InviteResponse struct {
	ID          int64               `json:"id"`
	EventID     int64               `json:"event_id"`
	Email       *string             `json:"email,omitempty"`
	Name        *string             `json:"name,omitempty"`
	MaxPlusOnes int                 `json:"max_plus_ones"`
	Status      models.InviteStatus `json:"status"`
	ExpiresAt   string              `json:"expires_at"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

type CreateInviteResponse struct {
	Invite  *InviteResponse `json:"invite"`
	Token   string          `json:"token"`
	RSVPURL string          `json:"rsvp_url"`
}

func (h *InviteHandlers) CreateInvite(w http.ResponseWriter, r *http.Request) {
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

	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}

	serviceReq := &invites.CreateIndividualInviteRequest{
		EventID:     eventID,
		Email:       req.Email,
		Name:        req.Name,
		MaxPlusOnes: req.MaxPlusOnes,
	}

	resp, err := h.service.CreateIndividualInvite(r.Context(), user, serviceReq)
	if err != nil {
		handleInviteServiceError(w, err)
		return
	}

	rsvpURL := fmt.Sprintf("%s/rsvp/%s", h.baseURL, resp.Token)

	response := CreateInviteResponse{
		Invite:  toInviteResponse(resp.Invite),
		Token:   resp.Token,
		RSVPURL: rsvpURL,
	}

	respondJSON(w, http.StatusCreated, response)
}

func toInviteResponse(invite *models.Invite) *InviteResponse {
	return &InviteResponse{
		ID:          invite.ID,
		EventID:     invite.EventID,
		Email:       invite.Email,
		Name:        invite.Name,
		MaxPlusOnes: invite.MaxPlusOnes,
		Status:      invite.Status,
		ExpiresAt:   invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   invite.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func handleInviteServiceError(w http.ResponseWriter, err error) {
	var notFoundErr *models.NotFoundError
	var permErr *models.PermissionDeniedError
	var validationErr *models.ValidationError
	var conflictErr *models.ConflictError

	switch {
	case errors.As(err, &notFoundErr):
		respondError(w, http.StatusNotFound, "event not found")
	case errors.As(err, &permErr):
		respondError(w, http.StatusForbidden, "permission denied")
	case errors.As(err, &validationErr):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.As(err, &conflictErr):
		respondError(w, http.StatusConflict, "email already invited to this event")
	default:
		respondError(w, http.StatusInternalServerError, "failed to create invite")
	}
}

func (h *ImportInviteHandlers) ImportInvites(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.UserFromContext(r.Context())
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

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	if header.Size > 1<<20 {
		respondError(w, http.StatusBadRequest, "file size exceeds 1MB limit")
		return
	}

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		respondError(w, http.StatusBadRequest, "file must be CSV format")
		return
	}

	csvData := make([]byte, header.Size)
	if _, err := file.Read(csvData); err != nil && err.Error() != "EOF" {
		respondError(w, http.StatusBadRequest, "failed to read file")
		return
	}

	expiresAt := time.Now().Add(60 * 24 * time.Hour)
	result, err := h.service.ImportCSV(r.Context(), eventID, csvData, 0, expiresAt)
	if err != nil {
		handleInviteServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}
