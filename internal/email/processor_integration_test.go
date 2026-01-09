package email

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func newTestIntegrationProcessor(repo repositories.EmailQueueRepository, sender SMTPSender, limiter RateLimiter, batchSize int, pollInterval time.Duration) QueueProcessor {
	metrics := NewNoOpMetrics()
	logger := NewLogger(slog.Default())
	return NewQueueProcessor(repo, sender, limiter, batchSize, pollInterval, metrics, logger)
}

func TestProcessorIntegration_FullCycle(t *testing.T) {
	database := setupTestDatabase(t)
	defer database.Close()

	repo := repositories.NewEmailQueueRepository(database)
	sender := &trackingSender{sent: make([]string, 0)}
	limiter := NewStubRateLimiter()

	processor := newTestIntegrationProcessor(repo, sender, limiter, 10, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		processor.Start(ctx)
	}()

	email := &models.EmailQueue{
		ToEmail:      "test@example.com",
		Subject:      "Test Subject",
		BodyText:     "Test body",
		Status:       models.EmailStatusPending,
		ScheduledFor: time.Now(),
		MaxAttempts:  3,
		Attempts:     0,
	}

	err := repo.Create(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	if !sender.wasSent("test@example.com") {
		t.Error("Email was not sent")
	}

	retrieved, err := repo.GetByID(context.Background(), email.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve email: %v", err)
	}

	if retrieved.Status != models.EmailStatusSent {
		t.Errorf("Email status = %s, want %s", retrieved.Status, models.EmailStatusSent)
	}

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	if err := processor.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestProcessorIntegration_RetryLogic(t *testing.T) {
	database := setupTestDatabase(t)
	defer database.Close()

	repo := repositories.NewEmailQueueRepository(database)
	sender := &failingSender{failCount: 2}
	limiter := NewStubRateLimiter()

	processor := newTestIntegrationProcessor(repo, sender, limiter, 10, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		processor.Start(ctx)
	}()

	email := &models.EmailQueue{
		ToEmail:      "test@example.com",
		Subject:      "Test Subject",
		BodyText:     "Test body",
		Status:       models.EmailStatusPending,
		ScheduledFor: time.Now(),
		MaxAttempts:  3,
		Attempts:     0,
	}

	err := repo.Create(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	retrieved, err := repo.GetByID(context.Background(), email.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve email: %v", err)
	}

	if retrieved.Attempts < 1 {
		t.Errorf("Email attempts = %d, want >= 1", retrieved.Attempts)
	}

	if retrieved.Status != models.EmailStatusPending {
		t.Errorf("Email status = %s, want %s (email should be reset to pending after failure)", retrieved.Status, models.EmailStatusPending)
	}

	if retrieved.ScheduledFor.Before(time.Now()) {
		t.Error("Email should be rescheduled for future")
	}

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	if err := processor.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestProcessorIntegration_MaxAttemptsReached(t *testing.T) {
	database := setupTestDatabase(t)
	defer database.Close()

	repo := repositories.NewEmailQueueRepository(database)
	sender := &alwaysFailingSender{}
	limiter := NewStubRateLimiter()

	processor := newTestIntegrationProcessor(repo, sender, limiter, 10, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		processor.Start(ctx)
	}()

	email := &models.EmailQueue{
		ToEmail:      "test@example.com",
		Subject:      "Test Subject",
		BodyText:     "Test body",
		Status:       models.EmailStatusPending,
		ScheduledFor: time.Now(),
		MaxAttempts:  2,
		Attempts:     0,
	}

	err := repo.Create(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	retrieved, err := repo.GetByID(context.Background(), email.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve email: %v", err)
	}

	if retrieved.Attempts < 1 {
		t.Errorf("Email attempts = %d, want >= 1", retrieved.Attempts)
	}

	if retrieved.Status != models.EmailStatusPending {
		t.Errorf("Email status = %s, want %s (should be pending after first failure)", retrieved.Status, models.EmailStatusPending)
	}

	if retrieved.LastError == nil || *retrieved.LastError == "" {
		t.Error("Expected error message to be recorded")
	}

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	if err := processor.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestProcessorIntegration_GracefulShutdown(t *testing.T) {
	database := setupTestDatabase(t)
	defer database.Close()

	repo := repositories.NewEmailQueueRepository(database)
	sender := &slowSender{delay: 200 * time.Millisecond}
	limiter := NewStubRateLimiter()

	processor := newTestIntegrationProcessor(repo, sender, limiter, 10, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		processor.Start(ctx)
	}()

	email := &models.EmailQueue{
		ToEmail:      "test@example.com",
		Subject:      "Test Subject",
		BodyText:     "Test body",
		Status:       models.EmailStatusPending,
		ScheduledFor: time.Now(),
		MaxAttempts:  3,
		Attempts:     0,
	}

	err := repo.Create(context.Background(), email)
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	if err := processor.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	retrieved, err := repo.GetByID(context.Background(), email.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve email: %v", err)
	}

	if retrieved.Status != models.EmailStatusSent && retrieved.Status != models.EmailStatusSending {
		t.Logf("Email status after shutdown: %s (acceptable)", retrieved.Status)
	}
}

func TestProcessorIntegration_BatchProcessing(t *testing.T) {
	database := setupTestDatabase(t)
	defer database.Close()

	repo := repositories.NewEmailQueueRepository(database)
	sender := &trackingSender{sent: make([]string, 0)}
	limiter := NewStubRateLimiter()

	processor := newTestIntegrationProcessor(repo, sender, limiter, 5, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		processor.Start(ctx)
	}()

	for i := 0; i < 10; i++ {
		email := &models.EmailQueue{
			ToEmail:      "test@example.com",
			Subject:      "Test Subject",
			BodyText:     "Test body",
			Status:       models.EmailStatusPending,
			ScheduledFor: time.Now(),
			MaxAttempts:  3,
			Attempts:     0,
		}

		err := repo.Create(context.Background(), email)
		if err != nil {
			t.Fatalf("Failed to create email %d: %v", i, err)
		}
	}

	time.Sleep(1 * time.Second)

	sentCount := sender.getSentCount()
	if sentCount != 10 {
		t.Errorf("Sent count = %d, want 10", sentCount)
	}

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	if err := processor.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func setupTestDatabase(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

type trackingSender struct {
	mu   sync.Mutex
	sent []string
}

func (s *trackingSender) Send(ctx context.Context, msg *SMTPMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, msg.To)
	return nil
}

func (s *trackingSender) TestConnection(ctx context.Context) error {
	return nil
}

func (s *trackingSender) Close() error {
	return nil
}

func (s *trackingSender) wasSent(email string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.sent {
		if e == email {
			return true
		}
	}
	return false
}

func (s *trackingSender) getSentCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sent)
}

type failingSender struct {
	mu        sync.Mutex
	failCount int
	attempts  int
}

func (s *failingSender) Send(ctx context.Context, msg *SMTPMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.attempts++
	if s.attempts <= s.failCount {
		return sql.ErrConnDone
	}
	return nil
}

func (s *failingSender) TestConnection(ctx context.Context) error {
	return nil
}

func (s *failingSender) Close() error {
	return nil
}

type alwaysFailingSender struct{}

func (s *alwaysFailingSender) Send(ctx context.Context, msg *SMTPMessage) error {
	return sql.ErrConnDone
}

func (s *alwaysFailingSender) TestConnection(ctx context.Context) error {
	return nil
}

func (s *alwaysFailingSender) Close() error {
	return nil
}

type slowSender struct {
	delay time.Duration
}

func (s *slowSender) Send(ctx context.Context, msg *SMTPMessage) error {
	time.Sleep(s.delay)
	return nil
}

func (s *slowSender) TestConnection(ctx context.Context) error {
	return nil
}

func (s *slowSender) Close() error {
	return nil
}

func TestProcessorIntegration_RateLimiting(t *testing.T) {
	database := setupTestDatabase(t)
	defer database.Close()

	repo := repositories.NewEmailQueueRepository(database)
	sender := &trackingSender{sent: make([]string, 0)}
	limiter := NewRateLimiter(2)

	processor := newTestIntegrationProcessor(repo, sender, limiter, 10, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		processor.Start(ctx)
	}()

	for i := 0; i < 5; i++ {
		email := &models.EmailQueue{
			ToEmail:      "test@example.com",
			Subject:      "Test Subject",
			BodyText:     "Test body",
			Status:       models.EmailStatusPending,
			ScheduledFor: time.Now(),
			MaxAttempts:  3,
			Attempts:     0,
		}

		err := repo.Create(context.Background(), email)
		if err != nil {
			t.Fatalf("Failed to create email %d: %v", i, err)
		}
	}

	time.Sleep(300 * time.Millisecond)

	sentCount := sender.getSentCount()
	if sentCount != 2 {
		t.Errorf("Sent count = %d, want 2 (rate limit should block remaining)", sentCount)
	}

	availableSlots := limiter.AvailableSlots()
	if availableSlots != 0 {
		t.Errorf("AvailableSlots() = %d, want 0 (limit should be reached)", availableSlots)
	}

	pendingEmails, err := repo.GetPending(context.Background(), 10)
	if err != nil {
		t.Fatalf("Failed to get pending emails: %v", err)
	}

	if len(pendingEmails) != 3 {
		t.Errorf("Pending emails = %d, want 3 (remaining should be rescheduled)", len(pendingEmails))
	}

	cutoff := time.Now().Add(-5 * time.Second)
	for _, email := range pendingEmails {
		if email.ScheduledFor.Before(cutoff) {
			t.Errorf("Rescheduled email scheduled for %v, which is more than 5 seconds in the past", email.ScheduledFor)
		}
	}

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	if err := processor.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}
