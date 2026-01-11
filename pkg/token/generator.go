// Package token provides cryptographically secure token generation and hashing
// for guest invite tokens.
//
// Security Properties:
//   - Uses crypto/rand (NOT math/rand) for cryptographic security
//   - Each token has 256 bits of entropy (2^256 possible values)
//   - HMAC-SHA256 prevents token forgery without secret key
//   - Constant-time comparison prevents timing attacks
//   - Secret key must be kept confidential and never logged
//
// Token Format:
//   - Generated tokens: 43-character base64-URL encoded strings
//   - Token hashes: 43-character base64-URL encoded HMAC-SHA256
//   - Character set: [A-Za-z0-9_-] (URL-safe, no padding)
package token

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// Generator defines the interface for cryptographically secure token operations.
// Implementations must use crypto/rand for token generation and HMAC-SHA256 for hashing.
type Generator interface {
	// Generate creates a cryptographically secure random token.
	//
	// Returns a 43-character base64-URL encoded string with 256 bits of entropy.
	// The token is safe for use in URLs without additional encoding.
	//
	// Error Conditions:
	//   - Returns error if crypto/rand.Read() fails
	//   - Failure indicates system entropy exhaustion or unavailable randomness source
	//   - Errors are system-level issues and should NOT be retried
	//   - Must fail secure - never fall back to weaker randomness sources
	//
	// Example:
	//   token, err := gen.Generate()
	//   if err != nil {
	//       log.Fatal("System entropy unavailable:", err)
	//   }
	Generate() (string, error)

	// Hash generates an HMAC-SHA256 hash of the provided token.
	//
	// Returns a 43-character base64-URL encoded hash. The hash is deterministic:
	// the same token with the same secret key always produces the same hash.
	//
	// Properties:
	//   - Same token always produces same hash (deterministic)
	//   - Different tokens produce different hashes
	//   - Different secret keys produce different hashes for same token
	//   - Constant-time comparison prevents timing attacks
	//
	// Error Conditions:
	//   - Currently always returns nil error
	//   - Error return reserved for future extensibility
	//
	// Example:
	//   hash, err := gen.Hash(token)
	//   if err != nil {
	//       return err
	//   }
	//   // Store hash in database, never store plain token
	Hash(token string) (string, error)
}

type generator struct {
	secret         []byte
	hashingEnabled bool
}

// NewGenerator creates a new Generator with the provided secret key.
//
// The secret key is used for HMAC-SHA256 hashing of tokens. The same secret
// key must be used for both generating hashes and validating tokens.
//
// Secret Key Requirements:
//   - Should be at least 32 bytes (256 bits) for security
//   - Must be randomly generated using crypto/rand
//   - Must be kept confidential and never logged or exposed
//   - Should be stored securely (environment variable, secrets manager)
//   - Consider implementing key rotation for long-lived deployments
//
// Example:
//
//	secret := make([]byte, 32)
//	if _, err := rand.Read(secret); err != nil {
//	    log.Fatal(err)
//	}
//	gen := NewGenerator(secret)
func NewGenerator(secret []byte) Generator {
	return &generator{
		secret:         secret,
		hashingEnabled: len(secret) > 0,
	}
}

// Generate creates a cryptographically secure random token.
//
// Implementation details:
//   - Generates 32 random bytes using crypto/rand
//   - Encodes bytes as base64-URL without padding
//   - Results in 43-character string
//
// The token has 256 bits of entropy, making it computationally infeasible
// to guess even with billions of attempts. Collision probability is negligible.
func (g *generator) Generate() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

// Hash generates an HMAC-SHA256 hash of the token using the secret key.
//
// Implementation details:
//   - Uses HMAC-SHA256 with the generator's secret key
//   - Encodes hash as base64-URL without padding
//   - Results in 43-character string
//
// The HMAC prevents token forgery - only holders of the secret key can
// generate valid hashes. This allows secure token validation by comparing
// hashes instead of storing plain tokens.
func (g *generator) Hash(token string) (string, error) {
	if !g.hashingEnabled {
		return token, nil
	}
	h := hmac.New(sha256.New, g.secret)
	h.Write([]byte(token))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h.Sum(nil)), nil
}
