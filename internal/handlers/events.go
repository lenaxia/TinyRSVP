package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EventHandlers struct {
	service events.Service
}

func NewEventHandlers(service events.Service) *EventHandlers {
	return &EventHandlers{
		service: service,
	}
}

func (h *EventHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/events", func(r chi.Router) {
		r.Post("/", h.CreateEvent)
		r.Get("/", h.ListEvents)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetEvent)
			r.Put("/", h.UpdateEvent)
			r.Delete("/", h.DeleteEvent)
			r.Post("/publish", h.PublishEvent)
			r.Post("/cancel", h.CancelEvent)
		})
	})
}

type CreateEventRequest struct {
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Timezone     string     `json:"timezone"`
	Location     *string    `json:"location,omitempty"`
	MaxPlusOnes  int        `json:"max_plus_ones"`
	RSVPDeadline *time.Time `json:"rsvp_deadline,omitempty"`
}

type UpdateEventRequest struct {
	Title        *string    `json:"title,omitempty"`
	Description  *string    `json:"description,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Timezone     *string    `json:"timezone,omitempty"`
	Location     *string    `json:"location,omitempty"`
	MaxPlusOnes  *int       `json:"max_plus_ones,omitempty"`
	RSVPDeadline *time.Time `json:"rsvp_deadline,omitempty"`
	Version      int        `json:"version"`
}

type CancelEventRequest struct {
	Reason string `json:"reason"`
}

type EventResponse struct {
	ID           int64              `json:"id"`
	Title        string             `json:"title"`
	Description  *string            `json:"description,omitempty"`
	StartTime    time.Time          `json:"start_time"`
	EndTime      *time.Time         `json:"end_time,omitempty"`
	Timezone     string             `json:"timezone"`
	Location     *string            `json:"location,omitempty"`
	Status       models.EventStatus `json:"status"`
	CreatedBy    int64              `json:"created_by"`
	Version      int                `json:"version"`
	MaxPlusOnes  int                `json:"max_plus_ones"`
	RSVPDeadline *time.Time         `json:"rsvp_deadline,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type ListEventsResponse struct {
	Events []*EventResponse `json:"events"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

func (h *EventHandlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	if err := validateCreateEventRequest(&req); err != nil {
		HandleError(w, r, NewBadRequestError(err.Error()))
		return
	}

	event := &models.Event{
		Title:        req.Title,
		Description:  req.Description,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Timezone:     req.Timezone,
		Location:     req.Location,
		MaxPlusOnes:  req.MaxPlusOnes,
		RSVPDeadline: req.RSVPDeadline,
	}

	if err := h.service.CreateEvent(r.Context(), event); err != nil {
		handleServiceError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, toEventResponse(event))
}

func (h *EventHandlers) GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	event, err := h.service.GetEvent(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, toEventResponse(event))
}

func (h *EventHandlers) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	if err := validateUpdateEventRequest(&req); err != nil {
		HandleError(w, r, NewBadRequestError(err.Error()))
		return
	}

	existing, err := h.service.GetEvent(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	event := &models.Event{
		ID:           id,
		Version:      req.Version,
		Title:        existing.Title,
		Description:  existing.Description,
		StartTime:    existing.StartTime,
		EndTime:      existing.EndTime,
		Timezone:     existing.Timezone,
		Location:     existing.Location,
		MaxPlusOnes:  existing.MaxPlusOnes,
		RSVPDeadline: existing.RSVPDeadline,
	}

	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Description != nil {
		event.Description = req.Description
	}
	if req.StartTime != nil {
		event.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		event.EndTime = req.EndTime
	}
	if req.Timezone != nil {
		event.Timezone = *req.Timezone
	}
	if req.Location != nil {
		event.Location = req.Location
	}
	if req.MaxPlusOnes != nil {
		event.MaxPlusOnes = *req.MaxPlusOnes
	}
	if req.RSVPDeadline != nil {
		event.RSVPDeadline = req.RSVPDeadline
	}

	if err := h.service.UpdateEvent(r.Context(), event); err != nil {
		handleServiceError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, toEventResponse(event))
}

func (h *EventHandlers) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	if err := h.service.DeleteEvent(r.Context(), id); err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EventHandlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			HandleError(w, r, NewBadRequestError("invalid limit parameter"))
			return
		}
		limit = parsedLimit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			HandleError(w, r, NewBadRequestError("invalid offset parameter"))
			return
		}
		offset = parsedOffset
	}

	filters := events.ListFilters{
		Limit:  limit,
		Offset: offset,
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status, err := parseEventStatus(statusStr)
		if err != nil {
			HandleError(w, r, NewBadRequestError("invalid status parameter"))
			return
		}
		filters.Status = &status
	}

	if creatorIDStr := r.URL.Query().Get("creator_id"); creatorIDStr != "" {
		creatorID, err := strconv.ParseInt(creatorIDStr, 10, 64)
		if err != nil {
			HandleError(w, r, NewBadRequestError("invalid creator_id parameter"))
			return
		}
		filters.CreatorID = &creatorID
	}

	eventList, err := h.service.ListEvents(r.Context(), filters)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	responses := make([]*EventResponse, len(eventList))
	for i, event := range eventList {
		responses[i] = toEventResponse(event)
	}

	total, err := h.service.CountEvents(r.Context(), filters)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	response := ListEventsResponse{
		Events: responses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *EventHandlers) PublishEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	if err := h.service.PublishEvent(r.Context(), id); err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EventHandlers) CancelEvent(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	var req CancelEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	if err := validateCancelEventRequest(&req); err != nil {
		HandleError(w, r, NewBadRequestError(err.Error()))
		return
	}

	if err := h.service.CancelEvent(r.Context(), id, req.Reason); err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parseEventID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid event ID format")
	}

	if id <= 0 {
		return 0, fmt.Errorf("event ID must be positive")
	}

	return id, nil
}

func parseEventStatus(statusStr string) (models.EventStatus, error) {
	switch statusStr {
	case "draft":
		return models.EventStatusDraft, nil
	case "published":
		return models.EventStatusPublished, nil
	case "cancelled":
		return models.EventStatusCancelled, nil
	case "archived":
		return models.EventStatusArchived, nil
	default:
		return "", fmt.Errorf("invalid status: must be 'draft', 'published', 'cancelled', or 'archived'")
	}
}

func validateCreateEventRequest(req *CreateEventRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(req.Title) < 3 || len(req.Title) > 200 {
		return fmt.Errorf("title must be between 3 and 200 characters")
	}
	if req.StartTime.IsZero() {
		return fmt.Errorf("start_time is required")
	}
	if req.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}
	if req.MaxPlusOnes < 0 || req.MaxPlusOnes > 10 {
		return fmt.Errorf("max_plus_ones must be between 0 and 10")
	}
	if req.Description != nil && len(*req.Description) > 5000 {
		return fmt.Errorf("description must not exceed 5000 characters")
	}
	if req.Location != nil && len(*req.Location) > 500 {
		return fmt.Errorf("location must not exceed 500 characters")
	}
	return nil
}

func validateUpdateEventRequest(req *UpdateEventRequest) error {
	if req.Version == 0 {
		return fmt.Errorf("version is required")
	}
	if req.Title != nil && (len(*req.Title) < 3 || len(*req.Title) > 200) {
		return fmt.Errorf("title must be between 3 and 200 characters")
	}
	if req.MaxPlusOnes != nil && (*req.MaxPlusOnes < 0 || *req.MaxPlusOnes > 10) {
		return fmt.Errorf("max_plus_ones must be between 0 and 10")
	}
	if req.Description != nil && len(*req.Description) > 5000 {
		return fmt.Errorf("description must not exceed 5000 characters")
	}
	if req.Location != nil && len(*req.Location) > 500 {
		return fmt.Errorf("location must not exceed 500 characters")
	}
	return nil
}

func validateCancelEventRequest(req *CancelEventRequest) error {
	if req.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	if len(req.Reason) < 10 || len(req.Reason) > 500 {
		return fmt.Errorf("reason must be between 10 and 500 characters")
	}
	return nil
}

func toEventResponse(event *models.Event) *EventResponse {
	return &EventResponse{
		ID:           event.ID,
		Title:        event.Title,
		Description:  event.Description,
		StartTime:    event.StartTime,
		EndTime:      event.EndTime,
		Timezone:     event.Timezone,
		Location:     event.Location,
		Status:       event.Status,
		CreatedBy:    event.CreatedBy,
		Version:      event.Version,
		MaxPlusOnes:  event.MaxPlusOnes,
		RSVPDeadline: event.RSVPDeadline,
		CreatedAt:    event.CreatedAt,
		UpdatedAt:    event.UpdatedAt,
	}
}

func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	HandleError(w, r, err)
}
