# User Story: Event HTTP Handlers

**Epic:** [02_EPIC_events.md](02_EPIC_events.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 6 hours
**Completed:** 2026-01-07

---

## User Story

As an **event manager**, I want **HTTP endpoints for event operations** so that **I can create, view, update, and manage events through the API**.

---

## Acceptance Criteria

- [x] POST /api/events - Create event
- [x] GET /api/events/:id - Get event by ID
- [x] PUT /api/events/:id - Update event
- [x] DELETE /api/events/:id - Delete event
- [x] GET /api/events - List events with filters
- [x] POST /api/events/:id/publish - Publish event
- [x] POST /api/events/:id/cancel - Cancel event
- [x] All endpoints require authentication
- [x] All endpoints enforce permissions
- [x] Request validation working
- [x] Error responses are consistent
- [x] All tests pass with timeout

---

## Technical Details

### API Endpoints

```
POST   /api/events              Create new event
GET    /api/events              List events (with filters)
GET    /api/events/:id          Get event by ID
PUT    /api/events/:id          Update event
DELETE /api/events/:id          Delete event (soft delete)
POST   /api/events/:id/publish  Publish draft event
POST   /api/events/:id/cancel   Cancel published event
```

### Request/Response Models

```go
type CreateEventRequest struct {
    Title        string     `json:"title" validate:"required,min=3,max=200"`
    Description  *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
    StartTime    time.Time  `json:"start_time" validate:"required"`
    EndTime      *time.Time `json:"end_time,omitempty"`
    Timezone     string     `json:"timezone" validate:"required"`
    Location     *string    `json:"location,omitempty" validate:"omitempty,max=500"`
    MaxPlusOnes  int        `json:"max_plus_ones" validate:"min=0,max=10"`
    RSVPDeadline *time.Time `json:"rsvp_deadline,omitempty"`
}

type UpdateEventRequest struct {
    Title        *string    `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
    Description  *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
    StartTime    *time.Time `json:"start_time,omitempty"`
    EndTime      *time.Time `json:"end_time,omitempty"`
    Timezone     *string    `json:"timezone,omitempty"`
    Location     *string    `json:"location,omitempty" validate:"omitempty,max=500"`
    MaxPlusOnes  *int       `json:"max_plus_ones,omitempty" validate:"omitempty,min=0,max=10"`
    RSVPDeadline *time.Time `json:"rsvp_deadline,omitempty"`
    Version      int        `json:"version" validate:"required"`
}

type CancelEventRequest struct {
    Reason string `json:"reason" validate:"required,min=10,max=500"`
}

type EventResponse struct {
    ID           int64              `json:"id"`
    Title        string             `json:"title"`
    Description  *string            `json:"description,omitempty"`
    StartTime    time.Time          `json:"start_time"`
    EndTime      *time.Time         `json:"end_time,omitempty"`
    Timezone     string             `json:"timezone"`
    Location     *string            `json:"location,omitempty"`
    Status       models.EventStatus `json:"status"`
    CreatedBy    int64              `json:"created_by"`
    Version      int                `json:"version"`
    MaxPlusOnes  int                `json:"max_plus_ones"`
    RSVPDeadline *time.Time         `json:"rsvp_deadline,omitempty"`
    CreatedAt    time.Time          `json:"created_at"`
    UpdatedAt    time.Time          `json:"updated_at"`
}

type ListEventsResponse struct {
    Events []*EventResponse `json:"events"`
    Total  int              `json:"total"`
    Limit  int              `json:"limit"`
    Offset int              `json:"offset"`
}
```

### Handler Structure

```go
type EventHandlers struct {
    service events.Service
}

func NewEventHandlers(service events.Service) *EventHandlers {
    return &EventHandlers{service: service}
}

func (h *EventHandlers) RegisterRoutes(r chi.Router) {
    r.Route("/api/events", func(r chi.Router) {
        r.Use(middleware.RequireAuth)
        
        r.Post("/", h.CreateEvent)
        r.Get("/", h.ListEvents)
        
        r.Route("/{id}", func(r chi.Router) {
            r.Get("/", h.GetEvent)
            r.Put("/", h.UpdateEvent)
            r.Delete("/", h.DeleteEvent)
            r.Post("/publish", h.PublishEvent)
            r.Post("/cancel", h.CancelEvent)
        })
    })
}
```

---

## Tasks

### Phase 1: Handler Setup (TDD)
- [x] Write test for handler constructor
- [x] Write test for route registration
- [x] Implement NewEventHandlers
- [x] Implement RegisterRoutes
- [x] Run tests (should pass)

### Phase 2: Create Event Handler (TDD)
- [x] Write test for valid create request
- [x] Write test for invalid JSON
- [x] Write test for validation errors
- [x] Write test for missing required fields
- [x] Write test for unauthorized user
- [x] Write test for service error
- [x] Implement CreateEvent handler
- [x] Run tests (should pass)

### Phase 3: Get Event Handler (TDD)
- [x] Write test for getting existing event
- [x] Write test for invalid event ID
- [x] Write test for non-existent event
- [x] Write test for unauthorized access
- [x] Implement GetEvent handler
- [x] Run tests (should pass)

### Phase 4: Update Event Handler (TDD)
- [x] Write test for valid update
- [x] Write test for partial update
- [x] Write test for invalid JSON
- [x] Write test for validation errors
- [x] Write test for version conflict
- [x] Write test for unauthorized update
- [x] Implement UpdateEvent handler
- [x] Run tests (should pass)

### Phase 5: List Events Handler (TDD)
- [x] Write test for listing all events
- [x] Write test for filtering by status
- [x] Write test for filtering by creator
- [x] Write test for pagination
- [x] Write test for invalid query params
- [x] Implement ListEvents handler
- [x] Run tests (should pass)

### Phase 6: Delete Event Handler (TDD)
- [x] Write test for deleting own event
- [x] Write test for deleting as admin
- [x] Write test for unauthorized delete
- [x] Write test for non-existent event
- [x] Implement DeleteEvent handler
- [x] Run tests (should pass)

### Phase 7: Lifecycle Handlers (TDD)
- [x] Write test for publishing draft event
- [x] Write test for publishing invalid state
- [x] Write test for unauthorized publish
- [x] Write test for cancelling with reason
- [x] Write test for cancelling invalid state
- [x] Write test for unauthorized cancel
- [x] Implement PublishEvent handler
- [x] Implement CancelEvent handler
- [x] Run tests (should pass)

### Phase 8: Integration Tests
- [x] Write integration test for full CRUD flow
- [x] Write integration test for lifecycle transitions
- [x] Write integration test for permission enforcement
- [x] Run integration tests

---

## Testing Requirements

### Unit Tests

```go
func TestEventHandlers_CreateEvent(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        user       *models.User
        setupMock  func(*events.MockService)
        wantStatus int
        wantBody   string
    }{
        {
            name: "valid event creation",
            body: `{
                "title": "Birthday Party",
                "start_time": "2026-06-15T14:00:00-07:00",
                "timezone": "America/Los_Angeles",
                "max_plus_ones": 2
            }`,
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            setupMock: func(m *events.MockService) {
                m.CreateEventFunc = func(ctx context.Context, e *models.Event) error {
                    e.ID = 1
                    return nil
                }
            },
            wantStatus: http.StatusCreated,
            wantBody:   `"id":1`,
        },
        {
            name: "invalid JSON",
            body: `{invalid json}`,
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            wantStatus: http.StatusBadRequest,
            wantBody:   "invalid request body",
        },
        {
            name: "missing required field",
            body: `{
                "start_time": "2026-06-15T14:00:00-07:00",
                "timezone": "America/Los_Angeles"
            }`,
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            wantStatus: http.StatusBadRequest,
            wantBody:   "title",
        },
        {
            name: "unauthorized user",
            body: `{
                "title": "Event",
                "start_time": "2026-06-15T14:00:00-07:00",
                "timezone": "America/Los_Angeles"
            }`,
            user: &models.User{
                ID:   1,
                Role: models.RoleGuest,
            },
            setupMock: func(m *events.MockService) {
                m.CreateEventFunc = func(ctx context.Context, e *models.Event) error {
                    return &models.PermissionDeniedError{}
                }
            },
            wantStatus: http.StatusForbidden,
            wantBody:   "permission denied",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := &events.MockService{}
            if tt.setupMock != nil {
                tt.setupMock(mockService)
            }
            
            handlers := NewEventHandlers(mockService)
            
            req := httptest.NewRequest("POST", "/api/events", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            
            ctx := auth.WithUser(req.Context(), tt.user)
            req = req.WithContext(ctx)
            
            w := httptest.NewRecorder()
            handlers.CreateEvent(w, req)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
            }
            
            if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
                t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
            }
        })
    }
}

func TestEventHandlers_UpdateEvent(t *testing.T) {
    tests := []struct {
        name       string
        eventID    string
        body       string
        user       *models.User
        setupMock  func(*events.MockService)
        wantStatus int
        wantBody   string
    }{
        {
            name:    "valid update",
            eventID: "1",
            body: `{
                "title": "Updated Title",
                "version": 1
            }`,
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            setupMock: func(m *events.MockService) {
                m.UpdateEventFunc = func(ctx context.Context, e *models.Event) error {
                    return nil
                }
            },
            wantStatus: http.StatusOK,
        },
        {
            name:    "version conflict",
            eventID: "1",
            body: `{
                "title": "Updated Title",
                "version": 1
            }`,
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            setupMock: func(m *events.MockService) {
                m.UpdateEventFunc = func(ctx context.Context, e *models.Event) error {
                    return &models.VersionConflictError{
                        ResourceType: "event",
                        ResourceID:   1,
                        Expected:     1,
                    }
                }
            },
            wantStatus: http.StatusConflict,
            wantBody:   "version conflict",
        },
        {
            name:    "invalid event ID",
            eventID: "invalid",
            body:    `{"title": "Updated", "version": 1}`,
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            wantStatus: http.StatusBadRequest,
            wantBody:   "invalid event ID",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := &events.MockService{}
            if tt.setupMock != nil {
                tt.setupMock(mockService)
            }
            
            handlers := NewEventHandlers(mockService)
            
            req := httptest.NewRequest("PUT", "/api/events/"+tt.eventID, strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("id", tt.eventID)
            req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
            
            ctx := auth.WithUser(req.Context(), tt.user)
            req = req.WithContext(ctx)
            
            w := httptest.NewRecorder()
            handlers.UpdateEvent(w, req)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
            }
            
            if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
                t.Errorf("Body = %q, want to contain %q", w.Body.String(), tt.wantBody)
            }
        })
    }
}

func TestEventHandlers_PublishEvent(t *testing.T) {
    tests := []struct {
        name       string
        eventID    string
        user       *models.User
        setupMock  func(*events.MockService)
        wantStatus int
        wantBody   string
    }{
        {
            name:    "successful publish",
            eventID: "1",
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            setupMock: func(m *events.MockService) {
                m.PublishEventFunc = func(ctx context.Context, id int64) error {
                    return nil
                }
            },
            wantStatus: http.StatusOK,
        },
        {
            name:    "invalid state transition",
            eventID: "1",
            user: &models.User{
                ID:   1,
                Role: models.RoleEventManager,
            },
            setupMock: func(m *events.MockService) {
                m.PublishEventFunc = func(ctx context.Context, id int64) error {
                    return fmt.Errorf("invalid state transition")
                }
            },
            wantStatus: http.StatusBadRequest,
            wantBody:   "invalid state transition",
        },
        {
            name:    "unauthorized",
            eventID: "1",
            user: &models.User{
                ID:   2,
                Role: models.RoleEventManager,
            },
            setupMock: func(m *events.MockService) {
                m.PublishEventFunc = func(ctx context.Context, id int64) error {
                    return &models.PermissionDeniedError{}
                }
            },
            wantStatus: http.StatusForbidden,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := &events.MockService{}
            if tt.setupMock != nil {
                tt.setupMock(mockService)
            }
            
            handlers := NewEventHandlers(mockService)
            
            req := httptest.NewRequest("POST", "/api/events/"+tt.eventID+"/publish", nil)
            
            rctx := chi.NewRouteContext()
            rctx.URLParams.Add("id", tt.eventID)
            req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
            
            ctx := auth.WithUser(req.Context(), tt.user)
            req = req.WithContext(ctx)
            
            w := httptest.NewRecorder()
            handlers.PublishEvent(w, req)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
            }
        })
    }
}
```

---

## Dependencies

**Depends on:** 
- 02_STORY_03_event_service.md - Event service layer
- Epic 01 (Auth) - Authentication middleware

**Blocks:** 
- Frontend event management UI
- Event API integration

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout (`go test -timeout 30s ./internal/handlers/...`)
- [x] Test coverage >= 85% (achieved 90.4%)
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] API documentation complete
- [x] Error responses consistent
- [x] Changes committed to git

---

## Implementation Notes

### Error Response Format

All error responses follow consistent format:

```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Validation failed",
        "details": {
            "title": "Title must be between 3 and 200 characters"
        }
    }
}
```

### Request Validation

Use struct tags for validation:
- `validate:"required"` - Field is required
- `validate:"min=3,max=200"` - Length constraints
- `validate:"omitempty"` - Optional field

### URL Parameters

Extract ID from URL using chi router:

```go
idStr := chi.URLParam(r, "id")
id, err := strconv.ParseInt(idStr, 10, 64)
```

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **HLD:** Section 7 (API Design)
- **LLD:** [lld/08_API_LLD.md](../lld/08_API_LLD.md)
- **Epic:** [02_EPIC_events.md](02_EPIC_events.md)
