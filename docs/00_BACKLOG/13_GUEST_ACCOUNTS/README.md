# Epic: Guest Accounts & Encryption at Rest

**Priority:** Medium  
**Status:** Not Started  
**Target Version:** v1  
**Estimated Effort:** 3-4 weeks

---

## Overview

Implement optional passwordless guest accounts so that returning guests can be recognized across multiple events without needing separate invite tokens each time. Simultaneously, encrypt all PII at rest across every table in the database using application-level AES-256-GCM encryption with HMAC-SHA256 blind indexes for searchable fields.

**Goals:**
- Allow guests to optionally create an account identified by email or phone number
- Authenticate guests via OTP (6-digit code delivered to email or SMS) — no passwords ever
- Link existing and future invites to a guest account for cross-event visibility and lost-invite recovery
- Encrypt all PII in the database (email, phone, name, IP address) with keys derived from an operator-supplied environment variable
- Zero changes to the existing staff authentication system (`internal/auth/`, `users` table roles, staff sessions)

**Non-Goals:**
- Passkeys (WebAuthn) — deferred to a future story
- TOTP — deferred
- OIDC for guests — deferred
- Token-only RSVP access is unchanged; accounts are purely additive

---

## Success Criteria

- [ ] `pkg/crypto/` package implements AES-256-GCM encrypt/decrypt and HMAC-SHA256 deterministic hash
- [ ] App fails to start with a clear, actionable error if `TINYRSVP_ENCRYPTION_KEY` is absent
- [ ] All PII columns in `users`, `invites`, `sessions`, and `email_queue` are encrypted at rest
- [ ] Encrypted fields are searchable via HMAC blind index (`WHERE email_hash = ?`)
- [ ] Guest can request OTP to email or phone, verify it, and receive a `tinyrsvp_guest` session cookie
- [ ] Guest session cookie is independent of staff session cookie (`tinyrsvp_session`)
- [ ] Logged-in guest can view all invites linked to their account at `/guest/account`
- [ ] RSVP confirmation page shows an optional account creation prompt when no guest session is present
- [ ] Existing `/rsvp/{token}` flow is completely unchanged — no account required to RSVP
- [ ] OTP requests are rate-limited to 3 per identifier per hour
- [ ] All new code follows TDD with multiple happy and unhappy path tests
- [ ] All tests pass with timeout (`go test -timeout 30s -race ./...`)

---

## User Stories

### Phase 1: Encryption Foundation
- [ ] [`13_STORY_01_crypto_package.md`](13_STORY_01_crypto_package.md) - `pkg/crypto/` Encryptor interface, AES-256-GCM, HKDF key derivation, HMAC blind hash
- [ ] [`13_STORY_02_encrypt_users.md`](13_STORY_02_encrypt_users.md) - Encrypt `users` table PII, update UserRepository, migration 000015
- [ ] [`13_STORY_03_encrypt_invites.md`](13_STORY_03_encrypt_invites.md) - Encrypt `invites` table PII, update InviteRepository
- [ ] [`13_STORY_04_encrypt_sessions.md`](13_STORY_04_encrypt_sessions.md) - Encrypt `sessions` table PII, update SessionRepository
- [ ] [`13_STORY_05_encrypt_email_queue.md`](13_STORY_05_encrypt_email_queue.md) - Encrypt `email_queue` PII, update EmailQueueRepository

### Phase 2: Guest Account Infrastructure
- [ ] [`13_STORY_06_guest_account_models.md`](13_STORY_06_guest_account_models.md) - Guest account models and DB migration 000016
- [ ] [`13_STORY_07_guest_repositories.md`](13_STORY_07_guest_repositories.md) - GuestAccountRepository, GuestSessionRepository, GuestOTPRepository

### Phase 3: Guest Auth Logic
- [ ] [`13_STORY_08_guestauth_package.md`](13_STORY_08_guestauth_package.md) - `internal/guestauth/` package: OTP logic, session manager, account service, OTP delivery interface + email implementation
- [ ] [`13_STORY_09_guest_auth_handlers.md`](13_STORY_09_guest_auth_handlers.md) - HTTP handlers and routing for `/guest/auth/*` and `/guest/account`
- [ ] [`13_STORY_10_guest_auth_middleware.md`](13_STORY_10_guest_auth_middleware.md) - `RequireGuestAuth` middleware

### Phase 4: Integration & Optional Extensions
- [ ] [`13_STORY_11_rsvp_optin_prompt.md`](13_STORY_11_rsvp_optin_prompt.md) - RSVP confirmation page opt-in prompt
- [ ] [`13_STORY_12_sms_otp_delivery.md`](13_STORY_12_sms_otp_delivery.md) - SMS OTP delivery implementation (optional, interface-based)

---

## Dependencies

**Depends on:**
- Epic 00 (Foundation) — database layer, migrations, config
- Epic 01 (Auth) — session infrastructure patterns, middleware patterns
- Epic 03 (Invites) — `invites` table; guest accounts link to invites
- Epic 04 (RSVP) — RSVP confirmation page modified in Story 11
- Epic 05 (Email) — `email.Service` used by OTP email delivery

**Blocks:** Nothing (additive feature; no existing epics depend on it)

**Story-level parallelism:**
- Stories 02, 03, 04, 05 can be worked in parallel after Story 01 is complete
- Stories 07–12 form a sequential dependency chain after Story 06

---

## Technical Overview

### Encryption Architecture

All PII is encrypted at the repository layer. Models always carry plaintext in-process. The `pkg/crypto/` `Encryptor` is injected into each repository that handles PII fields.

```
TINYRSVP_ENCRYPTION_KEY (env var, base64-encoded 32 bytes)
  │
  └─ HKDF-SHA256
       ├─ enc_key → AES-256-GCM → base64(nonce[12] || ciphertext)  [for storage/display]
       └─ idx_key → HMAC-SHA256 → base64(digest)                   [deterministic, for WHERE]
```

**Searchable fields** store both: ciphertext (read when displaying the value) and HMAC hash (used in `WHERE` lookups).  
**Display-only fields** store ciphertext only.

```
Write path:
  "Alice@Example.COM"
    → normalize → "alice@example.com"
    → Encrypt(normalized) → email_encrypted  (stored, varies each write)
    → Hash(normalized)    → email_hash       (stored, constant for same input)

Lookup path:
  input "Alice@Example.COM"
    → normalize → "alice@example.com"
    → Hash(normalized) → h
    → SELECT * FROM users WHERE email_hash = ?  (h)
    → Decrypt(email_encrypted) → "alice@example.com"
```

### Guest Auth Flow

```
┌─────────────────────────────────────────────────────────┐
│  Step 1: Request OTP                                     │
│                                                         │
│  POST /guest/auth/request-otp                           │
│  { "identifier": "alice@example.com" }                  │
│    → normalize + hash identifier                        │
│    → rate-limit check (≤3/hr per identifier)            │
│    → lookup or create guest_accounts row                │
│    → generate 6-digit code, store HMAC(code)           │
│    → deliver code via email (or SMS)                    │
│    → 200 OK                                             │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Step 2: Verify OTP                                      │
│                                                         │
│  POST /guest/auth/verify-otp                            │
│  { "identifier": "alice@example.com", "code": "123456" }│
│    → hash identifier → find pending OTP row             │
│    → constant-time compare HMAC(code) == code_hash      │
│    → check: not expired (15-min TTL)                    │
│    → check: not already used                            │
│    → mark used_at, create guest_session                 │
│    → set tinyrsvp_guest cookie (30d)                    │
│    → 200 OK                                             │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Step 3: View Account                                    │
│                                                         │
│  GET /guest/account  (tinyrsvp_guest cookie)            │
│    → RequireGuestAuth middleware validates session      │
│    → inject *models.GuestAccount into context           │
│    → list invites WHERE guest_account_id = ?            │
│    → render page                                        │
└─────────────────────────────────────────────────────────┘
```

### Role Hierarchy (updated)

```
Admin (full control)
  └─> Event Manager (own events only)
      └─> Guest with Account (own invites, read-only)
          └─> Anonymous Guest (token-based, no account)
```

Guest accounts have no access to any staff routes. `RequireGuestAuth` and `RequireAuth` are entirely separate middleware.

### New Tables (Migration 000016)

```
guest_accounts
  id, email_encrypted, email_hash, phone_encrypted, phone_hash,
  display_name_encrypted, created_at, updated_at
  CHECK (email_hash IS NOT NULL OR phone_hash IS NOT NULL)

guest_sessions
  id (32-byte random), guest_account_id → guest_accounts.id,
  created_at, expires_at (30d), last_accessed_at,
  ip_address_encrypted, user_agent_encrypted

guest_otp_codes
  id, guest_account_id → guest_accounts.id,
  identifier_hash, code_hash, purpose ('login'|'enroll'),
  created_at, expires_at (15m), used_at

invites.guest_account_id → guest_accounts.id  (added in 000016)
```

### Package Structure

```
pkg/
  crypto/
    crypto.go          # Encryptor interface, AES-256-GCM, HKDF, HMAC
    crypto_test.go

internal/
  guestauth/
    otp.go             # OTP generation, hashing, validation
    otp_test.go
    session.go         # GuestSessionManager
    session_test.go
    service.go         # GuestAccountService
    service_test.go
    delivery.go        # OTPDelivery interface + EmailOTPDelivery
    delivery_test.go
    ratelimit.go       # Per-identifier rate limiting
    ratelimit_test.go
  models/
    guest_account.go   # GuestAccount, GuestSession, GuestOTPCode
  db/repositories/
    guest_account_repository.go
    guest_session_repository.go
    guest_otp_repository.go
  handlers/
    guest_auth.go
    guest_auth_test.go
  middleware/
    guest.go           # RequireGuestAuth
    guest_test.go
```

---

## Technical Decisions

### Separate Auth Layer, Not Extended Staff Auth
Guest identity is structurally different from staff identity. Staff authenticate to manage events; guests authenticate to access their invitations. Merging them into the `users` table would pollute staff auth invariants (bootstrap admin logic, role CHECK constraints, OIDC subject column) and create risk of privilege escalation. The existing `internal/auth/` code is not touched.

### Application-Level Encryption, Not SQLCipher
The project targets both SQLite and PostgreSQL. SQLCipher is SQLite-only. Application-level AES-256-GCM works identically against both databases, requires no driver changes, and keeps encryption logic reviewable in Go code.

### HMAC Blind Index for Searchable Fields
AES-GCM with a random nonce produces different ciphertext on every write — SQL `WHERE email_encrypted = ?` is impossible. Storing `HMAC-SHA256(normalize(value), idx_key)` alongside the ciphertext enables exact-match lookups without storing plaintext. This is the same pattern already used in the `invites` table for `token_hash`.

### Single Master Key with HKDF Derivation
A single `TINYRSVP_ENCRYPTION_KEY` env var is simpler for homelab operators. Two derived subkeys (one for encryption, one for hashing) are separated via HKDF so that knowing one reveals nothing about the other.

### OTP over TOTP for Guest Enrollment
TOTP requires a shared secret that must be delivered out-of-band before it can be used — effectively the same delivery problem as a magic link for first-time enrollment. Starting with email/phone OTP eliminates this bootstrapping problem and covers the primary use case with no extra complexity.

### 15-Minute OTP TTL, Single-Use
Short TTL limits the window for code interception. Single-use enforcement (`used_at`) prevents replay attacks.

### 30-Day Guest Session TTL
Matches the invite token expiry window, keeping the guest session valid for as long as their invites are accessible.

---

## Security Considerations

### Encryption Key Management
- `TINYRSVP_ENCRYPTION_KEY` must be 32 random bytes, base64-encoded
- App refuses to start without it; error message includes `openssl rand -base64 32` generation hint
- Key is never stored in the database, never logged
- Key rotation is not in scope for this epic (track in Epic 10)

### OTP Security
- 6-digit numeric codes: 1-in-1,000,000 probability per attempt
- Rate limiting (3/hr per identifier) prevents enumeration
- Codes stored as `HMAC-SHA256(code, idx_key)` — never plaintext
- Constant-time comparison on verification prevents timing attacks
- `used_at` enforcement prevents replay

### Guest Session Security
- Session IDs are 32-byte cryptographically random values
- Cookie attributes: `tinyrsvp_guest`, HttpOnly, Secure, SameSite=Lax, 30-day Max-Age
- Sessions deleted on explicit logout or account deletion

### Isolation from Staff Auth
- `RequireGuestAuth` middleware uses a distinct context key from `RequireAuth`
- Guest sessions live in `guest_sessions`, not `sessions`
- No code path allows a guest session to satisfy `RequireAuth`
- Guest accounts have no `role` field and cannot be promoted

### PII Protection
- Plaintext PII is never written to the database
- HMAC blind indexes reveal nothing about plaintext beyond equality with a known value
- IP addresses in both `sessions` and `guest_sessions` are encrypted

---

## Testing Strategy

### Unit Tests
- `pkg/crypto/`: encrypt→decrypt roundtrip, hash determinism, wrong-key failure, tampered ciphertext, empty input, key too short
- OTP: correct length and format, uniqueness across samples, expired rejection, already-used rejection, wrong code rejection, constant-time compare
- `GuestAccountService`: lookup by email hash, lookup by phone hash, create new account, link invite to account
- Rate limiting: under limit passes, exactly at limit passes, over limit fails, window resets after 1 hour

### Integration Tests
- Repository roundtrips: create guest account, retrieve by hash, confirm raw DB row contains no plaintext
- Full OTP flow against real DB: request → verify → session created → session validated
- Session expiry: expired session rejected by middleware

### Regression Tests
- All existing repository tests pass after PII columns are renamed in migration 000015
- `/rsvp/{token}` flow works without a guest session cookie present

### Manual Tests
1. Set `TINYRSVP_ENCRYPTION_KEY` and start the app; confirm it starts
2. Unset `TINYRSVP_ENCRYPTION_KEY`; confirm app refuses to start with clear message
3. Request OTP to a real email; confirm delivery and 6-digit code
4. Verify OTP; confirm `tinyrsvp_guest` cookie is set
5. Navigate to `/guest/account`; confirm invites appear
6. Verify an expired code (wait 16 minutes) is rejected
7. Verify a used code cannot be reused
8. Submit an RSVP without a guest session; confirm flow is unchanged
9. Inspect raw SQLite DB; confirm no plaintext emails visible

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Lost encryption key makes all data unreadable | High | Document key backup requirement clearly in operator docs; recommend Docker secrets |
| HMAC blind index collision | Low | HMAC-SHA256 preimage resistance makes collision computationally infeasible |
| OTP brute force | Medium | Rate limiting (3/hr/identifier) + 15-min TTL + single-use enforcement |
| Guest session hijacking | Medium | HttpOnly + Secure cookie, 32-byte random session IDs |
| SMS provider unavailable | Low | SMS delivery is optional and gracefully disabled if not configured |
| Existing tests break after column renames in migration 000015 | Medium | Run full test suite after each story; fix all failures before continuing |
| Performance regression from per-row encryption | Low | AES-GCM operations are ~1µs; not on latency-critical path |

---

## References

- **HLD:** Section 4 (Authentication & Authorization), Section 13 (Database Schema), Section 16 (Security)
- **LLD:** [`lld/01_AUTH_LLD.md`](../../02_DESIGN/lld/01_AUTH_LLD.md) — session patterns and interfaces
- **Go stdlib:** `crypto/aes`, `crypto/cipher`, `crypto/hmac`, `crypto/sha256`
- **External:** `golang.org/x/crypto/hkdf`
- **Existing pattern reference:** `pkg/token/` — HMAC-SHA256 token hashing (same blind index pattern)

---

## Definition of Done

- [ ] All 12 user stories complete
- [ ] All PII columns encrypted in `users`, `invites`, `sessions`, `email_queue`, and all guest tables
- [ ] App fails to start without `TINYRSVP_ENCRYPTION_KEY`
- [ ] Guest OTP flow working end-to-end: request OTP → verify → session → view invites
- [ ] RSVP flow unchanged — no account required
- [ ] All existing tests still pass
- [ ] All new tests pass: `go test -timeout 30s -race ./...`
- [ ] No `map[string]interface{}` in new code
- [ ] Security review: no plaintext PII in raw DB rows
- [ ] Operator documentation updated (env var, key generation, Docker secrets guidance)
