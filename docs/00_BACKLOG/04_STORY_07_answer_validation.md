# User Story: Validate Preference Question Answers

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 6 hours

---

## User Story

As a **developer**, I want **comprehensive answer validation logic** so that **only valid answers are accepted for preference questions**.

---

## Acceptance Criteria

- [ ] Text answers validated (max 500 chars, trimmed)
- [ ] Select answers validated against question options
- [ ] Boolean answers validated (true/false only)
- [ ] Required questions must have answers
- [ ] Optional questions can be empty
- [ ] Answer type must match question type
- [ ] One answer per question enforced
- [ ] Clear error messages for validation failures
- [ ] All validation tests pass with timeout

---

## Technical Details

### Validator Interface

```go
package rsvp

type AnswerValidator interface {
    ValidateAnswers(ctx context.Context, answers []AnswerRequest, eventID int64) error
    ValidateAnswer(answer *AnswerRequest, question *models.PreferenceQuestion) error
}
```

### Validation Rules

**Text Answers:**
- Max 500 characters
- Trimmed whitespace
- Sanitized for XSS
- Required if question.required = true

**Select Answers:**
- Must match one of question.options
- Case-sensitive
- Required if question.required = true

**Boolean Answers:**
- Must be true or false
- Required if question.required = true

**General:**
- Only one answer type field populated
- Question must exist and belong to event
- One answer per question maximum

---

## Tasks

### Phase 1: Validator Implementation (TDD)
- [ ] Write test for valid text answer
- [ ] Write test for text too long
- [ ] Write test for valid select answer
- [ ] Write test for invalid select option
- [ ] Write test for valid boolean answer
- [ ] Write test for invalid boolean value
- [ ] Write test for missing required answer
- [ ] Write test for multiple answer types
- [ ] Write test for duplicate question answers
- [ ] Implement ValidateAnswers method
- [ ] Implement ValidateAnswer method
- [ ] Run tests (should pass)

### Phase 2: Integration
- [ ] Wire validator into RSVP service
- [ ] Test validation in submission flow
- [ ] Test validation in update flow
- [ ] Verify error messages

---

## Validation Logic

```go
func (v *validator) ValidateAnswers(ctx context.Context, answers []AnswerRequest, eventID int64) error {
    // Get all questions for event
    questions, err := v.questionRepo.GetByEventID(ctx, eventID)
    if err != nil {
        return err
    }
    
    questionMap := make(map[int64]*models.PreferenceQuestion)
    for _, q := range questions {
        questionMap[q.ID] = q
    }
    
    // Check required questions
    answeredQuestions := make(map[int64]bool)
    for _, ans := range answers {
        if answeredQuestions[ans.QuestionID] {
            return &models.ValidationError{
                Field:   "answers",
                Message: "Duplicate answer for question",
            }
        }
        answeredQuestions[ans.QuestionID] = true
        
        question, exists := questionMap[ans.QuestionID]
        if !exists {
            return &models.ValidationError{
                Field:   "answers",
                Message: "Invalid question ID",
            }
        }
        
        if err := v.ValidateAnswer(&ans, question); err != nil {
            return err
        }
    }
    
    // Check all required questions answered
    for _, q := range questions {
        if q.Required && !answeredQuestions[q.ID] {
            return &models.ValidationError{
                Field:   "answers",
                Message: fmt.Sprintf("Required question not answered: %s", q.QuestionText),
            }
        }
    }
    
    return nil
}

func (v *validator) ValidateAnswer(answer *AnswerRequest, question *models.PreferenceQuestion) error {
    // Check only one answer type
    populated := 0
    if answer.AnswerText != nil {
        populated++
    }
    if answer.AnswerOption != nil {
        populated++
    }
    if answer.AnswerBoolean != nil {
        populated++
    }
    
    if populated != 1 {
        return &models.ValidationError{
            Field:   "answers",
            Message: "Exactly one answer field must be provided",
        }
    }
    
    // Validate by type
    switch question.Type {
    case models.QuestionTypeText:
        if answer.AnswerText == nil {
            return &models.ValidationError{
                Field:   "answers",
                Message: "Text answer required for text question",
            }
        }
        if len(*answer.AnswerText) > 500 {
            return &models.ValidationError{
                Field:   "answers",
                Message: "Answer text cannot exceed 500 characters",
            }
        }
        
    case models.QuestionTypeSelect:
        if answer.AnswerOption == nil {
            return &models.ValidationError{
                Field:   "answers",
                Message: "Option answer required for select question",
            }
        }
        valid := false
        for _, opt := range question.Options {
            if opt == *answer.AnswerOption {
                valid = true
                break
            }
        }
        if !valid {
            return &models.ValidationError{
                Field:   "answers",
                Message: "Invalid option selected",
            }
        }
        
    case models.QuestionTypeBoolean:
        if answer.AnswerBoolean == nil {
            return &models.ValidationError{
                Field:   "answers",
                Message: "Boolean answer required for yes/no question",
            }
        }
    }
    
    return nil
}
```

---

## Testing Requirements

### Unit Tests

```go
func TestAnswerValidator_ValidateAnswer(t *testing.T) {
    tests := []struct {
        name     string
        answer   *AnswerRequest
        question *models.PreferenceQuestion
        wantErr  bool
        errMsg   string
    }{
        {
            name: "valid text answer",
            answer: &AnswerRequest{
                AnswerText: strPtr("Vegetarian"),
            },
            question: &models.PreferenceQuestion{
                Type: models.QuestionTypeText,
            },
            wantErr: false,
        },
        {
            name: "text too long",
            answer: &AnswerRequest{
                AnswerText: strPtr(strings.Repeat("a", 501)),
            },
            question: &models.PreferenceQuestion{
                Type: models.QuestionTypeText,
            },
            wantErr: true,
            errMsg:  "cannot exceed 500 characters",
        },
        {
            name: "valid select answer",
            answer: &AnswerRequest{
                AnswerOption: strPtr("red"),
            },
            question: &models.PreferenceQuestion{
                Type:    models.QuestionTypeSelect,
                Options: []string{"red", "blue", "green"},
            },
            wantErr: false,
        },
        {
            name: "invalid select option",
            answer: &AnswerRequest{
                AnswerOption: strPtr("purple"),
            },
            question: &models.PreferenceQuestion{
                Type:    models.QuestionTypeSelect,
                Options: []string{"red", "blue", "green"},
            },
            wantErr: true,
            errMsg:  "Invalid option",
        },
        {
            name: "valid boolean answer",
            answer: &AnswerRequest{
                AnswerBoolean: boolPtr(true),
            },
            question: &models.PreferenceQuestion{
                Type: models.QuestionTypeBoolean,
            },
            wantErr: false,
        },
        {
            name: "multiple answer types",
            answer: &AnswerRequest{
                AnswerText:   strPtr("text"),
                AnswerOption: strPtr("option"),
            },
            question: &models.PreferenceQuestion{
                Type: models.QuestionTypeText,
            },
            wantErr: true,
            errMsg:  "Exactly one answer field",
        },
    }
    
    validator := NewValidator(nil)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateAnswer(tt.answer, tt.question)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateAnswer() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && err != nil {
                if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.errMsg)
                }
            }
        })
    }
}
```

---

## Dependencies

**Depends on:**
- Story 00: RSVP Model
- Story 02_STORY_05: Preference Questions

**Blocks:**
- Story 02: RSVP Submission
- Story 06: Answer Submission
- Story 08: RSVP Updates

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Validator implemented and tested
- [ ] Unit tests passing (100% coverage)
- [ ] Integration tests passing
- [ ] Error messages clear
- [ ] Edge cases handled
- [ ] Documentation updated
- [ ] Code reviewed

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
