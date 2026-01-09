# User Story: XSS Prevention in Templates

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** Critical
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **security-conscious administrator**, I want **comprehensive XSS prevention in templates** so that **malicious user input cannot execute scripts or compromise guest browsers**.

---

## Acceptance Criteria

- [ ] Go html/template auto-escaping enabled
- [ ] All user input automatically escaped
- [ ] No use of template.HTML type (disables escaping)
- [ ] Script tags escaped in output
- [ ] Event handlers escaped in output
- [ ] JavaScript URLs escaped in output
- [ ] Data URLs sanitized
- [ ] SVG payloads escaped
- [ ] Security tests verify all XSS vectors blocked
- [ ] All tests pass with timeout
- [ ] Penetration testing completed

---

## Technical Details

### XSS Prevention Strategy

**Primary Defense:** Go html/template automatic escaping

```go
import "html/template"

func (r *renderer) RenderHTML(templateContent string, data interface{}) (string, error) {
    tmpl, err := template.New("html").Parse(templateContent)
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}
```

**Key Points:**
- Use `html/template` NOT `text/template` for HTML
- Never use `template.HTML` type
- Never use `template.JS` type
- Never use `template.CSS` type
- All variables automatically escaped

### Context-Aware Escaping

Go html/template provides context-aware escaping:

```html
<!-- HTML context: < > & " ' escaped -->
<div>{{.Event.Description}}</div>

<!-- Attribute context: quotes escaped -->
<img alt="{{.Event.Title}}">

<!-- URL context: URL encoding applied -->
<a href="/event/{{.Event.ID}}">Link</a>

<!-- JavaScript context: JSON encoding applied -->
<script>var title = {{.Event.Title}};</script>

<!-- CSS context: CSS escaping applied -->
<style>.title::before { content: "{{.Event.Title}}"; }</style>
```

### Escaping Rules

| Context | Input | Output |
|---------|-------|--------|
| HTML | `<script>alert('xss')</script>` | `&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;` |
| Attribute | `" onload="alert('xss')` | `&#34; onload=&#34;alert(&#39;xss&#39;)` |
| URL | `javascript:alert('xss')` | `#ZgotmplZ` (safe placeholder) |
| JavaScript | `'; alert('xss'); //` | `\u0027; alert(\u0027xss\u0027); //` |

---

## XSS Attack Vectors

### Vector 1: Script Tags

**Attack:**
```html
<script>alert('xss')</script>
```

**Prevention:**
```go
// Input in template variable
data.Event.Description = "<script>alert('xss')</script>"

// Template
<div>{{.Event.Description}}</div>

// Output (escaped)
<div>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</div>
```

### Vector 2: Event Handlers

**Attack:**
```html
<img src=x onerror=alert('xss')>
```

**Prevention:**
```go
// Input
data.Event.Location = "<img src=x onerror=alert('xss')>"

// Template
<p>{{.Event.Location}}</p>

// Output (escaped)
<p>&lt;img src=x onerror=alert(&#39;xss&#39;)&gt;</p>
```

### Vector 3: JavaScript URLs

**Attack:**
```html
javascript:alert('xss')
```

**Prevention:**
```go
// Input
data.RSVPURL = "javascript:alert('xss')"

// Template
<a href="{{.RSVPURL}}">RSVP</a>

// Output (sanitized)
<a href="#ZgotmplZ">RSVP</a>
```

### Vector 4: Data URLs

**Attack:**
```html
data:text/html,<script>alert('xss')</script>
```

**Prevention:**
```go
// Input
data.ImageURL = "data:text/html,<script>alert('xss')</script>"

// Template
<img src="{{.ImageURL}}">

// Output (sanitized)
<img src="#ZgotmplZ">
```

### Vector 5: SVG Payloads

**Attack:**
```html
<svg onload=alert('xss')>
```

**Prevention:**
```go
// Input
data.Event.Description = "<svg onload=alert('xss')>"

// Template
<div>{{.Event.Description}}</div>

// Output (escaped)
<div>&lt;svg onload=alert(&#39;xss&#39;)&gt;</div>
```

---

## Tasks

### Phase 1: Escaping Verification (TDD)
- [ ] Write test for script tag escaping
- [ ] Write test for event handler escaping
- [ ] Write test for JavaScript URL escaping
- [ ] Write test for data URL escaping
- [ ] Write test for SVG payload escaping
- [ ] Write test for attribute context escaping
- [ ] Write test for URL context escaping
- [ ] Verify all tests pass

### Phase 2: Context-Aware Escaping (TDD)
- [ ] Write test for HTML context
- [ ] Write test for attribute context
- [ ] Write test for URL context
- [ ] Write test for JavaScript context
- [ ] Write test for CSS context
- [ ] Verify context-aware escaping works
- [ ] Run tests (should pass)

### Phase 3: Dangerous Type Prevention (TDD)
- [ ] Write test rejecting template.HTML usage
- [ ] Write test rejecting template.JS usage
- [ ] Write test rejecting template.CSS usage
- [ ] Write test rejecting template.URL usage
- [ ] Implement type checking
- [ ] Run tests (should pass)

### Phase 4: Security Testing
- [ ] Test OWASP Top 10 XSS vectors
- [ ] Test polyglot payloads
- [ ] Test encoding bypass attempts
- [ ] Test mutation XSS (mXSS)
- [ ] Test DOM-based XSS
- [ ] Document all tested vectors

### Phase 5: Penetration Testing
- [ ] Manual testing with XSS payloads
- [ ] Test in multiple browsers
- [ ] Test with real email clients
- [ ] Verify no script execution
- [ ] Document findings

---

## XSS Test Vectors

### OWASP XSS Cheat Sheet Vectors

```go
var xssVectors = []string{
    // Basic vectors
    "<script>alert('xss')</script>",
    "<img src=x onerror=alert('xss')>",
    "<svg onload=alert('xss')>",
    
    // Event handlers
    "<body onload=alert('xss')>",
    "<input onfocus=alert('xss') autofocus>",
    "<select onfocus=alert('xss') autofocus>",
    
    // JavaScript URLs
    "javascript:alert('xss')",
    "jAvAsCrIpT:alert('xss')",
    "&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;alert('xss')",
    
    // Data URLs
    "data:text/html,<script>alert('xss')</script>",
    "data:text/html;base64,PHNjcmlwdD5hbGVydCgneHNzJyk8L3NjcmlwdD4=",
    
    // SVG vectors
    "<svg><script>alert('xss')</script></svg>",
    "<svg><animate onbegin=alert('xss') attributeName=x dur=1s>",
    
    // Encoding bypasses
    "<IMG SRC=&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;alert('xss')>",
    "<IMG SRC=&#x6A;&#x61;&#x76;&#x61;&#x73;&#x63;&#x72;&#x69;&#x70;&#x74;&#x3A;alert('xss')>",
    
    // Mutation XSS
    "<noscript><p title=\"</noscript><img src=x onerror=alert('xss')\">",
    
    // Polyglot
    "jaVasCript:/*-/*`/*\\`/*'/*\"/**/(/* */oNcliCk=alert() )//%0D%0A%0d%0a//</stYle/</titLe/</teXtarEa/</scRipt/--!>\\x3csVg/<sVg/oNloAd=alert()//\\x3e",
}
```

### Testing Each Vector

```go
func TestXSSPrevention_AllVectors(t *testing.T) {
    renderer := NewRenderer()
    
    template := "<div>{{.Payload}}</div>"
    
    for i, vector := range xssVectors {
        t.Run(fmt.Sprintf("vector_%d", i), func(t *testing.T) {
            data := struct{ Payload string }{Payload: vector}
            
            result, err := renderer.RenderHTML(template, data)
            if err != nil {
                t.Fatalf("RenderHTML() error = %v", err)
            }
            
            dangerousPatterns := []string{
                "<script",
                "javascript:",
                "onerror=",
                "onload=",
                "onclick=",
                "onfocus=",
                "onbegin=",
            }
            
            resultLower := strings.ToLower(result)
            for _, pattern := range dangerousPatterns {
                if strings.Contains(resultLower, pattern) {
                    t.Errorf("XSS not prevented: found %s in output: %s", pattern, result)
                }
            }
        })
    }
}
```

---

## Security Best Practices

### Do's ✅
- Use `html/template` for HTML rendering
- Use `text/template` for plain text
- Let Go handle all escaping automatically
- Validate template syntax before saving
- Test with comprehensive XSS vectors

### Don'ts ❌
- Never use `template.HTML` type
- Never use `template.JS` type
- Never use `template.CSS` type
- Never use `template.URL` type
- Never disable auto-escaping
- Never trust user input

### Template Safety Checklist
- [ ] Using html/template package
- [ ] No template.HTML usage
- [ ] No template.JS usage
- [ ] No template.CSS usage
- [ ] All variables auto-escaped
- [ ] XSS tests passing
- [ ] Security review completed

---

## Error Handling

| Error Condition | Error Type | Message |
|----------------|------------|---------|
| Dangerous type detected | `ValidationError` | "Template uses unsafe type: template.HTML" |
| Script tag in output | `SecurityError` | "XSS detected in rendered output" |
| Unescaped user input | `SecurityError` | "User input not properly escaped" |

---

## Testing Strategy

### Unit Tests

```go
func TestXSSPrevention_BasicVectors(t *testing.T) {
    renderer := NewRenderer()
    
    tests := []struct {
        name     string
        template string
        data     interface{}
        forbidden []string
    }{
        {
            name:     "script tag",
            template: "<div>{{.Input}}</div>",
            data:     struct{ Input string }{Input: "<script>alert('xss')</script>"},
            forbidden: []string{"<script>"},
        },
        {
            name:     "event handler",
            template: "<img alt='{{.Input}}'>",
            data:     struct{ Input string }{Input: "x' onerror='alert(1)"},
            forbidden: []string{"onerror="},
        },
        {
            name:     "javascript url",
            template: "<a href='{{.URL}}'>Link</a>",
            data:     struct{ URL string }{URL: "javascript:alert(1)"},
            forbidden: []string{"javascript:"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := renderer.RenderHTML(tt.template, tt.data)
            if err != nil {
                t.Fatalf("RenderHTML() error = %v", err)
            }
            
            for _, pattern := range tt.forbidden {
                if strings.Contains(strings.ToLower(result), strings.ToLower(pattern)) {
                    t.Errorf("Found forbidden pattern %s in output: %s", pattern, result)
                }
            }
        })
    }
}
```

### Integration Tests

```go
func TestXSSPrevention_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    service := setupTemplateService(t, db)
    
    event := &models.Event{
        Title:       "<script>alert('xss')</script>",
        Description: "<img src=x onerror=alert('xss')>",
        Location:    "javascript:alert('xss')",
    }
    
    template := &models.Template{
        Type:        models.TemplateTypeRSVPPage,
        HTMLContent: "<h1>{{.Event.Title}}</h1><p>{{.Event.Description}}</p><a href='{{.Event.Location}}'>Map</a>",
    }
    
    data := BuildRSVPPageData(event, nil, nil, "token")
    
    result, err := service.RenderTemplate(context.Background(), template, data)
    if err != nil {
        t.Fatalf("RenderTemplate() error = %v", err)
    }
    
    dangerousPatterns := []string{
        "<script>",
        "onerror=",
        "javascript:",
    }
    
    for _, pattern := range dangerousPatterns {
        if strings.Contains(result, pattern) {
            t.Errorf("XSS not prevented: found %s in output", pattern)
        }
    }
}
```

---

## XSS Prevention Mechanisms

### 1. Automatic HTML Escaping

```go
// Template
<div>{{.Event.Description}}</div>

// Input
Description: "<script>alert('xss')</script>"

// Output
<div>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</div>
```

### 2. Attribute Escaping

```go
// Template
<img alt="{{.Event.Title}}">

// Input
Title: "\" onerror=\"alert('xss')"

// Output
<img alt="&#34; onerror=&#34;alert(&#39;xss&#39;)">
```

### 3. URL Sanitization

```go
// Template
<a href="{{.RSVPURL}}">RSVP</a>

// Input
RSVPURL: "javascript:alert('xss')"

// Output
<a href="#ZgotmplZ">RSVP</a>
```

### 4. JavaScript Context Escaping

```go
// Template
<script>var title = {{.Event.Title}};</script>

// Input
Title: "'; alert('xss'); //"

// Output
<script>var title = "\u0027; alert(\u0027xss\u0027); //";</script>
```

---

## Tasks

### Phase 1: Escaping Verification (TDD)
- [ ] Write test for HTML escaping
- [ ] Write test for attribute escaping
- [ ] Write test for URL sanitization
- [ ] Write test for JavaScript escaping
- [ ] Write test for CSS escaping
- [ ] Verify all contexts handled
- [ ] Run tests (should pass)

### Phase 2: Attack Vector Testing (TDD)
- [ ] Write test for each OWASP vector
- [ ] Write test for script tags
- [ ] Write test for event handlers
- [ ] Write test for JavaScript URLs
- [ ] Write test for data URLs
- [ ] Write test for SVG payloads
- [ ] Write test for encoding bypasses
- [ ] Write test for mutation XSS
- [ ] Write test for polyglot payloads
- [ ] Run tests (should pass)

### Phase 3: Type Safety (TDD)
- [ ] Write test rejecting template.HTML
- [ ] Write test rejecting template.JS
- [ ] Write test rejecting template.CSS
- [ ] Write test rejecting template.URL
- [ ] Implement type checking
- [ ] Run tests (should pass)

### Phase 4: Integration Testing
- [ ] Test with real event data
- [ ] Test with malicious input
- [ ] Test all template types
- [ ] Test in multiple browsers
- [ ] Test in email clients
- [ ] Verify no script execution

### Phase 5: Security Audit
- [ ] Review all template usage
- [ ] Verify no template.HTML usage
- [ ] Verify all user input escaped
- [ ] Document security measures
- [ ] Create security checklist

---

## Security Testing

### Comprehensive XSS Test Suite

```go
func TestXSSPrevention_Comprehensive(t *testing.T) {
    renderer := NewRenderer()
    
    contexts := []struct {
        name     string
        template string
        field    string
    }{
        {"html", "<div>{{.Input}}</div>", "Input"},
        {"attribute", "<img alt='{{.Input}}'>", "Input"},
        {"url", "<a href='{{.Input}}'>Link</a>", "Input"},
        {"javascript", "<script>var x = {{.Input}};</script>", "Input"},
    }
    
    for _, ctx := range contexts {
        t.Run(ctx.name, func(t *testing.T) {
            for i, vector := range xssVectors {
                t.Run(fmt.Sprintf("vector_%d", i), func(t *testing.T) {
                    data := map[string]interface{}{
                        ctx.field: vector,
                    }
                    
                    result, err := renderer.RenderHTML(ctx.template, data)
                    if err != nil {
                        t.Fatalf("RenderHTML() error = %v", err)
                    }
                    
                    if containsDangerousPattern(result) {
                        t.Errorf("XSS not prevented in %s context: %s", ctx.name, result)
                    }
                })
            }
        })
    }
}

func containsDangerousPattern(html string) bool {
    dangerous := []string{
        "<script",
        "javascript:",
        "onerror=",
        "onload=",
        "onclick=",
        "onfocus=",
        "onmouseover=",
        "data:text/html",
    }
    
    htmlLower := strings.ToLower(html)
    for _, pattern := range dangerous {
        if strings.Contains(htmlLower, pattern) {
            return true
        }
    }
    
    return false
}
```

---

## Dependencies

**Depends on:**
- Story 01: Template Integration (for renderer)
- Story 02: Template Security (for validation)

**Blocks:**
- Story 03: Default Templates (must be XSS-safe)
- Story 04: Template CRUD (must enforce XSS prevention)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Auto-escaping verified
- [ ] All XSS vectors tested
- [ ] All unit tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] Security tests passing
- [ ] Penetration testing completed
- [ ] No dangerous types used
- [ ] Documentation updated
- [ ] Security review completed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11.4 (Template Security), Section 16.3 (Input Sanitization)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **Go Docs:** https://pkg.go.dev/html/template#hdr-Security_Model
- **OWASP:** https://cheatsheetseries.owasp.org/cheatsheets/XSS_Filter_Evasion_Cheat_Sheet.html
- **Story 01:** [06_STORY_01_template_integration.md](06_STORY_01_template_integration.md)
