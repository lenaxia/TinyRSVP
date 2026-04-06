package handlers

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestRSVPPage_TemplateRendersThemeData(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"
	name := "Test User"

	data := &RSVPPageData{
		Event: &models.Event{
			ID:        1,
			Title:     "Wedding Reception",
			StartTime: startTime,
			Timezone:  "America/Los_Angeles",
			Status:    models.EventStatusPublished,
		},
		Invite: &models.Invite{
			ID:          1,
			EventID:     1,
			Email:       &email,
			Name:        &name,
			MaxPlusOnes: 2,
		},
		Questions:      []*QuestionWithOptions{},
		Token:          "test-token",
		LocalStartTime: "Monday, January 13, 2026 at 6:00 PM PST",
		ThemeCategory:  "card",
		ThemeImageURL:  "/static/images/themes/wedding-elegance-header.svg",
		ThemeColor: template.HTML(`<style>
[data-event-theme] {
		  --theme-primary: #f4c2c2 !important;
}
[data-theme="dark"][data-event-theme] {
		  --theme-primary: #f4c2c2 !important;
}
</style>`),
		CSRFToken: "test-csrf-token",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, `data-event-theme="card"`) {
		t.Error("Template should include data-event-theme attribute")
	}

	if !strings.Contains(html, `/static/css/themes/card.css`) {
		t.Error("Template should include theme-specific CSS")
	}

	if !strings.Contains(html, `--theme-primary: #f4c2c2`) {
		t.Error("Template should include custom color override")
	}

	if !strings.Contains(html, `/static/images/themes/wedding-elegance-header.svg`) {
		t.Error("Template should include theme header image")
	}

	if !strings.Contains(html, `class="theme-header-image" src="/static/images/themes/wedding-elegance-header.svg"`) {
		t.Error("Template should render theme image tag")
	}
}

func TestRSVPPage_TemplateRendersWithoutTheme(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"

	data := &RSVPPageData{
		Event: &models.Event{
			ID:        1,
			Title:     "Simple Event",
			StartTime: startTime,
			Timezone:  "America/Los_Angeles",
			Status:    models.EventStatusPublished,
		},
		Invite: &models.Invite{
			ID:      1,
			EventID: 1,
			Email:   &email,
		},
		Questions:      []*QuestionWithOptions{},
		Token:          "test-token",
		LocalStartTime: "Monday, January 13, 2026 at 6:00 PM PST",
		ThemeCategory:  "",
		ThemeImageURL:  "",
		ThemeColor:     "",
		CSRFToken:      "test-csrf-token",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if strings.Contains(html, `data-event-theme=`) {
		t.Error("Template should not include data-event-theme when no theme")
	}

	if strings.Contains(html, `/static/css/themes/`) {
		t.Error("Template should not include theme-specific CSS when no theme")
	}

	if strings.Contains(html, `--theme-primary:`) {
		t.Error("Template should not include custom color when not provided")
	}

	if strings.Contains(html, `class="theme-header-image"`) {
		t.Error("Template should not render theme image section when no image")
	}
}

func TestRSVPPage_TemplateRendersPlainTheme(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"

	data := &RSVPPageData{
		Event: &models.Event{
			ID:        1,
			Title:     "Plain Event",
			StartTime: startTime,
			Timezone:  "America/Los_Angeles",
			Status:    models.EventStatusPublished,
		},
		Invite: &models.Invite{
			ID:      1,
			EventID: 1,
			Email:   &email,
		},
		Questions:      []*QuestionWithOptions{},
		Token:          "test-token",
		LocalStartTime: "Monday, January 13, 2026 at 6:00 PM PST",
		ThemeCategory:  "plain",
		ThemeImageURL:  "",
		ThemeColor:     "",
		CSRFToken:      "test-csrf-token",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, `data-event-theme="plain"`) {
		t.Error("Template should include data-event-theme for plain theme")
	}

	if !strings.Contains(html, `/static/css/themes/plain.css`) {
		t.Error("Template should include plain theme CSS")
	}

	if strings.Contains(html, `class="theme-header-image"`) {
		t.Error("Plain theme should not render image section")
	}
}

func TestRSVPPage_TemplateRendersCustomImageOnly(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"

	data := &RSVPPageData{
		Event: &models.Event{
			ID:        1,
			Title:     "Event with Custom Image",
			StartTime: startTime,
			Timezone:  "America/Los_Angeles",
			Status:    models.EventStatusPublished,
		},
		Invite: &models.Invite{
			ID:      1,
			EventID: 1,
			Email:   &email,
		},
		Questions:      []*QuestionWithOptions{},
		Token:          "test-token",
		LocalStartTime: "Monday, January 13, 2026 at 6:00 PM PST",
		ThemeCategory:  "card",
		ThemeImageURL:  "/uploads/my-custom-image.jpg",
		ThemeColor:     "",
		CSRFToken:      "test-csrf-token",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, `/uploads/my-custom-image.jpg`) {
		t.Error("Template should render custom image URL")
	}

	if strings.Contains(html, `--theme-primary:`) {
		t.Error("Template should not include color override when not provided")
	}
}

func TestRSVPPage_TemplateRendersCustomColorOnly(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/web/rsvp_page.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	startTime := time.Now().Add(24 * time.Hour)
	email := "test@example.com"

	data := &RSVPPageData{
		Event: &models.Event{
			ID:        1,
			Title:     "Event with Custom Color",
			StartTime: startTime,
			Timezone:  "America/Los_Angeles",
			Status:    models.EventStatusPublished,
		},
		Invite: &models.Invite{
			ID:      1,
			EventID: 1,
			Email:   &email,
		},
		Questions:      []*QuestionWithOptions{},
		Token:          "test-token",
		LocalStartTime: "Monday, January 13, 2026 at 6:00 PM PST",
		ThemeCategory:  "card",
		ThemeImageURL:  "",
		ThemeColor: template.HTML(`<style>
[data-event-theme] {
		  --theme-primary: #123456 !important;
}
[data-theme="dark"][data-event-theme] {
		  --theme-primary: #123456 !important;
}
</style>`),
		CSRFToken: "test-csrf-token",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, `--theme-primary: #123456`) {
		t.Error("Template should include custom color override")
	}

	if strings.Contains(html, `class="theme-header-image"`) {
		t.Error("Template should not render image section when no image URL")
	}
}
