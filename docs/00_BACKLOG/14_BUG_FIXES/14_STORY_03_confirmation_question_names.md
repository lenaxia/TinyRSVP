# STORY: Look Up Question Text in Confirmation Emails

**Epic:** 14 - Bug Fixes & Code Gaps  
**Story ID:** 14_STORY_03  
**Priority:** High  
**Estimated Effort:** 2 hours  
**Severity:** High — confirmation emails show "Question 1", "Question 2" instead of the actual question text for any event with preference questions

---

## Problem

`internal/email/confirmation_service.go:157`:

```go
Question: fmt.Sprintf("Question %d", answer.QuestionID),
```

When building the confirmation email template data, the question text is never looked up. Only `answer.QuestionID` (an integer) is available — the `RSVPAnswer` model does not embed the question text. The `PreferenceQuestion` repository is never consulted.

The test `TestSendConfirmationEmail_WithAnswers` explicitly asserts this broken behavior:

```go
if templateData.Answers[0].Question != "Question 1" {
    t.Errorf(...)
}
```

The test passes because it was written to match the bug.

---

## Acceptance Criteria

- [ ] Confirmation emails display actual question text (e.g., "Dietary requirements?") not "Question N"
- [ ] `ConfirmationService` accepts a `QuestionRepository` (or equivalent interface) and calls `GetByID` for each answer's `QuestionID`
- [ ] If a question cannot be found (deleted since RSVP), fall back gracefully (e.g., `"Question (deleted)"` or skip the answer)
- [ ] `TestSendConfirmationEmail_WithAnswers` is updated to inject mock questions and assert the real question text appears
- [ ] All confirmation service tests pass
- [ ] All 32 non-browser packages pass
- [ ] Update `docs/00_BACKLOG/05_EMAIL/README.md`: remove BUG-2
- [ ] Update `docs/00_BACKLOG/14_BUG_FIXES/README.md`: mark this story complete

---

## Technical Approach

### 1. Define a narrow interface

```go
// In internal/email/confirmation_service.go or a new file
type QuestionLookup interface {
    GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error)
}
```

### 2. Update `ConfirmationService`

```go
type ConfirmationService struct {
    renderer        TemplateRenderer
    emailRepo       repositories.EmailQueueRepository
    icsGenerator    ICSGenerator
    baseURL         string
    questionLookup  QuestionLookup  // NEW
}
```

### 3. Use it in `SendConfirmationEmail`

```go
for _, answer := range answers {
    questionText := fmt.Sprintf("Question %d", answer.QuestionID) // fallback
    if s.questionLookup != nil {
        if q, err := s.questionLookup.GetByID(ctx, answer.QuestionID); err == nil {
            questionText = q.Text
        }
    }
    // ...
}
```

### 4. Update constructor and `cmd/server/main.go`

Pass `questionRepo` into `NewConfirmationService(...)`.

### 5. Update test

```go
// Inject mock question lookup that returns real question text
mockQL := &mockQuestionLookup{
    questions: map[int64]string{1: "Dietary requirements?", 2: "T-shirt size?"},
}
service := NewConfirmationService(renderer, repo, generator, "https://...", mockQL)
// ...
if templateData.Answers[0].Question != "Dietary requirements?" {
    t.Errorf(...)
}
```

---

## Files to Change

- `internal/email/confirmation_service.go` — add `QuestionLookup`, use it
- `internal/email/confirmation_service_test.go` — update `TestSendConfirmationEmail_WithAnswers`
- `cmd/server/main.go` — pass `questionRepo` to `NewConfirmationService`

---

## Testing

```bash
go test -timeout 30s ./internal/email/...
go test -timeout 30s ./...
```

---

## Status

- **Status:** ✅ Complete — 2026-04-06

## Implementation Notes

- `confirmationService` gains `questionRepo repositories.QuestionRepository` (optional field).
- `NewConfirmationServiceWithQuestions` constructor wires it; `NewConfirmationService` leaves it nil (backward-compatible).
- `SendConfirmationEmail` calls `questionRepo.GetByEventID(ctx, event.ID)` — one query — and builds a `map[int64]string{questionID -> questionText}`.
- `prepareTemplateData` iterates answers and only includes those whose question ID appears in the map. Answers for deleted/missing questions are silently omitted — no "Question N" fallback.
- If no question repo is configured (`NewConfirmationService`) or `GetByEventID` errors, the map is empty and the Answers section is omitted entirely from the email. Email still sends.
- `{{if .Answers}}` guard in both `rsvp_confirmation.html` and `rsvp_confirmation.txt` handles empty/nil slice correctly — section is simply not rendered.
- `cmd/server/main.go` uses `NewConfirmationServiceWithQuestions(..., questionRepo)`.
- `TestSendConfirmationEmail_WithAnswers` updated: mock repo returns questions by `EventID`, asserts `"Dietary requirements?"` and `"T-shirt size?"`.
- `_FallsBackWhenQuestionNotFound`: Q1 (EventID 1) known, Q99 not in repo → 1 answer in output.
- `_NoQuestionRepo_OmitsAnswers`: no repo → `Answers` is nil/empty → section omitted.
- 32/32 packages pass.
