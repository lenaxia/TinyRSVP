# STORY: Wire Template Editor Routes into Router (or Delete)

**Epic:** 14 - Bug Fixes & Code Gaps  
**Story ID:** 14_STORY_05  
**Priority:** Medium  
**Estimated Effort:** 2 hours  
**Severity:** Medium — full feature implementation exists but is completely unreachable at runtime

---

## Problem

`internal/handlers/template_editor.go` implements a complete component editor UI:

- `GetEditorPage` — renders `/templates/{id}/edit`
- `GetComponents` — `GET /api/templates/{id}/components`
- `UpdateComponents` — `PUT /api/templates/{id}/components`
- `PreviewComponents` — `POST /api/templates/{id}/components/preview`
- `ValidateComponents` — `POST /api/templates/{id}/components/validate`

`NewTemplateEditorHandlers` is **never called** in `cmd/server/main.go`. `TemplateEditorHandlers` is **not in** the `RouterHandlers` struct in `internal/handlers/router.go`. None of the above routes are registered. The feature is silently dead.

The API spec in `docs/00_BACKLOG/08_API/README.md` does not list these routes (they were not in the original spec), but the handler was written and has its own tests.

---

## Acceptance Criteria

**Option A: Wire it up (preferred if the feature is wanted)**
- [ ] `NewTemplateEditorHandlers(...)` is called in `cmd/server/main.go` with all required dependencies
- [ ] `TemplateEditorHandlers` added to `RouterHandlers` struct
- [ ] Routes registered: `GET /templates/{id}/edit`, `GET|PUT /api/templates/{id}/components`, `POST /api/templates/{id}/components/preview`, `POST /api/templates/{id}/components/validate`
- [ ] `templates/web/template_editor.html` (which exists) is parsed and passed to the handler
- [ ] Existing handler tests continue to pass
- [ ] Manual test: visit `/templates/{id}/edit` for a valid template ID and confirm the editor page renders

**Option B: Delete it (if the feature is deferred)**
- [ ] `internal/handlers/template_editor.go` deleted
- [ ] All associated test files deleted
- [ ] No compile errors
- [ ] Story filed in Epic 10 backlog to re-implement when the feature is prioritised

**Decision must be made before starting this story.**

- [ ] All 32 non-browser packages pass
- [ ] Update `docs/00_BACKLOG/08_API/README.md`: remove ISSUE-2
- [ ] Update `docs/00_BACKLOG/14_BUG_FIXES/README.md`: mark this story complete

---

## Technical Approach (Option A)

### 1. Update `RouterHandlers`

```go
// internal/handlers/router.go
type RouterHandlers struct {
    // ... existing fields
    TemplateEditorHandler TemplateEditorHandlerInterface
}
```

### 2. Register routes in `SetupRoutes`

```go
// Authenticated routes — within the /templates group
r.Get("/{id}/edit", handlers.TemplateEditorHandler.GetEditorPage)
r.Get("/{id}/components", handlers.TemplateEditorHandler.GetComponents)
r.Put("/{id}/components", handlers.TemplateEditorHandler.UpdateComponents)
r.Post("/{id}/components/preview", handlers.TemplateEditorHandler.PreviewComponents)
r.Post("/{id}/components/validate", handlers.TemplateEditorHandler.ValidateComponents)
```

### 3. Wire in `cmd/server/main.go`

```go
templateEditorHandler := handlers.NewTemplateEditorHandlers(templateService, templateRepo)
templateEditorHandler.SetTemplates(templateEditorTemplates)
```

Parse `templateEditorTemplates` from `templates/web/template_editor.html` (which already exists).

---

## Files to Change

- `cmd/server/main.go` — instantiate handler, parse template
- `internal/handlers/router.go` — add to RouterHandlers, register routes

---

## Testing

```bash
go test -timeout 30s ./internal/handlers/...
go test -timeout 30s ./...
```

---

## Status

- **Status:** Not Started
