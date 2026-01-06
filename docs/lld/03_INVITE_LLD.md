# Domain 3: Invite & Token Management - Low-Level Design

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Implementation Ready  
**HLD Reference:** [Section 6 - Invite & Guest Access Model](../02_REVISED_HLD.md#6-invite--guest-access-model)

---

## 1. Overview

### 1.1 Purpose

Manages invite creation, token generation/validation, and invite lifecycle including bulk operations, token security, and guest access control.

### 1.2 Responsibilities

- Invite creation (individual, bulk CSV, manual)
- Token generation (256-bit cryptographically secure)
- Token hashing (HMAC-SHA256)
- Token validation (constant-time comparison)
- Token lifecycle (expiration, revocation, regeneration)
- Invite status tracking
- CSV import with validation
- Duplicate detection

### 1.3 Design Principles

- **Cryptographically Secure** - crypto/rand for token generation
- **HMAC-Based** - HMAC-SHA256 for token hashing
- **Constant-Time** - Prevent timing attacks
- **Fail Secure** - Deny access on validation error
- **Audit Everything** - Log all token operations

---

## 2. Package Structure

```
internal/
├── invites/
│   ├── service.go              # Invite service
│   ├── service_test.go
│   ├── csv.go                  # CSV import
│   ├── csv_test.go
│   └── validator.go            # Invite validation
│       └── validator_test.go
pkg/
└── token/
    ├── generator.go            # Token generation
    ├── generator_test.go
    ├── validator.go            # Token validation
    └── validator_test.go
```

---

## 3. Interfaces

### 3.1 Invite Service Interface

```go
package invites

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type Service interface {
    CreateInvite(ctx context.Context, invite *models.Invite) (token string, err error)
    CreateInviteBatch(ctx context.Context, invites []*models.Invite) ([]InviteResult, error)
    ImportCSV(ctx context.Context, eventID int64, csvData []byte) (*ImportResult, error)
    GetInvite(ctx context.Context, id int64) (*models.Invite, error)
    GetInviteByToken(ctx context.Context, token string) (*models.Invite, error)
    UpdateInvite(ctx context.Context, invite *models.Invite) error
    RevokeInvite(ctx context.Context, id int64) error
    RegenerateToken(ctx context.Context, id int64) (token string, err error)
    ListInvites(ctx context.Context, eventID int64) ([]*models.Invite, error)
    GetInviteStats(ctx context.Context, eventID int64) (*InviteStats, error)
}

type InviteResult struct {
    Invite *models.Invite
    Token  string
    Error  error
}

type ImportResult struct {
    Total      int
    Created    int
    Failed     int
    Duplicates int
    Errors     []ImportError
}

type ImportError struct {
    Row     int
    Email   string
    Message string
}

type InviteStats struct {
    Total      int
    Sent       int
    Viewed     int
    Responded  int
    Revoked    int
}
```

### 3.2 Token Generator Interface

```go
package token

type Generator interface {
    Generate() (string, error)
    Hash(token string) (string, error)
}
```

### 3.3 Token Validator Interface

```go
package token

type Validator interface {
    Validate(token, hash string) bool
}
```

---

## 4. Implementation

### 4.1 Token Generator

```go
package token

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
)

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
    return base64.URLEncoding.EncodeToString(b), nil
}

func (g *generator) Hash(token string) (string, error) {
    h := hmac.New(sha256.New, g.secret)
    h.Write([]byte(token))
    return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}
```

### 4.2 Token Validator

```go
package token

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
)

type validator struct {
    secret []byte
}

func NewValidator(secret []byte) Validator {
    return &validator{secret: secret}
}

func (v *validator) Validate(token, hash string) bool {
    h := hmac.New(sha256.New, v.secret)
    h.Write([]byte(token))
    expected := base64.URLEncoding.EncodeToString(h.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(hash))
}
```

### 4.3 CSV Import

```go
package invites

import (
    "encoding/csv"
    "fmt"
    "io"
    "strings"
)

func (s *service) ImportCSV(ctx context.Context, eventID int64, csvData []byte) (*ImportResult, error) {
    reader := csv.NewReader(strings.NewReader(string(csvData)))
    
    header, err := reader.Read()
    if err != nil {
        return nil, fmt.Errorf("failed to read CSV header: %w", err)
    }
    
    colMap := parseHeader(header)
    if colMap["email"] == -1 {
        return nil, fmt.Errorf("CSV must have 'email' column")
    }
    
    result := &ImportResult{}
    invites := []*models.Invite{}
    row := 1
    
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            result.Errors = append(result.Errors, ImportError{
                Row:     row,
                Message: err.Error(),
            })
            result.Failed++
            row++
            continue
        }
        
        email := strings.TrimSpace(record[colMap["email"]])
        if email == "" {
            result.Failed++
            row++
            continue
        }
        
        name := ""
        if colMap["name"] != -1 {
            name = strings.TrimSpace(record[colMap["name"]])
        }
        
        maxPlusOnes := 0
        if colMap["max_plus_ones"] != -1 {
            fmt.Sscanf(record[colMap["max_plus_ones"]], "%d", &maxPlusOnes)
        }
        
        invites = append(invites, &models.Invite{
            EventID:     eventID,
            Name:        &name,
            Email:       &email,
            MaxPlusOnes: maxPlusOnes,
        })
        
        row++
    }
    
    result.Total = len(invites)
    
    results, err := s.CreateInviteBatch(ctx, invites)
    if err != nil {
        return nil, err
    }
    
    for _, r := range results {
        if r.Error != nil {
            result.Failed++
            result.Errors = append(result.Errors, ImportError{
                Email:   *r.Invite.Email,
                Message: r.Error.Error(),
            })
        } else {
            result.Created++
        }
    }
    
    return result, nil
}
```

---

## 5. Security

**Token Generation:** crypto/rand (256-bit)  
**Token Hashing:** HMAC-SHA256 with secret key  
**Token Validation:** Constant-time comparison  
**Token Storage:** Hash only, never plain text

---

## 6. Testing

```go
func TestTokenGenerator_Generate(t *testing.T) {
    gen := NewGenerator([]byte("secret"))
    
    token1, err := gen.Generate()
    if err != nil {
        t.Fatal(err)
    }
    
    token2, err := gen.Generate()
    if err != nil {
        t.Fatal(err)
    }
    
    if token1 == token2 {
        t.Error("Tokens should be unique")
    }
    
    if len(token1) != 43 {
        t.Errorf("Expected 43 chars, got %d", len(token1))
    }
}

func TestTokenValidator_Validate(t *testing.T) {
    secret := []byte("secret")
    gen := NewGenerator(secret)
    val := NewValidator(secret)
    
    token, _ := gen.Generate()
    hash, _ := gen.Hash(token)
    
    if !val.Validate(token, hash) {
        t.Error("Valid token should validate")
    }
    
    if val.Validate("wrong", hash) {
        t.Error("Invalid token should not validate")
    }
}
```

---

**Document Status:** ✅ Complete

**Next Domain:** [Domain 4: RSVP & Preference Questions](04_RSVP_LLD.md)
