package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/invites"
)

type InviteWebHandlers struct {
	service   invites.InviteService
	eventRepo repositories.EventRepository
	templates *template.Template
}

func NewInviteWebHandlers(service invites.InviteService, eventRepo repositories.EventRepository) *InviteWebHandlers {
	return &InviteWebHandlers{
		service:   service,
		eventRepo: eventRepo,
	}
}

func (h *InviteWebHandlers) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *InviteWebHandlers) ListInvitesPage(w http.ResponseWriter, r *http.Request) {
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
		HandleError(w, r, err)
		return
	}

	if !user.IsAdmin() && event.CreatedBy != user.ID {
		HandleError(w, r, NewPermissionDeniedError("permission denied"))
		return
	}

	query := r.URL.Query()

	limit := 50
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	req := &invites.ListInvitesRequest{
		EventID: eventID,
		Limit:   limit,
		Offset:  (page - 1) * limit,
	}

	if status := query.Get("status"); status != "" && status != "all" {
		req.Status = &status
	}

	if search := query.Get("search"); search != "" {
		req.Search = &search
	}

	if sortBy := query.Get("sort_by"); sortBy != "" {
		req.SortBy = &sortBy
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		req.SortOrder = &sortOrder
	}

	resp, err := h.service.ListInvites(r.Context(), req)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	filter := query.Get("status")
	if filter == "" {
		filter = "all"
	}

	data := map[string]interface{}{
		"EventID":    eventID,
		"EventTitle": event.Title,
		"Invites":    resp.Invites,
		"Total":      resp.Total,
		"Stats":      resp.Stats,
		"Filter":     filter,
		"Search":     query.Get("search"),
		"Page":       page,
	}

	if h.templates == nil {
		HandleError(w, r, NewInternalError())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.Execute(w, data); err != nil {
		HandleError(w, r, err)
		return
	}
}
