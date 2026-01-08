package repositories

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupEmailQueueTest(t *testing.T) (EmailQueueRepository, func()) {
	t.Helper()

	database := setupTestDB(t)
	repo := NewEmailQueueRepository(database)

	cleanup := func() {
		if err := database.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}

	return repo, cleanup
}

func createTestEmailQueue(t *testing.T, overrides map[string]interface{}) *models.EmailQueue {
	t.Helper()

	email := &models.EmailQueue{
		ToEmail:      "test@example.com",
		Subject:      "Test Subject",
		BodyText:     "Test body text",
		Status:       models.EmailStatusPending,
		MaxAttempts:  4,
		ScheduledFor: time.Now(),
	}

	if v, ok := overrides["to_email"].(string); ok {
		email.ToEmail = v
	}
	if v, ok := overrides["to_name"].(*string); ok {
		email.ToName = v
	}
	if v, ok := overrides["subject"].(string); ok {
		email.Subject = v
	}
	if v, ok := overrides["body_text"].(string); ok {
		email.BodyText = v
	}
	if v, ok := overrides["body_html"].(*string); ok {
		email.BodyHTML = v
	}
	if v, ok := overrides["status"].(models.EmailStatus); ok {
		email.Status = v
	}
	if v, ok := overrides["max_attempts"].(int); ok {
		email.MaxAttempts = v
	}
	if v, ok := overrides["scheduled_for"].(time.Time); ok {
		email.ScheduledFor = v
	}

	return email
}

func TestEmailQueueRepository_Create(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name    string
		email   *models.EmailQueue
		wantErr bool
		errType interface{}
	}{
		{
			name:    "valid email",
			email:   createTestEmailQueue(t, nil),
			wantErr: false,
		},
		{
			name: "valid email with HTML body",
			email: createTestEmailQueue(t, map[string]interface{}{
				"body_html": stringPtr("<html><body>Test</body></html>"),
			}),
			wantErr: false,
		},
		{
			name: "valid email with to_name",
			email: createTestEmailQueue(t, map[string]interface{}{
				"to_name": stringPtr("Test User"),
			}),
			wantErr: false,
		},
		{
			name: "missing to_email",
			email: createTestEmailQueue(t, map[string]interface{}{
				"to_email": "",
			}),
			wantErr: true,
			errType: &models.ValidationError{},
		},
		{
			name: "missing subject",
			email: createTestEmailQueue(t, map[string]interface{}{
				"subject": "",
			}),
			wantErr: true,
			errType: &models.ValidationError{},
		},
		{
			name: "missing body_text",
			email: createTestEmailQueue(t, map[string]interface{}{
				"body_text": "",
			}),
			wantErr: true,
			errType: &models.ValidationError{},
		},
		{
			name: "invalid status",
			email: createTestEmailQueue(t, map[string]interface{}{
				"status": models.EmailStatus("invalid"),
			}),
			wantErr: true,
			errType: &models.ValidationError{},
		},
		{
			name: "negative max_attempts",
			email: createTestEmailQueue(t, map[string]interface{}{
				"max_attempts": -1,
			}),
			wantErr: true,
			errType: &models.ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.email)

			if tt.wantErr {
				if err == nil {
					t.Error("Create() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("Create() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
				return
			}

			if tt.email.ID == 0 {
				t.Error("Create() did not set ID")
			}

			if tt.email.CreatedAt.IsZero() {
				t.Error("Create() did not set CreatedAt")
			}
		})
	}
}

func TestEmailQueueRepository_GetByID(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email := createTestEmailQueue(t, nil)
	err := repo.Create(ctx, email)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType interface{}
	}{
		{
			name:    "existing email",
			id:      email.ID,
			wantErr: false,
		},
		{
			name:    "non-existent email",
			id:      99999,
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByID() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("GetByID() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("GetByID() unexpected error = %v", err)
				return
			}

			if result.ID != email.ID {
				t.Errorf("GetByID() ID = %d, want %d", result.ID, email.ID)
			}
			if result.ToEmail != email.ToEmail {
				t.Errorf("GetByID() ToEmail = %s, want %s", result.ToEmail, email.ToEmail)
			}
			if result.Subject != email.Subject {
				t.Errorf("GetByID() Subject = %s, want %s", result.Subject, email.Subject)
			}
		})
	}
}

func TestEmailQueueRepository_GetPending(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	pastTime := now.Add(-1 * time.Hour)
	futureTime := now.Add(1 * time.Hour)

	email1 := createTestEmailQueue(t, map[string]interface{}{
		"scheduled_for": pastTime,
		"to_email":      "past@example.com",
	})
	err := repo.Create(ctx, email1)
	if err != nil {
		t.Fatalf("Failed to create email1: %v", err)
	}

	email2 := createTestEmailQueue(t, map[string]interface{}{
		"scheduled_for": futureTime,
		"to_email":      "future@example.com",
	})
	err = repo.Create(ctx, email2)
	if err != nil {
		t.Fatalf("Failed to create email2: %v", err)
	}

	email3 := createTestEmailQueue(t, map[string]interface{}{
		"scheduled_for": pastTime,
		"status":        models.EmailStatusSent,
		"to_email":      "sent@example.com",
	})
	err = repo.Create(ctx, email3)
	if err != nil {
		t.Fatalf("Failed to create email3: %v", err)
	}

	tests := []struct {
		name      string
		maxCount  int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "get pending emails",
			maxCount:  10,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "limit to zero",
			maxCount:  0,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "limit to one",
			maxCount:  1,
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.GetPending(ctx, tt.maxCount)

			if tt.wantErr {
				if err == nil {
					t.Error("GetPending() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetPending() unexpected error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("GetPending() count = %d, want %d", len(results), tt.wantCount)
			}

			for _, email := range results {
				if email.Status != models.EmailStatusPending {
					t.Errorf("GetPending() returned non-pending email with status %s", email.Status)
				}
				if email.ScheduledFor.After(now) {
					t.Error("GetPending() returned email scheduled for future")
				}
			}
		})
	}
}

func TestEmailQueueRepository_MarkSending(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email := createTestEmailQueue(t, nil)
	err := repo.Create(ctx, email)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType interface{}
	}{
		{
			name:    "mark pending email as sending",
			id:      email.ID,
			wantErr: false,
		},
		{
			name:    "non-existent email",
			id:      99999,
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.MarkSending(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("MarkSending() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("MarkSending() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("MarkSending() unexpected error = %v", err)
				return
			}

			result, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("Failed to get email after marking: %v", err)
			}

			if result.Status != models.EmailStatusSending {
				t.Errorf("MarkSending() status = %s, want %s", result.Status, models.EmailStatusSending)
			}
		})
	}
}

func TestEmailQueueRepository_MarkSent(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email := createTestEmailQueue(t, nil)
	err := repo.Create(ctx, email)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType interface{}
	}{
		{
			name:    "mark email as sent",
			id:      email.ID,
			wantErr: false,
		},
		{
			name:    "non-existent email",
			id:      99999,
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.MarkSent(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("MarkSent() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("MarkSent() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("MarkSent() unexpected error = %v", err)
				return
			}

			result, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("Failed to get email after marking: %v", err)
			}

			if result.Status != models.EmailStatusSent {
				t.Errorf("MarkSent() status = %s, want %s", result.Status, models.EmailStatusSent)
			}
		})
	}
}

func TestEmailQueueRepository_MarkFailed(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email := createTestEmailQueue(t, nil)
	err := repo.Create(ctx, email)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	tests := []struct {
		name     string
		id       int64
		errorMsg string
		wantErr  bool
		errType  interface{}
	}{
		{
			name:     "mark email as failed",
			id:       email.ID,
			errorMsg: "SMTP connection failed",
			wantErr:  false,
		},
		{
			name:     "non-existent email",
			id:       99999,
			errorMsg: "test error",
			wantErr:  true,
			errType:  &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.MarkFailed(ctx, tt.id, tt.errorMsg)

			if tt.wantErr {
				if err == nil {
					t.Error("MarkFailed() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("MarkFailed() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("MarkFailed() unexpected error = %v", err)
				return
			}

			result, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("Failed to get email after marking: %v", err)
			}

			if result.Status != models.EmailStatusFailed {
				t.Errorf("MarkFailed() status = %s, want %s", result.Status, models.EmailStatusFailed)
			}
			if result.LastError == nil || *result.LastError != tt.errorMsg {
				t.Errorf("MarkFailed() last_error = %v, want %s", result.LastError, tt.errorMsg)
			}
		})
	}
}

func TestEmailQueueRepository_MarkCancelled(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email := createTestEmailQueue(t, nil)
	err := repo.Create(ctx, email)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType interface{}
	}{
		{
			name:    "mark email as cancelled",
			id:      email.ID,
			wantErr: false,
		},
		{
			name:    "non-existent email",
			id:      99999,
			wantErr: true,
			errType: &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.MarkCancelled(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("MarkCancelled() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("MarkCancelled() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("MarkCancelled() unexpected error = %v", err)
				return
			}

			result, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("Failed to get email after marking: %v", err)
			}

			if result.Status != models.EmailStatusCancelled {
				t.Errorf("MarkCancelled() status = %s, want %s", result.Status, models.EmailStatusCancelled)
			}
		})
	}
}

func TestEmailQueueRepository_IncrementAttempts(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email := createTestEmailQueue(t, nil)
	err := repo.Create(ctx, email)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	tests := []struct {
		name         string
		id           int64
		errorMsg     string
		wantAttempts int
		wantErr      bool
		errType      interface{}
	}{
		{
			name:         "first increment",
			id:           email.ID,
			errorMsg:     "First attempt failed",
			wantAttempts: 1,
			wantErr:      false,
		},
		{
			name:         "second increment",
			id:           email.ID,
			errorMsg:     "Second attempt failed",
			wantAttempts: 2,
			wantErr:      false,
		},
		{
			name:     "non-existent email",
			id:       99999,
			errorMsg: "test error",
			wantErr:  true,
			errType:  &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.IncrementAttempts(ctx, tt.id, tt.errorMsg)

			if tt.wantErr {
				if err == nil {
					t.Error("IncrementAttempts() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("IncrementAttempts() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("IncrementAttempts() unexpected error = %v", err)
				return
			}

			result, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("Failed to get email after increment: %v", err)
			}

			if result.Attempts != tt.wantAttempts {
				t.Errorf("IncrementAttempts() attempts = %d, want %d", result.Attempts, tt.wantAttempts)
			}
			if result.LastError == nil || *result.LastError != tt.errorMsg {
				t.Errorf("IncrementAttempts() last_error = %v, want %s", result.LastError, tt.errorMsg)
			}
			if result.LastAttemptAt == nil {
				t.Error("IncrementAttempts() did not set last_attempt_at")
			}
		})
	}
}

func TestEmailQueueRepository_Reschedule(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email := createTestEmailQueue(t, nil)
	err := repo.Create(ctx, email)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	newTime := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name         string
		id           int64
		scheduledFor time.Time
		wantErr      bool
		errType      interface{}
	}{
		{
			name:         "reschedule email",
			id:           email.ID,
			scheduledFor: newTime,
			wantErr:      false,
		},
		{
			name:         "non-existent email",
			id:           99999,
			scheduledFor: newTime,
			wantErr:      true,
			errType:      &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Reschedule(ctx, tt.id, tt.scheduledFor)

			if tt.wantErr {
				if err == nil {
					t.Error("Reschedule() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("Reschedule() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Reschedule() unexpected error = %v", err)
				return
			}

			result, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("Failed to get email after reschedule: %v", err)
			}

			if !result.ScheduledFor.Equal(tt.scheduledFor) {
				t.Errorf("Reschedule() scheduled_for = %v, want %v", result.ScheduledFor, tt.scheduledFor)
			}
		})
	}
}

func TestEmailQueueRepository_GetByStatus(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email1 := createTestEmailQueue(t, map[string]interface{}{
		"to_email": "pending1@example.com",
		"status":   models.EmailStatusPending,
	})
	err := repo.Create(ctx, email1)
	if err != nil {
		t.Fatalf("Failed to create email1: %v", err)
	}

	email2 := createTestEmailQueue(t, map[string]interface{}{
		"to_email": "pending2@example.com",
		"status":   models.EmailStatusPending,
	})
	err = repo.Create(ctx, email2)
	if err != nil {
		t.Fatalf("Failed to create email2: %v", err)
	}

	email3 := createTestEmailQueue(t, map[string]interface{}{
		"to_email": "sent@example.com",
		"status":   models.EmailStatusSent,
	})
	err = repo.Create(ctx, email3)
	if err != nil {
		t.Fatalf("Failed to create email3: %v", err)
	}

	tests := []struct {
		name      string
		status    models.EmailStatus
		limit     int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "get pending emails",
			status:    models.EmailStatusPending,
			limit:     10,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "get sent emails",
			status:    models.EmailStatusSent,
			limit:     10,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "get failed emails",
			status:    models.EmailStatusFailed,
			limit:     10,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "limit results",
			status:    models.EmailStatusPending,
			limit:     1,
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.GetByStatus(ctx, tt.status, tt.limit)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByStatus() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetByStatus() unexpected error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("GetByStatus() count = %d, want %d", len(results), tt.wantCount)
			}

			for _, email := range results {
				if email.Status != tt.status {
					t.Errorf("GetByStatus() returned email with status %s, want %s", email.Status, tt.status)
				}
			}
		})
	}
}

func TestEmailQueueRepository_GetByRecipient(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email1 := createTestEmailQueue(t, map[string]interface{}{
		"to_email": "user@example.com",
	})
	err := repo.Create(ctx, email1)
	if err != nil {
		t.Fatalf("Failed to create email1: %v", err)
	}

	email2 := createTestEmailQueue(t, map[string]interface{}{
		"to_email": "user@example.com",
	})
	err = repo.Create(ctx, email2)
	if err != nil {
		t.Fatalf("Failed to create email2: %v", err)
	}

	email3 := createTestEmailQueue(t, map[string]interface{}{
		"to_email": "other@example.com",
	})
	err = repo.Create(ctx, email3)
	if err != nil {
		t.Fatalf("Failed to create email3: %v", err)
	}

	tests := []struct {
		name      string
		email     string
		limit     int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "get emails for recipient",
			email:     "user@example.com",
			limit:     10,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "get emails for other recipient",
			email:     "other@example.com",
			limit:     10,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "no emails for recipient",
			email:     "none@example.com",
			limit:     10,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "limit results",
			email:     "user@example.com",
			limit:     1,
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.GetByRecipient(ctx, tt.email, tt.limit)

			if tt.wantErr {
				if err == nil {
					t.Error("GetByRecipient() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetByRecipient() unexpected error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("GetByRecipient() count = %d, want %d", len(results), tt.wantCount)
			}

			for _, email := range results {
				if email.ToEmail != tt.email {
					t.Errorf("GetByRecipient() returned email with to_email %s, want %s", email.ToEmail, tt.email)
				}
			}
		})
	}
}

func TestEmailQueueRepository_GetStats(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email1 := createTestEmailQueue(t, map[string]interface{}{
		"status": models.EmailStatusPending,
	})
	err := repo.Create(ctx, email1)
	if err != nil {
		t.Fatalf("Failed to create email1: %v", err)
	}

	email2 := createTestEmailQueue(t, map[string]interface{}{
		"status": models.EmailStatusPending,
	})
	err = repo.Create(ctx, email2)
	if err != nil {
		t.Fatalf("Failed to create email2: %v", err)
	}

	email3 := createTestEmailQueue(t, map[string]interface{}{
		"status": models.EmailStatusSent,
	})
	err = repo.Create(ctx, email3)
	if err != nil {
		t.Fatalf("Failed to create email3: %v", err)
	}

	email4 := createTestEmailQueue(t, map[string]interface{}{
		"status": models.EmailStatusFailed,
	})
	err = repo.Create(ctx, email4)
	if err != nil {
		t.Fatalf("Failed to create email4: %v", err)
	}

	stats, err := repo.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats() unexpected error = %v", err)
	}

	if stats.PendingCount != 2 {
		t.Errorf("GetStats() PendingCount = %d, want 2", stats.PendingCount)
	}
	if stats.SentCount != 1 {
		t.Errorf("GetStats() SentCount = %d, want 1", stats.SentCount)
	}
	if stats.FailedCount != 1 {
		t.Errorf("GetStats() FailedCount = %d, want 1", stats.FailedCount)
	}
	if stats.TotalCount != 4 {
		t.Errorf("GetStats() TotalCount = %d, want 4", stats.TotalCount)
	}
}

func TestEmailQueueRepository_UpdateStatus(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	email := createTestEmailQueue(t, nil)
	err := repo.Create(ctx, email)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	tests := []struct {
		name      string
		id        int64
		newStatus models.EmailStatus
		wantErr   bool
		errType   interface{}
	}{
		{
			name:      "update to sending",
			id:        email.ID,
			newStatus: models.EmailStatusSending,
			wantErr:   false,
		},
		{
			name:      "update to sent",
			id:        email.ID,
			newStatus: models.EmailStatusSent,
			wantErr:   false,
		},
		{
			name:      "non-existent email",
			id:        99999,
			newStatus: models.EmailStatusSent,
			wantErr:   true,
			errType:   &models.NotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateStatus(ctx, tt.id, tt.newStatus)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateStatus() expected error, got nil")
					return
				}
				if tt.errType != nil {
					if !isErrorType(err, tt.errType) {
						t.Errorf("UpdateStatus() error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateStatus() unexpected error = %v", err)
				return
			}

			result, err := repo.GetByID(ctx, tt.id)
			if err != nil {
				t.Fatalf("Failed to get email after update: %v", err)
			}

			if result.Status != tt.newStatus {
				t.Errorf("UpdateStatus() status = %s, want %s", result.Status, tt.newStatus)
			}
		})
	}
}

func TestEmailQueueRepository_ConcurrentAccess(t *testing.T) {
	repo, cleanup := setupEmailQueueTest(t)
	defer cleanup()

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		email := createTestEmailQueue(t, map[string]interface{}{
			"to_email":      "concurrent" + string(rune(i)) + "@example.com",
			"scheduled_for": time.Now().Add(-1 * time.Hour),
		})
		err := repo.Create(ctx, email)
		if err != nil {
			t.Fatalf("Failed to create test email %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup
	emailsSeen := make(map[int64]bool)
	var mu sync.Mutex
	duplicates := 0

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			emails, err := repo.GetPending(ctx, 5)
			if err != nil {
				t.Errorf("GetPending() error = %v", err)
				return
			}

			for _, email := range emails {
				mu.Lock()
				if emailsSeen[email.ID] {
					duplicates++
				}
				emailsSeen[email.ID] = true
				mu.Unlock()

				err := repo.MarkSending(ctx, email.ID)
				if err != nil {
					t.Logf("MarkSending() error = %v (expected for concurrent access)", err)
				}
			}
		}()
	}

	wg.Wait()

	if duplicates > 0 {
		t.Logf("Warning: %d duplicate emails seen in concurrent access (may indicate race condition)", duplicates)
	}
}

func isErrorType(err error, target interface{}) bool {
	switch target.(type) {
	case *models.ValidationError:
		var ve *models.ValidationError
		return sql.ErrNoRows != err && err != nil && (err.Error() == "validation error" || asValidationError(err, &ve))
	case *models.NotFoundError:
		var nfe *models.NotFoundError
		return sql.ErrNoRows == err || asNotFoundError(err, &nfe)
	case *models.ConflictError:
		var ce *models.ConflictError
		return asConflictError(err, &ce)
	default:
		return false
	}
}

func asValidationError(err error, target **models.ValidationError) bool {
	if err == nil {
		return false
	}
	ve, ok := err.(*models.ValidationError)
	if ok {
		*target = ve
		return true
	}
	return false
}

func asNotFoundError(err error, target **models.NotFoundError) bool {
	if err == nil {
		return false
	}
	nfe, ok := err.(*models.NotFoundError)
	if ok {
		*target = nfe
		return true
	}
	return false
}

func asConflictError(err error, target **models.ConflictError) bool {
	if err == nil {
		return false
	}
	ce, ok := err.(*models.ConflictError)
	if ok {
		*target = ce
		return true
	}
	return false
}
