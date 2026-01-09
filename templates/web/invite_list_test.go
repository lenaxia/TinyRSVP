package web

import (
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type InviteListData struct {
	EventID    int64
	EventTitle string
	Invites    []*models.Invite
	Total      int
	Page       int
	Stats      *repositories.InviteStats
	Filter     string
	Search     string
	Loading    bool
	Error      string
}

func getInviteListTemplate() (*template.Template, error) {
	return template.New("invite_list.html").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"div": func(a, b int) int { return a / b },
		"until": func(n int) []int {
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = i
			}
			return result
		},
	}).ParseFiles("invite_list.html")
}

func TestInviteListTemplateExists(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse invite_list.html: %v", err)
	}
	if tmpl == nil {
		t.Fatal("Template is nil")
	}
}

func TestInviteListTemplate_Structure(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	email1 := "user1@example.com"
	name1 := "User One"
	email2 := "user2@example.com"
	name2 := "User Two"
	now := time.Now()

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Invites: []*models.Invite{
			{
				ID:          1,
				EventID:     1,
				Email:       &email1,
				Name:        &name1,
				MaxPlusOnes: 2,
				Status:      models.InviteStatusSent,
				ExpiresAt:   now.Add(30 * 24 * time.Hour),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          2,
				EventID:     1,
				Email:       &email2,
				Name:        &name2,
				MaxPlusOnes: 1,
				Status:      models.InviteStatusDraft,
				ExpiresAt:   now.Add(30 * 24 * time.Hour),
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		Total: 2,
		Stats: &repositories.InviteStats{
			Total:     2,
			Draft:     1,
			Sent:      1,
			Viewed:    0,
			Responded: 0,
			Revoked:   0,
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"<head>",
		"<title>",
		"<body>",
		"invite-list",
		"invite-list-header",
		"invite-filters",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("template missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_EmptyState(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Invites:    []*models.Invite{},
		Total:      0,
		Stats: &repositories.InviteStats{
			Total: 0,
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"invite-empty",
		"No Invites Found",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("empty state missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_LoadingState(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Loading:    true,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"invite-list-loading",
		"loading-spinner",
		"Loading invites",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("loading state missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_ErrorState(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Error:      "Failed to load invites",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"invite-empty",
		"Error Loading Invites",
		"Failed to load invites",
		"Retry",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("error state missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_Filters(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Invites:    []*models.Invite{},
		Total:      0,
		Stats: &repositories.InviteStats{
			Total: 0,
		},
		Filter: "sent",
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"status-filter",
		"All Invites",
		"Draft",
		"Sent",
		"Viewed",
		"Responded",
		"Revoked",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("filters missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_Search(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Invites:    []*models.Invite{},
		Total:      0,
		Stats: &repositories.InviteStats{
			Total: 0,
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"invite-search",
		"search-input",
		"Search invites",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("search missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_InviteData(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	email := "test@example.com"
	name := "Test User"
	now := time.Now()
	sentAt := now.Add(-1 * time.Hour)

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
				SentAt:      &sentAt,
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

	html := buf.String()

	requiredElements := []string{
		"test@example.com",
		"Test User",
		"sent",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("invite data missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_BulkActions(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
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

	html := buf.String()

	requiredElements := []string{
		"bulk-actions",
		"select-all",
		"type=\"checkbox\"",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("bulk actions missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_IndividualActions(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
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

	html := buf.String()

	requiredElements := []string{
		"invite-actions",
		"Regenerate",
		"Revoke",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("individual actions missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_Stats(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Invites:    []*models.Invite{},
		Total:      10,
		Stats: &repositories.InviteStats{
			Total:     10,
			Draft:     2,
			Sent:      3,
			Viewed:    2,
			Responded: 2,
			Revoked:   1,
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"invite-stats",
		"Total",
		"Draft",
		"Sent",
		"Viewed",
		"Responded",
		"Revoked",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("stats missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_Pagination(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	invites := make([]*models.Invite, 0)
	for i := 0; i < 15; i++ {
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
		Total:      100,
		Page:       2,
		Stats: &repositories.InviteStats{
			Total: 100,
		},
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"invite-pagination",
		"pagination",
		"Previous",
		"Next",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("pagination missing required element: %s", elem)
		}
	}
}

func TestInviteListTemplate_ResponsiveClasses(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
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

	html := buf.String()

	requiredClasses := []string{
		"invite-list",
		"invite-table",
		"invite-card",
	}

	for _, class := range requiredClasses {
		if !strings.Contains(html, class) {
			t.Errorf("template missing responsive class: %s", class)
		}
	}
}

func TestInviteListTemplate_AccessibilityAttributes(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := InviteListData{
		EventID:    1,
		EventTitle: "Test Event",
		Loading:    true,
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredAttributes := []string{
		"role=",
		"aria-label=",
		"aria-live=",
	}

	for _, attr := range requiredAttributes {
		if !strings.Contains(html, attr) {
			t.Errorf("template missing accessibility attribute: %s", attr)
		}
	}
}

func TestInviteListTemplate_ExportButton(t *testing.T) {
	tmpl, err := getInviteListTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
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

	html := buf.String()

	requiredElements := []string{
		"Export",
		"export-btn",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("export functionality missing required element: %s", elem)
		}
	}
}
