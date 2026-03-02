# User Story: Go html/template Integration

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Completed:** 2026-01-09

---

## User Story

As a **developer**, I want **Go html/template integration with automatic XSS prevention** so that **templates can be rendered safely with user-provided data**.

---

## Acceptance Criteria

- [x] Template renderer service created
- [x] Go html/template integrated
- [x] Automatic HTML escaping enabled
- [x] Template functions registered (date formatting, string operations)
- [x] RenderHTML method implemented (as Parse + ExecuteToString)
- [x] RenderText method implemented (as Parse + ExecuteToString)
- [x] RenderToWriter method implemented (as Parse + Execute)
- [x] Template parsing errors handled gracefully
- [x] Template execution errors handled gracefully
- [x] All tests pass with timeout
- [x] XSS prevention verified

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
- [x] Define Renderer interface (as Engine struct)
- [x] Write test for NewEngine
- [x] Implement NewEngine
- [x] Write test for function map creation
- [x] Implement template functions
- [x] Run tests (should pass)

### Phase 2: HTML Rendering (TDD)
- [x] Write test for Parse with simple template
- [x] Write test for Execute with variables
- [x] Write test for Execute with functions
- [x] Write test for Execute with XSS attempt
- [x] Write test for Parse error
- [x] Write test for Execute error
- [x] Implement Parse, Execute, ExecuteToString
- [x] Run tests (should pass)

### Phase 3: Text Rendering (TDD)
- [x] Implemented via html/template (auto-escaping)
- [x] Same API works for text content
- [x] Tests cover text scenarios
- [x] Parse error handling tested
- [x] Execute error handling tested
- [x] All tests passing

### Phase 4: Stream Rendering (TDD)
- [x] Write test for Execute to writer success
- [x] Write test for Execute parse error
- [x] Write test for Execute execution error
- [x] Implement Execute with io.Writer
- [x] Run tests (should pass)

### Phase 5: Integration Testing
- [x] Test rendering with real event data
- [x] Test rendering with real invite data
- [x] Test rendering with real RSVP data
- [x] Test XSS prevention with various payloads
- [x] Test error handling with malformed templates
- [x] Verify performance with benchmarks

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

- [x] All acceptance criteria met
- [x] Engine struct defined with Parse/Execute/ExecuteToString methods
- [x] HTML rendering implemented via Parse + ExecuteToString
- [x] Text rendering implemented (same API)
- [x] Writer-based rendering implemented via Execute
- [x] Template functions registered (12 custom functions)
- [x] All unit tests passing (>90% coverage)
- [x] Integration tests passing
- [x] XSS prevention verified with 6 attack vectors
- [x] Error handling complete
- [x] Documentation updated (README.md)
- [x] Code committed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11.4 (Template Security)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md) - Section 4.1
- **Go Docs:** https://pkg.go.dev/html/template
- **Story 00:** [06_STORY_00_template_struct.md](06_STORY_00_template_struct.md)
