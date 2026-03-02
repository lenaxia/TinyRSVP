# User Story: Template Data Structure

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-09

---

## User Story

As a **developer**, I want **a strongly-typed template data structure** so that **templates can be stored, validated, and managed consistently throughout the application**.

---

## Acceptance Criteria

- [x] Template struct defined with all required fields
- [x] Template type enum created (invite_email, rsvp_page, confirmation_page)
- [x] Template repository interface defined
- [x] Database schema supports template storage
- [x] Template validation rules implemented
- [x] All fields properly typed (no map[string]interface{})
- [x] Created/updated timestamps tracked
- [x] Owner/event association supported
- [x] All tests pass with timeout
- [x] Documentation complete

---

## Technical Details

### Template Model

```go
package models

import (
    "time"
)

type TemplateType string

const (
    TemplateTypeInviteEmail      TemplateType = "invite_email"
    TemplateTypeRSVPPage         TemplateType = "rsvp_page"
    TemplateTypeConfirmationPage TemplateType = "confirmation_page"
)

type Template struct {
    ID          int64        `json:"id"`
    EventID     *int64       `json:"event_id,omitempty"`
    Name        string       `json:"name"`
    Type        TemplateType `json:"type"`
    Description string       `json:"description"`
    
    // Template Content
    HTMLContent string  `json:"html_content"`
    TextContent *string `json:"text_content,omitempty"`
    CSSContent  *string `json:"css_content,omitempty"`
    
    // Metadata
    IsDefault   bool      `json:"is_default"`
    IsActive    bool      `json:"is_active"`
    Version     int       `json:"version"`
    CreatedBy   int64     `json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func (t *Template) Validate() error {
    if t.Name == "" {
        return &ValidationError{Field: "name", Message: "Template name is required"}
    }
    
    if len(t.Name) < 3 || len(t.Name) > 100 {
        return &ValidationError{Field: "name", Message: "Template name must be 3-100 characters"}
    }
    
    if !t.Type.IsValid() {
        return &ValidationError{Field: "type", Message: "Invalid template type"}
    }
    
    if t.HTMLContent == "" {
        return &ValidationError{Field: "html_content", Message: "HTML content is required"}
    }
    
    if t.Type == TemplateTypeInviteEmail && t.TextContent == nil {
        return &ValidationError{Field: "text_content", Message: "Text content required for email templates"}
    }
    
    return nil
}

func (tt TemplateType) IsValid() bool {
    switch tt {
    case TemplateTypeInviteEmail, TemplateTypeRSVPPage, TemplateTypeConfirmationPage:
        return true
    default:
        return false
    }
}

func (tt TemplateType) String() string {
    return string(tt)
}
```

### Repository Interface

```go
package repositories

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type TemplateRepository interface {
    Create(ctx context.Context, template *models.Template) error
    GetByID(ctx context.Context, id int64) (*models.Template, error)
    GetByEventAndType(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error)
    GetDefaultByType(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
    List(ctx context.Context, filters *TemplateFilters) ([]*models.Template, error)
    Update(ctx context.Context, template *models.Template) error
    Delete(ctx context.Context, id int64) error
    SetActive(ctx context.Context, id int64, active bool) error
}

type TemplateFilters struct {
    EventID   *int64
    Type      *models.TemplateType
    IsDefault *bool
    IsActive  *bool
    CreatedBy *int64
    Limit     int
    Offset    int
}
```

---

## Database Schema

```sql
CREATE TABLE templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('invite_email', 'rsvp_page', 'confirmation_page')),
    description TEXT,
    
    html_content TEXT NOT NULL,
    text_content TEXT,
    css_content TEXT,
    
    is_default BOOLEAN NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    version INTEGER NOT NULL DEFAULT 1,
    
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX idx_templates_event_type ON templates(event_id, type);
CREATE INDEX idx_templates_type_default ON templates(type, is_default);
CREATE INDEX idx_templates_active ON templates(is_active);
CREATE INDEX idx_templates_created_by ON templates(created_by);
```

---

## Tasks

### Phase 1: Model Definition (TDD)
- [ ] Define TemplateType enum
- [ ] Write test for TemplateType.IsValid()
- [ ] Implement TemplateType validation
- [ ] Define Template struct
- [ ] Write test for Template.Validate()
- [ ] Implement Template validation
- [ ] Run tests (should pass)

### Phase 2: Repository Interface (TDD)
- [ ] Define TemplateRepository interface
- [ ] Define TemplateFilters struct
- [ ] Write test for Create operation
- [ ] Write test for GetByID operation
- [ ] Write test for GetByEventAndType operation
- [ ] Write test for GetDefaultByType operation
- [ ] Write test for List with filters
- [ ] Write test for Update operation
- [ ] Write test for Delete operation
- [ ] Write test for SetActive operation
- [ ] Implement repository methods
- [ ] Run tests (should pass)

### Phase 3: Database Migration
- [ ] Create migration file
- [ ] Write up migration SQL
- [ ] Write down migration SQL
- [ ] Test migration up
- [ ] Test migration down
- [ ] Verify schema created correctly

### Phase 4: Integration Testing
- [ ] Test template CRUD operations
- [ ] Test event association
- [ ] Test default template retrieval
- [ ] Test filtering by type
- [ ] Test filtering by active status
- [ ] Test concurrent operations
- [ ] Verify foreign key constraints

---

## Validation Rules

### Template Name
- Required field
- 3-100 characters
- Alphanumeric, spaces, hyphens allowed
- Must be unique per event

### Template Type
- Required field
- Must be one of: invite_email, rsvp_page, confirmation_page
- Cannot be changed after creation

### HTML Content
- Required field
- Must be valid Go template syntax
- Maximum 100KB size
- No script tags allowed

### Text Content
- Required for email templates
- Optional for page templates
- Plain text only
- Maximum 50KB size

### CSS Content
- Optional field
- Must pass CSS sanitization
- Maximum 50KB size
- No external imports allowed

### Event Association
- Optional (null for default templates)
- Must reference valid event
- Cascade delete when event deleted

---

## Business Logic

### Template Selection Priority

```go
func (r *repository) GetTemplateForEvent(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
    // 1. Try event-specific active template
    template, err := r.GetByEventAndType(ctx, eventID, templateType)
    if err == nil && template.IsActive {
        return template, nil
    }
    
    // 2. Fall back to default template
    return r.GetDefaultByType(ctx, templateType)
}
```

### Version Management

```go
func (r *repository) Update(ctx context.Context, template *models.Template) error {
    // Increment version on update
    template.Version++
    template.UpdatedAt = time.Now()
    
    return r.db.ExecContext(ctx, `
        UPDATE templates 
        SET html_content = ?, text_content = ?, css_content = ?,
            version = ?, updated_at = ?
        WHERE id = ? AND version = ?
    `, template.HTMLContent, template.TextContent, template.CSSContent,
       template.Version, template.UpdatedAt, template.ID, template.Version-1)
}
```

---

## Error Handling

| Error Condition | Error Type | Message |
|----------------|------------|---------|
| Invalid template type | `ValidationError` | "Invalid template type" |
| Missing name | `ValidationError` | "Template name is required" |
| Name too short/long | `ValidationError` | "Template name must be 3-100 characters" |
| Missing HTML content | `ValidationError` | "HTML content is required" |
| Missing text content (email) | `ValidationError` | "Text content required for email templates" |
| Invalid event ID | `ForeignKeyError` | "Event not found" |
| Template not found | `NotFoundError` | "Template not found" |
| Duplicate name | `ConflictError` | "Template name already exists for this event" |

---

## Testing Strategy

### Unit Tests

```go
func TestTemplate_Validate(t *testing.T) {
    tests := []struct {
        name     string
        template *models.Template
        wantErr  bool
        errField string
    }{
        {
            name: "valid invite email template",
            template: &models.Template{
                Name:        "Custom Invite",
                Type:        models.TemplateTypeInviteEmail,
                HTMLContent: "<html>{{.Event.Title}}</html>",
                TextContent: strPtr("Event: {{.Event.Title}}"),
            },
            wantErr: false,
        },
        {
            name: "missing name",
            template: &models.Template{
                Type:        models.TemplateTypeRSVPPage,
                HTMLContent: "<html></html>",
            },
            wantErr:  true,
            errField: "name",
        },
        {
            name: "invalid type",
            template: &models.Template{
                Name:        "Test",
                Type:        "invalid_type",
                HTMLContent: "<html></html>",
            },
            wantErr:  true,
            errField: "type",
        },
        {
            name: "email missing text content",
            template: &models.Template{
                Name:        "Email Template",
                Type:        models.TemplateTypeInviteEmail,
                HTMLContent: "<html></html>",
                TextContent: nil,
            },
            wantErr:  true,
            errField: "text_content",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.template.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && tt.errField != "" {
                var valErr *models.ValidationError
                if !errors.As(err, &valErr) || valErr.Field != tt.errField {
                    t.Errorf("Expected ValidationError for field %s", tt.errField)
                }
            }
        })
    }
}

func TestTemplateType_IsValid(t *testing.T) {
    tests := []struct {
        name string
        tt   models.TemplateType
        want bool
    }{
        {"invite email", models.TemplateTypeInviteEmail, true},
        {"rsvp page", models.TemplateTypeRSVPPage, true},
        {"confirmation page", models.TemplateTypeConfirmationPage, true},
        {"invalid", "invalid_type", false},
        {"empty", "", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.tt.IsValid(); got != tt.want {
                t.Errorf("IsValid() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

```go
func TestTemplateRepository_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewTemplateRepository(db)
    
    // Create test user and event
    user := createTestUser(t, db)
    event := createTestEvent(t, db, user.ID)
    
    // Test Create
    template := &models.Template{
        EventID:     &event.ID,
        Name:        "Custom Invite",
        Type:        models.TemplateTypeInviteEmail,
        HTMLContent: "<html>{{.Event.Title}}</html>",
        TextContent: strPtr("Event: {{.Event.Title}}"),
        IsDefault:   false,
        IsActive:    true,
        Version:     1,
        CreatedBy:   user.ID,
    }
    
    err := repo.Create(context.Background(), template)
    if err != nil {
        t.Fatalf("Create() error = %v", err)
    }
    
    if template.ID == 0 {
        t.Error("Expected template ID to be set")
    }
    
    // Test GetByID
    retrieved, err := repo.GetByID(context.Background(), template.ID)
    if err != nil {
        t.Fatalf("GetByID() error = %v", err)
    }
    
    if retrieved.Name != template.Name {
        t.Errorf("Name = %s, want %s", retrieved.Name, template.Name)
    }
    
    // Test GetByEventAndType
    eventTemplate, err := repo.GetByEventAndType(context.Background(), event.ID, models.TemplateTypeInviteEmail)
    if err != nil {
        t.Fatalf("GetByEventAndType() error = %v", err)
    }
    
    if eventTemplate.ID != template.ID {
        t.Error("Expected to retrieve the same template")
    }
    
    // Test Update
    template.HTMLContent = "<html>Updated</html>"
    err = repo.Update(context.Background(), template)
    if err != nil {
        t.Fatalf("Update() error = %v", err)
    }
    
    if template.Version != 2 {
        t.Errorf("Version = %d, want 2", template.Version)
    }
    
    // Test Delete
    err = repo.Delete(context.Background(), template.ID)
    if err != nil {
        t.Fatalf("Delete() error = %v", err)
    }
    
    _, err = repo.GetByID(context.Background(), template.ID)
    if err == nil {
        t.Error("Expected NotFoundError after delete")
    }
}
```

---

## Dependencies

**Depends on:**
- Epic 00: Foundation (database layer)
- Epic 01: Auth (user management)

**Blocks:**
- Story 01: Template Integration
- Story 02: Template Security
- Story 04: Template CRUD

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Template struct defined and validated
- [x] TemplateType enum implemented
- [x] Repository interface defined
- [x] Database migration created
- [x] All unit tests passing (93.4% coverage for models)
- [x] Integration tests passing
- [x] Documentation complete
- [x] Code reviewed
- [x] No linter warnings

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11 (Templates)
- **Database:** templates table schema
- **Models:** [`internal/models/template.go`](../../internal/models/template.go)
