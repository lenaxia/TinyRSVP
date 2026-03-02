package email

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type SMTPMessage struct {
	To          string
	ToName      *string
	Subject     string
	BodyText    string
	BodyHTML    string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type SMTPSender interface {
	Send(ctx context.Context, msg *SMTPMessage) error
	TestConnection(ctx context.Context) error
	Close() error
}

type RateLimiter interface {
	Allow() bool
	AvailableSlots() int
	WaitTime() time.Duration
	Record()
	Reset()
}

type QueueProcessor interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	ProcessBatch(ctx context.Context) error
}

type queueProcessor struct {
	repo         repositories.EmailQueueRepository
	sender       SMTPSender
	rateLimiter  RateLimiter
	batchSize    int
	pollInterval time.Duration
	stopChan     chan struct{}
	doneChan     chan struct{}
	metrics      Metrics
	logger       *Logger
}

func NewQueueProcessor(
	repo repositories.EmailQueueRepository,
	sender SMTPSender,
	rateLimiter RateLimiter,
	batchSize int,
	pollInterval time.Duration,
	metrics Metrics,
	logger *Logger,
) QueueProcessor {
	return &queueProcessor{
		repo:         repo,
		sender:       sender,
		rateLimiter:  rateLimiter,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		stopChan:     make(chan struct{}),
		doneChan:     make(chan struct{}),
		metrics:      metrics,
		logger:       logger,
	}
}

func (p *queueProcessor) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	p.logger.QueueProcessorStarted(p.pollInterval, p.batchSize)

	for {
		select {
		case <-ctx.Done():
			p.logger.QueueProcessorStopped()
			close(p.doneChan)
			return ctx.Err()

		case <-p.stopChan:
			p.logger.QueueProcessorStopped()
			close(p.doneChan)
			return nil

		case <-ticker.C:
			if err := p.ProcessBatch(ctx); err != nil {
				p.metrics.RecordProcessingError(err)
			}
		}
	}
}

func (p *queueProcessor) Stop(ctx context.Context) error {
	close(p.stopChan)

	select {
	case <-p.doneChan:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *queueProcessor) ProcessBatch(ctx context.Context) error {
	startTime := time.Now()

	availableSlots := p.rateLimiter.AvailableSlots()
	if availableSlots == 0 {
		p.metrics.RecordRateLimitHit()
		return nil
	}

	batchSize := min(p.batchSize, availableSlots)
	emails, err := p.repo.GetPending(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending emails: %w", err)
	}

	if len(emails) == 0 {
		return nil
	}

	p.metrics.RecordQueueSize(len(emails))

	for _, email := range emails {
		if err := p.processEmail(ctx, email); err != nil {
			p.metrics.RecordProcessingError(err)
		}
	}

	duration := time.Since(startTime)
	p.metrics.RecordBatchProcessed(len(emails), duration)
	p.logger.BatchProcessed(len(emails), duration)

	return nil
}

func (p *queueProcessor) processEmail(ctx context.Context, email *models.EmailQueue) error {
	startTime := time.Now()

	if err := p.repo.MarkSending(ctx, email.ID); err != nil {
		return fmt.Errorf("failed to mark as sending: %w", err)
	}

	p.logger.EmailSending(email.ID, email.ToEmail, email.Attempts+1)

	if !p.rateLimiter.Allow() {
		waitTime := p.rateLimiter.WaitTime()
		p.metrics.RecordRateLimitHit()
		p.metrics.RecordRateLimitWait(waitTime)
		p.logger.RateLimitHit(p.rateLimiter.AvailableSlots(), waitTime)

		if err := p.repo.UpdateStatus(ctx, email.ID, models.EmailStatusPending); err != nil {
			return fmt.Errorf("failed to reset status: %w", err)
		}
		return p.repo.Reschedule(ctx, email.ID, time.Now().Add(time.Minute))
	}

	if err := p.sendEmail(ctx, email); err != nil {
		return p.handleSendError(ctx, email, err)
	}

	p.rateLimiter.Record()

	duration := time.Since(startTime)
	p.metrics.RecordEmailSent(duration)
	p.metrics.RecordEmailDequeued()

	if err := p.repo.MarkSent(ctx, email.ID); err != nil {
		p.metrics.RecordProcessingError(err)
	}

	p.logger.EmailSent(email.ID, email.ToEmail, duration)
	return nil
}

func (p *queueProcessor) sendEmail(ctx context.Context, email *models.EmailQueue) error {
	attachments, err := email.GetAttachments()
	if err != nil {
		return fmt.Errorf("failed to get attachments: %w", err)
	}

	bodyHTML := ""
	if email.BodyHTML != nil {
		bodyHTML = *email.BodyHTML
	}

	msg := &SMTPMessage{
		To:          email.ToEmail,
		ToName:      email.ToName,
		Subject:     email.Subject,
		BodyText:    email.BodyText,
		BodyHTML:    bodyHTML,
		Attachments: convertAttachments(attachments),
	}

	return p.sender.Send(ctx, msg)
}

func (p *queueProcessor) handleSendError(ctx context.Context, email *models.EmailQueue, err error) error {
	p.metrics.RecordEmailFailed(err.Error())
	p.logger.EmailFailed(email.ID, email.ToEmail, email.Attempts+1, err)

	var permErr *PermanentError
	if errors.As(err, &permErr) {
		p.logger.EmailPermanentlyFailed(email.ID, email.ToEmail, email.Attempts+1, err)
		return p.repo.MarkFailed(ctx, email.ID, err.Error())
	}

	if err := p.repo.IncrementAttempts(ctx, email.ID, err.Error()); err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	p.metrics.RecordRetryAttempt(email.Attempts + 1)

	if email.Attempts+1 >= email.MaxAttempts {
		p.logger.EmailPermanentlyFailed(email.ID, email.ToEmail, email.Attempts+1, err)
		if err := p.repo.MarkFailed(ctx, email.ID, err.Error()); err != nil {
			p.metrics.RecordProcessingError(err)
		}
		return nil
	}

	backoff := calculateBackoff(email.Attempts + 1)
	scheduledFor := time.Now().Add(backoff)

	if err := p.repo.UpdateStatus(ctx, email.ID, models.EmailStatusPending); err != nil {
		return fmt.Errorf("failed to reset status to pending: %w", err)
	}

	if err := p.repo.Reschedule(ctx, email.ID, scheduledFor); err != nil {
		return fmt.Errorf("failed to reschedule: %w", err)
	}

	p.logger.EmailRetrying(email.ID, email.ToEmail, email.Attempts+1, backoff)

	return nil
}

func calculateBackoff(attempt int) time.Duration {
	var delay time.Duration

	switch attempt {
	case 1:
		delay = 1 * time.Minute
	case 2:
		delay = 5 * time.Minute
	case 3:
		delay = 15 * time.Minute
	default:
		delay = 30 * time.Minute
	}

	jitter := time.Duration(float64(delay) * 0.1 * (rand.Float64()*2 - 1))
	return delay + jitter
}

func convertAttachments(attachments []models.EmailAttachment) []Attachment {
	result := make([]Attachment, len(attachments))
	for i, att := range attachments {
		result[i] = Attachment{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			Data:        att.Content,
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
