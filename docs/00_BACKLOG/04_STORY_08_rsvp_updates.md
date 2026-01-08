# User Story: Update Existing RSVP

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **guest**, I want **to update my RSVP response** so that **I can change my mind before the deadline**.

---

## Acceptance Criteria

- [ ] Guest can update RSVP via same token link
- [ ] Existing RSVP loaded and displayed
- [ ] Guest can change response (yes/no/maybe)
- [ ] Guest can change plus ones
- [ ] Guest can update preference answers
- [ ] Deadline enforced (no updates after deadline)
- [ ] Updates saved atomically (transaction)
- [ ] Confirmation shown after update
- [ ] All tests pass with timeout

---

## Technical Details

### Endpoint
```
PUT /api/rsvp/:token
Content-Type: application/json

{
    "response": "maybe",
    "plus_ones": 1,
    "answers": [...]
}
```

### Service Method

```go
func (s *service) UpdateRSVP(ctx context.Context, token string, req *UpdateRSVPRequest) error {
    // 1. Validate token
    invite, err := s.inviteService.ValidateToken(ctx, token)
    if err != nil {
        return err
    }
    
    // 2. Get existing RSVP
    existing, err := s.rsvpRepo.GetByInviteID(ctx, invite.ID)
    if err != nil {
        return ErrRSVPNotFound
    }
    
    // 3. Check deadline
    event, err := s.eventRepo.GetByID(ctx, invite.EventID)
    if err != nil {
        return err
    }
    
    if event.RSVPDeadline != nil && time.Now().After(*event.RSVPDeadline) {
        return ErrDeadlinePassed
    }
    
    // 4. Validate request
    if err := s.validator.ValidateSubmission(req, invite, event); err != nil {
        return err
    }
    
    // 5. Update in transaction
    return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
        // Update RSVP
        existing.Response = req.Response
        existing.PlusOnes = req.PlusOnes
        if err := s.rsvpRepo.Update(ctx, existing); err != nil {
            return err
        }
        
        // Delete old answers
        if err := s.answerRepo.DeleteByRSVPID(ctx, existing.ID); err != nil {
            return err
        }
        
        // Create new answers
        for _, ansReq := range req.Answers {
            answer := &models.RSVPAnswer{
                RSVPID:        existing.ID,
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

- [ ] Implement UpdateRSVP service method
- [ ] Create PUT handler endpoint
- [ ] Load existing RSVP in page handler
- [ ] Pre-populate form with existing values
- [ ] Write tests for successful update
- [ ] Write tests for deadline enforcement
- [ ] Write tests for validation
- [ ] Integration test full update flow

---

## UI Changes

### RSVP Page with Existing Response

```html
{{if .ExistingRSVP}}
<div class="existing-rsvp-notice">
    <p>You previously responded: <strong>{{.ExistingRSVP.Response}}</strong></p>
    <p>You can update your response until {{.Event.RSVPDeadline}}</p>
</div>
{{end}}

<form method="POST" action="/api/rsvp/{{.Token}}">
    {{if .ExistingRSVP}}
    <input type="hidden" name="_method" value="PUT">
    {{end}}
    
    <!-- Form fields pre-populated with existing values -->
</form>
```

---

## Dependencies

**Depends on:**
- Story 00: RSVP Model
- Story 02: RSVP Submission
- Story 09: Deadline Enforcement

**Blocks:**
- None

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Service method implemented
- [ ] Handler implemented
- [ ] UI updated
- [ ] Tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
