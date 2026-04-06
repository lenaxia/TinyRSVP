package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

type EventWebHandlers struct {
	service         events.Service
	templateService templates.Service
	listTemplates   *template.Template
	formTemplates   *template.Template
	detailTemplates *template.Template
}

type EventListPageData struct {
	ActivePage string
	Events     []*models.Event
	Total      int
	Page       int
	Filter     string
	Sort       string
	Error      string
	Loading    bool
}

type EventFormPageData struct {
	ActivePage      string
	Event           *models.Event
	Questions       []*models.PreferenceQuestion
	Themes          []*models.Template
	SelectedThemeID int64
	Errors          map[string]string
	Error           string
	CSRFToken       string
}

type EventDetailPageData struct {
	ActivePage string
	Event      *models.Event
	CSRFToken  string
	Error      string
}

func NewEventWebHandlers(service events.Service, templateService templates.Service, listTmpl, formTmpl, detailTmpl *template.Template) *EventWebHandlers {
	return &EventWebHandlers{
		service:         service,
		templateService: templateService,
		listTemplates:   listTmpl,
		formTemplates:   formTmpl,
		detailTemplates: detailTmpl,
	}
}

func (h *EventWebHandlers) ListEventsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "list events",
			Resource: "Event",
		})
		return
	}

	if !auth.NewAuthorizationChecker().IsEventManager(user) {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "list events",
			Resource: "Event",
		})
		return
	}

	limit := 50
	offset := 0
	page := 1

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
			offset = (page - 1) * limit
		}
	}

	filters := events.ListFilters{
		Limit:  limit,
		Offset: offset,
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status, err := parseEventStatus(statusStr)
		if err == nil {
			filters.Status = &status
		}
	}

	eventList, err := h.service.ListEvents(r.Context(), filters)
	if err != nil {
		h.renderListPage(w, http.StatusOK, &EventListPageData{
			ActivePage: "events",
			Error:      "Failed to load events",
		})
		return
	}

	data := &EventListPageData{
		ActivePage: "events",
		Events:     eventList,
		Total:      len(eventList),
		Page:       page,
		Filter:     r.URL.Query().Get("status"),
		Sort:       r.URL.Query().Get("sort"),
	}

	h.renderListPage(w, http.StatusOK, data)
}

func (h *EventWebHandlers) NewEventForm(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "create event",
			Resource: "Event",
		})
		return
	}

	if !auth.NewAuthorizationChecker().CanCreateEvent(r.Context(), user) {
		HandleError(w, r, &models.PermissionDeniedError{
			Action:   "create event",
			Resource: "Event",
		})
		return
	}

	csrfToken := middleware.GetCSRFToken(r.Context())

	themes, defaultTheme, err := h.loadThemes(r.Context())
	if err != nil {
		// Log the error but continue - themes are optional
		// The form will still work without themes
		themes = []*models.Template{}
	}

	selectedThemeID := int64(0)
	if defaultTheme != nil {
		selectedThemeID = defaultTheme.ID
	}

	data := &EventFormPageData{
		ActivePage: "events",
		Event: &models.Event{
			MaxPlusOnes:    0,
			AllowMaybeRSVP: true,
		},
		Questions:       []*models.PreferenceQuestion{},
		Themes:          themes,
		SelectedThemeID: selectedThemeID,
		Errors:          make(map[string]string),
		CSRFToken:       csrfToken,
	}

	h.renderFormPage(w, http.StatusOK, data)
}

func (h *EventWebHandlers) EditEventForm(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	event, err := h.service.GetEvent(r.Context(), id)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	csrfToken := middleware.GetCSRFToken(r.Context())

	themes, defaultTheme, err := h.loadThemes(r.Context())
	if err != nil {
		// Log error but continue - themes are optional for form rendering
		themes = []*models.Template{}
		defaultTheme = nil
	}

	selectedThemeID := int64(0)
	if event.TemplateID != nil {
		selectedThemeID = *event.TemplateID
	} else if defaultTheme != nil {
		selectedThemeID = defaultTheme.ID
	}

	data := &EventFormPageData{
		ActivePage:      "events",
		Event:           event,
		Questions:       []*models.PreferenceQuestion{},
		Themes:          themes,
		SelectedThemeID: selectedThemeID,
		Errors:          make(map[string]string),
		CSRFToken:       csrfToken,
	}

	h.renderFormPage(w, http.StatusOK, data)
}

func (h *EventWebHandlers) GetEventPage(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	event, err := h.service.GetEvent(r.Context(), id)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	csrfToken := middleware.GetCSRFToken(r.Context())

	data := &EventDetailPageData{
		ActivePage: "events",
		Event:      event,
		CSRFToken:  csrfToken,
	}

	h.renderDetailPage(w, http.StatusOK, data)
}

func (h *EventWebHandlers) CreateEventFromForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HandleError(w, r, NewBadRequestError("invalid form data"))
		return
	}

	event, err := parseEventFormData(r.Form)
	if err != nil {
		HandleError(w, r, NewBadRequestError(err.Error()))
		return
	}

	if err := h.service.CreateEvent(r.Context(), event); err != nil {
		HandleError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/events/%d", event.ID), http.StatusSeeOther)
}

func (h *EventWebHandlers) UpdateEventFromForm(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	if err := r.ParseForm(); err != nil {
		HandleError(w, r, NewBadRequestError("invalid form data"))
		return
	}

	existing, err := h.service.GetEvent(r.Context(), id)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	versionStr := r.FormValue("version")
	if versionStr == "" {
		versionStr = "1"
	}
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid version"))
		return
	}

	event := &models.Event{
		ID:                     id,
		Version:                version,
		Title:                  existing.Title,
		Description:            existing.Description,
		StartTime:              existing.StartTime,
		EndTime:                existing.EndTime,
		Timezone:               existing.Timezone,
		Location:               existing.Location,
		MaxPlusOnes:            existing.MaxPlusOnes,
		RSVPDeadline:           existing.RSVPDeadline,
		AllowRSVPAfterDeadline: existing.AllowRSVPAfterDeadline,
		AllowMaybeRSVP:         existing.AllowMaybeRSVP,
		PrivateGuestList:       existing.PrivateGuestList,
		FamilyHeadcount:        existing.FamilyHeadcount,
		EventCapacity:          existing.EventCapacity,
		TemplateID:             existing.TemplateID,
		CustomThemeColor:       existing.CustomThemeColor,
		CustomThemeImageURL:    existing.CustomThemeImageURL,
	}

	if title := r.FormValue("title"); title != "" {
		event.Title = title
	}

	if desc := r.FormValue("description"); desc != "" {
		event.Description = &desc
	} else if r.Form.Has("description") {
		event.Description = nil
	}

	if loc := r.FormValue("location"); loc != "" {
		event.Location = &loc
	} else if r.Form.Has("location") {
		event.Location = nil
	}

	if startTimeStr := r.FormValue("start_time"); startTimeStr != "" {
		startTime, err := time.Parse("2006-01-02T15:04", startTimeStr)
		if err == nil {
			event.StartTime = startTime
		}
	}

	if endTimeStr := r.FormValue("end_time"); endTimeStr != "" {
		endTime, err := time.Parse("2006-01-02T15:04", endTimeStr)
		if err == nil {
			event.EndTime = &endTime
		}
	} else if r.Form.Has("end_time") {
		event.EndTime = nil
	}

	if tz := r.FormValue("timezone"); tz != "" {
		event.Timezone = tz
	}

	if maxPOStr := r.FormValue("max_plus_ones"); maxPOStr != "" {
		maxPO, err := strconv.Atoi(maxPOStr)
		if err == nil {
			event.MaxPlusOnes = maxPO
		}
	}

	if r.Form.Has("rsvp_settings_saved") {
		if deadlineStr := r.FormValue("rsvp_deadline"); deadlineStr != "" {
			deadline, err := time.Parse("2006-01-02T15:04", deadlineStr)
			if err == nil {
				event.RSVPDeadline = &deadline
			}
		} else {
			event.RSVPDeadline = nil
		}

		event.AllowRSVPAfterDeadline = r.FormValue("allow_rsvp_after_deadline") == "on"
		event.AllowMaybeRSVP = r.FormValue("allow_maybe_rsvp") == "on"
		event.PrivateGuestList = r.FormValue("private_guest_list") == "on"
		event.FamilyHeadcount = r.FormValue("family_headcount") == "on"

		if capacityStr := r.FormValue("event_capacity"); capacityStr != "" {
			capacity, err := strconv.Atoi(capacityStr)
			if err == nil && capacity > 0 {
				event.EventCapacity = &capacity
			}
		} else {
			event.EventCapacity = nil
		}
	}

	if friendlyName := strings.TrimSpace(r.FormValue("friendly_name")); friendlyName != "" {
		event.FriendlyName = &friendlyName
	} else if r.Form.Has("friendly_name") {
		event.FriendlyName = nil
	}

	if templateIDStr := strings.TrimSpace(r.FormValue("template_id")); templateIDStr != "" {
		templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
		if err == nil && templateID > 0 {
			event.TemplateID = &templateID
		}
	} else if r.Form.Has("template_id") {
		event.TemplateID = nil
	}

	if color := strings.TrimSpace(r.FormValue("custom_theme_color")); color != "" {
		event.CustomThemeColor = &color
	} else if r.Form.Has("custom_theme_color") {
		event.CustomThemeColor = nil
	}

	if imageURL := strings.TrimSpace(r.FormValue("custom_theme_image_url")); imageURL != "" {
		event.CustomThemeImageURL = &imageURL
	} else if r.Form.Has("custom_theme_image_url") {
		event.CustomThemeImageURL = nil
	}

	if err := h.service.UpdateEvent(r.Context(), event); err != nil {
		HandleError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/events/%d", id), http.StatusSeeOther)
}

func (h *EventWebHandlers) PublishEventAction(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	if err := h.service.PublishEvent(r.Context(), id); err != nil {
		HandleError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/events/%d", id), http.StatusSeeOther)
}

func (h *EventWebHandlers) CancelEventAction(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	if err := r.ParseForm(); err != nil {
		HandleError(w, r, NewBadRequestError("invalid form data"))
		return
	}

	reason := strings.TrimSpace(r.FormValue("reason"))
	if reason == "" {
		HandleError(w, r, NewBadRequestError("reason is required"))
		return
	}

	if len(reason) < 10 || len(reason) > 500 {
		HandleError(w, r, NewBadRequestError("reason must be between 10 and 500 characters"))
		return
	}

	if err := h.service.CancelEvent(r.Context(), id, reason); err != nil {
		HandleError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/events/%d", id), http.StatusSeeOther)
}

func (h *EventWebHandlers) DeleteEventAction(w http.ResponseWriter, r *http.Request) {
	id, err := parseEventID(chi.URLParam(r, "id"))
	if err != nil {
		HandleError(w, r, NewBadRequestError("invalid event ID"))
		return
	}

	if err := h.service.DeleteEvent(r.Context(), id); err != nil {
		HandleError(w, r, err)
		return
	}

	http.Redirect(w, r, "/events", http.StatusSeeOther)
}

func (h *EventWebHandlers) renderListPage(w http.ResponseWriter, status int, data *EventListPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.listTemplates != nil {
		if err := h.listTemplates.ExecuteTemplate(w, "event_list.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<body>
	%s
</body>
</html>`, func() string {
		if data.Error != "" {
			return fmt.Sprintf("<div>Error: %s</div>", data.Error)
		}
		if len(data.Events) == 0 {
			return "<div>No Events Found</div>"
		}
		var sb strings.Builder
		for _, event := range data.Events {
			sb.WriteString(fmt.Sprintf("<div>%s</div>", event.Title))
		}
		return sb.String()
	}())
}

func (h *EventWebHandlers) renderFormPage(w http.ResponseWriter, status int, data *EventFormPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.formTemplates != nil {
		if err := h.formTemplates.ExecuteTemplate(w, "event_form.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	title := "Create Event"
	if data.Event != nil && data.Event.ID > 0 {
		title = "Edit Event"
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<body>
	<h1>%s</h1>
	<form>
		<input type="hidden" name="csrf_token" value="%s">
	</form>
</body>
</html>`, title, data.CSRFToken)
}

func (h *EventWebHandlers) renderDetailPage(w http.ResponseWriter, status int, data *EventDetailPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if h.detailTemplates != nil {
		if err := h.detailTemplates.ExecuteTemplate(w, "event_detail.html", data); err != nil {
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<body>
	<h1>%s</h1>
</body>
</html>`, data.Event.Title)
}

func parseEventFormData(form url.Values) (*models.Event, error) {
	title := strings.TrimSpace(form.Get("title"))
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if len(title) < 3 || len(title) > 200 {
		return nil, fmt.Errorf("title must be between 3 and 200 characters")
	}

	startTimeStr := form.Get("start_time")
	if startTimeStr == "" {
		return nil, fmt.Errorf("start_time is required")
	}

	startTime, err := time.Parse("2006-01-02T15:04", startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format")
	}

	timezone := strings.TrimSpace(form.Get("timezone"))
	if timezone == "" {
		return nil, fmt.Errorf("timezone is required")
	}

	maxPlusOnesStr := form.Get("max_plus_ones")
	if maxPlusOnesStr == "" {
		maxPlusOnesStr = "0"
	}

	maxPlusOnes, err := strconv.Atoi(maxPlusOnesStr)
	if err != nil {
		return nil, fmt.Errorf("invalid max_plus_ones format")
	}

	if maxPlusOnes < 0 || maxPlusOnes > 10 {
		return nil, fmt.Errorf("max_plus_ones must be between 0 and 10")
	}

	event := &models.Event{
		Title:                  title,
		StartTime:              startTime,
		Timezone:               timezone,
		MaxPlusOnes:            maxPlusOnes,
		AllowMaybeRSVP:         form.Get("allow_maybe_rsvp") == "on",
		PrivateGuestList:       form.Get("private_guest_list") == "on",
		FamilyHeadcount:        form.Get("family_headcount") == "on",
		AllowRSVPAfterDeadline: form.Get("allow_rsvp_after_deadline") == "on",
	}

	if desc := strings.TrimSpace(form.Get("description")); desc != "" {
		event.Description = &desc
	}

	if loc := strings.TrimSpace(form.Get("location")); loc != "" {
		event.Location = &loc
	}

	if endTimeStr := form.Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse("2006-01-02T15:04", endTimeStr)
		if err == nil {
			event.EndTime = &endTime
		}
	}

	if deadlineStr := form.Get("rsvp_deadline"); deadlineStr != "" {
		deadline, err := time.Parse("2006-01-02T15:04", deadlineStr)
		if err == nil {
			event.RSVPDeadline = &deadline
		}
	}

	if capacityStr := form.Get("event_capacity"); capacityStr != "" {
		capacity, err := strconv.Atoi(capacityStr)
		if err == nil && capacity > 0 {
			event.EventCapacity = &capacity
		}
	}

	if friendlyName := strings.TrimSpace(form.Get("friendly_name")); friendlyName != "" {
		event.FriendlyName = &friendlyName
	}

	if templateIDStr := strings.TrimSpace(form.Get("template_id")); templateIDStr != "" {
		templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid template_id format")
		}
		if templateID <= 0 {
			return nil, fmt.Errorf("template_id must be positive")
		}
		event.TemplateID = &templateID
	}

	return event, nil
}

func (h *EventWebHandlers) loadThemes(ctx context.Context) ([]*models.Template, *models.Template, error) {
	if h.templateService == nil {
		return nil, nil, nil
	}

	rsvpPageType := models.TemplateTypeRSVPPage
	isActive := true

	themes, err := h.templateService.ListTemplates(ctx, &repositories.TemplateFilters{
		Type:     &rsvpPageType,
		IsActive: &isActive,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list themes: %w", err)
	}

	defaultTheme, err := h.templateService.GetDefaultTemplate(ctx, models.TemplateTypeRSVPPage)
	if err != nil {
		return themes, nil, nil
	}

	return themes, defaultTheme, nil
}
