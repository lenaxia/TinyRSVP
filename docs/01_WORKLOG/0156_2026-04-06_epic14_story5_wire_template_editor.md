# Worklog 0156 — Epic 14 Story 5: Wire Template Editor

**Date:** 2026-04-06  
**Session type:** Bug fix / feature wiring  
**Packages changed:** `internal/handlers/router.go`, `cmd/server/main.go`

---

## Story 05 — Template editor routes registered and reachable

**Root cause:** `TemplateEditorHandlers`, `EditorService`, `template_editor.html`, and all associated tests existed and compiled but were never instantiated or registered. `NewTemplateEditorHandlers` was never called in `main.go` and the routes were never added to the router.

**Investigation before touching anything:**
- `templates.EditorService` fully implemented in `internal/templates/editor_service.go` — `GetEditableTemplate`, `UpdateComponents`, `AddComponent`, `RemoveComponent`, `UpdateComponentProperty`, `ReorderComponents`, `PreviewChanges`. All backed by `TemplateRepository` methods that exist (`GetComponentConfig`, `UpdateComponentConfig`, `ValidateComponentConfig`).
- `template_editor.html` uses `{{template "base" .}}` — needs `base.html` and `navigation.html` in the parse set.
- `RegisterRoutes` hardcodes `/api/templates/{id}/components` in the path. Must be called on the root router `r`, not `apiRouter` (which is mounted at `/api` — would produce `/api/api/...`).
- Auth: handler methods check `auth.UserFromContext` directly. The `/api/...` routes will bypass `apiRouter`'s `RequireAuth` middleware but still return 401 on unauthenticated requests.
- `parseTemplateID` used internally is defined in `templates.go` — same package, no issue.

**Changes:**
- `router.go`: `TemplateEditorHandlers RouteRegistrar` added to `RouterHandlers`. `handlers.TemplateEditorHandlers.RegisterRoutes(r)` called on root router after `r.Mount("/api", apiRouter)`.
- `main.go`: `editorService := templates.NewEditorService(templateRepo)`; `templateEditorHandlers := handlers.NewTemplateEditorHandlers(editorService)`; `template_editor.html` parsed with base + navigation; `SetTemplates` called; `TemplateEditorHandlers: templateEditorHandlers` in `RouterHandlers` literal.

**Routes now live at:**
- `GET /templates/{id}/edit` — editor UI page
- `GET /api/templates/{id}/components` — get component config
- `PUT /api/templates/{id}/components` — update components
- `POST /api/templates/{id}/components/preview` — preview changes
- `GET /api/templates/{id}/components/validate` — validate config

**Epic 08 status:** ISSUE-2 resolved. All known issues fixed.

---

## Test results

32/32 non-browser packages pass (`go test -count=1 ./...` excluding `tests/ux`).
