package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

type TemplateEditorHandlers struct {
	editorService templates.EditorService
	templates     *template.Template
}

func NewTemplateEditorHandlers(editorService templates.EditorService) *TemplateEditorHandlers {
	return &TemplateEditorHandlers{
		editorService: editorService,
	}
}

func (h *TemplateEditorHandlers) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *TemplateEditorHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/templates/{id}/components", func(r chi.Router) {
		r.Get("/", h.GetComponents)
		r.Put("/", h.UpdateComponents)
		r.Post("/preview", h.PreviewComponents)
		r.Get("/validate", h.ValidateComponents)
	})

	r.Get("/templates/{id}/edit", h.GetEditorPage)
}

type UpdateComponentsRequest struct {
	Components []models.Component `json:"components"`
}

type PreviewComponentsRequest struct {
	Updates   []templates.ComponentUpdate `json:"updates"`
	Additions []models.Component          `json:"additions"`
	Removals  []string                    `json:"removals"`
}

type ComponentConfigResponse struct {
	Template        *TemplateResponse              `json:"template"`
	ComponentConfig *models.ComponentConfiguration `json:"component_config"`
}

type PreviewResponse struct {
	Preview *models.ComponentConfiguration `json:"preview"`
}

type ValidationResponse struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

func (h *TemplateEditorHandlers) GetComponents(w http.ResponseWriter, r *http.Request) {
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

	editable, err := h.editorService.GetEditableTemplate(r.Context(), id)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	response := ComponentConfigResponse{
		Template:        toTemplateResponse(editable.Template),
		ComponentConfig: editable.ComponentConfig,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *TemplateEditorHandlers) UpdateComponents(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateComponentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	if err := h.editorService.UpdateComponents(r.Context(), id, req.Components); err != nil {
		HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "components updated successfully",
	})
}

func (h *TemplateEditorHandlers) PreviewComponents(w http.ResponseWriter, r *http.Request) {
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

	var req PreviewComponentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	changes := &templates.ComponentChanges{
		Updates:   req.Updates,
		Additions: req.Additions,
		Removals:  req.Removals,
	}

	preview, err := h.editorService.PreviewChanges(r.Context(), id, changes)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	response := PreviewResponse{
		Preview: preview,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *TemplateEditorHandlers) ValidateComponents(w http.ResponseWriter, r *http.Request) {
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

	editable, err := h.editorService.GetEditableTemplate(r.Context(), id)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	if editable.ComponentConfig == nil {
		respondJSON(w, http.StatusOK, ValidationResponse{
			Valid:  true,
			Errors: []string{},
		})
		return
	}

	errors := h.validateComponentConfiguration(editable.ComponentConfig)

	response := ValidationResponse{
		Valid:  len(errors) == 0,
		Errors: errors,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *TemplateEditorHandlers) validateComponentConfiguration(config *models.ComponentConfiguration) []string {
	var errors []string

	if config.Version == "" {
		errors = append(errors, "version is required")
	}

	if len(config.Components) > 50 {
		errors = append(errors, "maximum 50 components allowed")
	}

	componentIDs := make(map[string]bool)
	for i, component := range config.Components {
		if component.ID == "" {
			errors = append(errors, fmt.Sprintf("component[%d]: ID is required", i))
		}

		if componentIDs[component.ID] {
			errors = append(errors, fmt.Sprintf("component[%d]: duplicate ID %s", i, component.ID))
		}
		componentIDs[component.ID] = true

		if !component.Type.IsValid() {
			errors = append(errors, fmt.Sprintf("component[%d]: invalid type %s", i, component.Type))
		}

		if !component.Position.Mode.IsValid() {
			errors = append(errors, fmt.Sprintf("component[%d]: invalid position mode %s", i, component.Position.Mode))
		}
	}

	return errors
}

func (h *TemplateEditorHandlers) GetEditorPage(w http.ResponseWriter, r *http.Request) {
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

	editable, err := h.editorService.GetEditableTemplate(r.Context(), id)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	data := struct {
		Template *models.Template
		User     *models.User
	}{
		Template: editable.Template,
		User:     user,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *TemplateEditorHandlers) renderPage(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "template_editor.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Template Editor - TinyRSVP</title>
</head>
<body>
    <h1>Template Editor</h1>
    <p>Template engine not initialized</p>
</body>
</html>`)
}

func parseTemplateIDFromPath(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid template ID format")
	}

	if id <= 0 {
		return 0, fmt.Errorf("template ID must be positive")
	}

	return id, nil
}
