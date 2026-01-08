# User Story: RSVP Deadline Enforcement

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 4 hours

---

## User Story

As a **host**, I want **RSVP deadlines strictly enforced** so that **guests cannot respond after the deadline**.

---

## Acceptance Criteria

- [ ] Deadline checked on RSVP submission
- [ ] Deadline checked on RSVP update
- [ ] Strict enforcement (no grace period)
- [ ] Clear error message when deadline passed
- [ ] Event details still visible after deadline
- [ ] RSVP form disabled after deadline
- [ ] Deadline displayed prominently on page
- [ ] Timezone-aware deadline checking
- [ ] All tests pass with timeout

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

- [ ] Implement deadline check function
- [ ] Add deadline check to submission flow
- [ ] Add deadline check to update flow
- [ ] Update UI to show deadline status
- [ ] Disable form after deadline
- [ ] Write tests for before deadline
- [ ] Write tests for after deadline
- [ ] Write tests for no deadline set
- [ ] Write tests for timezone edge cases

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

- [ ] All acceptance criteria met
- [ ] Deadline check implemented
- [ ] UI updated for deadline state
- [ ] Tests passing (100% coverage)
- [ ] Timezone handling correct
- [ ] Error messages clear
- [ ] Documentation updated

---

## References

- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
