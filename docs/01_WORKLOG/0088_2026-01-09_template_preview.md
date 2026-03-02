# Template Preview Implementation

**Date:** 2026-01-09  
**Story:** [06_STORY_06_template_preview.md](../00_BACKLOG/06_STORY_06_template_preview.md)  
**Status:** Complete

---

## Summary

Implemented template preview functionality allowing event managers to preview templates with sample data before saving. This enables users to verify template rendering and catch errors early in the template creation workflow.

---

## Implementation Details

### Phase 1: Test Data Generation

**Files Created:**
- [`internal/templates/preview.go`](../../internal/templates/preview.go) - Core preview functionality
- [`internal/templates/preview_test.go`](../../internal/templates/preview_test.go) - Unit tests

**Key Components:**
- `CreateTestData(templateType)` - Generates realistic test data for each template type
- Supports all three template types: invite_email, rsvp_page, confirmation_page
- Test data includes future dates, sample questions, and realistic content

### Phase 2: Preview Service

**Files Modified:**
- [`internal/templates/service.go`](../../internal/templates/service.go) - Added PreviewTemplate method

**Key Features:**
- `PreviewRequest` struct for API requests
- `PreviewResponse` struct for API responses
- Direct validation bypassing name requirements (preview doesn't save)
- Validates syntax, variables, and size limits
- Renders both HTML and optional text content
- Proper error handling with field-specific validation errors

### Phase 3: Handler Layer

**Files Modified:**
- [`internal/handlers/templates.go`](../../internal/handlers/templates.go) - Added PreviewTemplate handler
- [`internal/handlers/templates_test.go`](../../internal/handlers/templates_test.go) - Added handler tests

**API Endpoint:**
```
POST /api/templates/preview
Authorization: Required (Event Manager or Admin)
Content-Type: application/json

Request:
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
```

### Phase 4: Integration Testing

**Files Created:**
- [`internal/templates/preview_integration_test.go`](../../internal/templates/preview_integration_test.go)

**Test Coverage:**
- Preview with all template types
- All variable types (string, time, int, nested structs, arrays)
- All template functions (upper, lower, formatDateTime, truncate, default)
- Error handling for syntax errors, undefined variables, invalid types
- End-to-end workflow validation

---

## Test Results

All tests passing with timeout:
```bash
go test -timeout 30s ./internal/templates ./internal/handlers
ok  	github.com/lenaxia/tinyrsvp/internal/templates	0.144s
ok  	github.com/lenaxia/tinyrsvp/internal/handlers	0.936s
```

Full test suite:
```bash
go test -timeout 30s ./...
# All packages pass
```

---

## Key Design Decisions

1. **Validation Strategy**: Preview bypasses full template validation (which requires name) and only validates syntax, variables, and size limits

2. **Test Data Realism**: Test data uses future dates (30 days ahead) and realistic content to provide meaningful previews

3. **Error Handling**: Field-specific validation errors help users identify exactly where template issues occur

4. **No Persistence**: Preview does not save templates, allowing safe experimentation

5. **RBAC Enforcement**: Only authenticated users (event managers and admins) can preview templates

---

## Integration Points

**Depends On:**
- Story 00: Template Struct ✓
- Story 01: Template Integration ✓
- Story 02: Template Security ✓
- Story 05: Template Variables ✓

**Enables:**
- Story 04: Template CRUD (preview before save)
- Future UI template editor with live preview

---

## Files Modified

1. `internal/templates/preview.go` - New file with preview logic
2. `internal/templates/preview_test.go` - New file with unit tests
3. `internal/templates/preview_integration_test.go` - New file with integration tests
4. `internal/templates/service.go` - Added PreviewTemplate method to Service interface
5. `internal/handlers/templates.go` - Added PreviewTemplate handler and route
6. `internal/handlers/templates_test.go` - Added preview handler tests

---

## Next Steps

1. Consider adding UI template editor with live preview
2. Consider adding preview for CSS rendering
3. Consider adding preview with custom test data (user-provided)

---

## Notes

- All tests follow TDD approach (tests written first)
- Comprehensive coverage of happy paths, error cases, and edge cases
- Integration tests verify end-to-end functionality
- No technical debt introduced
- Code follows existing patterns and conventions
