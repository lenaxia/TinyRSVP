package token

import (
	"testing"
)

func TestGenerator_PlainTokenMode(t *testing.T) {
	gen := NewGenerator(nil)
	
	token, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}
	
	if token == "" {
		t.Fatal("Generate() returned empty token")
	}
	
	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v, want nil", err)
	}
	
	if hash != token {
		t.Errorf("Hash() in plain mode should return token unchanged, got %s, want %s", hash, token)
	}
}

func TestGenerator_PlainTokenMode_EmptySecret(t *testing.T) {
	gen := NewGenerator([]byte{})
	
	token := "test_token_123"
	
	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v, want nil", err)
	}
	
	if hash != token {
		t.Errorf("Hash() with empty secret should return token unchanged, got %s, want %s", hash, token)
	}
}

func TestGenerator_HMACMode_WithSecret(t *testing.T) {
	secret := []byte("my-secret-key-exactly-32-bytes!!")
	gen := NewGenerator(secret)
	
	token := "test_token_123"
	
	hash1, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v, want nil", err)
	}
	
	if hash1 == token {
		t.Error("Hash() with secret should not return plain token")
	}
	
	hash2, err := gen.Hash(token)
	if err != nil {
		t.Fatalf("Hash() error = %v, want nil", err)
	}
	
	if hash1 != hash2 {
		t.Error("Hash() should be deterministic - same token should produce same hash")
	}
}

func TestGenerator_HMACMode_DifferentSecrets(t *testing.T) {
	secret1 := []byte("secret-one-exactly-32-bytes!!!!!")
	secret2 := []byte("secret-two-exactly-32-bytes!!!!!")
	
	gen1 := NewGenerator(secret1)
	gen2 := NewGenerator(secret2)
	
	token := "test_token_123"
	
	hash1, err := gen1.Hash(token)
	if err != nil {
		t.Fatalf("gen1.Hash() error = %v, want nil", err)
	}
	
	hash2, err := gen2.Hash(token)
	if err != nil {
		t.Fatalf("gen2.Hash() error = %v, want nil", err)
	}
	
	if hash1 == hash2 {
		t.Error("Different secrets should produce different hashes for same token")
	}
}

func TestGenerator_ModeDetection(t *testing.T) {
	tests := []struct {
		name           string
		secret         []byte
		expectPlainMode bool
	}{
		{
			name:           "nil secret enables plain mode",
			secret:         nil,
			expectPlainMode: true,
		},
		{
			name:           "empty secret enables plain mode",
			secret:         []byte{},
			expectPlainMode: true,
		},
		{
			name:           "non-empty secret enables HMAC mode",
			secret:         []byte("secret"),
			expectPlainMode: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tt.secret)
			
			token := "test_token"
			hash, err := gen.Hash(token)
			if err != nil {
				t.Fatalf("Hash() error = %v, want nil", err)
			}
			
			isPlainMode := (hash == token)
			if isPlainMode != tt.expectPlainMode {
				t.Errorf("Plain mode = %v, want %v", isPlainMode, tt.expectPlainMode)
			}
		})
	}
}
