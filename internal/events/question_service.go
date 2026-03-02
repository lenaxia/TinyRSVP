package events

import (
	"context"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type QuestionService interface {
	AddQuestion(ctx context.Context, question *models.PreferenceQuestion) error
	UpdateQuestion(ctx context.Context, question *models.PreferenceQuestion) error
	DeleteQuestion(ctx context.Context, id int64) error
	GetQuestions(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
	ReorderQuestions(ctx context.Context, eventID int64, questionIDs []int64) error
}

type QuestionValidator interface {
	ValidateCreate(ctx context.Context, question *models.PreferenceQuestion) error
	ValidateUpdate(ctx context.Context, question *models.PreferenceQuestion) error
}

type questionService struct {
	eventRepo    repositories.EventRepository
	questionRepo repositories.QuestionRepository
	validator    QuestionValidator
	authz        auth.AuthorizationChecker
}

func NewQuestionService(
	eventRepo repositories.EventRepository,
	questionRepo repositories.QuestionRepository,
	validator QuestionValidator,
	authz auth.AuthorizationChecker,
) QuestionService {
	return &questionService{
		eventRepo:    eventRepo,
		questionRepo: questionRepo,
		validator:    validator,
		authz:        authz,
	}
}

func (s *questionService) AddQuestion(ctx context.Context, question *models.PreferenceQuestion) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "add question",
			Resource: "PreferenceQuestion",
		}
	}

	event, err := s.eventRepo.GetByID(ctx, question.EventID)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "add question",
			Resource: "PreferenceQuestion",
		}
	}

	if event.Status != models.EventStatusDraft {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot modify questions on published event",
		}
	}

	if err := s.validator.ValidateCreate(ctx, question); err != nil {
		return err
	}

	return s.questionRepo.Create(ctx, question)
}

func (s *questionService) UpdateQuestion(ctx context.Context, question *models.PreferenceQuestion) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "update question",
			Resource: "PreferenceQuestion",
		}
	}

	existing, err := s.questionRepo.GetByID(ctx, question.ID)
	if err != nil {
		return err
	}

	event, err := s.eventRepo.GetByID(ctx, existing.EventID)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "update question",
			Resource: "PreferenceQuestion",
			ID:       question.ID,
		}
	}

	if event.Status != models.EventStatusDraft {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot modify questions on published event",
		}
	}

	question.EventID = existing.EventID

	if err := s.validator.ValidateUpdate(ctx, question); err != nil {
		return err
	}

	return s.questionRepo.Update(ctx, question)
}

func (s *questionService) DeleteQuestion(ctx context.Context, id int64) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "delete question",
			Resource: "PreferenceQuestion",
		}
	}

	question, err := s.questionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	event, err := s.eventRepo.GetByID(ctx, question.EventID)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "delete question",
			Resource: "PreferenceQuestion",
			ID:       id,
		}
	}

	if event.Status != models.EventStatusDraft {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot modify questions on published event",
		}
	}

	return s.questionRepo.Delete(ctx, id)
}

func (s *questionService) GetQuestions(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, &models.PermissionDeniedError{
			Action:   "view questions",
			Resource: "PreferenceQuestion",
		}
	}

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if !s.authz.CanViewEvent(ctx, user, event) {
		return nil, &models.PermissionDeniedError{
			Action:   "view questions",
			Resource: "PreferenceQuestion",
		}
	}

	return s.questionRepo.GetByEventID(ctx, eventID)
}

func (s *questionService) ReorderQuestions(ctx context.Context, eventID int64, questionIDs []int64) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return &models.PermissionDeniedError{
			Action:   "reorder questions",
			Resource: "PreferenceQuestion",
		}
	}

	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	if !s.authz.CanEditEvent(ctx, user, event) {
		return &models.PermissionDeniedError{
			Action:   "reorder questions",
			Resource: "PreferenceQuestion",
		}
	}

	if event.Status != models.EventStatusDraft {
		return &models.ValidationError{
			Field:   "status",
			Message: "cannot modify questions on published event",
		}
	}

	if len(questionIDs) == 0 {
		return &models.ValidationError{
			Field:   "question_ids",
			Message: "at least one question ID required",
		}
	}

	return s.questionRepo.Reorder(ctx, eventID, questionIDs)
}
