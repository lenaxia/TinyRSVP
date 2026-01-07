# User Story: Token Validation

**Epic:** [03_EPIC_invites.md](03_EPIC_invites.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 0.5 days

---

## User Story

As a **system developer**, I want **constant-time token validation** so that **guest tokens are verified securely without timing attack vulnerabilities**.

---

## Acceptance Criteria

- [ ] Token validation uses constant-time comparison
- [ ] Validator interface allows for testing with mocks
- [ ] Validation compares computed hash with stored hash
- [ ] Invalid tokens return false (never panic or error)
- [ ] Validation is deterministic (same inputs = same result)
- [ ] Timing is constant regardless of token validity
- [ ] Validator can be initialized with secret key
- [ ] Validation handles malformed tokens gracefully

---

## Technical Details

### Package Location
- `pkg/token/validator.go`
- `pkg/token/validator_test.go`

### Interface Definition

```go
type Validator interface {
    Validate(token, hash string) bool
}
```

### Implementation Requirements

1. **Constant-Time Comparison**
   - Use `hmac.Equal()` for hash comparison
   - Never use `==` or `strings.Compare()`
   - Prevents timing attacks

2. **Hash Computation**
   - Compute HMAC-SHA256 of provided token
   - Use same secret key as generator
   - Base64-URL encode result

3. **Comparison**
   - Compare computed hash with stored hash
   - Use `hmac.Equal()` for constant-time comparison
   - Return boolean result

### Validation Flow

```
1. Receive token from URL
2. Compute HMAC-SHA256(secret, token)
3. Base64-URL encode computed hash
4. Constant-time compare with stored hash
5. Return true if match, false otherwise
```

---

## Subtasks

### Implementation
- [ ] Create `Validator` interface in `validator.go`
- [ ] Implement `validator` struct with secret key field
- [ ] Implement `NewValidator(secret []byte) Validator` constructor
- [ ] Implement `Validate(token, hash string) bool` method
- [ ] Use `hmac.Equal()` for constant-time comparison
- [ ] Handle edge cases (empty strings, malformed input)

### Testing
- [ ] Test valid token validation (should return true)
- [ ] Test invalid token validation (should return false)
- [ ] Test wrong secret key (should return false)
- [ ] Test malformed token (should return false)
- [ ] Test empty token (should return false)
- [ ] Test empty hash (should return false)
- [ ] Test timing consistency (constant-time verification)
- [ ] Benchmark validation performance

### Documentation
- [ ] Add package documentation
- [ ] Document security considerations
- [ ] Add usage examples
- [ ] Document timing attack prevention

---

## Dependencies

**Depends on:**
- Story 00: Token Generation (uses same HMAC algorithm)
- Story 01: Token Hashing (validates hashes)

**Blocks:**
- Story 03: Invite Model
- Story 04: Individual Invite

---

## Testing Strategy

### Unit Tests

1. **Valid Token Tests**
   ```go
   func TestValidator_Validate_ValidToken(t *testing.T) {
       secret := []byte("test-secret")
       gen := NewGenerator(secret)
       val := NewValidator(secret)
       
       token, _ := gen.Generate()
       hash, _ := gen.Hash(token)
       
       if !val.Validate(token, hash) {
           t.Error("Valid token should validate")
       }
   }
   ```

2. **Invalid Token Tests**
   ```go
   func TestValidator_Validate_InvalidToken(t *testing.T) {
       secret := []byte("test-secret")
       gen := NewGenerator(secret)
       val := NewValidator(secret)
       
       token, _ := gen.Generate()
       hash, _ := gen.Hash(token)
       
       if val.Validate("wrong-token", hash) {
           t.Error("Invalid token should not validate")
       }
   }
   ```

3. **Wrong Secret Tests**
   ```go
   func TestValidator_Validate_WrongSecret(t *testing.T) {
       gen := NewGenerator([]byte("secret1"))
       val := NewValidator([]byte("secret2"))
       
       token, _ := gen.Generate()
       hash, _ := gen.Hash(token)
       
       if val.Validate(token, hash) {
           t.Error("Wrong secret should not validate")
       }
   }
   ```

4. **Edge Case Tests**
   ```go
   func TestValidator_Validate_EdgeCases(t *testing.T) {
       val := NewValidator([]byte("secret"))
       
       tests := []struct {
           name  string
           token string
           hash  string
           want  bool
       }{
           {"empty token", "", "hash", false},
           {"empty hash", "token", "", false},
           {"both empty", "", "", false},
           {"malformed hash", "token", "not-base64!", false},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               got := val.Validate(tt.token, tt.hash)
               if got != tt.want {
                   t.Errorf("got %v, want %v", got, tt.want)
               }
           })
       }
   }
   ```

5. **Timing Attack Tests**
   ```go
   func TestValidator_Validate_ConstantTime(t *testing.T) {
       secret := []byte("test-secret")
       gen := NewGenerator(secret)
       val := NewValidator(secret)
       
       token, _ := gen.Generate()
       hash, _ := gen.Hash(token)
       
       // Measure time for valid token
       start := time.Now()
       for i := 0; i < 10000; i++ {
           val.Validate(token, hash)
       }
       validTime := time.Since(start)
       
       // Measure time for invalid token
       start = time.Now()
       for i := 0; i < 10000; i++ {
           val.Validate("wrong-token", hash)
       }
       invalidTime := time.Since(start)
       
       // Times should be similar (within 10%)
       ratio := float64(validTime) / float64(invalidTime)
       if ratio < 0.9 || ratio > 1.1 {
           t.Errorf("Timing difference too large: %v vs %v", validTime, invalidTime)
       }
   }
   ```

### Performance Tests

```go
func BenchmarkValidator_Validate(b *testing.B) {
    secret := []byte("test-secret")
    gen := NewGenerator(secret)
    val := NewValidator(secret)
    
    token, _ := gen.Generate()
    hash, _ := gen.Hash(token)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        val.Validate(token, hash)
    }
}
```

---

## Security Considerations

1. **Timing Attack Prevention**
   - MUST use `hmac.Equal()` for comparison
   - Never use `==`, `bytes.Equal()`, or `strings.Compare()`
   - These leak timing information about hash differences
   - Attackers can use timing to guess tokens

2. **Constant-Time Comparison**
   - `hmac.Equal()` compares all bytes regardless of match
   - Execution time independent of where mismatch occurs
   - Prevents timing-based token guessing

3. **Fail Secure**
   - Return false on any error or malformed input
   - Never panic or throw exceptions
   - Log validation failures for monitoring

4. **Secret Key Consistency**
   - Validator must use same secret as generator
   - Secret key rotation invalidates all tokens
   - Consider grace period for key rotation

---

## Validation Rules

- Token format: 43 characters, base64-URL encoded
- Hash format: 44 characters, base64-URL encoded
- Both token and hash must be non-empty
- Validation is case-sensitive
- No whitespace trimming (exact match required)

---

## Error Handling

Validation never returns errors, only boolean:
- Valid token + correct hash → `true`
- Invalid token → `false`
- Wrong secret → `false`
- Malformed input → `false`
- Empty input → `false`

Logging (for monitoring):
- Log validation failures with sanitized info
- Never log full tokens (only last 6 chars)
- Log hash mismatches for security monitoring

---

## Integration Points

### Invite Service
```go
func (s *service) GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
    // Find invite by computing hash
    hash, err := s.tokenGen.Hash(token)
    if err != nil {
        return nil, err
    }
    
    invite, err := s.repo.GetByTokenHash(ctx, hash)
    if err != nil {
        return nil, err
    }
    
    // Validate token (constant-time)
    if !s.tokenVal.Validate(token, invite.TokenHash) {
        return nil, models.ErrInvalidToken
    }
    
    return invite, nil
}
```

### HTTP Handler
```go
func (h *handler) RSVPPage(w http.ResponseWriter, r *http.Request) {
    token := chi.URLParam(r, "token")
    
    invite, err := h.inviteService.GetInviteByToken(r.Context(), token)
    if err != nil {
        http.Error(w, "Invalid or expired invite", http.StatusNotFound)
        return
    }
    
    // Render RSVP page
}
```

---

## References

- **HLD:** Section 6.2 (Token Security)
- **LLD:** [`lld/03_INVITE_LLD.md`](../lld/03_INVITE_LLD.md) Section 4.2
- **Story 00:** [03_STORY_00_token_generation.md](03_STORY_00_token_generation.md)
- **Story 01:** [03_STORY_01_token_hashing.md](03_STORY_01_token_hashing.md)
- **Go Docs:** `crypto/hmac`, `hmac.Equal()`
- **OWASP:** Timing Attack Prevention

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] Unit tests written and passing (>90% coverage)
- [ ] Timing attack tests pass
- [ ] Performance benchmarks run
- [ ] Security review passed
- [ ] Documentation complete
- [ ] Code reviewed
- [ ] No linter warnings
- [ ] Integration with invite service verified
