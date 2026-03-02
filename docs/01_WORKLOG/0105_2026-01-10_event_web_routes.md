# Event Web UI Routes Implementation

**Date:** 2026-01-10  
**Story:** [08_STORY_08_event_routes.md](../00_BACKLOG/08_STORY_08_event_routes.md)  
**Status:** ✅ Complete

---

## Summary

Implemented web UI routes for event management at `/events`, providing HTML form-based interfaces for creating, editing, viewing, and managing events. This complements the existing JSON API routes at `/api/events`.

---

## What Was Implemented

### New Files Created

1. **`internal/handlers/events_web.go`** (454 lines)
   - `EventWebHandlers` struct with template support
   - `ListEventsPage` - Renders event list with filters
   - `NewEventForm` - Renders new event creation form
   - `EditEventForm` - Renders event edit form
   - `GetEventPage` - Renders event detail view
   - `CreateEventFromForm` - Handles form submission for new events
   - `UpdateEventFromForm` - Handles form submission for event updates
   - `PublishEventAction` - Publishes draft events
   - `CancelEventAction` - Cancels events with reason
   - `DeleteEventAction` - Archives events
   - `parseEventFormData` - Parses and validates form data

2. **`internal/handlers/events_web_test.go`** (1,024 lines)
   - `TestNewEventWebHandlers` - Constructor test
   - `TestEventWebHandlers_ListEventsPage` - List page rendering tests
   - `TestEventWebHandlers_NewEventForm` - New form rendering tests
   - `TestEventWebHandlers_EditEventForm` - Edit form rendering tests
   - `TestEventWebHandlers_GetEventPage` - Detail page rendering tests
   - `TestEventWebHandlers_CreateEventFromForm` - Form submission tests
   - `TestEventWebHandlers_UpdateEventFromForm` - Update form tests
   - `TestEventWebHandlers_PublishEventAction` - Publish action tests
   - `TestEventWebHandlers_CancelEventAction` - Cancel action tests
   - `TestEventWebHandlers_DeleteEventAction` - Delete action tests
   - `TestEventWebHandlers_FormDataParsing` - Form parsing tests

3. **`internal/handlers/events_web_integration_test.go`** (569 lines)
   - `TestEventWebHandlers_FullWebUIFlow_Integration` - Complete lifecycle test
   - `TestEventWebHandlers_PermissionEnforcement_Integration` - Permission tests
   - `TestEventWebHandlers_RouterIntegration` - Route registration tests
   - `TestEventWebHandlers_CSRFProtection_Integration` - CSRF validation tests

### Files Modified

1. **`internal/handlers/router.go`**
   - Added `EventWebHandlerInterface` with 9 methods
   - Added `EventWebHandlers` field to `RouterHandlers` struct
   - Integrated web routes at `/events` with authentication middleware
   - Routes support GET (forms/views) and POST (actions)

2. **`internal/handlers/router_docs.go`**
   - Added "Web UI Event Management" section
   - Documented all 9 web routes
   - Distinguished from API routes

3. **`docs/00_BACKLOG/08_STORY_08_event_routes.md`**
   - Marked all acceptance criteria complete
   - Marked all tasks complete
   - Updated status to Complete
   - Added implementation notes

---

## Route Structure

### Web UI Routes (HTML)
```
GET  /events              → ListEventsPage (renders event_list.html)
GET  /events/new          → NewEventForm (renders event_form.html)
POST /events              → CreateEventFromForm (form data → redirect)
GET  /events/{id}         → GetEventPage (renders event_detail.html)
GET  /events/{id}/edit    → EditEventForm (renders event_form.html)
POST /events/{id}         → UpdateEventFromForm (form data → redirect)
POST /events/{id}/publish → PublishEventAction (action → redirect)
POST /events/{id}/cancel  → CancelEventAction (form data → redirect)
POST /events/{id}/delete  → DeleteEventAction (action → redirect)
```

### API Routes (JSON) - Already Existed
```
GET    /api/events           → ListEvents
POST   /api/events           → CreateEvent
GET    /api/events/{id}      → GetEvent
PUT    /api/events/{id}      → UpdateEvent
DELETE /api/events/{id}      → DeleteEvent
POST   /api/events/{id}/publish → PublishEvent
POST   /api/events/{id}/cancel  → CancelEvent
```

---

## Key Design Decisions

### 1. Separate Web and API Handlers

**Rationale:** Web UI and API have different concerns:
- Web UI: Form data, HTML templates, redirects, CSRF tokens
- API: JSON payloads, status codes, content negotiation

**Implementation:** Created separate `EventWebHandlers` alongside existing `EventHandlers`

### 2. Form Data vs JSON

**Web Routes:**
- Accept: `application/x-www-form-urlencoded`
- Parse with `r.ParseForm()` and `r.FormValue()`
- Return: HTML or redirects (303 See Other)

**API Routes:**
- Accept: `application/json`
- Parse with `json.NewDecoder(r.Body).Decode()`
- Return: JSON responses

### 3. CSRF Token Injection

**Implementation:**
- Extract token from context: `middleware.GetCSRFToken(r.Context())`
- Pass to template data structs
- Templates render: `<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">`
- Middleware validates on POST requests

### 4. HTTP Method Override

**Note:** Story originally specified `PUT /events/{id}` but web forms only support GET/POST.

**Solution:** Use `POST /events/{id}` for updates (standard HTML form pattern)

Alternative considered but not implemented:
- Method override via `_method` hidden field
- JavaScript to convert to PUT

**Decision:** Keep it simple with POST for web UI. API still uses PUT.

### 5. Redirect Pattern

**After Mutations:**
- Create → Redirect to `/events/{id}` (view new event)
- Update → Redirect to `/events/{id}` (view updated event)
- Publish → Redirect to `/events/{id}` (view published event)
- Cancel → Redirect to `/events/{id}` (view cancelled event)
- Delete → Redirect to `/events` (back to list)

**Status Code:** `303 See Other` (POST-Redirect-GET pattern)

---

## Testing Strategy

### Unit Tests (events_web_test.go)

**Coverage:**
- Constructor validation
- Each handler with multiple scenarios
- Happy paths and error paths
- Permission enforcement
- Input validation
- CSRF token injection
- Form data parsing

**Test Count:** 9 test functions, 40+ test cases

### Integration Tests (events_web_integration_test.go)

**Coverage:**
- Full CRUD lifecycle with real database
- Permission enforcement across users
- Router integration (all routes registered)
- CSRF protection validation

**Test Count:** 4 integration tests

**All Tests Passing:** ✅

---

## Permission Model

**List Events:**
- Requires: Event Manager or Admin role
- Admins see all events
- Managers see only their own events

**Create Event:**
- Requires: Event Manager or Admin role
- Creator automatically set from context

**View Event:**
- Requires: Event Manager or Admin role
- Must be creator or admin

**Edit Event:**
- Requires: Event Manager or Admin role
- Must be creator or admin

**Delete Event:**
- Requires: Event Manager or Admin role
- Must be creator or admin

**Publish/Cancel Event:**
- Requires: Event Manager or Admin role
- Must be creator or admin
- Validates state transitions

---

## Form Data Handling

### Required Fields
- `title` (3-200 characters)
- `start_time` (datetime-local format: YYYY-MM-DDTHH:MM)
- `timezone` (IANA timezone string)

### Optional Fields
- `description` (max 5000 characters)
- `location` (max 500 characters)
- `end_time` (datetime-local format)
- `rsvp_deadline` (datetime-local format)
- `max_plus_ones` (0-10, default 0)

### Special Fields
- `csrf_token` (required for all POST requests)
- `version` (required for updates, optimistic locking)
- `reason` (required for cancel action, 10-500 characters)

---

## Integration with Existing Code

### Templates
- Uses existing `templates/web/event_list.html`
- Uses existing `templates/web/event_form.html`
- Needs new `templates/web/event_detail.html` (fallback rendering provided)

### Services
- Uses existing `events.Service` interface
- All business logic in service layer
- Handlers are thin wrappers

### Middleware
- CSRF protection via `middleware.CSRF()`
- Authentication via `AuthMiddleware.RequireAuth()`
- Request ID, logging, security headers all applied

### Error Handling
- Uses centralized `HandleError(w, r, err)`
- Content negotiation (HTML vs JSON)
- Proper HTTP status codes

---

## What's Next

### Immediate Follow-ups
None required - story is complete.

### Future Enhancements
1. Create `templates/web/event_detail.html` for better event detail view
2. Add client-side form validation (JavaScript)
3. Add loading states for form submissions
4. Add success flash messages after actions
5. Add confirmation dialogs for destructive actions (cancel, delete)

### Related Stories
- **Blocks:** 08_STORY_09_event_ui.md (can now proceed with enhanced UI)
- **Depends on:** All dependencies satisfied

---

## Verification

### Test Results
```bash
$ go test -timeout 30s ./internal/handlers -run TestEventWeb
PASS
ok  	github.com/lenaxia/tinyrsvp/internal/handlers	0.051s
```

### All Handler Tests
```bash
$ go test -timeout 30s ./internal/handlers/...
ok  	github.com/lenaxia/tinyrsvp/internal/handlers	1.014s
```

---

## Notes

1. **CSRF Protection:** All mutation endpoints (POST) require valid CSRF tokens
2. **Permission Checks:** Enforced at service layer, not handler layer
3. **Form Parsing:** Handles both present-but-empty and missing fields correctly
4. **Redirects:** Follow POST-Redirect-GET pattern to prevent double submissions
5. **Templates:** Fallback rendering provided when templates not loaded
6. **Error Handling:** Consistent with rest of application (uses HandleError)

---

## Lessons Learned

1. **Separation of Concerns:** Keeping web UI and API handlers separate makes code cleaner
2. **Template Injection:** Passing templates to handlers enables better testing
3. **Form Data Parsing:** Need careful handling of optional fields (nil vs empty string)
4. **Integration Testing:** Real database tests catch issues unit tests miss
5. **CSRF Tokens:** Must be injected into all forms and validated on submission
