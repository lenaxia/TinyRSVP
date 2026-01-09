# User Story: Event CRUD Routes

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 2 days

---

## User Story

As an **event manager**, I want **to create, read, update, and delete events via HTTP API** so that **I can manage my events through the web interface**.

---

## Acceptance Criteria

- [ ] GET /events - List events
- [ ] GET /events/new - New event form
- [ ] POST /events - Create event
- [ ] GET /events/{id} - View event details
- [ ] GET /events/{id}/edit - Edit event form
- [ ] PUT /events/{id} - Update event
- [ ] DELETE /events/{id} - Delete (archive) event
- [ ] POST /events/{id}/publish - Publish draft event
- [ ] POST /events/{id}/cancel - Cancel event
- [ ] Permission checks on all routes
- [ ] Input validation
- [ ] CSRF protection on mutations
- [ ] Error handling

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

- [ ] Implement list events handler
- [ ] Implement new event form handler
- [ ] Implement create event handler
- [ ] Implement get event handler
- [ ] Implement edit event form handler
- [ ] Implement update event handler
- [ ] Implement delete event handler
- [ ] Implement publish event handler
- [ ] Implement cancel event handler
- [ ] Add permission checks
- [ ] Add input validation
- [ ] Test all routes
- [ ] Integration tests

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

- [ ] All acceptance criteria met
- [ ] All routes implemented
- [ ] Permission checks working
- [ ] Validation working
- [ ] Tests passing
- [ ] Documentation complete
