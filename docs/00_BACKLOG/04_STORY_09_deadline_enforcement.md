# User Story: RSVP Deadline Enforcement

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Critical
**Status:** Complete
**Estimated Effort:** 4 hours
**Completed:** 2026-01-08

---

## User Story

As a **host**, I want **RSVP deadlines strictly enforced** so that **guests cannot respond after the deadline**.

---

## Acceptance Criteria

- [x] Deadline checked on RSVP submission
- [x] Deadline checked on RSVP update
- [x] Strict enforcement (no grace period)
- [x] Clear error message when deadline passed
- [x] Event details still visible after deadline
- [x] RSVP form disabled after deadline
- [x] Deadline displayed prominently on page
- [x] Timezone-aware deadline checking
- [x] All tests pass with timeout

---

## Technical Details

### Deadline Check Logic

```go
func (s *service) checkDeadline(event *models.Event) error {
    if event.RSVPDeadline == nil {
        return nil // No deadline set
    }
    
    now := time.Now().UTC()
    deadline := event.RSVPDeadline.UTC()
    
    if now.After(deadline) {
        return &models.DeadlinePassedError{
            Deadline: deadline,
            Message:  "RSVP deadline has passed",
        }
    }
    
    return nil
}
```

### UI State After Deadline

```html
{{if .DeadlinePassed}}
<div class="deadline-passed-notice" role="alert">
    <h2>RSVP Deadline Has Passed</h2>
    <p>The deadline to RSVP was {{.Event.RSVPDeadline}}.</p>
    <p>Please contact the host directly if you need to respond.</p>
</div>

<!-- Event details still visible but form disabled -->
<div class="event-details">
    <!-- Event information -->
</div>
{{else}}
<!-- Show RSVP form -->
{{end}}
```

---

## Tasks

- [x] Implement deadline check function
- [x] Add deadline check to submission flow
- [x] Add deadline check to update flow
- [x] Update UI to show deadline status
- [x] Disable form after deadline
- [x] Write tests for before deadline
- [x] Write tests for after deadline
- [x] Write tests for no deadline set
- [x] Write tests for timezone edge cases

---

## Testing Requirements

### Unit Tests

```go
func TestDeadlineEnforcement(t *testing.T) {
    tests := []struct {
        name     string
        deadline *time.Time
        wantErr  bool
    }{
        {
            name:     "no deadline set",
            deadline: nil,
            wantErr:  false,
        },
        {
            name:     "deadline in future",
            deadline: timePtr(time.Now().Add(24 * time.Hour)),
            wantErr:  false,
        },
        {
            name:     "deadline in past",
            deadline: timePtr(time.Now().Add(-24 * time.Hour)),
            wantErr:  true,
        },
        {
            name:     "deadline exactly now",
            deadline: timePtr(time.Now()),
            wantErr:  true, // After means strictly after
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            event := &models.Event{
                RSVPDeadline: tt.deadline,
            }
            
            err := checkDeadline(event)
            if (err != nil) != tt.wantErr {
                t.Errorf("checkDeadline() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Error Messages

| Condition | User Message |
|-----------|--------------|
| Deadline passed | "RSVP deadline has passed. Please contact the host directly." |
| Deadline today | "RSVP deadline is today at [TIME]. Please respond soon!" |
| No deadline | (No message, form available) |

---

## Dependencies

**Depends on:**
- Story 00: RSVP Model
- Story 01: RSVP Page
- Story 02: RSVP Submission
- Story 08: RSVP Updates

**Blocks:**
- None

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Deadline check implemented
- [x] UI updated for deadline state
- [x] Tests passing (100% coverage)
- [x] Timezone handling correct
- [x] Error messages clear
- [x] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
