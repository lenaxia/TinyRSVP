package events

import (
	"context"
	"strings"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type mockQuestionRepository struct {
	createFunc       func(ctx context.Context, question *models.PreferenceQuestion) error
	getByIDFunc      func(ctx context.Context, id int64) (*models.PreferenceQuestion, error)
	getByEventIDFunc func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
	updateFunc       func(ctx context.Context, question *models.PreferenceQuestion) error
	deleteFunc       func(ctx context.Context, id int64) error
	reorderFunc      func(ctx context.Context, eventID int64, questionIDs []int64) error
}

func (m *mockQuestionRepository) Create(ctx context.Context, question *models.PreferenceQuestion) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, question)
	}
	question.ID = 1
	return nil
}

func (m *mockQuestionRepository) GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.PreferenceQuestion{ID: id}, nil
}

func (m *mockQuestionRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
	if m.getByEventIDFunc != nil {
		return m.getByEventIDFunc(ctx, eventID)
	}
	return []*models.PreferenceQuestion{}, nil
}

func (m *mockQuestionRepository) Update(ctx context.Context, question *models.PreferenceQuestion) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, question)
	}
	return nil
}

func (m *mockQuestionRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockQuestionRepository) Reorder(ctx context.Context, eventID int64, questionIDs []int64) error {
	if m.reorderFunc != nil {
		return m.reorderFunc(ctx, eventID, questionIDs)
	}
	return nil
}

type mockQuestionValidator struct {
	validateCreateFunc func(ctx context.Context, question *models.PreferenceQuestion) error
	validateUpdateFunc func(ctx context.Context, question *models.PreferenceQuestion) error
}

func (m *mockQuestionValidator) ValidateCreate(ctx context.Context, question *models.PreferenceQuestion) error {
	if m.validateCreateFunc != nil {
		return m.validateCreateFunc(ctx, question)
	}
	return nil
}

func (m *mockQuestionValidator) ValidateUpdate(ctx context.Context, question *models.PreferenceQuestion) error {
	if m.validateUpdateFunc != nil {
		return m.validateUpdateFunc(ctx, question)
	}
	return nil
}

func TestQuestionService_AddQuestion(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		question   *models.PreferenceQuestion
		setupMocks func(*mockEventRepository, *mockQuestionRepository, *mockQuestionValidator, *mockAuthorizationChecker)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "event owner adds text question",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "What is your dietary preference?",
				QuestionType: models.QuestionTypeText,
				Required:     true,
			},
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, v *mockQuestionValidator, a *mockAuthorizationChecker) {
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: false,
		},
		{
			name: "cannot add to published event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Question",
				QuestionType: models.QuestionTypeText,
			},
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, v *mockQuestionValidator, a *mockAuthorizationChecker) {
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errMsg:  "cannot modify questions on published event",
		},
		{
			name: "non-owner cannot add question",
			user: &models.User{
				ID:   2,
				Role: models.RoleEventManager,
			},
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Question",
				QuestionType: models.QuestionTypeText,
			},
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, v *mockQuestionValidator, a *mockAuthorizationChecker) {
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errMsg:  "permission denied",
		},
		{
			name: "validation error",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			question: &models.PreferenceQuestion{
				EventID:      1,
				QuestionText: "Bad",
				QuestionType: models.QuestionTypeText,
			},
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, v *mockQuestionValidator, a *mockAuthorizationChecker) {
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
				v.validateCreateFunc = func(ctx context.Context, question *models.PreferenceQuestion) error {
					return &models.ValidationError{
						Field:   "question_text",
						Message: "question text must be between 5 and 500 characters",
					}
				}
			},
			wantErr: true,
			errMsg:  "question text must be between 5 and 500 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := &mockEventRepository{}
			mockQuestionRepo := &mockQuestionRepository{}
			mockValidator := &mockQuestionValidator{}
			mockAuthz := &mockAuthorizationChecker{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockEventRepo, mockQuestionRepo, mockValidator, mockAuthz)
			}

			service := NewQuestionService(mockEventRepo, mockQuestionRepo, mockValidator, mockAuthz)

			ctx := auth.WithUser(context.Background(), tt.user)

			err := service.AddQuestion(ctx, tt.question)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddQuestion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestQuestionService_UpdateQuestion(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		question   *models.PreferenceQuestion
		setupMocks func(*mockEventRepository, *mockQuestionRepository, *mockQuestionValidator, *mockAuthorizationChecker)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "event owner updates question",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			question: &models.PreferenceQuestion{
				ID:           1,
				EventID:      1,
				QuestionText: "Updated question",
				QuestionType: models.QuestionTypeText,
			},
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, v *mockQuestionValidator, a *mockAuthorizationChecker) {
				qr.getByIDFunc = func(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
					return &models.PreferenceQuestion{
						ID:      id,
						EventID: 1,
					}, nil
				}
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: false,
		},
		{
			name: "cannot update on published event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			question: &models.PreferenceQuestion{
				ID:           1,
				EventID:      1,
				QuestionText: "Updated question",
				QuestionType: models.QuestionTypeText,
			},
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, v *mockQuestionValidator, a *mockAuthorizationChecker) {
				qr.getByIDFunc = func(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
					return &models.PreferenceQuestion{
						ID:      id,
						EventID: 1,
					}, nil
				}
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errMsg:  "cannot modify questions on published event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := &mockEventRepository{}
			mockQuestionRepo := &mockQuestionRepository{}
			mockValidator := &mockQuestionValidator{}
			mockAuthz := &mockAuthorizationChecker{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockEventRepo, mockQuestionRepo, mockValidator, mockAuthz)
			}

			service := NewQuestionService(mockEventRepo, mockQuestionRepo, mockValidator, mockAuthz)

			ctx := auth.WithUser(context.Background(), tt.user)

			err := service.UpdateQuestion(ctx, tt.question)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateQuestion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestQuestionService_DeleteQuestion(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		questionID int64
		setupMocks func(*mockEventRepository, *mockQuestionRepository, *mockAuthorizationChecker)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "event owner deletes question",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			questionID: 1,
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, a *mockAuthorizationChecker) {
				qr.getByIDFunc = func(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
					return &models.PreferenceQuestion{
						ID:      id,
						EventID: 1,
					}, nil
				}
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: false,
		},
		{
			name: "cannot delete from published event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			questionID: 1,
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, a *mockAuthorizationChecker) {
				qr.getByIDFunc = func(ctx context.Context, id int64) (*models.PreferenceQuestion, error) {
					return &models.PreferenceQuestion{
						ID:      id,
						EventID: 1,
					}, nil
				}
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errMsg:  "cannot modify questions on published event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := &mockEventRepository{}
			mockQuestionRepo := &mockQuestionRepository{}
			mockValidator := &mockQuestionValidator{}
			mockAuthz := &mockAuthorizationChecker{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockEventRepo, mockQuestionRepo, mockAuthz)
			}

			service := NewQuestionService(mockEventRepo, mockQuestionRepo, mockValidator, mockAuthz)

			ctx := auth.WithUser(context.Background(), tt.user)

			err := service.DeleteQuestion(ctx, tt.questionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteQuestion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestQuestionService_GetQuestions(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		eventID    int64
		setupMocks func(*mockEventRepository, *mockQuestionRepository, *mockAuthorizationChecker)
		wantErr    bool
	}{
		{
			name: "get questions for event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			eventID: 1,
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, a *mockAuthorizationChecker) {
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
					}, nil
				}
				a.CanViewEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return true
				}
				qr.getByEventIDFunc = func(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error) {
					return []*models.PreferenceQuestion{
						{ID: 1, EventID: eventID, QuestionText: "Question 1"},
						{ID: 2, EventID: eventID, QuestionText: "Question 2"},
					}, nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := &mockEventRepository{}
			mockQuestionRepo := &mockQuestionRepository{}
			mockValidator := &mockQuestionValidator{}
			mockAuthz := &mockAuthorizationChecker{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockEventRepo, mockQuestionRepo, mockAuthz)
			}

			service := NewQuestionService(mockEventRepo, mockQuestionRepo, mockValidator, mockAuthz)

			ctx := auth.WithUser(context.Background(), tt.user)

			questions, err := service.GetQuestions(ctx, tt.eventID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetQuestions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(questions) != 2 {
				t.Errorf("GetQuestions() returned %d questions, want 2", len(questions))
			}
		})
	}
}

func TestQuestionService_ReorderQuestions(t *testing.T) {
	tests := []struct {
		name        string
		user        *models.User
		eventID     int64
		questionIDs []int64
		setupMocks  func(*mockEventRepository, *mockQuestionRepository, *mockAuthorizationChecker)
		wantErr     bool
		errMsg      string
	}{
		{
			name: "event owner reorders questions",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			eventID:     1,
			questionIDs: []int64{3, 1, 2},
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, a *mockAuthorizationChecker) {
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusDraft,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: false,
		},
		{
			name: "cannot reorder on published event",
			user: &models.User{
				ID:   1,
				Role: models.RoleEventManager,
			},
			eventID:     1,
			questionIDs: []int64{3, 1, 2},
			setupMocks: func(er *mockEventRepository, qr *mockQuestionRepository, a *mockAuthorizationChecker) {
				er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
					return &models.Event{
						ID:        id,
						CreatedBy: 1,
						Status:    models.EventStatusPublished,
					}, nil
				}
				a.CanEditEventFunc = func(ctx context.Context, user *models.User, event *models.Event) bool {
					return user.ID == event.CreatedBy
				}
			},
			wantErr: true,
			errMsg:  "cannot modify questions on published event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventRepo := &mockEventRepository{}
			mockQuestionRepo := &mockQuestionRepository{}
			mockValidator := &mockQuestionValidator{}
			mockAuthz := &mockAuthorizationChecker{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockEventRepo, mockQuestionRepo, mockAuthz)
			}

			service := NewQuestionService(mockEventRepo, mockQuestionRepo, mockValidator, mockAuthz)

			ctx := auth.WithUser(context.Background(), tt.user)

			err := service.ReorderQuestions(ctx, tt.eventID, tt.questionIDs)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReorderQuestions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
