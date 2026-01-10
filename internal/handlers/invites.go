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
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
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
	service   invites.InviteService
	eventRepo repositories.EventRepository
	baseURL   string
}

func NewImportInviteHandlers(service invites.InviteService, eventRepo repositories.EventRepository, baseURL string) *ImportInviteHandlers {
	return &ImportInviteHandlers{
		service:   service,
		eventRepo: eventRepo,
		baseURL:   baseURL,
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
	Token       *string             `json:"token,omitempty"`
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
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	eventIDStr := chi.URLParam(r, "eventId")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	if req.Email == "" {
		HandleError(w, r, NewBadRequestError("email is required"))
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
		handleInviteServiceError(w, r, err)
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
		Token:       invite.Token,
		MaxPlusOnes: invite.MaxPlusOnes,
		Status:      invite.Status,
		ExpiresAt:   invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   invite.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func handleInviteServiceError(w http.ResponseWriter, r *http.Request, err error) {
	HandleError(w, r, err)
}

func (h *ImportInviteHandlers) ImportInvites(w http.ResponseWriter, r *http.Request) {
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
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			HandleError(w, r, NewNotFoundError("event not found"))
			return
		}
		HandleError(w, r, &APIError{StatusCode: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "failed to retrieve event"})
		return
	}

	if event.Status == models.EventStatusCancelled {
		HandleError(w, r, NewBadRequestError("cannot create invite for cancelled event"))
		return
	}

	if event.Status == models.EventStatusArchived {
		HandleError(w, r, NewBadRequestError("cannot create invite for archived event"))
		return
	}

	if !user.IsAdmin() && event.CreatedBy != user.ID {
		HandleError(w, r, NewPermissionDeniedError("permission denied"))
		return
	}

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		HandleError(w, r, NewBadRequestError("failed to parse multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		HandleError(w, r, NewBadRequestError("file is required"))
		return
	}
	defer file.Close()

	if header.Size > 1<<20 {
		HandleError(w, r, NewBadRequestError("file size exceeds 1MB limit"))
		return
	}

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		HandleError(w, r, NewBadRequestError("file must be CSV format"))
		return
	}

	csvData := make([]byte, header.Size)
	if _, err := file.Read(csvData); err != nil && err.Error() != "EOF" {
		HandleError(w, r, NewBadRequestError("failed to read file"))
		return
	}

	expiresAt := event.StartTime.Add(30 * 24 * time.Hour)
	result, err := h.service.ImportCSV(r.Context(), eventID, csvData, event.MaxPlusOnes, expiresAt)
	if err != nil {
		handleInviteServiceError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}
