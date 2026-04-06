# User Story: Encrypt `users` Table PII

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 4 hours  

---

## User Story

As a **system operator**, I want **the `users` table to store email and name as encrypted ciphertext with a searchable HMAC blind index** so that **PII is not exposed if the database file is compromised**.

---

## Acceptance Criteria

- [ ] Migration `000015` renames `users.email` → `users.email_encrypted`, adds `users.email_hash`
- [ ] Migration `000015` renames `users.name` → `users.name_encrypted`
- [ ] `UserRepository` accepts an `Encryptor` and encrypts on write, decrypts on read
- [ ] `GetByEmail` performs lookup via `WHERE email_hash = ?` using `Encryptor.Hash(normalize(email))`
- [ ] `models.User` struct fields remain `Email string` and `Name string` (plaintext in-process)
- [ ] Email normalization (lowercase + trim) applied before hashing and before encryption
- [ ] All existing `UserRepository` tests updated and passing
- [ ] `GetByOIDCSubject` and all other non-email queries unchanged
- [ ] All tests pass with timeout

---

## Technical Details

### Migration 000015 (SQLite)

```sql
-- migrations/sqlite/000015_encrypt_pii.up.sql
ALTER TABLE users RENAME COLUMN email TO email_encrypted;
ALTER TABLE users ADD COLUMN email_hash TEXT;
ALTER TABLE users RENAME COLUMN name TO name_encrypted;

DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_email_hash ON users(email_hash);
```

```sql
-- migrations/sqlite/000015_encrypt_pii.down.sql
ALTER TABLE users RENAME COLUMN email_encrypted TO email;
ALTER TABLE users DROP COLUMN email_hash;
ALTER TABLE users RENAME COLUMN name_encrypted TO name;

DROP INDEX IF EXISTS idx_users_email_hash;
CREATE UNIQUE INDEX idx_users_email ON users(email);
```

The same migration file also covers `invites`, `sessions`, and `email_queue` changes (Stories 03–05 are separate stories but share migration 000015). See Stories 03, 04, 05 for their respective SQL blocks within the same migration file.

### Repository Constructor Change

```go
// Before:
func NewUserRepository(db Database) UserRepository

// After:
func NewUserRepository(db Database, enc crypto.Encryptor) UserRepository
```

### Normalization

```go
func normalizeEmail(email string) string {
    return strings.ToLower(strings.TrimSpace(email))
}
```

Applied before both `enc.Hash()` and `enc.Encrypt()`.

### Write Path

```go
// In Create / Update:
normalized := normalizeEmail(user.Email)
emailHash, err := enc.Hash(normalized)        // for WHERE lookups
emailEnc, err  := enc.Encrypt(normalized)     // for storage

// INSERT INTO users (..., email_encrypted, email_hash, name_encrypted, ...)
// VALUES (..., ?, ?, ?, ...)
```

### Read Path

```go
// Scan into local vars, then decrypt:
var emailEnc, nameEnc string
row.Scan(..., &emailEnc, &nameEnc, ...)
user.Email, err = enc.Decrypt(emailEnc)
user.Name, err  = enc.Decrypt(nameEnc)
```

### GetByEmail

```go
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    hash := r.enc.Hash(normalizeEmail(email))
    query := `SELECT ... FROM users WHERE email_hash = ?`
    // scan + decrypt as above
}
```

---

## Tasks

### Phase 1: Migration (TDD)
- [ ] Write migration test: up migration creates `email_hash` column, removes old `email` column
- [ ] Write migration test: down migration reverses the change
- [ ] Add SQL to `000015_encrypt_pii.up.sql` for users table
- [ ] Add SQL to `000015_encrypt_pii.down.sql` for users table
- [ ] Run migration tests (should pass)

### Phase 2: Update UserRepository (TDD)
- [ ] Write test: `TestUserRepository_Create_EncryptsEmail` — stored row has no plaintext email
- [ ] Write test: `TestUserRepository_GetByEmail_UsesHash` — lookup by hash works
- [ ] Write test: `TestUserRepository_GetByEmail_Normalization` — mixed-case input still finds row
- [ ] Write test: `TestUserRepository_Update_ReEncrypts` — updating email re-encrypts correctly
- [ ] Write test: `TestUserRepository_List_DecryptsAll` — list returns plaintext email in structs
- [ ] Run tests (should fail — repository not updated yet)
- [ ] Update `UserRepository` constructor and all methods
- [ ] Run tests (should pass)

### Phase 3: Wire Encryptor at Startup
- [ ] Update `cmd/server/main.go` (or wherever repositories are constructed) to call `crypto.NewEncryptorFromEnv()` and pass to `NewUserRepository`
- [ ] Verify app fails to start with clear error when `TINYRSVP_ENCRYPTION_KEY` is unset
- [ ] Run full test suite

---

## Testing Requirements

```go
func TestUserRepository_Create_EncryptsEmail(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewUserRepository(db, enc)

    user := &models.User{
        Email: "Alice@Example.COM",
        Name:  "Alice",
        Role:  models.RoleEventManager,
    }
    err := repo.Create(context.Background(), user)
    if err != nil {
        t.Fatalf("Create: %v", err)
    }

    // Read raw row — email column must not be plaintext
    var rawEmail string
    db.QueryRow(context.Background(),
        "SELECT email_encrypted FROM users WHERE id = ?", user.ID,
    ).Scan(&rawEmail)

    if rawEmail == "alice@example.com" || rawEmail == "Alice@Example.COM" {
        t.Error("email stored as plaintext, expected ciphertext")
    }
}

func TestUserRepository_GetByEmail_Normalization(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewUserRepository(db, enc)

    _ = repo.Create(context.Background(), &models.User{
        Email: "alice@example.com",
        Role:  models.RoleEventManager,
    })

    tests := []string{"alice@example.com", "Alice@Example.COM", "ALICE@EXAMPLE.COM", "  alice@example.com  "}
    for _, input := range tests {
        user, err := repo.GetByEmail(context.Background(), input)
        if err != nil {
            t.Errorf("GetByEmail(%q): %v", input, err)
        }
        if user == nil || user.Email != "alice@example.com" {
            t.Errorf("GetByEmail(%q): expected alice@example.com, got %v", input, user)
        }
    }
}
```

---

## Dependencies

**Depends on:** Story 01 (crypto package)  
**Blocks:** Nothing directly, but all stories that construct `UserRepository` must be updated to pass `Encryptor`

**Note:** Migration 000015 is shared across Stories 02–05. Story 02 creates the file; Stories 03–05 append to it. Coordinate to avoid conflicts if working in parallel.

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All existing and new tests pass: `go test -timeout 30s -race ./internal/db/...`
- [ ] No plaintext PII visible in raw DB rows (verified by test)
- [ ] `go vet ./...` clean
