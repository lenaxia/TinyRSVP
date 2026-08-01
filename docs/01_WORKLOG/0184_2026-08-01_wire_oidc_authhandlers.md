# Worklog: Wire Up OIDC AuthHandlers

**Date:** 2026-08-01  
**Branch:** `feat/wire-oidc-authhandlers-v2`  
**PR:** #73

## Summary

Completed the OIDC auth refactor (C4). The `handlers.AuthHandlers` were built and tested but never wired — production used the legacy path, and `OIDCCallback` didn't create a session. This finishes the refactor: completes the callback flow, wires production, and removes the legacy handlers.

## Changes

- `AuthHandlers` now holds `userService` + `sessionMgr`; `NewAuthHandlers` takes all three deps.
- `OIDCCallback` completes the full session flow: `GetOrCreateUser` → `UpdateLastLogin` → `CreateSession` → `SetSessionCookie` → return-URL resolution (query → cookie) → cookie cleanup → redirect. Guards against nil result.
- `Logout` clears the session cookie via `ClearSessionCookie`.
- `main.go` wires `AuthHandlers`; `RouterHandlers` drops the legacy fields.
- Unified `AuthResult` to use `auth.AuthResult` (removed the duplicate `handlers.AuthResult`).
- Removed legacy `auth.LoginHandler`/`CallbackHandler`/`LogoutHandler` + their tests.
- Routes: `/login`, `/auth/oidc/login`, `/auth/oidc/callback`, `/logout`.

## Tests
- Auth callback tests (success, auth error, GetOrCreateUser error, CreateSession error, SetSessionCookie error, return-URL cookie fallback).
- Router tests updated for new routes.
- Full suite: all 39 non-browser packages pass; build/vet clean.

## Status
**Status:** ✅ Complete  
**Test Pass Rate:** All non-browser tests pass  
**Confidence:** HIGH
