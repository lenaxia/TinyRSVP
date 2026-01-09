# User Story: Go html/template Integration

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **developer**, I want **Go html/template integration with automatic XSS prevention** so that **templates can be rendered safely with user-provided data**.

---

## Acceptance Criteria

- [ ] Template renderer service created
- [ ] Go html/template integrated
- [ ] Automatic HTML escaping enabled
- [ ] Template functions registered (date formatting, string operations)
- [ ] RenderHTML method implemented
- [ ] RenderText method implemented
- [ ] RenderToWriter method implemented
- [ ] Template parsing errors handled gracefully
- [ ] Template execution errors handled gracefully
- [ ] All tests pass with timeout
- [ ] XSS prevention verified

---

## Technical Details

### Renderer Interface

```go
package templates

import (
    "io"
)

type Renderer interface {
    RenderHTML(templateContent string, data interface{}) (string, error)
    RenderText(templateContent string, data interface{}) (string, error)
    RenderToWriter(w io.Writer, templateContent string, data interface{}) error
}

type renderer struct {
    funcMap template.FuncMap
}

func NewRenderer() Renderer {
    return &renderer{
        funcMap: createFuncMap(),
    }
}
```

### Template Functions

```go
func createFuncMap() template.FuncMap {
    return template.FuncMap{
        "upper":      strings.ToUpper,
        "lower":      strings.ToLower,
        "title":      strings.Title,
        "formatDate": formatDate,
        "formatTime": formatTime,
        "add":        add,
        "sub":        sub,
    }
}

func formatDate(t time.Time, layout string) string {
    return t.Format(layout)
}

func formatTime(t time.Time) string {
    return t.Format("3:04 PM")
}

func add(a, b int) int {
    return a + b
}

func sub(a, b int) int {
    return a - b
}
```

### HTML Rendering

```go
func (r *renderer) RenderHTML(templateContent string, data interface{}) (string, error) {
    tmpl, err := template.New("html").Funcs(r.funcMap).Parse(templateContent)
    if err != nil {
        return "", &TemplateParseError{
            Message: "Failed to parse HTML template",
            Err:     err,
        }
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", &TemplateExecutionError{
            Message: "Failed to execute HTML template",
            Err:     err,
        }
    }
    
    return buf.String(), nil
}
```

### Text Rendering

```go
func (r *renderer) RenderText(templateContent string, data interface{}) (string, error) {
    tmpl, err := text/template.New("text").Funcs(r.funcMap).Parse(templateContent)
    if err != nil {
        return "", &TemplateParseError{
            Message: "Failed to parse text template",
            Err:     err,
        }
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", &TemplateExecutionError{
            Message: "Failed to execute text template",
            Err:     err,
        }
    }
    
    return buf.String(), nil
}
```

### Stream Rendering

```go
func (r *renderer) RenderToWriter(w io.Writer, templateContent string, data interface{}) error {
    tmpl, err := template.New("stream").Funcs(r.funcMap).Parse(templateContent)
    if err != nil {
        return &TemplateParseError{
            Message: "Failed to parse template",
            Err:     err,
        }
    }
    
    if err := tmpl.Execute(w, data); err != nil {
        return &TemplateExecutionError{
            Message: "Failed to execute template",
            Err:     err,
        }
    }
    
    return nil
}
```

---

## Tasks

### Phase 1: Renderer Setup (TDD)
- [ ] Define Renderer interface
- [ ] Write test for NewRenderer
- [ ] Implement NewRenderer
- [ ] Write test for function map creation
- [ ] Implement template functions
- [ ] Run tests (should pass)

### Phase 2: HTML Rendering (TDD)
- [ ] Write test for RenderHTML with simple template
- [ ] Write test for RenderHTML with variables
- [ ] Write test for RenderHTML with functions
- [ ] Write test for RenderHTML with XSS attempt
- [ ] Write test for RenderHTML parse error
- [ ] Write test for RenderHTML execution error
- [ ] Implement RenderHTML
- [ ] Run tests (should pass)

### Phase 3: Text Rendering (TDD)
- [ ] Write test for RenderText with simple template
- [ ] Write test for RenderText with variables
- [ ] Write test for RenderText with functions
- [ ] Write test for RenderText parse error
- [ ] Write test for RenderText execution error
- [ ] Implement RenderText
- [ ] Run tests (should pass)

### Phase 4: Stream Rendering (TDD)
- [ ] Write test for RenderToWriter success
- [ ] Write test for RenderToWriter parse error
- [ ] Write test for RenderToWriter execution error
- [ ] Implement RenderToWriter
- [ ] Run tests (should pass)

### Phase 5: Integration Testing
- [ ] Test rendering with real event data
- [ ] Test rendering with real invite data
- [ ] Test rendering with real RSVP data
- [ ] Test XSS prevention with various payloads
- [ ] Test error handling with malformed templates
- [ ] Verify performance with large templates

---

## Template Function Reference

### String Functions
- `{{.Text | upper}}` - Convert to uppercase
- `{{.Text | lower}}` - Convert to lowercase
- `{{.Text | title}}` - Convert to title case

### Date/Time Functions
- `{{.StartTime | formatDate "Jan 2, 2006"}}` - Format date
- `{{.StartTime | formatTime}}` - Format time (3:04 PM)

### Math Functions
- `{{add .Count 1}}` - Addition
- `{{sub .Total .Used}}` - Subtraction

### Control Flow
- `{{if .Condition}}...{{end}}` - Conditional
- `{{range .Items}}...{{end}}` - Loop
- `{{with .Value}}...{{end}}` - Scope

---

## Error Handling

### Template Parse Errors

```go
type TemplateParseError struct {
    Message string
    Err     error
}

func (e *TemplateParseError) Error() string {
    return fmt.Sprintf("%s: %v", e.Message, e.Err)
}
```

### Template Execution Errors

```go
type TemplateExecutionError struct {
    Message string
    Err     error
}

func (e *TemplateExecutionError) Error() string {
    return fmt.Sprintf("%s: %v", e.Message, e.Err)
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestRenderer_RenderHTML(t *testing.T) {
    renderer := NewRenderer()
    
    tests := []struct {
        name     string
        template string
        data     interface{}
        want     string
        wantErr  bool
    }{
        {
            name:     "simple text",
            template: "<h1>Hello World</h1>",
            data:     nil,
            want:     "<h1>Hello World</h1>",
            wantErr:  false,
        },
        {
            name:     "variable substitution",
            template: "<h1>{{.Title}}</h1>",
            data:     struct{ Title string }{Title: "Test Event"},
            want:     "<h1>Test Event</h1>",
            wantErr:  false,
        },
        {
            name:     "XSS prevention",
            template: "<p>{{.Description}}</p>",
            data:     struct{ Description string }{Description: "<script>alert('xss')</script>"},
            want:     "<p>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</p>",
            wantErr:  false,
        },
        {
            name:     "function call",
            template: "<h1>{{.Title | upper}}</h1>",
            data:     struct{ Title string }{Title: "test"},
            want:     "<h1>TEST</h1>",
            wantErr:  false,
        },
        {
            name:     "parse error",
            template: "{{.Title",
            data:     nil,
            want:     "",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := renderer.RenderHTML(tt.template, tt.data)
            if (err != nil) != tt.wantErr {
                t.Errorf("RenderHTML() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("RenderHTML() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestRenderer_XSSPrevention(t *testing.T) {
    renderer := NewRenderer()
    
    xssPayloads := []string{
        "<script>alert('xss')</script>",
        "<img src=x onerror=alert('xss')>",
        "<svg onload=alert('xss')>",
        "javascript:alert('xss')",
        "<iframe src='javascript:alert(\"xss\")'></iframe>",
    }
    
    template := "<div>{{.Payload}}</div>"
    
    for _, payload := range xssPayloads {
        t.Run(payload, func(t *testing.T) {
            data := struct{ Payload string }{Payload: payload}
            result, err := renderer.RenderHTML(template, data)
            if err != nil {
                t.Fatalf("RenderHTML() error = %v", err)
            }
            
            if strings.Contains(result, "<script>") ||
               strings.Contains(result, "onerror=") ||
               strings.Contains(result, "onload=") ||
               strings.Contains(result, "javascript:") {
                t.Errorf("XSS not prevented: %s", result)
            }
        })
    }
}
```

### Integration Tests

```go
func TestRenderer_Integration(t *testing.T) {
    renderer := NewRenderer()
    
    // Test with real event data
    event := &models.Event{
        Title:       "Birthday Party",
        Description: "Join us for cake & fun!",
        StartTime:   time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC),
        Location:    "123 Main St",
    }
    
    template := `
<html>
<head><title>{{.Title}}</title></head>
<body>
    <h1>{{.Title}}</h1>
    <p>{{.Description}}</p>
    <p>When: {{.StartTime | formatDate "January 2, 2006"}} at {{.StartTime | formatTime}}</p>
    <p>Where: {{.Location}}</p>
</body>
</html>
`
    
    result, err := renderer.RenderHTML(template, event)
    if err != nil {
        t.Fatalf("RenderHTML() error = %v", err)
    }
    
    if !strings.Contains(result, "Birthday Party") {
        t.Error("Expected title in output")
    }
    
    if !strings.Contains(result, "June 15, 2026") {
        t.Error("Expected formatted date in output")
    }
}
```

---

## Dependencies

**Depends on:**
- Story 00: Template Struct (for data models)

**Blocks:**
- Story 03: Default Templates (needs renderer)
- Story 04: Template CRUD (needs renderer)
- Story 06: Template Preview (needs renderer)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Renderer interface defined
- [ ] RenderHTML implemented
- [ ] RenderText implemented
- [ ] RenderToWriter implemented
- [ ] Template functions registered
- [ ] All unit tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] XSS prevention verified
- [ ] Error handling complete
- [ ] Documentation updated
- [ ] Code reviewed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11.4 (Template Security)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md) - Section 4.1
- **Go Docs:** https://pkg.go.dev/html/template
- **Story 00:** [06_STORY_00_template_struct.md](06_STORY_00_template_struct.md)
