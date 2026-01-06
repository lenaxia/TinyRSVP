# Domain 4: RSVP & Preference Questions - Low-Level Design

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Implementation Ready  
**HLD Reference:** [Section 7 - RSVP Model](../02_REVISED_HLD.md#7-rsvp-model), [Section 8 - Preference Questions](../02_REVISED_HLD.md#8-preference-questions)

---

## 1. Overview

### 1.1 Purpose

Manages RSVP submissions, updates, preference questions, and answer collection with validation and deadline enforcement.

### 1.2 Responsibilities

- RSVP submission and updates
- RSVP state transitions (yes/no/maybe)
- Plus ones validation
- RSVP deadline enforcement
- Preference question CRUD
- Answer validation and storage
- Question ordering
- Answer history (future)

### 1.3 Design Principles

- **Deadline Enforced** - Strict RSVP deadline checking
- **Validated** - Plus ones within limits
- **Atomic** - RSVP + answers in single transaction
- **Type-Safe** - Answer type matches question type
- **Audit Logged** - Track all RSVP changes

---

## 2. Package Structure

```
internal/
├── rsvp/
│   ├── service.go              # RSVP service
│   ├── service_test.go
│   ├── validator.go            # RSVP validation
│   ├── validator_test.go
│   ├── questions.go            # Question service
│   ├── questions_test.go
│   └── answers.go              # Answer service
│       └── answers_test.go
```

---

## 3. Interfaces

### 3.1 RSVP Service Interface

```go
package rsvp

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type Service interface {
    SubmitRSVP(ctx context.Context, req *RSVPRequest) error
    UpdateRSVP(ctx context.Context, inviteID int64, req *RSVPRequest) error
    GetRSVP(ctx context.Context, inviteID int64) (*RSVPWithAnswers, error)
    GetEventRSVPs(ctx context.Context, eventID int64) ([]*RSVPWithAnswers, error)
    GetRSVPStats(ctx context.Context, eventID int64) (*RSVPStats, error)
}

type RSVPRequest struct {
    InviteID int64
    Response RSVPResponse
    PlusOnes int
    Answers  []AnswerRequest
}

type RSVPResponse string

const (
    RSVPResponseYes   RSVPResponse = "yes"
    RSVPResponseNo    RSVPResponse = "no"
    RSVPResponseMaybe RSVPResponse = "maybe"
)

type AnswerRequest struct {
    QuestionID    int64
    AnswerText    *string
    AnswerOption  *string
    AnswerBoolean *bool
}

type RSVPWithAnswers struct {
    RSVP    *models.RSVP
    Answers []*models.RSVPAnswer
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

### 3.2 Question Service Interface

```go
package rsvp

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type QuestionService interface {
    CreateQuestion(ctx context.Context, question *models.PreferenceQuestion) error
    GetQuestion(ctx context.Context, id int64) (*models.PreferenceQuestion, error)
    GetEventQuestions(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
    UpdateQuestion(ctx context.Context, question *models.PreferenceQuestion) error
    DeleteQuestion(ctx context.Context, id int64) error
    ReorderQuestions(ctx context.Context, eventID int64, questionIDs []int64) error
}
```

---

## 4. Implementation

### 4.1 RSVP Service

```go
package rsvp

import (
    "context"
    "fmt"
    "time"
    
    "github.com/yourusername/tinyrsvp/internal/db/repositories"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type service struct {
    rsvpRepo     repositories.RSVPRepository
    inviteRepo   repositories.InviteRepository
    eventRepo    repositories.EventRepository
    answerRepo   repositories.AnswerRepository
    questionRepo repositories.QuestionRepository
    validator    Validator
    db           db.Database
}

func NewService(
    rsvpRepo repositories.RSVPRepository,
    inviteRepo repositories.InviteRepository,
    eventRepo repositories.EventRepository,
    answerRepo repositories.AnswerRepository,
    questionRepo repositories.QuestionRepository,
    validator Validator,
    database db.Database,
) Service {
    return &service{
        rsvpRepo:     rsvpRepo,
        inviteRepo:   inviteRepo,
        eventRepo:    eventRepo,
        answerRepo:   answerRepo,
        questionRepo: questionRepo,
        validator:    validator,
        db:           database,
    }
}

func (s *service) SubmitRSVP(ctx context.Context, req *RSVPRequest) error {
    invite, err := s.inviteRepo.GetByID(ctx, req.InviteID)
    if err != nil {
        return err
    }
    
    if !invite.CanRSVP() {
        return fmt.Errorf("cannot RSVP: invite revoked or expired")
    }
    
    event, err := s.eventRepo.GetByID(ctx, invite.EventID)
    if err != nil {
        return err
    }
    
    if event.RSVPDeadline != nil && time.Now().After(*event.RSVPDeadline) {
        return fmt.Errorf("RSVP deadline has passed")
    }
    
    if err := s.validator.ValidateRSVP(req, invite); err != nil {
        return err
    }
    
    return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
        rsvp := &models.RSVP{
            InviteID: invite.ID,
            Response: req.Response,
            PlusOnes: req.PlusOnes,
        }
        
        if err := s.rsvpRepo.Create(ctx, rsvp); err != nil {
            return err
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
                return err
            }
        }
        
        return s.inviteRepo.UpdateStatus(ctx, invite.ID, models.InviteStatusResponded)
    })
}
```

---

## 5. Validation

### 5.1 RSVP Validation

```go
func (v *validator) ValidateRSVP(req *RSVPRequest, invite *models.Invite) error {
    if req.Response != "yes" && req.Response != "no" && req.Response != "maybe" {
        return &models.ValidationError{
            Field:   "response",
            Message: "Response must be yes, no, or maybe",
        }
    }
    
    if req.PlusOnes < 0 || req.PlusOnes > invite.MaxPlusOnes {
        return &models.ValidationError{
            Field:   "plus_ones",
            Message: fmt.Sprintf("You can bring up to %d guest(s)", invite.MaxPlusOnes),
        }
    }
    
    if req.Response != "yes" && req.PlusOnes > 0 {
        req.PlusOnes = 0
    }
    
    return nil
}
```

---

## 6. Dependencies

**Internal:**
- Domain 3 (Invite) - RSVPs belong to invites
- Domain 2 (Event) - Questions belong to events, deadline checking
- Domain 7 (Database) - RSVP and question repositories

**Dependents:**
- Domain 5 (Email) - Confirmation emails

---

## 7. Testing

```go
func TestRSVPService_SubmitRSVP(t *testing.T) {
    tests := []struct {
        name    string
        req     *RSVPRequest
        invite  *models.Invite
        wantErr bool
    }{
        {
            name: "valid yes with plus ones",
            req: &RSVPRequest{
                Response: "yes",
                PlusOnes: 2,
            },
            invite: &models.Invite{
                MaxPlusOnes: 2,
                Status:      models.InviteStatusSent,
            },
            wantErr: false,
        },
        {
            name: "exceeds plus ones limit",
            req: &RSVPRequest{
                Response: "yes",
                PlusOnes: 5,
            },
            invite: &models.Invite{
                MaxPlusOnes: 2,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.SubmitRSVP(context.Background(), tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("SubmitRSVP() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

**Document Status:** ✅ Complete

**Next Domain:** [Domain 5: Email System](05_EMAIL_LLD.md)
