# User Story: `internal/guestauth/` Package

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 8 hours  

---

## User Story

As a **developer**, I want **an `internal/guestauth/` package that encapsulates OTP generation/validation, guest session management, account lookup/creation, and OTP delivery** so that **HTTP handlers have a clean, testable service layer to call**.

---

## Acceptance Criteria

- [ ] `GuestAccountService` interface and implementation in `service.go`
- [ ] `GuestSessionManager` interface and implementation in `session.go`
- [ ] OTP generation and validation logic in `otp.go`
- [ ] `OTPDelivery` interface in `delivery.go` with `EmailOTPDelivery` implementation
- [ ] Per-identifier rate limiting in `ratelimit.go`
- [ ] `GuestAccountService.RequestOTP` enforces rate limit, creates/finds account, generates and delivers OTP
- [ ] `GuestAccountService.VerifyOTP` validates code with constant-time compare, enforces TTL and single-use, creates session
- [ ] OTP codes are 6 numeric digits; stored as `HMAC-SHA256(code, idx_key)` via `Encryptor.Hash`
- [ ] `GuestSessionManager` uses cookie name `tinyrsvp_guest`, 30-day TTL, HttpOnly + Secure + SameSite=Lax
- [ ] All interfaces are mockable (used by handler tests in Story 09)
- [ ] All tests pass with timeout

---

## Technical Details

### Package Structure

```
internal/guestauth/
  otp.go           # generateOTP, hashOTP, ValidateOTP
  otp_test.go
  session.go       # GuestSessionManager interface + implementation
  session_test.go
  service.go       # GuestAccountService interface + implementation
  service_test.go
  delivery.go      # OTPDelivery interface + EmailOTPDelivery
  delivery_test.go
  ratelimit.go     # RateLimiter using GuestOTPRepository.CountRecentByIdentifier
  ratelimit_test.go
```

### GuestAccountService Interface

```go
type GuestAccountService interface {
    RequestOTP(ctx context.Context, identifier string) error
    VerifyOTP(ctx context.Context, identifier, code string) (*models.GuestSession, error)
    GetAccount(ctx context.Context, guestAccountID int64) (*models.GuestAccountView, error)
    LinkInviteToAccount(ctx context.Context, guestAccountID int64, inviteID int64) error
    ListInvites(ctx context.Context, guestAccountID int64) ([]*models.Invite, error)
    DeleteAccount(ctx context.Context, guestAccountID int64) error
}
```

### GuestSessionManager Interface

```go
type GuestSessionManager interface {
    CreateSession(ctx context.Context, guestAccountID int64, r *http.Request) (*models.GuestSession, error)
    GetSession(ctx context.Context, sessionID string) (*models.GuestSession, error)
    DeleteSession(ctx context.Context, sessionID string) error
    SetSessionCookie(w http.ResponseWriter, sessionID string) error
    ClearSessionCookie(w http.ResponseWriter) error
    GetSessionFromRequest(r *http.Request) (string, error)
}
```

Cookie name constant: `GuestSessionCookieName = "tinyrsvp_guest"`  
Session TTL constant: `GuestSessionDuration = 30 * 24 * time.Hour`

### OTP Functions (`otp.go`)

```go
const otpLength = 6

func generateOTP() (string, error)
// Returns a zero-padded 6-digit numeric string e.g. "042817"
// Uses crypto/rand to generate a number in [0, 1000000)

func hashOTP(enc crypto.Encryptor, code string) string
// Returns enc.Hash(code) — HMAC of the raw code string
```

### OTPDelivery Interface

```go
type OTPDelivery interface {
    Send(ctx context.Context, identifier, code string) error
}
```

`EmailOTPDelivery` wraps the existing `email.Service`. It detects whether `identifier` is an email address (contains `@`) and sends a templated OTP email.

`identifier` type detection (email vs phone) is based on the presence of `@`:

```go
func isEmail(identifier string) bool {
    return strings.Contains(identifier, "@")
}
```

### Normalization

```go
func normalizeIdentifier(identifier string) string {
    id := strings.TrimSpace(identifier)
    if isEmail(id) {
        return strings.ToLower(id)
    }
    // Phone: strip all non-digit characters except leading +
    // e.g. "+1 (555) 123-4567" → "+15551234567"
    return normalizePhone(id)
}
```

### RequestOTP Flow

```go
func (s *guestAccountService) RequestOTP(ctx context.Context, identifier string) error {
    normalized := normalizeIdentifier(identifier)
    idHash := s.enc.Hash(normalized)

    // Rate limit check
    if err := s.rateLimiter.Check(ctx, idHash); err != nil {
        return err // ErrRateLimitExceeded
    }

    // Find or create account
    account, err := s.findOrCreateAccount(ctx, normalized, idHash)
    if err != nil {
        return err
    }

    // Generate OTP
    code, err := generateOTP()
    if err != nil {
        return err
    }
    codeHash := hashOTP(s.enc, code)

    // Persist OTP
    otp := &models.GuestOTPCode{
        GuestAccountID: &account.ID,
        IdentifierHash: idHash,
        CodeHash:       codeHash,
        Purpose:        models.OTPPurposeLogin,
        ExpiresAt:      time.Now().Add(15 * time.Minute),
    }
    if err := s.otpRepo.Create(ctx, otp); err != nil {
        return err
    }

    // Deliver
    return s.delivery.Send(ctx, normalized, code)
}
```

### VerifyOTP Flow

```go
func (s *guestAccountService) VerifyOTP(ctx context.Context, identifier, code string) (*models.GuestSession, error) {
    normalized := normalizeIdentifier(identifier)
    idHash := s.enc.Hash(normalized)
    candidateHash := hashOTP(s.enc, code)

    otp, err := s.otpRepo.FindPending(ctx, idHash, models.OTPPurposeLogin)
    if err != nil || otp == nil {
        return nil, ErrInvalidOTP
    }
    if otp.IsExpired() {
        return nil, ErrOTPExpired
    }
    // Constant-time compare
    if !hmac.Equal([]byte(candidateHash), []byte(otp.CodeHash)) {
        return nil, ErrInvalidOTP
    }
    if err := s.otpRepo.MarkUsed(ctx, otp.ID); err != nil {
        return nil, err
    }
    return s.sessionMgr.CreateSession(ctx, *otp.GuestAccountID, nil)
}
```

### Error Types

```go
var (
    ErrRateLimitExceeded = errors.New("too many OTP requests; try again later")
    ErrInvalidOTP        = errors.New("invalid or expired code")
    ErrOTPExpired        = errors.New("code has expired")
)
```

---

## Tasks

### Phase 1: OTP (TDD)
- [ ] Write test: `TestGenerateOTP_Length` — always 6 characters
- [ ] Write test: `TestGenerateOTP_Numeric` — only digits
- [ ] Write test: `TestGenerateOTP_Unique` — 1000 samples have no duplicates (probabilistic)
- [ ] Write test: `TestHashOTP_Deterministic` — same code always same hash
- [ ] Run tests (should fail)
- [ ] Implement `otp.go`
- [ ] Run tests (should pass)

### Phase 2: Rate Limiter (TDD)
- [ ] Write test: `TestRateLimiter_UnderLimit` — 2 requests in 1 hour passes
- [ ] Write test: `TestRateLimiter_AtLimit` — 3 requests in 1 hour passes
- [ ] Write test: `TestRateLimiter_OverLimit` — 4 requests in 1 hour returns error
- [ ] Run tests (should fail)
- [ ] Implement `ratelimit.go`
- [ ] Run tests (should pass)

### Phase 3: GuestSessionManager (TDD)
- [ ] Write test: `TestGuestSessionManager_CreateAndGet`
- [ ] Write test: `TestGuestSessionManager_GetExpired` — returns error
- [ ] Write test: `TestGuestSessionManager_Delete`
- [ ] Write test: `TestGuestSessionManager_CookieName` — cookie name is `tinyrsvp_guest`
- [ ] Run tests (should fail)
- [ ] Implement `session.go`
- [ ] Run tests (should pass)

### Phase 4: GuestAccountService (TDD)
- [ ] Write test: `TestGuestAccountService_RequestOTP_RateLimited`
- [ ] Write test: `TestGuestAccountService_RequestOTP_CreatesAccount`
- [ ] Write test: `TestGuestAccountService_RequestOTP_ExistingAccount`
- [ ] Write test: `TestGuestAccountService_VerifyOTP_Valid`
- [ ] Write test: `TestGuestAccountService_VerifyOTP_WrongCode`
- [ ] Write test: `TestGuestAccountService_VerifyOTP_Expired`
- [ ] Write test: `TestGuestAccountService_VerifyOTP_AlreadyUsed`
- [ ] Run tests (should fail)
- [ ] Implement `service.go`
- [ ] Run tests (should pass)

### Phase 5: EmailOTPDelivery (TDD)
- [ ] Write test: `TestEmailOTPDelivery_Send_EmailIdentifier` — calls email.Service correctly
- [ ] Write test: `TestEmailOTPDelivery_Send_PhoneIdentifier` — returns ErrSMSNotConfigured
- [ ] Run tests (should fail)
- [ ] Implement `delivery.go`
- [ ] Run tests (should pass)

---

## Testing Requirements

```go
func TestGuestAccountService_VerifyOTP_AlreadyUsed(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockOTPRepo := mockrepos.NewMockGuestOTPRepository(ctrl)
    usedAt := time.Now().Add(-time.Minute)
    mockOTPRepo.EXPECT().
        FindPending(gomock.Any(), gomock.Any(), models.OTPPurposeLogin).
        Return(&models.GuestOTPCode{
            ID:       1,
            CodeHash: testHashCode("123456"),
            ExpiresAt: time.Now().Add(time.Hour),
            UsedAt:   &usedAt,
        }, nil)

    svc := guestauth.NewGuestAccountService(mockOTPRepo, nil, nil, testEnc)
    _, err := svc.VerifyOTP(context.Background(), "alice@example.com", "123456")
    if err == nil {
        t.Error("expected error for already-used OTP")
    }
}
```

---

## Dependencies

**Depends on:** Story 07 (repositories), Story 01 (crypto package), Epic 05 (email.Service)  
**Blocks:** Story 09 (handlers), Story 10 (middleware)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass: `go test -timeout 30s -race ./internal/guestauth/...`
- [ ] All interfaces exported and mockable
- [ ] `go vet ./...` clean
