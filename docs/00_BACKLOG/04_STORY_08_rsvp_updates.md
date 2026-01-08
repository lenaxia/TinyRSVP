# User Story: Update Existing RSVP

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Completed:** 2026-01-08

---

## User Story

As a **guest**, I want **to update my RSVP response** so that **I can change my mind before the deadline**.

---

## Acceptance Criteria

- [x] Guest can update RSVP via same token link
- [x] Existing RSVP loaded and displayed
- [x] Guest can change response (yes/no/maybe)
- [x] Guest can change plus ones
- [x] Guest can update preference answers
- [x] Deadline enforced (no updates after deadline)
- [x] Updates saved atomically (transaction)
- [x] Confirmation shown after update
- [x] All tests pass with timeout

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

- [x] Implement UpdateRSVP service method
- [x] Create PUT handler endpoint
- [x] Load existing RSVP in page handler
- [x] Pre-populate form with existing values
- [x] Write tests for successful update
- [x] Write tests for deadline enforcement
- [x] Write tests for validation
- [x] Integration test full update flow

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

- [x] All acceptance criteria met
- [x] Service method implemented
- [x] Handler implemented
- [x] UI updated (existing page handler loads RSVP)
- [x] Tests passing (>90% coverage)
- [x] Integration tests passing
- [x] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
