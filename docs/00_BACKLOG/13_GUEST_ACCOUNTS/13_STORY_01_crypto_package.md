# User Story: Crypto Package — Encryptor Interface, AES-256-GCM, HKDF, HMAC Blind Hash

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 4 hours  

---

## User Story

As a **developer**, I want **a reusable `pkg/crypto/` package that provides AES-256-GCM encryption and HMAC-SHA256 deterministic hashing** so that **all repositories can encrypt PII columns and perform searchable lookups without storing plaintext**.

---

## Acceptance Criteria

- [ ] `Encryptor` interface defined with `Encrypt`, `Decrypt`, and `Hash` methods
- [ ] `NewEncryptor(masterKey []byte)` derives two keys via HKDF-SHA256: one for AES-GCM, one for HMAC
- [ ] `Encrypt` produces unique ciphertext on each call (random 12-byte nonce) stored as `base64(nonce || ciphertext)`
- [ ] `Decrypt` reverses `Encrypt` correctly
- [ ] `Hash` is deterministic: same input always produces same output
- [ ] `Hash` and `Encrypt` use different derived keys (HKDF with distinct info labels)
- [ ] Decrypting with wrong key returns an error
- [ ] `NewEncryptorFromEnv()` reads `TINYRSVP_ENCRYPTION_KEY` (base64-encoded 32 bytes) and returns error if absent or malformed
- [ ] All tests pass with timeout

---

## Technical Details

### Package Location

```
pkg/crypto/
  crypto.go
  crypto_test.go
```

### Interface

```go
package crypto

type Encryptor interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
    Hash(value string) string
}
```

### Key Derivation

```go
import "golang.org/x/crypto/hkdf"

func deriveKey(masterKey []byte, info string) ([]byte, error) {
    h := hkdf.New(sha256.New, masterKey, nil, []byte(info))
    key := make([]byte, 32)
    _, err := io.ReadFull(h, key)
    return key, err
}

// Info labels (distinct, never reused):
const encKeyInfo = "tinyrsvp-enc-v1"
const idxKeyInfo = "tinyrsvp-idx-v1"
```

### Stored Format

```
Encrypt output: base64url( nonce[12 bytes] || ciphertext )
Hash output:    base64url( HMAC-SHA256(value, idx_key) )
```

### Constructor

```go
func NewEncryptor(masterKey []byte) (Encryptor, error)
func NewEncryptorFromEnv() (Encryptor, error) // reads TINYRSVP_ENCRYPTION_KEY
```

`NewEncryptorFromEnv` error message must include the generation hint:
`"TINYRSVP_ENCRYPTION_KEY is required; generate with: openssl rand -base64 32"`

---

## Tasks

### Phase 1: Interface and Key Derivation (TDD)
- [ ] Write test: `TestNewEncryptor_ValidKey` — constructs successfully
- [ ] Write test: `TestNewEncryptor_ShortKey` — returns error for key < 32 bytes
- [ ] Write test: `TestNewEncryptorFromEnv_Missing` — returns descriptive error
- [ ] Write test: `TestNewEncryptorFromEnv_Malformed` — invalid base64 returns error
- [ ] Write test: `TestNewEncryptorFromEnv_Valid` — sets env var, constructs successfully
- [ ] Run tests (should fail — package does not exist)
- [ ] Implement `pkg/crypto/crypto.go` with key derivation
- [ ] Run tests (should pass)

### Phase 2: Encrypt / Decrypt (TDD)
- [ ] Write test: `TestEncryptor_Roundtrip` — encrypt then decrypt returns original
- [ ] Write test: `TestEncryptor_UniqueNonce` — same plaintext encrypted twice produces different ciphertexts
- [ ] Write test: `TestEncryptor_EmptyString` — encrypts and decrypts empty string
- [ ] Write test: `TestEncryptor_WrongKey` — decrypt with different key returns error
- [ ] Write test: `TestEncryptor_TamperedCiphertext` — modified ciphertext returns error (GCM auth tag)
- [ ] Write test: `TestEncryptor_InvalidBase64` — garbage input to Decrypt returns error
- [ ] Run tests (should fail)
- [ ] Implement `Encrypt` and `Decrypt`
- [ ] Run tests (should pass)

### Phase 3: Hash (TDD)
- [ ] Write test: `TestEncryptor_HashDeterministic` — same value produces same hash across multiple calls
- [ ] Write test: `TestEncryptor_HashDistinct` — different values produce different hashes
- [ ] Write test: `TestEncryptor_HashDifferentFromEncrypt` — `Hash(x) != base64(Encrypt(x))` (different keys)
- [ ] Write test: `TestEncryptor_HashEmpty` — empty string produces consistent hash
- [ ] Run tests (should fail)
- [ ] Implement `Hash`
- [ ] Run tests (should pass)

---

## Testing Requirements

```go
func TestEncryptor_Roundtrip(t *testing.T) {
    tests := []struct {
        name      string
        plaintext string
    }{
        {"normal email", "alice@example.com"},
        {"empty string", ""},
        {"unicode", "用户@例子.com"},
        {"long value", strings.Repeat("a", 1000)},
        {"phone number", "+15551234567"},
    }

    enc, err := NewEncryptor(make([]byte, 32))
    if err != nil {
        t.Fatalf("NewEncryptor: %v", err)
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ct, err := enc.Encrypt(tt.plaintext)
            if err != nil {
                t.Fatalf("Encrypt: %v", err)
            }
            got, err := enc.Decrypt(ct)
            if err != nil {
                t.Fatalf("Decrypt: %v", err)
            }
            if got != tt.plaintext {
                t.Errorf("got %q, want %q", got, tt.plaintext)
            }
        })
    }
}

func TestEncryptor_WrongKey(t *testing.T) {
    key1 := make([]byte, 32)
    key2 := make([]byte, 32)
    key2[0] = 0xFF

    enc1, _ := NewEncryptor(key1)
    enc2, _ := NewEncryptor(key2)

    ct, err := enc1.Encrypt("secret")
    if err != nil {
        t.Fatalf("Encrypt: %v", err)
    }

    _, err = enc2.Decrypt(ct)
    if err == nil {
        t.Error("expected error decrypting with wrong key, got nil")
    }
}
```

---

## Dependencies

**Depends on:** Nothing (no internal dependencies; `golang.org/x/crypto` already in `go.sum` via `go-oidc`)  
**Blocks:** Stories 02, 03, 04, 05 (all PII encryption stories), Story 06 (guest account models)

**New dependency:** `golang.org/x/crypto/hkdf` — already transitively present; add explicit `require` if needed.

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass: `go test -timeout 30s -race ./pkg/crypto/...`
- [ ] No `map[string]interface{}` in implementation
- [ ] `go vet ./pkg/crypto/...` clean
- [ ] `go fmt ./pkg/crypto/...` applied
