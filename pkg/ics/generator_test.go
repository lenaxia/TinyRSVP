package ics

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	location := "123 Main St"
	description := "Test event description"
	endTime := time.Date(2026, 6, 15, 20, 0, 0, 0, time.UTC)

	event := &models.Event{
		ID:          1,
		Title:       "Test Event",
		Description: &description,
		Location:    &location,
		StartTime:   time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC),
		EndTime:     &endTime,
		Timezone:    "America/Los_Angeles",
		ICSSequence: 0,
	}

	icsData, err := gen.Generate(event, "https://example.com/rsvp/token123")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	icsStr := string(icsData)

	requiredFields := []string{
		"BEGIN:VCALENDAR",
		"END:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//TinyRSVP//EN",
		"METHOD:REQUEST",
		"BEGIN:VEVENT",
		"END:VEVENT",
		"UID:1@tinyrsvp",
		"SUMMARY:Test Event",
		"LOCATION:123 Main St",
		"DESCRIPTION:",
		"STATUS:CONFIRMED",
		"SEQUENCE:0",
	}

	for _, field := range requiredFields {
		if !strings.Contains(icsStr, field) {
			t.Errorf("ICS missing required field: %s", field)
		}
	}
}

func TestGenerator_Generate_MinimalEvent(t *testing.T) {
	gen := NewGenerator()

	event := &models.Event{
		ID:          2,
		Title:       "Minimal Event",
		StartTime:   time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC),
		Timezone:    "UTC",
		ICSSequence: 0,
	}

	icsData, err := gen.Generate(event, "https://example.com/rsvp/abc")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	icsStr := string(icsData)

	if !strings.Contains(icsStr, "BEGIN:VCALENDAR") {
		t.Error("Missing BEGIN:VCALENDAR")
	}

	if !strings.Contains(icsStr, "SUMMARY:Minimal Event") {
		t.Error("Missing event title")
	}

	if !strings.Contains(icsStr, "UID:2@tinyrsvp") {
		t.Error("Missing UID")
	}
}

func TestGenerator_Generate_WithEndTime(t *testing.T) {
	gen := NewGenerator()

	endTime := time.Date(2026, 6, 15, 21, 0, 0, 0, time.UTC)
	event := &models.Event{
		ID:          3,
		Title:       "Event With End",
		StartTime:   time.Date(2026, 6, 15, 19, 0, 0, 0, time.UTC),
		EndTime:     &endTime,
		Timezone:    "America/New_York",
		ICSSequence: 0,
	}

	icsData, err := gen.Generate(event, "https://example.com/rsvp/xyz")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	icsStr := string(icsData)

	if !strings.Contains(icsStr, "DTSTART") {
		t.Error("Missing DTSTART")
	}

	if !strings.Contains(icsStr, "DTEND") {
		t.Error("Missing DTEND")
	}
}

func TestGenerator_Generate_SpecialCharacters(t *testing.T) {
	gen := NewGenerator()

	location := "Room 123, Building A; Main Campus"
	description := "Event with special chars: comma, semicolon; backslash\\ and newline\ntext"

	event := &models.Event{
		ID:          4,
		Title:       "Special, Chars; Event\\Test",
		Description: &description,
		Location:    &location,
		StartTime:   time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC),
		Timezone:    "UTC",
		ICSSequence: 1,
	}

	icsData, err := gen.Generate(event, "https://example.com/rsvp/special")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	icsStr := string(icsData)

	if !strings.Contains(icsStr, "\\,") {
		t.Error("Comma should be escaped")
	}

	if !strings.Contains(icsStr, "\\;") {
		t.Error("Semicolon should be escaped")
	}

	if !strings.Contains(icsStr, "\\\\") {
		t.Error("Backslash should be escaped")
	}

	if !strings.Contains(icsStr, "\\n") {
		t.Error("Newline should be escaped")
	}
}

func TestGenerator_Generate_IncludesRSVPURL(t *testing.T) {
	gen := NewGenerator()

	description := "Original description"
	event := &models.Event{
		ID:          5,
		Title:       "URL Test Event",
		Description: &description,
		StartTime:   time.Date(2026, 9, 1, 14, 0, 0, 0, time.UTC),
		Timezone:    "UTC",
		ICSSequence: 0,
	}

	rsvpURL := "https://example.com/rsvp/urltest123"
	icsData, err := gen.Generate(event, rsvpURL)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	icsStr := string(icsData)

	if !strings.Contains(icsStr, rsvpURL) {
		t.Error("ICS should contain RSVP URL in description")
	}

	if !strings.Contains(icsStr, "Original description") {
		t.Error("ICS should contain original description")
	}
}

func TestGenerator_Generate_Alarm(t *testing.T) {
	gen := NewGenerator()

	event := &models.Event{
		ID:          6,
		Title:       "Alarm Test",
		StartTime:   time.Date(2026, 10, 1, 16, 0, 0, 0, time.UTC),
		Timezone:    "UTC",
		ICSSequence: 0,
	}

	icsData, err := gen.Generate(event, "https://example.com/rsvp/alarm")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	icsStr := string(icsData)

	alarmFields := []string{
		"BEGIN:VALARM",
		"END:VALARM",
		"TRIGGER:-PT24H",
		"ACTION:DISPLAY",
	}

	for _, field := range alarmFields {
		if !strings.Contains(icsStr, field) {
			t.Errorf("ICS missing alarm field: %s", field)
		}
	}
}

func TestGenerator_Generate_ValidICSFormat(t *testing.T) {
	gen := NewGenerator()

	event := &models.Event{
		ID:          7,
		Title:       "Format Test",
		StartTime:   time.Date(2026, 11, 1, 18, 0, 0, 0, time.UTC),
		Timezone:    "UTC",
		ICSSequence: 0,
	}

	icsData, err := gen.Generate(event, "https://example.com/rsvp/format")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	lines := bytes.Split(icsData, []byte("\r\n"))

	if len(lines) < 5 {
		t.Errorf("ICS should have multiple lines, got %d", len(lines))
	}

	firstLine := string(lines[0])
	if firstLine != "BEGIN:VCALENDAR" {
		t.Errorf("First line should be BEGIN:VCALENDAR, got %s", firstLine)
	}

	lastLine := string(lines[len(lines)-1])
	if lastLine != "END:VCALENDAR" && lastLine != "" {
		t.Errorf("Last non-empty line should be END:VCALENDAR, got %s", lastLine)
	}
}

func TestEscapeICS(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special chars",
			input: "simple text",
			want:  "simple text",
		},
		{
			name:  "comma",
			input: "text, with comma",
			want:  "text\\, with comma",
		},
		{
			name:  "semicolon",
			input: "text; with semicolon",
			want:  "text\\; with semicolon",
		},
		{
			name:  "backslash",
			input: "text\\with\\backslash",
			want:  "text\\\\with\\\\backslash",
		},
		{
			name:  "newline",
			input: "text\nwith\nnewline",
			want:  "text\\nwith\\nnewline",
		},
		{
			name:  "multiple special chars",
			input: "text, with; multiple\\ special\nchars",
			want:  "text\\, with\\; multiple\\\\ special\\nchars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeICS(tt.input)
			if got != tt.want {
				t.Errorf("escapeICS() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatICSTime(t *testing.T) {
	testTime := time.Date(2026, 6, 15, 18, 30, 45, 0, time.UTC)
	formatted := formatICSTime(testTime)

	expected := "20260615T183045Z"
	if formatted != expected {
		t.Errorf("formatICSTime() = %s, want %s", formatted, expected)
	}
}
