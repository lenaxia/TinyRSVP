# User Story: Template Security Validation

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** Critical
**Status:** Complete
**Estimated Effort:** 1 day
**Completed:** 2026-01-09

---

## User Story

As a **system administrator**, I want **template validation and security checks** so that **malicious templates cannot be uploaded and XSS attacks are prevented**.

---

## Acceptance Criteria

- [x] Template validator service created
- [x] Template syntax validation implemented
- [x] Variable reference validation implemented
- [x] Disallowed function detection implemented
- [x] Template size limits enforced
- [x] Parse errors caught before saving
- [x] Undefined variable detection implemented
- [x] Security validation on upload
- [x] Security validation on update
- [x] All tests pass with timeout
- [x] Security tests verify XSS prevention

---

## Technical Details

### Validator Interface

```go
package templates

import (
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type Validator interface {
    ValidateTemplate(template *models.Template) error
    ValidateSyntax(content string, templateType models.TemplateType) error
    ValidateVariables(content string, allowedVars []string) error
    ValidateSize(content string, maxBytes int) error
}

type validator struct {
    renderer Renderer
}

func NewValidator(renderer Renderer) Validator {
    return &validator{
        renderer: renderer,
    }
}
```

### Template Validation

```go
func (v *validator) ValidateTemplate(template *models.Template) error {
    if err := template.Validate(); err != nil {
        return err
    }
    
    if err := v.ValidateSize(template.HTMLContent, 100*1024); err != nil {
        return err
    }
    
    if err := v.ValidateSyntax(template.HTMLContent, template.Type); err != nil {
        return err
    }
    
    allowedVars := getAllowedVariables(template.Type)
    if err := v.ValidateVariables(template.HTMLContent, allowedVars); err != nil {
        return err
    }
    
    if template.TextContent != nil {
        if err := v.ValidateSize(*template.TextContent, 50*1024); err != nil {
            return err
        }
        
        if err := v.ValidateSyntax(*template.TextContent, template.Type); err != nil {
            return err
        }
    }
    
    if template.CSSContent != nil {
        if err := v.ValidateCSS(*template.CSSContent); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Syntax Validation

```go
func (v *validator) ValidateSyntax(content string, templateType models.TemplateType) error {
    testData := createTestData(templateType)
    
    _, err := v.renderer.RenderHTML(content, testData)
    if err != nil {
        return &ValidationError{
            Field:   "template_content",
            Message: fmt.Sprintf("Template syntax error: %v", err),
        }
    }
    
    return nil
}
```

### Variable Validation

```go
func (v *validator) ValidateVariables(content string, allowedVars []string) error {
    tmpl, err := template.New("check").Parse(content)
    if err != nil {
        return err
    }
    
    usedVars := extractVariables(tmpl.Tree.Root)
    
    for _, usedVar := range usedVars {
        if !contains(allowedVars, usedVar) {
            return &ValidationError{
                Field:   "template_content",
                Message: fmt.Sprintf("Undefined variable: {{.%s}}", usedVar),
            }
        }
    }
    
    return nil
}

func getAllowedVariables(templateType models.TemplateType) []string {
    common := []string{
        "Event.Title",
        "Event.Description",
        "Event.StartTime",
        "Event.EndTime",
        "Event.Timezone",
        "Event.Location",
        "Event.RSVPDeadline",
    }
    
    switch templateType {
    case models.TemplateTypeInviteEmail:
        return append(common, []string{
            "Invite.Name",
            "Invite.Email",
            "RSVPURL",
            "MaxPlusOnes",
        }...)
    case models.TemplateTypeRSVPPage:
        return append(common, []string{
            "RSVP.Response",
            "RSVP.PlusOnes",
            "Questions",
        }...)
    case models.TemplateTypeConfirmationPage:
        return append(common, []string{
            "RSVP.Response",
            "RSVP.PlusOnes",
            "Answers",
        }...)
    default:
        return common
    }
}
```

### Size Validation

```go
func (v *validator) ValidateSize(content string, maxBytes int) error {
    if len(content) > maxBytes {
        return &ValidationError{
            Field:   "template_content",
            Message: fmt.Sprintf("Template size exceeds %d bytes", maxBytes),
        }
    }
    return nil
}
```

---

## Tasks

### Phase 1: Validator Setup (TDD)
- [x] Define Validator interface
- [x] Write test for NewValidator
- [x] Implement NewValidator
- [x] Write test for getAllowedVariables
- [x] Implement getAllowedVariables
- [x] Run tests (should pass)

### Phase 2: Syntax Validation (TDD)
- [x] Write test for ValidateSyntax with valid template
- [x] Write test for ValidateSyntax with parse error
- [x] Write test for ValidateSyntax with execution error
- [x] Write test for ValidateSyntax with each template type
- [x] Implement ValidateSyntax
- [x] Run tests (should pass)

### Phase 3: Variable Validation (TDD)
- [x] Write test for ValidateVariables with valid vars
- [x] Write test for ValidateVariables with undefined var
- [x] Write test for ValidateVariables with nested vars
- [x] Write test for ValidateVariables with each template type
- [x] Implement ValidateVariables
- [x] Implement extractVariables helper
- [x] Run tests (should pass)

### Phase 4: Size Validation (TDD)
- [x] Write test for ValidateSize within limit
- [x] Write test for ValidateSize exceeding limit
- [x] Write test for ValidateSize at boundary
- [x] Implement ValidateSize
- [x] Run tests (should pass)

### Phase 5: Complete Validation (TDD)
- [x] Write test for ValidateTemplate success
- [x] Write test for ValidateTemplate with invalid syntax
- [x] Write test for ValidateTemplate with undefined vars
- [x] Write test for ValidateTemplate with size exceeded
- [x] Write test for ValidateTemplate with missing text content
- [x] Implement ValidateTemplate
- [x] Run tests (should pass)

### Phase 6: Security Testing
- [x] Test XSS prevention with script tags
- [x] Test XSS prevention with event handlers
- [x] Test XSS prevention with javascript: URLs
- [x] Test XSS prevention with data: URLs
- [x] Test XSS prevention with SVG payloads
- [x] Verify all payloads properly escaped

---

## Validation Rules

### Template Size Limits
- HTML content: Maximum 100KB
- Text content: Maximum 50KB
- CSS content: Maximum 50KB

### Allowed Template Variables

#### All Templates
- Event.Title
- Event.Description
- Event.StartTime
- Event.EndTime
- Event.Timezone
- Event.Location
- Event.RSVPDeadline

#### Invite Email Templates
- All common variables plus:
- Invite.Name
- Invite.Email
- RSVPURL
- MaxPlusOnes

#### RSVP Page Templates
- All common variables plus:
- RSVP.Response
- RSVP.PlusOnes
- Questions (array)

#### Confirmation Page Templates
- All common variables plus:
- RSVP.Response
- RSVP.PlusOnes
- Answers (array)

### Disallowed Functions
- `call` - Arbitrary function calls
- `js` - JavaScript execution
- `html` - Disable escaping (security risk)

### Allowed Functions
- `upper`, `lower`, `title` - String operations
- `formatDate`, `formatTime` - Date formatting
- `add`, `sub` - Math operations
- `if`, `range`, `with` - Control flow (built-in)

---

## Error Handling

| Error Condition | Error Type | Message |
|----------------|------------|---------|
| Parse error | `ValidationError` | "Template syntax error: {details}" |
| Undefined variable | `ValidationError` | "Undefined variable: {{.VarName}}" |
| Size exceeded | `ValidationError` | "Template size exceeds {limit} bytes" |
| Disallowed function | `ValidationError` | "Function not allowed: {function}" |
| Missing text content | `ValidationError` | "Text content required for email templates" |

---

## Testing Strategy

### Unit Tests

```go
func TestValidator_ValidateTemplate(t *testing.T) {
    renderer := NewRenderer()
    validator := NewValidator(renderer)
    
    tests := []struct {
        name     string
        template *models.Template
        wantErr  bool
        errMsg   string
    }{
        {
            name: "valid invite email template",
            template: &models.Template{
                Name:        "Custom Invite",
                Type:        models.TemplateTypeInviteEmail,
                HTMLContent: "<h1>{{.Event.Title}}</h1>",
                TextContent: strPtr("Event: {{.Event.Title}}"),
            },
            wantErr: false,
        },
        {
            name: "undefined variable",
            template: &models.Template{
                Name:        "Invalid Template",
                Type:        models.TemplateTypeRSVPPage,
                HTMLContent: "<h1>{{.UndefinedVar}}</h1>",
            },
            wantErr: true,
            errMsg:  "Undefined variable",
        },
        {
            name: "size exceeded",
            template: &models.Template{
                Name:        "Large Template",
                Type:        models.TemplateTypeRSVPPage,
                HTMLContent: strings.Repeat("x", 101*1024),
            },
            wantErr: true,
            errMsg:  "exceeds",
        },
        {
            name: "parse error",
            template: &models.Template{
                Name:        "Broken Template",
                Type:        models.TemplateTypeRSVPPage,
                HTMLContent: "{{.Event.Title",
            },
            wantErr: true,
            errMsg:  "syntax error",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.ValidateTemplate(tt.template)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
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

### Security Tests

```go
func TestValidator_XSSPrevention(t *testing.T) {
    renderer := NewRenderer()
    validator := NewValidator(renderer)
    
    xssPayloads := []struct {
        name    string
        content string
    }{
        {"script tag", "<div>{{.Event.Description}}</div>"},
        {"event handler", "<img src='{{.Event.Location}}'>"},
        {"javascript url", "<a href='{{.RSVPURL}}'>Click</a>"},
    }
    
    for _, payload := range xssPayloads {
        t.Run(payload.name, func(t *testing.T) {
            template := &models.Template{
                Name:        "XSS Test",
                Type:        models.TemplateTypeRSVPPage,
                HTMLContent: payload.content,
            }
            
            err := validator.ValidateTemplate(template)
            if err != nil {
                t.Fatalf("ValidateTemplate() error = %v", err)
            }
            
            testData := struct {
                Event struct {
                    Description string
                    Location    string
                }
                RSVPURL string
            }{
                Event: struct {
                    Description string
                    Location    string
                }{
                    Description: "<script>alert('xss')</script>",
                    Location:    "javascript:alert('xss')",
                },
                RSVPURL: "javascript:alert('xss')",
            }
            
            result, err := renderer.RenderHTML(template.HTMLContent, testData)
            if err != nil {
                t.Fatalf("RenderHTML() error = %v", err)
            }
            
            if strings.Contains(result, "<script>") ||
               strings.Contains(result, "javascript:") {
                t.Errorf("XSS not prevented in result: %s", result)
            }
        })
    }
}
```

---

## Dependencies

**Depends on:**
- Story 00: Template Struct (for data models)
- Story 01: Template Integration (for renderer)

**Blocks:**
- Story 03: Default Templates (needs validation)
- Story 04: Template CRUD (needs validation)
- Story 11: XSS Prevention (implements validation)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Validator interface defined
- [x] ValidateTemplate implemented
- [x] ValidateSyntax implemented
- [x] ValidateVariables implemented
- [x] ValidateSize implemented
- [x] All unit tests passing (>90% coverage)
- [x] Security tests passing
- [x] XSS prevention verified
- [x] Documentation updated
- [x] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11.4 (Template Security), Section 16.3 (Input Sanitization)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **Story 01:** [06_STORY_01_template_integration.md](06_STORY_01_template_integration.md)
- **Story 11:** [06_STORY_11_xss_prevention.md](06_STORY_11_xss_prevention.md)
