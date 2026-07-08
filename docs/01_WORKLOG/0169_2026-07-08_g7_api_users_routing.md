# Worklog 0169: Replace Hand-Rolled /api/users Routing with Chi Routes (G7)

**Date:** 2026-07-08  
**Epic:** 10 (Technical Debt)  
**Branch:** `cleanup/G7-api-users-routing`  
**PR:** #42

---

## Summary

Replaced the hand-rolled path-parsing routing for `/api/users/{id}` with proper chi routes. Eliminated ~60 lines of manual path splitting, `_method` override handling, and per-method auth middleware wrapping.

## Changes

### Router (`internal/handlers/router.go`)

**Before** (60 lines): manual `strings.TrimPrefix`, `strings.Split`, switch on method, `RequireAuth(RequireAdmin(http.HandlerFunc(...)))` repeated 3 times.

**After** (20 lines):
```go
r.Route("/api/users", func(r chi.Router) {
    r.Use(AuthMiddleware.RequireAuth)
    r.Use(AuthMiddleware.RequireAdmin)
    r.Use(methodOverrideMiddleware)
    r.Get("/", ListUsers)
    r.Get("/{id}", GetUser)
    r.Patch("/{id}", UpdateUserRole)
    r.Delete("/{id}", DeleteUser)
})
```

### Handler signatures (`internal/handlers/users.go`)

Changed `GetUser`, `UpdateUserRole`, `DeleteUser` from taking `(w, r, userID string)` to `(w, r)` only, reading `userID` from `chi.URLParam(r, "id")` — matching the pattern used by all other handlers.

### Tests

- `internal/handlers/users_test.go` and `users_integration_test.go`: added `reqWithUserIDParam` helper to set chi URL params on test requests
- `tests/e2e/auth_flow_test.go`: updated to use `chi.NewRouteContext` for URL param injection
- `internal/handlers/router_real_handlers_test.go`: updated mock signatures

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** Full suite green
