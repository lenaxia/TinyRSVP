# Worklog: Preference Questions CRUD Implementation

**Date:** 2026-01-08  
**Story:** Epic 02 Story 05 - Preference Questions CRUD  
**Status:** ✅ Complete

---

## Summary

Implemented complete CRUD functionality for preference questions, allowing event managers to add custom questions to events that guests answer during RSVP. The implementation follows TDD principles with comprehensive test coverage across all layers.

---

## What Was Implemented

### 1. Model Layer (`internal/models/preference_question.go`)
- **PreferenceQuestion struct** with fields:
  - ID, EventID, QuestionText, QuestionType
  - Required, DisplayOrder, Options
  - CreatedAt, UpdatedAt timestamps
- **QuestionType enum** with three types:
  - `text` - Free-form text input
  - `single_choice` - Select one option
  - `multiple_choice` - Select multiple options
- **Helper methods**:
  - `ParseOptions()` - Parse JSON options array
  - `SetOptions()` - Set options from string array

### 2. Validation Layer (`internal/events/question_validator.go`)
- **QuestionValidator** with comprehensive validation:
  - Question text: 5-500 characters, no leading/trailing whitespace
  - Question type: Must be valid enum value
  - Options validation:
    - Required for choice questions
    - Not allowed for text questions
    - 2-10 options required
    - Each option 1-200 characters
    - No duplicate options
- **ValidateCreate** and **ValidateUpdate** methods

### 3. Repository Layer (`internal/db/repositories/question_repository.go`)
- **QuestionRepository interface** with methods:
  - `Create` - Insert new question
  - `GetByID` - Retrieve by ID
  - `GetByEventID` - Get all questions for event (ordered)
  - `Update` - Update existing question
  - `Delete` - Remove question
  - `Reorder` - Atomic reordering with transaction
- Proper error handling with NotFoundError
- Transaction-based reordering ensures atomicity

### 4. Service Layer (`internal/events/question_service.go`)
- **QuestionService interface** with business logic:
  - `AddQuestion` - Add question with authorization
  - `UpdateQuestion` - Update with permission checks
  - `DeleteQuestion` - Delete with authorization
  - `GetQuestions` - List questions with view permission
  - `ReorderQuestions` - Reorder with validation
- **Authorization enforcement**:
  - Only event owners can modify questions
  - Admins have full access
  - Questions can only be modified on draft events
- **State validation**:
  - Cannot modify questions on published/cancelled/archived events

### 5. HTTP Handlers (`internal/handlers/questions.go`)
- **REST API endpoints**:
  - `POST /api/events/:id/questions` - Create question
  - `GET /api/events/:id/questions` - List questions
  - `PUT /api/events/:id/questions/:qid` - Update question
  - `DELETE /api/events/:id/questions/:qid` - Delete question
  - `POST /api/events/:id/questions/reorder` - Reorder questions
- **Request/Response types**:
  - CreateQuestionRequest
  - UpdateQuestionRequest
  - ReorderQuestionsRequest
- **Error handling**:
  - 400 for validation errors
  - 403 for permission denied
  - 404 for not found
  - 500 for internal errors

### 6. Database Migration (`migrations/sqlite/000005_update_preference_questions.up.sql`)
- Updated `preference_questions` table schema:
  - Added `updated_at` column
  - Changed question_type CHECK constraint from ('text', 'select', 'boolean') to ('text', 'single_choice', 'multiple_choice')
  - Data migration for existing records
- Corresponding down migration for rollback

---

## Test Coverage

### Unit Tests
- ✅ Model tests (3 test functions, 14 test cases)
- ✅ Validator tests (2 test functions, 23 test cases)
- ✅ Repository tests (6 test functions)
- ✅ Service tests (4 test functions, 10 test cases)
- ✅ Handler tests (5 test functions, 15 test cases)

### Integration Tests
- ✅ Full lifecycle test (6 subtests)
- ✅ Published event restrictions (4 subtests)
- ✅ Validation errors (3 subtests)

**Total:** 65+ test cases, all passing

---

## Key Design Decisions

### 1. Question Types
Chose three question types based on common RSVP needs:
- **text**: Flexible for dietary restrictions, special requests, etc.
- **single_choice**: For yes/no, attendance confirmation, meal selection
- **multiple_choice**: For selecting multiple dietary restrictions, interests

### 2. Options Storage
Store options as JSON array in database for flexibility:
```json
["Option 1", "Option 2", "Option 3"]
```
- Easy to parse and validate
- Supports variable number of options
- No additional tables needed

### 3. Display Order
Use integer `display_order` field for question ordering:
- Simple and efficient
- Atomic reordering with transactions
- Ordered by display_order ASC, then ID ASC

### 4. State Restrictions
Only allow question modifications on draft events:
- Prevents confusion for guests who already RSVP'd
- Maintains data integrity
- Clear business rule enforcement

### 5. Authorization
Reuse existing event authorization:
- `CanEditEvent` for modifications
- `CanViewEvent` for reading
- Consistent with event management permissions

---

## Files Created

### Production Code
1. `internal/models/preference_question.go` - Model definition
2. `internal/models/preference_question_test.go` - Model tests
3. `internal/events/question_validator.go` - Validation logic
4. `internal/events/question_validator_test.go` - Validator tests
5. `internal/db/repositories/question_repository.go` - Data access
6. `internal/db/repositories/question_repository_test.go` - Repository tests
7. `internal/events/question_service.go` - Business logic
8. `internal/events/question_service_test.go` - Service tests
9. `internal/handlers/questions.go` - HTTP handlers
10. `internal/handlers/questions_test.go` - Handler tests
11. `internal/handlers/questions_integration_test.go` - Integration tests

### Database Migrations
12. `migrations/sqlite/000005_update_preference_questions.up.sql` - Schema update
13. `migrations/sqlite/000005_update_preference_questions.down.sql` - Rollback

**Total:** 13 files, ~2,800 lines of code

---

## API Endpoints

### Create Question
```http
POST /api/events/:id/questions
Content-Type: application/json

{
  "question_text": "What is your dietary preference?",
  "question_type": "text",
  "required": true
}
```

### List Questions
```http
GET /api/events/:id/questions
```

### Update Question
```http
PUT /api/events/:id/questions/:qid
Content-Type: application/json

{
  "question_text": "Updated question",
  "question_type": "text",
  "required": false
}
```

### Delete Question
```http
DELETE /api/events/:id/questions/:qid
```

### Reorder Questions
```http
POST /api/events/:id/questions/reorder
Content-Type: application/json

{
  "question_ids": [3, 1, 2]
}
```

---

## Testing Results

```bash
# All question-related tests
go test -timeout 30s ./internal/models ./internal/events ./internal/db/repositories ./internal/handlers -run "Question"

ok  	github.com/lenaxia/tinyrsvp/internal/models	0.006s
ok  	github.com/lenaxia/tinyrsvp/internal/events	0.007s
ok  	github.com/lenaxia/tinyrsvp/internal/db/repositories	0.049s
ok  	github.com/lenaxia/tinyrsvp/internal/handlers	0.044s
```

All tests passing ✅

---

## Next Steps

### Immediate
1. ✅ Update story status in backlog
2. ✅ Commit all changes
3. ✅ Create worklog entry

### Future Work
- Epic 02 Story 06: Auto-archiving (next story in epic)
- RSVP answer collection (Epic 04) will use these questions
- Frontend UI for question management (Epic 07)

---

## Dependencies Satisfied

**This story depended on:**
- ✅ 02_STORY_03_event_service.md - Event service (complete)
- ✅ 02_STORY_04_event_handlers.md - Event handlers (complete)

**This story unblocks:**
- Epic 04: RSVP response collection
- Epic 04: Guest preference submission

---

## Lessons Learned

### What Went Well
1. TDD approach caught issues early
2. Clear separation of concerns across layers
3. Reusing existing authorization patterns
4. Transaction-based reordering ensures data consistency

### Challenges
1. Database schema mismatch required migration
2. Mock setup required careful attention to return values
3. Integration test path resolution needed adjustment

### Improvements for Next Time
1. Verify database schema matches spec before starting
2. Create shared test helpers for common mock setups
3. Document migration strategy upfront

---

## Commits

1. `cd66b77` - feat: add PreferenceQuestion model and validator
2. `0b775c1` - feat: add PreferenceQuestion repository with full CRUD
3. `422fd9b` - feat: add PreferenceQuestion service layer
4. `a0346a9` - feat: add PreferenceQuestion HTTP handlers
5. `ea1b8d7` - feat: add PreferenceQuestion integration tests

---

**Status:** ✅ Complete  
**All Tests:** ✅ Passing  
**Ready for:** Next story in Epic 02
