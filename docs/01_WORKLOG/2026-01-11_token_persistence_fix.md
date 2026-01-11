# Token Persistence Fix - 2026-01-11

## Issue
After server reboot, invite tokens became invalid with error: "Invite not found or has been revoked" (HTTP 404)

## Root Cause
The `TOKEN_SECRET` environment variable was not set, causing the application to generate a **random HMAC secret on each startup**. Since invite tokens are stored as `HMAC-SHA256(token, secret)` hashes, changing the secret made existing tokens unfindable.

**Flow:**
```
Server Start #1:
  TOKEN_SECRET not set → Random secret "ABC123"
  Create invite → Hash = HMAC-SHA256(token, "ABC123") → Store in DB

Server Reboot:
  TOKEN_SECRET not set → NEW random secret "XYZ789"
  
Access invite:
  Compute hash = HMAC-SHA256(token, "XYZ789")
  Query DB → NOT FOUND (DB has hash from "ABC123")
```

## Solution Implemented

### 1. Hardcoded Fallback Secret
When `TOKEN_SECRET` is not set, the system now uses a **hardcoded fallback secret** instead of generating a random one. This ensures tokens persist across restarts.

**Location:** [`internal/config/config.go:244`](../internal/config/config.go:244)

```go
func getHardcodedTokenSecret() string {
    return "tinyrsvp_default_token_secret_change_in_production_da8f152a3cc3d58054cb988a463344503ad1ad09fba718a8a5e6e9513d16040f"
}
```

### 2. Configurable Token Hashing
Added `TOKEN_HASHING_ENABLED` environment variable (default: `true`)

**Modes:**

| Mode | Config | Behavior | Security | Operations |
|------|--------|----------|----------|------------|
| **HMAC (Default)** | `TOKEN_HASHING_ENABLED=true`<br>`TOKEN_SECRET=<custom>` | Tokens hashed with custom secret | High - DB breach doesn't expose usable tokens | Requires secret management |
| **HMAC Fallback** | `TOKEN_HASHING_ENABLED=true`<br>`TOKEN_SECRET` not set | Tokens hashed with hardcoded secret | Medium - Tokens persist but use known secret | Simple - works out of box |
| **Plain Token** | `TOKEN_HASHING_ENABLED=false` | Tokens stored in plain text | Lower - DB breach exposes usable tokens | Simplest - no secret needed |

### 3. Warning Messages
The system now displays clear warnings on startup:

**When TOKEN_SECRET not set (HMAC mode):**
```
WARNING: TOKEN_SECRET not set - using hardcoded fallback
WARNING: Tokens will persist across restarts but use a known secret
WARNING: For production, set TOKEN_SECRET environment variable
WARNING: Generate with: openssl rand -hex 32
```

**When hashing disabled:**
```
WARNING: Token hashing disabled (TOKEN_HASHING_ENABLED=false)
WARNING: Invite tokens will be stored in plain text in the database
WARNING: This reduces security but simplifies operations
```

## Configuration

### Production Deployment (Recommended)
```yaml
environment:
  - TOKEN_SECRET=<generate-with-openssl-rand-hex-32>
  - TOKEN_HASHING_ENABLED=true  # default, can omit
```

### Homelab/Development (Acceptable)
```yaml
environment:
  # TOKEN_SECRET not set - uses hardcoded fallback
  # Tokens persist across restarts
```

### Plain Token Mode (Simplest)
```yaml
environment:
  - TOKEN_HASHING_ENABLED=false
  # No secret needed, tokens stored in plain text
```

## Files Changed

1. [`docker-compose.yml`](../docker-compose.yml) - Added TOKEN_SECRET with fallback
2. [`docker-compose.test.yml`](../docker-compose.test.yml) - Added fixed TOKEN_SECRET for tests
3. [`.gitleaks.toml`](../.gitleaks.toml) - Allowlisted test secret
4. [`internal/config/config.go`](../internal/config/config.go) - Implemented configurable strategy
5. [`pkg/token/generator.go`](../pkg/token/generator.go) - Added plain token mode support
6. [`cmd/server/main.go`](../cmd/server/main.go) - Updated token generator initialization

## Tests Added

1. [`internal/config/token_strategy_test.go`](../internal/config/token_strategy_test.go) - Config strategy tests
2. [`pkg/token/plain_mode_test.go`](../pkg/token/plain_mode_test.go) - Token generator mode tests

**All tests pass:** ✅

## Migration Notes

**Existing Invites:**
- Invites created before this fix are **invalid** (hashed with random secrets)
- Must regenerate tokens for existing invites
- Or delete old database and start fresh

**No Breaking Changes:**
- Default behavior uses hardcoded fallback (tokens persist)
- Existing production deployments with TOKEN_SECRET continue working
- New deployments work out-of-box

## Security Considerations

### HMAC Mode (Default)
**Pros:**
- Database breach doesn't expose usable tokens
- Defense in depth
- Prevents token forgery

**Cons:**
- Requires secret management
- Lost secret = all tokens invalid

### Plain Token Mode
**Pros:**
- No secret management
- Simpler operations
- Tokens naturally persist

**Cons:**
- Database breach = immediate access
- Less defense in depth

**Recommendation:** For homelab use with cryptographically random tokens (256-bit entropy), either mode is acceptable. Choose based on operational preferences.

## Verification

To verify the fix:
1. Create an invite
2. Note the token from the URL
3. Restart the server
4. Access the same invite URL
5. Should load successfully (previously showed 404)

## Related Issues

- Original bug report: Invite tokens invalid after reboot
- Log entry: `GET /rsvp/aAPewpl9F_bE4c9YAiOOA8obl_x5R_0MMGf6pqfvlkQ 404`
