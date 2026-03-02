# Epic: Invite & Token Management

**Priority:** High  
**Status:** Not Started  
**Target Version:** v0  
**Estimated Effort:** 1.5 weeks

---

## Overview

Implement secure token-based guest access system. Support individual invites, bulk CSV import, and manual token generation. Ensure cryptographically secure tokens with HMAC-SHA256 hashing and proper lifecycle management.

**Goal:** Enable event managers to invite guests via email or manual distribution, with secure token-based access that doesn't require guest accounts.

---

## Success Criteria

- [ ] Event managers can create individual invites with email
- [ ] Bulk CSV import supports up to 500 guests
- [ ] Tokens are cryptographically secure (256-bit)
- [ ] Tokens hashed with HMAC-SHA256 in database
- [ ] Token validation uses constant-time comparison
- [ ] Tokens expire 30 days after event date
- [ ] Tokens can be revoked by event manager
- [ ] Tokens can be regenerated if compromised
- [ ] Invite status tracked (draft/sent/viewed/responded)

---

## User Stories

### Phase 1: Token Infrastructure
- [ ] [`03_STORY_00_token_generation.md`](03_STORY_token_generation.md) - Cryptographically secure token generation
- [ ] [`03_STORY_01_token_hashing.md`](03_STORY_token_hashing.md) - HMAC-SHA256 hashing implementation
- [ ] [`03_STORY_02_token_validation.md`](03_STORY_token_validation.md) - Constant-time token validation

### Phase 2: Invite Management
- [ ] [`03_STORY_03_invite_model.md`](03_STORY_invite_model.md) - Invite struct and repository
- [ ] [`03_STORY_04_individual_invite.md`](03_STORY_individual_invite.md) - Create single invite
- [ ] [`03_STORY_05_bulk_csv_import.md`](03_STORY_bulk_csv_import.md) - CSV import with validation
- [ ] [`03_STORY_06_manual_invite.md`](03_STORY_manual_invite.md) - Generate invite without email

### Phase 3: Token Lifecycle
- [ ] [`03_STORY_07_token_expiration.md`](03_STORY_token_expiration.md) - Token expiration and cleanup
- [ ] [`03_STORY_08_token_revocation.md`](03_STORY_token_revocation.md) - Revoke compromised tokens
- [ ] [`03_STORY_09_token_regeneration.md`](03_STORY_token_regeneration.md) - Regenerate tokens

### Phase 4: Invite Status
- [ ] [`03_STORY_10_invite_tracking.md`](03_STORY_invite_tracking.md) - Track invite status transitions
- [ ] [`03_STORY_11_invite_listing.md`](03_STORY_invite_listing.md) - List and filter invites

---

## Dependencies

**Depends on:** Epic 00 (Foundation), Epic 01 (Auth), Epic 02 (Events)  
**Blocks:** Epic 04 (RSVP), Epic 05 (Email)

---

## Technical Overview

### Token Generation & Hashing

```
1. Generate 32 random bytes (crypto/rand)
2. Base64-URL encode → token (43 chars)
3. HMAC-SHA256(secret, token) → hash
4. Store hash in database (never plain token)
5. Return token to caller (one-time display)
```

### Token Validation

```
1. Receive token from URL
2. HMAC-SHA256(secret, token) → computed_hash
3. Constant-time compare with stored hash
4. Check invite not revoked
5. Check token not expired
6. Grant access if all checks pass
```

### RSVP URL Format

```
https://rsvp.example.com/rsvp/{token}

Example:
https://rsvp.example.com/rsvp/a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5k
```

### CSV Import Format

```csv
name,email,max_plus_ones
John Doe,john@example.com,2
Jane Smith,jane@example.com,1
Bob Johnson,bob@example.com,0
```

---

## Technical Decisions

### HMAC-SHA256 vs Bcrypt
- Tokens are random (not user passwords)
- HMAC provides authentication + integrity
- Constant-time comparison prevents timing attacks
- Faster than bcrypt (better guest experience)
- Appropriate for high-entropy tokens

### Token Length: 256 bits
- 2^256 possible values (unguessable)
- 43 characters base64-URL encoded
- URL-safe (no special characters)
- Fits in database VARCHAR(255)

### Secret Key Management
- Generated on first startup (crypto/rand)
- Stored in database config table
- Never logged or exposed
- Rotation invalidates all tokens

### Invite Status Tracking
- DRAFT: Created but not sent
- SENT: Email sent to guest
- VIEWED: Guest opened RSVP page
- RESPONDED: Guest submitted RSVP
- REVOKED: Cancelled by organizer

---

## Security Considerations

### Token Security
- Cryptographically random generation
- HMAC prevents forgery
- Constant-time comparison prevents timing attacks
- Tokens never logged (only last 6 chars shown in UI)
- HTTPS required for transmission

### CSV Import Security
- Max 500 rows per upload
- Email validation
- Duplicate detection
- Malicious content scanning
- Error reporting without exposing data

### Token Lifecycle
- Automatic expiration (event date + 30 days)
- Manual revocation support
- Regeneration invalidates old token
- Cleanup job removes expired tokens

---

## Validation Rules

### Email Format
- Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
- Max 255 characters
- Local part < 64 chars
- Domain < 255 chars

### Guest Name
- Optional
- Max 100 characters
- Sanitized for XSS

### Max Plus Ones
- Integer 0-10
- Cannot exceed event.max_plus_ones
- Default: event.max_plus_ones

### CSV Validation
- Header row required
- Email column required
- Max 500 rows
- Duplicate emails rejected
- Invalid emails reported

---

## References

- **HLD:** Section 6 (Invite & Guest Access Model)
- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md)
- **Database:** invites table
- **Security:** Section 16.3 (Input Sanitization), 16.4 (CSRF Protection)

---

## Testing Strategy

### Unit Tests
- Token generation randomness
- HMAC hashing correctness
- Constant-time comparison
- CSV parsing and validation
- Invite status transitions

### Integration Tests
- Full invite creation flow
- Bulk import with various CSV formats
- Token validation with database
- Expiration and cleanup
- Revocation and regeneration

### Security Tests
- Timing attack resistance
- Token forgery attempts
- CSV injection attempts
- Duplicate token prevention

### Edge Cases
- Empty CSV
- Malformed CSV
- Duplicate emails
- Invalid email formats
- Expired tokens
- Revoked tokens

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Token collision | Critical | Use crypto/rand, 256-bit tokens |
| Timing attacks | High | Constant-time comparison |
| CSV injection | Medium | Sanitize all input, validate format |
| Token leakage | High | Never log tokens, HTTPS required |
| Secret key compromise | Critical | Secure storage, rotation support |

---

## Definition of Done

- [ ] All user stories complete
- [ ] Token generation cryptographically secure
- [ ] HMAC hashing implemented correctly
- [ ] Constant-time validation working
- [ ] CSV import handles 500+ rows
- [ ] All validation rules enforced
- [ ] Token lifecycle fully functional
- [ ] Security review passed
- [ ] All tests passing
- [ ] Documentation updated
