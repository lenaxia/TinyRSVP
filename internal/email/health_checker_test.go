package email

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockHealthRepo struct {
	stats *repositories.EmailQueueStats
	err   error
}

func (m *mockHealthRepo) GetStats(ctx context.Context) (*repositories.EmailQueueStats, error) {
	return m.stats, m.err
}

func (m *mockHealthRepo) Create(ctx context.Context, email *models.EmailQueue) error {
	return nil
}

func (m *mockHealthRepo) GetByID(ctx context.Context, id int64) (*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockHealthRepo) GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockHealthRepo) GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockHealthRepo) GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *mockHealthRepo) UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error {
	return nil
}

func (m *mockHealthRepo) IncrementAttempts(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockHealthRepo) MarkSending(ctx context.Context, id int64) error {
	return nil
}

func (m *mockHealthRepo) MarkSent(ctx context.Context, id int64) error {
	return nil
}

func (m *mockHealthRepo) MarkFailed(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (m *mockHealthRepo) MarkCancelled(ctx context.Context, id int64) error {
	return nil
}

func (m *mockHealthRepo) Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error {
	return nil
}

type mockHealthSender struct {
	testErr error
}

func (m *mockHealthSender) Send(ctx context.Context, msg *SMTPMessage) error {
	return nil
}

func (m *mockHealthSender) TestConnection(ctx context.Context) error {
	return m.testErr
}

func (m *mockHealthSender) Close() error {
	return nil
}

func TestHealthChecker_Check(t *testing.T) {
	t.Run("healthy system", func(t *testing.T) {
		repo := &mockHealthRepo{
			stats: &repositories.EmailQueueStats{
				PendingCount: 50,
				SendingCount: 5,
				FailedCount:  2,
			},
		}
		sender := &mockHealthSender{}
		checker := NewHealthChecker(repo, sender)

		err := checker.Check(context.Background())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		repo := &mockHealthRepo{
			err: errors.New("database connection failed"),
		}
		sender := &mockHealthSender{}
		checker := NewHealthChecker(repo, sender)

		err := checker.Check(context.Background())
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if !errors.Is(err, errors.New("database connection failed")) && err.Error() != "database check failed: database connection failed" {
			t.Errorf("Expected database error, got: %v", err)
		}
	})

	t.Run("queue backlog too large", func(t *testing.T) {
		repo := &mockHealthRepo{
			stats: &repositories.EmailQueueStats{
				PendingCount: 1500,
			},
		}
		sender := &mockHealthSender{}
		checker := NewHealthChecker(repo, sender)

		err := checker.Check(context.Background())
		if err == nil {
			t.Error("Expected error for large backlog, got nil")
		}
	})
}

func TestHealthChecker_GetStatus(t *testing.T) {
	t.Run("healthy status", func(t *testing.T) {
		repo := &mockHealthRepo{
			stats: &repositories.EmailQueueStats{
				PendingCount: 100,
				SendingCount: 10,
				FailedCount:  5,
			},
		}
		sender := &mockHealthSender{}
		checker := NewHealthChecker(repo, sender)

		status, err := checker.GetStatus(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if !status.Healthy {
			t.Error("Expected healthy status")
		}

		if status.QueueSize != 100 {
			t.Errorf("Expected queue size 100, got %d", status.QueueSize)
		}

		if status.SendingCount != 10 {
			t.Errorf("Expected sending count 10, got %d", status.SendingCount)
		}

		if status.FailedCount != 5 {
			t.Errorf("Expected failed count 5, got %d", status.FailedCount)
		}

		if len(status.Issues) != 0 {
			t.Errorf("Expected no issues, got: %v", status.Issues)
		}
	})

	t.Run("unhealthy - large backlog", func(t *testing.T) {
		repo := &mockHealthRepo{
			stats: &repositories.EmailQueueStats{
				PendingCount: 1500,
				SendingCount: 10,
				FailedCount:  5,
			},
		}
		sender := &mockHealthSender{}
		checker := NewHealthChecker(repo, sender)

		status, err := checker.GetStatus(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if status.Healthy {
			t.Error("Expected unhealthy status")
		}

		if len(status.Issues) == 0 {
			t.Error("Expected issues to be reported")
		}

		found := false
		for _, issue := range status.Issues {
			if issue == "Queue backlog too large" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected 'Queue backlog too large' issue, got: %v", status.Issues)
		}
	})

	t.Run("unhealthy - too many failures", func(t *testing.T) {
		repo := &mockHealthRepo{
			stats: &repositories.EmailQueueStats{
				PendingCount: 50,
				SendingCount: 5,
				FailedCount:  150,
			},
		}
		sender := &mockHealthSender{}
		checker := NewHealthChecker(repo, sender)

		status, err := checker.GetStatus(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if status.Healthy {
			t.Error("Expected unhealthy status")
		}

		found := false
		for _, issue := range status.Issues {
			if issue == "Too many failed emails" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected 'Too many failed emails' issue, got: %v", status.Issues)
		}
	})

	t.Run("multiple issues", func(t *testing.T) {
		repo := &mockHealthRepo{
			stats: &repositories.EmailQueueStats{
				PendingCount: 2000,
				SendingCount: 10,
				FailedCount:  200,
			},
		}
		sender := &mockHealthSender{}
		checker := NewHealthChecker(repo, sender)

		status, err := checker.GetStatus(context.Background())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if status.Healthy {
			t.Error("Expected unhealthy status")
		}

		if len(status.Issues) != 2 {
			t.Errorf("Expected 2 issues, got %d: %v", len(status.Issues), status.Issues)
		}
	})

	t.Run("database error", func(t *testing.T) {
		repo := &mockHealthRepo{
			err: errors.New("database error"),
		}
		sender := &mockHealthSender{}
		checker := NewHealthChecker(repo, sender)

		_, err := checker.GetStatus(context.Background())
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}
