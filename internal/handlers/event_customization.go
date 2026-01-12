package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EventCustomizationHandlers struct {
	service   events.CustomizationService
	templates *template.Template
}

func NewEventCustomizationHandlers(service events.CustomizationService) *EventCustomizationHandlers {
	return &EventCustomizationHandlers{
		service: service,
	}
}

func (h *EventCustomizationHandlers) SetTemplates(templates *template.Template) {
	h.templates = templates
}

func (h *EventCustomizationHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/events/{id}/template/customization", func(r chi.Router) {
		r.Get("/", h.GetCustomization)
		r.Put("/", h.UpdateCustomization)
		r.Post("/preview", h.PreviewCustomization)
		r.Delete("/", h.ResetCustomization)
	})
}

func (h *EventCustomizationHandlers) CustomizationPage(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	data := map[string]interface{}{
		"EventID": eventID,
	}

	if h.templates != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.templates.ExecuteTemplate(w, "event_customization.html", data); err != nil {
			HandleError(w, r, err)
			return
		}
	} else {
		http.Error(w, "Templates not configured", http.StatusInternalServerError)
	}
}

func (h *EventCustomizationHandlers) GetCustomization(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	data, err := h.service.GetEventCustomization(r.Context(), eventID)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func (h *EventCustomizationHandlers) UpdateCustomization(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	var overrides models.ComponentOverrides
	if err := json.NewDecoder(r.Body).Decode(&overrides); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	if err := h.service.UpdateEventCustomization(r.Context(), eventID, &overrides); err != nil {
		HandleError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Customization updated successfully",
	})
}

func (h *EventCustomizationHandlers) PreviewCustomization(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	var overrides models.ComponentOverrides
	if err := json.NewDecoder(r.Body).Decode(&overrides); err != nil {
		HandleError(w, r, NewBadRequestError("invalid request body"))
		return
	}

	config, err := h.service.PreviewEventCustomization(r.Context(), eventID, &overrides)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

func (h *EventCustomizationHandlers) ResetCustomization(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	if err := h.service.ResetEventCustomization(r.Context(), eventID); err != nil {
		HandleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
