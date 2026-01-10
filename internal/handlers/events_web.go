package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type EventWebHandlers struct {
	service   events.Service
	templates *template.Template
}

type EventListPageData struct {
	Events  []*models.Event
	Total   int
	Page    int
	Filter  string
	Sort    string
	Error   string
	Loading bool
}

type EventFormPageData struct {
	Event     *models.Event
	Questions []*models.PreferenceQuestion
	Errors    map[string]string
	Error     string
	CSRFToken string
}

type EventDetailPageData struct {
	Event     *models.Event
	CSRFToken string
	Error     string
}

func NewEventWebHandlers(service events.Service, templates *template.Template) *EventWebHandlers {
	return &EventWebHandlers{
		service:   service,
		templates: templates,
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
			Error: "Failed to load events",
		})
		return
	}

	data := &EventListPageData{
		Events: eventList,
		Total:  len(eventList),
		Page:   page,
		Filter: r.URL.Query().Get("status"),
		Sort:   r.URL.Query().Get("sort"),
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

	data := &EventFormPageData{
		Event: &models.Event{
			MaxPlusOnes: 0,
		},
		Questions: []*models.PreferenceQuestion{},
		Errors:    make(map[string]string),
		CSRFToken: csrfToken,
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

	data := &EventFormPageData{
		Event:     event,
		Questions: []*models.PreferenceQuestion{},
		Errors:    make(map[string]string),
		CSRFToken: csrfToken,
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
		Event:     event,
		CSRFToken: csrfToken,
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
		ID:           id,
		Version:      version,
		Title:        existing.Title,
		Description:  existing.Description,
		StartTime:    existing.StartTime,
		EndTime:      existing.EndTime,
		Timezone:     existing.Timezone,
		Location:     existing.Location,
		MaxPlusOnes:  existing.MaxPlusOnes,
		RSVPDeadline: existing.RSVPDeadline,
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

	if deadlineStr := r.FormValue("rsvp_deadline"); deadlineStr != "" {
		deadline, err := time.Parse("2006-01-02T15:04", deadlineStr)
		if err == nil {
			event.RSVPDeadline = &deadline
		}
	} else if r.Form.Has("rsvp_deadline") {
		event.RSVPDeadline = nil
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

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "event_list.html", data); err != nil {
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

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "event_form.html", data); err != nil {
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

	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "event_detail.html", data); err != nil {
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
		Title:       title,
		StartTime:   startTime,
		Timezone:    timezone,
		MaxPlusOnes: maxPlusOnes,
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

	return event, nil
}
