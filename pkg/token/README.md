# Token Package

## Purpose

The token package provides cryptographically secure token generation and hashing for guest invite tokens in TinyRSVP. This package is designed to generate unguessable, URL-safe tokens that can be safely distributed to guests via email and used in URLs.

## Security Considerations

### Cryptographic Randomness
- Uses `crypto/rand` (NOT `math/rand`) for token generation
- Draws from OS entropy source (e.g., `/dev/urandom` on Linux)
- Fails safely if entropy is unavailable
- Each token has 256 bits of entropy (2^256 possible values)
- Collision probability is negligible even with billions of tokens

### HMAC-SHA256 Hashing
- Tokens are hashed using HMAC-SHA256 before database storage
- HMAC prevents forgery - only holders of the secret key can generate valid hashes
- SHA256 provides 256-bit hash output
- Constant-time comparison prevents timing attacks
- Different secret keys produce different hashes for the same token

### Secret Key Management
- Secret key MUST be at least 32 bytes (256 bits) for security
- Secret key should be randomly generated using `crypto/rand`
- Secret key MUST be kept confidential and never logged
- Secret key MUST be stored securely (environment variable, secrets manager)
- Consider implementing key rotation strategy for long-lived deployments

### URL Safety
- Tokens use base64-URL encoding (RFC 4648)
- No padding characters (=) included
- Safe for use in URL paths without additional encoding
- Character set: `[A-Za-z0-9_-]`

## Interface

```go
type Generator interface {
    Generate() (string, error)
    Hash(token string) (string, error)
}
```

### Methods

#### `Generate() (string, error)`
Generates a cryptographically secure random token.

**Returns:**
- `string`: 43-character base64-URL encoded token
- `error`: Non-nil if random number generation fails

**Error Conditions:**
- Returns error if `crypto/rand.Read()` fails
- This typically indicates system entropy exhaustion
- Errors should be treated as fatal - do not retry
- Log error and fail the operation

#### `Hash(token string) (string, error)`
Generates HMAC-SHA256 hash of a token using the secret key.

**Parameters:**
- `token`: The token string to hash

**Returns:**
- `string`: 43-character base64-URL encoded hash
- `error`: Currently always nil, reserved for future use

**Properties:**
- Same token always produces same hash (deterministic)
- Different tokens produce different hashes
- Different secret keys produce different hashes for same token

## Usage Examples

### Basic Usage

```go
package main

import (
    "crypto/rand"
    "fmt"
    "log"
    
    "github.com/lenaxia/tinyrsvp/pkg/token"
)

func main() {
    secret := make([]byte, 32)
    if _, err := rand.Read(secret); err != nil {
        log.Fatal("Failed to generate secret:", err)
    }
    
    gen := token.NewGenerator(secret)
    
    inviteToken, err := gen.Generate()
    if err != nil {
        log.Fatal("Failed to generate token:", err)
    }
    
    tokenHash, err := gen.Hash(inviteToken)
    if err != nil {
        log.Fatal("Failed to hash token:", err)
    }
    
    fmt.Printf("Invite Token: %s\n", inviteToken)
    fmt.Printf("Token Hash: %s\n", tokenHash)
}
```

### Integration with Invite Service

```go
type InviteService struct {
    tokenGen token.Generator
    repo     InviteRepository
}

func (s *InviteService) CreateInvite(email string) (*Invite, error) {
    inviteToken, err := s.tokenGen.Generate()
    if err != nil {
        return nil, fmt.Errorf("failed to generate invite token: %w", err)
    }
    
    tokenHash, err := s.tokenGen.Hash(inviteToken)
    if err != nil {
        return nil, fmt.Errorf("failed to hash token: %w", err)
    }
    
    invite := &Invite{
        Email:     email,
        TokenHash: tokenHash,
        CreatedAt: time.Now(),
    }
    
    if err := s.repo.Create(invite); err != nil {
        return nil, err
    }
    
    invite.Token = inviteToken
    return invite, nil
}

func (s *InviteService) ValidateToken(token string) (*Invite, error) {
    tokenHash, err := s.tokenGen.Hash(token)
    if err != nil {
        return nil, fmt.Errorf("failed to hash token: %w", err)
    }
    
    invite, err := s.repo.FindByTokenHash(tokenHash)
    if err != nil {
        return nil, err
    }
    
    return invite, nil
}
```

### Testing with Mock Generator

```go
type MockGenerator struct {
    GenerateFunc func() (string, error)
    HashFunc     func(token string) (string, error)
}

func (m *MockGenerator) Generate() (string, error) {
    if m.GenerateFunc != nil {
        return m.GenerateFunc()
    }
    return "mock-token-43-characters-long-for-testing", nil
}

func (m *MockGenerator) Hash(token string) (string, error) {
    if m.HashFunc != nil {
        return m.HashFunc(token)
    }
    return "mock-hash-43-characters-long-for-testing!", nil
}
```

## Token Format

### Generated Token
```
Input:  32 random bytes from crypto/rand
Output: "a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5kL7mN9p" (43 chars)
Format: [A-Za-z0-9_-]{43}
```

### Token Hash
```
Input:  Token string + Secret key
Output: "xK2mP9qR5sT8vW1yZ3aC5dF7gH9jL2nO4pQ6rS8tU0v" (43 chars)
Format: [A-Za-z0-9_-]{43}
```

## Error Handling

### Generate() Errors

| Condition | Error | Action |
|-----------|-------|--------|
| Entropy unavailable | `failed to generate random bytes: ...` | Log error, fail operation, alert ops team |
| System under load | `failed to generate random bytes: ...` | Do NOT retry, fail secure |

### Hash() Errors

Currently, `Hash()` does not return errors but the signature includes `error` for future extensibility.

## Performance

Benchmarks on typical hardware:
- `Generate()`: ~50-100 µs per operation
- `Hash()`: ~5-10 µs per operation
- `GenerateAndHash()`: ~55-110 µs per operation

Token generation is suitable for real-time request handling with no caching required.

## Integration Guidance

### Initialization
1. Generate or load secret key at application startup
2. Create single `Generator` instance with secret key
3. Inject generator into services that need token operations
4. Do NOT create multiple generators with different secrets unless required

### Storage
- Store only token hashes in database, NEVER plain tokens
- Plain tokens should only exist in memory during generation
- Plain tokens should be sent to guests via email immediately
- Plain tokens should not be logged or persisted

### Validation
1. Receive token from guest (URL parameter or form field)
2. Hash the received token using `Hash()`
3. Look up invite by token hash in database
4. Verify invite is valid (not expired, not revoked)
5. Process RSVP if valid

### Key Rotation
If implementing key rotation:
1. Generate new secret key
2. Create new generator with new key
3. Use new generator for new tokens
4. Keep old generator(s) for validating existing tokens
5. Gradually phase out old keys as tokens expire

## Dependencies

- `crypto/rand` - Cryptographic random number generation
- `crypto/hmac` - HMAC message authentication
- `crypto/sha256` - SHA256 hashing
- `encoding/base64` - Base64-URL encoding

## Testing

Run tests with timeout:
```bash
go test -timeout 30s ./pkg/token/...
```

Run tests with coverage:
```bash
go test -timeout 30s -cover ./pkg/token/...
```

Run benchmarks:
```bash
go test -timeout 30s -bench=. ./pkg/token/...
```

## References

- [RFC 4648](https://tools.ietf.org/html/rfc4648) - Base64-URL encoding
- [FIPS 198-1](https://csrc.nist.gov/publications/detail/fips/198/1/final) - HMAC specification
- [FIPS 180-4](https://csrc.nist.gov/publications/detail/fips/180/4/final) - SHA-256 specification
- Go `crypto/rand` documentation
- Story 00: Token Generation ([`docs/00_BACKLOG/03_STORY_00_token_generation.md`](../../docs/00_BACKLOG/03_STORY_00_token_generation.md))
