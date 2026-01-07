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

func TestIntegration_ValidatorInterface(t *testing.T) {
	var _ Validator = (*validator)(nil)

	secret := []byte("test-secret")
	val := NewValidator(secret)

	if val == nil {
		t.Fatal("NewValidator() returned nil")
	}

	result := val.Validate("test-token", "test-hash")
	if result {
		t.Error("Expected false for arbitrary token/hash pair")
	}
}

func TestIntegration_GeneratorAndValidatorWorkflow(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("Failed to generate secret: %v", err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if !val.Validate(token, hash) {
		t.Error("Validator should validate token generated and hashed by same secret")
	}

	wrongToken := "wrong-token-that-was-not-generated"
	if val.Validate(wrongToken, hash) {
		t.Error("Validator should reject wrong token")
	}

	wrongHash := "wrong-hash-that-does-not-match"
	if val.Validate(token, wrongHash) {
		t.Error("Validator should reject wrong hash")
	}

	if val.Validate("", hash) {
		t.Error("Validator should reject empty token")
	}

	if val.Validate(token, "") {
		t.Error("Validator should reject empty hash")
	}
}

func TestIntegration_GeneratorValidatorInterop(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("Failed to generate secret: %v", err)
	}

	gen1 := NewGenerator(secret)
	gen2 := NewGenerator(secret)
	val1 := NewValidator(secret)
	val2 := NewValidator(secret)

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

	if !val1.Validate(token, hash1) {
		t.Error("val1 should validate token hashed by gen1")
	}

	if !val1.Validate(token, hash2) {
		t.Error("val1 should validate token hashed by gen2")
	}

	if !val2.Validate(token, hash1) {
		t.Error("val2 should validate token hashed by gen1")
	}

	if !val2.Validate(token, hash2) {
		t.Error("val2 should validate token hashed by gen2")
	}
}

func TestIntegration_ConcurrentValidation(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatalf("Failed to generate secret: %v", err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	tokens := make([]string, 10)
	hashes := make([]string, 10)

	for i := 0; i < 10; i++ {
		token, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		hash, err := gen.Hash(token)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}
		tokens[i] = token
		hashes[i] = hash
	}

	concurrency := 50
	results := make(chan bool, concurrency)
	errors := make(chan string, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			tokenIdx := idx % 10
			if !val.Validate(tokens[tokenIdx], hashes[tokenIdx]) {
				errors <- "validation failed for valid token"
				return
			}

			wrongToken := "wrong-token-" + tokens[tokenIdx]
			if val.Validate(wrongToken, hashes[tokenIdx]) {
				errors <- "validation succeeded for invalid token"
				return
			}

			results <- true
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		select {
		case err := <-errors:
			t.Fatalf("Concurrent validation error: %s", err)
		case <-results:
		}
	}
}

func TestIntegration_ValidatorWithDifferentSecrets(t *testing.T) {
	secret1 := make([]byte, 32)
	if _, err := rand.Read(secret1); err != nil {
		t.Fatalf("Failed to generate secret1: %v", err)
	}

	secret2 := make([]byte, 32)
	if _, err := rand.Read(secret2); err != nil {
		t.Fatalf("Failed to generate secret2: %v", err)
	}

	gen1 := NewGenerator(secret1)
	gen2 := NewGenerator(secret2)
	val1 := NewValidator(secret1)
	val2 := NewValidator(secret2)

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

	if hash1 == hash2 {
		t.Error("Different secrets should produce different hashes")
	}

	if !val1.Validate(token, hash1) {
		t.Error("val1 should validate token hashed with secret1")
	}

	if val1.Validate(token, hash2) {
		t.Error("val1 should NOT validate token hashed with secret2")
	}

	if val2.Validate(token, hash1) {
		t.Error("val2 should NOT validate token hashed with secret1")
	}

	if !val2.Validate(token, hash2) {
		t.Error("val2 should validate token hashed with secret2")
	}
}
