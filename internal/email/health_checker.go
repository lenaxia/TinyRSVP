package email

import (
	"context"
	"fmt"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
)

type HealthChecker struct {
	repo   repositories.EmailQueueRepository
	sender SMTPSender
}

func NewHealthChecker(repo repositories.EmailQueueRepository, sender SMTPSender) *HealthChecker {
	return &HealthChecker{
		repo:   repo,
		sender: sender,
	}
}

func (h *HealthChecker) Check(ctx context.Context) error {
	stats, err := h.repo.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("database check failed: %w", err)
	}

	if stats.PendingCount > 1000 {
		return fmt.Errorf("queue backlog too large: %d pending", stats.PendingCount)
	}

	return nil
}

func (h *HealthChecker) GetStatus(ctx context.Context) (*HealthStatus, error) {
	stats, err := h.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	status := &HealthStatus{
		Healthy:      true,
		QueueSize:    stats.PendingCount,
		SendingCount: stats.SendingCount,
		FailedCount:  stats.FailedCount,
		CheckedAt:    time.Now(),
	}

	if stats.PendingCount > 1000 {
		status.Healthy = false
		status.Issues = append(status.Issues, "Queue backlog too large")
	}

	if stats.FailedCount > 100 {
		status.Healthy = false
		status.Issues = append(status.Issues, "Too many failed emails")
	}

	return status, nil
}

type HealthStatus struct {
	Healthy      bool      `json:"healthy"`
	QueueSize    int       `json:"queue_size"`
	SendingCount int       `json:"sending_count"`
	FailedCount  int       `json:"failed_count"`
	Issues       []string  `json:"issues,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`
}
