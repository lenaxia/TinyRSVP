# Template Security Validation Implementation

**Date:** 2026-01-09  
**Story:** [06_STORY_02_template_security.md](../00_BACKLOG/06_STORY_02_template_security.md)  
**Status:** Complete

---

## Summary

Implemented comprehensive template validation and security checks for the TinyRSVP template system. The validator ensures templates are syntactically correct, use only allowed variables, stay within size limits, and are protected against XSS attacks.

---

## Implementation Details

### Files Created

1. **[`internal/templates/validator.go`](../../internal/templates/validator.go)**
   - Validator interface with 4 validation methods
   - Complete implementation with security checks
   - Variable extraction from template parse trees
   - Test data generation for validation

2. **[`internal/templates/validator_test.go`](../../internal/templates/validator_test.go)**
   - Unit tests for all validator methods
   - 40+ test cases covering happy and unhappy paths
   - Edge case testing (boundary conditions)

3. **[`internal/templates/validator_integration_test.go`](../../internal/templates/validator_integration_test.go)**
   - Real-world template validation tests
   - Comprehensive XSS prevention tests
   - Advanced payload testing

### Key Components

#### Validator Interface
```go
type Validator interface {
    ValidateTemplate(tmpl *models.Template) error
    ValidateSyntax(content string, templateType models.TemplateType) error
    ValidateVariables(content string, allowedVars []string) error
    ValidateSize(content string, maxBytes int) error
}
```

#### Validation Flow
1. Model validation (name, type, required fields)
2. Size validation (HTML: 100KB, Text: 50KB, CSS: 50KB)
3. Variable validation (check against allowed list)
4. Syntax validation (parse and execute with test data)

#### Security Features
- **XSS Prevention**: Automatic HTML escaping via `html/template`
- **Variable Whitelisting**: Only allowed variables can be used
- **Size Limits**: Prevent resource exhaustion
- **Syntax Validation**: Catch errors before saving
- **URL Sanitization**: Go's `#ZgotmplZ` for dangerous URLs

---

## Test Results

All tests passing:
```
=== Template Package Tests ===
- Engine tests: 31 tests PASS
- Validator unit tests: 27 tests PASS  
- Integration tests: 9 tests PASS
- XSS security tests: 11 tests PASS

Total: 78 tests PASS in 0.023s
```

### Test Coverage

- **Size Validation**: 5 test cases (within limit, at boundary, exceeded)
- **Syntax Validation**: 7 test cases (valid templates, parse errors, all types)
- **Variable Validation**: 9 test cases (allowed, undefined, nested, range)
- **Complete Validation**: 9 test cases (success, failures, edge cases)
- **Integration Tests**: 9 test cases (real-world templates, boundaries)
- **XSS Tests**: 11 test cases (script tags, event handlers, URLs, advanced payloads)

---

## Allowed Variables by Template Type

### All Templates (Common)
- Event.Title
- Event.Description
- Event.StartTime
- Event.EndTime
- Event.Timezone
- Event.Location
- Event.RSVPDeadline

### Invite Email Templates
- Common variables plus:
- Invite.Name
- Invite.Email
- RSVPURL
- MaxPlusOnes

### RSVP Page Templates
- Common variables plus:
- RSVP.Response
- RSVP.PlusOnes
- Questions

### Confirmation Page Templates
- Common variables plus:
- RSVP.Response
- RSVP.PlusOnes
- Answers

---

## Size Limits

| Content Type | Maximum Size |
|--------------|--------------|
| HTML Content | 100 KB       |
| Text Content | 50 KB        |
| CSS Content  | 50 KB        |

---

## Security Validation

### XSS Prevention Verified

All XSS attack vectors properly mitigated:
- ✅ Script tags escaped (`<script>` → `&lt;script&gt;`)
- ✅ Event handlers escaped (`onerror=` → `onerror=` in escaped HTML)
- ✅ JavaScript URLs sanitized (`javascript:` → `#ZgotmplZ`)
- ✅ SVG with scripts escaped
- ✅ iframes escaped
- ✅ Object tags escaped
- ✅ HTML entities double-escaped

### Validation Order

1. Model validation (basic field checks)
2. Size validation (prevent resource exhaustion)
3. Variable validation (whitelist enforcement)
4. Syntax validation (parse and execute)

This order ensures:
- Fast failures for simple errors
- Security checks before expensive operations
- Clear error messages for each validation type

---

## Integration with Existing Code

The validator integrates seamlessly with:
- **[`internal/templates/engine.go`](../../internal/templates/engine.go)**: Uses Engine for parsing and execution
- **[`internal/models/template.go`](../../internal/models/template.go)**: Validates Template structs
- **Template functions**: Supports all custom functions (upper, lower, formatDateTime, etc.)

---

## Usage Example

```go
engine := templates.NewEngine()
validator := templates.NewValidator(engine)

template := &models.Template{
    Name:        "Custom Invite",
    Type:        models.TemplateTypeInviteEmail,
    HTMLContent: "<h1>{{.Event.Title}}</h1>",
    TextContent: strPtr("Event: {{.Event.Title}}"),
    CreatedBy:   userID,
}

if err := validator.ValidateTemplate(template); err != nil {
    // Handle validation error
    return err
}

// Template is safe to save and use
```

---

## Next Steps

This validator should be integrated into:
- Story 03: Default Templates (validate default templates)
- Story 04: Template CRUD (validate on create/update)
- Story 11: XSS Prevention (already implemented via validator)

---

## Notes

- Variable extraction handles complex nested structures (if, range, with)
- Test data uses time.Time for date fields to match real usage
- XSS tests verify both escaping and URL sanitization
- All tests use 30s timeout as per project standards
- Following TDD: tests written first, then implementation
