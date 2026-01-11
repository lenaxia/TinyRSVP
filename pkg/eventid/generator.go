package eventid

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const (
	// IDLength is the length of generated event IDs
	IDLength = 10

	// base62 character set (0-9, A-Z, a-z)
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	ErrInvalidIDLength    = errors.New("invalid event ID length")
	ErrInvalidIDCharacter = errors.New("event ID contains invalid characters")
	ErrEmptyID            = errors.New("event ID cannot be empty")
)

// GenerateEventID generates a cryptographically random 10-character event ID
// using base62 encoding (0-9, A-Z, a-z). This provides approximately 60 bits
// of entropy, making IDs extremely difficult to guess while remaining shorter
// than UUIDs.
//
// Example output: "aBcD123456"
func GenerateEventID() (string, error) {
	id := make([]byte, IDLength)
	maxIdx := big.NewInt(int64(len(base62Chars)))

	for i := 0; i < IDLength; i++ {
		num, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		id[i] = base62Chars[num.Int64()]
	}

	return string(id), nil
}

// ValidateEventID validates that an event ID meets the required format:
// - Exactly 10 characters long
// - Contains only base62 characters (0-9, A-Z, a-z)
func ValidateEventID(id string) error {
	if id == "" {
		return ErrEmptyID
	}

	if len(id) != IDLength {
		return fmt.Errorf("%w: expected %d characters, got %d", ErrInvalidIDLength, IDLength, len(id))
	}

	for _, c := range id {
		if !isValidBase62Char(c) {
			return fmt.Errorf("%w: '%c'", ErrInvalidIDCharacter, c)
		}
	}

	return nil
}

// isValidBase62Char checks if a character is valid in base62 encoding
func isValidBase62Char(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z')
}
