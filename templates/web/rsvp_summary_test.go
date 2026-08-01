package web

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type RSVPSummaryData struct {
	ActivePage    string
	IsAdmin       bool
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

func parseRSVPSummaryTemplate() (*template.Template, error) {
	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}

	return template.New("rsvp_summary.html").Funcs(funcMap).ParseFiles(
		"partials/base.html",
		"partials/navigation.html",
		"rsvp_summary.html",
	)
}

func TestRSVPSummaryTemplate_ValidData(t *testing.T) {
	tmpl, err := parseRSVPSummaryTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 100,
		YesCount:     60,
		NoCount:      20,
		MaybeCount:   10,
		NoResponse:   10,
		TotalGuests:  75,
	}

	data := &RSVPSummaryData{
		ActivePage:   "events",
		Event:        event,
		Stats:        stats,
		ResponseRate: 90.0,
		EventID:      1,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Test Event") {
		t.Error("Expected template to contain event title")
	}

	if !strings.Contains(html, "100") {
		t.Error("Expected template to contain total invites count")
	}

	if !strings.Contains(html, "60") {
		t.Error("Expected template to contain yes count")
	}

	if !strings.Contains(html, "90") {
		t.Error("Expected template to contain response rate")
	}
}

func TestRSVPSummaryTemplate_EmptyStats(t *testing.T) {
	tmpl, err := parseRSVPSummaryTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 0,
		YesCount:     0,
		NoCount:      0,
		MaybeCount:   0,
		NoResponse:   0,
		TotalGuests:  0,
	}

	data := &RSVPSummaryData{
		ActivePage:   "events",
		Event:        event,
		Stats:        stats,
		ResponseRate: 0.0,
		EventID:      1,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Test Event") {
		t.Error("Expected template to contain event title")
	}

	if !strings.Contains(html, "0") {
		t.Error("Expected template to show zero counts")
	}
}

func TestRSVPSummaryTemplate_ErrorState(t *testing.T) {
	tmpl, err := parseRSVPSummaryTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := &RSVPSummaryData{
		ActivePage: "events",
		Error:      "Failed to load RSVP data",
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Failed to load RSVP data") {
		t.Error("Expected template to display error message")
	}
}

func TestRSVPSummaryTemplate_LoadingState(t *testing.T) {
	tmpl, err := parseRSVPSummaryTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := &RSVPSummaryData{
		ActivePage: "events",
		Loading:    true,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Loading") || !strings.Contains(html, "loading") {
		t.Error("Expected template to display loading state")
	}
}

func TestRSVPSummaryTemplate_WithQuestionStats(t *testing.T) {
	tmpl, err := parseRSVPSummaryTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 50,
		YesCount:     30,
		NoCount:      10,
		MaybeCount:   5,
		NoResponse:   5,
		TotalGuests:  40,
	}

	question := &models.PreferenceQuestion{
		ID:           1,
		EventID:      1,
		QuestionText: "Dietary restrictions?",
		QuestionType: models.QuestionTypeSingleChoice,
	}

	questionStats := map[int64]*QuestionStat{
		1: {
			Question: question,
			Answers: map[string]int{
				"Vegetarian": 15,
				"Vegan":      10,
				"None":       25,
			},
		},
	}

	data := &RSVPSummaryData{
		ActivePage:    "events",
		Event:         event,
		Stats:         stats,
		ResponseRate:  90.0,
		EventID:       1,
		QuestionStats: questionStats,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Dietary restrictions?") {
		t.Error("Expected template to contain question text")
	}
}

func TestRSVPSummaryTemplate_ResponseRateCalculation(t *testing.T) {
	tmpl, err := parseRSVPSummaryTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 100,
		YesCount:     50,
		NoCount:      25,
		MaybeCount:   15,
		NoResponse:   10,
		TotalGuests:  65,
	}

	data := &RSVPSummaryData{
		ActivePage:   "events",
		Event:        event,
		Stats:        stats,
		ResponseRate: 90.0,
		EventID:      1,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "90") {
		t.Error("Expected template to display response rate percentage")
	}
}

func TestRSVPSummaryTemplate_ExportButton(t *testing.T) {
	tmpl, err := parseRSVPSummaryTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 50,
		YesCount:     30,
		NoCount:      10,
		MaybeCount:   5,
		NoResponse:   5,
		TotalGuests:  40,
	}

	data := &RSVPSummaryData{
		ActivePage:   "events",
		Event:        event,
		Stats:        stats,
		ResponseRate: 90.0,
		EventID:      1,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "Export") || !strings.Contains(html, "export") {
		t.Error("Expected template to contain export button")
	}
}

func TestRSVPSummaryTemplate_FilterByResponseType(t *testing.T) {
	tmpl, err := parseRSVPSummaryTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	event := &models.Event{
		ID:        1,
		Title:     "Test Event",
		StartTime: startTime,
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusPublished,
	}

	stats := &repositories.RSVPStats{
		TotalInvites: 50,
		YesCount:     30,
		NoCount:      10,
		MaybeCount:   5,
		NoResponse:   5,
		TotalGuests:  40,
	}

	data := &RSVPSummaryData{
		ActivePage:   "events",
		Event:        event,
		Stats:        stats,
		ResponseRate: 90.0,
		EventID:      1,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "filter") && !strings.Contains(html, "Filter") {
		t.Error("Expected template to contain filter functionality")
	}
}
