# Worklog: Remove Dead Code

**Date:** 2026-08-01  
**Branch:** `chore/remove-dead-code`

## Summary

Removed genuinely dead code identified in the deep tech-debt review. Every item was verified to have zero production callers before deletion.

## Removed

1. **`audit_log` table** — created in migration 000001 but never read/written by any Go code (no repository, no model, no queries). Dropped via migration 000015.
2. **`config` table + `ConfigRepository` + `models.Config`** — full CRUD repository with 9 tests + generated mock, but `main.go` never instantiated it. Completely orphaned. Dropped table via migration 000015; deleted repository, model, test, and mock.
3. **`pkg/token/validator.go`** — ~400 lines implementing constant-time token comparison via `Validator.Validate`. Never called in production (the app validates tokens by hashing + DB lookup via `InviteRepository.GetByTokenHash`). Deleted validator + 2 test files.
4. **`internal/handlers/router_docs.go`** — 150-line `RouterDocumentation` constant. Never referenced, printed, or served. Content had drifted (references non-existent middleware). Deleted.
5. **`templates/web/not_implemented.html`** — orphaned template. The `/not-implemented` route serves an inline constant instead. Never parsed or rendered. Deleted.
6. **Migration 000015** — `DROP TABLE IF EXISTS audit_log; DROP TABLE IF EXISTS config;` with a down migration that recreates them.

## Verification

All removals verified via `grep -rn` returning zero production callers. Migration tests updated to not expect the dropped tables. Full suite: all 39 non-browser packages pass; build/vet clean.

## Not removed (intentional)

- Dead interface methods (e.g. `EventRepository.GetComponentOverrides`, `AnswerRepository.Update`, `EmailQueueRepository.MarkCancelled`) — lower value, higher churn (cascades to mocks). Left for a future cleanup pass.
- Dead static assets (`layout.css`, `navigation.css`) — low value, separate concern.
- `SECURITY_HMAC_SECRET` config — flagged for product decision, not dead code per se (validated + displayed).

## Status
**Status:** ✅ Complete  
**Test Pass Rate:** All non-browser tests pass  
**Confidence:** HIGH
