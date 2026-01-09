package email

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type MockEmailQueueRepository struct {
	GetPendingFunc      func(ctx context.Context, maxCount int) ([]*models.EmailQueue, error)
	MarkSendingFunc     func(ctx context.Context, id int64) error
	MarkSentFunc        func(ctx context.Context, id int64) error
	MarkFailedFunc      func(ctx context.Context, id int64, errorMsg string) error
	IncrementAttemptsFunc func(ctx context.Context, id int64, errorMsg string) error
	RescheduleFunc      func(ctx context.Context, id int64, scheduledFor time.Time) error
}

func (m *MockEmailQueueRepository) Create(ctx context.Context, email *models.EmailQueue) error {
	return nil
}

func (m *MockEmailQueueRepository) GetByID(ctx context.Context, id int64) (*models.EmailQueue, error) {
	return nil, nil
}

func (m *MockEmailQueueRepository) GetPending(ctx context.Context, maxCount int) ([]*models.EmailQueue, error) {
	if m.GetPendingFunc != nil {
		return m.GetPendingFunc(ctx, maxCount)
	}
	return []*models.EmailQueue{}, nil
}

func (m *MockEmailQueueRepository) GetByStatus(ctx context.Context, status models.EmailStatus, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *MockEmailQueueRepository) GetByRecipient(ctx context.Context, email string, limit int) ([]*models.EmailQueue, error) {
	return nil, nil
}

func (m *MockEmailQueueRepository) UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error {
	return nil
}

func (m *MockEmailQueueRepository) IncrementAttempts(ctx context.Context, id int64, errorMsg string) error {
	if m.IncrementAttemptsFunc != nil {
		return m.IncrementAttemptsFunc(ctx, id, errorMsg)
	}
	return nil
}

func (m *MockEmailQueueRepository) MarkSending(ctx context.Context, id int64) error {
	if m.MarkSendingFunc != nil {
		return m.MarkSendingFunc(ctx, id)
	}
	return nil
}

func (m *MockEmailQueueRepository) MarkSent(ctx context.Context, id int64) error {
	if m.MarkSentFunc != nil {
		return m.MarkSentFunc(ctx, id)
	}
	return nil
}

func (m *MockEmailQueueRepository) MarkFailed(ctx context.Context, id int64, errorMsg string) error {
	if m.MarkFailedFunc != nil {
		return m.MarkFailedFunc(ctx, id, errorMsg)
	}
	return nil
}

func (m *MockEmailQueueRepository) MarkCancelled(ctx context.Context, id int64) error {
	return nil
}

func (m *MockEmailQueueRepository) Reschedule(ctx context.Context, id int64, scheduledFor time.Time) error {
	if m.RescheduleFunc != nil {
		return m.RescheduleFunc(ctx, id, scheduledFor)
	}
	return nil
}

func (m *MockEmailQueueRepository) GetStats(ctx context.Context) (*repositories.EmailQueueStats, error) {
	return nil, nil
}

type MockSMTPSender struct {
	SendFunc           func(ctx context.Context, msg *SMTPMessage) error
	TestConnectionFunc func(ctx context.Context) error
	CloseFunc          func() error
}

func (m *MockSMTPSender) Send(ctx context.Context, msg *SMTPMessage) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, msg)
	}
	return nil
}

func (m *MockSMTPSender) TestConnection(ctx context.Context) error {
	if m.TestConnectionFunc != nil {
		return m.TestConnectionFunc(ctx)
	}
	return nil
}

func (m *MockSMTPSender) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

type MockRateLimiter struct {
	AllowFunc          func() bool
	AvailableSlotsFunc func() int
	WaitTimeFunc       func() time.Duration
	RecordFunc         func()
	ResetFunc          func()
}

func (m *MockRateLimiter) Allow() bool {
	if m.AllowFunc != nil {
		return m.AllowFunc()
	}
	return true
}

func (m *MockRateLimiter) AvailableSlots() int {
	if m.AvailableSlotsFunc != nil {
		return m.AvailableSlotsFunc()
	}
	return 100
}

func (m *MockRateLimiter) WaitTime() time.Duration {
	if m.WaitTimeFunc != nil {
		return m.WaitTimeFunc()
	}
	return 0
}

func (m *MockRateLimiter) Record() {
	if m.RecordFunc != nil {
		m.RecordFunc()
	}
}

func (m *MockRateLimiter) Reset() {
	if m.ResetFunc != nil {
		m.ResetFunc()
	}
}

func TestNewQueueProcessor(t *testing.T) {
	repo := &MockEmailQueueRepository{}
	sender := &MockSMTPSender{}
	limiter := &MockRateLimiter{}

	processor := NewQueueProcessor(repo, sender, limiter, 10, time.Minute)

	if processor == nil {
		t.Fatal("NewQueueProcessor() returned nil")
	}
}

func TestQueueProcessor_ProcessBatch_NoEmails(t *testing.T) {
	mockRepo := &MockEmailQueueRepository{
		GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
			return []*models.EmailQueue{}, nil
		},
	}

	mockLimiter := &MockRateLimiter{
		AvailableSlotsFunc: func() int { return 10 },
	}

	processor := NewQueueProcessor(mockRepo, nil, mockLimiter, 10, time.Minute)

	err := processor.ProcessBatch(context.Background())
	if err != nil {
		t.Errorf("ProcessBatch() error = %v, want nil", err)
	}
}

func TestQueueProcessor_ProcessBatch_Success(t *testing.T) {
	toName := "Test User"
	email := &models.EmailQueue{
		ID:          1,
		ToEmail:     "test@example.com",
		ToName:      &toName,
		Subject:     "Test",
		BodyText:    "Test body",
		Status:      models.EmailStatusPending,
		Attempts:    0,
		MaxAttempts: 4,
	}

	markSendingCalled := false
	markSentCalled := false
	sendCalled := false

	mockRepo := &MockEmailQueueRepository{
		GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
			return []*models.EmailQueue{email}, nil
		},
		MarkSendingFunc: func(ctx context.Context, id int64) error {
			markSendingCalled = true
			if id != 1 {
				t.Errorf("MarkSending() id = %d, want 1", id)
			}
			return nil
		},
		MarkSentFunc: func(ctx context.Context, id int64) error {
			markSentCalled = true
			if id != 1 {
				t.Errorf("MarkSent() id = %d, want 1", id)
			}
			return nil
		},
	}

	mockSender := &MockSMTPSender{
		SendFunc: func(ctx context.Context, msg *SMTPMessage) error {
			sendCalled = true
			if msg.To != "test@example.com" {
				t.Errorf("Send() To = %s, want test@example.com", msg.To)
			}
			return nil
		},
	}

	mockLimiter := &MockRateLimiter{
		AllowFunc:          func() bool { return true },
		AvailableSlotsFunc: func() int { return 10 },
	}

	processor := NewQueueProcessor(mockRepo, mockSender, mockLimiter, 10, time.Minute)

	err := processor.ProcessBatch(context.Background())
	if err != nil {
		t.Errorf("ProcessBatch() error = %v, want nil", err)
	}

	if !markSendingCalled {
		t.Error("MarkSending() was not called")
	}
	if !markSentCalled {
		t.Error("MarkSent() was not called")
	}
	if !sendCalled {
		t.Error("Send() was not called")
	}
}

func TestQueueProcessor_ProcessBatch_SendFailure(t *testing.T) {
	email := &models.EmailQueue{
		ID:          1,
		ToEmail:     "test@example.com",
		Subject:     "Test",
		BodyText:    "Test body",
		Status:      models.EmailStatusPending,
		Attempts:    0,
		MaxAttempts: 4,
	}

	incrementCalled := false
	rescheduleCalled := false

	mockRepo := &MockEmailQueueRepository{
		GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
			return []*models.EmailQueue{email}, nil
		},
		MarkSendingFunc: func(ctx context.Context, id int64) error {
			return nil
		},
		IncrementAttemptsFunc: func(ctx context.Context, id int64, errorMsg string) error {
			incrementCalled = true
			if errorMsg == "" {
				t.Error("IncrementAttempts() errorMsg is empty")
			}
			return nil
		},
		RescheduleFunc: func(ctx context.Context, id int64, scheduledFor time.Time) error {
			rescheduleCalled = true
			if scheduledFor.Before(time.Now()) {
				t.Error("Reschedule() scheduledFor is in the past")
			}
			return nil
		},
	}

	mockSender := &MockSMTPSender{
		SendFunc: func(ctx context.Context, msg *SMTPMessage) error {
			return errors.New("SMTP connection failed")
		},
	}

	mockLimiter := &MockRateLimiter{
		AllowFunc:          func() bool { return true },
		AvailableSlotsFunc: func() int { return 10 },
	}

	processor := NewQueueProcessor(mockRepo, mockSender, mockLimiter, 10, time.Minute)

	err := processor.ProcessBatch(context.Background())
	if err != nil {
		t.Errorf("ProcessBatch() error = %v, want nil", err)
	}

	if !incrementCalled {
		t.Error("IncrementAttempts() was not called")
	}
	if !rescheduleCalled {
		t.Error("Reschedule() was not called")
	}
}

func TestQueueProcessor_ProcessBatch_MaxAttemptsReached(t *testing.T) {
	email := &models.EmailQueue{
		ID:          1,
		ToEmail:     "test@example.com",
		Subject:     "Test",
		BodyText:    "Test body",
		Status:      models.EmailStatusPending,
		Attempts:    3,
		MaxAttempts: 4,
	}

	markFailedCalled := false

	mockRepo := &MockEmailQueueRepository{
		GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
			return []*models.EmailQueue{email}, nil
		},
		MarkSendingFunc: func(ctx context.Context, id int64) error {
			return nil
		},
		IncrementAttemptsFunc: func(ctx context.Context, id int64, errorMsg string) error {
			return nil
		},
		MarkFailedFunc: func(ctx context.Context, id int64, errorMsg string) error {
			markFailedCalled = true
			if errorMsg == "" {
				t.Error("MarkFailed() errorMsg is empty")
			}
			return nil
		},
	}

	mockSender := &MockSMTPSender{
		SendFunc: func(ctx context.Context, msg *SMTPMessage) error {
			return errors.New("SMTP connection failed")
		},
	}

	mockLimiter := &MockRateLimiter{
		AllowFunc:          func() bool { return true },
		AvailableSlotsFunc: func() int { return 10 },
	}

	processor := NewQueueProcessor(mockRepo, mockSender, mockLimiter, 10, time.Minute)

	err := processor.ProcessBatch(context.Background())
	if err != nil {
		t.Errorf("ProcessBatch() error = %v, want nil", err)
	}

	if !markFailedCalled {
		t.Error("MarkFailed() was not called")
	}
}

func TestQueueProcessor_ProcessBatch_RateLimitZeroSlots(t *testing.T) {
	getPendingCalled := false

	mockRepo := &MockEmailQueueRepository{
		GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
			getPendingCalled = true
			return []*models.EmailQueue{}, nil
		},
	}

	mockLimiter := &MockRateLimiter{
		AvailableSlotsFunc: func() int { return 0 },
	}

	processor := NewQueueProcessor(mockRepo, nil, mockLimiter, 10, time.Minute)

	err := processor.ProcessBatch(context.Background())
	if err != nil {
		t.Errorf("ProcessBatch() error = %v, want nil", err)
	}

	if getPendingCalled {
		t.Error("GetPending() should not be called when rate limit is 0")
	}
}

func TestQueueProcessor_ProcessBatch_BatchSizeLimiting(t *testing.T) {
	mockRepo := &MockEmailQueueRepository{
		GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
			if max != 5 {
				t.Errorf("GetPending() max = %d, want 5", max)
			}
			return []*models.EmailQueue{}, nil
		},
	}

	mockLimiter := &MockRateLimiter{
		AvailableSlotsFunc: func() int { return 5 },
	}

	processor := NewQueueProcessor(mockRepo, nil, mockLimiter, 10, time.Minute)

	err := processor.ProcessBatch(context.Background())
	if err != nil {
		t.Errorf("ProcessBatch() error = %v, want nil", err)
	}
}

func TestQueueProcessor_GracefulShutdown(t *testing.T) {
	mockRepo := &MockEmailQueueRepository{}
	mockLimiter := &MockRateLimiter{}

	processor := NewQueueProcessor(mockRepo, nil, mockLimiter, 10, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- processor.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	cancel()

	select {
	case err := <-errChan:
		if err != context.Canceled {
			t.Errorf("Start() error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not stop within timeout")
	}
}

func TestQueueProcessor_Stop(t *testing.T) {
	mockRepo := &MockEmailQueueRepository{}
	mockLimiter := &MockRateLimiter{}

	processor := NewQueueProcessor(mockRepo, nil, mockLimiter, 10, 100*time.Millisecond)

	go func() {
		_ = processor.Start(context.Background())
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := processor.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() error = %v, want nil", err)
	}
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name    string
		attempt int
		base    time.Duration
	}{
		{"first attempt", 1, 1 * time.Minute},
		{"second attempt", 2, 5 * time.Minute},
		{"third attempt", 3, 15 * time.Minute},
		{"fourth attempt", 4, 30 * time.Minute},
		{"fifth attempt", 5, 30 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBackoff(tt.attempt)
			
			maxJitter := time.Duration(float64(tt.base) * 0.1)
			minDelay := tt.base - maxJitter
			maxDelay := tt.base + maxJitter

			if got < minDelay || got > maxDelay {
				t.Errorf("calculateBackoff(%d) = %v, want range [%v, %v]", tt.attempt, got, minDelay, maxDelay)
			}
		})
	}
}

func TestQueueProcessor_ProcessBatch_PermanentError(t *testing.T) {
	email := &models.EmailQueue{
		ID:          1,
		ToEmail:     "test@example.com",
		Subject:     "Test",
		BodyText:    "Test body",
		Status:      models.EmailStatusPending,
		Attempts:    0,
		MaxAttempts: 4,
	}

	markFailedCalled := false
	rescheduleCalled := false

	mockRepo := &MockEmailQueueRepository{
		GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
			return []*models.EmailQueue{email}, nil
		},
		MarkSendingFunc: func(ctx context.Context, id int64) error {
			return nil
		},
		IncrementAttemptsFunc: func(ctx context.Context, id int64, errorMsg string) error {
			return nil
		},
		MarkFailedFunc: func(ctx context.Context, id int64, errorMsg string) error {
			markFailedCalled = true
			return nil
		},
		RescheduleFunc: func(ctx context.Context, id int64, scheduledFor time.Time) error {
			rescheduleCalled = true
			return nil
		},
	}

	mockSender := &MockSMTPSender{
		SendFunc: func(ctx context.Context, msg *SMTPMessage) error {
			return &PermanentError{Err: errors.New("mailbox unavailable")}
		},
	}

	mockLimiter := &MockRateLimiter{
		AllowFunc:          func() bool { return true },
		AvailableSlotsFunc: func() int { return 10 },
	}

	processor := NewQueueProcessor(mockRepo, mockSender, mockLimiter, 10, time.Minute)

	err := processor.ProcessBatch(context.Background())
	if err != nil {
		t.Errorf("ProcessBatch() error = %v, want nil", err)
	}

	if !markFailedCalled {
		t.Error("MarkFailed() was not called for permanent error")
	}
	if rescheduleCalled {
		t.Error("Reschedule() should not be called for permanent error")
	}
}

func TestQueueProcessor_ProcessBatch_TransientError(t *testing.T) {
	email := &models.EmailQueue{
		ID:          1,
		ToEmail:     "test@example.com",
		Subject:     "Test",
		BodyText:    "Test body",
		Status:      models.EmailStatusPending,
		Attempts:    0,
		MaxAttempts: 4,
	}

	markFailedCalled := false
	rescheduleCalled := false

	mockRepo := &MockEmailQueueRepository{
		GetPendingFunc: func(ctx context.Context, max int) ([]*models.EmailQueue, error) {
			return []*models.EmailQueue{email}, nil
		},
		MarkSendingFunc: func(ctx context.Context, id int64) error {
			return nil
		},
		IncrementAttemptsFunc: func(ctx context.Context, id int64, errorMsg string) error {
			return nil
		},
		MarkFailedFunc: func(ctx context.Context, id int64, errorMsg string) error {
			markFailedCalled = true
			return nil
		},
		RescheduleFunc: func(ctx context.Context, id int64, scheduledFor time.Time) error {
			rescheduleCalled = true
			return nil
		},
	}

	mockSender := &MockSMTPSender{
		SendFunc: func(ctx context.Context, msg *SMTPMessage) error {
			return &TransientError{Err: errors.New("mailbox busy")}
		},
	}

	mockLimiter := &MockRateLimiter{
		AllowFunc:          func() bool { return true },
		AvailableSlotsFunc: func() int { return 10 },
	}

	processor := NewQueueProcessor(mockRepo, mockSender, mockLimiter, 10, time.Minute)

	err := processor.ProcessBatch(context.Background())
	if err != nil {
		t.Errorf("ProcessBatch() error = %v, want nil", err)
	}

	if markFailedCalled {
		t.Error("MarkFailed() should not be called for transient error")
	}
	if !rescheduleCalled {
		t.Error("Reschedule() was not called for transient error")
	}
}

func TestCalculateBackoff_Jitter(t *testing.T) {
	delays := make(map[time.Duration]bool)

	for i := 0; i < 100; i++ {
		delay := calculateBackoff(1)
		delays[delay] = true
	}

	if len(delays) < 10 {
		t.Errorf("Jitter not working, only %d unique delays in 100 attempts", len(delays))
	}

	for delay := range delays {
		expectedBase := 1 * time.Minute
		maxJitter := time.Duration(float64(expectedBase) * 0.1)
		minDelay := expectedBase - maxJitter
		maxDelay := expectedBase + maxJitter

		if delay < minDelay || delay > maxDelay {
			t.Errorf("Delay %v outside expected range [%v, %v]", delay, minDelay, maxDelay)
		}
	}
}
