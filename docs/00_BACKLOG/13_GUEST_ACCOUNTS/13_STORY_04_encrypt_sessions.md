# User Story: Encrypt `sessions` Table PII

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 3 hours  

---

## User Story

As a **system operator**, I want **the `sessions` table to store IP address and user agent as encrypted ciphertext** so that **visitor metadata is not exposed if the database file is compromised**.

---

## Acceptance Criteria

- [ ] Migration `000015` renames `sessions.ip_address` → `sessions.ip_address_encrypted`
- [ ] Migration `000015` renames `sessions.user_agent` → `sessions.user_agent_encrypted`
- [ ] `SessionRepository` accepts an `Encryptor` and encrypts on write, decrypts on read
- [ ] Both fields are nullable — `nil` values stored as SQL NULL, not as encrypted empty strings
- [ ] `models.Session` struct fields remain `IPAddress *string` and `UserAgent *string` (plaintext in-process)
- [ ] No blind index needed for either field (neither is used in a `WHERE` clause)
- [ ] All existing `SessionRepository` tests updated and passing
- [ ] All tests pass with timeout

---

## Technical Details

### Migration 000015 (SQLite) — sessions block

Append to the shared `000015_encrypt_pii.up.sql`:

```sql
-- sessions table
ALTER TABLE sessions RENAME COLUMN ip_address TO ip_address_encrypted;
ALTER TABLE sessions RENAME COLUMN user_agent TO user_agent_encrypted;
```

Down:

```sql
ALTER TABLE sessions RENAME COLUMN ip_address_encrypted TO ip_address;
ALTER TABLE sessions RENAME COLUMN user_agent_encrypted TO user_agent;
```

### Repository Constructor Change

```go
// Before:
func NewSessionRepository(db Database) SessionRepository

// After:
func NewSessionRepository(db Database, enc crypto.Encryptor) SessionRepository
```

### Write Path

```go
ipEnc, err  := encryptNullable(enc, session.IPAddress)
uaEnc, err  := encryptNullable(enc, session.UserAgent)
// INSERT ... ip_address_encrypted, user_agent_encrypted
```

### Read Path

```go
var ipEnc, uaEnc *string
row.Scan(..., &ipEnc, &uaEnc)
session.IPAddress, err = decryptNullable(enc, ipEnc)
session.UserAgent, err  = decryptNullable(enc, uaEnc)
```

The `encryptNullable` / `decryptNullable` helpers are the same pattern introduced in Story 03.

---

## Tasks

### Phase 1: Migration
- [ ] Append sessions SQL blocks to `000015_encrypt_pii.up.sql`
- [ ] Append sessions SQL blocks to `000015_encrypt_pii.down.sql`
- [ ] Verify migration applies and rolls back cleanly

### Phase 2: Update SessionRepository (TDD)
- [ ] Write test: `TestSessionRepository_Create_EncryptsIPAddress` — raw row contains no plaintext IP
- [ ] Write test: `TestSessionRepository_Create_NullIP` — NULL IP stored as SQL NULL
- [ ] Write test: `TestSessionRepository_GetByID_DecryptsFields` — retrieved session has plaintext IP and UA
- [ ] Write test: `TestSessionRepository_Update_ReEncrypts` — updating session re-encrypts correctly
- [ ] Run tests (should fail)
- [ ] Update `SessionRepository` constructor and all methods
- [ ] Run tests (should pass)

### Phase 3: Wire Encryptor
- [ ] Update `cmd/server/main.go` to pass `Encryptor` to `NewSessionRepository`
- [ ] Run full test suite

---

## Testing Requirements

```go
func TestSessionRepository_Create_EncryptsIPAddress(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewSessionRepository(db, enc)

    ip := "192.168.1.1"
    ua := "Mozilla/5.0"
    session := &models.Session{
        ID:        "sess123",
        UserID:    1,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
        IPAddress: &ip,
        UserAgent: &ua,
    }
    err := repo.Create(context.Background(), session)
    if err != nil {
        t.Fatalf("Create: %v", err)
    }

    var rawIP *string
    db.QueryRow(context.Background(),
        "SELECT ip_address_encrypted FROM sessions WHERE id = ?", session.ID,
    ).Scan(&rawIP)

    if rawIP != nil && *rawIP == "192.168.1.1" {
        t.Error("IP address stored as plaintext, expected ciphertext")
    }
}

func TestSessionRepository_GetByID_DecryptsFields(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewSessionRepository(db, enc)

    ip := "10.0.0.1"
    ua := "TestAgent/1.0"
    original := &models.Session{
        ID:        "sess456",
        UserID:    1,
        ExpiresAt: time.Now().Add(time.Hour),
        IPAddress: &ip,
        UserAgent: &ua,
    }
    _ = repo.Create(context.Background(), original)

    got, err := repo.GetByID(context.Background(), "sess456")
    if err != nil {
        t.Fatalf("GetByID: %v", err)
    }
    if got.IPAddress == nil || *got.IPAddress != ip {
        t.Errorf("IPAddress: got %v, want %q", got.IPAddress, ip)
    }
    if got.UserAgent == nil || *got.UserAgent != ua {
        t.Errorf("UserAgent: got %v, want %q", got.UserAgent, ua)
    }
}
```

---

## Dependencies

**Depends on:** Story 01 (crypto package), Story 02 (migration 000015 file created)  
**Blocks:** Nothing directly

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All existing and new tests pass: `go test -timeout 30s -race ./internal/db/...`
- [ ] No plaintext IP address visible in raw DB rows (verified by test)
- [ ] `go vet ./...` clean
