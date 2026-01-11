package web

import (
	"html/template"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func getTemplateWithFuncs() (*template.Template, error) {
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

func TestRSVPSummaryTemplateFileExists(t *testing.T) {
	if _, err := os.Stat("rsvp_summary.html"); os.IsNotExist(err) {
		t.Fatal("rsvp_summary.html template does not exist")
	}
}

func TestRSVPSummaryTemplateValidHTML(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
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

	var buf strings.Builder
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	htmlContent := buf.String()

	if !strings.Contains(htmlContent, "<!DOCTYPE html>") {
		t.Error("Missing DOCTYPE declaration")
	}

	if !strings.Contains(htmlContent, "<html") {
		t.Error("Missing html tag")
	}

	if !strings.Contains(htmlContent, "</html>") {
		t.Error("Missing closing html tag")
	}

	if !strings.Contains(htmlContent, "<head>") {
		t.Error("Missing head tag")
	}

	if !strings.Contains(htmlContent, "<body>") {
		t.Error("Missing body tag")
	}
}

func TestRSVPSummaryTemplateMetaTags(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
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

	var buf strings.Builder
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	htmlContent := buf.String()

	requiredMeta := []string{
		`<meta charset="UTF-8">`,
		`<meta name="viewport"`,
	}

	for _, meta := range requiredMeta {
		if !strings.Contains(htmlContent, meta) {
			t.Errorf("Missing required meta tag: %s", meta)
		}
	}
}

func TestRSVPSummaryTemplateIncludesCSS(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
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

	var buf strings.Builder
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	htmlContent := buf.String()

	requiredCSS := []string{
		"/static/css/variables.css",
		"/static/css/typography.css",
		"/static/css/colors.css",
		"/static/css/spacing.css",
		"/static/css/grid.css",
		"/static/css/buttons.css",
		"/static/css/rsvp_summary.css",
	}

	for _, css := range requiredCSS {
		if !strings.Contains(htmlContent, css) {
			t.Errorf("Missing CSS file: %s", css)
		}
	}
}

func TestRSVPSummaryTemplateCanParse(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
	if err != nil {
		t.Fatalf("Failed to parse rsvp_summary.html: %v", err)
	}

	if tmpl == nil {
		t.Fatal("Template is nil")
	}
}

func TestRSVPSummaryTemplateRendersWithData(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
	if err != nil {
		t.Fatalf("Failed to parse rsvp_summary.html: %v", err)
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

	var buf strings.Builder
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Test Event") {
		t.Error("Template should render event title")
	}

	if !strings.Contains(output, "50") {
		t.Error("Template should render total invites")
	}

	if !strings.Contains(output, "30") {
		t.Error("Template should render yes count")
	}
}

func TestRSVPSummaryTemplateAccessibility(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
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

	var buf strings.Builder
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	htmlContent := buf.String()

	accessibilityFeatures := []string{
		`lang="en"`,
		"<title>",
		"<main",
		"<header",
		"<section",
		"aria-label",
		"role=",
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(htmlContent, feature) {
			t.Errorf("Missing accessibility feature: %s", feature)
		}
	}
}

func TestRSVPSummaryTemplateHasSemanticHTML(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
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

	var buf strings.Builder
	err = tmpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	htmlContent := buf.String()

	semanticTags := []string{
		"<main",
		"<header",
		"<section",
		"<h1",
		"<h2",
		"<h3",
	}

	for _, tag := range semanticTags {
		if !strings.Contains(htmlContent, tag) {
			t.Errorf("Missing semantic HTML tag: %s", tag)
		}
	}
}

func TestRSVPSummaryTemplateUsesGoTemplating(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	templateSyntax := []string{
		"{{if",
		"{{range",
		"{{.EventID}}",
		"{{.Error}}",
		".Loading",
		".Stats",
		"{{end}}",
	}

	for _, syntax := range templateSyntax {
		if !strings.Contains(htmlContent, syntax) {
			t.Errorf("Missing Go template syntax: %s", syntax)
		}
	}
}

func TestRSVPSummaryTemplateStatsFields(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	statsFields := []string{
		"{{.Stats.TotalInvites}}",
		"{{.Stats.YesCount}}",
		"{{.Stats.NoCount}}",
		"{{.Stats.MaybeCount}}",
		"{{.Stats.NoResponse}}",
		"{{.Stats.TotalGuests}}",
	}

	for _, field := range statsFields {
		if !strings.Contains(htmlContent, field) {
			t.Errorf("Missing stats field: %s", field)
		}
	}
}

func TestRSVPSummaryTemplateConditionalRendering(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
	if err != nil {
		t.Fatalf("Failed to parse rsvp_summary.html: %v", err)
	}

	tests := []struct {
		name        string
		data        RSVPSummaryData
		contains    []string
		notContains []string
	}{
		{
			name: "normal_state",
			data: RSVPSummaryData{
				Event: &models.Event{
					ID:        1,
					Title:     "Test Event",
					StartTime: time.Now().Add(24 * time.Hour),
					Timezone:  "America/Los_Angeles",
					Status:    models.EventStatusPublished,
				},
				Stats: &repositories.RSVPStats{
					TotalInvites: 50,
					YesCount:     30,
					NoCount:      10,
					MaybeCount:   5,
					NoResponse:   5,
					TotalGuests:  40,
				},
				ResponseRate: 90.0,
				EventID:      1,
			},
			contains:    []string{"stats-grid", "response-rate", "chart-container"},
			notContains: []string{"rsvp-summary-error", "rsvp-summary-loading"},
		},
		{
			name: "loading_state",
			data: RSVPSummaryData{
				Loading: true,
			},
			contains:    []string{"rsvp-summary-loading", "Loading RSVP summary"},
			notContains: []string{"stats-grid", "response-rate"},
		},
		{
			name: "error_state",
			data: RSVPSummaryData{
				Error: "Failed to load RSVP data",
			},
			contains:    []string{"rsvp-summary-error", "Error Loading RSVP Summary", "Failed to load RSVP data"},
			notContains: []string{"stats-grid", "response-rate"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := tmpl.ExecuteTemplate(&buf, "base", tt.data)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			output := buf.String()

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q", expected)
				}
			}

			for _, notExpected := range tt.notContains {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output to NOT contain %q", notExpected)
				}
			}
		})
	}
}

func TestRSVPSummaryTemplateHasBackLink(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, `href="/events/{{.EventID}}"`) {
		t.Error("Missing back to event link")
	}

	if !strings.Contains(htmlContent, "Back to Event") {
		t.Error("Missing back to event text")
	}
}

func TestRSVPSummaryTemplateHasExportButton(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, "export-btn") {
		t.Error("Missing export button class")
	}

	if !strings.Contains(htmlContent, "Export") {
		t.Error("Missing export button text")
	}
}

func TestRSVPSummaryTemplateHasFilters(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	filterOptions := []string{
		`value="all"`,
		`value="yes"`,
		`value="no"`,
		`value="maybe"`,
		`value="pending"`,
	}

	for _, option := range filterOptions {
		if !strings.Contains(htmlContent, option) {
			t.Errorf("Missing filter option: %s", option)
		}
	}
}

func TestRSVPSummaryTemplateHasChartVisualization(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	chartElements := []string{
		"chart-container",
		"chart-bars",
		"chart-bar-yes",
		"chart-bar-no",
		"chart-bar-maybe",
		"chart-bar-pending",
	}

	for _, element := range chartElements {
		if !strings.Contains(htmlContent, element) {
			t.Errorf("Missing chart element: %s", element)
		}
	}
}

func TestRSVPSummaryTemplateHasResponseRateCircle(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	circleElements := []string{
		"response-rate-circle",
		"response-rate-svg",
		"<svg",
		"<circle",
		"response-rate-percentage",
	}

	for _, element := range circleElements {
		if !strings.Contains(htmlContent, element) {
			t.Errorf("Missing response rate circle element: %s", element)
		}
	}
}

func TestRSVPSummaryTemplateRendersQuestionStats(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
	if err != nil {
		t.Fatalf("Failed to parse rsvp_summary.html: %v", err)
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
		ActivePage:   "events",
		Event:         event,
		Stats:         stats,
		ResponseRate:  90.0,
		EventID:       1,
		QuestionStats: questionStats,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Dietary restrictions?") {
		t.Error("Template should render question text")
	}

	if !strings.Contains(output, "questions-grid") {
		t.Error("Template should include questions grid")
	}
}

func TestRSVPSummaryTemplateHasStatCards(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	statCards := []string{
		"stat-card-total",
		"stat-card-yes",
		"stat-card-no",
		"stat-card-maybe",
		"stat-card-pending",
		"stat-card-guests",
	}

	for _, card := range statCards {
		if !strings.Contains(htmlContent, card) {
			t.Errorf("Missing stat card: %s", card)
		}
	}
}

func TestRSVPSummaryTemplateHandlesZeroStats(t *testing.T) {
	tmpl, err := getTemplateWithFuncs()
	if err != nil {
		t.Fatalf("Failed to parse rsvp_summary.html: %v", err)
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

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "0") {
		t.Error("Template should render zero values")
	}
}

func TestRSVPSummaryTemplateHasViewInvitesLink(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, `href="/events/{{.EventID}}/invites"`) {
		t.Error("Missing view invites link")
	}

	if !strings.Contains(htmlContent, "View Invites") {
		t.Error("Missing view invites text")
	}
}

func TestRSVPSummaryTemplateResponseRateDisplay(t *testing.T) {
	content, err := os.ReadFile("rsvp_summary.html")
	if err != nil {
		t.Fatalf("Failed to read rsvp_summary.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, "{{.ResponseRate}}") {
		t.Error("Missing response rate field")
	}

	if !strings.Contains(htmlContent, "Response Rate") {
		t.Error("Missing response rate label")
	}
}
