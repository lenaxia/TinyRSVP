# STORY: Render Invite Email Template Instead of Hardcoded Plaintext

**Epic:** 14 - Bug Fixes & Code Gaps  
**Story ID:** 14_STORY_02  
**Priority:** High  
**Estimated Effort:** 3 hours  
**Severity:** High — invite emails sent to guests are always plain-text with no event details, ignoring the HTML template system

---

## Problem

`internal/invites/service.go:275`:

```go
bodyText := fmt.Sprintf("Hello %s,\n\nYou've been invited to an event.\n\nRSVP here: %s\n\nThis link expires on %s.",
    guestName, rsvpURL, expiresAt.Format("January 2, 2006"))
```

`SendInvite` hard-codes a minimal plaintext body. The template system seeds a `TemplateTypeInviteEmail` template into the database (`internal/templates/seeder.go`) with a proper HTML invite template, but `SendInvite` never calls it. Every invite email sent through the system is plain-text, regardless of any event template customization.

The test `TestSendInvite_Success` only checks that the RSVP URL appears in `BodyText` — it does not verify template rendering — so it passes and masks this bug.

---

## Acceptance Criteria

- [ ] `SendInvite` renders the event's `TemplateTypeInviteEmail` template (HTML + text parts) using the template service
- [ ] If no custom template exists for the event, the default seeded invite email template is used as fallback
- [ ] The queued email has both `BodyHTML` and `BodyText` populated
- [ ] `TestSendInvite_Success` is updated to assert that template rendering was invoked (not just that the URL appears)
- [ ] All existing invite send tests pass
- [ ] All 32 non-browser packages pass
- [ ] Update `docs/00_BACKLOG/05_EMAIL/README.md`: remove BUG-1, update success criteria
- [ ] Update `docs/00_BACKLOG/14_BUG_FIXES/README.md`: mark this story complete

---

## Technical Approach

### 1. Add template service dependency to `inviteService`

`inviteService` currently takes `generator` and `repo`. Add a `templateService templates.Service` field (or a narrower interface with just `RenderInviteEmail`).

### 2. Render in `SendInvite`

```go
// Look up the event's invite email template (or default)
tmpl, err := s.templateService.GetTemplateForEvent(ctx, invite.EventID, models.TemplateTypeInviteEmail)
if err != nil {
    // fall back to plaintext if template unavailable
}

data := &templates.InviteEmailData{
    GuestName:  guestName,
    EventTitle: event.Title,
    RSVPURL:    rsvpURL,
    ExpiresAt:  invite.ExpiresAt,
    // ...
}

bodyHTML, _ := s.templateService.RenderHTML(ctx, tmpl.ID, data)
bodyText, _ := s.templateService.RenderText(ctx, tmpl.ID, data)
```

### 3. Update `SendInviteRequest` or `inviteService` constructor

`NewInviteService(...)` should accept a `templates.Service` or a dedicated `InviteEmailRenderer` interface (preferred for testability).

---

## Files to Change

- `internal/invites/service.go` — update `SendInvite`, update constructor
- `internal/invites/service_send_test.go` — update `TestSendInvite_Success`, add template rendering assertions
- `cmd/server/main.go` — pass template service into `NewInviteService`

---

## Testing

```bash
go test -timeout 30s ./internal/invites/...
go test -timeout 30s ./...
```

---

## Status

- **Status:** ✅ Complete — 2026-04-06

## Implementation Notes

- Added `RenderEmailTemplate(ctx, eventID, templateType, data)` to `templates.Service` interface and `*service` implementation (`internal/templates/service.go`). Uses `GetTemplateForEvent` (per-event with system default fallback) + `Engine.Parse` + `Engine.ExecuteToString` for both HTML and text bodies.
- `inviteService` gains a `templateService templates.Service` field. Constructor `NewInviteServiceWithTemplates` wires it; `NewInviteService` leaves it nil (plaintext fallback, backward compatible).
- `SendInvite` builds `map[string]interface{}{"Event", "Invite", "RSVPURL", "MaxPlusOnes"}` matching the variables used in `internal/templates/defaults/invite_email.html/.txt`, calls `RenderEmailTemplate`, and falls back to plaintext if nil or on any render error.
- Subject line updated to `"You're Invited: {EventTitle}"` when event is available.
- `cmd/server/main.go` uses `NewInviteServiceWithTemplates(tokenGenerator, inviteRepo, templateService)`.
- Generated mock (`internal/testutil/mocks/services/mock_template_service.go`) regenerated to include `RenderEmailTemplate`.
- `internal/handlers/templates_test.go` local `mockTemplateService` stub updated with no-op `RenderEmailTemplate`.
- 3 new tests in `internal/invites/service_send_test.go`: template path, fallback on render error, no-service path.
- 32/32 packages pass.
