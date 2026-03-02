package rsvp

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupTestDB(t *testing.T) db.Database {
	t.Helper()
	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return database
}

type mockInviteService struct {
	getInviteByTokenFunc func(ctx context.Context, token string) (*models.Invite, error)
}

func (m *mockInviteService) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	if m.getInviteByTokenFunc != nil {
		return m.getInviteByTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func hashToken(token string) string {
	return "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"
}

func TestService_SubmitRSVP_ValidYesWithPlusOnes(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

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

	updatedInvite, err := inviteRepo.GetByID(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to get updated invite: %v", err)
	}

	if updatedInvite.Status != models.InviteStatusResponded {
		t.Errorf("Expected invite status 'responded', got '%s'", updatedInvite.Status)
	}
}

func TestService_SubmitRSVP_InvalidToken(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return nil, errors.New("invalid token")
		},
	}

	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

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
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	past := time.Now().Add(-24 * time.Hour)

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{
				ID:        1,
				EventID:   1,
				ExpiresAt: past,
				Status:    models.InviteStatusSent,
			}, nil
		},
	}

	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "expiredtoken", req)

	if err == nil {
		t.Fatal("Expected error for expired invite")
	}

	if err.Error() != "invite has expired" {
		t.Errorf("Expected 'invite has expired' error, got '%v'", err)
	}
}

func TestService_SubmitRSVP_RevokedInvite(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

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

	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "revokedtoken", req)

	if err == nil {
		t.Fatal("Expected error for revoked invite")
	}

	if err.Error() != "invite has been revoked" {
		t.Errorf("Expected 'invite has been revoked' error, got '%v'", err)
	}
}

func TestService_SubmitRSVP_InvalidResponse(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "INVALID",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for invalid response")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "response" {
		t.Errorf("Expected field 'response', got '%s'", validationErr.Field)
	}
}

func TestService_SubmitRSVP_NegativePlusOnes(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: -1,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for negative plus_ones")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "plus_ones" {
		t.Errorf("Expected field 'plus_ones', got '%s'", validationErr.Field)
	}
}

func TestService_SubmitRSVP_ExceedMaxPlusOnes(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 5,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for exceeding max plus_ones")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "plus_ones" {
		t.Errorf("Expected field 'plus_ones', got '%s'", validationErr.Field)
	}
}

func TestService_SubmitRSVP_DeadlinePassed(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(48 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &past,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for deadline passed")
	}

	var deadlineErr *models.DeadlinePassedError
	if !errors.As(err, &deadlineErr) {
		t.Errorf("Expected DeadlinePassedError, got %T: %v", err, err)
	}
}

func TestService_SubmitRSVP_DuplicateRSVP(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	existingRSVP := &models.RSVP{
		InviteID: invite.ID,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}
	if err := rsvpRepo.Create(ctx, existingRSVP); err != nil {
		t.Fatalf("Failed to create existing RSVP: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

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
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusCancelled,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for cancelled event")
	}

	if err.Error() != "event has been cancelled" {
		t.Errorf("Expected 'event has been cancelled' error, got '%v'", err)
	}
}

func TestService_SubmitRSVP_MissingRequiredAnswer(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Required question",
		QuestionType: models.QuestionTypeText,
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

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
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "answers" {
		t.Errorf("Expected field 'answers', got '%s'", validationErr.Field)
	}
}

func TestService_SubmitRSVP_InvalidAnswerType(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Text question",
		QuestionType: models.QuestionTypeText,
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID:   question.ID,
				AnswerOption: strPtr("wrong type"),
			},
		},
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for wrong answer type")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}
}

func TestService_SubmitRSVP_AutoCorrectPlusOnesForNo(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 5,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "no",
		PlusOnes: 3,
	}

	rsvp, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp.PlusOnes != 0 {
		t.Errorf("Expected plus_ones to be auto-corrected to 0, got %d", rsvp.PlusOnes)
	}
}

func TestService_SubmitRSVP_WithAnswers(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	textQuestion := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Text question",
		QuestionType: models.QuestionTypeText,
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, textQuestion); err != nil {
		t.Fatalf("Failed to create text question: %v", err)
	}

	choiceQuestion := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Choice question",
		QuestionType: models.QuestionTypeSingleChoice,
		Options:      strPtr(`["Option A","Option B"]`),
		Required:     false,
		DisplayOrder: 2,
	}
	if err := questionRepo.Create(ctx, choiceQuestion); err != nil {
		t.Fatalf("Failed to create choice question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 1,
		Answers: []AnswerRequest{
			{
				QuestionID: textQuestion.ID,
				AnswerText: strPtr("My text answer"),
			},
			{
				QuestionID:   choiceQuestion.ID,
				AnswerOption: strPtr("Option A"),
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

	answers, err := answerRepo.GetByRSVPID(ctx, rsvp.ID)
	if err != nil {
		t.Fatalf("Failed to get answers: %v", err)
	}

	if len(answers) != 2 {
		t.Errorf("Expected 2 answers, got %d", len(answers))
	}
}

func TestService_SubmitRSVP_TransactionRollback(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Required question",
		QuestionType: models.QuestionTypeText,
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 1,
		Answers: []AnswerRequest{
			{
				QuestionID: 99999,
				AnswerText: strPtr("Answer to non-existent question"),
			},
		},
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for invalid question ID")
	}

	existingRSVP, err := rsvpRepo.GetByInviteID(ctx, invite.ID)
	if err == nil {
		t.Errorf("Expected no RSVP to be created due to rollback, but found RSVP with ID %d", existingRSVP.ID)
	}

	updatedInvite, err := inviteRepo.GetByID(ctx, invite.ID)
	if err != nil {
		t.Fatalf("Failed to get invite: %v", err)
	}

	if updatedInvite.Status != models.InviteStatusSent {
		t.Errorf("Expected invite status to remain 'sent' after rollback, got '%s'", updatedInvite.Status)
	}
}

func TestService_SubmitRSVP_TextAnswerTooLong(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Text question",
		QuestionType: models.QuestionTypeText,
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	longText := strings.Repeat("a", 501)
	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID: question.ID,
				AnswerText: strPtr(longText),
			},
		},
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for text answer exceeding 500 characters")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "answers" {
		t.Errorf("Expected field 'answers', got '%s'", validationErr.Field)
	}

	if !strings.Contains(validationErr.Message, "500 characters") {
		t.Errorf("Expected error message about 500 characters, got '%s'", validationErr.Message)
	}

	existingRSVP, err := rsvpRepo.GetByInviteID(ctx, invite.ID)
	if err == nil {
		t.Errorf("Expected no RSVP to be saved, but found RSVP with ID %d", existingRSVP.ID)
	}
}

func TestService_SubmitRSVP_InvalidChoiceOption(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Choice question",
		QuestionType: models.QuestionTypeSingleChoice,
		Options:      strPtr(`["Option A","Option B"]`),
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID:   question.ID,
				AnswerOption: strPtr("Invalid Option"),
			},
		},
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for invalid choice option")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "answers" {
		t.Errorf("Expected field 'answers', got '%s'", validationErr.Field)
	}

	if !strings.Contains(validationErr.Message, "invalid option") {
		t.Errorf("Expected error message about invalid option, got '%s'", validationErr.Message)
	}

	existingRSVP, err := rsvpRepo.GetByInviteID(ctx, invite.ID)
	if err == nil {
		t.Errorf("Expected no RSVP to be saved, but found RSVP with ID %d", existingRSVP.ID)
	}
}

func TestService_SubmitRSVP_EmptyTextForRequired(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Required text question",
		QuestionType: models.QuestionTypeText,
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID: question.ID,
				AnswerText: strPtr(""),
			},
		},
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for empty text answer on required question")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Field != "answers" {
		t.Errorf("Expected field 'answers', got '%s'", validationErr.Field)
	}

	if !strings.Contains(validationErr.Message, "non-empty") {
		t.Errorf("Expected error message about non-empty text, got '%s'", validationErr.Message)
	}
}

func TestService_SubmitRSVP_MultipleChoiceValidation(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Multiple choice question",
		QuestionType: models.QuestionTypeMultipleChoice,
		Options:      strPtr(`["A","B","C"]`),
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID:   question.ID,
				AnswerOption: strPtr("B"),
			},
		},
	}

	rsvp, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error for valid multiple choice answer, got %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	invite2 := &models.Invite{
		EventID:   event.ID,
		TokenHash: "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ9876543",
		Email:     strPtr("guest2@example.com"),
		Name:      strPtr("Test Guest 2"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite2); err != nil {
		t.Fatalf("Failed to create second test invite: %v", err)
	}

	inviteService2 := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite2, nil
		},
	}

	service2 := NewService(database, inviteService2, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req2 := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID:   question.ID,
				AnswerOption: strPtr("Invalid"),
			},
		},
	}

	_, err = service2.SubmitRSVP(ctx, "validtoken2", req2)

	if err == nil {
		t.Fatal("Expected error for invalid multiple choice option")
	}

	var validationErr *models.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Expected ValidationError, got %T", err)
	}
}

func TestService_SubmitRSVP_CaseSensitiveOptions(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Case sensitive question",
		QuestionType: models.QuestionTypeSingleChoice,
		Options:      strPtr(`["Red","Blue","Green"]`),
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:   event.ID,
		TokenHash: hashToken("validtoken"),
		Email:     strPtr("guest@example.com"),
		Name:      strPtr("Test Guest"),

		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
		Answers: []AnswerRequest{
			{
				QuestionID:   question.ID,
				AnswerOption: strPtr("red"),
			},
		},
	}

	_, err := service.SubmitRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Logf("Case sensitivity enforced: lowercase 'red' rejected when option is 'Red'")
		var validationErr *models.ValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("Expected ValidationError, got %T", err)
		}
		if !strings.Contains(validationErr.Message, "invalid option") {
			t.Errorf("Expected error message about invalid option, got '%s'", validationErr.Message)
		}
	} else {
		t.Fatal("Expected error for case mismatch - options should be case-sensitive")
	}
}

func TestService_UpdateRSVP_ValidUpdate(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:          "Test Event",
		Description:    strPtr("Test Description"),
		StartTime:      future,
		Timezone:       "UTC",
		Status:         models.EventStatusPublished,
		RSVPDeadline:   &rsvpDeadline,
		AllowMaybeRSVP: true,
		CreatedBy:      1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       strPtr("guest@example.com"),
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusResponded,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	existingRSVP := &models.RSVP{
		InviteID: invite.ID,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}
	if err := rsvpRepo.Create(ctx, existingRSVP); err != nil {
		t.Fatalf("Failed to create existing RSVP: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "maybe",
		PlusOnes: 2,
		Answers:  []AnswerRequest{},
	}

	rsvp, err := service.UpdateRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	if rsvp.Response != models.RSVPResponseMaybe {
		t.Errorf("Expected response 'maybe', got '%s'", rsvp.Response)
	}

	if rsvp.PlusOnes != 2 {
		t.Errorf("Expected plus_ones 2, got %d", rsvp.PlusOnes)
	}
}

func TestService_UpdateRSVP_NoExistingRSVP(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       strPtr("guest@example.com"),
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusSent,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 1,
	}

	_, err := service.UpdateRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for no existing RSVP")
	}

	var notFoundErr *models.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("Expected NotFoundError, got %T: %v", err, err)
	}
}

func TestService_UpdateRSVP_DeadlinePassed(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(48 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &past,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       strPtr("guest@example.com"),
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusResponded,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	existingRSVP := &models.RSVP{
		InviteID: invite.ID,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}
	if err := rsvpRepo.Create(ctx, existingRSVP); err != nil {
		t.Fatalf("Failed to create existing RSVP: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "no",
		PlusOnes: 0,
	}

	_, err := service.UpdateRSVP(ctx, "validtoken", req)

	if err == nil {
		t.Fatal("Expected error for deadline passed")
	}

	var deadlineErr *models.DeadlinePassedError
	if !errors.As(err, &deadlineErr) {
		t.Errorf("Expected DeadlinePassedError, got %T: %v", err, err)
	}
}

func TestService_UpdateRSVP_WithAnswers(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	questionRepo := repositories.NewQuestionRepository(database)
	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Text question",
		QuestionType: models.QuestionTypeText,
		Required:     true,
		DisplayOrder: 1,
	}
	if err := questionRepo.Create(ctx, question); err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       strPtr("guest@example.com"),
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusResponded,
		MaxPlusOnes: 2,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	existingRSVP := &models.RSVP{
		InviteID: invite.ID,
		Response: models.RSVPResponseYes,
		PlusOnes: 1,
	}
	if err := rsvpRepo.Create(ctx, existingRSVP); err != nil {
		t.Fatalf("Failed to create existing RSVP: %v", err)
	}

	answerRepo := repositories.NewAnswerRepository(database)
	oldAnswer := &models.RSVPAnswer{
		RSVPID:     existingRSVP.ID,
		QuestionID: question.ID,
		AnswerText: strPtr("Old answer"),
	}
	if err := answerRepo.Create(ctx, oldAnswer); err != nil {
		t.Fatalf("Failed to create old answer: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 2,
		Answers: []AnswerRequest{
			{
				QuestionID: question.ID,
				AnswerText: strPtr("New answer"),
			},
		},
	}

	rsvp, err := service.UpdateRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp == nil {
		t.Fatal("Expected RSVP to be returned")
	}

	answers, err := answerRepo.GetByRSVPID(ctx, rsvp.ID)
	if err != nil {
		t.Fatalf("Failed to get answers: %v", err)
	}

	if len(answers) != 1 {
		t.Errorf("Expected 1 answer, got %d", len(answers))
	}

	if answers[0].AnswerText == nil || *answers[0].AnswerText != "New answer" {
		t.Errorf("Expected answer text 'New answer', got '%v'", answers[0].AnswerText)
	}
}

func TestService_UpdateRSVP_InvalidToken(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return nil, errors.New("invalid token")
		},
	}

	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "yes",
		PlusOnes: 0,
	}

	_, err := service.UpdateRSVP(ctx, "invalidtoken", req)

	if err == nil {
		t.Fatal("Expected error for invalid token")
	}
}

func TestService_UpdateRSVP_ChangeToNo(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	future := time.Now().Add(48 * time.Hour)
	rsvpDeadline := time.Now().Add(24 * time.Hour)

	eventRepo := repositories.NewEventRepository(database)
	event := &models.Event{
		Title:        "Test Event",
		Description:  strPtr("Test Description"),
		StartTime:    future,
		Timezone:     "UTC",
		Status:       models.EventStatusPublished,
		RSVPDeadline: &rsvpDeadline,
		CreatedBy:    1,
	}
	if err := eventRepo.Create(ctx, event); err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	inviteRepo := repositories.NewInviteRepository(database)
	invite := &models.Invite{
		EventID:     event.ID,
		TokenHash:   hashToken("validtoken"),
		Email:       strPtr("guest@example.com"),
		Name:        strPtr("Test Guest"),
		Status:      models.InviteStatusResponded,
		MaxPlusOnes: 5,
		ExpiresAt:   future,
	}
	if err := inviteRepo.Create(ctx, invite); err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}

	rsvpRepo := repositories.NewRSVPRepository(database)
	existingRSVP := &models.RSVP{
		InviteID: invite.ID,
		Response: models.RSVPResponseYes,
		PlusOnes: 3,
	}
	if err := rsvpRepo.Create(ctx, existingRSVP); err != nil {
		t.Fatalf("Failed to create existing RSVP: %v", err)
	}

	inviteService := &mockInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return invite, nil
		},
	}

	answerRepo := repositories.NewAnswerRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)

	service := NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)

	req := &SubmitRSVPRequest{
		Response: "no",
		PlusOnes: 2,
	}

	rsvp, err := service.UpdateRSVP(ctx, "validtoken", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if rsvp.Response != models.RSVPResponseNo {
		t.Errorf("Expected response 'no', got '%s'", rsvp.Response)
	}

	if rsvp.PlusOnes != 0 {
		t.Errorf("Expected plus_ones to be auto-corrected to 0, got %d", rsvp.PlusOnes)
	}
}

func TestCheckDeadline_NoDeadlineSet(t *testing.T) {
	event := &models.Event{
		RSVPDeadline: nil,
	}

	err := checkDeadline(event)

	if err != nil {
		t.Errorf("Expected no error when deadline is not set, got %v", err)
	}
}

func TestCheckDeadline_DeadlineInFuture(t *testing.T) {
	future := time.Now().UTC().Add(24 * time.Hour)
	event := &models.Event{
		RSVPDeadline: &future,
	}

	err := checkDeadline(event)

	if err != nil {
		t.Errorf("Expected no error when deadline is in future, got %v", err)
	}
}

func TestCheckDeadline_DeadlineInPast(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour)
	event := &models.Event{
		RSVPDeadline: &past,
	}

	err := checkDeadline(event)

	if err == nil {
		t.Fatal("Expected error when deadline is in past")
	}

	var deadlineErr *models.DeadlinePassedError
	if !errors.As(err, &deadlineErr) {
		t.Errorf("Expected DeadlinePassedError, got %T", err)
	}

	if deadlineErr != nil && !deadlineErr.Deadline.Equal(past) {
		t.Errorf("Expected deadline %v, got %v", past, deadlineErr.Deadline)
	}
}

func TestCheckDeadline_DeadlineExactlyNow(t *testing.T) {
	now := time.Now().UTC()
	event := &models.Event{
		RSVPDeadline: &now,
	}

	err := checkDeadline(event)

	if err == nil {
		t.Fatal("Expected error when deadline is exactly now")
	}

	var deadlineErr *models.DeadlinePassedError
	if !errors.As(err, &deadlineErr) {
		t.Errorf("Expected DeadlinePassedError, got %T", err)
	}
}

func TestCheckDeadline_TimezoneAware(t *testing.T) {
	tests := []struct {
		name     string
		deadline time.Time
		wantErr  bool
	}{
		{
			name:     "UTC past deadline",
			deadline: time.Now().UTC().Add(-1 * time.Hour),
			wantErr:  true,
		},
		{
			name:     "UTC future deadline",
			deadline: time.Now().UTC().Add(1 * time.Hour),
			wantErr:  false,
		},
		{
			name:     "PST past deadline converted to UTC",
			deadline: time.Date(2026, 1, 1, 10, 0, 0, 0, time.FixedZone("PST", -8*3600)),
			wantErr:  true,
		},
		{
			name:     "EST future deadline converted to UTC",
			deadline: time.Now().In(time.FixedZone("EST", -5*3600)).Add(2 * time.Hour),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &models.Event{
				RSVPDeadline: &tt.deadline,
			}

			err := checkDeadline(event)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				var deadlineErr *models.DeadlinePassedError
				if !errors.As(err, &deadlineErr) {
					t.Errorf("Expected DeadlinePassedError, got %T", err)
				}
			}
		})
	}
}
