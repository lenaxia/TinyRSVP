# Worklog 0172: Remove Dead DB-Backed HMAC Secret (G12)

**Date:** 2026-07-08  
**Epic:** 10 (Technical Debt)  
**Branch:** `cleanup/G12-remove-dead-hmac-db-secret`

---

## Summary

Removed the unused DB-backed HMAC secret methods (`GetHMACSecret`, `SetHMACSecret`, `generateAndStoreHMACSecret`) from `ConfigRepository`. These were dead code — defined but never called outside of tests.

## Root Cause

The codebase had two sources of truth for the HMAC secret:
1. **Environment variable** `SECURITY_HMAC_SECRET` → `config.Security.HMACSecretKey` — used by `main.go` for token generation
2. **Database `config` table** key `"hmac_secret"` → `ConfigRepository.GetHMACSecret()` — **never called in production code**

The DB-backed methods were defined in the repository interface and implemented, but no production code ever called `GetHMACSecret` or `SetHMACSecret`. They existed only in tests and the repository itself (circular: `GetHMACSecret` calls `generateAndStoreHMACSecret` which calls `SetHMACSecret`).

## Fix

Removed from `ConfigRepository`:
- `GetHMACSecret(ctx) ([]byte, error)` — dead code
- `SetHMACSecret(ctx, secret []byte) error` — dead code
- `generateAndStoreHMACSecret(ctx) ([]byte, error)` — private, dead code
- `const hmacSecretKey = "hmac_secret"` — unused constant

Removed unused imports: `crypto/rand`, `encoding/base64`.

Removed 3 tests that tested the dead methods. Regenerated `MockConfigRepository`.

The HMAC secret now comes exclusively from the `SECURITY_HMAC_SECRET` environment variable, with no DB fallback. This is the correct single source of truth.

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** Full suite green
