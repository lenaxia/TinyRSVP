# User Story: Template Rendering Service

**Epic:** [05_EPIC_email.md](05_EPIC_email.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 1 day

---

## User Story

As a **system**, I want **a template rendering service for emails** so that **email content can be generated from templates with dynamic data**.

---

## Acceptance Criteria

- [x] Render HTML email templates
- [x] Render plain text email templates
- [x] Support template data injection
- [x] Support template functions (date formatting, etc.)
- [x] Template caching for performance
- [x] Template validation on load
- [x] Error handling for missing templates
- [x] Error handling for invalid data
- [x] All tests pass with timeout
- [x] Integration tests with real templates

---

## Technical Details

### Template Renderer Interface

```go
package email

import (
    "context"
)

type TemplateRenderer interface {
    // RenderHTML renders an HTML email template
    RenderHTML(ctx context.Context, templateName string, data interface{}) (string, error)
    
    // RenderText renders a plain text email template
    RenderText(ctx context.Context, templateName string, data interface{}) (string, error)
    
    // LoadTemplates loads all email templates
    LoadTemplates() error
    
    // ReloadTemplates reloads templates (for development)
    ReloadTemplates() error
}

type TemplateConfig struct {
    TemplateDir string
    CacheEnabled bool
}
```

### Implementation

```go
package email

import (
    "context"
    "fmt"
    "html/template"
    "path/filepath"
    "sync"
    "text/template" as textTemplate
    "time"
)

type templateRenderer struct {
    config        *TemplateConfig
    htmlTemplates map[string]*template.Template
    textTemplates map[string]*textTemplate.Template
    mu            sync.RWMutex
}

func NewTemplateRenderer(config *TemplateConfig) (TemplateRenderer, error) {
    if config.TemplateDir == "" {
        return nil, fmt.Errorf("template directory is required")
    }
    
    r := &templateRenderer{
        config:        config,
        htmlTemplates: make(map[string]*template.Template),
        textTemplates: make(map[string]*textTemplate.Template),
    }
    
    if err := r.LoadTemplates(); err != nil {
        return nil, fmt.Errorf("failed to load templates: %w", err)
    }
    
    return r, nil
}

func (r *templateRenderer) RenderHTML(ctx context.Context, templateName string, data interface{}) (string, error) {
    r.mu.RLock()
    tmpl, ok := r.htmlTemplates[templateName]
    r.mu.RUnlock()
    
    if !ok {
        return "", fmt.Errorf("HTML template not found: %s", templateName)
    }
    
    var buf strings.Builder
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("failed to execute HTML template: %w", err)
    }
    
    return buf.String(), nil
}

func (r *templateRenderer) RenderText(ctx context.Context, templateName string, data interface{}) (string, error) {
    r.mu.RLock()
    tmpl, ok := r.textTemplates[templateName]
    r.mu.RUnlock()
    
    if !ok {
        return "", fmt.Errorf("text template not found: %s", templateName)
    }
    
    var buf strings.Builder
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("failed to execute text template: %w", err)
    }
    
    return buf.String(), nil
}

func (r *templateRenderer) LoadTemplates() error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Load HTML templates
    htmlPattern := filepath.Join(r.config.TemplateDir, "*.html")
    htmlFiles, err := filepath.Glob(htmlPattern)
    if err != nil {
        return fmt.Errorf("failed to glob HTML templates: %w", err)
    }
    
    for _, file := range htmlFiles {
        name := filepath.Base(file)
        name = strings.TrimSuffix(name, ".html")
        
        tmpl, err := template.New(name).Funcs(templateFuncs()).ParseFiles(file)
        if err != nil {
            return fmt.Errorf("failed to parse HTML template %s: %w", file, err)
        }
        
        r.htmlTemplates[name] = tmpl
    }
    
    // Load text templates
    textPattern := filepath.Join(r.config.TemplateDir, "*.txt")
    textFiles, err := filepath.Glob(textPattern)
    if err != nil {
        return fmt.Errorf("failed to glob text templates: %w", err)
    }
    
    for _, file := range textFiles {
        name := filepath.Base(file)
        name = strings.TrimSuffix(name, ".txt")
        
        tmpl, err := textTemplate.New(name).Funcs(textTemplateFuncs()).ParseFiles(file)
        if err != nil {
            return fmt.Errorf("failed to parse text template %s: %w", file, err)
        }
        
        r.textTemplates[name] = tmpl
    }
    
    return nil
}

func (r *templateRenderer) ReloadTemplates() error {
    return r.LoadTemplates()
}

func templateFuncs() template.FuncMap {
    return template.FuncMap{
        "formatDate": func(t time.Time, layout string) string {
            return t.Format(layout)
        },
        "formatDateTime": func(t time.Time) string {
            return t.Format("Monday, January 2, 2006 at 3:04 PM")
        },
        "title": strings.Title,
        "upper": strings.ToUpper,
        "lower": strings.ToLower,
    }
}

func textTemplateFuncs() textTemplate.FuncMap {
    return textTemplate.FuncMap{
        "formatDate": func(t time.Time, layout string) string {
            return t.Format(layout)
        },
        "formatDateTime": func(t time.Time) string {
            return t.Format("Monday, January 2, 2006 at 3:04 PM")
        },
        "title": strings.Title,
        "upper": strings.ToUpper,
        "lower": strings.ToLower,
    }
}
```

---

## Tasks

### Phase 1: Interface & Configuration (TDD)
- [x] Define TemplateRenderer interface
- [x] Define TemplateConfig struct
- [x] Write test for renderer initialization
- [x] Implement renderer struct
- [x] Write test for config validation
- [x] Implement config validation

### Phase 2: Template Loading (TDD)
- [x] Write test for loading HTML templates
- [x] Implement HTML template loading
- [x] Write test for loading text templates
- [x] Implement text template loading
- [x] Write test for template caching
- [x] Implement template caching
- [x] Write test for reload functionality
- [x] Implement reload functionality

### Phase 3: Template Rendering (TDD)
- [x] Write test for HTML rendering
- [x] Implement RenderHTML method
- [x] Write test for text rendering
- [x] Implement RenderText method
- [x] Write test for data injection
- [x] Verify data injection works
- [x] Write test for missing templates
- [x] Implement error handling

### Phase 4: Template Functions (TDD)
- [x] Write test for date formatting
- [x] Implement date formatting function
- [x] Write test for string functions
- [x] Implement string functions
- [x] Write test for custom functions
- [x] Add custom template functions

### Phase 5: Integration Testing
- [x] Test with real email templates
- [x] Test with complex data structures
- [x] Test template caching performance
- [x] Test concurrent rendering
- [x] Test template reload
- [x] Test error scenarios

---

## Dependencies

**Depends on:**
- Go standard library: `html/template`, `text/template`
- Email templates in `templates/email/`

**Blocks:**
- Story 02: Email Queue Processor (needs rendered content)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Interface defined and documented
- [x] Implementation complete
- [x] All unit tests passing
- [x] Integration tests passing
- [x] Template functions working
- [x] Code reviewed
- [x] Documentation updated

---

## Test Requirements

### Unit Tests

```go
func TestTemplateRenderer_RenderHTML(t *testing.T) {
    config := &TemplateConfig{
        TemplateDir: "testdata/templates",
        CacheEnabled: true,
    }
    
    renderer, err := NewTemplateRenderer(config)
    if err != nil {
        t.Fatal(err)
    }
    
    data := struct {
        Name  string
        Event string
    }{
        Name:  "John Doe",
        Event: "Birthday Party",
    }
    
    html, err := renderer.RenderHTML(context.Background(), "test_template", data)
    if err != nil {
        t.Errorf("RenderHTML() error = %v", err)
    }
    
    if !strings.Contains(html, "John Doe") {
        t.Error("Rendered HTML missing expected data")
    }
}

func TestTemplateRenderer_RenderText(t *testing.T) {
    // Test plain text rendering
}

func TestTemplateRenderer_MissingTemplate(t *testing.T) {
    // Test error handling for missing templates
}

func TestTemplateRenderer_InvalidData(t *testing.T) {
    // Test error handling for invalid data
}

func TestTemplateRenderer_TemplateFunctions(t *testing.T) {
    tests := []struct {
        name     string
        template string
        data     interface{}
        want     string
    }{
        {
            name:     "formatDate",
            template: "{{formatDate .Date \"2006-01-02\"}}",
            data:     struct{ Date time.Time }{Date: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
            want:     "2026-01-15",
        },
        {
            name:     "title",
            template: "{{title .Text}}",
            data:     struct{ Text string }{Text: "hello world"},
            want:     "Hello World",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test template function
        })
    }
}
```

---

## Template Examples

### HTML Template (rsvp_confirmation.html)

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>RSVP Confirmation</title>
</head>
<body>
    <h1>Hi {{.GuestName}},</h1>
    <p>Your RSVP has been confirmed!</p>
    
    <h2>Event Details</h2>
    <p><strong>{{.EventTitle}}</strong></p>
    <p>{{formatDateTime .EventDate}}</p>
    <p>{{.EventLocation}}</p>
    
    <h2>Your Response</h2>
    <p>Status: {{title .Response}}</p>
    {{if gt .PlusOnes 0}}
    <p>Guests: {{.PlusOnes}}</p>
    {{end}}
</body>
</html>
```

### Text Template (rsvp_confirmation.txt)

```text
Hi {{.GuestName}},

Your RSVP has been confirmed!

EVENT DETAILS
-------------
{{.EventTitle}}
{{formatDateTime .EventDate}}
{{.EventLocation}}

YOUR RESPONSE
-------------
Status: {{title .Response}}
{{if gt .PlusOnes 0}}Guests: {{.PlusOnes}}{{end}}
```

---

## Template Functions

### Date/Time Functions
- `formatDate` - Format date with custom layout
- `formatDateTime` - Format date/time in readable format
- `formatTime` - Format time only

### String Functions
- `title` - Title case
- `upper` - Uppercase
- `lower` - Lowercase
- `truncate` - Truncate string

### Utility Functions
- `default` - Default value if nil
- `join` - Join array with separator
- `dict` - Create dictionary for nested templates

---

## Performance Considerations

- Template caching enabled by default
- Concurrent rendering safe (RWMutex)
- Template parsing done once at startup
- Reload only in development mode

---

## References

- **Epic:** [05_EPIC_email.md](05_EPIC_email.md)
- **LLD:** [lld/05_EMAIL_LLD.md](../lld/05_EMAIL_LLD.md)
- **Templates:** [`templates/email/`](../../templates/email/)
- **Go Docs:** `html/template`, `text/template`
