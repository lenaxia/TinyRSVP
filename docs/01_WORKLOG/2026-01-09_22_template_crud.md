# Template CRUD Implementation

**Date:** 2026-01-09  
**Story:** [06_STORY_04_template_crud.md](../00_BACKLOG/06_STORY_04_template_crud.md)  
**Status:** Complete

## Summary

Implemented complete CRUD operations for template management with comprehensive testing and RBAC enforcement.

## What Was Implemented

### 1. Repository Extensions
- **File:** `internal/db/repositories/template_repository.go`
- Added `IsTemplateInUse(ctx, id)` - checks if template is used by any events
- Added `SetDefault(ctx, id)` - sets template as default for its type (transactional)
- Both methods include comprehensive unit tests

### 2. Template Service
- **File:** `internal/templates/service.go`
- **Tests:** `internal/templates/service_test.go`
- **Integration Tests:** `internal/templates/service_integration_test.go`

**Methods Implemented:**
- `CreateTemplate` - creates new template with validation and auth
- `GetTemplate` - retrieves template by ID
- `GetTemplateForEvent` - gets event-specific or default template
- `GetDefaultTemplate` - retrieves default template by type
- `UpdateTemplate` - updates template with permission checks
- `DeleteTemplate` - deletes template with safety checks
- `SetActive` - activates/deactivates template
- `SetDefault` - sets template as default (admin only)
- `ListTemplates` - lists templates with flexible filtering

**RBAC Rules:**
- Event managers can create templates
- Users can only edit/delete their own templates
- Admins can edit/delete any template
- Only admins can edit default templates
- Only admins can set templates as default

**Safety Checks:**
- Cannot delete default system templates
- Cannot delete templates in use by events
- Template validation enforced on create/update
- Proper error types for all failure scenarios

### 3. HTTP Handlers
- **File:** `internal/handlers/templates.go`
- **Tests:** `internal/handlers/templates_test.go`
- **Integration Tests:** `internal/handlers/templates_integration_test.go`

**Endpoints Implemented:**
- `POST /api/templates` - create template
- `GET /api/templates/:id` - get template
- `PUT /api/templates/:id` - update template
- `DELETE /api/templates/:id` - delete template
- `GET /api/templates` - list templates with filters
- `POST /api/templates/:id/set-active` - set active status
- `POST /api/templates/:id/set-default` - set as default

**Query Parameters for List:**
- `type` - filter by template type
- `event_id` - filter by event
- `is_default` - filter by default status
- `is_active` - filter by active status
- `limit` - pagination limit (default 50, max 100)
- `offset` - pagination offset

### 4. Error Types
- **File:** `internal/models/errors.go`
- Added `UnauthorizedError` - for authentication failures
- Added `ForbiddenError` - for permission denials

### 5. Test Helpers
- **File:** `internal/templates/test_helpers.go`
- Mock repository and validator for service tests
- Helper functions for string operations

## Test Coverage

### Unit Tests
- **Service:** 7 test functions covering all CRUD operations
- **Handlers:** 4 test functions covering all HTTP endpoints
- **Repository:** 2 new test functions for new methods

### Integration Tests
- **Service:** 8 comprehensive integration tests
  - Full CRUD flow
  - Permission enforcement
  - Cannot delete default templates
  - Cannot delete templates in use
  - Set default functionality
  - Get template for event (with fallback)
  - List templates with filters
  - Concurrent operations

- **Handlers:** 3 integration tests
  - Full CRUD flow via HTTP
  - Permission enforcement via HTTP
  - List with filters via HTTP

**All tests pass with timeout (30s)**

## Key Design Decisions

1. **Service sets CreatedBy before validation** - ensures validation passes for required field
2. **SetDefault uses transaction** - ensures atomicity when unsetting old default
3. **GetTemplateForEvent has fallback logic** - returns default if event template inactive/missing
4. **Handlers check auth at entry** - fail fast for unauthorized requests
5. **UpdateTemplate fetches existing first** - preserves fields not in update request

## Integration Points

The template service integrates with:
- `internal/db/repositories` - for data persistence
- `internal/auth` - for user context and permissions
- `internal/models` - for data structures and error types

The handlers integrate with:
- `internal/templates` - for business logic
- `github.com/go-chi/chi/v5` - for routing

## Next Steps

To complete template functionality:
1. Wire up handlers in main.go router
2. Add RBAC middleware to template routes
3. Consider adding template preview endpoint (Story 06)
4. Add template versioning/history if needed

## Files Created/Modified

**Created:**
- `internal/templates/service.go`
- `internal/templates/service_test.go`
- `internal/templates/service_integration_test.go`
- `internal/templates/test_helpers.go`
- `internal/handlers/templates.go`
- `internal/handlers/templates_test.go`
- `internal/handlers/templates_integration_test.go`

**Modified:**
- `internal/db/repositories/template_repository.go` - added IsTemplateInUse, SetDefault
- `internal/db/repositories/template_repository_test.go` - added tests for new methods
- `internal/models/errors.go` - added UnauthorizedError, ForbiddenError
- `internal/templates/seeder_test.go` - added missing interface methods to mock
- `internal/templates/README.md` - documented new service
- `docs/00_BACKLOG/06_STORY_04_template_crud.md` - marked complete

## Test Results

```
✓ All repository tests passing
✓ All service unit tests passing (7 tests)
✓ All service integration tests passing (8 tests)
✓ All handler unit tests passing (4 tests)
✓ All handler integration tests passing (3 tests)
✓ Total: 22+ new tests, all passing
```
