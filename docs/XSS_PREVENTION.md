# XSS Prevention in TinyRSVP Templates

## Overview

TinyRSVP uses Go's `html/template` package which provides automatic context-aware XSS prevention. All user input is automatically escaped based on the context where it appears in the template.

## Security Model

### Automatic Escaping

Go's `html/template` automatically escapes all template variables based on their context:

- **HTML Context**: `<div>{{.Content}}</div>` - Tags and special characters are escaped
- **Attribute Context**: `<img alt='{{.Title}}'>` - Quotes and special characters are escaped
- **URL Context**: `<a href='{{.URL}}'>` - Dangerous URLs are sanitized to `#ZgotmplZ`
- **JavaScript Context**: `<script>var x = {{.Data}};</script>` - JavaScript strings are properly escaped

### Context-Aware Escaping Examples

#### HTML Context
```go
// Input
Description: "<script>alert('xss')</script>"

// Template
<div>{{.Event.Description}}</div>

// Output (safe)
<div>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</div>
```

#### Attribute Context
```go
// Input
Title: "\" onerror=\"alert('xss')"

// Template
<img alt="{{.Event.Title}}">

// Output (safe)
<img alt="&#34; onerror=&#34;alert(&#39;xss&#39;)">
```

#### URL Context
```go
// Input
URL: "javascript:alert('xss')"

// Template
<a href="{{.RSVPURL}}">RSVP</a>

// Output (safe - dangerous URL replaced)
<a href="#ZgotmplZ">RSVP</a>
```

## Removed Dangerous Functions

The following functions have been **removed** from the template engine as they bypass XSS protection:

- `safeHTML` - Would allow unescaped HTML (removed)
- `safeURL` - Would allow dangerous URLs (removed)
- `safeCSS` - Would allow unescaped CSS (removed)

Any attempt to use these functions will result in a template parse error.

## Testing

### Comprehensive Test Coverage

XSS prevention is tested with:

1. **OWASP Top 10 XSS Vectors**: 38+ attack vectors including:
   - Script tags
   - Event handlers (onerror, onload, onclick, etc.)
   - JavaScript URLs
   - Data URLs
   - SVG payloads
   - Encoding bypasses
   - Mutation XSS
   - Polyglot payloads

2. **Context-Aware Tests**: Verify escaping in:
   - HTML context
   - Attribute context
   - URL context
   - JavaScript context

3. **Integration Tests**: End-to-end testing with:
   - All template types (invite_email, rsvp_page, confirmation_page)
   - Real-world scenarios
   - Service-level rendering

### Test Files

- `internal/templates/xss_test.go` - Unit tests for XSS prevention
- `internal/templates/xss_integration_test.go` - Integration tests
- `internal/templates/engine_test.go` - Engine-level XSS tests

## Best Practices

### Do's ✅

- Use `html/template` for all HTML rendering
- Let Go handle all escaping automatically
- Test templates with malicious input
- Use strongly-typed data structures

### Don'ts ❌

- Never use `template.HTML` type
- Never use `template.JS` type
- Never use `template.CSS` type
- Never use `template.URL` type
- Never disable auto-escaping
- Never trust user input

## Security Guarantees

When using TinyRSVP templates correctly:

1. **All user input is automatically escaped** based on context
2. **Script tags cannot execute** - they are rendered as text
3. **Event handlers cannot execute** - they are rendered as text
4. **JavaScript URLs are sanitized** - replaced with safe placeholder
5. **Data URLs are sanitized** - replaced with safe placeholder
6. **No bypass functions exist** - safeHTML/URL/CSS removed

## Verification

To verify XSS prevention is working:

```bash
# Run XSS-specific tests
go test -timeout 30s -v ./internal/templates -run TestXSSPrevention

# Run all template tests
go test -timeout 30s ./internal/templates/...
```

All tests should pass, confirming that:
- Dangerous patterns are escaped
- Context-aware escaping works
- No bypass functions exist
- Integration scenarios are safe

## References

- [Go html/template Security Model](https://pkg.go.dev/html/template#hdr-Security_Model)
- [OWASP XSS Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
- [OWASP XSS Filter Evasion](https://cheatsheetseries.owasp.org/cheatsheets/XSS_Filter_Evasion_Cheat_Sheet.html)
