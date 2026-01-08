# User Story: Submit Preference Question Answers

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Medium
**Status:** Not Started
**Estimated Effort:** 4 hours

---

## User Story

As a **guest**, I want **my preference question answers to be saved with my RSVP** so that **the host receives my preferences**.

---

## Acceptance Criteria

- [ ] Answers submitted with RSVP in single request
- [ ] Answers saved atomically with RSVP (transaction)
- [ ] Answer type matches question type
- [ ] Required questions must have answers
- [ ] Optional questions can be skipped
- [ ] One answer per question maximum
- [ ] Answers validated before save
- [ ] Clear error messages for validation failures
- [ ] All tests pass with timeout

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
            "answer_boolean": true
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
                RSVPID:        rsvp.ID,
                QuestionID:    ansReq.QuestionID,
                AnswerText:    ansReq.AnswerText,
                AnswerOption:  ansReq.AnswerOption,
                AnswerBoolean: ansReq.AnswerBoolean,
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

- [ ] Implement answer validation
- [ ] Implement atomic save logic
- [ ] Write tests for valid answers
- [ ] Write tests for missing required answers
- [ ] Write tests for invalid answer types
- [ ] Write tests for duplicate answers
- [ ] Integration test full flow

---

## Validation Rules

- Required questions must have answers
- Answer type must match question type
- Text: max 500 characters
- Option: must match question options
- Boolean: must be true or false
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

- [ ] All acceptance criteria met
- [ ] Service logic implemented
- [ ] Tests passing (>90% coverage)
- [ ] Transaction handling working
- [ ] Error handling complete
- [ ] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
