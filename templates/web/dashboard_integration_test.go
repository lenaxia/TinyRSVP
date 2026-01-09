package web

import (
	"html/template"
	"os"
	"strings"
	"testing"
)

type DashboardStats struct {
	TotalEvents     int
	DraftEvents     int
	PublishedEvents int
	TotalInvites    int
	PendingInvites  int
	TotalRSVPs      int
	AcceptedRSVPs   int
	DeclinedRSVPs   int
	ResponseRate    int
}

type DashboardActivity struct {
	Icon        string
	Title       string
	Description string
	Time        string
}

type DashboardData struct {
	Stats      DashboardStats
	Activities []DashboardActivity
	Loading    bool
	Error      string
}

func TestDashboardTemplateExists(t *testing.T) {
	if _, err := os.Stat("dashboard.html"); os.IsNotExist(err) {
		t.Fatal("dashboard.html template does not exist")
	}
}

func TestDashboardTemplateValidHTML(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

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

func TestDashboardTemplateMetaTags(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	requiredMeta := []string{
		`<meta charset="UTF-8">`,
		`<meta name="viewport"`,
	}

	for _, meta := range requiredMeta {
		t.Run("meta_"+meta, func(t *testing.T) {
			if !strings.Contains(htmlContent, meta) {
				t.Errorf("Missing required meta tag: %s", meta)
			}
		})
	}
}

func TestDashboardTemplateIncludesCSS(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	requiredCSS := []string{
		"/static/css/variables.css",
		"/static/css/typography.css",
		"/static/css/colors.css",
		"/static/css/spacing.css",
		"/static/css/grid.css",
		"/static/css/buttons.css",
		"/static/css/navigation.css",
		"/static/css/dashboard.css",
	}

	for _, css := range requiredCSS {
		t.Run("includes_"+css, func(t *testing.T) {
			if !strings.Contains(htmlContent, css) {
				t.Errorf("Missing CSS file: %s", css)
			}
		})
	}
}

func TestDashboardTemplateHasDashboardLayout(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	requiredClasses := []string{
		`class="dashboard"`,
		`class="dashboard-sidebar"`,
		`class="dashboard-main"`,
		`class="dashboard-header"`,
	}

	for _, class := range requiredClasses {
		t.Run("has_"+class, func(t *testing.T) {
			if !strings.Contains(htmlContent, class) {
				t.Errorf("Missing required class: %s", class)
			}
		})
	}
}

func TestDashboardTemplateHasStatsCards(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	requiredClasses := []string{
		`class="stats-grid"`,
		`class="stats-card"`,
		`class="stats-card-title"`,
		`class="stats-card-value"`,
		`class="stats-card-subtitle"`,
	}

	for _, class := range requiredClasses {
		t.Run("has_"+class, func(t *testing.T) {
			if !strings.Contains(htmlContent, class) {
				t.Errorf("Missing required class: %s", class)
			}
		})
	}
}

func TestDashboardTemplateHasActivityFeed(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	requiredClasses := []string{
		`class="activity-feed"`,
		`class="activity-feed-title"`,
		`class="activity-item"`,
		`class="activity-item-icon"`,
		`class="activity-item-content"`,
		`class="activity-item-title"`,
		`class="activity-item-description"`,
		`class="activity-item-time"`,
	}

	for _, class := range requiredClasses {
		t.Run("has_"+class, func(t *testing.T) {
			if !strings.Contains(htmlContent, class) {
				t.Errorf("Missing required class: %s", class)
			}
		})
	}
}

func TestDashboardTemplateHasStates(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	requiredStates := []string{
		`class="empty-state"`,
		`class="loading-state"`,
		`class="error-state"`,
	}

	for _, state := range requiredStates {
		t.Run("has_"+state, func(t *testing.T) {
			if !strings.Contains(htmlContent, state) {
				t.Errorf("Missing required state: %s", state)
			}
		})
	}
}

func TestDashboardTemplateHasQuickActions(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, `class="quick-actions"`) {
		t.Error("Missing quick-actions section")
	}

	if !strings.Contains(htmlContent, `href="/events/new"`) {
		t.Error("Missing create event link")
	}
}

func TestDashboardTemplateCanParse(t *testing.T) {
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse dashboard.html: %v", err)
	}

	if tmpl == nil {
		t.Fatal("Template is nil")
	}
}

func TestDashboardTemplateRendersWithData(t *testing.T) {
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse dashboard.html: %v", err)
	}

	data := DashboardData{
		Stats: DashboardStats{
			TotalEvents:     10,
			DraftEvents:     2,
			PublishedEvents: 8,
			TotalInvites:    50,
			PendingInvites:  10,
			TotalRSVPs:      40,
			AcceptedRSVPs:   35,
			DeclinedRSVPs:   5,
			ResponseRate:    80,
		},
		Activities: []DashboardActivity{
			{
				Icon:        "📧",
				Title:       "Invitation sent",
				Description: "50 invitations sent for Summer BBQ",
				Time:        "2 hours ago",
			},
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "10") {
		t.Error("Template should render TotalEvents")
	}

	if !strings.Contains(output, "50") {
		t.Error("Template should render TotalInvites")
	}

	if !strings.Contains(output, "40") {
		t.Error("Template should render TotalRSVPs")
	}

	if !strings.Contains(output, "80%") {
		t.Error("Template should render ResponseRate")
	}

	if !strings.Contains(output, "Summer BBQ") {
		t.Error("Template should render activity description")
	}
}

func TestDashboardTemplateRendersEmptyState(t *testing.T) {
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse dashboard.html: %v", err)
	}

	data := DashboardData{
		Stats: DashboardStats{
			TotalEvents:  0,
			TotalInvites: 0,
			TotalRSVPs:   0,
		},
		Activities: []DashboardActivity{},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "empty-state") {
		t.Error("Template should render empty state when no activities")
	}

	if !strings.Contains(output, "No Recent Activity") {
		t.Error("Template should show empty state message")
	}
}

func TestDashboardTemplateRendersLoadingState(t *testing.T) {
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse dashboard.html: %v", err)
	}

	data := DashboardData{
		Loading: true,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "loading-state") {
		t.Error("Template should render loading state")
	}

	if !strings.Contains(output, "Loading dashboard") {
		t.Error("Template should show loading message")
	}
}

func TestDashboardTemplateRendersErrorState(t *testing.T) {
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse dashboard.html: %v", err)
	}

	data := DashboardData{
		Error: "Failed to load dashboard data",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "error-state") {
		t.Error("Template should render error state")
	}

	if !strings.Contains(output, "Failed to load dashboard data") {
		t.Error("Template should show error message")
	}
}

func TestDashboardTemplateHasNavigation(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	requiredLinks := []string{
		`href="/dashboard"`,
		`href="/events"`,
		`href="/invites"`,
		`href="/settings"`,
	}

	for _, link := range requiredLinks {
		t.Run("has_link_"+link, func(t *testing.T) {
			if !strings.Contains(htmlContent, link) {
				t.Errorf("Missing navigation link: %s", link)
			}
		})
	}
}

func TestDashboardTemplateAccessibility(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	accessibilityFeatures := []string{
		`lang="en"`,
		"<title>",
		"<nav",
		"<main",
		"<aside",
		"<header",
		"<section",
	}

	for _, feature := range accessibilityFeatures {
		t.Run("accessibility_"+feature, func(t *testing.T) {
			if !strings.Contains(htmlContent, feature) {
				t.Errorf("Missing accessibility feature: %s", feature)
			}
		})
	}
}

func TestDashboardTemplateHasSemanticHTML(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	semanticTags := []string{
		"<nav",
		"<main",
		"<aside",
		"<header",
		"<section",
		"<h1>",
		"<h2",
		"<h3",
		"<time",
	}

	for _, tag := range semanticTags {
		t.Run("semantic_"+tag, func(t *testing.T) {
			if !strings.Contains(htmlContent, tag) {
				t.Errorf("Missing semantic HTML tag: %s", tag)
			}
		})
	}
}

func TestDashboardTemplateUsesGoTemplating(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	templateSyntax := []string{
		"{{if",
		"{{range",
		"{{.Stats",
		"{{.Error}}",
		".Loading",
		"{{end}}",
	}

	for _, syntax := range templateSyntax {
		t.Run("template_syntax_"+syntax, func(t *testing.T) {
			if !strings.Contains(htmlContent, syntax) {
				t.Errorf("Missing Go template syntax: %s", syntax)
			}
		})
	}
}

func TestDashboardTemplateStatsFields(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	statsFields := []string{
		"{{.Stats.TotalEvents}}",
		"{{.Stats.DraftEvents}}",
		"{{.Stats.PublishedEvents}}",
		"{{.Stats.TotalInvites}}",
		"{{.Stats.PendingInvites}}",
		"{{.Stats.TotalRSVPs}}",
		"{{.Stats.AcceptedRSVPs}}",
		"{{.Stats.DeclinedRSVPs}}",
		"{{.Stats.ResponseRate}}",
	}

	for _, field := range statsFields {
		t.Run("stats_field_"+field, func(t *testing.T) {
			if !strings.Contains(htmlContent, field) {
				t.Errorf("Missing stats field: %s", field)
			}
		})
	}
}

func TestDashboardTemplateActivityFields(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	activityFields := []string{
		"{{.Icon}}",
		"{{.Title}}",
		"{{.Description}}",
		"{{.Time}}",
	}

	for _, field := range activityFields {
		t.Run("activity_field_"+field, func(t *testing.T) {
			if !strings.Contains(htmlContent, field) {
				t.Errorf("Missing activity field: %s", field)
			}
		})
	}
}

func TestDashboardTemplateRendersMultipleActivities(t *testing.T) {
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse dashboard.html: %v", err)
	}

	data := DashboardData{
		Stats: DashboardStats{
			TotalEvents:  5,
			TotalInvites: 25,
			TotalRSVPs:   20,
			ResponseRate: 80,
		},
		Activities: []DashboardActivity{
			{
				Icon:        "📧",
				Title:       "Invitation sent",
				Description: "25 invitations sent for Event A",
				Time:        "1 hour ago",
			},
			{
				Icon:        "✅",
				Title:       "RSVP received",
				Description: "John Doe accepted invitation",
				Time:        "2 hours ago",
			},
			{
				Icon:        "📅",
				Title:       "Event created",
				Description: "New event: Event A",
				Time:        "3 hours ago",
			},
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	activityCount := strings.Count(output, `class="activity-item"`)
	if activityCount != 3 {
		t.Errorf("Expected 3 activity items, got %d", activityCount)
	}

	if !strings.Contains(output, "Event A") {
		t.Error("Template should render activity descriptions")
	}

	if !strings.Contains(output, "John Doe") {
		t.Error("Template should render activity descriptions")
	}
}

func TestDashboardTemplateConditionalRendering(t *testing.T) {
	tmpl, err := template.ParseFiles("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse dashboard.html: %v", err)
	}

	tests := []struct {
		name     string
		data     DashboardData
		contains []string
		notContains []string
	}{
		{
			name: "normal_state",
			data: DashboardData{
				Stats: DashboardStats{TotalEvents: 5},
				Activities: []DashboardActivity{
					{Title: "Test Activity"},
				},
			},
			contains: []string{"stats-grid", "activity-item"},
			notContains: []string{"empty-state", "loading-state", "error-state"},
		},
		{
			name: "empty_state",
			data: DashboardData{
				Stats: DashboardStats{TotalEvents: 0},
				Activities: []DashboardActivity{},
			},
			contains: []string{"empty-state", "No Recent Activity"},
			notContains: []string{"activity-item"},
		},
		{
			name: "loading_state",
			data: DashboardData{
				Loading: true,
			},
			contains: []string{"loading-state", "Loading dashboard"},
			notContains: []string{"stats-grid", "activity-feed"},
		},
		{
			name: "error_state",
			data: DashboardData{
				Error: "Database connection failed",
			},
			contains: []string{"error-state", "Database connection failed"},
			notContains: []string{"stats-grid", "activity-feed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := tmpl.Execute(&buf, tt.data)
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

func TestDashboardTemplateHasNavLinks(t *testing.T) {
	content, err := os.ReadFile("dashboard.html")
	if err != nil {
		t.Fatalf("Failed to read dashboard.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, `class="nav-link active"`) {
		t.Error("Dashboard link should be marked as active")
	}

	if !strings.Contains(htmlContent, `class="logo"`) {
		t.Error("Missing logo link")
	}
}
