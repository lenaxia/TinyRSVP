# User Story: Guest Account, Session, and OTP Repositories

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 6 hours  

---

## User Story

As a **developer**, I want **`GuestAccountRepository`, `GuestSessionRepository`, and `GuestOTPRepository` with full CRUD and domain-specific queries** so that **the guest auth service layer has a clean, testable data access interface**.

---

## Acceptance Criteria

- [ ] `GuestAccountRepository` interface and implementation in `internal/db/repositories/guest_account_repository.go`
- [ ] `GuestSessionRepository` interface and implementation in `internal/db/repositories/guest_session_repository.go`
- [ ] `GuestOTPRepository` interface and implementation in `internal/db/repositories/guest_otp_repository.go`
- [ ] All three repositories accept an `Encryptor` and handle encrypt-on-write / decrypt-on-read for their PII fields
- [ ] `GuestAccountRepository.GetByEmailHash` and `GetByPhoneHash` perform `WHERE *_hash = ?` lookups
- [ ] `GuestAccountRepository.LinkInvite` sets `invites.guest_account_id`
- [ ] `GuestAccountRepository.ListInvitesByAccount` returns all invites for a guest account
- [ ] `GuestOTPRepository.FindPending` returns the most recent unused, non-expired OTP for an identifier hash
- [ ] `GuestSessionRepository.DeleteExpired` removes expired sessions (called by cleanup job)
- [ ] All repository methods use parameterized queries
- [ ] Integration tests use a real SQLite DB via `testutil.SetupTestDBWithMigrations`
- [ ] All tests pass with timeout

---

## Technical Details

### GuestAccountRepository Interface

```go
type GuestAccountRepository interface {
    Create(ctx context.Context, account *models.GuestAccount) error
    GetByID(ctx context.Context, id int64) (*models.GuestAccountView, error)
    GetByEmailHash(ctx context.Context, emailHash string) (*models.GuestAccountView, error)
    GetByPhoneHash(ctx context.Context, phoneHash string) (*models.GuestAccountView, error)
    Update(ctx context.Context, account *models.GuestAccount) error
    Delete(ctx context.Context, id int64) error
    LinkInvite(ctx context.Context, guestAccountID int64, inviteID int64) error
    ListInvitesByAccount(ctx context.Context, guestAccountID int64) ([]*models.Invite, error)
}
```

`GetByID`, `GetByEmailHash`, `GetByPhoneHash` return `*models.GuestAccountView` which includes the decrypted `Email`, `Phone`, and `DisplayName` fields populated by the repository after decryption.

### GuestSessionRepository Interface

```go
type GuestSessionRepository interface {
    Create(ctx context.Context, session *models.GuestSession) error
    GetByID(ctx context.Context, id string) (*models.GuestSession, error)
    UpdateLastAccessed(ctx context.Context, id string) error
    Delete(ctx context.Context, id string) error
    DeleteByGuestAccountID(ctx context.Context, guestAccountID int64) error
    DeleteExpired(ctx context.Context) (int64, error)
}
```

### GuestOTPRepository Interface

```go
type GuestOTPRepository interface {
    Create(ctx context.Context, otp *models.GuestOTPCode) error
    FindPending(ctx context.Context, identifierHash string, purpose models.OTPPurpose) (*models.GuestOTPCode, error)
    MarkUsed(ctx context.Context, id int64) error
    CountRecentByIdentifier(ctx context.Context, identifierHash string, since time.Time) (int, error)
    DeleteExpired(ctx context.Context) (int64, error)
}
```

`FindPending` returns the most recent OTP row where `identifier_hash = ?` AND `used_at IS NULL` AND `expires_at > NOW()`.

`CountRecentByIdentifier` is used for rate limiting: counts OTP rows for a given identifier created since a given time.

### Encrypt/Decrypt in GuestAccountRepository

```go
func (r *guestAccountRepository) Create(ctx context.Context, account *models.GuestAccount) error {
    // account.EmailHash and account.PhoneHash are pre-computed by the service layer
    // account.EmailEncrypted and account.PhoneEncrypted are pre-computed by the service layer
    // Repository stores them as-is; encryption is the caller's responsibility
    // DisplayNameEncrypted is also pre-computed
}
```

**Design note:** For guest accounts, the service layer (Story 08) computes hashes and ciphertexts before calling the repository, because the service holds the raw identifier needed to normalize and hash. The repository stores and retrieves the opaque encrypted/hashed values. On reads, the repository decrypts and populates `GuestAccountView`. This keeps normalization logic in one place (the service).

---

## Tasks

### Phase 1: GuestAccountRepository (TDD)
- [ ] Write test: `TestGuestAccountRepo_CreateAndGetByEmailHash`
- [ ] Write test: `TestGuestAccountRepo_GetByPhoneHash`
- [ ] Write test: `TestGuestAccountRepo_GetByID_DecryptsDisplayName`
- [ ] Write test: `TestGuestAccountRepo_LinkInvite`
- [ ] Write test: `TestGuestAccountRepo_ListInvitesByAccount`
- [ ] Write test: `TestGuestAccountRepo_Delete_CascadesSessionsAndOTPs`
- [ ] Run tests (should fail)
- [ ] Implement `guest_account_repository.go`
- [ ] Run tests (should pass)

### Phase 2: GuestSessionRepository (TDD)
- [ ] Write test: `TestGuestSessionRepo_CreateAndGetByID`
- [ ] Write test: `TestGuestSessionRepo_GetByID_NotFound`
- [ ] Write test: `TestGuestSessionRepo_UpdateLastAccessed`
- [ ] Write test: `TestGuestSessionRepo_Delete`
- [ ] Write test: `TestGuestSessionRepo_DeleteByGuestAccountID`
- [ ] Write test: `TestGuestSessionRepo_DeleteExpired`
- [ ] Run tests (should fail)
- [ ] Implement `guest_session_repository.go`
- [ ] Run tests (should pass)

### Phase 3: GuestOTPRepository (TDD)
- [ ] Write test: `TestGuestOTPRepo_CreateAndFindPending`
- [ ] Write test: `TestGuestOTPRepo_FindPending_ExpiredReturnsNil`
- [ ] Write test: `TestGuestOTPRepo_FindPending_UsedReturnsNil`
- [ ] Write test: `TestGuestOTPRepo_MarkUsed`
- [ ] Write test: `TestGuestOTPRepo_CountRecentByIdentifier`
- [ ] Write test: `TestGuestOTPRepo_DeleteExpired`
- [ ] Run tests (should fail)
- [ ] Implement `guest_otp_repository.go`
- [ ] Run tests (should pass)

---

## Testing Requirements

```go
func TestGuestAccountRepo_CreateAndGetByEmailHash(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewGuestAccountRepository(db, enc)

    emailHash := enc.Hash("alice@example.com")
    emailEnc, _ := enc.Encrypt("alice@example.com")

    account := &models.GuestAccount{
        EmailEncrypted: &emailEnc,
        EmailHash:      &emailHash,
    }
    err := repo.Create(context.Background(), account)
    if err != nil {
        t.Fatalf("Create: %v", err)
    }
    if account.ID == 0 {
        t.Fatal("expected ID to be set after Create")
    }

    got, err := repo.GetByEmailHash(context.Background(), emailHash)
    if err != nil {
        t.Fatalf("GetByEmailHash: %v", err)
    }
    if got.Email == nil || *got.Email != "alice@example.com" {
        t.Errorf("Email: got %v, want alice@example.com", got.Email)
    }
}

func TestGuestOTPRepo_FindPending_ExpiredReturnsNil(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewGuestOTPRepository(db, enc)

    otp := &models.GuestOTPCode{
        IdentifierHash: enc.Hash("alice@example.com"),
        CodeHash:       enc.Hash("123456"),
        Purpose:        models.OTPPurposeLogin,
        ExpiresAt:      time.Now().Add(-time.Minute), // already expired
    }
    _ = repo.Create(context.Background(), otp)

    got, err := repo.FindPending(context.Background(), otp.IdentifierHash, models.OTPPurposeLogin)
    if err != nil {
        t.Fatalf("FindPending: %v", err)
    }
    if got != nil {
        t.Error("expected nil for expired OTP, got a result")
    }
}
```

---

## Dependencies

**Depends on:** Story 06 (models and migration 000016)  
**Blocks:** Story 08 (guestauth package)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass: `go test -timeout 30s -race ./internal/db/repositories/...`
- [ ] No plaintext PII in raw DB rows (verified by tests)
- [ ] All queries use parameterized placeholders
- [ ] `go vet ./...` clean
