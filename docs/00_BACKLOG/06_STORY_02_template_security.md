# User Story: Template Security Validation

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **system administrator**, I want **template validation and security checks** so that **malicious templates cannot be uploaded and XSS attacks are prevented**.

---

## Acceptance Criteria

- [ ] Template validator service created
- [ ] Template syntax validation implemented
- [ ] Variable reference validation implemented
- [ ] Disallowed function detection implemented
- [ ] Template size limits enforced
- [ ] Parse errors caught before saving
- [ ] Undefined variable detection implemented
- [ ] Security validation on upload
- [ ] Security validation on update
- [ ] All tests pass with timeout
- [ ] Security tests verify XSS prevention

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
- [ ] Define Validator interface
- [ ] Write test for NewValidator
- [ ] Implement NewValidator
- [ ] Write test for getAllowedVariables
- [ ] Implement getAllowedVariables
- [ ] Run tests (should pass)

### Phase 2: Syntax Validation (TDD)
- [ ] Write test for ValidateSyntax with valid template
- [ ] Write test for ValidateSyntax with parse error
- [ ] Write test for ValidateSyntax with execution error
- [ ] Write test for ValidateSyntax with each template type
- [ ] Implement ValidateSyntax
- [ ] Run tests (should pass)

### Phase 3: Variable Validation (TDD)
- [ ] Write test for ValidateVariables with valid vars
- [ ] Write test for ValidateVariables with undefined var
- [ ] Write test for ValidateVariables with nested vars
- [ ] Write test for ValidateVariables with each template type
- [ ] Implement ValidateVariables
- [ ] Implement extractVariables helper
- [ ] Run tests (should pass)

### Phase 4: Size Validation (TDD)
- [ ] Write test for ValidateSize within limit
- [ ] Write test for ValidateSize exceeding limit
- [ ] Write test for ValidateSize at boundary
- [ ] Implement ValidateSize
- [ ] Run tests (should pass)

### Phase 5: Complete Validation (TDD)
- [ ] Write test for ValidateTemplate success
- [ ] Write test for ValidateTemplate with invalid syntax
- [ ] Write test for ValidateTemplate with undefined vars
- [ ] Write test for ValidateTemplate with size exceeded
- [ ] Write test for ValidateTemplate with missing text content
- [ ] Implement ValidateTemplate
- [ ] Run tests (should pass)

### Phase 6: Security Testing
- [ ] Test XSS prevention with script tags
- [ ] Test XSS prevention with event handlers
- [ ] Test XSS prevention with javascript: URLs
- [ ] Test XSS prevention with data: URLs
- [ ] Test XSS prevention with SVG payloads
- [ ] Verify all payloads properly escaped

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

- [ ] All acceptance criteria met
- [ ] Validator interface defined
- [ ] ValidateTemplate implemented
- [ ] ValidateSyntax implemented
- [ ] ValidateVariables implemented
- [ ] ValidateSize implemented
- [ ] All unit tests passing (>90% coverage)
- [ ] Security tests passing
- [ ] XSS prevention verified
- [ ] Documentation updated
- [ ] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11.4 (Template Security), Section 16.3 (Input Sanitization)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **Story 01:** [06_STORY_01_template_integration.md](06_STORY_01_template_integration.md)
- **Story 11:** [06_STORY_11_xss_prevention.md](06_STORY_11_xss_prevention.md)
