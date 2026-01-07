package token

import (
	"crypto/rand"
	"testing"
)

func TestIntegration_FullTokenWorkflow(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("Failed to generate secret: %v", err)
	}

	gen := NewGenerator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(token) != 43 {
		t.Errorf("Generate() token length = %d, want 43", len(token))
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if len(hash) != 43 {
		t.Errorf("Hash() length = %d, want 43", len(hash))
	}

	hash2, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() second call error = %v", err)
	}

	if hash != hash2 {
		t.Error("Hash() should be deterministic for same token")
	}
}

func TestIntegration_MultipleGeneratorsWithSameSecret(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("Failed to generate secret: %v", err)
	}

	gen1 := NewGenerator(secret)
	gen2 := NewGenerator(secret)

	token, err := gen1.Generate()
	if err != nil {
		t.Fatalf("gen1.Generate() error = %v", err)
	}

	hash1, err := gen1.Hash(token)
	if err != nil {
		t.Fatalf("gen1.Hash() error = %v", err)
	}

	hash2, err := gen2.Hash(token)
	if err != nil {
		t.Fatalf("gen2.Hash() error = %v", err)
	}

	if hash1 != hash2 {
		t.Error("Different generators with same secret should produce same hash")
	}
}

func TestIntegration_TokenValidationWorkflow(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("Failed to generate secret: %v", err)
	}

	gen := NewGenerator(secret)

	validToken, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	validHash, err := gen.Hash(validToken)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	receivedToken := validToken
	receivedHash, err := gen.Hash(receivedToken)
	if err != nil {
		t.Fatalf("Hash() for validation error = %v", err)
	}

	if receivedHash != validHash {
		t.Error("Token validation failed: hashes should match")
	}

	invalidToken := "invalid-token-that-was-not-generated"
	invalidHash, err := gen.Hash(invalidToken)
	if err != nil {
		t.Fatalf("Hash() for invalid token error = %v", err)
	}

	if invalidHash == validHash {
		t.Error("Invalid token should not produce same hash as valid token")
	}
}

func TestIntegration_ConcurrentTokenGeneration(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("Failed to generate secret: %v", err)
	}

	gen := NewGenerator(secret)

	concurrency := 10
	tokens := make(chan string, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			token, err := gen.Generate()
			if err != nil {
				errors <- err
				return
			}
			tokens <- token
		}()
	}

	seen := make(map[string]bool)
	for i := 0; i < concurrency; i++ {
		select {
		case err := <-errors:
			t.Fatalf("Concurrent generation error: %v", err)
		case token := <-tokens:
			if seen[token] {
				t.Errorf("Duplicate token generated: %s", token)
			}
			seen[token] = true
		}
	}

	if len(seen) != concurrency {
		t.Errorf("Expected %d unique tokens, got %d", concurrency, len(seen))
	}
}

func TestIntegration_PackageImportability(t *testing.T) {
	var _ Generator = (*generator)(nil)

	secret := []byte("test-secret")
	gen := NewGenerator(secret)

	if gen == nil {
		t.Fatal("NewGenerator() returned nil")
	}

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if token == "" {
		t.Error("Generate() returned empty token")
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash == "" {
		t.Error("Hash() returned empty hash")
	}
}
