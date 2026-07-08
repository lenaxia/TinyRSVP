# Worklog 0171: Router God-Object Refactor (G6)

**Date:** 2026-07-08  
**Epic:** 10 (Technical Debt)  
**Branch:** `cleanup/G6-router-refactor`  
**PR:** #45

---

## Summary

Split the 400-line `NewRouter` monolith into 7 focused route-registration functions in a new `router_setup.go` file. `NewRouter` is now a 65-line orchestrator that calls the helpers.

## Changes

### Before
`internal/handlers/router.go`: 716 lines
- `NewRouter`: 400 lines, ~42 conditional `if handlers.X != nil` blocks
- Hard to find any specific route
- Adding a new route required scrolling through the entire function

### After
`internal/handlers/router.go`: 389 lines (types, interfaces, helpers)
`internal/handlers/router_setup.go`: 367 lines (route registration)

7 extracted functions:
1. `setupMiddleware` — global middleware chain
2. `registerInfrastructureRoutes` — health, ready, metrics, CSP
3. `registerAuthRoutes` — login, logout, OIDC, forward-auth
4. `registerPageRoutes` — dashboard, admin, settings, metrics pages
5. `registerAPIRoutes` — /api/* events, invites, images, templates, users
6. `registerRSVPRoutes` — RSVP, confirmation, unsubscribe, calendar
7. `registerStaticRoutes` — static files, assets

## What didn't change
- `RouterHandlers` struct — unchanged
- All routing behavior — identical
- All existing tests pass without modification

## Status
**Status:** ✅ Complete  
**Test Pass Rate:** Full suite green
