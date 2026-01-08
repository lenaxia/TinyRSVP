# User Story: Submit Preference Question Answers

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 4 hours
**Completed:** 2026-01-08

---

## User Story

As a **guest**, I want **my preference question answers to be saved with my RSVP** so that **the host receives my preferences**.

---

## Acceptance Criteria

- [x] Answers submitted with RSVP in single request
- [x] Answers saved atomically with RSVP (transaction)
- [x] Answer type matches question type
- [x] Required questions must have answers
- [x] Optional questions can be skipped
- [x] One answer per question maximum
- [x] Answers validated before save
- [x] Clear error messages for validation failures
- [x] All tests pass with timeout

---

## Technical Details

### Request Structure

```json
{
    "response": "yes",
    "plus_ones": 2,
    "answers": [
        {
            "question_id": 1,
            "answer_text": "Vegetarian"
        },
        {
            "question_id": 2,
            "answer_option": "red"
        },
        {
            "question_id": 3,
            "answer_option": "Yes"
        }
    ]
}
```

### Service Logic

```go
func (s *service) SubmitRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) error {
    // ... validate token, event, deadline ...
    
    // Validate answers
    if err := s.validateAnswers(ctx, req.Answers, event.ID); err != nil {
        return err
    }
    
    // Save in transaction
    return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
        // Create RSVP
        rsvp := &models.RSVP{...}
        if err := s.rsvpRepo.Create(ctx, rsvp); err != nil {
            return err
        }
        
        // Create answers
        for _, ansReq := range req.Answers {
            answer := &models.RSVPAnswer{
                RSVPID:       rsvp.ID,
                QuestionID:   ansReq.QuestionID,
                AnswerText:   ansReq.AnswerText,
                AnswerOption: ansReq.AnswerOption,
            }
            if err := s.answerRepo.Create(ctx, answer); err != nil {
                return err
            }
        }
        
        return nil
    })
}
```

---

## Tasks

- [x] Implement answer validation
- [x] Implement atomic save logic
- [x] Write tests for valid answers
- [x] Write tests for missing required answers
- [x] Write tests for invalid answer types
- [x] Write tests for duplicate answers
- [x] Integration test full flow

---

## Validation Rules

- Required questions must have answers
- Answer type must match question type (text, single_choice, multiple_choice)
- Text: max 500 characters
- Single choice/Multiple choice: must match one of the question's options
- One answer per question

---

## Dependencies

**Depends on:**
- Story 00: RSVP Model
- Story 02: RSVP Submission
- Story 05: Question Display
- Story 07: Answer Validation

**Blocks:**
- Story 08: RSVP Updates

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Service logic implemented
- [x] Tests passing (>90% coverage)
- [x] Transaction handling working
- [x] Error handling complete
- [x] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
