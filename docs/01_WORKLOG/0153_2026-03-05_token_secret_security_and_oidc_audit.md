# 0153 — TOKEN_SECRET Security Fix + OIDC Audit + Doc Cleanup

**Date:** 2026-03-05
**Session goal:** Fix TOKEN_SECRET hardcoded fallback (security risk), create `.env.example`, audit OIDC implementation status, and add honest disclaimers for unimplemented features.

---

## Work Done

### 1. TOKEN_SECRET Security Fix

**Problem:** `internal/config/config.go` had `getHardcodedTokenSecret()` returning a known hex string as a fallback when `TOKEN_SECRET` was unset. Anyone who deployed without setting the env var was silently using a public, predictable secret.

**Fix:**
- Removed `getHardcodedTokenSecret()` entirely
- `validateToken()` now returns a hard error if `TOKEN_SECRET` is empty when hashing is enabled
- `docker-compose.yml` changed from `${TOKEN_SECRET:-<hex>}` to `${TOKEN_SECRET:?TOKEN_SECRET is required. Generate with: openssl rand -hex 32}`

**Config tests updated:**
- 4 tests in `config_test.go` that called `Load()` without `TOKEN_SECRET` set now provide one
- `TestConfig_TokenStrategy_HMAC_WithHardcodedFallback` replaced with `TestConfig_TokenStrategy_HMAC_NoSecret_FailsToStart`

---

### 2. `.env.example` Created

New annotated file covering all 35 env vars with sections:
- Server
- Database (SQLite only note)
- Token Security
- Authentication (OIDC disclaimer + Forward Auth)
- Email / SMTP
- Storage (local only, S3 planned)
- Performance tuning
- Logging

---

### 3. S3 / PostgreSQL Documentation Fix

`internal/storage/provider.go:53` hard-returns `"s3 storage provider not yet implemented"`.
`internal/db/` only has SQLite migrations — no postgres path.

Removed all implied functionality from README and `.env.example`. Both are now clearly marked as "planned for v1".

---

### 4. OIDC Implementation Audit

**Finding:** OIDC is **fully implemented** at the code level:
- `internal/auth/oidc.go`: Complete login redirect, state cookie CSRF protection, callback with OAuth2 code exchange, ID token verification via `go-oidc`, claims parsing
- `internal/auth/handlers.go`: Login, callback, and logout HTTP handlers wired to the Authenticator interface
- `cmd/server/main.go:236-251`: OIDC conditionally initialized at startup from config

**Gap:** No integration tests against a real provider. The callback flow (code exchange, token verification) has no test coverage beyond construction/input validation with a mock OIDC discovery endpoint. Forward Auth is the tested path in this beta.

**Docs updated:**
- `.env.example`: Added "implemented but not integration-tested" note to Option A OIDC section; removed "recommended for most setups"
- `README.md`: Updated feature bullet, tech stack line, and security section to say "Forward Auth (tested) or OIDC (implemented, not integration-tested in beta)"

---

## Files Changed

| File | Change |
|------|--------|
| `internal/config/config.go` | Removed `getHardcodedTokenSecret()`; hard fail on empty `TOKEN_SECRET` |
| `internal/config/config_test.go` | Added `TOKEN_SECRET` to 4 tests |
| `internal/config/token_strategy_test.go` | Replaced hardcoded-fallback test with `NoSecret_FailsToStart` |
| `docker-compose.yml` | `TOKEN_SECRET` now uses `:?` hard-fail syntax |
| `.env.example` | Created — full annotated 35-var example |
| `README.md` | Roadmap fixed, S3/PG removed, OIDC disclaimer added |

---

## Test Status

All 33 packages passing (`go test -count=1 ./...`).

---

## Next Steps

- Consider bootstrapping first admin without external auth provider being required first (UX gap for /r/selfhosted users)
- CHANGELOG.md or tagged release
- LICENSE file (currently "TBD")
- Epic 09 (Security/OWASP audit) — needed before public internet deployment
