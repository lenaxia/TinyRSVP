# Worklog 0157 — Epic 14 Stories 5 & 6 + Revalidation Fixes

**Date:** 2026-04-06  
**Session type:** Bug fix / feature wiring  
**Packages changed:** `internal/handlers`, `internal/email`, `internal/rsvp`, `cmd/server`

---

## Story 05 — Wire template editor into router

**Root cause:** `TemplateEditorHandlers`, `EditorService`, `template_editor.html`, and all tests existed and compiled but were never instantiated or registered. `NewTemplateEditorHandlers` was never called in `main.go` and no routes were registered.

**Revalidation caught two additional issues:**

1. **Pre-existing bug in `GetEditorPage`**: the data struct passed to `renderPage` was missing `ActivePage string`. `navigation.html:20` evaluates `.ActivePage` on every page — without it, `ExecuteTemplate` returns an error and the handler writes "Failed to render page". All existing tests loaded the template set without `base.html`/`navigation.html` or bypassed `SetTemplates`, masking this bug. Fixed by adding `ActivePage: "templates"` to the data struct.

2. **No production template set test**: added `TestGetEditorPage_ProductionTemplateSet_RendersEditorPage` which mirrors the exact `ParseFiles` call from `main.go` and proves the real template renders without the fallback error.

**Wiring details:**
- `router.go`: `TemplateEditorHandlers RouteRegistrar` field added to `RouterHandlers`. `RegisterRoutes(r)` called on root router (not `apiRouter`) — `RegisterRoutes` hardcodes `/api/` prefix in paths; mounting on `apiRouter` would produce `/api/api/...`.
- `main.go`: `templates.NewEditorService(templateRepo)` → `handlers.NewTemplateEditorHandlers(editorService)` → parse `template_editor.html` with base+navigation → `SetTemplates` → `RouterHandlers.TemplateEditorHandlers`.
- Auth note: editor API routes bypass `apiRouter`'s `RequireAuth` middleware. Handler checks `auth.UserFromContext` directly and returns 401. Acceptable for now; tracked in Epic 09.

**Routes live at:**
- `GET /templates/{id}/edit`
- `GET|PUT /api/templates/{id}/components`
- `POST /api/templates/{id}/components/preview`
- `GET /api/templates/{id}/components/validate`

---

## Story 06 — Remove MockService from production email package

**Root cause:** `internal/email/service.go` contained `MockService` struct — test infrastructure compiled into the production binary.

**Approach:** 
- Added local `mockEmailService` (unexported) directly in `internal/rsvp/service_email_test.go` — same package, test file only.
- `email` package import removed from the test file.
- 4 usages migrated; exported field names → unexported.
- `MockService` and its method deleted from `service.go`.
- The generated `testutil/mocks/services/MockEmailService` (gomock) was not used — the existing tests use a function-field pattern; switching to gomock's `EXPECT()` would require a broader refactor beyond this story's scope.

---

## Test results

32/32 non-browser packages pass (`go test -count=1 ./...` excluding `tests/ux`).

## Epic 14 Status

Stories 02–06 complete. Story 01 (`X-Test-User-ID` bypass removal) deferred — appropriate in test/development mode. Will be addressed in Epic 09 (Security Audit) before public deployment.
