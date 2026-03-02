# User Story: Template Preview

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-09

---

## User Story

As an **event manager**, I want **to preview templates before saving** so that **I can verify the template renders correctly with sample data**.

---

## Acceptance Criteria

- [x] Preview endpoint created
- [x] Preview uses sample test data
- [x] Preview renders HTML templates
- [x] Preview renders text templates
- [x] Preview shows validation errors
- [x] Preview does not save template
- [x] Preview works for all template types
- [x] Preview accessible from template editor
- [x] All tests pass with timeout
- [x] Error handling for invalid templates

---

## Technical Details

### Preview Service

```go
package templates

import (
    "context"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

type PreviewRequest struct {
    Type        models.TemplateType `json:"type"`
    HTMLContent string              `json:"html_content"`
    TextContent *string             `json:"text_content,omitempty"`
    CSSContent  *string             `json:"css_content,omitempty"`
}

type PreviewResponse struct {
    HTMLPreview string            `json:"html_preview"`
    TextPreview string            `json:"text_preview,omitempty"`
    Errors      []ValidationError `json:"errors,omitempty"`
}

func (s *service) PreviewTemplate(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error) {
    template := &models.Template{
        Type:        req.Type,
        HTMLContent: req.HTMLContent,
        TextContent: req.TextContent,
        CSSContent:  req.CSSContent,
    }
    
    response := &PreviewResponse{}
    
    if err := s.validator.ValidateTemplate(template); err != nil {
        return nil, err
    }
    
    testData := createTestData(req.Type)
    
    htmlPreview, err := s.renderer.RenderHTML(req.HTMLContent, testData)
    if err != nil {
        return nil, &TemplateExecutionError{
            Message: "Failed to render HTML preview",
            Err:     err,
        }
    }
    response.HTMLPreview = htmlPreview
    
    if req.TextContent != nil {
        textPreview, err := s.renderer.RenderText(*req.TextContent, testData)
        if err != nil {
            return nil, &TemplateExecutionError{
                Message: "Failed to render text preview",
                Err:     err,
            }
        }
        response.TextPreview = textPreview
    }
    
    return response, nil
}
```

### Test Data Generation

```go
func createTestData(templateType models.TemplateType) interface{} {
    baseEvent := struct {
        Title        string
        Description  string
        StartTime    time.Time
        EndTime      *time.Time
        Timezone     string
        Location     string
        RSVPDeadline *time.Time
    }{
        Title:       "Sample Event",
        Description: "This is a sample event for template preview",
        StartTime:   time.Now().Add(30 * 24 * time.Hour),
        Timezone:    "America/Los_Angeles",
        Location:    "123 Main Street, City, State 12345",
    }
    
    endTime := baseEvent.StartTime.Add(3 * time.Hour)
    baseEvent.EndTime = &endTime
    
    deadline := baseEvent.StartTime.Add(-7 * 24 * time.Hour)
    baseEvent.RSVPDeadline = &deadline
    
    switch templateType {
    case models.TemplateTypeInviteEmail:
        return &InviteEmailData{
            Event: baseEvent,
            Invite: struct {
                Name  string
                Email string
            }{
                Name:  "John Doe",
                Email: "john@example.com",
            },
            RSVPURL:     "https://rsvp.example.com/rsvp/sample-token-preview",
            MaxPlusOnes: 2,
        }
        
    case models.TemplateTypeRSVPPage:
        return &RSVPPageData{
            Event:       baseEvent,
            Token:       "sample-token-preview",
            MaxPlusOnes: 2,
            Questions: []QuestionData{
                {
                    ID:           1,
                    QuestionText: "Dietary restrictions?",
                    QuestionType: "text",
                    Required:     false,
                },
                {
                    ID:           2,
                    QuestionText: "Preferred meal",
                    QuestionType: "select",
                    Options: []OptionData{
                        {Value: "chicken", Label: "Chicken"},
                        {Value: "fish", Label: "Fish"},
                        {Value: "vegetarian", Label: "Vegetarian"},
                    },
                    Required: true,
                },
            },
        }
        
    case models.TemplateTypeConfirmationPage:
        return &ConfirmationPageData{
            Event: baseEvent,
            Token: "sample-token-preview",
            RSVP: struct {
                Response string
                PlusOnes int
            }{
                Response: "yes",
                PlusOnes: 2,
            },
            Answers: []AnswerData{
                {
                    QuestionText:  "Dietary restrictions?",
                    AnswerDisplay: "Vegetarian",
                },
                {
                    QuestionText:  "Preferred meal",
                    AnswerDisplay: "Vegetarian",
                },
            },
        }
        
    default:
        return nil
    }
}
```

---

## API Endpoint

### Preview Template
```
POST /api/templates/preview
Authorization: Required (Event Manager or Admin)
Content-Type: application/json

{
    "type": "invite_email",
    "html_content": "<h1>{{.Event.Title}}</h1>",
    "text_content": "{{.Event.Title}}",
    "css_content": "h1 { color: blue; }"
}

Response 200 OK:
{
    "html_preview": "<h1>Sample Event</h1>",
    "text_preview": "Sample Event"
}

Response 400 Bad Request (validation error):
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Template syntax error: unexpected EOF",
        "field": "html_content"
    }
}
```

---

## Tasks

### Phase 1: Preview Service (TDD)
- [x] Define PreviewRequest struct
- [x] Define PreviewResponse struct
- [x] Write test for PreviewTemplate with valid template
- [x] Write test for PreviewTemplate with invalid syntax
- [x] Write test for PreviewTemplate with undefined variable
- [x] Write test for PreviewTemplate for each type
- [x] Implement PreviewTemplate
- [x] Run tests (should pass)

### Phase 2: Test Data Generation (TDD)
- [x] Write test for createTestData for invite email
- [x] Write test for createTestData for RSVP page
- [x] Write test for createTestData for confirmation page
- [x] Implement createTestData
- [x] Run tests (should pass)

### Phase 3: Handler Layer (TDD)
- [x] Create preview handler
- [x] Write test for POST /api/templates/preview
- [x] Write test for invalid JSON
- [x] Write test for validation errors
- [x] Write test for unauthorized access
- [x] Implement preview handler
- [x] Run tests (should pass)

### Phase 4: Integration Testing
- [x] Test preview with default templates
- [x] Test preview with custom templates
- [x] Test preview with all variable types
- [x] Test preview with all functions
- [x] Test error handling

---

## UI Integration

### Template Editor with Preview

```html
<div class="template-editor">
    <div class="editor-pane">
        <label>HTML Content</label>
        <textarea id="html-content" rows="20"></textarea>
        
        <label>Text Content</label>
        <textarea id="text-content" rows="10"></textarea>
        
        <button id="preview-btn">Preview</button>
        <button id="save-btn">Save</button>
    </div>
    
    <div class="preview-pane">
        <h3>Preview</h3>
        <div id="preview-container"></div>
    </div>
</div>

<script>
document.getElementById('preview-btn').addEventListener('click', async () => {
    const response = await fetch('/api/templates/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            type: 'invite_email',
            html_content: document.getElementById('html-content').value,
            text_content: document.getElementById('text-content').value
        })
    });
    
    const data = await response.json();
    document.getElementById('preview-container').innerHTML = data.html_preview;
});
</script>
```

---

## Testing Strategy

### Unit Tests

```go
func TestTemplateService_PreviewTemplate(t *testing.T) {
    renderer := NewRenderer()
    validator := NewValidator(renderer)
    service := NewService(nil, validator, renderer)
    
    tests := []struct {
        name    string
        req     *PreviewRequest
        wantErr bool
    }{
        {
            name: "valid invite email",
            req: &PreviewRequest{
                Type:        models.TemplateTypeInviteEmail,
                HTMLContent: "<h1>{{.Event.Title}}</h1>",
                TextContent: strPtr("{{.Event.Title}}"),
            },
            wantErr: false,
        },
        {
            name: "invalid syntax",
            req: &PreviewRequest{
                Type:        models.TemplateTypeRSVPPage,
                HTMLContent: "{{.Event.Title",
            },
            wantErr: true,
        },
        {
            name: "undefined variable",
            req: &PreviewRequest{
                Type:        models.TemplateTypeRSVPPage,
                HTMLContent: "{{.UndefinedVar}}",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := auth.WithUserID(context.Background(), 1)
            
            resp, err := service.PreviewTemplate(ctx, tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("PreviewTemplate() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if !tt.wantErr && resp.HTMLPreview == "" {
                t.Error("Expected non-empty HTML preview")
            }
        })
    }
}

func TestCreateTestData(t *testing.T) {
    tests := []struct {
        name         string
        templateType models.TemplateType
        wantNil      bool
    }{
        {"invite email", models.TemplateTypeInviteEmail, false},
        {"rsvp page", models.TemplateTypeRSVPPage, false},
        {"confirmation page", models.TemplateTypeConfirmationPage, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            data := createTestData(tt.templateType)
            if (data == nil) != tt.wantNil {
                t.Errorf("createTestData() = nil: %v, wantNil %v", data == nil, tt.wantNil)
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
- Story 02: Template Security (for validation)
- Story 05: Template Variables (for data structures)

**Blocks:**
- Story 04: Template CRUD (preview before save)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] PreviewTemplate implemented
- [x] Test data generation implemented
- [x] Preview handler implemented
- [x] All unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] Error handling complete
- [x] Documentation updated
- [x] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11 (Templates & Customization)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **Story 05:** [06_STORY_05_template_variables.md](06_STORY_05_template_variables.md)
