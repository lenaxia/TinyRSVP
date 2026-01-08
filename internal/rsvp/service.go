package rsvp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/email"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

var (
	ErrDuplicateRSVP = errors.New("you have already responded to this invite")
)

type InviteService interface {
	GetInviteByToken(ctx context.Context, token string) (*models.Invite, error)
}

type InviteRepository interface {
	Update(ctx context.Context, invite *models.Invite) error
}

type Service interface {
	SubmitRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) (*models.RSVP, error)
	UpdateRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) (*models.RSVP, error)
}

type service struct {
	db                db.Database
	inviteService     InviteService
	inviteRepo        InviteRepository
	eventRepo         repositories.EventRepository
	rsvpRepo          repositories.RSVPRepository
	answerRepo        repositories.AnswerRepository
	questionRepo      repositories.QuestionRepository
	plusOnesValidator PlusOnesValidator
	emailService      email.Service
}

func NewService(
	database db.Database,
	inviteService InviteService,
	inviteRepo InviteRepository,
	eventRepo repositories.EventRepository,
	rsvpRepo repositories.RSVPRepository,
	answerRepo repositories.AnswerRepository,
	questionRepo repositories.QuestionRepository,
) Service {
	return &service{
		db:                database,
		inviteService:     inviteService,
		inviteRepo:        inviteRepo,
		eventRepo:         eventRepo,
		rsvpRepo:          rsvpRepo,
		answerRepo:        answerRepo,
		questionRepo:      questionRepo,
		plusOnesValidator: NewValidator(),
		emailService:      nil,
	}
}

func NewServiceWithEmail(
	database db.Database,
	inviteService InviteService,
	inviteRepo InviteRepository,
	eventRepo repositories.EventRepository,
	rsvpRepo repositories.RSVPRepository,
	answerRepo repositories.AnswerRepository,
	questionRepo repositories.QuestionRepository,
	emailService email.Service,
) Service {
	return &service{
		db:                database,
		inviteService:     inviteService,
		inviteRepo:        inviteRepo,
		eventRepo:         eventRepo,
		rsvpRepo:          rsvpRepo,
		answerRepo:        answerRepo,
		questionRepo:      questionRepo,
		plusOnesValidator: NewValidator(),
		emailService:      emailService,
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

func checkDeadline(event *models.Event) error {
	if event.RSVPDeadline == nil {
		return nil
	}

	now := time.Now().UTC()
	deadline := event.RSVPDeadline.UTC()

	if !deadline.After(now) {
		return &models.DeadlinePassedError{
			Deadline: deadline,
			Message:  "RSVP deadline has passed",
		}
	}

	return nil
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

	if err := checkDeadline(event); err != nil {
		return nil, err
	}

	if req.Response == "no" && req.PlusOnes > 0 {
		req.PlusOnes = 0
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

	var rsvp *models.RSVP
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		rsvpModel := &models.RSVP{
			InviteID: invite.ID,
			Response: models.RSVPResponse(req.Response),
			PlusOnes: req.PlusOnes,
		}

		if err := rsvpModel.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		result, err := tx.ExecContext(ctx,
			`INSERT INTO rsvps (invite_id, response, plus_ones, created_at, updated_at)
			 VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			rsvpModel.InviteID, rsvpModel.Response, rsvpModel.PlusOnes)
		if err != nil {
			return fmt.Errorf("failed to create RSVP: %w", err)
		}

		rsvpID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get RSVP ID: %w", err)
		}
		rsvpModel.ID = rsvpID

		for _, ansReq := range req.Answers {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO rsvp_answers (rsvp_id, question_id, answer_text, answer_option, answer_boolean, created_at)
				 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
				rsvpID, ansReq.QuestionID, ansReq.AnswerText, ansReq.AnswerOption, ansReq.AnswerBoolean)
			if err != nil {
				return fmt.Errorf("failed to create answer: %w", err)
			}
		}

		_, err = tx.ExecContext(ctx,
			`UPDATE invites SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			models.InviteStatusResponded, invite.ID)
		if err != nil {
			return fmt.Errorf("failed to update invite status: %w", err)
		}

		err = tx.QueryRowContext(ctx,
			`SELECT id, invite_id, response, plus_ones, created_at, updated_at FROM rsvps WHERE id = ?`,
			rsvpID).Scan(&rsvpModel.ID, &rsvpModel.InviteID, &rsvpModel.Response,
			&rsvpModel.PlusOnes, &rsvpModel.CreatedAt, &rsvpModel.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to retrieve created RSVP: %w", err)
		}

		rsvp = rsvpModel
		return nil
	})

	if err != nil {
		return nil, err
	}

	if s.emailService != nil && invite.Email != nil {
		answers, err := s.answerRepo.GetByRSVPID(ctx, rsvp.ID)
		if err != nil {
			log.Printf("Failed to get answers for confirmation email: %v", err)
		} else {
			go func() {
				if err := s.emailService.SendConfirmationEmail(context.Background(), token, rsvp, invite, event, answers); err != nil {
					log.Printf("Failed to send confirmation email: %v", err)
				}
			}()
		}
	}

	return rsvp, nil
}

func (s *service) validateRequest(ctx context.Context, req *SubmitRSVPRequest, invite *models.Invite, event *models.Event) error {
	response := models.RSVPResponse(req.Response)
	if !response.Valid() {
		return &models.ValidationError{
			Field:   "response",
			Message: "response must be yes, no, or maybe",
		}
	}

	if err := s.plusOnesValidator.ValidatePlusOnes(req.PlusOnes, response, invite); err != nil {
		return err
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
		if len(*ansReq.AnswerText) == 0 && question.Required {
			return &models.ValidationError{
				Field:   "answers",
				Message: fmt.Sprintf("question %d requires a non-empty text answer", question.ID),
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

func (s *service) UpdateRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) (*models.RSVP, error) {
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

	existing, err := s.rsvpRepo.GetByInviteID(ctx, invite.ID)
	if err != nil {
		return nil, err
	}

	event, err := s.eventRepo.GetByID(ctx, invite.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if event.Status == models.EventStatusCancelled {
		return nil, errors.New("event has been cancelled")
	}

	if err := checkDeadline(event); err != nil {
		return nil, err
	}

	if req.Response == "no" && req.PlusOnes > 0 {
		req.PlusOnes = 0
	}

	if err := s.validateRequest(ctx, req, invite, event); err != nil {
		return nil, err
	}

	var rsvp *models.RSVP
	err = s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		existing.Response = models.RSVPResponse(req.Response)
		existing.PlusOnes = req.PlusOnes

		if err := existing.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		_, err := tx.ExecContext(ctx,
			`UPDATE rsvps SET response = ?, plus_ones = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			existing.Response, existing.PlusOnes, existing.ID)
		if err != nil {
			return fmt.Errorf("failed to update RSVP: %w", err)
		}

		_, err = tx.ExecContext(ctx,
			`DELETE FROM rsvp_answers WHERE rsvp_id = ?`,
			existing.ID)
		if err != nil {
			return fmt.Errorf("failed to delete old answers: %w", err)
		}

		for _, ansReq := range req.Answers {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO rsvp_answers (rsvp_id, question_id, answer_text, answer_option, answer_boolean, created_at)
				 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
				existing.ID, ansReq.QuestionID, ansReq.AnswerText, ansReq.AnswerOption, ansReq.AnswerBoolean)
			if err != nil {
				return fmt.Errorf("failed to create answer: %w", err)
			}
		}

		err = tx.QueryRowContext(ctx,
			`SELECT id, invite_id, response, plus_ones, created_at, updated_at FROM rsvps WHERE id = ?`,
			existing.ID).Scan(&existing.ID, &existing.InviteID, &existing.Response,
			&existing.PlusOnes, &existing.CreatedAt, &existing.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to retrieve updated RSVP: %w", err)
		}

		rsvp = existing
		return nil
	})

	if err != nil {
		return nil, err
	}

	if s.emailService != nil && invite.Email != nil {
		answers, err := s.answerRepo.GetByRSVPID(ctx, rsvp.ID)
		if err != nil {
			log.Printf("Failed to get answers for confirmation email: %v", err)
		} else {
			go func() {
				if err := s.emailService.SendConfirmationEmail(context.Background(), token, rsvp, invite, event, answers); err != nil {
					log.Printf("Failed to send confirmation email: %v", err)
				}
			}()
		}
	}

	return rsvp, nil
}
