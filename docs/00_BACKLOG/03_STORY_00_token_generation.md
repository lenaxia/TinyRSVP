# User Story: Token Generation

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 0.5 days
**Completed:** 2026-01-07

---

## User Story

As a **system developer**, I want **cryptographically secure token generation** so that **guest invite tokens are unguessable and secure**.

---

## Acceptance Criteria

- [x] Tokens generated using `crypto/rand` (not `math/rand`)
- [x] Tokens are 256 bits (32 bytes) of random data
- [x] Tokens are base64-URL encoded (43 characters)
- [x] Each token is unique (no collisions)
- [x] Token generation fails safely if randomness unavailable
- [x] Token format is URL-safe (no special characters requiring encoding)
- [x] Generator interface allows for testing with mocks
- [x] Generator can be initialized with secret key for HMAC

---

## Technical Details

### Package Location
- `pkg/token/generator.go`
- `pkg/token/generator_test.go`

### Interface Definition

```go
type Generator interface {
    Generate() (string, error)
    Hash(token string) (string, error)
}
```

### Implementation Requirements

1. **Token Generation**
   - Use `crypto/rand.Read()` to generate 32 random bytes
   - Base64-URL encode the bytes
   - Return 43-character string
   - Handle errors from random number generator

2. **Token Hashing**
   - Use HMAC-SHA256 with secret key
   - Base64-URL encode the hash
   - Return hash string for database storage

3. **Secret Key Management**
   - Accept secret key in constructor
   - Store secret key securely in struct
   - Never expose secret key in logs or errors

### Token Format

```
Input:  32 random bytes
Output: "a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p" (43 chars)
```

---

## Subtasks

### Implementation
- [x] Create `pkg/token/` directory
- [x] Define `Generator` interface in `generator.go`
- [x] Implement `generator` struct with secret key field
- [x] Implement `NewGenerator(secret []byte) Generator` constructor
- [x] Implement `Generate() (string, error)` method
- [x] Implement `Hash(token string) (string, error)` method
- [x] Handle errors from `crypto/rand.Read()`

### Testing
- [x] Test token uniqueness (generate 1000 tokens, verify no duplicates)
- [x] Test token length (exactly 43 characters)
- [x] Test token format (URL-safe base64)
- [x] Test hash consistency (same token produces same hash)
- [x] Test hash uniqueness (different tokens produce different hashes)
- [x] Test error handling (mock crypto/rand failure)
- [x] Test with different secret keys (different secrets produce different hashes)
- [x] Benchmark token generation performance

### Documentation
- [x] Add package documentation
- [x] Document security considerations
- [x] Add usage examples
- [x] Document error conditions

---

## Dependencies

**Depends on:**
- None (foundation package)

**Blocks:**
- Story 01: Token Hashing
- Story 02: Token Validation
- Story 03: Invite Model

---

## Testing Strategy

### Unit Tests

1. **Token Generation Tests**
   ```go
   func TestGenerator_Generate_Uniqueness(t *testing.T)
   func TestGenerator_Generate_Length(t *testing.T)
   func TestGenerator_Generate_URLSafe(t *testing.T)
   func TestGenerator_Generate_Error(t *testing.T)
   ```

2. **Token Hashing Tests**
   ```go
   func TestGenerator_Hash_Consistency(t *testing.T)
   func TestGenerator_Hash_Uniqueness(t *testing.T)
   func TestGenerator_Hash_DifferentSecrets(t *testing.T)
   ```

3. **Edge Cases**
   - Empty secret key
   - Nil secret key
   - Very long secret key
   - Concurrent token generation

### Performance Tests

```go
func BenchmarkGenerator_Generate(b *testing.B)
func BenchmarkGenerator_Hash(b *testing.B)
```

---

## Security Considerations

1. **Randomness Source**
   - MUST use `crypto/rand`, never `math/rand`
   - `crypto/rand` uses OS entropy source
   - Fails safely if entropy unavailable

2. **Token Entropy**
   - 256 bits = 2^256 possible values
   - Collision probability negligible
   - Unguessable even with billions of attempts

3. **Secret Key**
   - Never log secret key
   - Never expose in error messages
   - Store securely in memory
   - Consider key rotation strategy

4. **URL Safety**
   - Base64-URL encoding (RFC 4648)
   - No padding characters
   - Safe for URL paths without encoding

---

## Validation Rules

- Token length: exactly 43 characters
- Token format: `[A-Za-z0-9_-]{43}`
- Hash length: 43 characters (base64-URL encoded SHA256 without padding)
- Secret key: minimum 32 bytes recommended

---

## Error Handling

| Error Condition | Error Type | HTTP Status | User Message |
|----------------|------------|-------------|--------------|
| Random generation fails | `InternalError` | 500 | "Failed to generate secure token" |
| Invalid secret key | `ConfigError` | 500 | "Invalid token secret configuration" |

---

## References

- **HLD:** Section 6.2 (Token Security)
- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md) Section 4.1
- **Go Docs:** `crypto/rand`, `crypto/hmac`, `crypto/sha256`
- **RFC 4648:** Base64-URL encoding

---

## Definition of Done

- [x] All acceptance criteria met
- [x] Unit tests written and passing (87.5% coverage)
- [x] Performance benchmarks run
- [x] Security review passed
- [x] Documentation complete
- [x] Code reviewed
- [x] No linter warnings
- [ ] Integration with invite service verified (deferred to Story 03)
