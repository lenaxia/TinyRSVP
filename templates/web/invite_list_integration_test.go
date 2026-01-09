package web

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestInviteListTemplateFileExists(t *testing.T) {
	if _, err := os.Stat("invite_list.html"); os.IsNotExist(err) {
		t.Fatal("invite_list.html template does not exist")
	}
}

func TestInviteListTemplateValidHTML(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
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

func TestInviteListTemplateMetaTags(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

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

func TestInviteListTemplateIncludesCSS(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	requiredCSS := []string{
		"/static/css/variables.css",
		"/static/css/typography.css",
		"/static/css/colors.css",
		"/static/css/spacing.css",
		"/static/css/grid.css",
		"/static/css/buttons.css",
		"/static/css/forms.css",
		"/static/css/navigation.css",
		"/static/css/invite_list.css",
	}

	for _, css := range requiredCSS {
		if !strings.Contains(htmlContent, css) {
			t.Errorf("Missing CSS file: %s", css)
		}
	}
}

func TestInviteListTemplateCanParse(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse invite_list.html: %v", err)
	}

	if tmpl == nil {
		t.Fatal("Template is nil")
	}
}

func TestInviteListTemplateRendersWithData(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse invite_list.html: %v", err)
	}

	email := "test@example.com"
	name := "Test User"
	now := time.Now()

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Invites: []*models.Invite{
			{
				ID:          1,
				EventID:     1,
				Email:       &email,
				Name:        &name,
				MaxPlusOnes: 2,
				Status:      models.InviteStatusSent,
				ExpiresAt:   now.Add(30 * 24 * time.Hour),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		Total: 1,
		Stats: &repositories.InviteStats{
			Total: 1,
			Sent:  1,
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Test Event") {
		t.Error("Template should render event title")
	}

	if !strings.Contains(output, "test@example.com") {
		t.Error("Template should render invite email")
	}

	if !strings.Contains(output, "Test User") {
		t.Error("Template should render invite name")
	}
}

func TestInviteListTemplateAccessibility(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

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

func TestInviteListTemplateHasSemanticHTML(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	semanticTags := []string{
		"<main",
		"<header",
		"<section",
		"<h1",
		"<h2",
		"<table",
		"<thead",
		"<tbody",
		"<article",
		"<time",
	}

	for _, tag := range semanticTags {
		if !strings.Contains(htmlContent, tag) {
			t.Errorf("Missing semantic HTML tag: %s", tag)
		}
	}
}

func TestInviteListTemplateUsesGoTemplating(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	templateSyntax := []string{
		"{{if",
		"{{range",
		"{{.EventID}}",
		"{{.EventTitle}}",
		"{{.Error}}",
		".Loading",
		".Invites",
		"{{.Stats",
		"{{end}}",
	}

	for _, syntax := range templateSyntax {
		if !strings.Contains(htmlContent, syntax) {
			t.Errorf("Missing Go template syntax: %s", syntax)
		}
	}
}

func TestInviteListTemplateInviteFields(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	inviteFields := []string{
		"{{.ID}}",
		"{{.Name}}",
		"{{.Email}}",
		"{{.Status}}",
		"{{.MaxPlusOnes}}",
		"{{.SentAt",
	}

	for _, field := range inviteFields {
		if !strings.Contains(htmlContent, field) {
			t.Errorf("Missing invite field: %s", field)
		}
	}
}

func TestInviteListTemplateStatsFields(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	statsFields := []string{
		"{{.Stats.Total}}",
		"{{.Stats.Draft}}",
		"{{.Stats.Sent}}",
		"{{.Stats.Viewed}}",
		"{{.Stats.Responded}}",
		"{{.Stats.Revoked}}",
	}

	for _, field := range statsFields {
		if !strings.Contains(htmlContent, field) {
			t.Errorf("Missing stats field: %s", field)
		}
	}
}

func TestInviteListTemplateRendersMultipleInvites(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse invite_list.html: %v", err)
	}

	invites := make([]*models.Invite, 0)
	for i := 0; i < 3; i++ {
		email := "user@example.com"
		name := "User"
		invites = append(invites, &models.Invite{
			ID:          int64(i + 1),
			EventID:     1,
			Email:       &email,
			Name:        &name,
			MaxPlusOnes: 0,
			Status:      models.InviteStatusSent,
			ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Invites:    invites,
		Total:      3,
		Stats: &repositories.InviteStats{
			Total: 3,
			Sent:  3,
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	inviteRowCount := strings.Count(output, `class="invite-table-row"`)
	if inviteRowCount != 3 {
		t.Errorf("Expected 3 invite table rows, got %d", inviteRowCount)
	}

	inviteCardCount := strings.Count(output, `class="invite-card"`)
	if inviteCardCount != 3 {
		t.Errorf("Expected 3 invite cards, got %d", inviteCardCount)
	}
}

func TestInviteListTemplateConditionalRendering(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse invite_list.html: %v", err)
	}

	tests := []struct {
		name        string
		data        InviteListData
		contains    []string
		notContains []string
	}{
		{
			name: "normal_state",
			data: InviteListData{
				EventID:    1,
				EventTitle: "Test Event",
				Invites: []*models.Invite{
					{
						ID:          1,
						EventID:     1,
						MaxPlusOnes: 0,
						Status:      models.InviteStatusSent,
						ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				},
				Total: 1,
				Stats: &repositories.InviteStats{Total: 1},
			},
			contains:    []string{"invite-table", "invite-card"},
			notContains: []string{"invite-empty", "invite-list-loading"},
		},
		{
			name: "empty_state",
			data: InviteListData{
				EventID:    1,
				EventTitle: "Test Event",
				Invites:    []*models.Invite{},
				Total:      0,
				Stats:      &repositories.InviteStats{Total: 0},
			},
			contains:    []string{"invite-empty", "No Invites Found"},
			notContains: []string{"invite-table-row", "invite-card"},
		},
		{
			name: "loading_state",
			data: InviteListData{
				EventID:    1,
				EventTitle: "Test Event",
				Loading:    true,
			},
			contains:    []string{"invite-list-loading", "Loading invites"},
			notContains: []string{"invite-table", "invite-card"},
		},
		{
			name: "error_state",
			data: InviteListData{
				EventID:    1,
				EventTitle: "Test Event",
				Error:      "Failed to load invites",
			},
			contains:    []string{"invite-empty", "Error Loading Invites", "Failed to load invites"},
			notContains: []string{"invite-table", "invite-card"},
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

func TestInviteListTemplateHasBackLink(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, `href="/events/{{.EventID}}"`) {
		t.Error("Missing back to event link")
	}

	if !strings.Contains(htmlContent, "Back to Event") {
		t.Error("Missing back to event text")
	}
}

func TestInviteListTemplateHasCreateInviteLink(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, `href="/events/{{.EventID}}/invites/new"`) {
		t.Error("Missing create invite link")
	}

	if !strings.Contains(htmlContent, "Create Invite") {
		t.Error("Missing create invite text")
	}
}

func TestInviteListTemplateHasFilters(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	filterOptions := []string{
		`value="all"`,
		`value="draft"`,
		`value="sent"`,
		`value="viewed"`,
		`value="responded"`,
		`value="revoked"`,
	}

	for _, option := range filterOptions {
		if !strings.Contains(htmlContent, option) {
			t.Errorf("Missing filter option: %s", option)
		}
	}
}

func TestInviteListTemplateHasActionButtons(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	actionButtons := []string{
		"Regenerate",
		"Revoke",
		"Export",
		"Send Selected",
		"Revoke Selected",
	}

	for _, button := range actionButtons {
		if !strings.Contains(htmlContent, button) {
			t.Errorf("Missing action button: %s", button)
		}
	}
}

func TestInviteListTemplateHasBulkSelection(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	bulkSelectionElements := []string{
		`id="select-all"`,
		`type="checkbox"`,
		"Select All",
		"bulk-actions",
	}

	for _, element := range bulkSelectionElements {
		if !strings.Contains(htmlContent, element) {
			t.Errorf("Missing bulk selection element: %s", element)
		}
	}
}

func TestInviteListTemplateHasTableAndCards(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	if !strings.Contains(htmlContent, `class="invite-table"`) {
		t.Error("Missing table for desktop view")
	}

	if !strings.Contains(htmlContent, `class="invite-cards"`) {
		t.Error("Missing cards for mobile view")
	}

	if !strings.Contains(htmlContent, `class="invite-card"`) {
		t.Error("Missing individual card class")
	}
}

func TestInviteListTemplateRendersAllStatuses(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse invite_list.html: %v", err)
	}

	statuses := []models.InviteStatus{
		models.InviteStatusDraft,
		models.InviteStatusSent,
		models.InviteStatusViewed,
		models.InviteStatusResponded,
		models.InviteStatusRevoked,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			email := "test@example.com"
			name := "Test User"
			now := time.Now()

			data := InviteListData{
				EventID:    1,
				EventTitle: "Test Event",
				Invites: []*models.Invite{
					{
						ID:          1,
						EventID:     1,
						Email:       &email,
						Name:        &name,
						MaxPlusOnes: 0,
						Status:      status,
						ExpiresAt:   now.Add(30 * 24 * time.Hour),
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				},
				Total: 1,
				Stats: &repositories.InviteStats{Total: 1},
			}

			var buf strings.Builder
			err := tmpl.Execute(&buf, data)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, string(status)) {
				t.Errorf("Template should render status: %s", status)
			}

			expectedClass := "invite-status-" + string(status)
			if !strings.Contains(output, expectedClass) {
				t.Errorf("Template should include status class: %s", expectedClass)
			}
		})
	}
}

func TestInviteListTemplateDataAttributes(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	dataAttributes := []string{
		"data-invite-id",
		"data-action",
	}

	for _, attr := range dataAttributes {
		if !strings.Contains(htmlContent, attr) {
			t.Errorf("Missing data attribute: %s", attr)
		}
	}
}

func TestInviteListTemplateHandlesNullValues(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse invite_list.html: %v", err)
	}

	now := time.Now()

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Invites: []*models.Invite{
			{
				ID:          1,
				EventID:     1,
				Email:       nil,
				Name:        nil,
				MaxPlusOnes: 0,
				Status:      models.InviteStatusDraft,
				SentAt:      nil,
				ExpiresAt:   now.Add(30 * 24 * time.Hour),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		Total: 1,
		Stats: &repositories.InviteStats{Total: 1},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template with nil values: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "—") {
		t.Error("Template should render placeholder for null values")
	}
}

func TestInviteListTemplateHasPaginationLogic(t *testing.T) {
	content, err := os.ReadFile("invite_list.html")
	if err != nil {
		t.Fatalf("Failed to read invite_list.html: %v", err)
	}

	htmlContent := string(content)

	paginationElements := []string{
		"{{if gt .Total 50}}",
		"{{if gt .Page 1}}",
		"{{if lt .Page",
		"pagination-prev",
		"pagination-next",
		"?page=",
	}

	for _, element := range paginationElements {
		if !strings.Contains(htmlContent, element) {
			t.Errorf("Missing pagination element: %s", element)
		}
	}
}
