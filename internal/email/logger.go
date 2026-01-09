package email

import (
	"log/slog"
	"time"
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger(logger *slog.Logger) *Logger {
	return &Logger{logger: logger}
}

func (l *Logger) EmailQueued(emailID int64, recipient string) {
	l.logger.Info("email queued",
		slog.Int64("email_id", emailID),
		slog.String("recipient", recipient),
	)
}

func (l *Logger) EmailSending(emailID int64, recipient string, attempt int) {
	l.logger.Info("sending email",
		slog.Int64("email_id", emailID),
		slog.String("recipient", recipient),
		slog.Int("attempt", attempt),
	)
}

func (l *Logger) EmailSent(emailID int64, recipient string, duration time.Duration) {
	l.logger.Info("email sent successfully",
		slog.Int64("email_id", emailID),
		slog.String("recipient", recipient),
		slog.Duration("duration", duration),
	)
}

func (l *Logger) EmailFailed(emailID int64, recipient string, attempt int, err error) {
	l.logger.Error("email send failed",
		slog.Int64("email_id", emailID),
		slog.String("recipient", recipient),
		slog.Int("attempt", attempt),
		slog.String("error", err.Error()),
	)
}

func (l *Logger) EmailRetrying(emailID int64, recipient string, attempt int, backoff time.Duration) {
	l.logger.Warn("retrying email",
		slog.Int64("email_id", emailID),
		slog.String("recipient", recipient),
		slog.Int("attempt", attempt),
		slog.Duration("backoff", backoff),
	)
}

func (l *Logger) EmailPermanentlyFailed(emailID int64, recipient string, attempts int, err error) {
	l.logger.Error("email permanently failed",
		slog.Int64("email_id", emailID),
		slog.String("recipient", recipient),
		slog.Int("total_attempts", attempts),
		slog.String("error", err.Error()),
	)
}

func (l *Logger) RateLimitHit(available int, waitTime time.Duration) {
	l.logger.Warn("rate limit hit",
		slog.Int("available_slots", available),
		slog.Duration("wait_time", waitTime),
	)
}

func (l *Logger) BatchProcessed(count int, duration time.Duration) {
	l.logger.Info("batch processed",
		slog.Int("email_count", count),
		slog.Duration("duration", duration),
	)
}

func (l *Logger) QueueProcessorStarted(interval time.Duration, batchSize int) {
	l.logger.Info("queue processor started",
		slog.Duration("poll_interval", interval),
		slog.Int("batch_size", batchSize),
	)
}

func (l *Logger) QueueProcessorStopped() {
	l.logger.Info("queue processor stopped")
}
