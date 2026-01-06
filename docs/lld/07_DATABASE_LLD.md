# Domain 7: Database & Persistence - Low-Level Design

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Implementation Ready  
**HLD Reference:** [Section 13 - Database Schema](../02_REVISED_HLD.md#13-database-schema)

---

## 1. Overview

### 1.1 Purpose

Provides the foundational data persistence layer for TinyRSVP, including database connection management, repository pattern implementation, transaction handling, and migrations.

### 1.2 Responsibilities

- Database connection lifecycle management
- SQL query execution with parameterization
- Transaction management with rollback support
- Repository pattern implementation for all entities
- Database migration execution
- Connection pooling and optimization
- Audit logging persistence

### 1.3 Design Principles

- **Repository Pattern** - Abstract database implementation
- **Interface-Based** - All repositories implement common interface
- **Transaction Support** - Atomic operations across repositories
- **Type Safety** - Strongly-typed models, no `map[string]interface{}`
- **Error Wrapping** - Context-preserving error handling
- **Prepared Statements** - SQL injection prevention

---

## 2. Package Structure

```
internal/
├── db/
│   ├── db.go                    # Database connection and utilities
│   ├── db_test.go               # Database tests
│   ├── transaction.go           # Transaction management
│   ├── transaction_test.go
│   ├── migrations.go            # Migration execution
│   ├── migrations_test.go
│   └── repositories/            # Repository implementations
│       ├── user_repository.go
│       ├── user_repository_test.go
│       ├── session_repository.go
│       ├── session_repository_test.go
│       ├── event_repository.go
│       ├── event_repository_test.go
│       ├── invite_repository.go
│       ├── invite_repository_test.go
│       ├── rsvp_repository.go
│       ├── rsvp_repository_test.go
│       ├── question_repository.go
│       ├── question_repository_test.go
│       ├── email_queue_repository.go
│       ├── email_queue_repository_test.go
│       ├── template_repository.go
│       ├── template_repository_test.go
│       ├── audit_log_repository.go
│       └── audit_log_repository_test.go
├── models/
│   ├── user.go                  # User model
│   ├── session.go               # Session model
│   ├── event.go                 # Event model
│   ├── invite.go                # Invite model
│   ├── rsvp.go                  # RSVP model
│   ├── question.go              # Question model
│   ├── email.go                 # Email queue model
│   ├── template.go              # Template model
│   ├── audit.go                 # Audit log model
│   └── errors.go                # Domain error types
migrations/
└── sqlite/
    ├── 001_initial_schema.up.sql
    ├── 001_initial_schema.down.sql
    ├── 002_add_audit_log.up.sql
    └── 002_add_audit_log.down.sql
```

---

## 3. Data Models

### 3.1 User Model

```go
package models

import "time"

type UserRole string

const (
    RoleAdmin        UserRole = "admin"
    RoleEventManager UserRole = "event_manager"
)

type User struct {
    ID           int64      `db:"id" json:"id"`
    Email        string     `db:"email" json:"email"`
    Name         string     `db:"name" json:"name"`
    Role         UserRole   `db:"role" json:"role"`
    OIDCSubject  *string    `db:"oidc_subject" json:"oidc_subject,omitempty"`
    CreatedAt    time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
    LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
}

func (u *User) IsAdmin() bool {
    return u.Role == RoleAdmin
}

func (u *User) IsEventManager() bool {
    return u.Role == RoleEventManager || u.Role == RoleAdmin
}
```

### 3.2 Session Model

```go
package models

import "time"

type Session struct {
    ID             string    `db:"id" json:"id"`
    UserID         int64     `db:"user_id" json:"user_id"`
    CreatedAt      time.Time `db:"created_at" json:"created_at"`
    ExpiresAt      time.Time `db:"expires_at" json:"expires_at"`
    LastAccessedAt time.Time `db:"last_accessed_at" json:"last_accessed_at"`
    IPAddress      *string   `db:"ip_address" json:"ip_address,omitempty"`
    UserAgent      *string   `db:"user_agent" json:"user_agent,omitempty"`
}

func (s *Session) IsExpired() bool {
    return time.Now().After(s.ExpiresAt)
}
```

### 3.3 Event Model

```go
package models

import "time"

type EventStatus string

const (
    EventStatusDraft     EventStatus = "draft"
    EventStatusPublished EventStatus = "published"
    EventStatusCancelled EventStatus = "cancelled"
    EventStatusArchived  EventStatus = "archived"
)

type Event struct {
    ID           int64        `db:"id" json:"id"`
    Title        string       `db:"title" json:"title"`
    Description  *string      `db:"description" json:"description,omitempty"`
    Location     *string      `db:"location" json:"location,omitempty"`
    StartTime    time.Time    `db:"start_time" json:"start_time"`
    EndTime      *time.Time   `db:"end_time" json:"end_time,omitempty"`
    Timezone     string       `db:"timezone" json:"timezone"`
    RSVPDeadline *time.Time   `db:"rsvp_deadline" json:"rsvp_deadline,omitempty"`
    MaxPlusOnes  int          `db:"max_plus_ones" json:"max_plus_ones"`
    Status       EventStatus  `db:"status" json:"status"`
    TemplateID   *int64       `db:"template_id" json:"template_id,omitempty"`
    CreatedBy    int64        `db:"created_by" json:"created_by"`
    CreatedAt    time.Time    `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
    Version      int          `db:"version" json:"version"`
    ICSSequence  int          `db:"ics_sequence" json:"ics_sequence"`
}

func (e *Event) CanEdit() bool {
    return e.Status == EventStatusDraft || e.Status == EventStatusPublished
}

func (e *Event) CanDelete() bool {
    return e.Status == EventStatusDraft || e.Status == EventStatusPublished
}

func (e *Event) IsActive() bool {
    return e.Status == EventStatusPublished
}
```

### 3.4 Invite Model

```go
package models

import "time"

type InviteStatus string

const (
    InviteStatusDraft     InviteStatus = "draft"
    InviteStatusSent      InviteStatus = "sent"
    InviteStatusViewed    InviteStatus = "viewed"
    InviteStatusResponded InviteStatus = "responded"
    InviteStatusRevoked   InviteStatus = "revoked"
)

type Invite struct {
    ID           int64        `db:"id" json:"id"`
    EventID      int64        `db:"event_id" json:"event_id"`
    Name         *string      `db:"name" json:"name,omitempty"`
    Email        *string      `db:"email" json:"email,omitempty"`
    TokenHash    string       `db:"token_hash" json:"-"`
    MaxPlusOnes  int          `db:"max_plus_ones" json:"max_plus_ones"`
    Status       InviteStatus `db:"status" json:"status"`
    SentAt       *time.Time   `db:"sent_at" json:"sent_at,omitempty"`
    ViewedAt     *time.Time   `db:"viewed_at" json:"viewed_at,omitempty"`
    Unsubscribed bool         `db:"unsubscribed" json:"unsubscribed"`
    EmailInvalid bool         `db:"email_invalid" json:"email_invalid"`
    CreatedAt    time.Time    `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
    ExpiresAt    time.Time    `db:"expires_at" json:"expires_at"`
}

func (i *Invite) IsRevoked() bool {
    return i.Status == InviteStatusRevoked
}

func (i *Invite) IsExpired() bool {
    return time.Now().After(i.ExpiresAt)
}

func (i *Invite) CanRSVP() bool {
    return !i.IsRevoked() && !i.IsExpired()
}
```

### 3.5 RSVP Model

```go
package models

import "time"

type RSVPResponse string

const (
    RSVPResponseYes   RSVPResponse = "yes"
    RSVPResponseNo    RSVPResponse = "no"
    RSVPResponseMaybe RSVPResponse = "maybe"
)

type RSVP struct {
    ID        int64        `db:"id" json:"id"`
    InviteID  int64        `db:"invite_id" json:"invite_id"`
    Response  RSVPResponse `db:"response" json:"response"`
    PlusOnes  int          `db:"plus_ones" json:"plus_ones"`
    CreatedAt time.Time    `db:"created_at" json:"created_at"`
    UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}

func (r *RSVP) IsAttending() bool {
    return r.Response == RSVPResponseYes
}

func (r *RSVP) TotalAttendees() int {
    if r.IsAttending() {
        return 1 + r.PlusOnes
    }
    return 0
}
```

### 3.6 Question and Answer Models

```go
package models

import (
    "encoding/json"
    "time"
)

type QuestionType string

const (
    QuestionTypeText    QuestionType = "text"
    QuestionTypeSelect  QuestionType = "select"
    QuestionTypeBoolean QuestionType = "boolean"
)

type SelectOption struct {
    Value string `json:"value"`
    Label string `json:"label"`
}

type PreferenceQuestion struct {
    ID           int64          `db:"id" json:"id"`
    EventID      int64          `db:"event_id" json:"event_id"`
    QuestionText string         `db:"question_text" json:"question_text"`
    QuestionType QuestionType   `db:"question_type" json:"question_type"`
    Options      json.RawMessage `db:"options" json:"options,omitempty"`
    Required     bool           `db:"required" json:"required"`
    DisplayOrder int            `db:"display_order" json:"display_order"`
    CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

func (q *PreferenceQuestion) GetOptions() ([]SelectOption, error) {
    if q.QuestionType != QuestionTypeSelect {
        return nil, nil
    }
    var options []SelectOption
    if err := json.Unmarshal(q.Options, &options); err != nil {
        return nil, err
    }
    return options, nil
}

type RSVPAnswer struct {
    ID            int64      `db:"id" json:"id"`
    RSVPID        int64      `db:"rsvp_id" json:"rsvp_id"`
    QuestionID    int64      `db:"question_id" json:"question_id"`
    AnswerText    *string    `db:"answer_text" json:"answer_text,omitempty"`
    AnswerOption  *string    `db:"answer_option" json:"answer_option,omitempty"`
    AnswerBoolean *bool      `db:"answer_boolean" json:"answer_boolean,omitempty"`
    CreatedAt     time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
```

### 3.7 Email Queue Model

```go
package models

import (
    "encoding/json"
    "time"
)

type EmailStatus string

const (
    EmailStatusPending   EmailStatus = "pending"
    EmailStatusSending   EmailStatus = "sending"
    EmailStatusSent      EmailStatus = "sent"
    EmailStatusFailed    EmailStatus = "failed"
    EmailStatusCancelled EmailStatus = "cancelled"
)

type EmailAttachment struct {
    Filename    string `json:"filename"`
    ContentType string `json:"content_type"`
    Content     []byte `json:"content"`
}

type EmailQueue struct {
    ID            int64             `db:"id" json:"id"`
    ToEmail       string            `db:"to_email" json:"to_email"`
    ToName        *string           `db:"to_name" json:"to_name,omitempty"`
    Subject       string            `db:"subject" json:"subject"`
    BodyText      string            `db:"body_text" json:"body_text"`
    BodyHTML      *string           `db:"body_html" json:"body_html,omitempty"`
    Attachments   json.RawMessage   `db:"attachments" json:"attachments,omitempty"`
    Status        EmailStatus       `db:"status" json:"status"`
    Attempts      int               `db:"attempts" json:"attempts"`
    MaxAttempts   int               `db:"max_attempts" json:"max_attempts"`
    LastAttemptAt *time.Time        `db:"last_attempt_at" json:"last_attempt_at,omitempty"`
    LastError     *string           `db:"last_error" json:"last_error,omitempty"`
    ScheduledFor  time.Time         `db:"scheduled_for" json:"scheduled_for"`
    CreatedAt     time.Time         `db:"created_at" json:"created_at"`
}

func (e *EmailQueue) GetAttachments() ([]EmailAttachment, error) {
    if len(e.Attachments) == 0 {
        return nil, nil
    }
    var attachments []EmailAttachment
    if err := json.Unmarshal(e.Attachments, &attachments); err != nil {
        return nil, err
    }
    return attachments, nil
}

func (e *EmailQueue) ShouldRetry() bool {
    return e.Status == EmailStatusPending && e.Attempts < e.MaxAttempts
}
```

### 3.8 Template Model

```go
package models

import "time"

type TemplateType string

const (
    TemplateTypeInviteEmail      TemplateType = "invite_email"
    TemplateTypeRSVPPage         TemplateType = "rsvp_page"
    TemplateTypeConfirmationPage TemplateType = "confirmation_page"
)

type Template struct {
    ID          int64        `db:"id" json:"id"`
    Name        string       `db:"name" json:"name"`
    Type        TemplateType `db:"type" json:"type"`
    HTMLContent string       `db:"html_content" json:"html_content"`
    TextContent *string      `db:"text_content" json:"text_content,omitempty"`
    CSSContent  *string      `db:"css_content" json:"css_content,omitempty"`
    IsDefault   bool         `db:"is_default" json:"is_default"`
    CreatedBy   *int64       `db:"created_by" json:"created_by,omitempty"`
    CreatedAt   time.Time    `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
}
```

### 3.9 Audit Log Model

```go
package models

import (
    "encoding/json"
    "time"
)

type AuditLog struct {
    ID           int64           `db:"id" json:"id"`
    Timestamp    time.Time       `db:"timestamp" json:"timestamp"`
    UserID       *int64          `db:"user_id" json:"user_id,omitempty"`
    Action       string          `db:"action" json:"action"`
    ResourceType string          `db:"resource_type" json:"resource_type"`
    ResourceID   *int64          `db:"resource_id" json:"resource_id,omitempty"`
    Details      json.RawMessage `db:"details" json:"details,omitempty"`
    IPAddress    *string         `db:"ip_address" json:"ip_address,omitempty"`
    UserAgent    *string         `db:"user_agent" json:"user_agent,omitempty"`
}
```

### 3.10 Config Model

```go
package models

import "time"

type Config struct {
    Key       string    `db:"key" json:"key"`
    Value     string    `db:"value" json:"value"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
```

### 3.11 Error Types

```go
package models

import "fmt"

type NotFoundError struct {
    Resource string
    ID       interface{}
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %v", e.Resource, e.ID)
}

type ConflictError struct {
    Resource string
    Field    string
    Value    interface{}
}

func (e *ConflictError) Error() string {
    return fmt.Sprintf("%s conflict on %s: %v", e.Resource, e.Field, e.Value)
}

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

type OptimisticLockError struct {
    Resource        string
    ID              int64
    ExpectedVersion int
    ActualVersion   int
}

func (e *OptimisticLockError) Error() string {
    return fmt.Sprintf("%s %d was modified (expected version %d, got %d)", 
        e.Resource, e.ID, e.ExpectedVersion, e.ActualVersion)
}
```

---

## 4. Interfaces

### 4.1 Database Interface

```go
package db

import (
    "context"
    "database/sql"
)

type Database interface {
    DB() *sql.DB
    Close() error
    Ping(ctx context.Context) error
    WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error
    Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
}
```

### 4.2 Repository Interface

```go
package db

import "context"

type Repository interface {
    Create(ctx context.Context, entity interface{}) error
    GetByID(ctx context.Context, id int64) (interface{}, error)
    Update(ctx context.Context, entity interface{}) error
    Delete(ctx context.Context, id int64) error
}
```

### 4.3 User Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id int64) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    GetByOIDCSubject(ctx context.Context, subject string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, limit, offset int) ([]*models.User, error)
    Count(ctx context.Context) (int, error)
    IsFirstUser(ctx context.Context) (bool, error)
    UpdateLastLogin(ctx context.Context, userID int64) error
}
```

### 4.4 Session Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type SessionRepository interface {
    Create(ctx context.Context, session *models.Session) error
    GetByID(ctx context.Context, id string) (*models.Session, error)
    GetByUserID(ctx context.Context, userID int64) ([]*models.Session, error)
    Update(ctx context.Context, session *models.Session) error
    Delete(ctx context.Context, id string) error
    DeleteByUserID(ctx context.Context, userID int64) error
    DeleteExpired(ctx context.Context) (int64, error)
    UpdateLastAccessed(ctx context.Context, id string) error
}
```

### 4.5 Event Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type EventRepository interface {
    Create(ctx context.Context, event *models.Event) error
    GetByID(ctx context.Context, id int64) (*models.Event, error)
    Update(ctx context.Context, event *models.Event) error
    UpdateWithVersion(ctx context.Context, event *models.Event, expectedVersion int) error
    Delete(ctx context.Context, id int64) error
    ListByCreator(ctx context.Context, creatorID int64, limit, offset int) ([]*models.Event, error)
    ListByStatus(ctx context.Context, status models.EventStatus, limit, offset int) ([]*models.Event, error)
    ListAll(ctx context.Context, limit, offset int) ([]*models.Event, error)
    Count(ctx context.Context) (int, error)
    CountByCreator(ctx context.Context, creatorID int64) (int, error)
    UpdateStatus(ctx context.Context, id int64, status models.EventStatus) error
    IncrementICSSequence(ctx context.Context, id int64) error
    GetEventsToArchive(ctx context.Context) ([]*models.Event, error)
}
```

### 4.6 Invite Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type InviteRepository interface {
    Create(ctx context.Context, invite *models.Invite) error
    CreateBatch(ctx context.Context, invites []*models.Invite) error
    GetByID(ctx context.Context, id int64) (*models.Invite, error)
    GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error)
    GetByEventID(ctx context.Context, eventID int64) ([]*models.Invite, error)
    GetByEmail(ctx context.Context, eventID int64, email string) (*models.Invite, error)
    Update(ctx context.Context, invite *models.Invite) error
    UpdateStatus(ctx context.Context, id int64, status models.InviteStatus) error
    UpdateTokenHash(ctx context.Context, id int64, tokenHash string) error
    Delete(ctx context.Context, id int64) error
    CountByEventID(ctx context.Context, eventID int64) (int, error)
    CountByStatus(ctx context.Context, eventID int64, status models.InviteStatus) (int, error)
    MarkViewed(ctx context.Context, id int64) error
    GetExpiredTokens(ctx context.Context) ([]*models.Invite, error)
}
```

### 4.7 RSVP Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type RSVPRepository interface {
    Create(ctx context.Context, rsvp *models.RSVP) error
    GetByID(ctx context.Context, id int64) (*models.RSVP, error)
    GetByInviteID(ctx context.Context, inviteID int64) (*models.RSVP, error)
    Update(ctx context.Context, rsvp *models.RSVP) error
    Delete(ctx context.Context, id int64) error
    GetByEventID(ctx context.Context, eventID int64) ([]*models.RSVP, error)
    CountByResponse(ctx context.Context, eventID int64, response models.RSVPResponse) (int, error)
    GetAttendeeCount(ctx context.Context, eventID int64) (int, error)
}
```

### 4.8 Question Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type QuestionRepository interface {
    Create(ctx context.Context, question *models.PreferenceQuestion) error
    GetByID(ctx context.Context, id int64) (*models.PreferenceQuestion, error)
    GetByEventID(ctx context.Context, eventID int64) ([]*models.PreferenceQuestion, error)
    Update(ctx context.Context, question *models.PreferenceQuestion) error
    Delete(ctx context.Context, id int64) error
    Reorder(ctx context.Context, eventID int64, questionIDs []int64) error
}

type AnswerRepository interface {
    Create(ctx context.Context, answer *models.RSVPAnswer) error
    CreateBatch(ctx context.Context, answers []*models.RSVPAnswer) error
    GetByRSVPID(ctx context.Context, rsvpID int64) ([]*models.RSVPAnswer, error)
    GetByQuestionID(ctx context.Context, questionID int64) ([]*models.RSVPAnswer, error)
    Update(ctx context.Context, answer *models.RSVPAnswer) error
    Delete(ctx context.Context, id int64) error
    DeleteByRSVPID(ctx context.Context, rsvpID int64) error
}
```

### 4.9 Email Queue Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type EmailQueueRepository interface {
    Create(ctx context.Context, email *models.EmailQueue) error
    GetByID(ctx context.Context, id int64) (*models.EmailQueue, error)
    GetPending(ctx context.Context, limit int) ([]*models.EmailQueue, error)
    Update(ctx context.Context, email *models.EmailQueue) error
    UpdateStatus(ctx context.Context, id int64, status models.EmailStatus) error
    IncrementAttempts(ctx context.Context, id int64, errorMsg string) error
    MarkSent(ctx context.Context, id int64) error
    MarkFailed(ctx context.Context, id int64, errorMsg string) error
    Delete(ctx context.Context, id int64) error
    GetByStatus(ctx context.Context, status models.EmailStatus, limit, offset int) ([]*models.EmailQueue, error)
    CountByStatus(ctx context.Context, status models.EmailStatus) (int, error)
}
```

### 4.10 Template Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type TemplateRepository interface {
    Create(ctx context.Context, template *models.Template) error
    GetByID(ctx context.Context, id int64) (*models.Template, error)
    GetByType(ctx context.Context, templateType models.TemplateType) ([]*models.Template, error)
    GetDefault(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
    Update(ctx context.Context, template *models.Template) error
    Delete(ctx context.Context, id int64) error
    SetDefault(ctx context.Context, id int64) error
    List(ctx context.Context, limit, offset int) ([]*models.Template, error)
}
```

### 4.11 Audit Log Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
    "time"
)

type AuditLogRepository interface {
    Create(ctx context.Context, log *models.AuditLog) error
    GetByID(ctx context.Context, id int64) (*models.AuditLog, error)
    GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.AuditLog, error)
    GetByResource(ctx context.Context, resourceType string, resourceID int64, limit, offset int) ([]*models.AuditLog, error)
    GetRecent(ctx context.Context, limit int) ([]*models.AuditLog, error)
    DeleteOlderThan(ctx context.Context, timestamp time.Time) (int64, error)
}
```

### 4.12 Config Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type ConfigRepository interface {
    Get(ctx context.Context, key string) (*models.Config, error)
    Set(ctx context.Context, key, value string) error
    Delete(ctx context.Context, key string) error
    GetAll(ctx context.Context) ([]*models.Config, error)
    GetHMACSecret(ctx context.Context) ([]byte, error)
    SetHMACSecret(ctx context.Context, secret []byte) error
}
```

### 4.13 Migrator Interface

```go
package db

import "context"

type Migrator interface {
    Up(ctx context.Context) error
    Down(ctx context.Context) error
    Version(ctx context.Context) (uint, bool, error)
    Steps(ctx context.Context, n int) error
}
```

---

## 5. Implementation Details

### 5.1 Database Connection

```go
package db

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
)

type Config struct {
    Type         string
    Path         string
    MaxOpenConns int
    MaxIdleConns int
    MaxLifetime  time.Duration
}

type database struct {
    db *sql.DB
}

func NewDatabase(cfg Config) (Database, error) {
    var dsn string
    
    switch cfg.Type {
    case "sqlite":
        dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000", cfg.Path)
    default:
        return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
    }
    
    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.MaxLifetime)
    
    if err := db.PingContext(context.Background()); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return &database{db: db}, nil
}

func (d *database) DB() *sql.DB {
    return d.db
}

func (d *database) Close() error {
    return d.db.Close()
}

func (d *database) Ping(ctx context.Context) error {
    return d.db.PingContext(ctx)
}

func (d *database) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
    tx, err := d.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()
    
    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
        }
        return err
    }
    
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return nil
}

func (d *database) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    return d.db.ExecContext(ctx, query, args...)
}

func (d *database) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    return d.db.QueryContext(ctx, query, args...)
}

func (d *database) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
    return d.db.QueryRowContext(ctx, query, args...)
}
```

### 5.2 Migration Execution

```go
package db

import (
    "context"
    "database/sql"
    "fmt"
    
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

type migrator struct {
    migrate *migrate.Migrate
}

func NewMigrator(db *sql.DB, migrationsPath string) (Migrator, error) {
    driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to create migration driver: %w", err)
    }
    
    m, err := migrate.NewWithDatabaseInstance(
        fmt.Sprintf("file://%s", migrationsPath),
        "sqlite3",
        driver,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create migrator: %w", err)
    }
    
    return &migrator{migrate: m}, nil
}

func (m *migrator) Up(ctx context.Context) error {
    if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("migration up failed: %w", err)
    }
    return nil
}

func (m *migrator) Down(ctx context.Context) error {
    if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("migration down failed: %w", err)
    }
    return nil
}

func (m *migrator) Version(ctx context.Context) (uint, bool, error) {
    return m.migrate.Version()
}

func (m *migrator) Steps(ctx context.Context, n int) error {
    if err := m.migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("migration steps failed: %w", err)
    }
    return nil
}
```

---

## 6. Dependencies

### 6.1 External Libraries

```go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)
```

### 6.2 Internal Dependencies

- None (foundation layer)

### 6.3 Dependents

- Domain 1 (Auth) - User and session storage
- Domain 2 (Event) - Event persistence
- Domain 3 (Invite) - Invite persistence
- Domain 4 (RSVP) - RSVP and answer persistence
- Domain 5 (Email) - Email queue storage
- Domain 6 (Template) - Template storage

---

## 7. Validation & Error Handling

### 7.1 Centralized Validation Rules

**See HLD Section 14 for complete validation specifications**

**Event Validation:**
- Title: 3-200 characters
- Description: Max 5000 characters
- Location: Max 500 characters
- Start time: Must be future, valid ISO 8601
- End time: Must be after start time, within 7 days
- Timezone: Valid IANA timezone
- RSVP deadline: Before start time, in future
- Max plus ones: 0-10

**Invite Validation:**
- Name: Max 100 characters
- Email: Valid format, max 255 characters
- Max plus ones: 0-10, <= event.max_plus_ones

**RSVP Validation:**
- Response: Must be yes/no/maybe
- Plus ones: 0 to invite.max_plus_ones
- Deadline: Must be before event.rsvp_deadline

**Question Validation:**
- Question text: 5-500 characters
- Question type: text/select/boolean
- Options: 2-20 for select type

**Answer Validation:**
- Text: Max 500 characters
- Select: Must match option value
- Boolean: Must be true/false
- Required: Must have answer

### 7.2 Error Type Mapping

**Database Errors → Domain Errors:**
```go
func mapDBError(err error) error {
    if errors.Is(err, sql.ErrNoRows) {
        return &models.NotFoundError{}
    }
    if isUniqueConstraintError(err) {
        return &models.ConflictError{}
    }
    return err
}
```

**Domain Errors → HTTP Status:**
- `NotFoundError` → 404
- `ValidationError` → 400
- `ConflictError` → 409
- `OptimisticLockError` → 409
- `PermissionError` → 403
- Other errors → 500

### 7.3 Security Checklist

**Per HLD Section 16:**

**Transport Security:**
- ✅ HTTPS required (Secure cookie flag)
- ✅ HSTS header (31536000 seconds)
- ✅ Trusted proxy IP validation

**Session Security:**
- ✅ HttpOnly cookies
- ✅ SameSite=Lax
- ✅ 32-byte random session IDs
- ✅ 7-day expiration

**Token Security:**
- ✅ 256-bit cryptographically secure tokens
- ✅ HMAC-SHA256 hashing
- ✅ Constant-time comparison
- ✅ Tokens never logged

**Input Security:**
- ✅ Parameterized queries (SQL injection prevention)
- ✅ html/template auto-escaping (XSS prevention)
- ✅ Path traversal prevention
- ✅ File upload validation

**CSRF Protection:**
- ✅ CSRF tokens per session
- ✅ Validated on POST/PUT/DELETE
- ✅ SameSite cookies

**Secrets Management:**
- ✅ Never log secrets
- ✅ Never return in API responses
- ✅ Masked in admin UI
- ✅ Rotation support

## 8. Testing Strategy

### 8.1 Unit Tests

**Test Approach:**
```go
func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    tests := []struct {
        name    string
        user    *models.User
        wantErr bool
    }{
        {
            name: "valid user",
            user: &models.User{
                Email: "test@example.com",
                Name:  "Test User",
                Role:  models.RoleEventManager,
            },
            wantErr: false,
        },
        {
            name: "duplicate email",
            user: &models.User{
                Email: "test@example.com",
                Name:  "Another User",
                Role:  models.RoleEventManager,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.Create(context.Background(), tt.user)
            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 7.2 Test Database Setup

```go
func setupTestDB(t *testing.T) Database {
    t.Helper()
    
    db, err := NewDatabase(Config{
        Type: "sqlite",
        Path: ":memory:",
        MaxOpenConns: 1,
        MaxIdleConns: 1,
    })
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    
    migrator, err := NewMigrator(db.DB(), "../../migrations/sqlite")
    if err != nil {
        t.Fatalf("Failed to create migrator: %v", err)
    }
    
    if err := migrator.Up(context.Background()); err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }
    
    return db
}
```

---

## 8. Security Considerations

### 8.1 SQL Injection Prevention

**Always Use Parameterized Queries:**
```go
query := `SELECT * FROM users WHERE email = ?`
rows, err := db.Query(ctx, query, email)
```

**Never Concatenate User Input:**
```go
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
```

### 8.2 Sensitive Data Handling

**Never Log:**
- Token hashes
- HMAC secrets
- Session IDs
- Passwords

---

## 9. Performance Considerations

### 9.1 Connection Pooling

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 9.2 Indexes

All indexes defined in migration files for optimal query performance.

---

## 10. Error Scenarios

### 10.1 Connection Errors

**Handling:** Fail fast on startup with clear error message

### 10.2 Transaction Errors

**Handling:** Retry deadlocks up to 3 times with exponential backoff

### 10.3 Constraint Violations

**Handling:** Convert to domain errors and return to caller

---

## 11. Examples

### 11.1 Basic CRUD

```go
func ExampleUserRepository() {
    db, _ := NewDatabase(Config{Type: "sqlite", Path: "test.db"})
    repo := NewUserRepository(db)
    ctx := context.Background()
    
    user := &models.User{
        Email: "john@example.com",
        Name:  "John Doe",
        Role:  models.RoleEventManager,
    }
    
    if err := repo.Create(ctx, user); err != nil {
        log.Fatal(err)
    }
}
```

---

## 12. Open Questions

**None** - All design decisions finalized in HLD

---

**Document Status:** ✅ Complete and Ready for Implementation