package rsvp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

var (
	ErrDeadlinePassed = errors.New("RSVP deadline has passed")
	ErrDuplicateRSVP  = errors.New("you have already responded to this invite")
)

type InviteService interface {
	GetInviteByToken(ctx context.Context, token string) (*models.Invite, error)
	UpdateStatus(ctx context.Context, inviteID int64, status models.InviteStatus) error
}

type Database interface {
	ExecInTransaction(ctx context.Context, fn func(context.Context) error) error
}

type Service interface {
	SubmitRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) (*models.RSVP, error)
}

type service struct {
	inviteService InviteService
	eventRepo     repositories.EventRepository
	rsvpRepo      repositories.RSVPRepository
	answerRepo    repositories.AnswerRepository
	questionRepo  repositories.QuestionRepository
	db            Database
}

func NewService(
	inviteService InviteService,
	eventRepo repositories.EventRepository,
	rsvpRepo repositories.RSVPRepository,
	answerRepo repositories.AnswerRepository,
	questionRepo repositories.QuestionRepository,
	db Database,
) Service {
	return &service{
		inviteService: inviteService,
		eventRepo:     eventRepo,
		rsvpRepo:      rsvpRepo,
		answerRepo:    answerRepo,
		questionRepo:  questionRepo,
		db:            db,
	}
}

type SubmitRSVPRequest struct {
	Response string          `json:"response"`
	PlusOnes int             `json:"plus_ones"`
	Answers  []AnswerRequest `json:"answers"`
}

type AnswerRequest struct {
	QuestionID    int64   `json:"question_id"`
	AnswerText    *string `json:"answer_text,omitempty"`
	AnswerOption  *string `json:"answer_option,omitempty"`
	AnswerBoolean *bool   `json:"answer_boolean,omitempty"`
}

func (s *service) SubmitRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) (*models.RSVP, error) {
	invite, err := s.inviteService.GetInviteByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if !invite.ExpiresAt.IsZero() && invite.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("invite has expired")
	}

	if invite.Status == models.InviteStatusRevoked {
		return nil, errors.New("invite has been revoked")
	}

	event, err := s.eventRepo.GetByID(ctx, invite.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if event.Status == models.EventStatusCancelled {
		return nil, errors.New("event has been cancelled")
	}

	if event.RSVPDeadline != nil && event.RSVPDeadline.Before(time.Now()) {
		return nil, ErrDeadlinePassed
	}

	if err := s.validateRequest(ctx, req, invite, event); err != nil {
		return nil, err
	}

	existing, err := s.rsvpRepo.GetByInviteID(ctx, invite.ID)
	if err != nil {
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("failed to check existing RSVP: %w", err)
		}
	}

	if existing != nil {
		return nil, ErrDuplicateRSVP
	}

	if req.Response == "no" && req.PlusOnes > 0 {
		req.PlusOnes = 0
	}

	rsvp := &models.RSVP{
		InviteID: invite.ID,
		Response: models.RSVPResponse(req.Response),
		PlusOnes: req.PlusOnes,
	}

	if s.db != nil {
		err = s.db.ExecInTransaction(ctx, func(txCtx context.Context) error {
			return s.createRSVPWithAnswers(txCtx, rsvp, req.Answers, invite.ID)
		})
	} else {
		err = s.createRSVPWithAnswers(ctx, rsvp, req.Answers, invite.ID)
	}

	if err != nil {
		return nil, err
	}

	return rsvp, nil
}

func (s *service) createRSVPWithAnswers(ctx context.Context, rsvp *models.RSVP, answers []AnswerRequest, inviteID int64) error {
	if err := s.rsvpRepo.Create(ctx, rsvp); err != nil {
		return fmt.Errorf("failed to create RSVP: %w", err)
	}

	for _, ansReq := range answers {
		answer := &models.RSVPAnswer{
			RSVPID:        rsvp.ID,
			QuestionID:    ansReq.QuestionID,
			AnswerText:    ansReq.AnswerText,
			AnswerOption:  ansReq.AnswerOption,
			AnswerBoolean: ansReq.AnswerBoolean,
		}

		if err := s.answerRepo.Create(ctx, answer); err != nil {
			return fmt.Errorf("failed to create answer: %w", err)
		}
	}

	if err := s.inviteService.UpdateStatus(ctx, inviteID, models.InviteStatusResponded); err != nil {
		return fmt.Errorf("failed to update invite status: %w", err)
	}

	return nil
}

func (s *service) validateRequest(ctx context.Context, req *SubmitRSVPRequest, invite *models.Invite, event *models.Event) error {
	response := models.RSVPResponse(req.Response)
	if !response.Valid() {
		return &models.ValidationError{
			Field:   "response",
			Message: "response must be yes, no, or maybe",
		}
	}

	if req.PlusOnes < 0 {
		return &models.ValidationError{
			Field:   "plus_ones",
			Message: "plus_ones cannot be negative",
		}
	}

	if req.PlusOnes > invite.MaxPlusOnes {
		return &models.ValidationError{
			Field:   "plus_ones",
			Message: fmt.Sprintf("you can bring up to %d guest(s)", invite.MaxPlusOnes),
		}
	}

	questions, err := s.questionRepo.GetByEventID(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("failed to get questions: %w", err)
	}

	answerMap := make(map[int64]AnswerRequest)
	for _, ans := range req.Answers {
		answerMap[ans.QuestionID] = ans
	}

	for _, q := range questions {
		if q.Required {
			if _, ok := answerMap[q.ID]; !ok {
				return &models.ValidationError{
					Field:   "answers",
					Message: "please answer all required questions",
				}
			}
		}
	}

	for _, ansReq := range req.Answers {
		question, err := s.questionRepo.GetByID(ctx, ansReq.QuestionID)
		if err != nil {
			return fmt.Errorf("failed to get question: %w", err)
		}

		if err := s.validateAnswer(ansReq, question); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) validateAnswer(ansReq AnswerRequest, question *models.PreferenceQuestion) error {
	switch question.QuestionType {
	case models.QuestionTypeText:
		if ansReq.AnswerText == nil {
			return &models.ValidationError{
				Field:   "answers",
				Message: fmt.Sprintf("question %d requires a text answer", question.ID),
			}
		}
		if len(*ansReq.AnswerText) > 500 {
			return &models.ValidationError{
				Field:   "answers",
				Message: "text answer cannot exceed 500 characters",
			}
		}

	case models.QuestionTypeSingleChoice, models.QuestionTypeMultipleChoice:
		if ansReq.AnswerOption == nil {
			return &models.ValidationError{
				Field:   "answers",
				Message: fmt.Sprintf("question %d requires a selection", question.ID),
			}
		}
		options, err := question.ParseOptions()
		if err != nil {
			return fmt.Errorf("failed to parse question options: %w", err)
		}
		valid := false
		for _, opt := range options {
			if opt == *ansReq.AnswerOption {
				valid = true
				break
			}
		}
		if !valid {
			return &models.ValidationError{
				Field:   "answers",
				Message: fmt.Sprintf("invalid option for question %d", question.ID),
			}
		}
	}

	return nil
}
