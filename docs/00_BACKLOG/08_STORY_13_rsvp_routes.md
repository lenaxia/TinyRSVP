# User Story: RSVP Routes (Guest-Facing)

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)
**Priority:** High
**Status:** Complete (100%)
**Estimated Effort:** 1.5 days
**Validation Date:** 2026-01-10
**Validation Document:** [2026-01-10_44_story_13_validation.md](../01_WORKLOG/2026-01-10_44_story_13_validation.md)
**Completion Date:** 2026-01-10

---

## User Story

As a **guest**, I want **to submit my RSVP via a unique token link** so that **I can respond to an event invitation without creating an account**.

---

## Acceptance Criteria

- [x] GET /rsvp/{token} - RSVP page (no auth required)
- [x] POST /rsvp/{token} - Submit RSVP (no auth required)
- [x] GET /rsvp/{token}/confirm - Confirmation page
- [x] GET /unsubscribe/{token} - Unsubscribe from reminders
- [x] Token validation
- [x] Event details displayed
- [x] Response options (attending/not attending/maybe)
- [x] Plus ones input
- [x] Preference questions
- [x] Deadline enforcement
- [x] Update existing RSVP
- [x] Rate limiting

**Status: 12/12 criteria met (100%)**

---

## Technical Details

### Routes
```go
r.Route("/rsvp/{token}", func(r chi.Router) {
    r.Use(RateLimit(100)) // No auth, so rate limit aggressively
    
    r.Get("/", handlers.RSVPPage)
    r.Post("/", handlers.SubmitRSVP)
    r.Get("/confirm", handlers.RSVPConfirmation)
})

r.Get("/unsubscribe/{token}", handlers.Unsubscribe)
```

---

## Tasks

- [x] Implement RSVP page handler (internal/handlers/rsvp.go:83-199)
- [x] Implement RSVP submission handler (internal/handlers/rsvp.go:258-285)
- [x] Implement confirmation page handler (internal/handlers/rsvp.go:445-596)
- [x] Implement unsubscribe handler (internal/handlers/rsvp.go:599-731)
- [x] Add token validation
- [x] Add deadline checking
- [x] Add plus ones validation
- [x] Test RSVP flow
- [x] Test token expiration
- [x] Test deadline enforcement

**Status: 10/10 tasks complete (100%)**

---

## Dependencies

**Depends on:** 
- 08_STORY_00_router_setup.md
- 08_STORY_05_rate_limiting.md
- Epic 04 (RSVP service)

**Blocks:** 08_STORY_14_rsvp_ui.md

---

## Handler Examples

### RSVP Page
```go
func (h *Handlers) RSVPPage(w http.ResponseWriter, r *http.Request) {
    token := chi.URLParam(r, "token")
    
    invite, event, err := h.rsvp.GetInviteByToken(r.Context(), token)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    // Check deadline
    if time.Now().After(event.RSVPDeadline) {
        h.templates.Render(w, "rsvp/deadline_passed.html", nil)
        return
    }
    
    // Get existing RSVP if any
    existingRSVP, _ := h.rsvp.GetByInvite(r.Context(), invite.ID)
    
    data := struct {
        Event        *models.Event
        Invite       *models.Invite
        ExistingRSVP *models.RSVP
        Questions    []*models.PreferenceQuestion
    }{
        Event:        event,
        Invite:       invite,
        ExistingRSVP: existingRSVP,
        Questions:    event.Questions,
    }
    
    h.templates.Render(w, "rsvp/form.html", data)
}
```

### Submit RSVP
```go
func (h *Handlers) SubmitRSVP(w http.ResponseWriter, r *http.Request) {
    token := chi.URLParam(r, "token")
    
    var req SubmitRSVPRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        HandleError(w, r, NewValidationError("Invalid request"))
        return
    }
    
    rsvp, err := h.rsvp.Submit(r.Context(), token, &req)
    if err != nil {
        HandleError(w, r, err)
        return
    }
    
    // Redirect to confirmation
    http.Redirect(w, r, fmt.Sprintf("/rsvp/%s/confirm", token), http.StatusSeeOther)
}
```

---

## Testing Strategy

```go
func TestRSVPPage_ValidToken(t *testing.T)
func TestRSVPPage_InvalidToken(t *testing.T)
func TestRSVPPage_ExpiredToken(t *testing.T)
func TestRSVPPage_DeadlinePassed(t *testing.T)
func TestSubmitRSVP_Success(t *testing.T)
func TestSubmitRSVP_Validation(t *testing.T)
func TestSubmitRSVP_PlusOnesExceeded(t *testing.T)
func TestRSVPConfirmation(t *testing.T)
func TestUnsubscribe(t *testing.T)
```

---

## Security Considerations

- No authentication required (token-based access)
- Aggressive rate limiting
- Token validation on every request
- Prevent token enumeration
- Log suspicious activity

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)
- **RSVP Epic:** [04_EPIC_rsvp.md](04_EPIC_rsvp.md)

---

## Definition of Done

- [x] All acceptance criteria met (12/12)
- [x] All routes implemented (5/5)
- [x] Token validation working
- [x] Deadline enforcement working
- [x] Tests passing (all RSVP and unsubscribe tests pass)
- [x] Documentation complete (validation doc created, story updated)

---

## Implementation Status

### Completed ✅
- RSVP page with full feature set (token validation, event details, questions, deadline checks)
- RSVP submission with comprehensive error handling
- RSVP updates for existing responses
- Confirmation page with answer display
- Unsubscribe route and handler with full error handling
- CSRF protection integration
- Rate limiting via middleware
- Extensive test coverage (unit, integration, and end-to-end tests)
- Template rendering support for all pages

### Files
- **Handler:** internal/handlers/rsvp.go
- **Tests:** internal/handlers/rsvp_*_test.go
- **Service:** internal/invites/service.go (UnsubscribeFromReminders method)
- **Service Tests:** internal/invites/service_unsubscribe_test.go
- **Router:** internal/handlers/router.go (lines 490-515)
- **Template:** templates/web/unsubscribe.html
- **Validation:** docs/01_WORKLOG/2026-01-10_44_story_13_validation.md

---

## Completion Summary

**Completed:** 2026-01-10

### Implementation Details

1. **Service Layer** (internal/invites/service.go)
   - Added `UnsubscribeFromReminders(ctx context.Context, token string) error` to InviteService interface
   - Implemented method with token validation, expiry checking, revocation checking
   - Idempotent: returns success if already unsubscribed
   - Comprehensive unit tests (7 test cases)

2. **Handler Layer** (internal/handlers/rsvp.go)
   - Added `Unsubscribe(w http.ResponseWriter, r *http.Request)` to RSVPHandlerInterface
   - Implemented in RSVPHandler with full error handling
   - Handles: invalid token, expired token, revoked invite, event not found
   - Unit tests (9 test cases) and integration tests (6 test cases)

3. **Router** (internal/handlers/router.go)
   - Added route: `GET /unsubscribe/{token}`
   - No authentication required (token-based access)
   - Subject to rate limiting

4. **Template** (templates/web/unsubscribe.html)
   - Mobile-first responsive design
   - Success and error states
   - Accessible with ARIA labels
   - Consistent with other RSVP templates

5. **Testing**
   - Service layer: 7 unit tests
   - Handler layer: 9 unit tests
   - Integration: 6 end-to-end tests
   - All tests passing
   - Coverage: success paths, error paths, edge cases

### Test Coverage
- ✅ Valid token unsubscribe
- ✅ Invalid token handling
- ✅ Expired token handling
- ✅ Revoked invite handling
- ✅ Already unsubscribed (idempotent)
- ✅ Database errors
- ✅ Event not found
- ✅ Hash errors
- ✅ Update errors
