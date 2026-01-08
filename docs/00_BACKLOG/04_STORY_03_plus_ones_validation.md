# User Story: Plus Ones Validation Logic

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 4 hours

---

## User Story

As a **developer**, I want **comprehensive plus ones validation logic** so that **guests cannot exceed their allowed plus ones limit**.

---

## Acceptance Criteria

- [x] Plus ones validated against invite.max_plus_ones
- [x] Plus ones must be >= 0
- [x] Plus ones automatically set to 0 for "no" response
- [x] Plus ones cannot exceed event.max_plus_ones
- [x] Clear error messages for validation failures
- [x] Validation occurs before database save
- [x] All validation tests pass with timeout
- [x] Edge cases handled (null, negative, overflow)

---

## Technical Details

### Validator Interface

```go
package rsvp

type PlusOnesValidator interface {
    ValidatePlusOnes(plusOnes int, response RSVPResponse, invite *models.Invite) error
}

type validator struct{}

func NewValidator() PlusOnesValidator {
    return &validator{}
}
```

### Validation Rules

1. **Non-negative**: Plus ones >= 0
2. **Within invite limit**: Plus ones <= invite.max_plus_ones
3. **Auto-correction**: If response = "no", force plus ones = 0
4. **Type safety**: Must be integer (no floats, strings)

---

## Tasks

### Phase 1: Validator Implementation (TDD)
- [x] Write test for valid plus ones (0 to limit)
- [x] Write test for negative plus ones
- [x] Write test for exceeding invite limit
- [x] Write test for "no" response with plus ones
- [x] Write test for "maybe" response with plus ones
- [x] Write test for zero max_plus_ones
- [x] Write test for boundary values
- [x] Implement ValidatePlusOnes method
- [x] Run tests (should pass)

### Phase 2: Integration with RSVP Service
- [x] Wire validator into RSVP service
- [x] Test validation in submission flow
- [x] Test validation in update flow
- [x] Verify error messages

### Phase 3: Edge Cases
- [x] Test with max int value
- [x] Test with invite.max_plus_ones = 0
- [x] Test with invite.max_plus_ones = 10 (max)
- [x] Test concurrent validations

---

## Validation Logic

```go
func (v *validator) ValidatePlusOnes(plusOnes int, response RSVPResponse, invite *models.Invite) error {
    // Auto-correct for "no" response
    if response == RSVPResponseNo && plusOnes > 0 {
        return &models.ValidationError{
            Field:   "plus_ones",
            Message: "Cannot bring guests when declining",
        }
    }
    
    // Check non-negative
    if plusOnes < 0 {
        return &models.ValidationError{
            Field:   "plus_ones",
            Message: "Plus ones cannot be negative",
        }
    }
    
    // Check against invite limit
    if plusOnes > invite.MaxPlusOnes {
        return &models.ValidationError{
            Field:   "plus_ones",
            Message: fmt.Sprintf("You can bring up to %d guest(s)", invite.MaxPlusOnes),
        }
    }
    
    return nil
}
```

---

## Testing Requirements

### Unit Tests

```go
func TestPlusOnesValidator_ValidatePlusOnes(t *testing.T) {
    tests := []struct {
        name      string
        plusOnes  int
        response  RSVPResponse
        invite    *models.Invite
        wantErr   bool
        errMsg    string
    }{
        {
            name:     "valid within limit",
            plusOnes: 2,
            response: RSVPResponseYes,
            invite:   &models.Invite{MaxPlusOnes: 5},
            wantErr:  false,
        },
        {
            name:     "zero plus ones",
            plusOnes: 0,
            response: RSVPResponseYes,
            invite:   &models.Invite{MaxPlusOnes: 5},
            wantErr:  false,
        },
        {
            name:     "at maximum limit",
            plusOnes: 5,
            response: RSVPResponseYes,
            invite:   &models.Invite{MaxPlusOnes: 5},
            wantErr:  false,
        },
        {
            name:     "exceeds limit",
            plusOnes: 6,
            response: RSVPResponseYes,
            invite:   &models.Invite{MaxPlusOnes: 5},
            wantErr:  true,
            errMsg:   "can bring up to 5 guest(s)",
        },
        {
            name:     "negative plus ones",
            plusOnes: -1,
            response: RSVPResponseYes,
            invite:   &models.Invite{MaxPlusOnes: 5},
            wantErr:  true,
            errMsg:   "cannot be negative",
        },
        {
            name:     "no response with plus ones",
            plusOnes: 2,
            response: RSVPResponseNo,
            invite:   &models.Invite{MaxPlusOnes: 5},
            wantErr:  true,
            errMsg:   "Cannot bring guests when declining",
        },
        {
            name:     "maybe response with plus ones",
            plusOnes: 2,
            response: RSVPResponseMaybe,
            invite:   &models.Invite{MaxPlusOnes: 5},
            wantErr:  false,
        },
        {
            name:     "zero max allowed",
            plusOnes: 1,
            response: RSVPResponseYes,
            invite:   &models.Invite{MaxPlusOnes: 0},
            wantErr:  true,
            errMsg:   "can bring up to 0 guest(s)",
        },
    }
    
    validator := NewValidator()
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidatePlusOnes(tt.plusOnes, tt.response, tt.invite)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePlusOnes() error = %v, wantErr %v", err, tt.wantErr)
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

## Error Messages

| Condition | User-Friendly Message |
|-----------|----------------------|
| Negative value | "Plus ones cannot be negative" |
| Exceeds limit | "You can bring up to X guest(s)" |
| No response with guests | "Cannot bring guests when declining" |
| Zero allowed | "This invite does not allow plus ones" |

---

## Dependencies

**Depends on:**
- Story 00: RSVP Model (for data structures)
- Epic 03: Invites (for invite.max_plus_ones)

**Blocks:**
- Story 02: RSVP Submission (needs validation)
- Story 04: Plus Ones UI (needs validation logic)
- Story 08: RSVP Updates (needs validation)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Validator implemented and tested
- [x] Unit tests passing (100% coverage)
- [x] Integration tests passing
- [x] Error messages clear and actionable
- [x] Edge cases handled
- [x] Documentation updated
- [x] Code reviewed
- [x] No linter warnings

---

## References

- **HLD:** Section 7.4 (Plus Ones Validation)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md) - Section 5.1
- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
