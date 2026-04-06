# STORY: Fix Unsubscribe Page — Add Template to rsvpPageTemplates

**Epic:** 14 - Bug Fixes & Code Gaps  
**Story ID:** 14_STORY_04  
**Priority:** High  
**Estimated Effort:** 1 hour  
**Severity:** High — the unsubscribe success page always renders "Failed to render page" (garbled body); the DB unsubscribe itself succeeds

---

## Problem

`cmd/server/main.go:410-418` parses `rsvpPageTemplates` from only three files:

```go
rsvpPageTemplates, err := template.New("rsvp_page.html").Funcs(funcMap).ParseFiles(
    "templates/web/partials/base.html",
    "templates/web/partials/navigation.html",
    "templates/web/rsvp_page.html",
)
```

`internal/handlers/rsvp.go:969` then tries:

```go
if err := h.templates.ExecuteTemplate(w, "unsubscribe.html", data); err != nil {
    http.Error(w, "Failed to render page", http.StatusInternalServerError)
}
```

`unsubscribe.html` is not in the template set, so `ExecuteTemplate` returns an error. Because `w.WriteHeader(status)` was already called before this point, `http.Error` cannot change the status code — it writes "Failed to render page\n" as the response body. The guest sees a blank or garbled page instead of the unsubscribe confirmation.

Proven by test:
```
CONFIRMED: unsubscribe.html not in template set: html/template: "unsubscribe.html" is undefined
```

The styled `templates/web/unsubscribe.html` exists and is correct — it simply is never loaded.

---

## Acceptance Criteria

- [ ] `rsvpPageTemplates` includes `templates/web/unsubscribe.html`
- [ ] A successful unsubscribe renders the styled `unsubscribe.html` page, not the inline fallback
- [ ] An unsubscribe error (invalid token, not found) also renders `unsubscribe.html` with the error message populated
- [ ] New test: `TestRSVPHandler_Unsubscribe_SuccessPage_RendersTemplate` verifies that after a successful unsubscribe the response body contains content from the template (not "Failed to render page")
- [ ] All 32 non-browser packages pass
- [ ] Update `docs/00_BACKLOG/05_EMAIL/README.md`: remove BUG-3
- [ ] Update `docs/00_BACKLOG/14_BUG_FIXES/README.md`: mark this story complete

---

## Technical Approach

One-line fix in `cmd/server/main.go`:

```go
rsvpPageTemplates, err := template.New("rsvp_page.html").Funcs(funcMap).ParseFiles(
    "templates/web/partials/base.html",
    "templates/web/partials/navigation.html",
    "templates/web/rsvp_page.html",
    "templates/web/unsubscribe.html",   // ADD THIS
)
```

Verify that `unsubscribe.html` uses the same template structure (extends `base.html`) and that `UnsubscribePageData` provides all fields the template references.

---

## Files to Change

- `cmd/server/main.go` — add `unsubscribe.html` to ParseFiles call
- `internal/handlers/rsvp_confirmation_test.go` or new test file — add unsubscribe page render test

---

## Testing

```bash
go test -run TestRSVPHandler_Unsubscribe -v ./internal/handlers/...
go test -timeout 30s ./...
```

Manual verification:
1. Start server with valid SMTP config
2. Send an invite to a test address
3. Click the unsubscribe link
4. Confirm the styled unsubscribe page renders (not "Failed to render page")

---

## Status

- **Status:** Not Started
