package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EventCustomizationHandlers struct {
	service events.CustomizationService
}

func NewEventCustomizationHandlers(service events.CustomizationService) *EventCustomizationHandlers {
	return &EventCustomizationHandlers{
		service: service,
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
