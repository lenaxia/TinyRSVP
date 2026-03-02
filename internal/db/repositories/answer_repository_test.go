package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupAnswerTestDB(t *testing.T) db.Database {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	_, err = database.Exec(ctx, `
		INSERT INTO users (id, email, name, role, created_at, updated_at)
		VALUES (1, 'test@example.com', 'Test User', 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return database
}

func createTestRSVP(t *testing.T, db db.Database) (int64, int64, int64) {
	t.Helper()

	ctx := context.Background()

	result, err := db.Exec(ctx, `
		INSERT INTO events (title, start_time, timezone, status, created_by, max_plus_ones)
		VALUES ('Test Event', ?, 'America/Los_Angeles', 'published', 1, 2)
	`, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}
	eventID, _ := result.LastInsertId()

	tokenHash := fmt.Sprintf("test_token_%d", time.Now().UnixNano())
	result, err = db.Exec(ctx, `
		INSERT INTO invites (event_id, name, email, token_hash, max_plus_ones, status, expires_at)
		VALUES (?, 'Test Guest', 'guest@example.com', ?, 2, 'sent', ?)
	`, eventID, tokenHash, time.Now().Add(48*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create test invite: %v", err)
	}
	inviteID, _ := result.LastInsertId()

	result, err = db.Exec(ctx, `
		INSERT INTO rsvps (invite_id, response, plus_ones)
		VALUES (?, 'yes', 1)
	`, inviteID)
	if err != nil {
		t.Fatalf("Failed to create test RSVP: %v", err)
	}
	rsvpID, _ := result.LastInsertId()

	return eventID, inviteID, rsvpID
}

func createTestQuestion(t *testing.T, db db.Database, eventID int64, questionType string) int64 {
	t.Helper()

	ctx := context.Background()

	validType := questionType
	if questionType == "select" {
		validType = "single_choice"
	} else if questionType == "boolean" {
		validType = "single_choice"
	}

	result, err := db.Exec(ctx, `
		INSERT INTO preference_questions (event_id, question_text, question_type, required, display_order)
		VALUES (?, 'Test Question', ?, 0, 0)
	`, eventID, validType)
	if err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	id, _ := result.LastInsertId()
	return id
}

func TestNewAnswerRepository(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	if repo == nil {
		t.Fatal("NewAnswerRepository returned nil")
	}
}

func TestAnswerRepository_Create_TextAnswer(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	eventID, _, rsvpID := createTestRSVP(t, database)
	questionID := createTestQuestion(t, database, eventID, "text")

	ctx := context.Background()
	text := "This is my answer"
	answer := &models.RSVPAnswer{
		RSVPID:     rsvpID,
		QuestionID: questionID,
		AnswerText: &text,
	}

	err := repo.Create(ctx, answer)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if answer.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}
	if answer.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestAnswerRepository_Create_OptionAnswer(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	eventID, _, rsvpID := createTestRSVP(t, database)
	questionID := createTestQuestion(t, database, eventID, "select")

	ctx := context.Background()
	option := "option1"
	answer := &models.RSVPAnswer{
		RSVPID:       rsvpID,
		QuestionID:   questionID,
		AnswerOption: &option,
	}

	err := repo.Create(ctx, answer)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if answer.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}
}

func TestAnswerRepository_Create_BooleanAnswer(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	eventID, _, rsvpID := createTestRSVP(t, database)
	questionID := createTestQuestion(t, database, eventID, "boolean")

	ctx := context.Background()
	boolVal := true
	answer := &models.RSVPAnswer{
		RSVPID:        rsvpID,
		QuestionID:    questionID,
		AnswerBoolean: &boolVal,
	}

	err := repo.Create(ctx, answer)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if answer.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}
}

func TestAnswerRepository_GetByRSVPID(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	eventID, _, rsvpID := createTestRSVP(t, database)

	q1 := createTestQuestion(t, database, eventID, "text")
	q2 := createTestQuestion(t, database, eventID, "boolean")

	ctx := context.Background()

	text := "Answer 1"
	answer1 := &models.RSVPAnswer{
		RSVPID:     rsvpID,
		QuestionID: q1,
		AnswerText: &text,
	}
	repo.Create(ctx, answer1)

	boolVal := false
	answer2 := &models.RSVPAnswer{
		RSVPID:        rsvpID,
		QuestionID:    q2,
		AnswerBoolean: &boolVal,
	}
	repo.Create(ctx, answer2)

	answers, err := repo.GetByRSVPID(ctx, rsvpID)
	if err != nil {
		t.Fatalf("GetByRSVPID() error = %v", err)
	}

	if len(answers) != 2 {
		t.Errorf("GetByRSVPID() returned %d answers, want 2", len(answers))
	}
}

func TestAnswerRepository_GetByQuestionID(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	eventID, _, rsvpID := createTestRSVP(t, database)
	questionID := createTestQuestion(t, database, eventID, "text")

	ctx := context.Background()

	text := "Answer from guest"
	answer := &models.RSVPAnswer{
		RSVPID:     rsvpID,
		QuestionID: questionID,
		AnswerText: &text,
	}
	repo.Create(ctx, answer)

	answers, err := repo.GetByQuestionID(ctx, questionID)
	if err != nil {
		t.Fatalf("GetByQuestionID() error = %v", err)
	}

	if len(answers) != 1 {
		t.Errorf("GetByQuestionID() returned %d answers, want 1", len(answers))
	}

	if answers[0].QuestionID != questionID {
		t.Errorf("QuestionID = %d, want %d", answers[0].QuestionID, questionID)
	}
}

func TestAnswerRepository_Update(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	eventID, _, rsvpID := createTestRSVP(t, database)
	questionID := createTestQuestion(t, database, eventID, "text")

	ctx := context.Background()

	text := "Original answer"
	answer := &models.RSVPAnswer{
		RSVPID:     rsvpID,
		QuestionID: questionID,
		AnswerText: &text,
	}
	repo.Create(ctx, answer)

	time.Sleep(1100 * time.Millisecond)

	newText := "Updated answer"
	answer.AnswerText = &newText

	err := repo.Update(ctx, answer)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	answers, _ := repo.GetByRSVPID(ctx, rsvpID)
	if len(answers) == 0 {
		t.Fatal("No answers found after update")
	}

	if *answers[0].AnswerText != newText {
		t.Errorf("AnswerText = %s, want %s", *answers[0].AnswerText, newText)
	}

	if !answers[0].UpdatedAt.After(answers[0].CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestAnswerRepository_DeleteByRSVPID(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	eventID, _, rsvpID := createTestRSVP(t, database)

	q1 := createTestQuestion(t, database, eventID, "text")
	q2 := createTestQuestion(t, database, eventID, "boolean")

	ctx := context.Background()

	text := "Answer 1"
	answer1 := &models.RSVPAnswer{
		RSVPID:     rsvpID,
		QuestionID: q1,
		AnswerText: &text,
	}
	repo.Create(ctx, answer1)

	boolVal := true
	answer2 := &models.RSVPAnswer{
		RSVPID:        rsvpID,
		QuestionID:    q2,
		AnswerBoolean: &boolVal,
	}
	repo.Create(ctx, answer2)

	err := repo.DeleteByRSVPID(ctx, rsvpID)
	if err != nil {
		t.Fatalf("DeleteByRSVPID() error = %v", err)
	}

	answers, _ := repo.GetByRSVPID(ctx, rsvpID)
	if len(answers) != 0 {
		t.Errorf("GetByRSVPID() returned %d answers after delete, want 0", len(answers))
	}
}

func TestAnswerRepository_Create_DuplicateAnswer(t *testing.T) {
	database := setupAnswerTestDB(t)
	defer database.Close()

	repo := NewAnswerRepository(database)
	eventID, _, rsvpID := createTestRSVP(t, database)
	questionID := createTestQuestion(t, database, eventID, "text")

	ctx := context.Background()

	text1 := "First answer"
	answer1 := &models.RSVPAnswer{
		RSVPID:     rsvpID,
		QuestionID: questionID,
		AnswerText: &text1,
	}
	repo.Create(ctx, answer1)

	text2 := "Second answer"
	answer2 := &models.RSVPAnswer{
		RSVPID:     rsvpID,
		QuestionID: questionID,
		AnswerText: &text2,
	}

	err := repo.Create(ctx, answer2)
	if err == nil {
		t.Error("Expected error for duplicate (rsvp_id, question_id), got nil")
	}
}
