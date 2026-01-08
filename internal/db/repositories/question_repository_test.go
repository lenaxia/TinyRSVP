package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupQuestionTestDB(t *testing.T) db.Database {
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

func TestQuestionRepository_Create(t *testing.T) {
	database := setupQuestionTestDB(t)
	defer database.Close()

	repo := NewQuestionRepository(database)
	eventRepo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 2,
	}
	err := eventRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	tests := []struct {
		name     string
		question *models.PreferenceQuestion
		wantErr  bool
	}{
		{
			name: "create text question",
			question: &models.PreferenceQuestion{
				EventID:      event.ID,
				QuestionText: "What is your dietary preference?",
				QuestionType: models.QuestionTypeText,
				Required:     true,
				DisplayOrder: 1,
			},
			wantErr: false,
		},
		{
			name: "create single choice question",
			question: &models.PreferenceQuestion{
				EventID:      event.ID,
				QuestionText: "Will you attend?",
				QuestionType: models.QuestionTypeSingleChoice,
				Required:     true,
				DisplayOrder: 2,
				Options:      stringPtr(`["Yes", "No", "Maybe"]`),
			},
			wantErr: false,
		},
		{
			name: "create multiple choice question",
			question: &models.PreferenceQuestion{
				EventID:      event.ID,
				QuestionText: "Select dietary restrictions",
				QuestionType: models.QuestionTypeMultipleChoice,
				Required:     false,
				DisplayOrder: 3,
				Options:      stringPtr(`["Vegetarian", "Vegan", "Gluten-free"]`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.question)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.question.ID == 0 {
					t.Error("Expected ID to be set after create")
				}
				if tt.question.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.question.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestQuestionRepository_GetByID(t *testing.T) {
	database := setupQuestionTestDB(t)
	defer database.Close()

	repo := NewQuestionRepository(database)
	eventRepo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 2,
	}
	err := eventRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Test question",
		QuestionType: models.QuestionTypeText,
		Required:     false,
		DisplayOrder: 0,
	}
	err = repo.Create(context.Background(), question)
	if err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "get existing question",
			id:      question.ID,
			wantErr: false,
		},
		{
			name:    "get non-existent question",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.ID != tt.id {
					t.Errorf("GetByID() ID = %v, want %v", got.ID, tt.id)
				}
				if got.QuestionText != question.QuestionText {
					t.Errorf("GetByID() QuestionText = %v, want %v", got.QuestionText, question.QuestionText)
				}
			}
		})
	}
}

func TestQuestionRepository_GetByEventID(t *testing.T) {
	database := setupQuestionTestDB(t)
	defer database.Close()

	repo := NewQuestionRepository(database)
	eventRepo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 2,
	}
	err := eventRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	q1 := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Question 1",
		QuestionType: models.QuestionTypeText,
		DisplayOrder: 1,
	}
	repo.Create(context.Background(), q1)

	q2 := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Question 2",
		QuestionType: models.QuestionTypeText,
		DisplayOrder: 2,
	}
	repo.Create(context.Background(), q2)

	questions, err := repo.GetByEventID(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("GetByEventID() error = %v", err)
	}

	if len(questions) != 2 {
		t.Errorf("GetByEventID() returned %d questions, want 2", len(questions))
	}

	if questions[0].DisplayOrder > questions[1].DisplayOrder {
		t.Error("Questions not ordered by display_order")
	}
}

func TestQuestionRepository_Update(t *testing.T) {
	database := setupQuestionTestDB(t)
	defer database.Close()

	repo := NewQuestionRepository(database)
	eventRepo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 2,
	}
	err := eventRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Original text",
		QuestionType: models.QuestionTypeText,
		Required:     false,
		DisplayOrder: 0,
	}
	err = repo.Create(context.Background(), question)
	if err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	question.QuestionText = "Updated question text"
	question.Required = true

	err = repo.Update(context.Background(), question)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repo.GetByID(context.Background(), question.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if updated.QuestionText != "Updated question text" {
		t.Errorf("Update() QuestionText = %v, want %v", updated.QuestionText, "Updated question text")
	}

	if !updated.Required {
		t.Error("Update() Required should be true")
	}
}

func TestQuestionRepository_Delete(t *testing.T) {
	database := setupQuestionTestDB(t)
	defer database.Close()

	repo := NewQuestionRepository(database)
	eventRepo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 2,
	}
	err := eventRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	question := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Test question",
		QuestionType: models.QuestionTypeText,
		Required:     false,
		DisplayOrder: 0,
	}
	err = repo.Create(context.Background(), question)
	if err != nil {
		t.Fatalf("Failed to create test question: %v", err)
	}

	err = repo.Delete(context.Background(), question.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.GetByID(context.Background(), question.ID)
	if err == nil {
		t.Error("Expected error when getting deleted question")
	}
}

func TestQuestionRepository_Reorder(t *testing.T) {
	database := setupQuestionTestDB(t)
	defer database.Close()

	repo := NewQuestionRepository(database)
	eventRepo := NewEventRepository(database)

	event := &models.Event{
		Title:       "Test Event",
		StartTime:   time.Now().Add(24 * time.Hour),
		Timezone:    "America/Los_Angeles",
		Status:      models.EventStatusDraft,
		CreatedBy:   1,
		MaxPlusOnes: 2,
	}
	err := eventRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	q1 := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Question 1",
		QuestionType: models.QuestionTypeText,
		DisplayOrder: 0,
	}
	repo.Create(context.Background(), q1)

	q2 := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Question 2",
		QuestionType: models.QuestionTypeText,
		DisplayOrder: 1,
	}
	repo.Create(context.Background(), q2)

	q3 := &models.PreferenceQuestion{
		EventID:      event.ID,
		QuestionText: "Question 3",
		QuestionType: models.QuestionTypeText,
		DisplayOrder: 2,
	}
	repo.Create(context.Background(), q3)

	newOrder := []int64{q3.ID, q1.ID, q2.ID}

	err = repo.Reorder(context.Background(), event.ID, newOrder)
	if err != nil {
		t.Fatalf("Reorder() error = %v", err)
	}

	questions, err := repo.GetByEventID(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("GetByEventID() error = %v", err)
	}

	if len(questions) != 3 {
		t.Fatalf("Expected 3 questions, got %d", len(questions))
	}

	if questions[0].ID != q3.ID {
		t.Errorf("First question ID = %v, want %v", questions[0].ID, q3.ID)
	}
	if questions[1].ID != q1.ID {
		t.Errorf("Second question ID = %v, want %v", questions[1].ID, q1.ID)
	}
	if questions[2].ID != q2.ID {
		t.Errorf("Third question ID = %v, want %v", questions[2].ID, q2.ID)
	}
}

