# User Story: RSVP Routes (Guest-Facing)

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 1.5 days

---

## User Story

As a **guest**, I want **to submit my RSVP via a unique token link** so that **I can respond to an event invitation without creating an account**.

---

## Acceptance Criteria

- [ ] GET /rsvp/{token} - RSVP page (no auth required)
- [ ] POST /rsvp/{token} - Submit RSVP (no auth required)
- [ ] GET /rsvp/{token}/confirm - Confirmation page
- [ ] GET /unsubscribe/{token} - Unsubscribe from reminders
- [ ] Token validation
- [ ] Event details displayed
- [ ] Response options (attending/not attending/maybe)
- [ ] Plus ones input
- [ ] Preference questions
- [ ] Deadline enforcement
- [ ] Update existing RSVP
- [ ] Rate limiting

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

- [ ] Implement RSVP page handler
- [ ] Implement RSVP submission handler
- [ ] Implement confirmation page handler
- [ ] Implement unsubscribe handler
- [ ] Add token validation
- [ ] Add deadline checking
- [ ] Add plus ones validation
- [ ] Test RSVP flow
- [ ] Test token expiration
- [ ] Test deadline enforcement

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

- [ ] All acceptance criteria met
- [ ] All routes implemented
- [ ] Token validation working
- [ ] Deadline enforcement working
- [ ] Tests passing
- [ ] Documentation complete
