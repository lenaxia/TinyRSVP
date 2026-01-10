package web

import (
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestEventDetailTemplate(t *testing.T) {
	tmpl, err := template.New("event_detail.html").ParseFiles("event_detail.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	now := time.Now()
	endTime := now.Add(2 * time.Hour)
	deadline := now.Add(-24 * time.Hour)
	description := "Test event description"
	location := "123 Main St"

	tests := []struct {
		name      string
		data      interface{}
		wantTexts []string
	}{
		{
			name: "complete event with all fields",
			data: map[string]interface{}{
				"Event": &models.Event{
					ID:           1,
					Title:        "Test Event",
					Description:  &description,
					Location:     &location,
					StartTime:    now,
					EndTime:      &endTime,
					Timezone:     "America/Los_Angeles",
					MaxPlusOnes:  2,
					RSVPDeadline: &deadline,
					Status:       models.EventStatusPublished,
				},
				"CSRFToken": "test-csrf-token",
			},
			wantTexts: []string{
				"Test Event",
				"Test event description",
				"123 Main St",
				"Published",
				"test-csrf-token",
			},
		},
		{
			name: "minimal event with required fields only",
			data: map[string]interface{}{
				"Event": &models.Event{
					ID:        2,
					Title:     "Minimal Event",
					StartTime: now,
					Timezone:  "UTC",
					Status:    models.EventStatusDraft,
				},
				"CSRFToken": "csrf-token-2",
			},
			wantTexts: []string{
				"Minimal Event",
				"Draft",
				"csrf-token-2",
			},
		},
		{
			name: "cancelled event with reason",
			data: map[string]interface{}{
				"Event": &models.Event{
					ID:        3,
					Title:     "Cancelled Event",
					StartTime: now,
					Timezone:  "UTC",
					Status:    models.EventStatusCancelled,
				},
				"CSRFToken": "csrf-token-3",
			},
			wantTexts: []string{
				"Cancelled Event",
				"Cancelled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			if err := tmpl.Execute(&buf, tt.data); err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			output := buf.String()

			for _, want := range tt.wantTexts {
				if !strings.Contains(output, want) {
					t.Errorf("Template output missing expected text %q", want)
				}
			}

			if !strings.Contains(output, "<!DOCTYPE html>") {
				t.Error("Template output missing DOCTYPE")
			}

			if !strings.Contains(output, "<html") {
				t.Error("Template output missing html tag")
			}
		})
	}
}

func TestEventDetailTemplate_CSRFToken(t *testing.T) {
	tmpl, err := template.New("event_detail.html").ParseFiles("event_detail.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"Event": &models.Event{
			ID:        1,
			Title:     "Test",
			StartTime: time.Now(),
			Timezone:  "UTC",
			Status:    models.EventStatusPublished,
		},
		"CSRFToken": "unique-csrf-token-12345",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "unique-csrf-token-12345") {
		t.Error("CSRF token not found in template output")
	}

	if !strings.Contains(output, `name="csrf_token"`) {
		t.Error("CSRF token input field not found")
	}
}

func TestEventDetailTemplate_ActionButtons(t *testing.T) {
	tmpl, err := template.New("event_detail.html").ParseFiles("event_detail.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	tests := []struct {
		name        string
		status      models.EventStatus
		wantButtons []string
	}{
		{
			name:   "draft event shows publish and delete",
			status: models.EventStatusDraft,
			wantButtons: []string{
				"Edit",
				"Publish",
				"Delete",
			},
		},
		{
			name:   "published event shows cancel and edit",
			status: models.EventStatusPublished,
			wantButtons: []string{
				"Edit",
				"Cancel",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{
				"Event": &models.Event{
					ID:        1,
					Title:     "Test",
					StartTime: time.Now(),
					Timezone:  "UTC",
					Status:    tt.status,
				},
				"CSRFToken": "test-token",
			}

			var buf strings.Builder
			if err := tmpl.Execute(&buf, data); err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			output := buf.String()

			for _, button := range tt.wantButtons {
				if !strings.Contains(output, button) {
					t.Errorf("Expected button %q not found in output", button)
				}
			}
		})
	}
}
