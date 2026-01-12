package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	
	r.Get("/api/themes/preview", h.HandleThemePreview)
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
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	templateType := models.TemplateType(req.Type)
	if !templateType.IsValid() {
		HandleError(w, r, NewBadRequestError("invalid template type"))
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
		handleTemplateServiceError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"template": toTemplateResponse(template),
	})
}

func (h *TemplateHandlers) GetTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid template ID"))
		return
	}

	template, err := h.service.GetTemplate(r.Context(), id)
	if err != nil {
		handleTemplateServiceError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"template": toTemplateResponse(template),
	})
}

func (h *TemplateHandlers) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid template ID"))
		return
	}

	var req UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	existing, err := h.service.GetTemplate(r.Context(), id)
	if err != nil {
		handleTemplateServiceError(w, r, err)
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
		handleTemplateServiceError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"template": toTemplateResponse(template),
	})
}

func (h *TemplateHandlers) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid template ID"))
		return
	}

	if err := h.service.DeleteTemplate(r.Context(), id); err != nil {
		handleTemplateServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TemplateHandlers) ListTemplates(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	filters := &repositories.TemplateFilters{
		Limit: 50,
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		templateType := models.TemplateType(typeStr)
		if !templateType.IsValid() {
			HandleError(w, r, NewBadRequestError("invalid template type"))
			return
		}
		filters.Type = &templateType
	}

	if eventIDStr := r.URL.Query().Get("event_id"); eventIDStr != "" {
		eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
		if err != nil {
			HandleError(w, r, NewBadRequestError("invalid event_id parameter"))
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
			HandleError(w, r, NewBadRequestError("invalid limit parameter"))
			return
		}
		filters.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			HandleError(w, r, NewBadRequestError("invalid offset parameter"))
			return
		}
		filters.Offset = offset
	}

	templateList, err := h.service.ListTemplates(r.Context(), filters)
	if err != nil {
		handleTemplateServiceError(w, r, err)
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
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid template ID"))
		return
	}

	var req SetActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	if err := h.service.SetActive(r.Context(), id, req.Active); err != nil {
		handleTemplateServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TemplateHandlers) SetDefault(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	id, err := parseTemplateID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid template ID"))
		return
	}

	if err := h.service.SetDefault(r.Context(), id); err != nil {
		handleTemplateServiceError(w, r, err)
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

func handleTemplateServiceError(w http.ResponseWriter, r *http.Request, err error) {
	HandleError(w, r, err)
}

func (h *TemplateHandlers) PreviewTemplate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user.ID == 0 {
		HandleError(w, r, NewUnauthorizedError("authentication required"))
		return
	}

	var req templates.PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	resp, err := h.service.PreviewTemplate(r.Context(), &req)
	if err != nil {
		handleTemplateServiceError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *TemplateHandlers) HandleThemePreview(w http.ResponseWriter, r *http.Request) {
	themeIDStr := r.URL.Query().Get("theme_id")
	if themeIDStr == "" {
		http.Error(w, "theme_id parameter is required", http.StatusBadRequest)
		return
	}

	themeID, err := strconv.ParseInt(themeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid theme ID", http.StatusBadRequest)
		return
	}

	template, err := h.service.GetTemplate(r.Context(), themeID)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Sample Event Title"
	}

	location := r.URL.Query().Get("location")
	if location == "" {
		location = "Sample Location"
	}

	description := r.URL.Query().Get("description")
	if description == "" {
		description = "This is a sample event description to show how your theme will look."
	}

	startTimeStr := r.URL.Query().Get("start_time")
	var startTime time.Time
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			startTime = time.Now().Add(7 * 24 * time.Hour)
		}
	} else {
		startTime = time.Now().Add(7 * 24 * time.Hour)
	}

	themeMode := r.URL.Query().Get("theme_mode")
	if themeMode == "" {
		themeMode = "light"
	}

	customImageURL := r.URL.Query().Get("custom_image_url")
	customColor := r.URL.Query().Get("custom_color")

	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'self'; base-uri 'self'; form-action 'self'")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	dataTheme := themeMode
	eventTheme := template.Category
	if eventTheme == "" {
		eventTheme = "modern"
	}

	headerImageHTML := ""
	if customImageURL != "" {
		headerImageHTML = fmt.Sprintf(`
	               <div class="event-header-image">
	                   <img src="%s" alt="Event header image" />
	               </div>`, customImageURL)
	}

	customColorCSS := ""
	if customColor != "" && isValidHexColor(customColor) {
		customColorCSS = fmt.Sprintf(`
	   <style>
	       :root {
	           --primary-color: %s;
	           --primary-color-hover: %s;
	           --primary-color-alpha: %s33;
	       }
	   </style>`, customColor, customColor, customColor)
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en" data-theme="%s" data-event-theme="%s">
<head>
	   <meta charset="UTF-8">
	   <meta name="viewport" content="width=device-width, initial-scale=1.0">
	   <title>Theme Preview</title>
	   <link rel="stylesheet" href="/static/css/variables.css">
	   <link rel="stylesheet" href="/static/css/typography.css">
	   <link rel="stylesheet" href="/static/css/colors.css">
	   <link rel="stylesheet" href="/static/css/buttons.css">
	   <link rel="stylesheet" href="/static/css/forms.css">
	   <link rel="stylesheet" href="/static/css/rsvp_page.css">
	   <style>
	       .event-header-image {
	           width: 100%%;
	           max-height: 400px;
	           overflow: hidden;
	           margin-bottom: 2rem;
	           border-radius: 8px;
	       }
	       .event-header-image img {
	           width: 100%%;
	           height: 100%%;
	           object-fit: cover;
	       }
	   </style>%s
</head>
<body>
	   <div class="rsvp-page">
	       <div class="rsvp-container">
	           <article class="event-details">%s
	               <header>
	                   <h1 class="event-title">%s</h1>
	               </header>
	               <section class="event-info">
	                   <div class="event-info-item">
	                       <div class="event-info-content">
	                           <div class="event-info-label">Date & Time</div>
	                           <time>%s</time>
	                       </div>
	                   </div>
	                   <div class="event-info-item">
	                       <div class="event-info-content">
	                           <div class="event-info-label">Location</div>
	                           <address class="event-location">%s</address>
	                       </div>
	                   </div>
	               </section>
	               <section class="event-description">
	                   <h2>About This Event</h2>
	                   <p>%s</p>
	               </section>
	           </article>
	           <form class="rsvp-form">
	               <h2 class="rsvp-form-title">Please Respond</h2>
	               <div class="form-group">
	                   <fieldset>
	                       <legend class="form-label">Will you attend?</legend>
	                       <div class="response-options">
	                           <div class="response-option">
	                               <input type="radio" name="response" value="yes" id="response_yes">
	                               <label for="response_yes">Yes, I'll be there</label>
	                           </div>
	                           <div class="response-option">
	                               <input type="radio" name="response" value="maybe" id="response_maybe">
	                               <label for="response_maybe">Maybe</label>
	                           </div>
	                           <div class="response-option">
	                               <input type="radio" name="response" value="no" id="response_no">
	                               <label for="response_no">No, I can't make it</label>
	                           </div>
	                       </div>
	                   </fieldset>
	               </div>
	               <div class="rsvp-actions">
	                   <button type="button" class="btn btn-primary" disabled>Submit RSVP (Preview Only)</button>
	               </div>
	           </form>
	       </div>
	   </div>
</body>
</html>`, dataTheme, eventTheme, customColorCSS, headerImageHTML, title, startTime.Format("Monday, January 2, 2006 at 3:04 PM MST"), location, description)
}

func isValidHexColor(color string) bool {
	if len(color) != 7 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
