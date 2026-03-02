# User Story: Guest-Facing RSVP Page

**Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 1 day
**Actual Effort:** 2 hours
**Completed:** 2026-01-08

---

## User Story

As a **guest**, I want **to view an RSVP page with event details** so that **I can understand the event and prepare to respond**.

---

## Acceptance Criteria

- [x] Guest can access RSVP page via token link
- [x] Token validation occurs before page load
- [x] Event details displayed (title, date, time, location)
- [x] Timezone converted to event's local time
- [x] RSVP deadline displayed prominently
- [x] Page shows if guest has already responded
- [x] Mobile-responsive design (320px-767px)
- [x] Works without JavaScript (progressive enhancement)
- [x] Clear error messages for invalid/expired tokens
- [x] Page loads quickly (inline CSS, no external resources)
- [x] Accessible (WCAG 2.1 AA compliant)

---

## Technical Details

### Route
```
GET /rsvp/:token
```

### Handler Location
- [`internal/handlers/rsvp.go`](../../internal/handlers/rsvp.go)
- [`internal/handlers/rsvp_test.go`](../../internal/handlers/rsvp_test.go)

### Template Location
- [`templates/web/rsvp_page.html`](../../templates/web/rsvp_page.html)

### Page Flow

```
Guest clicks invite link
         ↓
GET /rsvp/:token
         ↓
Validate token
         ↓
Load invite + event
         ↓
Check existing RSVP
         ↓
Render RSVP page
```

### Template Data Structure

```go
type RSVPPageData struct {
    Event           *models.Event
    Invite          *models.Invite
    ExistingRSVP    *models.RSVP
    Questions       []*models.PreferenceQuestion
    Token           string
    DeadlinePassed  bool
    EventPassed     bool
    LocalStartTime  string
    LocalEndTime    string
    TimeUntilEvent  string
    CanUpdate       bool
}
```

---

## Tasks

### Phase 1: Handler Implementation (TDD)
- [x] Create RSVP handler struct
- [x] Write test for valid token
- [x] Write test for invalid token
- [x] Write test for expired token
- [x] Write test for revoked invite
- [x] Write test for cancelled event
- [x] Write test for archived event
- [x] Write test for existing RSVP
- [x] Implement GetRSVPPage handler
- [x] Run tests (should pass)

### Phase 2: Template Creation
- [x] Create base RSVP page template
- [x] Add event details section
- [x] Add date/time display with timezone
- [x] Add location information
- [x] Add RSVP deadline display
- [x] Add "already responded" message
- [x] Add error message display
- [x] Test template rendering

### Phase 3: Mobile-First CSS
- [x] Create mobile base styles (320px-767px)
- [x] Style event details card
- [x] Style date/time display
- [x] Style deadline warning
- [x] Add tablet styles (768px-1023px)
- [x] Add desktop styles (1024px+)
- [x] Test on multiple screen sizes

### Phase 4: Accessibility
- [x] Add semantic HTML structure
- [x] Add ARIA labels where needed
- [x] Keyboard navigation supported
- [x] Screen reader compatible
- [x] Color contrast ratios verified
- [x] Focus indicators added

### Phase 5: Integration Testing
- [x] Test full page load flow
- [x] Test with various event states
- [x] Test with various invite states
- [x] Test timezone conversion
- [x] Test mobile responsiveness
- [x] Performance verified

---

## Page Layout

### Mobile (320px-767px)

```
┌─────────────────────────┐
│  Event Title            │
├─────────────────────────┤
│  📅 Date & Time         │
│  🕐 Local: ...          │
│  🌍 Timezone: ...       │
├─────────────────────────┤
│  📍 Location            │
│  Address details        │
├─────────────────────────┤
│  ⏰ RSVP by: ...        │
│  (X days remaining)     │
├─────────────────────────┤
│  Description            │
│  Event details...       │
├─────────────────────────┤
│  [RSVP Form Below]      │
└─────────────────────────┘
```

### Desktop (1024px+)

```
┌───────────────────────────────────────┐
│           Event Title                  │
├──────────────────┬────────────────────┤
│  Event Details   │  RSVP Form         │
│  📅 Date         │                    │
│  🕐 Time         │  [Form fields]     │
│  📍 Location     │                    │
│  ⏰ Deadline     │                    │
│                  │                    │
│  Description     │                    │
└──────────────────┴────────────────────┘
```

---

## Token Validation

### Validation Checks
1. Token format valid (40 characters, alphanumeric)
2. Token hash exists in database
3. Invite not revoked
4. Invite not expired
5. Event not cancelled
6. Event not archived

### Error Messages

| Condition | Message |
|-----------|---------|
| Invalid token format | "Invalid invite link" |
| Token not found | "Invite not found or has been revoked" |
| Invite revoked | "This invite has been revoked" |
| Invite expired | "This invite has expired" |
| Event cancelled | "This event has been cancelled" |
| Event archived | "This event is no longer active" |

---

## Timezone Handling

### Server-Side
```go
func formatEventTime(event *models.Event, timezone string) (string, error) {
    loc, err := time.LoadLocation(event.Timezone)
    if err != nil {
        return "", err
    }
    
    eventTime := event.StartTime.In(loc)
    
    // Format: "Monday, January 15, 2026 at 6:00 PM PST"
    return eventTime.Format("Monday, January 2, 2006 at 3:04 PM MST"), nil
}
```

### Client-Side Enhancement (Optional)
```javascript
// Progressive enhancement: convert to user's local timezone
document.querySelectorAll('[data-utc-time]').forEach(el => {
    const utcTime = el.dataset.utcTime;
    const localTime = new Date(utcTime).toLocaleString();
    el.textContent = localTime;
});
```

---

## Testing Strategy

### Unit Tests

```go
func TestRSVPHandler_GetRSVPPage(t *testing.T) {
    tests := []struct {
        name       string
        token      string
        setupMock  func(*mocks.MockInviteService)
        wantStatus int
        wantBody   string
    }{
        {
            name:  "valid token shows event",
            token: "validtoken123",
            setupMock: func(m *mocks.MockInviteService) {
                m.ValidateTokenFunc = func(ctx context.Context, token string) (*models.Invite, error) {
                    return &models.Invite{
                        ID:      1,
                        EventID: 1,
                        Status:  models.InviteStatusSent,
                    }, nil
                }
            },
            wantStatus: 200,
            wantBody:   "Birthday Party",
        },
        {
            name:  "invalid token shows error",
            token: "invalidtoken",
            setupMock: func(m *mocks.MockInviteService) {
                m.ValidateTokenFunc = func(ctx context.Context, token string) (*models.Invite, error) {
                    return nil, models.ErrInvalidToken
                }
            },
            wantStatus: 404,
            wantBody:   "Invite not found",
        },
        {
            name:  "revoked invite shows error",
            token: "revokedtoken",
            setupMock: func(m *mocks.MockInviteService) {
                m.ValidateTokenFunc = func(ctx context.Context, token string) (*models.Invite, error) {
                    return &models.Invite{
                        Status: models.InviteStatusRevoked,
                    }, nil
                }
            },
            wantStatus: 403,
            wantBody:   "invite has been revoked",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := &mocks.MockInviteService{}
            tt.setupMock(mockService)
            
            handler := NewRSVPHandler(mockService, nil, nil)
            
            req := httptest.NewRequest("GET", "/rsvp/"+tt.token, nil)
            w := httptest.NewRecorder()
            
            handler.GetRSVPPage(w, req)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
            }
            
            if !strings.Contains(w.Body.String(), tt.wantBody) {
                t.Errorf("Body doesn't contain %q", tt.wantBody)
            }
        })
    }
}
```

### Integration Tests

```go
func TestRSVPPage_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    // Create test event
    event := createTestEvent(t, db)
    
    // Create test invite
    invite, token := createTestInvite(t, db, event.ID)
    
    // Make request
    req := httptest.NewRequest("GET", "/rsvp/"+token, nil)
    w := httptest.NewRecorder()
    
    handler.ServeHTTP(w, req)
    
    // Verify response
    if w.Code != 200 {
        t.Fatalf("Status = %d, want 200", w.Code)
    }
    
    body := w.Body.String()
    
    // Check event details present
    if !strings.Contains(body, event.Title) {
        t.Error("Event title not found in response")
    }
    
    if !strings.Contains(body, event.Location) {
        t.Error("Event location not found in response")
    }
}
```

---

## Performance Requirements

- Page load: <2 seconds
- Time to Interactive: <3 seconds
- First Contentful Paint: <1 second
- Total page weight: <100KB (excluding images)

### Optimization Strategies
- Inline critical CSS
- Defer non-critical JavaScript
- Optimize images (WebP format, lazy loading)
- Minimize HTTP requests
- Enable gzip compression

---

## Accessibility Requirements

### WCAG 2.1 AA Compliance
- Semantic HTML5 elements
- Proper heading hierarchy (h1, h2, h3)
- ARIA labels for icons
- Keyboard navigation support
- Screen reader tested
- Color contrast ratio >= 4.5:1
- Focus indicators visible
- Touch targets >= 44px

### Example Semantic HTML
```html
<article class="event-details" role="main">
    <header>
        <h1>{{.Event.Title}}</h1>
    </header>
    
    <section class="event-info">
        <h2 class="sr-only">Event Information</h2>
        
        <div class="event-datetime">
            <svg aria-hidden="true">...</svg>
            <time datetime="{{.Event.StartTime}}">
                {{.LocalStartTime}}
            </time>
        </div>
        
        <div class="event-location">
            <svg aria-hidden="true">...</svg>
            <address>{{.Event.Location}}</address>
        </div>
    </section>
</article>
```

---

## Dependencies

**Depends on:**
- Story 00: RSVP Model (for data structures)
- Epic 03: Invites (for token validation)
- Epic 02: Events (for event data)

**Blocks:**
- Story 02: RSVP Submission (needs page to submit from)
- Story 04: Plus Ones UI (needs page to add UI to)
- Story 05: Question Display (needs page to display on)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Handler implemented and tested
- [x] Template created and rendering correctly
- [x] Mobile-responsive design working
- [x] Accessibility requirements met
- [x] Unit tests passing (17 tests total)
- [x] Integration tests passing
- [x] Performance targets met
- [x] Error handling complete
- [x] Documentation updated
- [x] Code committed
- [x] No linter warnings

---

## References

- **HLD:** Section 7.2 (RSVP Flow)
- **LLD:** [lld/04_RSVP_LLD.md](../lld/04_RSVP_LLD.md)
- **Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)
- **Frontend:** README-LLM.md Section "Frontend: Plain CSS + Vanilla JavaScript"
