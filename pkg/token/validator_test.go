package token

import (
	"crypto/rand"
	"testing"
	"time"
)

func TestNewValidator(t *testing.T) {
	secret := []byte("test-secret-key")
	val := NewValidator(secret)

	if val == nil {
		t.Fatal("NewValidator returned nil")
	}
}

func TestValidator_Validate_ValidToken(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatal(err)
	}

	if !val.Validate(token, hash) {
		t.Error("Valid token should validate successfully")
	}
}

func TestValidator_Validate_InvalidToken(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatal(err)
	}

	if val.Validate("wrong-token", hash) {
		t.Error("Invalid token should not validate")
	}
}

func TestValidator_Validate_WrongSecret(t *testing.T) {
	secret1 := make([]byte, 32)
	secret2 := make([]byte, 32)
	if _, err := rand.Read(secret1); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(secret2); err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(secret1)
	val := NewValidator(secret2)

	token, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatal(err)
	}

	if val.Validate(token, hash) {
		t.Error("Token with wrong secret should not validate")
	}
}

func TestValidator_Validate_EdgeCases(t *testing.T) {
	secret := []byte("test-secret")
	val := NewValidator(secret)

	tests := []struct {
		name  string
		token string
		hash  string
		want  bool
	}{
		{
			name:  "empty token",
			token: "",
			hash:  "some-hash",
			want:  false,
		},
		{
			name:  "empty hash",
			token: "some-token",
			hash:  "",
			want:  false,
		},
		{
			name:  "both empty",
			token: "",
			hash:  "",
			want:  false,
		},
		{
			name:  "malformed hash",
			token: "some-token",
			hash:  "not-base64!@#$%",
			want:  false,
		},
		{
			name:  "whitespace token",
			token: "   ",
			hash:  "some-hash",
			want:  false,
		},
		{
			name:  "very long token",
			token: string(make([]byte, 10000)),
			hash:  "some-hash",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := val.Validate(tt.token, tt.hash)
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_Validate_MultipleTokens(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	tokens := make([]string, 10)
	hashes := make([]string, 10)

	for i := 0; i < 10; i++ {
		token, err := gen.Generate()
		if err != nil {
			t.Fatal(err)
		}
		hash, err := gen.Hash(token)
		if err != nil {
			t.Fatal(err)
		}
		tokens[i] = token
		hashes[i] = hash
	}

	for i := 0; i < 10; i++ {
		if !val.Validate(tokens[i], hashes[i]) {
			t.Errorf("Token %d should validate", i)
		}
	}

	for i := 0; i < 10; i++ {
		wrongIdx := (i + 1) % 10
		if val.Validate(tokens[i], hashes[wrongIdx]) {
			t.Errorf("Token %d with wrong hash should not validate", i)
		}
	}
}

func TestValidator_Validate_Deterministic(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		result := val.Validate(token, hash)
		if !result {
			t.Errorf("Validation should be deterministic, failed on iteration %d", i)
		}
	}
}

func TestValidator_Validate_ConstantTime(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}

	hash, err := gen.Hash(token)
	if err != nil {
		t.Fatal(err)
	}

	iterations := 50000

	var validTimes []time.Duration
	var invalidTimes []time.Duration

	for run := 0; run < 5; run++ {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			val.Validate(token, hash)
		}
		validTimes = append(validTimes, time.Since(start))

		start = time.Now()
		for i := 0; i < iterations; i++ {
			val.Validate("wrong-token-with-same-length-as-real", hash)
		}
		invalidTimes = append(invalidTimes, time.Since(start))
	}

	var avgValid, avgInvalid time.Duration
	for i := 0; i < 5; i++ {
		avgValid += validTimes[i]
		avgInvalid += invalidTimes[i]
	}
	avgValid /= 5
	avgInvalid /= 5

	ratio := float64(avgValid) / float64(avgInvalid)
	if ratio < 0.5 || ratio > 2.0 {
		t.Logf("Avg valid time: %v, Avg invalid time: %v, Ratio: %.2f", avgValid, avgInvalid, ratio)
		t.Errorf("Timing difference too large (ratio %.2f), possible timing attack vulnerability", ratio)
	}
}

func TestValidator_Validate_DifferentHashLengths(t *testing.T) {
	secret := []byte("test-secret")
	val := NewValidator(secret)
	gen := NewGenerator(secret)

	token, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		hash string
		want bool
	}{
		{
			name: "short hash",
			hash: "abc",
			want: false,
		},
		{
			name: "long hash",
			hash: string(make([]byte, 1000)),
			want: false,
		},
		{
			name: "single character",
			hash: "a",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := val.Validate(token, tt.hash)
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_Validate_SpecialCharacters(t *testing.T) {
	secret := []byte("test-secret")
	val := NewValidator(secret)

	tests := []struct {
		name  string
		token string
		hash  string
		want  bool
	}{
		{
			name:  "null bytes in token",
			token: "token\x00with\x00nulls",
			hash:  "some-hash",
			want:  false,
		},
		{
			name:  "unicode in token",
			token: "token-with-unicode-🎉",
			hash:  "some-hash",
			want:  false,
		},
		{
			name:  "newlines in token",
			token: "token\nwith\nnewlines",
			hash:  "some-hash",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := val.Validate(tt.token, tt.hash)
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkValidator_Validate(b *testing.B) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		b.Fatal(err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	token, err := gen.Generate()
	if err != nil {
		b.Fatal(err)
	}

	hash, err := gen.Hash(token)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Validate(token, hash)
	}
}

func BenchmarkValidator_Validate_Invalid(b *testing.B) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		b.Fatal(err)
	}

	gen := NewGenerator(secret)
	val := NewValidator(secret)

	hash, err := gen.Hash("some-token")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Validate("wrong-token", hash)
	}
}
