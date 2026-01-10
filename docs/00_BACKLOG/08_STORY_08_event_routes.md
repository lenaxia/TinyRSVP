# User Story: Event CRUD Routes

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 2 days
**Actual Effort:** 1 day
**Completed:** 2026-01-10

---

## User Story

As an **event manager**, I want **to create, read, update, and delete events via HTTP API** so that **I can manage my events through the web interface**.

---

## Acceptance Criteria

- [x] GET /events - List events
- [x] GET /events/new - New event form
- [x] POST /events - Create event
- [x] GET /events/{id} - View event details
- [x] GET /events/{id}/edit - Edit event form
- [x] POST /events/{id} - Update event (using POST with form data)
- [x] POST /events/{id}/delete - Delete (archive) event
- [x] POST /events/{id}/publish - Publish draft event
- [x] POST /events/{id}/cancel - Cancel event
- [x] Permission checks on all routes
- [x] Input validation
- [x] CSRF protection on mutations
- [x] Error handling

---

## Technical Details

### Routes
```go
r.Route("/events", func(r chi.Router) {
    r.Use(RequireAuth)
    
    r.Get("/", handlers.ListEvents)
    r.Get("/new", handlers.NewEventForm)
    r.Post("/", handlers.CreateEvent)
    
    r.Route("/{id}", func(r chi.Router) {
        r.Get("/", handlers.GetEvent)
        r.Get("/edit", handlers.EditEventForm)
        r.Put("/", handlers.UpdateEvent)
        r.Delete("/", handlers.DeleteEvent)
        r.Post("/publish", handlers.PublishEvent)
        r.Post("/cancel", handlers.CancelEvent)
    })
})
```

---

## Tasks

- [x] Implement list events handler
- [x] Implement new event form handler
- [x] Implement create event handler
- [x] Implement get event handler
- [x] Implement edit event form handler
- [x] Implement update event handler
- [x] Implement delete event handler
- [x] Implement publish event handler
- [x] Implement cancel event handler
- [x] Add permission checks
- [x] Add input validation
- [x] Test all routes
- [x] Integration tests

---

## Dependencies

**Depends on:** 
- 08_STORY_00_router_setup.md
- 08_STORY_02_error_handling.md
- 08_STORY_04_csrf_protection.md
- Epic 02 (Events service)

**Blocks:** 08_STORY_09_event_ui.md

---

## Handler Examples

### List Events
```go
func (h *Handlers) ListEvents(w http.ResponseWriter, r *http.Request) {
    user := GetUser(r.Context())
    
    events, err := h.events.ListByUser(r.Context(), user.ID)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    h.templates.Render(w, "events/list.html", map[string]interface{}{
        "Events": events,
    })
}
```

### Create Event
```go
func (h *Handlers) CreateEvent(w http.ResponseWriter, r *http.Request) {
    var req CreateEventRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        HandleError(w, r, NewValidationError("Invalid request body"))
        return
    }
    
    user := GetUser(r.Context())
    req.CreatedBy = user.ID
    
    event, err := h.events.Create(r.Context(), &req)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(event)
}
```

---

## Testing Strategy

```go
func TestListEvents(t *testing.T)
func TestCreateEvent_Success(t *testing.T)
func TestCreateEvent_Validation(t *testing.T)
func TestGetEvent_Success(t *testing.T)
func TestGetEvent_NotFound(t *testing.T)
func TestUpdateEvent_Success(t *testing.T)
func TestUpdateEvent_PermissionDenied(t *testing.T)
func TestDeleteEvent_Success(t *testing.T)
func TestPublishEvent_Success(t *testing.T)
func TestCancelEvent_Success(t *testing.T)
```

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **Events Epic:** [02_EPIC_events.md](02_EPIC_events.md)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All routes implemented
- [x] Permission checks working
- [x] Validation working
- [x] Tests passing
- [x] Documentation complete

## Implementation Notes

**Files Created:**
- `internal/handlers/events_web.go` - Web UI handlers for event management
- `internal/handlers/events_web_test.go` - Unit tests for web handlers
- `internal/handlers/events_web_integration_test.go` - Integration tests

**Files Modified:**
- `internal/handlers/router.go` - Added EventWebHandlerInterface and web routes at /events
- `internal/handlers/router_docs.go` - Updated documentation with web UI routes

**Key Features:**
- Web UI routes at `/events` (separate from API routes at `/api/events`)
- HTML form rendering with CSRF token injection
- Form data parsing and validation
- Permission enforcement on all routes
- Proper error handling with content negotiation
- Full lifecycle support (create, edit, publish, cancel, delete)
- Comprehensive test coverage (unit + integration)

**Route Structure:**
```
GET  /events              - List events page
GET  /events/new          - New event form
POST /events              - Create event from form
GET  /events/{id}         - View event details
GET  /events/{id}/edit    - Edit event form
POST /events/{id}         - Update event from form
POST /events/{id}/publish - Publish event
POST /events/{id}/cancel  - Cancel event (requires reason)
POST /events/{id}/delete  - Delete event
```

**Testing:**
- 9 test functions with multiple test cases each
- Full CRUD flow integration test
- Permission enforcement integration test
- Router integration test
- CSRF protection integration test
- Form data parsing tests
- All tests passing ✅
