package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lenaxia/tinyrsvp/internal/invites"
)

type CleanupHandler struct {
	service invites.InviteService
}

func NewCleanupHandler(service invites.InviteService) *CleanupHandler {
	return &CleanupHandler{
		service: service,
	}
}

func (h *CleanupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.service.CleanupExpiredTokens(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": count,
		"message": "Expired tokens cleaned up successfully",
	})
}
