# User Story: Encrypt `invites` Table PII

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 4 hours  

---

## User Story

As a **system operator**, I want **the `invites` table to store guest email and name as encrypted ciphertext with a searchable HMAC blind index** so that **guest PII is not exposed if the database file is compromised**.

---

## Acceptance Criteria

- [ ] Migration `000015` renames `invites.email` → `invites.email_encrypted`, adds `invites.email_hash`
- [ ] Migration `000015` renames `invites.name` → `invites.name_encrypted`
- [ ] `InviteRepository` accepts an `Encryptor` and encrypts on write, decrypts on read
- [ ] `GetByEmail` performs lookup via `WHERE email_hash = ?`
- [ ] Email normalization applied before hashing and encryption (same as Story 02)
- [ ] `name` field is nullable — `nil` name is stored as SQL NULL, not as encrypted empty string
- [ ] All existing `InviteRepository` tests updated and passing
- [ ] All tests pass with timeout

---

## Technical Details

### Migration 000015 (SQLite) — invites block

Append to the same `000015_encrypt_pii.up.sql` created in Story 02:

```sql
-- invites table
ALTER TABLE invites RENAME COLUMN email TO email_encrypted;
ALTER TABLE invites ADD COLUMN email_hash TEXT;
ALTER TABLE invites RENAME COLUMN name TO name_encrypted;

DROP INDEX IF EXISTS idx_invites_email;
CREATE INDEX idx_invites_email_hash ON invites(email_hash);
```

Down:

```sql
ALTER TABLE invites RENAME COLUMN email_encrypted TO email;
ALTER TABLE invites DROP COLUMN email_hash;
ALTER TABLE invites RENAME COLUMN name_encrypted TO name;

DROP INDEX IF EXISTS idx_invites_email_hash;
CREATE INDEX idx_invites_email ON invites(email);
```

### Nullable Name Handling

```go
func encryptNullable(enc crypto.Encryptor, s *string) (*string, error) {
    if s == nil {
        return nil, nil
    }
    ct, err := enc.Encrypt(*s)
    if err != nil {
        return nil, err
    }
    return &ct, nil
}

func decryptNullable(enc crypto.Encryptor, ct *string) (*string, error) {
    if ct == nil {
        return nil, nil
    }
    pt, err := enc.Decrypt(*ct)
    if err != nil {
        return nil, err
    }
    return &pt, nil
}
```

### Nullable Email Hash

When `invites.email` is NULL (invite created without email, e.g. manual token), `email_hash` is also NULL. Only non-NULL emails get hashed.

```go
func hashNullableEmail(enc crypto.Encryptor, email *string) *string {
    if email == nil {
        return nil
    }
    h := enc.Hash(normalizeEmail(*email))
    return &h
}
```

---

## Tasks

### Phase 1: Migration
- [ ] Append invites SQL blocks to `000015_encrypt_pii.up.sql`
- [ ] Append invites SQL blocks to `000015_encrypt_pii.down.sql`
- [ ] Run migration tests (should pass)

### Phase 2: Update InviteRepository (TDD)
- [ ] Write test: `TestInviteRepository_Create_EncryptsEmail` — raw row has no plaintext email
- [ ] Write test: `TestInviteRepository_Create_NullEmail` — NULL email stored as NULL, not ciphertext
- [ ] Write test: `TestInviteRepository_GetByEmail_UsesHash` — lookup by hash finds invite
- [ ] Write test: `TestInviteRepository_GetByEmail_Normalization` — case-insensitive lookup works
- [ ] Write test: `TestInviteRepository_NullName_StoredAsNull` — nil name stored as SQL NULL
- [ ] Write test: `TestInviteRepository_List_DecryptsAll` — list returns plaintext values
- [ ] Run tests (should fail)
- [ ] Update `InviteRepository` constructor and all methods
- [ ] Run tests (should pass)

### Phase 3: Update Constructor Wiring
- [ ] Update `cmd/server/main.go` to pass `Encryptor` to `NewInviteRepository`
- [ ] Run full test suite

---

## Testing Requirements

```go
func TestInviteRepository_Create_NullEmail(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewInviteRepository(db, enc)

    invite := &models.Invite{
        EventID: 1,
        Token:   "tok",
        // Email intentionally nil
    }
    err := repo.Create(context.Background(), invite)
    if err != nil {
        t.Fatalf("Create: %v", err)
    }

    var emailEnc *string
    db.QueryRow(context.Background(),
        "SELECT email_encrypted FROM invites WHERE id = ?", invite.ID,
    ).Scan(&emailEnc)

    if emailEnc != nil {
        t.Errorf("expected NULL email_encrypted, got %q", *emailEnc)
    }
}

func TestInviteRepository_GetByEmail_Normalization(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewInviteRepository(db, enc)

    email := "Bob@Example.COM"
    _ = repo.Create(context.Background(), &models.Invite{
        EventID: 1,
        Token:   "tok",
        Email:   &email,
    })

    for _, input := range []string{"bob@example.com", "BOB@EXAMPLE.COM", "Bob@Example.COM"} {
        invite, err := repo.GetByEmail(context.Background(), 1, input)
        if err != nil || invite == nil {
            t.Errorf("GetByEmail(%q): %v", input, err)
        }
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
- [ ] No plaintext PII visible in raw DB rows (verified by test)
- [ ] `go vet ./...` clean
