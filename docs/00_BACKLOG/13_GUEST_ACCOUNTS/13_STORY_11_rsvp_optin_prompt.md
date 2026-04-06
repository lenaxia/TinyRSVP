# User Story: RSVP Confirmation Page Opt-In Prompt

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** Medium  
**Status:** Not Started  
**Estimated Effort:** 4 hours  

---

## User Story

As a **guest who has just submitted an RSVP**, I want **to see an optional prompt to create a free account** so that **I can optionally save my identity to view all my invitations in one place, without being required to do so**.

---

## Acceptance Criteria

- [ ] After a successful RSVP submission, the confirmation page renders an opt-in prompt if no `tinyrsvp_guest` cookie is present
- [ ] If a `tinyrsvp_guest` cookie is present (guest already logged in), the prompt is not shown
- [ ] The prompt is visually low-key — clearly secondary to the confirmation message
- [ ] If the invite has a known email address, the identifier input is pre-filled with that email
- [ ] Submitting the prompt POSTs to `/guest/auth/request-otp` (the existing handler from Story 09)
- [ ] The RSVP submission flow itself is **completely unchanged** — no guest session is required to submit an RSVP
- [ ] The confirmation page handler does not depend on `RequireGuestAuth`
- [ ] Template tests verify prompt is shown/hidden correctly based on session presence
- [ ] All tests pass with timeout

---

## Technical Details

### Template Data Change

The confirmation page template data struct gains an optional `GuestOptIn` field:

```go
// In the confirmation page view model (wherever it's currently defined):
type RSVPConfirmationData struct {
    // ... existing fields ...
    GuestOptIn *GuestOptInPrompt
}

type GuestOptInPrompt struct {
    PrefilledEmail string // empty string if invite has no email
}
```

`GuestOptIn` is `nil` when the guest already has a `tinyrsvp_guest` session cookie. It is non-nil (with `PrefilledEmail` set if available) when no session cookie is present.

### Handler Change

In the existing RSVP confirmation handler (in `internal/handlers/rsvp.go` or equivalent):

```go
func (h *RSVPHandler) HandleConfirmation(w http.ResponseWriter, r *http.Request) {
    // ... existing RSVP lookup logic ...

    var guestOptIn *GuestOptInPrompt
    if _, err := r.Cookie("tinyrsvp_guest"); err != nil {
        // No guest session — show prompt
        guestOptIn = &GuestOptInPrompt{}
        if invite.Email != nil {
            guestOptIn.PrefilledEmail = *invite.Email
        }
    }

    data := RSVPConfirmationData{
        // ... existing fields ...
        GuestOptIn: guestOptIn,
    }
    // render template
}
```

### Template Change

In the RSVP confirmation HTML template, add at the bottom of the confirmation content:

```html
{{if .GuestOptIn}}
<section class="guest-optin">
    <p>Want to see all your invitations in one place?</p>
    <form method="POST" action="/guest/auth/request-otp">
        <input
            type="text"
            name="identifier"
            placeholder="Email or phone number"
            value="{{.GuestOptIn.PrefilledEmail}}"
            autocomplete="email"
        >
        <button type="submit">Send me a login code</button>
    </form>
</section>
{{end}}
```

No JavaScript required. The form submits directly to the existing `request-otp` handler. On success the user gets the OTP flow; on error the existing handler returns an error page. After verifying their OTP they are redirected to `/guest/account`.

---

## Tasks

### Phase 1: View Model and Handler (TDD)
- [ ] Write test: `TestRSVPConfirmation_NoGuestCookie_ShowsPrompt` — `GuestOptIn` is non-nil in template data
- [ ] Write test: `TestRSVPConfirmation_WithGuestCookie_HidesPrompt` — `GuestOptIn` is nil
- [ ] Write test: `TestRSVPConfirmation_PrefilledEmail_FromInvite` — pre-fill when invite has email
- [ ] Write test: `TestRSVPConfirmation_NoEmail_EmptyPrefill` — pre-fill is empty string when invite has no email
- [ ] Run tests (should fail)
- [ ] Update confirmation handler to populate `GuestOptIn`
- [ ] Run tests (should pass)

### Phase 2: Template
- [ ] Add `{{if .GuestOptIn}}` block to confirmation page template
- [ ] Style the opt-in section as visually secondary (smaller text, muted colors, below main content)
- [ ] Write template render test: prompt HTML present when `GuestOptIn` non-nil
- [ ] Write template render test: prompt HTML absent when `GuestOptIn` nil
- [ ] Run tests (should pass)

### Phase 3: Regression
- [ ] Run the full RSVP flow test suite — confirm no existing tests broken
- [ ] Manually verify: submit RSVP without guest cookie → prompt appears with pre-filled email
- [ ] Manually verify: submit RSVP with guest cookie → prompt absent

---

## Testing Requirements

```go
func TestRSVPConfirmation_NoGuestCookie_ShowsPrompt(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    invite := &models.Invite{ID: 1, EventID: 1, Email: testutil.StringPtr("alice@example.com")}

    mockRepo := mockrepos.NewMockInviteRepository(ctrl)
    mockRepo.EXPECT().GetByToken(gomock.Any(), "tok123").Return(invite, nil)

    handler := handlers.NewRSVPHandler(mockRepo /* ... */)

    r := httptest.NewRequest(http.MethodGet, "/rsvp/tok123/confirmation", nil)
    // No tinyrsvp_guest cookie
    w := httptest.NewRecorder()

    handler.HandleConfirmation(w, r)

    body := w.Body.String()
    if !strings.Contains(body, "guest-optin") {
        t.Error("expected guest opt-in prompt in response body")
    }
    if !strings.Contains(body, "alice@example.com") {
        t.Error("expected pre-filled email in opt-in form")
    }
}

func TestRSVPConfirmation_WithGuestCookie_HidesPrompt(t *testing.T) {
    // same setup, but add tinyrsvp_guest cookie to request
    // assert "guest-optin" is NOT in body
}
```

---

## Dependencies

**Depends on:** Story 09 (guest auth handlers — `/guest/auth/request-otp` route must exist)  
**Blocks:** Nothing

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass: `go test -timeout 30s -race ./internal/handlers/...`
- [ ] RSVP submission flow unchanged — no account required (verified by existing tests still passing)
- [ ] Prompt styled as secondary content, not blocking or prominent
- [ ] `go vet ./...` clean
