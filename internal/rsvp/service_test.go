package rsvp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockInviteService struct {
	getInviteByTokenFunc func(ctx context.Context, token string) (*models.Invite, error)
	updateStatusFunc     func(ctx context.Context, inviteID int64, status models.InviteStatus) error
}

func (m *mockInviteService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	if m.getInviteByTokenFunc != nil {
		return m.getInviteByTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockInviteService) UpdateStatus(ctx context.Context, inviteID int64, status models.InviteStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, inviteID, status)
	}
	return nil
}

type mockEventRepository struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.Event, error)
}

func (m *mockEventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockEventRepository) Create(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockEventRepository) Update(ctx context.Context, event *models.Event) error {
	return errors.New("not implemented")
}

func (m *mockEventRepository) UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error {
	return errors.New("not implemented")
}

func (m *mockEventRepository) UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error {
	return errors.New("not implemented")
}

func (m *mockEventRepository) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (m *mockEventRepository) List(ctx context.Context, filters repositories.ListFilters) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockEventRepository) GetByStatus(ctx context.Context, status models.EventStatus) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

func (m *mockEventRepository) GetEventsToArchive(ctx context.Context, daysAfterEvent int) ([]*models.Event, error) {
	return nil, errors.New("not implemented")
}

type mockRSVPRepository struct {
	createFunc        func(ctx context.Context, rsvp *models.RSVP) error
	getByInviteIDFunc func(ctx context.Context, inviteID int64) (*models.RSVP, error)
}

func (m *mockRSVPRepository) Create(ctx context.Context, rsvp *models.RSVP) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, rsvp)
	}
	return nil
}

func (m *mockRSVPRepository) GetByInviteID(ctx context.Context, inviteID int64) (*models.RSVP, error) {
	if m.getByInviteIDFunc != nil {
		return m.getByInviteIDFunc(ctx, inviteID)
	}
	return nil, &models.NotFoundError{Resource: "rsvp", ID: inviteID}
}

func (m *mockRSVPRepository) GetByID(ctx context.Context, id int64) (*models.RSVP, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRSVPRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.RSVP, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRSVPRepository) Update(ctx context.Context, rsvp *models.RSVP) error {
	return errors.New("not implemented")
}

func (m *mockRSVPRepository) GetStats(ctx context.Context, eventID int64) (*repositories.RSVPStats, error) {
	return nil, errors.New("not implemented")
}

type mockAnswerRepository struct {
	createFunc func(ctx context.Context, answer *models.RSVPAnswer) error
}

func (m *mockAnswerRepository) Create(ctx context.Context, answer *models.RSVPAnswer) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, answer)
	}
	return nil
}

func (m *mockAnswerRepository) GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAnswerRepository) GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAnswerRepository) Update(ctx context.Context, answer *models.RSVPAnswer) error {
	return errors.New("not implemented")
}

func (m *mockAnswerRepository) DeleteByRSVPID(ctx context.Context, rsvpID int64) error {
	return errors.New("not implemented")
}

type mockQuestionRepository struct {
	getByIDFunc      func(ctx context.Context, id int64) (*models.PreferenceQuestion, error)
	getByEventIDFunc func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
}

func (m *mockQuestionRepository) GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, &models.NotFoundError{Resource: "question", ID: id}
}

func (m *mockQuestionRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	if m.getByEventIDFunc != nil {
		return m.getByEventIDFunc(ctx, eventID)
	}
	return []*models.PreferenceQuestion{}, nil
}

func (m *mockQuestionRepository) Create(ctx context.Context, question *models.PreferenceQuestion) error {
	return errors.New("not implemented")
}

func (m *mockQuestionRepository) Update(ctx context.Context, question *models.PreferenceQuestion) error {
	return errors.New("not implemented")
}

func (m *mockQuestionRepository) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (m *mockQuestionRepository) Reorder(ctx context.Context, eventID int64, questionIDs []int64) error {
	return errors.New("not implemented")
}

type mockDatabase struct {
	execInTransactionFunc func(ctx context.Context, fn func(context.Context) error) error
}

func (m *mockDatabase) ExecInTransaction(ctx context.Context, fn func(context.Context) error) error {
	if m.execInTransactionFunc != nil {
		return m.execInTransactionFunc(ctx, fn)
	}
	return fn(ctx)
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestService_SubmitRSVP_ValidYesWithPlusOnes(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:          1,
				EventID:     1,
				MaxPlusOnes: 2,
				Status:      models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
				StartTime:    future,
			}, nil
		},
	}

	rsvpRepo := &mockRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return nil, &models.NotFoundError{Resource: "rsvp", ID: inviteID}
		},
		createFunc: func(ctx context.Context, rsvp *models.RSVP) error {
			rsvp.ID = 1
			rsvp.CreatedAt = time.Now()
			rsvp.UpdatedAt = time.Now()
			return nil
		},
	}

	questionRepo := &mockQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	answerRepo := &mockAnswerRepository{}
	db := &mockDatabase{}

	service := NewService(inviteService, eventRepo, rsvpRepo, answerRepo, questionRepo, db)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 2,
		Answers:  []AnswerRequest{},
	}

	rsvp, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	if rsvp.Response != models.RSVPResponseYes {
		t.Errorf("Expected response 'yes', got '%s'", rsvp.Response)
	}

	if rsvp.PlusOnes != 2 {
		t.Errorf("Expected plus_ones 2, got %d", rsvp.PlusOnes)
	}
}

func TestService_SubmitRSVP_InvalidToken(t *testing.T) {
	ctx := context.Background()

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return nil, errors.New("invalid token")
		},
	}

	service := NewService(inviteService, nil, nil, nil, nil, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "invalidtoken", req)

	if err == nil {
		t.Fatal("Expected error for invalid token")
	}
}

func TestService_SubmitRSVP_ExpiredInvite(t *testing.T) {
	ctx := context.Background()

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			past := time.Now().Add(-24 * time.Hour)
			return &models.Invite{
				ID:        1,
				EventID:   1,
				ExpiresAt: past,
				Status:    models.InviteStatusSent,
			}, nil
		},
	}

	service := NewService(inviteService, nil, nil, nil, nil, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "expiredtoken", req)

	if err == nil {
		t.Fatal("Expected error for expired invite")
	}
}

func TestService_SubmitRSVP_RevokedInvite(t *testing.T) {
	ctx := context.Background()

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusRevoked,
			}, nil
		},
	}

	service := NewService(inviteService, nil, nil, nil, nil, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "revokedtoken", req)

	if err == nil {
		t.Fatal("Expected error for revoked invite")
	}
}

func TestService_SubmitRSVP_InvalidResponse(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
			}, nil
		},
	}

	service := NewService(inviteService, eventRepo, nil, nil, nil, nil)

	testCases := []struct {
		name     string
		response string
	}{
		{"empty response", ""},
		{"invalid response", "invalid"},
		{"uppercase yes", "YES"},
		{"mixed case", "Yes"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &SubmitRSVPRequest{
				Response: tc.response,
				PlusOnes: 0,
			}

			_, err := service.SubmitRSVP(ctx, "validtoken", req)

			if err == nil {
				t.Errorf("Expected error for response '%s'", tc.response)
			}
		})
	}
}

func TestService_SubmitRSVP_PlusOnesExceedLimit(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:          1,
				EventID:     1,
				MaxPlusOnes: 2,
				Status:      models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
			}, nil
		},
	}

	questionRepo := &mockQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	service := NewService(inviteService, eventRepo, nil, nil, questionRepo, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 5,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for plus ones exceeding limit")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestService_SubmitRSVP_NegativePlusOnes(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
			}, nil
		},
	}

	questionRepo := &mockQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	service := NewService(inviteService, eventRepo, nil, nil, questionRepo, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: -1,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for negative plus ones")
	}
}

func TestService_SubmitRSVP_DeadlinePassed(t *testing.T) {
	ctx := context.Background()
	past := time.Now().Add(-24 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &past,
			}, nil
		},
	}

	service := NewService(inviteService, eventRepo, nil, nil, nil, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for deadline passed")
	}

	if !errors.Is(err, ErrDeadlinePassed) {
		t.Errorf("Expected ErrDeadlinePassed, got %v", err)
	}
}

func TestService_SubmitRSVP_DuplicateRSVP(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
			}, nil
		},
	}

	rsvpRepo := &mockRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return &models.RSVP{
				ID:       1,
				InviteID: inviteID,
				Response: models.RSVPResponseYes,
			}, nil
		},
	}

	questionRepo := &mockQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	service := NewService(inviteService, eventRepo, rsvpRepo, nil, questionRepo, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for duplicate RSVP")
	}

	if !errors.Is(err, ErrDuplicateRSVP) {
		t.Errorf("Expected ErrDuplicateRSVP, got %v", err)
	}
}

func TestService_SubmitRSVP_CancelledEvent(t *testing.T) {
	ctx := context.Background()

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:     1,
				Status: models.EventStatusCancelled,
			}, nil
		},
	}

	service := NewService(inviteService, eventRepo, nil, nil, nil, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for cancelled event")
	}
}

func TestService_SubmitRSVP_MissingRequiredAnswer(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
			}, nil
		},
	}

	rsvpRepo := &mockRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return nil, &models.NotFoundError{Resource: "rsvp", ID: inviteID}
		},
	}

	questionRepo := &mockQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{
				{
					ID:       1,
					EventID:  1,
					QuestionType: models.QuestionTypeText,
					Required: true,
				},
			}, nil
		},
	}

	service := NewService(inviteService, eventRepo, rsvpRepo, nil, questionRepo, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers:  []AnswerRequest{},
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for missing required answer")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestService_SubmitRSVP_InvalidAnswerType(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusSent,
			}, nil
		},
		updateStatusFunc: func(ctx context.Context, inviteID int64, status models.InviteStatus) error {
			return nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
			}, nil
		},
	}

	rsvpRepo := &mockRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return nil, &models.NotFoundError{Resource: "rsvp", ID: inviteID}
		},
	}

	questionRepo := &mockQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			q := &models.PreferenceQuestion{
				ID:           1,
				EventID:      1,
				QuestionType: models.QuestionTypeSingleChoice,
				Required:     true,
			}
			q.SetOptions([]string{"red", "blue", "green"})
			return []*models.PreferenceQuestion{q}, nil
		},
		getByIDFunc: func(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
			q := &models.PreferenceQuestion{
				ID:           1,
				EventID:      1,
				QuestionType: models.QuestionTypeSingleChoice,
				Required:     true,
			}
			q.SetOptions([]string{"red", "blue", "green"})
			return q, nil
		},
	}

	service := NewService(inviteService, eventRepo, rsvpRepo, nil, questionRepo, nil)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID: 1,
				AnswerText: strPtr("text answer for single choice question"),
			},
		},
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for invalid answer type")
	}
}

func TestService_SubmitRSVP_NoResponseAutoCorrectsPlusOnes(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:          1,
				EventID:     1,
				MaxPlusOnes: 2,
				Status:      models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
			}, nil
		},
	}

	rsvpRepo := &mockRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return nil, &models.NotFoundError{Resource: "rsvp", ID: inviteID}
		},
		createFunc: func(ctx context.Context, rsvp *models.RSVP) error {
			if rsvp.PlusOnes != 0 {
				t.Errorf("Expected plus_ones to be auto-corrected to 0 for 'no' response, got %d", rsvp.PlusOnes)
			}
			rsvp.ID = 1
			rsvp.CreatedAt = time.Now()
			rsvp.UpdatedAt = time.Now()
			return nil
		},
	}

	questionRepo := &mockQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			return []*models.PreferenceQuestion{}, nil
		},
	}

	answerRepo := &mockAnswerRepository{}
	db := &mockDatabase{}

	service := NewService(inviteService, eventRepo, rsvpRepo, answerRepo, questionRepo, db)

	req := &SubmitRSVPRequest{
		Response: "no",
		PlusOnes: 2,
	}

	rsvp, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp.PlusOnes != 0 {
		t.Errorf("Expected plus_ones to be 0 for 'no' response, got %d", rsvp.PlusOnes)
	}
}

func TestService_SubmitRSVP_WithValidAnswers(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:      1,
				EventID: 1,
				Status:  models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:           1,
				Status:       models.EventStatusPublished,
				RSVPDeadline: &future,
			}, nil
		},
	}

	rsvpRepo := &mockRSVPRepository{
		getByInviteIDFunc: func(ctx context.Context, inviteID int64) (*models.RSVP, error) {
			return nil, &models.NotFoundError{Resource: "rsvp", ID: inviteID}
		},
		createFunc: func(ctx context.Context, rsvp *models.RSVP) error {
			rsvp.ID = 1
			rsvp.CreatedAt = time.Now()
			rsvp.UpdatedAt = time.Now()
			return nil
		},
	}

	questionRepo := &mockQuestionRepository{
		getByEventIDFunc: func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
			q1 := &models.PreferenceQuestion{
				ID:           1,
				EventID:      1,
				QuestionType: models.QuestionTypeText,
				Required:     true,
			}
			q2 := &models.PreferenceQuestion{
				ID:           2,
				EventID:      1,
				QuestionType: models.QuestionTypeSingleChoice,
				Required:     false,
			}
			q2.SetOptions([]string{"red", "blue", "green"})
			q3 := &models.PreferenceQuestion{
				ID:           3,
				EventID:      1,
				QuestionType: models.QuestionTypeText,
				Required:     false,
			}
			return []*models.PreferenceQuestion{q1, q2, q3}, nil
		},
		getByIDFunc: func(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
			switch id {
			case 1:
				return &models.PreferenceQuestion{
					ID:           1,
					EventID:      1,
					QuestionType: models.QuestionTypeText,
					Required:     true,
				}, nil
			case 2:
				q := &models.PreferenceQuestion{
					ID:           2,
					EventID:      1,
					QuestionType: models.QuestionTypeSingleChoice,
					Required:     false,
				}
				q.SetOptions([]string{"red", "blue", "green"})
				return q, nil
			case 3:
				return &models.PreferenceQuestion{
					ID:           3,
					EventID:      1,
					QuestionType: models.QuestionTypeText,
					Required:     false,
				}, nil
			}
			return nil, &models.NotFoundError{Resource: "question", ID: id}
		},
	}

	answerCreated := 0
	answerRepo := &mockAnswerRepository{
		createFunc: func(ctx context.Context, answer *models.RSVPAnswer) error {
			answerCreated++
			answer.ID = int64(answerCreated)
			answer.CreatedAt = time.Now()
			answer.UpdatedAt = time.Now()
			return nil
		},
	}

	db := &mockDatabase{}

	service := NewService(inviteService, eventRepo, rsvpRepo, answerRepo, questionRepo, db)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID: 1,
				AnswerText: strPtr("Vegetarian"),
			},
			{
				QuestionID: 2,
				AnswerOption: strPtr("red"),
			},
			{
				QuestionID: 3,
				AnswerText: strPtr("No dietary restrictions"),
			},
		},
	}

	rsvp, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	if answerCreated != 3 {
		t.Errorf("Expected 3 answers to be created, got %d", answerCreated)
	}
}
