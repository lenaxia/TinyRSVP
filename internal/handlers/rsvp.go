package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/rsvp"
	"github.com/lenaxia/tinyrsvp/pkg/ics"
)

type RSVPInviteService interface {
	GetInviteByToken(ctx context.Context, token string) (*models.Invite, error)
	MarkInviteViewed(ctx context.Context, inviteID int64) error
	UnsubscribeFromReminders(ctx context.Context, token string) error
}

type RSVPService interface {
	SubmitRSVP(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error)
	UpdateRSVP(ctx context.Context, token string, req *rsvp.SubmitRSVPRequest) (*models.RSVP, error)
}

type RSVPHandler struct {
	inviteService         RSVPInviteService
	eventRepo             repositories.EventRepository
	rsvpRepo              repositories.RSVPRepository
	questionRepo          repositories.QuestionRepository
	answerRepo            repositories.AnswerRepository
	rsvpService           RSVPService
	templateRepo          repositories.TemplateRepository
	templateService       TemplateService
	customizationService  CustomizationService
	templates             *template.Template
	confirmationTemplates *template.Template
}

type TemplateService interface {
	RenderRSVPPage(w io.Writer, event *models.Event, template *models.Template) error
}

type CustomizationService interface {
	GetEventCustomization(ctx context.Context, eventID int64) (*events.EventCustomizationData, error)
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

func (h *RSVPHandler) SetAnswerRepository(repo repositories.AnswerRepository) {
	h.answerRepo = repo
}

func (h *RSVPHandler) SetTemplateRepository(repo repositories.TemplateRepository) {
	h.templateRepo = repo
}

func (h *RSVPHandler) SetTemplateService(service TemplateService) {
	h.templateService = service
}

func (h *RSVPHandler) SetCustomizationService(service CustomizationService) {
	h.customizationService = service
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
	CSRFToken      string
	ThemeCategory  string
	ThemeImageURL  string
	ThemeColor     template.HTML
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

	now := time.Now().UTC()
	deadlinePassed := false
	if event.RSVPDeadline != nil {
		deadline := event.RSVPDeadline.UTC()
		if !deadline.After(now) {
			deadlinePassed = true
		}
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

	theme, err := h.getEventTheme(r.Context(), event)
	if err != nil {
		h.renderError(w, http.StatusInternalServerError, "Failed to load theme")
		return
	}

	themeCategory := ""
	if theme != nil && theme.Category != models.CategoryPlain {
		themeCategory = getThemeSlug(TemplateCategory(theme.Category))
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
		ErrorMessage:   r.URL.Query().Get("error"),
		CSRFToken:      middleware.GetCSRFToken(r.Context()),
		ThemeCategory:  themeCategory,
		ThemeImageURL:  h.getThemeImageURL(event, theme),
		ThemeColor:     h.getThemeColor(event, theme),
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

	if h.templateService != nil && data.Event != nil && h.templateRepo != nil {
		theme, err := h.getEventTheme(context.Background(), data.Event)
		if err == nil && theme != nil {
			var mergedConfig *models.ComponentConfiguration
			if h.customizationService != nil {
				customization, err := h.customizationService.GetEventCustomization(context.Background(), data.Event.ID)
				if err == nil && customization != nil {
					mergedConfig = customization.MergedConfig
				}
			}

			if mergedConfig != nil {
				mergedTemplate := *theme
				configJSON, _ := json.Marshal(mergedConfig)
				configStr := string(configJSON)
				mergedTemplate.ComponentConfig = &configStr

				var buf bytes.Buffer
				if err := h.templateService.RenderRSVPPage(&buf, data.Event, &mergedTemplate); err == nil {
					w.Write(buf.Bytes())
					return
				}
			} else {
				var buf bytes.Buffer
				if err := h.templateService.RenderRSVPPage(&buf, data.Event, theme); err == nil {
					w.Write(buf.Bytes())
					return
				}
			}
		}
	}

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

func (h *RSVPHandler) SetConfirmationTemplates(tmpl *template.Template) {
	h.confirmationTemplates = tmpl
}

func (h *RSVPHandler) getEventTheme(ctx context.Context, event *models.Event) (*models.Template, error) {
	if h.templateRepo == nil {
		return nil, nil
	}

	if event.TemplateID != nil {
		theme, err := h.templateRepo.GetByID(ctx, *event.TemplateID)
		if err == nil {
			return theme, nil
		}
	}

	return h.templateRepo.GetDefaultByType(ctx, models.TemplateTypeRSVPPage)
}

func (h *RSVPHandler) getThemeImageURL(event *models.Event, theme *models.Template) string {
	if event.CustomThemeImageURL != nil && *event.CustomThemeImageURL != "" {
		return *event.CustomThemeImageURL
	}

	if theme != nil && theme.ImageURL != nil {
		return *theme.ImageURL
	}

	return ""
}

func (h *RSVPHandler) getThemeColor(event *models.Event, theme *models.Template) template.HTML {
	// Don't override colors when a full named theme is active — the theme defines its own palette.
	// Only inject the custom color for plain/no-theme pages.
	if theme != nil && theme.Category != models.CategoryPlain && theme.Category != models.CategoryPlainText {
		return ""
	}

	if event.CustomThemeColor != nil && *event.CustomThemeColor != "" {
		color := *event.CustomThemeColor
		if !isValidHexColor(color) {
			return ""
		}
		valid, _ := validateCustomColorContrast(color)
		if !valid {
			return ""
		}
		return template.HTML(generateColorOverrideCSS(color))
	}

	return ""
}

func (h *RSVPHandler) parseRSVPRequest(r *http.Request) (*rsvp.SubmitRSVPRequest, error) {
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		var req rsvp.SubmitRSVPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, err
		}
		return &req, nil
	}

	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	req := &rsvp.SubmitRSVPRequest{
		Response: r.FormValue("response"),
		Answers:  []rsvp.AnswerRequest{},
	}

	if plusOnesStr := r.FormValue("plus_ones"); plusOnesStr != "" {
		plusOnes, err := strconv.Atoi(plusOnesStr)
		if err != nil {
			return nil, fmt.Errorf("invalid plus_ones value")
		}
		req.PlusOnes = plusOnes
	}

	if adultsStr := r.FormValue("adults_count"); adultsStr != "" {
		adults, err := strconv.Atoi(adultsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid adults_count value")
		}
		req.AdultsCount = &adults
	}

	if kidsStr := r.FormValue("kids_count"); kidsStr != "" {
		kids, err := strconv.Atoi(kidsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid kids_count value")
		}
		req.KidsCount = &kids
	}

	answerMap := make(map[int64]*rsvp.AnswerRequest)

	for key, values := range r.Form {
		if !strings.HasPrefix(key, "answers[") {
			continue
		}

		key = strings.TrimPrefix(key, "answers[")
		closeBracket := strings.Index(key, "]")
		if closeBracket == -1 {
			continue
		}

		questionIDStr := key[:closeBracket]
		questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
		if err != nil {
			continue
		}

		fieldType := key[closeBracket+1:]
		fieldType = strings.TrimPrefix(fieldType, "[")
		fieldType = strings.TrimSuffix(fieldType, "]")
		fieldType = strings.TrimSuffix(fieldType, "[]")

		if _, exists := answerMap[questionID]; !exists {
			answerMap[questionID] = &rsvp.AnswerRequest{
				QuestionID: questionID,
			}
		}

		answer := answerMap[questionID]

		switch fieldType {
		case "text":
			if len(values) > 0 && values[0] != "" {
				text := values[0]
				if len(text) > 2000 {
					text = text[:2000]
				}
				answer.AnswerText = &text
			}
		case "option":
			if len(values) > 0 && values[0] != "" {
				option := values[0]
				answer.AnswerOption = &option
			}
		case "options":
			if len(values) > 0 {
				combined := strings.Join(values, ", ")
				answer.AnswerOption = &combined
			}
		}
	}

	for _, answer := range answerMap {
		req.Answers = append(req.Answers, *answer)
	}

	return req, nil
}

func (h *RSVPHandler) SubmitRSVP(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "token is required",
		})
		return
	}

	req, err := h.parseRSVPRequest(r)
	if err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	var existingRSVP *models.RSVP
	var result *models.RSVP

	if h.rsvpRepo != nil {
		invite, inviteErr := h.inviteService.GetInviteByToken(r.Context(), token)
		if inviteErr != nil {
			h.handleSubmitError(w, r, token, inviteErr)
			return
		}

		existingRSVP, err = h.rsvpRepo.GetByInviteID(r.Context(), invite.ID)
		if err != nil {
			var notFoundErr *models.NotFoundError
			if !errors.As(err, &notFoundErr) {
				h.handleSubmitError(w, r, token, fmt.Errorf("failed to check RSVP status"))
				return
			}
		}
	}

	if existingRSVP != nil {
		result, err = h.rsvpService.UpdateRSVP(r.Context(), token, req)
		if err != nil {
			h.handleUpdateError(w, r, token, err)
			return
		}
	} else {
		result, err = h.rsvpService.SubmitRSVP(r.Context(), token, req)
		if err != nil {
			h.handleSubmitError(w, r, token, err)
			return
		}
	}

	acceptHeader := r.Header.Get("Accept")
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(acceptHeader, "application/json") || strings.Contains(contentType, "application/json") {
		statusCode := http.StatusCreated
		if existingRSVP != nil {
			statusCode = http.StatusOK
		}
		h.respondJSON(w, statusCode, map[string]interface{}{
			"rsvp":    result,
			"message": "RSVP submitted successfully",
		})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/rsvp/%s/confirmation", token), http.StatusSeeOther)
}

func (h *RSVPHandler) isJSONRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") ||
		strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

func (h *RSVPHandler) handleSubmitError(w http.ResponseWriter, r *http.Request, token string, err error) {
	var msg string

	var validationErr *models.ValidationError
	var deadlineErr *models.DeadlinePassedError

	switch {
	case errors.As(err, &validationErr):
		msg = validationErr.Message
	case errors.As(err, &deadlineErr):
		msg = deadlineErr.Message
	case errors.Is(err, rsvp.ErrDuplicateRSVP):
		msg = "you have already responded to this invite"
	case strings.Contains(err.Error(), "expired"):
		msg = "this invite has expired"
	case strings.Contains(err.Error(), "revoked"):
		msg = "this invite has been revoked"
	case strings.Contains(err.Error(), "cancelled"):
		msg = "this event has been cancelled"
	default:
		msg = "failed to save RSVP, please try again"
	}

	if h.isJSONRequest(r) {
		status := http.StatusInternalServerError
		if errors.As(err, &validationErr) {
			status = http.StatusBadRequest
		} else if errors.As(err, &deadlineErr) || strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "revoked") {
			status = http.StatusForbidden
		} else if errors.Is(err, rsvp.ErrDuplicateRSVP) {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "cancelled") {
			status = http.StatusBadRequest
		}
		h.respondJSON(w, status, map[string]string{"error": msg})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/rsvp/%s?error=%s", token, url.QueryEscape(msg)), http.StatusSeeOther)
}

func (h *RSVPHandler) handleUpdateError(w http.ResponseWriter, r *http.Request, token string, err error) {
	var msg string

	var validationErr *models.ValidationError
	var notFoundErr *models.NotFoundError
	var deadlineErr *models.DeadlinePassedError

	switch {
	case errors.As(err, &validationErr):
		msg = validationErr.Message
	case errors.As(err, &notFoundErr):
		msg = "no existing RSVP found to update"
	case errors.As(err, &deadlineErr):
		msg = deadlineErr.Message
	case strings.Contains(err.Error(), "expired"):
		msg = "this invite has expired"
	case strings.Contains(err.Error(), "revoked"):
		msg = "this invite has been revoked"
	case strings.Contains(err.Error(), "cancelled"):
		msg = "this event has been cancelled"
	default:
		msg = "failed to update RSVP, please try again"
	}

	if h.isJSONRequest(r) {
		status := http.StatusInternalServerError
		if errors.As(err, &validationErr) {
			status = http.StatusBadRequest
		} else if errors.As(err, &notFoundErr) {
			status = http.StatusNotFound
		} else if errors.As(err, &deadlineErr) || strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "revoked") {
			status = http.StatusForbidden
		} else if strings.Contains(err.Error(), "cancelled") {
			status = http.StatusBadRequest
		}
		h.respondJSON(w, status, map[string]string{"error": msg})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/rsvp/%s?error=%s", token, url.QueryEscape(msg)), http.StatusSeeOther)
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

	req, err := h.parseRSVPRequest(r)
	if err != nil {
		h.respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.rsvpService.UpdateRSVP(r.Context(), token, req)
	if err != nil {
		h.handleUpdateError(w, r, token, err)
		return
	}

	acceptHeader := r.Header.Get("Accept")
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(acceptHeader, "application/json") || strings.Contains(contentType, "application/json") {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{
			"rsvp":    result,
			"message": "RSVP updated successfully",
		})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/rsvp/%s/confirmation", token), http.StatusSeeOther)
}

type AnswerWithQuestion struct {
	Answer   *models.RSVPAnswer
	Question *models.PreferenceQuestion
}

type ConfirmationPageData struct {
	ActivePage           string
	Event                *models.Event
	Invite               *models.Invite
	RSVP                 *models.RSVP
	AnswersWithQuestions []*AnswerWithQuestion
	Token                string
	CanUpdate            bool
	LocalStartTime       string
	LocalEndTime         string
	ErrorMessage         string
	CSRFToken            string
}

func (h *RSVPHandler) GetConfirmationPage(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.renderConfirmationError(w, http.StatusNotFound, "Invalid invite link")
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
			h.renderConfirmationError(w, http.StatusNotFound, "Event not found")
			return
		}
		h.renderConfirmationError(w, http.StatusInternalServerError, "Failed to load event")
		return
	}

	if event.Status == models.EventStatusCancelled {
		h.renderConfirmationError(w, http.StatusGone, "This event has been cancelled")
		return
	}

	if event.Status == models.EventStatusArchived {
		h.renderConfirmationError(w, http.StatusGone, "This event is no longer active")
		return
	}

	existingRSVP, err := h.rsvpRepo.GetByInviteID(r.Context(), invite.ID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			h.renderConfirmationError(w, http.StatusNotFound, "No RSVP found for this invite")
			return
		}
		h.renderConfirmationError(w, http.StatusInternalServerError, "Failed to load RSVP")
		return
	}

	questions, err := h.questionRepo.GetByEventID(r.Context(), event.ID)
	if err != nil {
		h.renderConfirmationError(w, http.StatusInternalServerError, "Failed to load questions")
		return
	}

	questionMap := make(map[int64]*models.PreferenceQuestion)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	var answers []*models.RSVPAnswer
	if h.answerRepo != nil {
		answers, err = h.answerRepo.GetByRSVPID(r.Context(), existingRSVP.ID)
		if err != nil {
			h.renderConfirmationError(w, http.StatusInternalServerError, "Failed to load answers")
			return
		}
	}

	answersWithQuestions := make([]*AnswerWithQuestion, 0, len(answers))
	for _, answer := range answers {
		if question, ok := questionMap[answer.QuestionID]; ok {
			answersWithQuestions = append(answersWithQuestions, &AnswerWithQuestion{
				Answer:   answer,
				Question: question,
			})
		}
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

	now := time.Now().UTC()
	deadlinePassed := false
	if event.RSVPDeadline != nil {
		deadline := event.RSVPDeadline.UTC()
		if !deadline.After(now) {
			deadlinePassed = true
		}
	}

	eventPassed := event.StartTime.Before(now)
	canUpdate := !deadlinePassed && !eventPassed

	data := &ConfirmationPageData{
		Event:                event,
		Invite:               invite,
		RSVP:                 existingRSVP,
		AnswersWithQuestions: answersWithQuestions,
		Token:                token,
		CanUpdate:            canUpdate,
		LocalStartTime:       localStartTime,
		LocalEndTime:         localEndTime,
		CSRFToken:            middleware.GetCSRFToken(r.Context()),
	}

	h.renderConfirmationPage(w, http.StatusOK, data)
}

func (h *RSVPHandler) renderConfirmationError(w http.ResponseWriter, status int, message string) {
	data := &ConfirmationPageData{
		ErrorMessage: message,
	}
	h.renderConfirmationPage(w, status, data)
}

func (h *RSVPHandler) renderConfirmationPage(w http.ResponseWriter, status int, data *ConfirmationPageData) {
	if h.confirmationTemplates != nil {
		// Buffer the render so we can still write a clean error response if
		// template execution fails (writing w.WriteHeader before Execute would
		// make any subsequent http.Error a no-op — headers already sent).
		var buf bytes.Buffer
		if err := h.confirmationTemplates.ExecuteTemplate(&buf, "confirmation.html", data); err != nil {
			fmt.Printf("ERROR: confirmation template execution failed: %v\n", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		buf.WriteTo(w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
	   <meta charset="UTF-8">
	   <meta name="viewport" content="width=device-width, initial-scale=1.0">
	   <title>RSVP Confirmation</title>
</head>
<body>
	   <h1>RSVP Confirmation</h1>
	   %s
</body>
</html>`, func() string {
		if data.ErrorMessage != "" {
			return fmt.Sprintf("<p>Error: %s</p>", data.ErrorMessage)
		}
		if data.Event != nil && data.RSVP != nil {
			return fmt.Sprintf("<p>Thank you for your RSVP to %s! Your response: %s</p>", data.Event.Title, data.RSVP.Response)
		}
		return "<p>Loading...</p>"
	}())
}

type UnsubscribePageData struct {
	Event        *models.Event
	Invite       *models.Invite
	ErrorMessage string
	Success      bool
}

func (h *RSVPHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.renderUnsubscribeError(w, http.StatusNotFound, "Invalid unsubscribe link")
		return
	}

	invite, err := h.inviteService.GetInviteByToken(r.Context(), token)
	if err != nil {
		h.handleUnsubscribeInviteError(w, err)
		return
	}

	event, err := h.eventRepo.GetByID(r.Context(), invite.EventID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			h.renderUnsubscribeError(w, http.StatusNotFound, "Event not found")
			return
		}
		h.renderUnsubscribeError(w, http.StatusInternalServerError, "Failed to load event")
		return
	}

	if err := h.inviteService.UnsubscribeFromReminders(r.Context(), token); err != nil {
		h.renderUnsubscribeError(w, http.StatusInternalServerError, "Failed to unsubscribe")
		return
	}

	data := &UnsubscribePageData{
		Event:   event,
		Invite:  invite,
		Success: true,
	}

	h.renderUnsubscribePage(w, http.StatusOK, data)
}

func (h *RSVPHandler) handleUnsubscribeInviteError(w http.ResponseWriter, err error) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "expired") {
		h.renderUnsubscribeError(w, http.StatusGone, "This invite has expired")
		return
	}
	if strings.Contains(errMsg, "revoked") {
		h.renderUnsubscribeError(w, http.StatusForbidden, "This invite has been revoked")
		return
	}

	var notFoundErr *models.NotFoundError
	if errors.As(err, &notFoundErr) {
		h.renderUnsubscribeError(w, http.StatusNotFound, "Invite not found")
		return
	}

	h.renderUnsubscribeError(w, http.StatusNotFound, "Invite not found or has been revoked")
}

func (h *RSVPHandler) renderUnsubscribeError(w http.ResponseWriter, status int, message string) {
	data := &UnsubscribePageData{
		ErrorMessage: message,
		Success:      false,
	}
	h.renderUnsubscribePage(w, status, data)
}

func (h *RSVPHandler) renderUnsubscribePage(w http.ResponseWriter, status int, data *UnsubscribePageData) {
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

func (h *RSVPHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	invite, err := h.inviteService.GetInviteByToken(r.Context(), token)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			http.Error(w, "invite not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load invite", http.StatusInternalServerError)
		return
	}

	event, err := h.eventRepo.GetByID(r.Context(), invite.EventID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load event", http.StatusInternalServerError)
		return
	}

	scheme := resolveScheme(r.Header.Get("X-Forwarded-Proto"))
	rsvpURL := fmt.Sprintf("%s://%s/rsvp/%s", scheme, r.Host, token)

	generator := ics.NewGenerator()
	data, err := generator.Generate(event, rsvpURL)
	if err != nil {
		http.Error(w, "failed to generate calendar file", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s.ics", strings.ReplaceAll(event.Title, " ", "_"))
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func resolveScheme(forwardedProto string) string {
	if forwardedProto == "http" || forwardedProto == "https" {
		return forwardedProto
	}
	return "https"
}
