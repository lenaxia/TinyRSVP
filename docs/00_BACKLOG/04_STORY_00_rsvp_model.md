# User Story: RSVP Model and Repository

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 4 hours

---

## User Story

As a **developer**, I want **a complete RSVP model with repository layer** so that **RSVP data can be properly structured, validated, and persisted**.

---

## Acceptance Criteria

- [ ] RSVP struct matches database schema
- [ ] RSVPResponse enum with all response types defined
- [ ] RSVP repository interface defined
- [ ] RSVP repository implementation with CRUD operations
- [ ] RSVPAnswer model for preference question answers
- [ ] Answer repository interface and implementation
- [ ] Unique constraint enforced (one RSVP per invite)
- [ ] Plus ones validation (0 to invite.max_plus_ones)
- [ ] Response validation (yes/no/maybe only)
- [ ] All repository tests pass with timeout
- [ ] Integration tests for RSVP creation with answers

---

## Technical Details

### RSVP Model Structure

```go
type RSVP struct {
    ID        int64        `db:"id" json:"id"`
    InviteID  int64        `db:"invite_id" json:"invite_id"`
    Response  RSVPResponse `db:"response" json:"response"`
    PlusOnes  int          `db:"plus_ones" json:"plus_ones"`
    CreatedAt time.Time    `db:"created_at" json:"created_at"`
    UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}

type RSVPResponse string

const (
    RSVPResponseYes   RSVPResponse = "yes"
    RSVPResponseNo    RSVPResponse = "no"
    RSVPResponseMaybe RSVPResponse = "maybe"
)
```

### RSVPAnswer Model Structure

```go
type RSVPAnswer struct {
    ID            int64     `db:"id" json:"id"`
    RSVPID        int64     `db:"rsvp_id" json:"rsvp_id"`
    QuestionID    int64     `db:"question_id" json:"question_id"`
    AnswerText    *string   `db:"answer_text" json:"answer_text,omitempty"`
    AnswerOption  *string   `db:"answer_option" json:"answer_option,omitempty"`
    AnswerBoolean *bool     `db:"answer_boolean" json:"answer_boolean,omitempty"`
    CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
```

### Repository Interfaces

```go
package repositories

type RSVPRepository interface {
    Create(ctx context.Context, rsvp *models.RSVP) error
    GetByID(ctx context.Context, id int64) (*models.RSVP, error)
    GetByInviteID(ctx context.Context, inviteID int64) (*models.RSVP, error)
    GetByEventID(ctx context.Context, eventID int64) ([]*models.RSVP, error)
    Update(ctx context.Context, rsvp *models.RSVP) error
    GetStats(ctx context.Context, eventID int64) (*RSVPStats, error)
}

type AnswerRepository interface {
    Create(ctx context.Context, answer *models.RSVPAnswer) error
    GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error)
    GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error)
    Update(ctx context.Context, answer *models.RSVPAnswer) error
    DeleteByRSVPID(ctx context.Context, rsvpID int64) error
}

type RSVPStats struct {
    TotalInvites int
    YesCount     int
    NoCount      int
    MaybeCount   int
    NoResponse   int
    TotalGuests  int
}
```

---

## Tasks

### Phase 1: RSVP Model (TDD)
- [ ] Create RSVP struct in [`internal/models/rsvp.go`](../../internal/models/rsvp.go)
- [ ] Define RSVPResponse enum with constants
- [ ] Write tests for RSVP validation
- [ ] Add validation methods to RSVP model
- [ ] Create RSVPAnswer struct in same file
- [ ] Write tests for answer type validation

### Phase 2: RSVP Repository (TDD)
- [ ] Create repository interface in [`internal/db/repositories/rsvp_repository.go`](../../internal/db/repositories/rsvp_repository.go)
- [ ] Write test for Create with valid RSVP
- [ ] Write test for Create with duplicate invite_id (should fail)
- [ ] Write test for GetByID
- [ ] Write test for GetByInviteID
- [ ] Write test for GetByEventID
- [ ] Write test for Update
- [ ] Write test for GetStats
- [ ] Implement repository methods
- [ ] Run tests (should pass)

### Phase 3: Answer Repository (TDD)
- [ ] Create repository interface in [`internal/db/repositories/answer_repository.go`](../../internal/db/repositories/answer_repository.go)
- [ ] Write test for Create with text answer
- [ ] Write test for Create with option answer
- [ ] Write test for Create with boolean answer
- [ ] Write test for GetByRSVPID
- [ ] Write test for GetByQuestionID
- [ ] Write test for Update
- [ ] Write test for DeleteByRSVPID
- [ ] Implement repository methods
- [ ] Run tests (should pass)

### Phase 4: Integration Tests
- [ ] Write integration test for RSVP creation
- [ ] Write integration test for RSVP with multiple answers
- [ ] Write integration test for duplicate RSVP prevention
- [ ] Write integration test for GetStats calculation
- [ ] Run all tests with timeout

### Phase 5: Documentation
- [ ] Document RSVP model in README
- [ ] Document repository interfaces
- [ ] Add usage examples
- [ ] Update database schema documentation

---

## Testing Requirements

### Unit Tests

```go
func TestRSVP_Validate(t *testing.T) {
    tests := []struct {
        name    string
        rsvp    *models.RSVP
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid yes response",
            rsvp: &models.RSVP{
                InviteID: 1,
                Response: models.RSVPResponseYes,
                PlusOnes: 2,
            },
            wantErr: false,
        },
        {
            name: "invalid response",
            rsvp: &models.RSVP{
                InviteID: 1,
                Response: "invalid",
                PlusOnes: 0,
            },
            wantErr: true,
            errMsg:  "response must be yes, no, or maybe",
        },
        {
            name: "negative plus ones",
            rsvp: &models.RSVP{
                InviteID: 1,
                Response: models.RSVPResponseYes,
                PlusOnes: -1,
            },
            wantErr: true,
            errMsg:  "plus ones cannot be negative",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.rsvp.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if tt.wantErr && err != nil {
                if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("Validate() error = %q, want to contain %q", err.Error(), tt.errMsg)
                }
            }
        })
    }
}

func TestRSVPRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewRSVPRepository(db)
    
    rsvp := &models.RSVP{
        InviteID: 1,
        Response: models.RSVPResponseYes,
        PlusOnes: 2,
    }
    
    err := repo.Create(context.Background(), rsvp)
    if err != nil {
        t.Fatalf("Create() error = %v", err)
    }
    
    if rsvp.ID == 0 {
        t.Error("Expected ID to be set after creation")
    }
    
    // Test duplicate prevention
    duplicate := &models.RSVP{
        InviteID: 1,
        Response: models.RSVPResponseNo,
        PlusOnes: 0,
    }
    
    err = repo.Create(context.Background(), duplicate)
    if err == nil {
        t.Error("Expected error for duplicate invite_id, got nil")
    }
}

func TestRSVPRepository_GetStats(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewRSVPRepository(db)
    
    // Create test data
    createTestRSVPs(t, db, 1, 5, 3, 2) // eventID, yes, no, maybe
    
    stats, err := repo.GetStats(context.Background(), 1)
    if err != nil {
        t.Fatalf("GetStats() error = %v", err)
    }
    
    if stats.YesCount != 5 {
        t.Errorf("YesCount = %d, want 5", stats.YesCount)
    }
    if stats.NoCount != 3 {
        t.Errorf("NoCount = %d, want 3", stats.NoCount)
    }
    if stats.MaybeCount != 2 {
        t.Errorf("MaybeCount = %d, want 2", stats.MaybeCount)
    }
}
```

---

## Validation Rules

### RSVP Response
- Required
- Must be one of: "yes", "no", "maybe"
- Case-sensitive
- Stored as lowercase

### Plus Ones
- Integer >= 0
- Must be validated against invite.max_plus_ones at service layer
- Automatically set to 0 if response is "no"

### Answer Types
- Text: Max 500 characters, trimmed
- Option: Must match one of question's options
- Boolean: Must be true or false
- Only one answer type field should be populated per answer

---

## Database Schema

### rsvps Table
```sql
CREATE TABLE rsvps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invite_id INTEGER NOT NULL UNIQUE,
    response TEXT NOT NULL CHECK(response IN ('yes', 'no', 'maybe')),
    plus_ones INTEGER NOT NULL DEFAULT 0 CHECK(plus_ones >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (invite_id) REFERENCES invites(id) ON DELETE CASCADE
);

CREATE INDEX idx_rsvps_invite_id ON rsvps(invite_id);
```

### rsvp_answers Table
```sql
CREATE TABLE rsvp_answers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rsvp_id INTEGER NOT NULL,
    question_id INTEGER NOT NULL,
    answer_text TEXT,
    answer_option TEXT,
    answer_boolean INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (rsvp_id) REFERENCES rsvps(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES preference_questions(id) ON DELETE CASCADE,
    UNIQUE(rsvp_id, question_id)
);

CREATE INDEX idx_rsvp_answers_rsvp_id ON rsvp_answers(rsvp_id);
CREATE INDEX idx_rsvp_answers_question_id ON rsvp_answers(question_id);
```

---

## Dependencies

**Depends on:**
- Epic 00 (Foundation) - Complete
- Epic 02 (Events) - Complete (for event lookup)
- Epic 03 (Invites) - Complete (for invite relationship)
- Story 02_STORY_05 (Preference Questions) - Complete

**Blocks:**
- All other RSVP stories in Epic 04

**External Dependencies:**
- SQLite database with foreign key support

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/db/repositories/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] UNIQUE constraint on invite_id enforced
- [ ] Repository methods documented
- [ ] Changes committed to git

---

## Implementation Notes

### RSVP Stats Calculation

The GetStats method should efficiently calculate:
- Total invites for event (from invites table)
- Count of each response type
- Total guests (sum of 1 + plus_ones for "yes" responses)
- No response count (total invites - RSVPs)

```go
func (r *rsvpRepository) GetStats(ctx context.Context, eventID int64) (*RSVPStats, error) {
    query := `
        SELECT 
            COUNT(DISTINCT i.id) as total_invites,
            COUNT(CASE WHEN r.response = 'yes' THEN 1 END) as yes_count,
            COUNT(CASE WHEN r.response = 'no' THEN 1 END) as no_count,
            COUNT(CASE WHEN r.response = 'maybe' THEN 1 END) as maybe_count,
            COALESCE(SUM(CASE WHEN r.response = 'yes' THEN 1 + r.plus_ones ELSE 0 END), 0) as total_guests
        FROM invites i
        LEFT JOIN rsvps r ON i.id = r.invite_id
        WHERE i.event_id = ?
    `
    
    var stats RSVPStats
    err := r.db.QueryRowContext(ctx, query, eventID).Scan(
        &stats.TotalInvites,
        &stats.YesCount,
        &stats.NoCount,
        &stats.MaybeCount,
        &stats.TotalGuests,
    )
    
    stats.NoResponse = stats.TotalInvites - stats.YesCount - stats.NoCount - stats.MaybeCount
    
    return &stats, err
}
```

### Answer Type Validation

Only one answer field should be populated:

```go
func (a *RSVPAnswer) Validate() error {
    populated := 0
    if a.AnswerText != nil {
        populated++
    }
    if a.AnswerOption != nil {
        populated++
    }
    if a.AnswerBoolean != nil {
        populated++
    }
    
    if populated != 1 {
        return fmt.Errorf("exactly one answer field must be populated")
    }
    
    return nil
}
```

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **HLD:** Section 7 (RSVP Model)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md) - Section 4.1
- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **Database:** [lld/07_DATABASE_LLD.md](../lld/07_DATABASE_LLD.md)
