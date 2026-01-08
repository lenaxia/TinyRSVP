package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEmailQueue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		email   *EmailQueue
		wantErr bool
	}{
		{
			name: "valid email",
			email: &EmailQueue{
				ToEmail:      "test@example.com",
				Subject:      "Test Subject",
				BodyText:     "Test body",
				Status:       EmailStatusPending,
				MaxAttempts:  4,
				ScheduledFor: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing to_email",
			email: &EmailQueue{
				Subject:      "Test Subject",
				BodyText:     "Test body",
				Status:       EmailStatusPending,
				MaxAttempts:  4,
				ScheduledFor: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing subject",
			email: &EmailQueue{
				ToEmail:      "test@example.com",
				BodyText:     "Test body",
				Status:       EmailStatusPending,
				MaxAttempts:  4,
				ScheduledFor: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing body_text",
			email: &EmailQueue{
				ToEmail:      "test@example.com",
				Subject:      "Test Subject",
				Status:       EmailStatusPending,
				MaxAttempts:  4,
				ScheduledFor: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			email: &EmailQueue{
				ToEmail:      "test@example.com",
				Subject:      "Test Subject",
				BodyText:     "Test body",
				Status:       "invalid",
				MaxAttempts:  4,
				ScheduledFor: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "negative max_attempts",
			email: &EmailQueue{
				ToEmail:      "test@example.com",
				Subject:      "Test Subject",
				BodyText:     "Test body",
				Status:       EmailStatusPending,
				MaxAttempts:  -1,
				ScheduledFor: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.email.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailQueue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailQueue_SetAttachments(t *testing.T) {
	email := &EmailQueue{}
	attachments := []EmailAttachment{
		{
			Filename:    "test.ics",
			ContentType: "text/calendar",
			Content:     []byte("test content"),
		},
	}

	err := email.SetAttachments(attachments)
	if err != nil {
		t.Fatalf("SetAttachments() error = %v", err)
	}

	if email.Attachments == nil {
		t.Fatal("Attachments should not be nil")
	}

	retrieved, err := email.GetAttachments()
	if err != nil {
		t.Fatalf("GetAttachments() error = %v", err)
	}

	if len(retrieved) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(retrieved))
	}

	if retrieved[0].Filename != "test.ics" {
		t.Errorf("Expected filename 'test.ics', got '%s'", retrieved[0].Filename)
	}

	if retrieved[0].ContentType != "text/calendar" {
		t.Errorf("Expected content type 'text/calendar', got '%s'", retrieved[0].ContentType)
	}

	if string(retrieved[0].Content) != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", string(retrieved[0].Content))
	}
}

func TestEmailQueue_GetAttachments_Empty(t *testing.T) {
	email := &EmailQueue{}

	attachments, err := email.GetAttachments()
	if err != nil {
		t.Fatalf("GetAttachments() error = %v", err)
	}

	if len(attachments) != 0 {
		t.Errorf("Expected 0 attachments, got %d", len(attachments))
	}
}

func TestEmailQueue_GetAttachments_InvalidJSON(t *testing.T) {
	email := &EmailQueue{
		Attachments: []byte("invalid json"),
	}

	_, err := email.GetAttachments()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestEmailStatus_Valid(t *testing.T) {
	tests := []struct {
		name   string
		status EmailStatus
		want   bool
	}{
		{"pending", EmailStatusPending, true},
		{"sending", EmailStatusSending, true},
		{"sent", EmailStatusSent, true},
		{"failed", EmailStatusFailed, true},
		{"cancelled", EmailStatusCancelled, true},
		{"invalid", "invalid", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.want {
				t.Errorf("EmailStatus.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailAttachment_MarshalJSON(t *testing.T) {
	attachment := EmailAttachment{
		Filename:    "test.txt",
		ContentType: "text/plain",
		Content:     []byte("hello world"),
	}

	data, err := json.Marshal(attachment)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded EmailAttachment
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Filename != attachment.Filename {
		t.Errorf("Filename = %v, want %v", decoded.Filename, attachment.Filename)
	}

	if decoded.ContentType != attachment.ContentType {
		t.Errorf("ContentType = %v, want %v", decoded.ContentType, attachment.ContentType)
	}

	if string(decoded.Content) != string(attachment.Content) {
		t.Errorf("Content = %v, want %v", string(decoded.Content), string(attachment.Content))
	}
}
