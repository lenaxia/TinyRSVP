# User Story: RSVP Submission Endpoint

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **guest**, I want **to submit my RSVP response** so that **the host knows whether I'm attending**.

---

## Acceptance Criteria

- [ ] Guest can submit RSVP via POST endpoint
- [ ] Response validated (yes/no/maybe only)
- [ ] Plus ones validated against invite limit
- [ ] Token validated before submission
- [ ] Deadline enforced (no submission after deadline)
- [ ] RSVP and answers saved atomically (transaction)
- [ ] Invite status updated to "responded"
- [ ] Duplicate RSVP prevented (one per invite)
- [ ] Clear error messages for validation failures
- [ ] Success response includes RSVP details
- [ ] All tests pass with timeout

---

## Technical Details

### Endpoint
```
POST /api/rsvp/:token
Content-Type: application/json

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

Response 201 Created:
{
    "rsvp": {
        "id": 123,
        "invite_id": 1,
        "response": "yes",
        "plus_ones": 2,
        "created_at": "2026-01-08T20:00:00Z"
    },
    "message": "RSVP submitted successfully"
}
```

### Service Interface

```go
package rsvp

type Service interface {
    SubmitRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) (*models.RSVP, error)
}

type SubmitRSVPRequest struct {
    Response string          `json:"response"`
    PlusOnes int             `json:"plus_ones"`
    Answers  []AnswerRequest `json:"answers"`
}

type AnswerRequest struct {
    QuestionID    int64   `json:"question_id"`
    AnswerText    *string `json:"answer_text,omitempty"`
    AnswerOption  *string `json:"answer_option,omitempty"`
    AnswerBoolean *bool   `json:"answer_boolean,omitempty"`
}
```

---

## Tasks

### Phase 1: Service Layer (TDD)
- [ ] Create RSVP service interface
- [ ] Write test for valid RSVP submission
- [ ] Write test for invalid token
- [ ] Write test for expired token
- [ ] Write test for revoked invite
- [ ] Write test for invalid response value
- [ ] Write test for plus ones exceeding limit
- [ ] Write test for negative plus ones
- [ ] Write test for deadline passed
- [ ] Write test for duplicate RSVP
- [ ] Write test for cancelled event
- [ ] Write test for missing required answers
- [ ] Write test for invalid answer types
- [ ] Implement SubmitRSVP method
- [ ] Run tests (should pass)

### Phase 2: Handler Layer (TDD)
- [ ] Create RSVP handler
- [ ] Write test for successful submission
- [ ] Write test for invalid JSON
- [ ] Write test for missing required fields
- [ ] Write test for validation errors
- [ ] Write test for service errors
- [ ] Implement POST handler
- [ ] Run tests (should pass)

### Phase 3: Validation Logic
- [ ] Implement response validation
- [ ] Implement plus ones validation
- [ ] Implement deadline checking
- [ ] Implement answer validation
- [ ] Test all validation rules

### Phase 4: Transaction Management
- [ ] Implement atomic RSVP + answers save
- [ ] Test transaction rollback on error
- [ ] Test invite status update
- [ ] Verify data consistency

### Phase 5: Integration Testing
- [ ] Test full submission flow
- [ ] Test with various response types
- [ ] Test with multiple answers
- [ ] Test error scenarios
- [ ] Test concurrent submissions

---

## Validation Rules

### Response Validation
- Required field
- Must be exactly: "yes", "no", or "maybe"
- Case-sensitive (lowercase only)
- No whitespace allowed

### Plus Ones Validation
- Must be integer >= 0
- Cannot exceed invite.max_plus_ones
- Automatically set to 0 if response is "no"
- Validated before database save

### Deadline Validation
- Check event.rsvp_deadline if set
- Compare with current time (UTC)
- Strict enforcement (no grace period)
- Clear error message if past deadline

### Answer Validation
- All required questions must have answers
- Answer type must match question type
- Text answers: max 500 characters
- Option answers: must match question options
- Boolean answers: must be true or false
- One answer per question maximum

---

## Business Logic

### RSVP Submission Flow

```go
func (s *service) SubmitRSVP(ctx context.Context, token string, req *SubmitRSVPRequest) (*models.RSVP, error) {
    // 1. Validate and get invite
    invite, err := s.inviteService.ValidateToken(ctx, token)
    if err != nil {
        return nil, err
    }
    
    // 2. Get event and check deadline
    event, err := s.eventRepo.GetByID(ctx, invite.EventID)
    if err != nil {
        return nil, err
    }
    
    if event.RSVPDeadline != nil && time.Now().After(*event.RSVPDeadline) {
        return nil, ErrDeadlinePassed
    }
    
    // 3. Validate request
    if err := s.validator.ValidateSubmission(req, invite, event); err != nil {
        return nil, err
    }
    
    // 4. Check for existing RSVP
    existing, _ := s.rsvpRepo.GetByInviteID(ctx, invite.ID)
    if existing != nil {
        return nil, ErrDuplicateRSVP
    }
    
    // 5. Save in transaction
    return s.db.WithTransaction(ctx, func(tx *sql.Tx) (*models.RSVP, error) {
        rsvp := &models.RSVP{
            InviteID: invite.ID,
            Response: models.RSVPResponse(req.Response),
            PlusOnes: req.PlusOnes,
        }
        
        if err := s.rsvpRepo.Create(ctx, rsvp); err != nil {
            return nil, err
        }
        
        for _, ansReq := range req.Answers {
            answer := &models.RSVPAnswer{
                RSVPID:        rsvp.ID,
                QuestionID:    ansReq.QuestionID,
                AnswerText:    ansReq.AnswerText,
                AnswerOption:  ansReq.AnswerOption,
                AnswerBoolean: ansReq.AnswerBoolean,
            }
            if err := s.answerRepo.Create(ctx, answer); err != nil {
                return nil, err
            }
        }
        
        if err := s.inviteRepo.UpdateStatus(ctx, invite.ID, models.InviteStatusResponded); err != nil {
            return nil, err
        }
        
        return rsvp, nil
    })
}
```

### Plus Ones Auto-Correction

```go
func (v *validator) ValidateSubmission(req *SubmitRSVPRequest, invite *models.Invite, event *models.Event) error {
    // Auto-correct plus ones for "no" response
    if req.Response == "no" && req.PlusOnes > 0 {
        req.PlusOnes = 0
    }
    
    // Validate plus ones limit
    if req.PlusOnes < 0 {
        return &models.ValidationError{
            Field:   "plus_ones",
            Message: "Plus ones cannot be negative",
        }
    }
    
    if req.PlusOnes > invite.MaxPlusOnes {
        return &models.ValidationError{
            Field:   "plus_ones",
            Message: fmt.Sprintf("You can bring up to %d guest(s)", invite.MaxPlusOnes),
        }
    }
    
    return nil
}
```

---

## Error Handling

| Error Condition | Error Type | HTTP Status | User Message |
|----------------|------------|-------------|--------------|
| Invalid token | `InvalidTokenError` | 404 | "Invalid invite link" |
| Expired token | `ExpiredTokenError` | 403 | "This invite has expired" |
| Revoked invite | `RevokedInviteError` | 403 | "This invite has been revoked" |
| Invalid response | `ValidationError` | 400 | "Response must be yes, no, or maybe" |
| Plus ones exceeded | `ValidationError` | 400 | "You can bring up to X guest(s)" |
| Deadline passed | `DeadlinePassedError` | 403 | "RSVP deadline has passed" |
| Duplicate RSVP | `ConflictError` | 409 | "You have already responded to this invite" |
| Cancelled event | `ValidationError` | 400 | "This event has been cancelled" |
| Missing required answer | `ValidationError` | 400 | "Please answer all required questions" |
| Invalid answer type | `ValidationError` | 400 | "Invalid answer format for question" |
| Database error | `InternalError` | 500 | "Failed to save RSVP. Please try again" |

---

## Testing Strategy

### Unit Tests

```go
func TestRSVPService_SubmitRSVP(t *testing.T) {
    tests := []struct {
        name      string
        token     string
        req       *SubmitRSVPRequest
        setupMock func(*mocks.MockInviteService, *mocks.MockEventRepo)
        wantErr   bool
        errType   error
    }{
        {
            name:  "valid yes with plus ones",
            token: "validtoken",
            req: &SubmitRSVPRequest{
                Response: "yes",
                PlusOnes: 2,
                Answers:  []AnswerRequest{},
            },
            setupMock: func(is *mocks.MockInviteService, er *mocks.MockEventRepo) {
                is.ValidateTokenFunc = func(ctx context.Context, token string) (*models.Invite, error) {
                    return &models.Invite{
                        ID:          1,
                        EventID:     1,
                        MaxPlusOnes: 2,
                        Status:      models.InviteStatusSent,
                    }, nil
                }
                er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
                    future := time.Now().Add(24 * time.Hour)
                    return &models.Event{
                        ID:           1,
                        RSVPDeadline: &future,
                        Status:       models.EventStatusPublished,
                    }, nil
                }
            },
            wantErr: false,
        },
        {
            name:  "plus ones exceed limit",
            token: "validtoken",
            req: &SubmitRSVPRequest{
                Response: "yes",
                PlusOnes: 5,
            },
            setupMock: func(is *mocks.MockInviteService, er *mocks.MockEventRepo) {
                is.ValidateTokenFunc = func(ctx context.Context, token string) (*models.Invite, error) {
                    return &models.Invite{
                        ID:          1,
                        EventID:     1,
                        MaxPlusOnes: 2,
                    }, nil
                }
            },
            wantErr: true,
            errType: &models.ValidationError{},
        },
        {
            name:  "deadline passed",
            token: "validtoken",
            req: &SubmitRSVPRequest{
                Response: "yes",
                PlusOnes: 0,
            },
            setupMock: func(is *mocks.MockInviteService, er *mocks.MockEventRepo) {
                is.ValidateTokenFunc = func(ctx context.Context, token string) (*models.Invite, error) {
                    return &models.Invite{ID: 1, EventID: 1}, nil
                }
                er.GetByIDFunc = func(ctx context.Context, id int64) (*models.Event, error) {
                    past := time.Now().Add(-24 * time.Hour)
                    return &models.Event{
                        ID:           1,
                        RSVPDeadline: &past,
                    }, nil
                }
            },
            wantErr: true,
            errType: ErrDeadlinePassed,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockInvite := &mocks.MockInviteService{}
            mockEvent := &mocks.MockEventRepo{}
            tt.setupMock(mockInvite, mockEvent)
            
            service := NewService(mockInvite, mockEvent, nil, nil, nil)
            
            _, err := service.SubmitRSVP(context.Background(), tt.token, tt.req)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("SubmitRSVP() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if tt.wantErr && tt.errType != nil {
                if !errors.As(err, &tt.errType) {
                    t.Errorf("Error type = %T, want %T", err, tt.errType)
                }
            }
        })
    }
}
```

### Integration Tests

```go
func TestRSVPSubmission_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    // Create test event
    event := createTestEvent(t, db, time.Now().Add(48*time.Hour))
    
    // Create test invite
    invite, token := createTestInvite(t, db, event.ID, 2)
    
    // Create test questions
    q1 := createTestQuestion(t, db, event.ID, "text", true)
    q2 := createTestQuestion(t, db, event.ID, "select", false)
    
    // Submit RSVP
    req := &SubmitRSVPRequest{
        Response: "yes",
        PlusOnes: 2,
        Answers: []AnswerRequest{
            {
                QuestionID: q1.ID,
                AnswerText: strPtr("Vegetarian"),
            },
            {
                QuestionID: q2.ID,
                AnswerOption: strPtr("red"),
            },
        },
    }
    
    service := setupService(t, db)
    rsvp, err := service.SubmitRSVP(context.Background(), token, req)
    
    if err != nil {
        t.Fatalf("SubmitRSVP() error = %v", err)
    }
    
    // Verify RSVP saved
    if rsvp.ID == 0 {
        t.Error("Expected RSVP ID to be set")
    }
    
    if rsvp.Response != "yes" {
        t.Errorf("Response = %s, want yes", rsvp.Response)
    }
    
    // Verify answers saved
    answers, err := service.GetAnswers(context.Background(), rsvp.ID)
    if err != nil {
        t.Fatalf("GetAnswers() error = %v", err)
    }
    
    if len(answers) != 2 {
        t.Errorf("Got %d answers, want 2", len(answers))
    }
    
    // Verify invite status updated
    updatedInvite, err := service.GetInvite(context.Background(), invite.ID)
    if err != nil {
        t.Fatalf("GetInvite() error = %v", err)
    }
    
    if updatedInvite.Status != models.InviteStatusResponded {
        t.Errorf("Invite status = %s, want responded", updatedInvite.Status)
    }
}
```

---

## Dependencies

**Depends on:**
- Story 00: RSVP Model (for data structures)
- Story 01: RSVP Page (for UI to submit from)
- Epic 03: Invites (for token validation)
- Story 02_STORY_05: Preference Questions (for answers)

**Blocks:**
- Story 08: RSVP Updates (needs submission logic)
- Story 10: Confirmation Page (needs successful submission)
- Story 11: Confirmation Email (needs submission event)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Service layer implemented and tested
- [ ] Handler layer implemented and tested
- [ ] Validation logic complete
- [ ] Transaction management working
- [ ] Unit tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] Error handling complete
- [ ] Documentation updated
- [ ] Code reviewed
- [ ] No linter warnings

---

## References

- **HLD:** Section 7.3 (RSVP Submission)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md) - Section 4.1
- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **API:** [lld/08_API_LLD.md](../lld/08_API_LLD.md)
