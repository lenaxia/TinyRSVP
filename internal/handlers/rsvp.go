package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/rsvp"
)

type RSVPInviteService interface {
	GetInviteByToken(ctx context.Context, token string) (*models.Invite, error)
	MarkInviteViewed(ctx context.Context, inviteID int64) error
}

type RSVPService interface {
	SubmitRSVP(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error)
	UpdateRSVP(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error)
}

type RSVPHandler struct {
	inviteService RSVPInviteService
	eventRepo     repositories.EventRepository
	rsvpRepo      repositories.RSVPRepository
	questionRepo  repositories.QuestionRepository
	rsvpService   RSVPService
	templates     *template.Template
}

func NewRSVPHandler(
	inviteService RSVPInviteService,
	eventRepo repositories.EventRepository,
	rsvpRepo repositories.RSVPRepository,
	questionRepo repositories.QuestionRepository,
) *RSVPHandler {
	return &RSVPHandler{
		inviteService: inviteService,
		eventRepo:     eventRepo,
		rsvpRepo:      rsvpRepo,
		questionRepo:  questionRepo,
	}
}

func (h *RSVPHandler) SetRSVPService(service RSVPService) {
	h.rsvpService = service
}

type QuestionWithOptions struct {
	*models.PreferenceQuestion
	ParsedOptions []string
}

type RSVPPageData struct {
	Event          *models.Event
	Invite         *models.Invite
	ExistingRSVP   *models.RSVP
	Questions      []*QuestionWithOptions
	Token          string
	DeadlinePassed bool
	EventPassed    bool
	LocalStartTime string
	LocalEndTime   string
	TimeUntilEvent string
	CanUpdate      bool
	ErrorMessage   string
}

func (h *RSVPHandler) GetRSVPPage(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.renderError(w, http.StatusNotFound, "Invalid invite link")
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

	if event.Status == models.EventStatusCancelled {
		h.renderError(w, http.StatusGone, "This event has been cancelled")
		return
	}

	if event.Status == models.EventStatusArchived {
		h.renderError(w, http.StatusGone, "This event is no longer active")
		return
	}

	existingRSVP, err := h.rsvpRepo.GetByInviteID(r.Context(), invite.ID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			h.renderError(w, http.StatusInternalServerError, "Failed to check RSVP status")
			return
		}
	}

	questions, err := h.questionRepo.GetByEventID(r.Context(), event.ID)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to load questions")
		return
	}

	questionsWithOptions := make([]*QuestionWithOptions, len(questions))
	for i, q := range questions {
		opts, _ := q.ParseOptions()
		questionsWithOptions[i] = &QuestionWithOptions{
			PreferenceQuestion: q,
			ParsedOptions:      opts,
		}
	}

	if err := h.inviteService.MarkInviteViewed(r.Context(), invite.ID); err != nil {
	}

	loc, err := time.LoadLocation(event.Timezone)
	if err != nil {
		loc = time.UTC
	}

	localStartTime := event.StartTime.In(loc).Format("Monday, January 2, 2006 at 3:04 PM MST")
	localEndTime := ""
	if event.EndTime != nil {
		localEndTime = event.EndTime.In(loc).Format("3:04 PM MST")
	}

	now := time.Now()
	deadlinePassed := false
	if event.RSVPDeadline != nil && event.RSVPDeadline.Before(now) {
		deadlinePassed = true
	}

	eventPassed := event.StartTime.Before(now)
	canUpdate := existingRSVP != nil && !deadlinePassed && !eventPassed

	timeUntilEvent := ""
	if !eventPassed {
		duration := event.StartTime.Sub(now)
		days := int(duration.Hours() / 24)
		if days > 0 {
			timeUntilEvent = fmt.Sprintf("%d days", days)
		} else {
			hours := int(duration.Hours())
			if hours > 0 {
				timeUntilEvent = fmt.Sprintf("%d hours", hours)
			} else {
				timeUntilEvent = "less than an hour"
			}
		}
	}

	data := &RSVPPageData{
		Event:          event,
		Invite:         invite,
		ExistingRSVP:   existingRSVP,
		Questions:      questionsWithOptions,
		Token:          token,
		DeadlinePassed: deadlinePassed,
		EventPassed:    eventPassed,
		LocalStartTime: localStartTime,
		LocalEndTime:   localEndTime,
		TimeUntilEvent: timeUntilEvent,
		CanUpdate:      canUpdate,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *RSVPHandler) handleInviteError(w http.ResponseWriter, err error) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "expired") {
		h.renderError(w, http.StatusGone, "This invite has expired")
		return
	}
	if strings.Contains(errMsg, "revoked") {
		h.renderError(w, http.StatusForbidden, "This invite has been revoked")
		return
	}
	h.renderError(w, http.StatusNotFound, "Invite not found or has been revoked")
}

func (h *RSVPHandler) renderError(w http.ResponseWriter, status int, message string) {
	data := &RSVPPageData{
		ErrorMessage: message,
	}
	h.renderPage(w, status, data)
}

func (h *RSVPHandler) renderPage(w http.ResponseWriter, status int, data *RSVPPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "rsvp_page.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RSVP</title>
</head>
<body>
    <h1>RSVP Page</h1>
    %s
</body>
</html>`, func() string {
		if data.ErrorMessage != "" {
			return fmt.Sprintf("<p>Error: %s</p>", data.ErrorMessage)
		}
		if data.Event != nil {
			return fmt.Sprintf("<p>Event: %s</p>", data.Event.Title)
		}
		return "<p>Loading...</p>"
	}())
}

func (h *RSVPHandler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

func (h *RSVPHandler) SubmitRSVP(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "token is required",
		})
		return
	}

	var req rsvp.SubmitRSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.rsvpService.SubmitRSVP(r.Context(), token, &req)
	if err != nil {
		h.handleSubmitError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"rsvp":    result,
		"message": "RSVP submitted successfully",
	})
}

func (h *RSVPHandler) handleSubmitError(w http.ResponseWriter, err error) {
	var validationErr *models.ValidationError
	if errors.As(err, &validationErr) {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": validationErr.Message,
			"field": validationErr.Field,
		})
		return
	}

	if errors.Is(err, rsvp.ErrDeadlinePassed) {
		h.respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "RSVP deadline has passed",
		})
		return
	}

	if errors.Is(err, rsvp.ErrDuplicateRSVP) {
		h.respondJSON(w, http.StatusConflict, map[string]string{
			"error": "you have already responded to this invite",
		})
		return
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "expired") {
		h.respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "this invite has expired",
		})
		return
	}

	if strings.Contains(errMsg, "revoked") {
		h.respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "this invite has been revoked",
		})
		return
	}

	if strings.Contains(errMsg, "cancelled") {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "this event has been cancelled",
		})
		return
	}

	h.respondJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "failed to save RSVP, please try again",
	})
}

func (h *RSVPHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *RSVPHandler) UpdateRSVP(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "token is required",
		})
		return
	}

	var req rsvp.SubmitRSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.rsvpService.UpdateRSVP(r.Context(), token, &req)
	if err != nil {
		h.handleUpdateError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"rsvp":    result,
		"message": "RSVP updated successfully",
	})
}

func (h *RSVPHandler) handleUpdateError(w http.ResponseWriter, err error) {
	var validationErr *models.ValidationError
	if errors.As(err, &validationErr) {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": validationErr.Message,
			"field": validationErr.Field,
		})
		return
	}

	var notFoundErr *models.NotFoundError
	if errors.As(err, &notFoundErr) {
		h.respondJSON(w, http.StatusNotFound, map[string]string{
			"error": "no existing RSVP found to update",
		})
		return
	}

	if errors.Is(err, rsvp.ErrDeadlinePassed) {
		h.respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "RSVP deadline has passed",
		})
		return
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "expired") {
		h.respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "this invite has expired",
		})
		return
	}

	if strings.Contains(errMsg, "revoked") {
		h.respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "this invite has been revoked",
		})
		return
	}

	if strings.Contains(errMsg, "cancelled") {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "this event has been cancelled",
		})
		return
	}

	h.respondJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "failed to update RSVP, please try again",
	})
}
