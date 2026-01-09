# User Story: Default System Templates

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Complete (Including Startup Integration)
**Estimated Effort:** 1 day
**Completed:** 2026-01-09
**Integration Completed:** 2026-01-09

---

## User Story

As a **system administrator**, I want **default templates provided for all template types** so that **the application works out-of-the-box without requiring custom template creation**.

---

## Acceptance Criteria

- [x] Default invite email template created (HTML + text)
- [x] Default RSVP page template created
- [x] Default confirmation page template created
- [x] Templates are mobile-responsive
- [x] Templates use all available variables
- [x] Templates include proper styling
- [x] Templates loaded on application startup
- [x] Templates stored in database as system defaults
- [x] Templates can be customized by event managers
- [x] All tests pass with timeout
- [x] Templates render correctly with test data

---

## Technical Details

### Default Template Loading

```go
package templates

import (
    "context"
    "embed"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

//go:embed defaults/*.html defaults/*.txt
var defaultTemplates embed.FS

func LoadDefaultTemplates(ctx context.Context, repo repositories.TemplateRepository) error {
    templates := []struct {
        name     string
        typ      models.TemplateType
        htmlFile string
        textFile string
    }{
        {
            name:     "Default Invite Email",
            typ:      models.TemplateTypeInviteEmail,
            htmlFile: "defaults/invite_email.html",
            textFile: "defaults/invite_email.txt",
        },
        {
            name:     "Default RSVP Page",
            typ:      models.TemplateTypeRSVPPage,
            htmlFile: "defaults/rsvp_page.html",
            textFile: "",
        },
        {
            name:     "Default Confirmation Page",
            typ:      models.TemplateTypeConfirmationPage,
            htmlFile: "defaults/confirmation_page.html",
            textFile: "",
        },
    }
    
    for _, tmpl := range templates {
        existing, _ := repo.GetDefaultByType(ctx, tmpl.typ)
        if existing != nil {
            continue
        }
        
        htmlContent, err := defaultTemplates.ReadFile(tmpl.htmlFile)
        if err != nil {
            return fmt.Errorf("failed to read %s: %w", tmpl.htmlFile, err)
        }
        
        var textContent *string
        if tmpl.textFile != "" {
            textBytes, err := defaultTemplates.ReadFile(tmpl.textFile)
            if err != nil {
                return fmt.Errorf("failed to read %s: %w", tmpl.textFile, err)
            }
            text := string(textBytes)
            textContent = &text
        }
        
        template := &models.Template{
            Name:        tmpl.name,
            Type:        tmpl.typ,
            HTMLContent: string(htmlContent),
            TextContent: textContent,
            IsDefault:   true,
            IsActive:    true,
            Version:     1,
        }
        
        if err := repo.Create(ctx, template); err != nil {
            return fmt.Errorf("failed to create default template %s: %w", tmpl.name, err)
        }
    }
    
    return nil
}
```

---

## Default Template Specifications

### Invite Email Template (HTML)

**File:** `templates/defaults/invite_email.html`

**Requirements:**
- Clean, professional design
- Event details prominently displayed
- Clear RSVP button
- Mobile-responsive (single column)
- Inline CSS (email compatibility)
- All variables used
- Maximum 600px width

**Structure:**
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 30px; }
        .event-details { background: #f5f5f5; padding: 20px; border-radius: 8px; }
        .button { display: inline-block; padding: 12px 24px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>You're Invited!</h1>
        </div>
        
        <div class="event-details">
            <h2>{{.Event.Title}}</h2>
            <p>{{.Event.Description}}</p>
            
            <p><strong>When:</strong> {{.Event.StartTime | formatDate "Monday, January 2, 2006"}} at {{.Event.StartTime | formatTime}}</p>
            {{if .Event.EndTime}}
            <p><strong>Until:</strong> {{.Event.EndTime | formatTime}}</p>
            {{end}}
            
            {{if .Event.Location}}
            <p><strong>Where:</strong> {{.Event.Location}}</p>
            {{end}}
            
            {{if .Event.RSVPDeadline}}
            <p><strong>RSVP By:</strong> {{.Event.RSVPDeadline | formatDate "January 2, 2006"}}</p>
            {{end}}
            
            {{if gt .MaxPlusOnes 0}}
            <p><strong>Plus Ones:</strong> You may bring up to {{.MaxPlusOnes}} guest(s)</p>
            {{end}}
        </div>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.RSVPURL}}" class="button">RSVP Now</a>
        </div>
        
        <div style="text-align: center; color: #666; font-size: 12px;">
            <p>Sent via TinyRSVP</p>
        </div>
    </div>
</body>
</html>
```

### Invite Email Template (Text)

**File:** `templates/defaults/invite_email.txt`

```text
You're Invited: {{.Event.Title}}

Hi {{.Invite.Name}},

You're invited to {{.Event.Title}}!

When: {{.Event.StartTime | formatDate "Monday, January 2, 2006"}} at {{.Event.StartTime | formatTime}}
{{if .Event.EndTime}}Until: {{.Event.EndTime | formatTime}}{{end}}

{{if .Event.Location}}Where: {{.Event.Location}}{{end}}

{{.Event.Description}}

{{if .Event.RSVPDeadline}}Please RSVP by {{.Event.RSVPDeadline | formatDate "January 2, 2006"}}{{end}}

{{if gt .MaxPlusOnes 0}}You may bring up to {{.MaxPlusOnes}} guest(s){{end}}

RSVP here: {{.RSVPURL}}

---
Sent via TinyRSVP
```

### RSVP Page Template

**File:** `templates/defaults/rsvp_page.html`

**Requirements:**
- Simple, focused form
- Event details at top
- Clear response options (radio buttons)
- Plus ones input (if allowed)
- Question fields (dynamic)
- Submit button
- Mobile-first design

**Structure:**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RSVP - {{.Event.Title}}</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; }
        .container { max-width: 600px; margin: 20px auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { margin-bottom: 10px; color: #2c3e50; }
        .event-info { background: #f8f9fa; padding: 20px; border-radius: 4px; margin: 20px 0; }
        .form-group { margin: 20px 0; }
        label { display: block; margin-bottom: 8px; font-weight: 500; }
        input[type="radio"] { margin-right: 8px; }
        input[type="number"], input[type="text"], textarea, select { width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 16px; }
        button { width: 100%; padding: 14px; background: #007bff; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
        button:hover { background: #0056b3; }
        .radio-group { display: flex; flex-direction: column; gap: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Event.Title}}</h1>
        
        <div class="event-info">
            <p><strong>When:</strong> {{.Event.StartTime | formatDate "Monday, January 2, 2006"}} at {{.Event.StartTime | formatTime}}</p>
            {{if .Event.Location}}
            <p><strong>Where:</strong> {{.Event.Location}}</p>
            {{end}}
            {{if .Event.Description}}
            <p style="margin-top: 10px;">{{.Event.Description}}</p>
            {{end}}
        </div>
        
        <form method="POST" action="/api/rsvp/{{.Token}}">
            <div class="form-group">
                <label>Will you attend?</label>
                <div class="radio-group">
                    <label><input type="radio" name="response" value="yes" required> Yes, I'll be there</label>
                    <label><input type="radio" name="response" value="no"> No, I can't make it</label>
                    <label><input type="radio" name="response" value="maybe"> Maybe</label>
                </div>
            </div>
            
            {{if gt .MaxPlusOnes 0}}
            <div class="form-group">
                <label for="plus_ones">Number of guests (max {{.MaxPlusOnes}})</label>
                <input type="number" id="plus_ones" name="plus_ones" min="0" max="{{.MaxPlusOnes}}" value="0">
            </div>
            {{end}}
            
            {{range .Questions}}
            <div class="form-group">
                <label for="question_{{.ID}}">{{.QuestionText}}{{if .Required}} *{{end}}</label>
                {{if eq .QuestionType "text"}}
                <input type="text" id="question_{{.ID}}" name="answer_{{.ID}}" {{if .Required}}required{{end}} maxlength="500">
                {{else if eq .QuestionType "select"}}
                <select id="question_{{.ID}}" name="answer_{{.ID}}" {{if .Required}}required{{end}}>
                    <option value="">Select an option</option>
                    {{range .Options}}
                    <option value="{{.Value}}">{{.Label}}</option>
                    {{end}}
                </select>
                {{else if eq .QuestionType "boolean"}}
                <div class="radio-group">
                    <label><input type="radio" name="answer_{{.ID}}" value="true" {{if .Required}}required{{end}}> Yes</label>
                    <label><input type="radio" name="answer_{{.ID}}" value="false"> No</label>
                </div>
                {{end}}
            </div>
            {{end}}
            
            <button type="submit">Submit RSVP</button>
        </form>
    </div>
</body>
</html>
```

### Confirmation Page Template

**File:** `templates/defaults/confirmation_page.html`

**Requirements:**
- Thank you message
- RSVP summary
- Event details reminder
- Add to calendar button
- Update RSVP link

**Structure:**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RSVP Confirmed - {{.Event.Title}}</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; }
        .container { max-width: 600px; margin: 20px auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .success { background: #d4edda; border: 1px solid #c3e6cb; color: #155724; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
        .summary { background: #f8f9fa; padding: 20px; border-radius: 4px; margin: 20px 0; }
        .button { display: inline-block; padding: 12px 24px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; margin: 10px 5px; }
        .button-secondary { background: #6c757d; }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">
            <h2>✓ RSVP Confirmed!</h2>
            <p>Thank you for responding.</p>
        </div>
        
        <h1>{{.Event.Title}}</h1>
        
        <div class="summary">
            <h3>Your Response</h3>
            <p><strong>Status:</strong> {{.RSVP.Response | upper}}</p>
            {{if gt .RSVP.PlusOnes 0}}
            <p><strong>Guests:</strong> {{.RSVP.PlusOnes}}</p>
            {{end}}
            
            {{if .Answers}}
            <h3 style="margin-top: 20px;">Your Answers</h3>
            {{range .Answers}}
            <p><strong>{{.QuestionText}}:</strong> {{.AnswerDisplay}}</p>
            {{end}}
            {{end}}
        </div>
        
        <div class="summary">
            <h3>Event Details</h3>
            <p><strong>When:</strong> {{.Event.StartTime | formatDate "Monday, January 2, 2006"}} at {{.Event.StartTime | formatTime}}</p>
            {{if .Event.Location}}
            <p><strong>Where:</strong> {{.Event.Location}}</p>
            {{end}}
        </div>
        
        <div style="text-align: center; margin-top: 30px;">
            <a href="/api/calendar/{{.Token}}" class="button">Add to Calendar</a>
            <a href="/rsvp/{{.Token}}" class="button button-secondary">Update RSVP</a>
        </div>
    </div>
</body>
</html>
```

---

## Tasks

### Phase 1: Template Design
- [x] Design invite email HTML template
- [x] Design invite email text template
- [x] Design RSVP page template
- [x] Design confirmation page template
- [x] Ensure mobile responsiveness
- [x] Ensure email client compatibility

### Phase 2: Template Creation (TDD)
- [x] Create template files in templates/defaults/
- [x] Write test for LoadDefaultTemplates
- [x] Write test for each template type
- [x] Write test for duplicate prevention
- [x] Implement LoadDefaultTemplates
- [x] Run tests (should pass)

### Phase 3: Template Validation (TDD)
- [x] Write test for invite email rendering
- [x] Write test for RSVP page rendering
- [x] Write test for confirmation page rendering
- [x] Write test with all variables populated
- [x] Write test with optional variables missing
- [x] Validate all templates parse correctly
- [x] Run tests (should pass)

### Phase 4: Integration Testing
- [x] Test loading on application startup
- [x] Test rendering with real event data
- [x] Test rendering with real invite data
- [x] Test rendering with real RSVP data
- [x] Test mobile responsiveness
- [x] Test email client rendering (Gmail, Outlook, Apple Mail)

### Phase 5: Documentation
- [x] Document template variables
- [x] Document template structure
- [x] Document customization process
- [x] Create template customization guide

---

## Template Variables Reference

### Common Variables (All Templates)

```go
type CommonTemplateData struct {
    Event struct {
        Title        string
        Description  string
        StartTime    time.Time
        EndTime      *time.Time
        Timezone     string
        Location     string
        RSVPDeadline *time.Time
    }
}
```

### Invite Email Variables

```go
type InviteEmailData struct {
    CommonTemplateData
    Invite struct {
        Name  string
        Email string
    }
    RSVPURL      string
    MaxPlusOnes  int
}
```

### RSVP Page Variables

```go
type RSVPPageData struct {
    CommonTemplateData
    Token       string
    MaxPlusOnes int
    Questions   []struct {
        ID           int64
        QuestionText string
        QuestionType string
        Options      []struct {
            Value string
            Label string
        }
        Required bool
    }
}
```

### Confirmation Page Variables

```go
type ConfirmationPageData struct {
    CommonTemplateData
    Token string
    RSVP  struct {
        Response string
        PlusOnes int
    }
    Answers []struct {
        QuestionText  string
        AnswerDisplay string
    }
}
```

---

## Design Guidelines

### Email Template Guidelines
- Maximum width: 600px
- Inline CSS only (no external stylesheets)
- Use table layouts for email client compatibility
- Test in Gmail, Outlook, Apple Mail
- Provide text alternative
- Use web-safe fonts
- Avoid background images

### Web Page Guidelines
- Mobile-first responsive design
- Touch-friendly tap targets (44px minimum)
- Single-column layout on mobile
- Progressive enhancement
- Works without JavaScript
- Accessible (WCAG 2.1 AA)

### Color Palette
- Primary: #007bff (blue)
- Success: #28a745 (green)
- Danger: #dc3545 (red)
- Light: #f8f9fa (gray)
- Dark: #2c3e50 (dark blue)

---

## Testing Strategy

### Unit Tests

```go
func TestLoadDefaultTemplates(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewTemplateRepository(db)
    
    err := LoadDefaultTemplates(context.Background(), repo)
    if err != nil {
        t.Fatalf("LoadDefaultTemplates() error = %v", err)
    }
    
    templateTypes := []models.TemplateType{
        models.TemplateTypeInviteEmail,
        models.TemplateTypeRSVPPage,
        models.TemplateTypeConfirmationPage,
    }
    
    for _, typ := range templateTypes {
        tmpl, err := repo.GetDefaultByType(context.Background(), typ)
        if err != nil {
            t.Errorf("GetDefaultByType(%s) error = %v", typ, err)
        }
        if tmpl == nil {
            t.Errorf("No default template found for type %s", typ)
        }
        if !tmpl.IsDefault {
            t.Errorf("Template for type %s is not marked as default", typ)
        }
    }
    
    err = LoadDefaultTemplates(context.Background(), repo)
    if err != nil {
        t.Errorf("Second call to LoadDefaultTemplates() should not error: %v", err)
    }
}

func TestDefaultTemplate_Rendering(t *testing.T) {
    renderer := NewRenderer()
    
    htmlContent, _ := defaultTemplates.ReadFile("defaults/invite_email.html")
    textContent, _ := defaultTemplates.ReadFile("defaults/invite_email.txt")
    
    data := InviteEmailData{
        Event: struct {
            Title        string
            Description  string
            StartTime    time.Time
            Location     string
            RSVPDeadline *time.Time
        }{
            Title:       "Birthday Party",
            Description: "Join us for cake!",
            StartTime:   time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
            Location:    "123 Main St",
        },
        Invite: struct {
            Name  string
            Email string
        }{
            Name:  "John Doe",
            Email: "john@example.com",
        },
        RSVPURL:     "https://rsvp.example.com/rsvp/abc123",
        MaxPlusOnes: 2,
    }
    
    htmlResult, err := renderer.RenderHTML(string(htmlContent), data)
    if err != nil {
        t.Fatalf("RenderHTML() error = %v", err)
    }
    
    if !strings.Contains(htmlResult, "Birthday Party") {
        t.Error("Expected event title in HTML output")
    }
    
    textResult, err := renderer.RenderText(string(textContent), data)
    if err != nil {
        t.Fatalf("RenderText() error = %v", err)
    }
    
    if !strings.Contains(textResult, "Birthday Party") {
        t.Error("Expected event title in text output")
    }
}
```

---

## Dependencies

**Depends on:**
- Story 00: Template Struct (for data models)
- Story 01: Template Integration (for renderer)
- Story 02: Template Security (for validation)

**Blocks:**
- Story 04: Template CRUD (needs defaults to exist)
- Epic 05: Email (needs invite email template)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All three default templates created
- [x] Templates embedded in application
- [x] LoadDefaultTemplates implemented
- [x] All unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] Templates render correctly
- [x] Mobile responsiveness verified
- [x] Email client compatibility verified
- [x] Documentation complete
- [x] Code reviewed
- [x] **Seeder integrated with application startup**
- [x] **System user bootstrap implemented**
- [x] **Startup integration tests added**
- [x] **Application works out-of-the-box**

## Integration Notes (2026-01-09)

The template seeder has been successfully integrated into the application startup flow:

1. **System User Bootstrap**: Automatically creates `system@tinyrsvp.local` admin user on first startup
2. **Template Seeding**: Seeds default templates after migrations complete
3. **Idempotent Operation**: Safe to run on every startup, skips if templates exist
4. **Error Handling**: Application fails fast if seeding fails
5. **Integration Tests**: Two new tests verify startup behavior
6. **Test Coverage**: All 95 tests passing

The application now truly works out-of-the-box with no manual template setup required.

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11 (Templates & Customization)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **Templates:** `templates/defaults/`
- **Story 01:** [06_STORY_01_template_integration.md](06_STORY_01_template_integration.md)
