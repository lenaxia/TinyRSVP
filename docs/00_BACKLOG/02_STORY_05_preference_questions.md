# User Story: Preference Questions CRUD

**Epic:** [02_EPIC_events.md](02_EPIC_events.md)
**Priority:** Medium
**Status:** ✅ Complete
**Estimated Effort:** 6 hours
**Actual Effort:** 6 hours
**Completed:** 2026-01-08

---

## User Story

As an **event manager**, I want **to add preference questions to events** so that **I can collect additional information from guests when they RSVP**.

---

## Acceptance Criteria

- [x] PreferenceQuestion model defined
- [x] Question types supported: text, single_choice, multiple_choice
- [x] Add question to event
- [x] Update question
- [x] Delete question
- [x] Reorder questions
- [x] Get questions for event
- [x] Validate question structure
- [x] All tests pass with timeout

---

## Technical Details

### PreferenceQuestion Model

```go
type QuestionType string

const (
    QuestionTypeText           QuestionType = "text"
    QuestionTypeSingleChoice   QuestionType = "single_choice"
    QuestionTypeMultipleChoice QuestionType = "multiple_choice"
)

type PreferenceQuestion struct {
    ID          int64        `db:"id" json:"id"`
    EventID     int64        `db:"event_id" json:"event_id"`
    QuestionText string      `db:"question_text" json:"question_text"`
    QuestionType QuestionType `db:"question_type" json:"question_type"`
    Required    bool         `db:"required" json:"required"`
    DisplayOrder int         `db:"display_order" json:"display_order"`
    Options     *string      `db:"options" json:"options,omitempty"`
    CreatedAt   time.Time    `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
}
```

### Repository Interface

```go
type PreferenceQuestionRepository interface {
    Create(ctx context.Context, question *PreferenceQuestion) error
    GetByID(ctx context.Context, id int64) (*PreferenceQuestion, error)
    GetByEventID(ctx context.Context, eventID int64) ([]*PreferenceQuestion, error)
    Update(ctx context.Context, question *PreferenceQuestion) error
    Delete(ctx context.Context, id int64) error
    Reorder(ctx context.Context, eventID int64, questionIDs []int64) error
}
```

### Service Interface

```go
type PreferenceQuestionService interface {
    AddQuestion(ctx context.Context, question *PreferenceQuestion) error
    UpdateQuestion(ctx context.Context, question *PreferenceQuestion) error
    DeleteQuestion(ctx context.Context, id int64) error
    GetQuestions(ctx context.Context, eventID int64) ([]*PreferenceQuestion, error)
    ReorderQuestions(ctx context.Context, eventID int64, questionIDs []int64) error
}
```

### Validation Rules

**Question Text:**
- Required
- 5-500 characters
- No leading/trailing whitespace

**Question Type:**
- Must be one of: text, single_choice, multiple_choice
- Required

**Options:**
- Required for single_choice and multiple_choice
- Must be valid JSON array
- 2-10 options
- Each option 1-200 characters

**Display Order:**
- Auto-assigned on creation
- Can be reordered

---

## Tasks

### Phase 1: Model and Validation (TDD)
- [x] Write test for PreferenceQuestion struct
- [x] Write test for question type validation
- [x] Write test for text question validation
- [x] Write test for single choice validation
- [x] Write test for multiple choice validation
- [x] Write test for options JSON parsing
- [x] Implement question validator
- [x] Run tests (should pass)

### Phase 2: Repository (TDD)
- [x] Write test for creating question
- [x] Write test for getting question by ID
- [x] Write test for getting questions by event
- [x] Write test for updating question
- [x] Write test for deleting question
- [x] Write test for reordering questions
- [x] Write test for foreign key constraints
- [x] Implement repository methods
- [x] Run tests (should pass)

### Phase 3: Service Layer (TDD)
- [x] Write test for adding question to event
- [x] Write test for adding question to non-existent event
- [x] Write test for adding question without permission
- [x] Write test for updating question
- [x] Write test for deleting question
- [x] Write test for reordering questions
- [x] Write test for getting questions
- [x] Implement service methods
- [x] Run tests (should pass)

### Phase 4: HTTP Handlers (TDD)
- [x] Write test for POST /api/events/:id/questions
- [x] Write test for GET /api/events/:id/questions
- [x] Write test for PUT /api/events/:id/questions/:qid
- [x] Write test for DELETE /api/events/:id/questions/:qid
- [x] Write test for POST /api/events/:id/questions/reorder
- [x] Implement handlers
- [x] Run tests (should pass)

### Phase 5: Integration Tests
- [x] Write integration test for full question lifecycle
- [x] Write integration test for question ordering
- [x] Write integration test for published event restrictions
- [x] Run integration tests

---

## Testing Requirements

### Unit Tests

```go
func TestPreferenceQuestionValidator_ValidateCreate(t *testing.T) {
    tests := []struct {
        name    string
        question *models.PreferenceQuestion
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid text question",
            question: &models.PreferenceQuestion{
                QuestionText: "What is your dietary preference?",
                QuestionType: models.QuestionTypeText,
                Required:     true,
            },
            wantErr: false,
        },
        {
            name: "valid single choice",
            question: &models.PreferenceQuestion{
                QuestionText: "Will you attend?",
                QuestionType: models.QuestionTypeSingleChoice,
                Required:     true,
                Options:      stringPtr(`["Yes", "No", "Maybe"]`),
            },
            wantErr: false,
        },
        {
            name: "valid multiple choice",
            question: &models.PreferenceQuestion{
                QuestionText: "Select dietary restrictions",
                QuestionType: models.QuestionTypeMultipleChoice,
                Required:     false,
                Options:      stringPtr(`["Vegetarian", "Vegan", "Gluten-free", "None"]`),
            },
            wantErr: false,
        },
        {
            name: "question text too short",
            question: &models.PreferenceQuestion{
                QuestionText: "Why?",
                QuestionType: models.QuestionTypeText,
            },
            wantErr: true,
            errMsg:  "question text must be between 5 and 500 characters",
        },
        {
            name: "choice question without options",
            question: &models.PreferenceQuestion{
                QuestionText: "Choose one",
                QuestionType: models.QuestionTypeSingleChoice,
            },
            wantErr: true,
            errMsg:  "options required for choice questions",
        },
        {
            name: "too few options",
            question: &models.PreferenceQuestion{
                QuestionText: "Choose one",
                QuestionType: models.QuestionTypeSingleChoice,
                Options:      stringPtr(`["Only one"]`),
            },
            wantErr: true,
            errMsg:  "must have 2-10 options",
        },
        {
            name: "invalid JSON options",
            question: &models.PreferenceQuestion{
                QuestionText: "Choose one",
                QuestionType: models.QuestionTypeSingleChoice,
                Options:      stringPtr(`invalid json`),
            },
            wantErr: true,
            errMsg:  "invalid options JSON",
        },
    }
    
    validator := NewPreferenceQuestionValidator()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateCreate(context.Background(), tt.question)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
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

func TestPreferenceQuestionService_AddQuestion(t *testing.T) {
    tests := []struct {
        name      string
        user      *models.User
        eventID   int64
        question  *models.PreferenceQuestion
        setupMock func(*repositories.MockEventRepository, *repositories.MockPreferenceQuestionRepository)
        wantErr   bool
        errMsg    string
    }{
        {
            name: "event owner adds question",
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            eventID: 1,
            question: &models.PreferenceQuestion{
                QuestionText: "What is your dietary preference?",
                QuestionType: models.QuestionTypeText,
                Required:     true,
            },
            setupMock: func(er *repositories.MockEventRepository, qr *repositories.MockPreferenceQuestionRepository) {
                er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
                    return &models.Event{
                        ID:        id,
                        CreatedBy: 1,
                        Status:    models.EventStatusDraft,
                    }, nil
                }
                qr.CreateFunc = func(ctx context.Context, q *models.PreferenceQuestion) error {
                    q.ID = 1
                    return nil
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
            eventID: 1,
            question: &models.PreferenceQuestion{
                QuestionText: "Question",
                QuestionType: models.QuestionTypeText,
            },
            setupMock: func(er *repositories.MockEventRepository, qr *repositories.MockPreferenceQuestionRepository) {
                er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
                    return &models.Event{
                        ID:        id,
                        CreatedBy: 1,
                        Status:    models.EventStatusPublished,
                    }, nil
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
            eventID: 1,
            question: &models.PreferenceQuestion{
                QuestionText: "Question",
                QuestionType: models.QuestionTypeText,
            },
            setupMock: func(er *repositories.MockEventRepository, qr *repositories.MockPreferenceQuestionRepository) {
                er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
                    return &models.Event{
                        ID:        id,
                        CreatedBy: 1,
                    }, nil
                }
            },
            wantErr: true,
            errMsg:  "permission denied",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockEventRepo := &repositories.MockEventRepository{}
            mockQuestionRepo := &repositories.MockPreferenceQuestionRepository{}
            
            if tt.setupMock != nil {
                tt.setupMock(mockEventRepo, mockQuestionRepo)
            }
            
            service := NewPreferenceQuestionService(mockEventRepo, mockQuestionRepo, mockValidator, mockAuthz)
            
            ctx := auth.WithUser(context.Background(), tt.user)
            tt.question.EventID = tt.eventID
            
            err := service.AddQuestion(ctx, tt.question)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("AddQuestion() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Dependencies

**Depends on:** 
- 02_STORY_03_event_service.md - Event service
- 02_STORY_04_event_handlers.md - Event handlers

**Blocks:** 
- RSVP response collection
- Guest preference submission

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout
- [x] Test coverage >= 85%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] Question validation working
- [x] Documentation complete
- [x] Changes committed to git

---

## Implementation Notes

### Options Storage

Options are stored as JSON array in database:

```json
["Option 1", "Option 2", "Option 3"]
```

Parse and validate on read/write.

### Display Order

Questions are ordered by `display_order` field. When reordering:
1. Validate all question IDs belong to event
2. Update display_order for each question
3. Use transaction to ensure atomicity

### Cascade Delete

When event is deleted, all associated questions are deleted via foreign key constraint.

---

## References

- **HLD:** Section 8 (Preference Questions)
- **LLD:** [lld/02_EVENT_LLD.md](../lld/02_EVENT_LLD.md)
- **Epic:** [02_EPIC_events.md](02_EPIC_events.md)
