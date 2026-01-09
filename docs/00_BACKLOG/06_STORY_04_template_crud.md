# User Story: Template CRUD Operations

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1.5 days
**Completed:** 2026-01-09

---

## User Story

As an **event manager**, I want **to create, edit, and delete custom templates** so that **I can customize the appearance of invitations and RSVP pages for my events**.

---

## Acceptance Criteria

- [x] Event managers can create custom templates
- [x] Event managers can edit their own templates
- [x] Event managers can delete their own templates
- [x] Event managers can list their templates
- [x] Event managers can set template as active/inactive
- [x] Admins can manage all templates
- [x] Template validation enforced on create/update
- [x] Cannot delete template if in use by events
- [x] Cannot delete default system templates
- [x] All tests pass with timeout
- [x] RBAC permissions enforced

---

## Technical Details

### Service Interface

```go
package templates

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type Service interface {
    CreateTemplate(ctx context.Context, template *models.Template) error
    GetTemplate(ctx context.Context, id int64) (*models.Template, error)
    GetTemplateForEvent(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error)
    GetDefaultTemplate(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
    UpdateTemplate(ctx context.Context, template *models.Template) error
    DeleteTemplate(ctx context.Context, id int64) error
    SetActive(ctx context.Context, id int64, active bool) error
    SetDefault(ctx context.Context, id int64) error
    ListTemplates(ctx context.Context, filters *TemplateFilters) ([]*models.Template, error)
}

type service struct {
    repo      repositories.TemplateRepository
    validator Validator
    renderer  Renderer
}

func NewService(repo repositories.TemplateRepository, validator Validator, renderer Renderer) Service {
    return &service{
        repo:      repo,
        validator: validator,
        renderer:  renderer,
    }
}
```

### Create Template

```go
func (s *service) CreateTemplate(ctx context.Context, template *models.Template) error {
    if err := s.validator.ValidateTemplate(template); err != nil {
        return err
    }
    
    userID := auth.GetUserID(ctx)
    if userID == 0 {
        return &models.UnauthorizedError{Message: "Authentication required"}
    }
    
    template.CreatedBy = userID
    template.CreatedAt = time.Now()
    template.UpdatedAt = time.Now()
    template.Version = 1
    template.IsActive = true
    
    if err := s.repo.Create(ctx, template); err != nil {
        return fmt.Errorf("failed to create template: %w", err)
    }
    
    return nil
}
```

### Update Template

```go
func (s *service) UpdateTemplate(ctx context.Context, template *models.Template) error {
    existing, err := s.repo.GetByID(ctx, template.ID)
    if err != nil {
        return err
    }
    
    userID := auth.GetUserID(ctx)
    role := auth.GetUserRole(ctx)
    
    if role != models.RoleAdmin && existing.CreatedBy != userID {
        return &models.ForbiddenError{Message: "You can only edit your own templates"}
    }
    
    if existing.IsDefault && role != models.RoleAdmin {
        return &models.ForbiddenError{Message: "Only admins can edit default templates"}
    }
    
    if err := s.validator.ValidateTemplate(template); err != nil {
        return err
    }
    
    template.UpdatedAt = time.Now()
    template.Version = existing.Version + 1
    
    if err := s.repo.Update(ctx, template); err != nil {
        return fmt.Errorf("failed to update template: %w", err)
    }
    
    return nil
}
```

### Delete Template

```go
func (s *service) DeleteTemplate(ctx context.Context, id int64) error {
    template, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }
    
    userID := auth.GetUserID(ctx)
    role := auth.GetUserRole(ctx)
    
    if role != models.RoleAdmin && template.CreatedBy != userID {
        return &models.ForbiddenError{Message: "You can only delete your own templates"}
    }
    
    if template.IsDefault {
        return &models.ValidationError{
            Field:   "template",
            Message: "Cannot delete default system templates",
        }
    }
    
    inUse, err := s.repo.IsTemplateInUse(ctx, id)
    if err != nil {
        return err
    }
    
    if inUse {
        return &models.ValidationError{
            Field:   "template",
            Message: "Cannot delete template that is in use by events",
        }
    }
    
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("failed to delete template: %w", err)
    }
    
    return nil
}
```

### Get Template for Event

```go
func (s *service) GetTemplateForEvent(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
    template, err := s.repo.GetByEventAndType(ctx, eventID, templateType)
    if err == nil && template.IsActive {
        return template, nil
    }
    
    return s.repo.GetDefaultByType(ctx, templateType)
}
```

---

## API Endpoints

### Create Template
```
POST /api/templates
Authorization: Required (Event Manager or Admin)
Content-Type: application/json

{
    "name": "Custom Invite",
    "type": "invite_email",
    "description": "Custom invitation template",
    "html_content": "<html>...</html>",
    "text_content": "...",
    "css_content": "...",
    "event_id": 123
}

Response 201 Created:
{
    "template": {
        "id": 1,
        "name": "Custom Invite",
        "type": "invite_email",
        "is_default": false,
        "is_active": true,
        "created_at": "2026-01-09T00:00:00Z"
    }
}
```

### Get Template
```
GET /api/templates/:id
Authorization: Required (Event Manager or Admin)

Response 200 OK:
{
    "template": {
        "id": 1,
        "name": "Custom Invite",
        "type": "invite_email",
        "html_content": "<html>...</html>",
        "text_content": "...",
        "css_content": "...",
        "is_default": false,
        "is_active": true,
        "created_by": 1,
        "created_at": "2026-01-09T00:00:00Z",
        "updated_at": "2026-01-09T00:00:00Z"
    }
}
```

### Update Template
```
PUT /api/templates/:id
Authorization: Required (Owner or Admin)
Content-Type: application/json

{
    "name": "Updated Template",
    "html_content": "<html>...</html>",
    "text_content": "...",
    "css_content": "..."
}

Response 200 OK:
{
    "template": {
        "id": 1,
        "version": 2,
        "updated_at": "2026-01-09T01:00:00Z"
    }
}
```

### Delete Template
```
DELETE /api/templates/:id
Authorization: Required (Owner or Admin)

Response 204 No Content
```

### List Templates
```
GET /api/templates?type=invite_email&event_id=123
Authorization: Required (Event Manager or Admin)

Response 200 OK:
{
    "templates": [
        {
            "id": 1,
            "name": "Custom Invite",
            "type": "invite_email",
            "is_default": false,
            "is_active": true,
            "created_by": 1,
            "created_at": "2026-01-09T00:00:00Z"
        }
    ],
    "total": 1
}
```

---

## Tasks

### Phase 1: Service Layer (TDD)
- [ ] Define Service interface
- [ ] Write test for CreateTemplate success
- [ ] Write test for CreateTemplate validation error
- [ ] Write test for CreateTemplate unauthorized
- [ ] Write test for UpdateTemplate success
- [ ] Write test for UpdateTemplate not owner
- [ ] Write test for UpdateTemplate default template
- [ ] Write test for DeleteTemplate success
- [ ] Write test for DeleteTemplate in use
- [ ] Write test for DeleteTemplate default template
- [ ] Write test for GetTemplateForEvent
- [ ] Implement all service methods
- [ ] Run tests (should pass)

### Phase 2: Handler Layer (TDD)
- [ ] Create template handlers
- [ ] Write test for POST /api/templates
- [ ] Write test for GET /api/templates/:id
- [ ] Write test for PUT /api/templates/:id
- [ ] Write test for DELETE /api/templates/:id
- [ ] Write test for GET /api/templates (list)
- [ ] Write test for RBAC enforcement
- [ ] Implement all handlers
- [ ] Run tests (should pass)

### Phase 3: Repository Methods (TDD)
- [ ] Write test for IsTemplateInUse
- [ ] Write test for GetByEventAndType
- [ ] Write test for SetActive
- [ ] Write test for SetDefault
- [ ] Implement repository methods
- [ ] Run tests (should pass)

### Phase 4: Integration Testing
- [ ] Test full create flow
- [ ] Test full update flow
- [ ] Test full delete flow
- [ ] Test permission enforcement
- [ ] Test template selection for events
- [ ] Test concurrent operations

---

## Validation Rules

### Create Template
- User must be authenticated
- Template must pass validation
- Name must be unique per user
- Type must be valid
- HTML content required
- Text content required for email templates

### Update Template
- User must own template or be admin
- Cannot edit default templates (admin only)
- Template must pass validation
- Version must match (optimistic locking)

### Delete Template
- User must own template or be admin
- Cannot delete default templates
- Cannot delete if in use by events
- Cascade delete associated assets

### Set Default
- Admin only
- Only one default per type
- Previous default becomes non-default

---

## Error Handling

| Error Condition | Error Type | HTTP Status | Message |
|----------------|------------|-------------|---------|
| Unauthorized | `UnauthorizedError` | 401 | "Authentication required" |
| Not owner | `ForbiddenError` | 403 | "You can only edit your own templates" |
| Edit default (non-admin) | `ForbiddenError` | 403 | "Only admins can edit default templates" |
| Delete default | `ValidationError` | 400 | "Cannot delete default system templates" |
| Delete in use | `ValidationError` | 400 | "Cannot delete template in use by events" |
| Template not found | `NotFoundError` | 404 | "Template not found" |
| Validation error | `ValidationError` | 400 | Field-specific message |
| Version conflict | `ConflictError` | 409 | "Template was modified. Please refresh and try again" |

---

## Testing Strategy

### Unit Tests

```go
func TestTemplateService_CreateTemplate(t *testing.T) {
    tests := []struct {
        name     string
        template *models.Template
        userID   int64
        wantErr  bool
        errType  error
    }{
        {
            name: "valid template",
            template: &models.Template{
                Name:        "Custom Invite",
                Type:        models.TemplateTypeInviteEmail,
                HTMLContent: "<h1>{{.Event.Title}}</h1>",
                TextContent: strPtr("{{.Event.Title}}"),
            },
            userID:  1,
            wantErr: false,
        },
        {
            name: "unauthorized",
            template: &models.Template{
                Name:        "Custom Invite",
                Type:        models.TemplateTypeInviteEmail,
                HTMLContent: "<h1>{{.Event.Title}}</h1>",
            },
            userID:  0,
            wantErr: true,
            errType: &models.UnauthorizedError{},
        },
        {
            name: "validation error",
            template: &models.Template{
                Name:        "",
                Type:        models.TemplateTypeInviteEmail,
                HTMLContent: "<h1>{{.Event.Title}}</h1>",
            },
            userID:  1,
            wantErr: true,
            errType: &models.ValidationError{},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := auth.WithUserID(context.Background(), tt.userID)
            
            mockRepo := &mocks.MockTemplateRepository{}
            mockValidator := &mocks.MockValidator{
                ValidateTemplateFunc: func(tmpl *models.Template) error {
                    return tmpl.Validate()
                },
            }
            mockRenderer := &mocks.MockRenderer{}
            
            service := NewService(mockRepo, mockValidator, mockRenderer)
            
            err := service.CreateTemplate(ctx, tt.template)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateTemplate() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if tt.wantErr && tt.errType != nil {
                if !errors.As(err, &tt.errType) {
                    t.Errorf("Error type = %T, want %T", err, tt.errType)
                }
            }
        })
    }
}

func TestTemplateService_DeleteTemplate(t *testing.T) {
    tests := []struct {
        name       string
        templateID int64
        userID     int64
        role       models.Role
        template   *models.Template
        inUse      bool
        wantErr    bool
        errMsg     string
    }{
        {
            name:       "delete own template",
            templateID: 1,
            userID:     1,
            role:       models.RoleEventManager,
            template: &models.Template{
                ID:        1,
                CreatedBy: 1,
                IsDefault: false,
            },
            inUse:   false,
            wantErr: false,
        },
        {
            name:       "delete default template",
            templateID: 1,
            userID:     1,
            role:       models.RoleAdmin,
            template: &models.Template{
                ID:        1,
                IsDefault: true,
            },
            wantErr: true,
            errMsg:  "Cannot delete default",
        },
        {
            name:       "delete template in use",
            templateID: 1,
            userID:     1,
            role:       models.RoleEventManager,
            template: &models.Template{
                ID:        1,
                CreatedBy: 1,
                IsDefault: false,
            },
            inUse:   true,
            wantErr: true,
            errMsg:  "in use",
        },
        {
            name:       "delete other user's template",
            templateID: 1,
            userID:     2,
            role:       models.RoleEventManager,
            template: &models.Template{
                ID:        1,
                CreatedBy: 1,
                IsDefault: false,
            },
            wantErr: true,
            errMsg:  "only delete your own",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := auth.WithUserID(context.Background(), tt.userID)
            ctx = auth.WithUserRole(ctx, tt.role)
            
            mockRepo := &mocks.MockTemplateRepository{
                GetByIDFunc: func(ctx context.Context, id int64) (*models.Template, error) {
                    return tt.template, nil
                },
                IsTemplateInUseFunc: func(ctx context.Context, id int64) (bool, error) {
                    return tt.inUse, nil
                },
                DeleteFunc: func(ctx context.Context, id int64) error {
                    return nil
                },
            }
            
            service := NewService(mockRepo, nil, nil)
            
            err := service.DeleteTemplate(ctx, tt.templateID)
            if (err != nil) != tt.wantErr {
                t.Errorf("DeleteTemplate() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if tt.wantErr && tt.errMsg != "" {
                if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("Error message = %v, want to contain %v", err, tt.errMsg)
                }
            }
        })
    }
}
```

### Integration Tests

```go
func TestTemplateService_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewTemplateRepository(db)
    validator := NewValidator(NewRenderer())
    renderer := NewRenderer()
    service := NewService(repo, validator, renderer)
    
    user := createTestUser(t, db, models.RoleEventManager)
    ctx := auth.WithUserID(context.Background(), user.ID)
    ctx = auth.WithUserRole(ctx, user.Role)
    
    template := &models.Template{
        Name:        "Custom Invite",
        Type:        models.TemplateTypeInviteEmail,
        Description: "My custom template",
        HTMLContent: "<h1>{{.Event.Title}}</h1>",
        TextContent: strPtr("{{.Event.Title}}"),
    }
    
    err := service.CreateTemplate(ctx, template)
    if err != nil {
        t.Fatalf("CreateTemplate() error = %v", err)
    }
    
    if template.ID == 0 {
        t.Error("Expected template ID to be set")
    }
    
    retrieved, err := service.GetTemplate(ctx, template.ID)
    if err != nil {
        t.Fatalf("GetTemplate() error = %v", err)
    }
    
    if retrieved.Name != template.Name {
        t.Errorf("Name = %s, want %s", retrieved.Name, template.Name)
    }
    
    template.HTMLContent = "<h1>Updated</h1>"
    err = service.UpdateTemplate(ctx, template)
    if err != nil {
        t.Fatalf("UpdateTemplate() error = %v", err)
    }
    
    if template.Version != 2 {
        t.Errorf("Version = %d, want 2", template.Version)
    }
    
    err = service.DeleteTemplate(ctx, template.ID)
    if err != nil {
        t.Fatalf("DeleteTemplate() error = %v", err)
    }
    
    _, err = service.GetTemplate(ctx, template.ID)
    if err == nil {
        t.Error("Expected NotFoundError after delete")
    }
}
```

---

## Tasks

### Phase 1: Service Layer (TDD)
- [x] Define Service interface
- [x] Write test for CreateTemplate
- [x] Write test for GetTemplate
- [x] Write test for UpdateTemplate
- [x] Write test for DeleteTemplate
- [x] Write test for SetActive
- [x] Write test for SetDefault
- [x] Write test for ListTemplates
- [x] Write test for GetTemplateForEvent
- [x] Implement all service methods
- [x] Run tests (should pass)

### Phase 2: Handler Layer (TDD)
- [x] Create template handlers
- [x] Write test for POST handler
- [x] Write test for GET handler
- [x] Write test for PUT handler
- [x] Write test for DELETE handler
- [x] Write test for LIST handler
- [x] Write test for RBAC enforcement
- [x] Implement all handlers
- [x] Run tests (should pass)

### Phase 3: Repository Extensions (TDD)
- [x] Write test for IsTemplateInUse
- [x] Write test for SetDefault
- [x] Implement repository methods
- [x] Run tests (should pass)

### Phase 4: Integration Testing
- [x] Test full CRUD flow
- [x] Test permission enforcement
- [x] Test template selection logic
- [x] Test concurrent updates
- [x] Test version conflicts

---

## Dependencies

**Depends on:**
- Story 00: Template Struct (for data models)
- Story 01: Template Integration (for renderer)
- Story 02: Template Security (for validation)
- Story 03: Default Templates (for system templates)
- Epic 01: Auth (for RBAC)

**Blocks:**
- Story 06: Template Preview (needs CRUD)
- Epic 05: Email (needs template management)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Service layer implemented
- [x] Handler layer implemented
- [x] Repository methods implemented
- [x] All unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] RBAC enforcement verified
- [x] Error handling complete
- [x] Documentation updated
- [x] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11 (Templates & Customization)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md) - Section 3.1
- **API:** [lld/08_API_LLD.md](../lld/08_API_LLD.md)
- **Story 00:** [06_STORY_00_template_struct.md](06_STORY_00_template_struct.md)
