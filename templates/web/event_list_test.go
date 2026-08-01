package web

import (
	"html/template"
	"strings"
	"testing"
	"time"
)

type EventListEvent struct {
	ID               int
	Title            string
	Description      string
	Location         string
	StartTime        time.Time
	Status           string
	InviteCount      int
	RSVPCount        int
	AcceptCount      int
	PrivateGuestList bool
	CreatedBy        int64
}

type EventListData struct {
	Events       []EventListEvent
	Loading      bool
	Error        string
	Filter       string
	Sort         string
	Page         int
	Total        int
	PageSize     int
	IsAdmin      bool
	CurrentUserID int64
}

func getEventListTemplate() (*template.Template, error) {
	return parseWithBase("event_list.html")
}

func TestEventListTemplateExists(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse event_list.html: %v", err)
	}
	if tmpl == nil {
		t.Fatal("Template is nil")
	}
}

func TestEventListTemplateRenders(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events: []EventListEvent{
			{
				ID:          1,
				Title:       "Test Event",
				Description: "Test Description",
				Location:    "Test Location",
				StartTime:   time.Now(),
				Status:      "published",
				InviteCount: 50,
				RSVPCount:   30,
				AcceptCount: 25,
			},
		},
		Loading: false,
		Error:   "",
		Filter:  "all",
		Sort:    "date",
		Page:    1,
		Total:   1,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	if html == "" {
		t.Fatal("Template rendered empty string")
	}
}

func TestEventListTemplateStructure(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events: []EventListEvent{
			{
				ID:          1,
				Title:       "Test Event",
				Description: "Test Description",
				Location:    "Test Location",
				StartTime:   time.Now(),
				Status:      "published",
				InviteCount: 50,
				RSVPCount:   30,
				AcceptCount: 25,
			},
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"event-list",
		"event-filters",
		"event-search",
		"event-cards",
		"event-card",
	}

	for _, element := range requiredElements {
		if !strings.Contains(html, element) {
			t.Errorf("Template missing required element: %s", element)
		}
	}
}

func TestEventListTemplateLoadingState(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Loading: true,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "event-list-loading") {
		t.Error("Template should show loading state")
	}
}

func TestEventListTemplateEmptyState(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events:  []EventListEvent{},
		Loading: false,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "event-empty") {
		t.Error("Template should show empty state when no events")
	}
}

func TestEventListTemplateErrorState(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Error: "Failed to load events",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Failed to load events") {
		t.Error("Template should display error message")
	}
}

func TestEventListTemplateEventCard(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events: []EventListEvent{
			{
				ID:          1,
				Title:       "Birthday Party",
				Description: "Join us for a celebration",
				Location:    "123 Main St",
				StartTime:   time.Now(),
				Status:      "published",
				InviteCount: 50,
				RSVPCount:   30,
				AcceptCount: 25,
			},
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	eventDetails := []string{
		"Birthday Party",
		"Join us for a celebration",
		"123 Main St",
		"published",
	}

	for _, detail := range eventDetails {
		if !strings.Contains(html, detail) {
			t.Errorf("Template should display event detail: %s", detail)
		}
	}
}

func TestEventListTemplateFilters(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events: []EventListEvent{
			{ID: 1, Title: "Test Event", Status: "published"},
		},
		Filter: "published",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	filterOptions := []string{
		"draft",
		"published",
		"archived",
	}

	for _, option := range filterOptions {
		if !strings.Contains(html, option) {
			t.Errorf("Template should include filter option: %s", option)
		}
	}
}

func TestEventListTemplateAccessibility(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events: []EventListEvent{
			{
				ID:          1,
				Title:       "Test Event",
				Description: "Test Description",
				Location:    "Test Location",
				StartTime:   time.Now(),
				Status:      "published",
			},
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	accessibilityFeatures := []string{
		"<label",
		"aria-label",
		"<main",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(html, feature) {
			t.Errorf("Template should include accessibility feature: %s", feature)
		}
	}
}

func TestEventListTemplateSemanticHTML(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events: []EventListEvent{
			{
				ID:    1,
				Title: "Test Event",
			},
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	semanticElements := []string{
		"<main",
		"<section",
		"<header",
		"<h1",
		"<h2",
	}

	for _, element := range semanticElements {
		if !strings.Contains(html, element) {
			t.Errorf("Template should use semantic HTML element: %s", element)
		}
	}
}

func TestEventListTemplatePagination(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events: []EventListEvent{
			{ID: 1, Title: "Event 1"},
		},
		Page:     2,
		Total:    50,
		PageSize: 10,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "pagination") {
		t.Error("Template should include pagination")
	}
}

func TestEventListTemplateMultipleEvents(t *testing.T) {
	tmpl, err := getEventListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := EventListData{
		Events: []EventListEvent{
			{ID: 1, Title: "Event 1", Status: "published"},
			{ID: 2, Title: "Event 2", Status: "draft"},
			{ID: 3, Title: "Event 3", Status: "archived"},
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	for _, event := range data.Events {
		if !strings.Contains(html, event.Title) {
			t.Errorf("Template should display event: %s", event.Title)
		}
	}
}
