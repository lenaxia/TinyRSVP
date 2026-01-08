package email

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type SMTPMessage struct {
	To          string
	ToName      string
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
}

type RateLimiter interface {
	Allow() bool
	AvailableSlots() int
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
}

func NewQueueProcessor(
	repo repositories.EmailQueueRepository,
	sender SMTPSender,
	rateLimiter RateLimiter,
	batchSize int,
	pollInterval time.Duration,
) QueueProcessor {
	return &queueProcessor{
		repo:         repo,
		sender:       sender,
		rateLimiter:  rateLimiter,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		stopChan:     make(chan struct{}),
		doneChan:     make(chan struct{}),
	}
}

func (p *queueProcessor) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	log.Printf("Email queue processor started (interval: %v, batch: %d)",
		p.pollInterval, p.batchSize)

	for {
		select {
		case <-ctx.Done():
			log.Println("Email queue processor stopped (context cancelled)")
			close(p.doneChan)
			return ctx.Err()

		case <-p.stopChan:
			log.Println("Email queue processor stopped (shutdown requested)")
			close(p.doneChan)
			return nil

		case <-ticker.C:
			if err := p.ProcessBatch(ctx); err != nil {
				log.Printf("Error processing email batch: %v", err)
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
	availableSlots := p.rateLimiter.AvailableSlots()
	if availableSlots == 0 {
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

	log.Printf("Processing %d emails", len(emails))

	for _, email := range emails {
		if err := p.processEmail(ctx, email); err != nil {
			log.Printf("Failed to process email %d: %v", email.ID, err)
		}
	}

	return nil
}

func (p *queueProcessor) processEmail(ctx context.Context, email *models.EmailQueue) error {
	if err := p.repo.MarkSending(ctx, email.ID); err != nil {
		return fmt.Errorf("failed to mark as sending: %w", err)
	}

	if !p.rateLimiter.Allow() {
		if err := p.repo.UpdateStatus(ctx, email.ID, models.EmailStatusPending); err != nil {
			return fmt.Errorf("failed to reset status: %w", err)
		}
		return p.repo.Reschedule(ctx, email.ID, time.Now().Add(time.Minute))
	}

	if err := p.sendEmail(ctx, email); err != nil {
		return p.handleSendError(ctx, email, err)
	}

	if err := p.repo.MarkSent(ctx, email.ID); err != nil {
		log.Printf("Warning: email sent but failed to mark as sent: %v", err)
	}

	log.Printf("Email %d sent successfully to %s", email.ID, email.ToEmail)
	return nil
}

func (p *queueProcessor) sendEmail(ctx context.Context, email *models.EmailQueue) error {
	attachments, err := email.GetAttachments()
	if err != nil {
		return fmt.Errorf("failed to get attachments: %w", err)
	}

	toName := ""
	if email.ToName != nil {
		toName = *email.ToName
	}

	bodyHTML := ""
	if email.BodyHTML != nil {
		bodyHTML = *email.BodyHTML
	}

	msg := &SMTPMessage{
		To:          email.ToEmail,
		ToName:      toName,
		Subject:     email.Subject,
		BodyText:    email.BodyText,
		BodyHTML:    bodyHTML,
		Attachments: convertAttachments(attachments),
	}

	return p.sender.Send(ctx, msg)
}

func (p *queueProcessor) handleSendError(ctx context.Context, email *models.EmailQueue, err error) error {
	if err := p.repo.IncrementAttempts(ctx, email.ID, err.Error()); err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	if email.Attempts+1 >= email.MaxAttempts {
		if err := p.repo.MarkFailed(ctx, email.ID, err.Error()); err != nil {
			log.Printf("Failed to mark email as failed: %v", err)
		}
		log.Printf("Email %d permanently failed after %d attempts: %v",
			email.ID, email.Attempts+1, err)
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

	log.Printf("Email %d rescheduled for retry in %v (attempt %d/%d)",
		email.ID, backoff, email.Attempts+1, email.MaxAttempts)

	return nil
}

func calculateBackoff(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 1 * time.Minute
	case 2:
		return 5 * time.Minute
	case 3:
		return 15 * time.Minute
	default:
		return 30 * time.Minute
	}
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
