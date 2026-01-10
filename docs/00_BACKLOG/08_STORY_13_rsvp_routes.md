# User Story: RSVP Routes (Guest-Facing)

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)
**Priority:** High
**Status:** Mostly Complete (92%) - Missing Unsubscribe Route
**Estimated Effort:** 1.5 days
**Validation Date:** 2026-01-10
**Validation Document:** [2026-01-10_44_story_13_validation.md](../01_WORKLOG/2026-01-10_44_story_13_validation.md)

---

## User Story

As a **guest**, I want **to submit my RSVP via a unique token link** so that **I can respond to an event invitation without creating an account**.

---

## Acceptance Criteria

- [x] GET /rsvp/{token} - RSVP page (no auth required)
- [x] POST /rsvp/{token} - Submit RSVP (no auth required)
- [x] GET /rsvp/{token}/confirm - Confirmation page
- [ ] GET /unsubscribe/{token} - Unsubscribe from reminders **MISSING**
- [x] Token validation
- [x] Event details displayed
- [x] Response options (attending/not attending/maybe)
- [x] Plus ones input
- [x] Preference questions
- [x] Deadline enforcement
- [x] Update existing RSVP
- [x] Rate limiting

**Status: 11/12 criteria met (92%)**

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
- [ ] Implement unsubscribe handler **MISSING**
- [x] Add token validation
- [x] Add deadline checking
- [x] Add plus ones validation
- [x] Test RSVP flow
- [x] Test token expiration
- [x] Test deadline enforcement

**Status: 9/10 tasks complete (90%)**

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

- [ ] All acceptance criteria met (11/12 - missing unsubscribe)
- [ ] All routes implemented (4/5 - missing unsubscribe)
- [x] Token validation working
- [x] Deadline enforcement working
- [x] Tests passing (all RSVP tests pass)
- [x] Documentation complete (validation doc created)

---

## Implementation Status

### Completed ✅
- RSVP page with full feature set (token validation, event details, questions, deadline checks)
- RSVP submission with comprehensive error handling
- RSVP updates for existing responses
- Confirmation page with answer display
- CSRF protection integration
- Rate limiting via middleware
- Extensive test coverage (integration and unit tests)
- Template rendering support

### Missing ❌
- Unsubscribe route and handler
- Unsubscribe tests
- Unsubscribe template

### Files
- **Handler:** internal/handlers/rsvp.go
- **Tests:** internal/handlers/rsvp_*_test.go
- **Router:** internal/handlers/router.go (lines 490-512)
- **Validation:** docs/01_WORKLOG/2026-01-10_44_story_13_validation.md

---

## Next Steps

To complete this story:
1. Add `Unsubscribe(w http.ResponseWriter, r *http.Request)` to RSVPHandlerInterface
2. Implement unsubscribe handler in rsvp.go
3. Add route `r.Get("/unsubscribe/{token}", handlers.RSVPHandler.Unsubscribe)` to router.go
4. Create unsubscribe template
5. Write tests for unsubscribe functionality
6. Update story status to Complete
