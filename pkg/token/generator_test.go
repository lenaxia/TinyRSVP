package token

import (
	"encoding/base64"
	"regexp"
	"strings"
	"testing"
)

func TestGenerator_Generate_Length(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(token) != 43 {
		t.Errorf("Generate() token length = %d, want 43", len(token))
	}
}

func TestGenerator_Generate_URLSafe(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	urlSafePattern := regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
	if !urlSafePattern.MatchString(token) {
		t.Errorf("Generate() token = %q is not URL-safe base64", token)
	}

	if strings.Contains(token, "+") || strings.Contains(token, "/") || strings.Contains(token, "=") {
		t.Errorf("Generate() token = %q contains non-URL-safe characters", token)
	}
}

func TestGenerator_Generate_Uniqueness(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	tokens := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		token, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() iteration %d error = %v", i, err)
		}

		if tokens[token] {
			t.Errorf("Generate() produced duplicate token: %q", token)
		}
		tokens[token] = true
	}

	if len(tokens) != iterations {
		t.Errorf("Generate() produced %d unique tokens, want %d", len(tokens), iterations)
	}
}

func TestGenerator_Generate_MultipleInstances(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen1 := NewGenerator(secret)
	gen2 := NewGenerator(secret)

	token1, err := gen1.Generate()
	if err != nil {
		t.Fatalf("gen1.Generate() error = %v", err)
	}

	token2, err := gen2.Generate()
	if err != nil {
		t.Fatalf("gen2.Generate() error = %v", err)
	}

	if token1 == token2 {
		t.Error("Different generator instances should produce different tokens")
	}
}

func TestGenerator_Hash_Consistency(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	token := "test-token-value"

	hash1, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() first call error = %v", err)
	}

	hash2, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() second call error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Hash() inconsistent: first = %q, second = %q", hash1, hash2)
	}
}

func TestGenerator_Hash_Uniqueness(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	token1 := "test-token-1"
	token2 := "test-token-2"

	hash1, err := gen.Hash(token1)
	if err != nil {
		t.Fatalf("Hash(token1) error = %v", err)
	}

	hash2, err := gen.Hash(token2)
	if err != nil {
		t.Fatalf("Hash(token2) error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("Hash() should produce different hashes for different tokens")
	}
}

func TestGenerator_Hash_DifferentSecrets(t *testing.T) {
	secret1 := []byte("secret-key-1-32-bytes-long!!!!")
	secret2 := []byte("secret-key-2-32-bytes-long!!!!")

	gen1 := NewGenerator(secret1)
	gen2 := NewGenerator(secret2)

	token := "test-token-value"

	hash1, err := gen1.Hash(token)
	if err != nil {
		t.Fatalf("gen1.Hash() error = %v", err)
	}

	hash2, err := gen2.Hash(token)
	if err != nil {
		t.Fatalf("gen2.Hash() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("Hash() with different secrets should produce different hashes")
	}
}

func TestGenerator_Hash_Length(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	token := "test-token-value"

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if len(hash) != 43 {
		t.Errorf("Hash() length = %d, want 43 (base64-URL encoded SHA256 without padding)", len(hash))
	}
}

func TestGenerator_Hash_URLSafe(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	token := "test-token-value"

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	_, err = base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(hash)
	if err != nil {
		t.Errorf("Hash() = %q is not valid base64-URL encoding: %v", hash, err)
	}

	if strings.Contains(hash, "+") || strings.Contains(hash, "/") || strings.Contains(hash, "=") {
		t.Errorf("Hash() = %q contains non-URL-safe characters", hash)
	}
}

func TestGenerator_Hash_EmptyToken(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	hash, err := gen.Hash("")
	if err != nil {
		t.Fatalf("Hash(\"\") error = %v", err)
	}

	if hash == "" {
		t.Error("Hash(\"\") should return a hash, not empty string")
	}
}

func TestNewGenerator_NilSecret(t *testing.T) {
	gen := NewGenerator(nil)

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() with nil secret error = %v", err)
	}

	if token == "" {
		t.Error("Generate() with nil secret should still generate token")
	}
}

func TestNewGenerator_EmptySecret(t *testing.T) {
	gen := NewGenerator([]byte{})

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() with empty secret error = %v", err)
	}

	if token == "" {
		t.Error("Generate() with empty secret should still generate token")
	}
}

func TestNewGenerator_ShortSecret(t *testing.T) {
	secret := []byte("short")
	gen := NewGenerator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() with short secret error = %v", err)
	}

	if token == "" {
		t.Error("Generate() with short secret should still generate token")
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() with short secret error = %v", err)
	}

	if hash == "" {
		t.Error("Hash() with short secret should still generate hash")
	}
}

func TestNewGenerator_LongSecret(t *testing.T) {
	secret := []byte(strings.Repeat("a", 256))
	gen := NewGenerator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() with long secret error = %v", err)
	}

	if token == "" {
		t.Error("Generate() with long secret should still generate token")
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() with long secret error = %v", err)
	}

	if hash == "" {
		t.Error("Hash() with long secret should still generate hash")
	}
}

func TestGenerator_IntegrationFlow(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if len(token) != 43 {
		t.Errorf("token length = %d, want 43", len(token))
	}

	if len(hash) != 43 {
		t.Errorf("hash length = %d, want 43", len(hash))
	}

	hash2, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() second call error = %v", err)
	}

	if hash != hash2 {
		t.Error("Hash() should be consistent for same token")
	}
}

func BenchmarkGenerator_Generate(b *testing.B) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_Hash(b *testing.B) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)
	token := "test-token-value-for-benchmarking"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Hash(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_GenerateAndHash(b *testing.B) {
	secret := []byte("test-secret-key-32-bytes-long!!")
	gen := NewGenerator(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token, err := gen.Generate()
		if err != nil {
			b.Fatal(err)
		}
		_, err = gen.Hash(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}
