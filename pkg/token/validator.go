package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Validator defines the interface for constant-time token validation.
// Implementations must use constant-time comparison to prevent timing attacks.
type Validator interface {
	// Validate performs constant-time validation of a token against its hash.
	//
	// This method computes the HMAC-SHA256 hash of the provided token using
	// the validator's secret key, then performs a constant-time comparison
	// with the provided hash.
	//
	// Security Properties:
	//   - Uses hmac.Equal() for constant-time comparison
	//   - Execution time is independent of where hash mismatch occurs
	//   - Prevents timing-based token guessing attacks
	//   - Never panics or returns errors, always returns boolean
	//
	// Parameters:
	//   - token: The token string to validate (43 characters, base64-URL)
	//   - hash: The expected hash to compare against (43 characters, base64-URL)
	//
	// Returns:
	//   - true if token is valid and hash matches
	//   - false for any error condition (invalid token, wrong secret, malformed input)
	//
	// Edge Cases:
	//   - Empty token or hash returns false
	//   - Malformed input returns false
	//   - Wrong secret key returns false
	//   - Never panics, always returns boolean
	//
	// Example:
	//   if validator.Validate(token, storedHash) {
	//       // Token is valid, proceed with RSVP
	//   } else {
	//       // Token is invalid, return 404
	//   }
	Validate(token, hash string) bool
}

type validator struct {
	secret []byte
}

// NewValidator creates a new Validator with the provided secret key.
//
// The secret key must be the same key used by the Generator that created
// the token hashes. Using a different secret key will cause all validations
// to fail.
//
// Secret Key Requirements:
//   - Should be at least 32 bytes (256 bits) for security
//   - Must be the same key used for token generation
//   - Must be kept confidential and never logged
//   - Should be stored securely (environment variable, secrets manager)
//
// Example:
//
//	secret := []byte(os.Getenv("TOKEN_SECRET"))
//	validator := NewValidator(secret)
func NewValidator(secret []byte) Validator {
	return &validator{secret: secret}
}

// Validate performs constant-time validation of a token against its hash.
//
// Implementation Details:
//  1. Computes HMAC-SHA256 of the provided token
//  2. Base64-URL encodes the computed hash
//  3. Uses hmac.Equal() for constant-time comparison with provided hash
//  4. Returns true only if hashes match exactly
//
// The constant-time comparison is critical for security. Using standard
// comparison operators (==, bytes.Equal, strings.Compare) would leak timing
// information that attackers could use to guess valid tokens.
//
// Timing Attack Prevention:
//   - hmac.Equal() compares all bytes regardless of match
//   - Execution time is independent of where mismatch occurs
//   - Prevents attackers from using timing to guess tokens
//
// Error Handling:
//   - Returns false for any error condition
//   - Never panics or throws exceptions
//   - Fail secure - invalid input is treated as invalid token
func (v *validator) Validate(token, hash string) bool {
	if token == "" || hash == "" {
		return false
	}

	h := hmac.New(sha256.New, v.secret)
	h.Write([]byte(token))
	computedHash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(computedHash), []byte(hash))
}
