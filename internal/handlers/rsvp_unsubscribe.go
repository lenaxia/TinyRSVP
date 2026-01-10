package handlers

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type UnsubscribeInviteService interface {
	GetInviteByToken(ctx context.Context, token string) (*models.Invite, error)
	UnsubscribeFromReminders(ctx context.Context, token string) error
}

type UnsubscribeHandler struct {
	inviteService UnsubscribeInviteService
	eventRepo     repositories.EventRepository
	templates     *template.Template
}

func NewUnsubscribeHandler(
	inviteService UnsubscribeInviteService,
	eventRepo repositories.EventRepository,
) *UnsubscribeHandler {
	return &UnsubscribeHandler{
		inviteService: inviteService,
		eventRepo:     eventRepo,
	}
}

func (h *UnsubscribeHandler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

type UnsubscribePageData struct {
	Event        *models.Event
	Invite       *models.Invite
	ErrorMessage string
	Success      bool
}

func (h *UnsubscribeHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.renderError(w, http.StatusNotFound, "Invalid unsubscribe link")
		return
	}

	invite, err := h.inviteService.GetInviteByToken(r.Context(), token)
	if err != nil {
		h.handleInviteError(w, err)
		return
	}

	event, err := h.eventRepo.GetByID(r.Context(), invite.EventID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			h.renderError(w, http.StatusNotFound, "Event not found")
			return
		}
		h.renderError(w, http.StatusInternalServerError, "Failed to load event")
		return
	}

	if err := h.inviteService.UnsubscribeFromReminders(r.Context(), token); err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to unsubscribe")
		return
	}

	data := &UnsubscribePageData{
		Event:   event,
		Invite:  invite,
		Success: true,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *UnsubscribeHandler) handleInviteError(w http.ResponseWriter, err error) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "expired") {
		h.renderError(w, http.StatusGone, "This invite has expired")
		return
	}
	if strings.Contains(errMsg, "revoked") {
		h.renderError(w, http.StatusForbidden, "This invite has been revoked")
		return
	}
	
	var notFoundErr *models.NotFoundError
	if errors.As(err, &notFoundErr) {
		h.renderError(w, http.StatusNotFound, "Invite not found")
		return
	}
	
	h.renderError(w, http.StatusNotFound, "Invite not found or has been revoked")
}

func (h *UnsubscribeHandler) renderError(w http.ResponseWriter, status int, message string) {
	data := &UnsubscribePageData{
		ErrorMessage: message,
		Success:      false,
	}
	h.renderPage(w, status, data)
}

func (h *UnsubscribeHandler) renderPage(w http.ResponseWriter, status int, data *UnsubscribePageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "unsubscribe.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Unsubscribe</title>
</head>
<body>
    <h1>Unsubscribe from Reminders</h1>
    %s
</body>
</html>`, func() string {
		if data.ErrorMessage != "" {
			return fmt.Sprintf("<p>Error: %s</p>", data.ErrorMessage)
		}
		if data.Success && data.Event != nil {
			return fmt.Sprintf("<p>You have been unsubscribed from reminders for %s.</p>", data.Event.Title)
		}
		return "<p>Processing...</p>"
	}())
}
