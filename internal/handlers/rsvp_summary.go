package handlers

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type RSVPSummaryHandler struct {
	eventRepo    repositories.EventRepository
	rsvpRepo     repositories.RSVPRepository
	questionRepo repositories.QuestionRepository
	answerRepo   repositories.AnswerRepository
	templates    *template.Template
}

func NewRSVPSummaryHandler(
	eventRepo repositories.EventRepository,
	rsvpRepo repositories.RSVPRepository,
	questionRepo repositories.QuestionRepository,
	answerRepo repositories.AnswerRepository,
) *RSVPSummaryHandler {
	return &RSVPSummaryHandler{
		eventRepo:    eventRepo,
		rsvpRepo:     rsvpRepo,
		questionRepo: questionRepo,
		answerRepo:   answerRepo,
	}
}

func (h *RSVPSummaryHandler) SetTemplates(tmpl *template.Template) {
	h.templates = tmpl
}

type RSVPSummaryData struct {
	Event         *models.Event
	Stats         *repositories.RSVPStats
	RSVPs         []*models.RSVP
	ResponseRate  float64
	QuestionStats map[int64]*QuestionStat
	EventID       int64
	Error         string
	Loading       bool
}

type QuestionStat struct {
	Question *models.PreferenceQuestion
	Answers  map[string]int
}

func (h *RSVPSummaryHandler) GetRSVPSummary(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		h.renderError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil || eventID <= 0 {
		h.renderError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	event, err := h.eventRepo.GetByID(r.Context(), eventID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			h.renderError(w, http.StatusNotFound, "event not found")
			return
		}
		h.renderError(w, http.StatusInternalServerError, "failed to retrieve event")
		return
	}

	if !user.IsAdmin() && event.CreatedBy != user.ID {
		h.renderError(w, http.StatusForbidden, "permission denied")
		return
	}

	stats, err := h.rsvpRepo.GetStats(r.Context(), eventID)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "failed to retrieve RSVP statistics")
		return
	}

	responseRate := 0.0
	if stats.TotalInvites > 0 {
		responded := stats.YesCount + stats.NoCount + stats.MaybeCount
		responseRate = (float64(responded) / float64(stats.TotalInvites)) * 100
	}

	rsvps, err := h.rsvpRepo.GetByEventID(r.Context(), eventID)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "failed to retrieve RSVPs")
		return
	}

	questionStats, err := h.buildQuestionStats(r.Context(), eventID, rsvps)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "failed to retrieve question statistics")
		return
	}

	data := &RSVPSummaryData{
		Event:         event,
		Stats:         stats,
		RSVPs:         rsvps,
		ResponseRate:  responseRate,
		QuestionStats: questionStats,
		EventID:       eventID,
	}

	h.renderPage(w, http.StatusOK, data)
}

func (h *RSVPSummaryHandler) buildQuestionStats(ctx context.Context, eventID int64, rsvps []*models.RSVP) (map[int64]*QuestionStat, error) {
	questions, err := h.questionRepo.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		return nil, nil
	}

	questionStats := make(map[int64]*QuestionStat)
	for _, q := range questions {
		questionStats[q.ID] = &QuestionStat{
			Question: q,
			Answers:  make(map[string]int),
		}
	}

	for _, rsvp := range rsvps {
		answers, err := h.answerRepo.GetByRSVPID(ctx, rsvp.ID)
		if err != nil {
			continue
		}

		for _, answer := range answers {
			if stat, ok := questionStats[answer.QuestionID]; ok {
				if answer.AnswerText != nil {
					stat.Answers[*answer.AnswerText]++
				}
			}
		}
	}

	return questionStats, nil
}

func (h *RSVPSummaryHandler) renderError(w http.ResponseWriter, status int, message string) {
	data := &RSVPSummaryData{
		Error: message,
	}
	h.renderPage(w, status, data)
}

func (h *RSVPSummaryHandler) renderPage(w http.ResponseWriter, status int, data *RSVPSummaryData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "rsvp_summary.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RSVP Summary</title>
</head>
<body>
    <h1>RSVP Summary</h1>
    %s
</body>
</html>`, func() string {
		if data.Error != "" {
			return fmt.Sprintf("<p>Error: %s</p>", data.Error)
		}
		if data.Event != nil && data.Stats != nil {
			return fmt.Sprintf("<p>Event: %s | Total Invites: %d | Yes: %d | No: %d | Maybe: %d</p>",
				data.Event.Title, data.Stats.TotalInvites, data.Stats.YesCount, data.Stats.NoCount, data.Stats.MaybeCount)
		}
		return "<p>Loading...</p>"
	}())
}
