# Worklog 0155 — Epic 14 Story 4: Fix Unsubscribe Page Template

**Date:** 2026-04-06  
**Session type:** Bug fix  
**Packages changed:** `cmd/server`, `internal/handlers`

---

## Story 04 — Unsubscribe page renders styled HTML instead of garbled body

**Root cause:** `cmd/server/main.go` built `rsvpPageTemplates` from only three files (`base.html`, `navigation.html`, `rsvp_page.html`). `renderUnsubscribePage` called `h.templates.ExecuteTemplate(w, "unsubscribe.html", data)` — template not in set, error returned. Because `w.WriteHeader(status)` was already called, `http.Error` wrote "Failed to render page\n" as the body without changing the status code.

**Fix:** Added `"templates/web/unsubscribe.html"` to the `ParseFiles` call in `main.go:410-415`. One line.

**Compatibility verified:**
- `unsubscribe.html` is a standalone HTML document with no `{{define}}` blocks — `ParseFiles` names it `"unsubscribe.html"` by basename, matching the `ExecuteTemplate` call exactly.
- Uses only `.Success`, `.ErrorMessage`, `.Event.Title`, `.Event.Description` — no custom funcmap functions required.
- No conflict with `base.html`'s `{{define "base"}}` named template in the same set.

**Why the existing test didn't catch this:**
`TestUnsubscribeHandler_Integration_Success` loaded `unsubscribe.html` in isolation via `template.ParseFiles("../../templates/web/unsubscribe.html")` and set it directly on the handler. This bypassed the production template construction entirely, masking the bug.

**New test:** `TestUnsubscribeHandler_ProductionTemplateSet_RendersUnsubscribePage` in `rsvp_unsubscribe_integration_test.go` — mirrors the exact `ParseFiles` call from `main.go` (including base.html, navigation.html, rsvp_page.html, unsubscribe.html), runs a real unsubscribe request, and asserts:
1. Status 200
2. Body does not contain "Failed to render page"
3. Body contains "Unsubscribed Successfully" (from the styled template)
4. Body contains the event title

**Epic 05 status:** All three email bugs now fixed. Epic 05 marked ✅ Complete.

---

## Test results

32/32 non-browser packages pass (`go test -count=1 ./...` excluding `tests/ux`).
