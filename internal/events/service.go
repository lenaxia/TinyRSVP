package events

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/pkg/eventid"
)

type Service interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	GetEvent(ctx context.Context, id int64) (*models.Event, error)
	UpdateEvent(ctx context.Context, event *models.Event) error
	DeleteEvent(ctx context.Context, id int64) error
	ListEvents(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error)
	ListEventsWithStats(ctx context.Context, filters repositories.ListFilters) ([]*models.EventWithStats, error)
	CountEvents(ctx context.Context, filters repositories.ListFilters) (int, error)
	PublishEvent(ctx context.Context, id int64) error
	CancelEvent(ctx context.Context, id int64, reason string) error
	ArchiveEvent(ctx context.Context, id int64) error
	GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error)
}

// ListFilters is an alias for repositories.ListFilters so the service
// interface and callers can use it without importing the repositories
// package directly. Eliminates the prior duplicate struct that required
// manual field-by-field copying.
type ListFilters = repositories.ListFilters

type service struct {
	repo       repositories.EventRepository
	inviteRepo repositories.InviteRepository
	validator  Validator
	authz      auth.AuthorizationChecker
}

func NewService(
	repo repositories.EventRepository,
	inviteRepo repositories.InviteRepository,
	validator Validator,
	authz auth.AuthorizationChecker,
) Service {
	return &service{
		repo:       repo,
		inviteRepo: inviteRepo,
		validator:  validator,
		authz:      authz,
	}
}

func (s *service) CreateEvent(ctx context.Context, event *models.Event) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "create event",
			Resource: "Event",
		}
	}

	if !s.authz.CanCreateEvent(ctx, user) {
		return &models.PermissionDeniedError{
			Action:   "create event",
			Resource: "Event",
		}
	}

	if err := s.validator.ValidateCreate(ctx, event); err != nil {
		return err
	}

	publicID, err := eventid.GenerateEventID()
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	event.PublicID = &publicID
	event.CreatedBy = user.ID
	event.Status = models.EventStatusDraft
	event.Version = 1

	if err := s.repo.Create(ctx, event); err != nil {
		return err
	}

	return nil
}

func (s *service) GetEvent(ctx context.Context, id int64) (*models.Event, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "view event",
			Resource: "Event",
		}
	}

	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !s.authz.CanViewEvent(ctx, user, event) {
		return nil, &models.PermissionDeniedError{
			Action:   "view event",
			Resource: "Event",
			ID:       id,
		}
	}

	return event, nil
}

func (s *service) UpdateEvent(ctx context.Context, event *models.Event) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "update event",
			Resource: "Event",
		}
	}

	existing, err := s.repo.GetByID(ctx, event.ID)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, existing) {
		return &models.PermissionDeniedError{
			Action:   "update event",
			Resource: "Event",
			ID:       event.ID,
		}
	}

	event.Status = existing.Status

	if err := s.validator.ValidateUpdate(ctx, event); err != nil {
		return err
	}

	startTimeChanged := !existing.StartTime.Equal(event.StartTime)

	if err := s.repo.UpdateWithVersion(ctx, event, event.Version); err != nil {
		return err
	}

	// Cascade start_time changes to invite expiry dates.
	// Invites expire at event.StartTime + 30 days; when the event is rescheduled
	// the old expiry becomes stale and links appear expired prematurely.
	if startTimeChanged && s.inviteRepo != nil {
		newExpiry := event.StartTime.Add(30 * 24 * 60 * 60 * 1e9) // 30 days
		if err := s.inviteRepo.UpdateExpiresAtByEventID(ctx, event.ID, newExpiry); err != nil {
			// Non-fatal: log but don't fail the event update
			slog.Warn("failed to update invite expiries for event", "event_id", event.ID, "error", err)
		}
	}

	return nil
}

func (s *service) DeleteEvent(ctx context.Context, id int64) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "delete event",
			Resource: "Event",
		}
	}

	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !s.authz.CanDeleteEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "delete event",
			Resource: "Event",
			ID:       id,
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *service) ListEvents(ctx context.Context, filters ListFilters) ([]*models.Event, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "list events",
			Resource: "Event",
		}
	}

	if !s.authz.IsEventManager(user) {
		return nil, &models.PermissionDeniedError{
			Action:   "list events",
			Resource: "Event",
		}
	}

	if !s.authz.IsAdmin(user) {
		filters.CreatorID = &user.ID
	}

	events, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *service) ListEventsWithStats(ctx context.Context, filters repositories.ListFilters) ([]*models.EventWithStats, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "list events",
			Resource: "Event",
		}
	}

	if !s.authz.IsEventManager(user) {
		return nil, &models.PermissionDeniedError{
			Action:   "list events",
			Resource: "Event",
		}
	}

	if !s.authz.IsAdmin(user) {
		filters.CreatorID = &user.ID
	}

	return s.repo.ListWithStats(ctx, filters)
}

// CountEvents returns the total count of events matching filters, honoring the
// same authz scoping (non-admins see only their own). Ignores Limit/Offset so
// the count reflects the full result set for pagination.
func (s *service) CountEvents(ctx context.Context, filters repositories.ListFilters) (int, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return 0, &models.PermissionDeniedError{
			Action:   "count events",
			Resource: "Event",
		}
	}

	if !s.authz.IsEventManager(user) {
		return 0, &models.PermissionDeniedError{
			Action:   "count events",
			Resource: "Event",
		}
	}

	if !s.authz.IsAdmin(user) {
		filters.CreatorID = &user.ID
	}

	return s.repo.CountByFilters(ctx, filters)
}

func (s *service) PublishEvent(ctx context.Context, id int64) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "publish event",
			Resource: "Event",
		}
	}

	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "publish event",
			Resource: "Event",
			ID:       id,
		}
	}

	if err := s.validator.ValidateStateTransition(event.Status, models.EventStatusPublished); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, models.EventStatusPublished); err != nil {
		return err
	}

	return nil
}

func (s *service) CancelEvent(ctx context.Context, id int64, reason string) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "cancel event",
			Resource: "Event",
		}
	}

	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "cancel event",
			Resource: "Event",
			ID:       id,
		}
	}

	if err := s.validator.ValidateStateTransition(event.Status, models.EventStatusCancelled); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, models.EventStatusCancelled); err != nil {
		return err
	}

	return nil
}

func (s *service) ArchiveEvent(ctx context.Context, id int64) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "archive event",
			Resource: "Event",
		}
	}

	if !s.authz.IsAdmin(user) {
		return &models.PermissionDeniedError{
			Action:   "archive event",
			Resource: "Event",
			ID:       id,
		}
	}

	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.validator.ValidateStateTransition(event.Status, models.EventStatusArchived); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, models.EventStatusArchived); err != nil {
		return err
	}

	return nil
}

func (s *service) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "get events to archive",
			Resource: "Event",
		}
	}

	if !s.authz.IsAdmin(user) {
		return nil, &models.PermissionDeniedError{
			Action:   "get events to archive",
			Resource: "Event",
		}
	}

	events, err := s.repo.GetEventsToArchive(ctx, daysAfterEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to get events to archive: %w", err)
	}

	return events, nil
}
