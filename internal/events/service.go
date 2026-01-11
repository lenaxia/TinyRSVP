package events

import (
	"context"
	"fmt"

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
	ListEvents(ctx context.Context, filters ListFilters) ([]*models.Event, error)
	PublishEvent(ctx context.Context, id int64) error
	CancelEvent(ctx context.Context, id int64, reason string) error
	ArchiveEvent(ctx context.Context, id int64) error
	GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error)
}

type ListFilters struct {
	CreatorID *int64
	Status    *models.EventStatus
	Limit     int
	Offset    int
}

type service struct {
	repo      repositories.EventRepository
	validator Validator
	authz     auth.AuthorizationChecker
}

func NewService(
	repo repositories.EventRepository,
	validator Validator,
	authz auth.AuthorizationChecker,
) Service {
	return &service{
		repo:      repo,
		validator: validator,
		authz:     authz,
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

	if err := s.repo.UpdateWithVersion(ctx, event, event.Version); err != nil {
		return err
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

	repoFilters := repositories.ListFilters{
		CreatorID: filters.CreatorID,
		Status:    filters.Status,
		Limit:     filters.Limit,
		Offset:    filters.Offset,
	}

	if !s.authz.IsAdmin(user) {
		repoFilters.CreatorID = &user.ID
	}

	events, err := s.repo.List(ctx, repoFilters)
	if err != nil {
		return nil, err
	}

	return events, nil
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
