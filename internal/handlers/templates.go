package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

type TemplateHandlers struct {
	service templates.Service
}

func NewTemplateHandlers(service templates.Service) *TemplateHandlers {
	return &TemplateHandlers{
		service: service,
	}
}

func (h *TemplateHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/templates", func(r chi.Router) {
		r.Post("/", h.CreateTemplate)
		r.Get("/", h.ListTemplates)
		r.Post("/preview", h.PreviewTemplate)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetTemplate)
			r.Put("/", h.UpdateTemplate)
			r.Delete("/", h.DeleteTemplate)
			r.Post("/set-active", h.SetActive)
			r.Post("/set-default", h.SetDefault)
		})
	})
}

type CreateTemplateRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	HTMLContent string  `json:"html_content"`
	TextContent *string `json:"text_content,omitempty"`
	CSSContent  *string `json:"css_content,omitempty"`
	EventID     *int64  `json:"event_id,omitempty"`
}

type UpdateTemplateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	HTMLContent *string `json:"html_content,omitempty"`
	TextContent *string `json:"text_content,omitempty"`
	CSSContent  *string `json:"css_content,omitempty"`
}

type SetActiveRequest struct {
	Active bool `json:"active"`
}

type TemplateResponse struct {
	ID          int64   `json:"id"`
	EventID     *int64  `json:"event_id,omitempty"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	HTMLContent string  `json:"html_content"`
	TextContent *string `json:"text_content,omitempty"`
	CSSContent  *string `json:"css_content,omitempty"`
	IsDefault   bool    `json:"is_default"`
	IsActive    bool    `json:"is_active"`
	Version     int     `json:"version"`
	CreatedBy   int64   `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ListTemplatesResponse struct {
	Templates []*TemplateResponse `json:"templates"`
	Total     int                 `json:"total"`
}

func (h *TemplateHandlers) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	templateType := models.TemplateType(req.Type)
	if !templateType.IsValid() {
		respondError(w, http.StatusBadRequest, "invalid template type")
		return
	}

	template := &models.Template{
		EventID:     req.EventID,
		Name:        req.Name,
		Type:        templateType,
		Description: req.Description,
		HTMLContent: req.HTMLContent,
		TextContent: req.TextContent,
		CSSContent:  req.CSSContent,
	}

	if err := h.service.CreateTemplate(r.Context(), template); err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"template": toTemplateResponse(template),
	})
}

func (h *TemplateHandlers) GetTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	template, err := h.service.GetTemplate(r.Context(), id)
	if err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"template": toTemplateResponse(template),
	})
}

func (h *TemplateHandlers) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	var req UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := h.service.GetTemplate(r.Context(), id)
	if err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	template := &models.Template{
		ID:          id,
		EventID:     existing.EventID,
		Name:        existing.Name,
		Type:        existing.Type,
		Description: existing.Description,
		HTMLContent: existing.HTMLContent,
		TextContent: existing.TextContent,
		CSSContent:  existing.CSSContent,
		CreatedBy:   existing.CreatedBy,
	}

	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Description != nil {
		template.Description = *req.Description
	}
	if req.HTMLContent != nil {
		template.HTMLContent = *req.HTMLContent
	}
	if req.TextContent != nil {
		template.TextContent = req.TextContent
	}
	if req.CSSContent != nil {
		template.CSSContent = req.CSSContent
	}

	if err := h.service.UpdateTemplate(r.Context(), template); err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"template": toTemplateResponse(template),
	})
}

func (h *TemplateHandlers) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	if err := h.service.DeleteTemplate(r.Context(), id); err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TemplateHandlers) ListTemplates(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	filters := &repositories.TemplateFilters{
		Limit: 50,
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		templateType := models.TemplateType(typeStr)
		if !templateType.IsValid() {
			respondError(w, http.StatusBadRequest, "invalid template type")
			return
		}
		filters.Type = &templateType
	}

	if eventIDStr := r.URL.Query().Get("event_id"); eventIDStr != "" {
		eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid event_id parameter")
			return
		}
		filters.EventID = &eventID
	}

	if isDefaultStr := r.URL.Query().Get("is_default"); isDefaultStr != "" {
		isDefault := isDefaultStr == "true"
		filters.IsDefault = &isDefault
	}

	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filters.IsActive = &isActive
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			respondError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		filters.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			respondError(w, http.StatusBadRequest, "invalid offset parameter")
			return
		}
		filters.Offset = offset
	}

	templateList, err := h.service.ListTemplates(r.Context(), filters)
	if err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	responses := make([]*TemplateResponse, len(templateList))
	for i, tmpl := range templateList {
		responses[i] = toTemplateResponse(tmpl)
	}

	response := ListTemplatesResponse{
		Templates: responses,
		Total:     len(responses),
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *TemplateHandlers) SetActive(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	var req SetActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.SetActive(r.Context(), id, req.Active); err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TemplateHandlers) SetDefault(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	if err := h.service.SetDefault(r.Context(), id); err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parseTemplateID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid template ID format")
	}

	if id <= 0 {
		return 0, fmt.Errorf("template ID must be positive")
	}

	return id, nil
}

func toTemplateResponse(template *models.Template) *TemplateResponse {
	return &TemplateResponse{
		ID:          template.ID,
		EventID:     template.EventID,
		Name:        template.Name,
		Type:        string(template.Type),
		Description: template.Description,
		HTMLContent: template.HTMLContent,
		TextContent: template.TextContent,
		CSSContent:  template.CSSContent,
		IsDefault:   template.IsDefault,
		IsActive:    template.IsActive,
		Version:     template.Version,
		CreatedBy:   template.CreatedBy,
		CreatedAt:   template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   template.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func handleTemplateServiceError(w http.ResponseWriter, err error) {
	var notFoundErr *models.NotFoundError
	var forbiddenErr *models.ForbiddenError
	var unauthorizedErr *models.UnauthorizedError
	var validationErr *models.ValidationError

	switch {
	case errors.As(err, &notFoundErr):
		respondError(w, http.StatusNotFound, "template not found")
	case errors.As(err, &forbiddenErr):
		respondError(w, http.StatusForbidden, err.Error())
	case errors.As(err, &unauthorizedErr):
		respondError(w, http.StatusUnauthorized, err.Error())
	case errors.As(err, &validationErr):
		respondError(w, http.StatusBadRequest, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *TemplateHandlers) PreviewTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req templates.PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.PreviewTemplate(r.Context(), &req)
	if err != nil {
		handleTemplateServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}
