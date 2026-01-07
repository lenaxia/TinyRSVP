package token

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Generator interface {
	Generate() (string, error)
	Hash(token string) (string, error)
}

type generator struct {
	secret []byte
}

func NewGenerator(secret []byte) Generator {
	return &generator{secret: secret}
}

func (g *generator) Generate() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

func (g *generator) Hash(token string) (string, error) {
	h := hmac.New(sha256.New, g.secret)
	h.Write([]byte(token))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h.Sum(nil)), nil
}
