# User Story: Encrypt `email_queue` Table PII

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 3 hours  

---

## User Story

As a **system operator**, I want **the `email_queue` table to store recipient email and name as encrypted ciphertext with a searchable HMAC blind index** so that **queued recipient data is not exposed if the database file is compromised**.

---

## Acceptance Criteria

- [ ] Migration `000015` renames `email_queue.to_email` → `email_queue.to_email_encrypted`, adds `email_queue.to_email_hash`
- [ ] Migration `000015` renames `email_queue.to_name` → `email_queue.to_name_encrypted`
- [ ] `EmailQueueRepository` accepts an `Encryptor` and encrypts on write, decrypts on read
- [ ] `to_email_hash` is used for any `WHERE to_email = ?` lookups (e.g. bounce handling, unsubscribe)
- [ ] `to_name` is nullable — `nil` stored as SQL NULL
- [ ] `body_text` and `body_html` are **not** encrypted (operational debuggability; not raw PII)
- [ ] `models.EmailQueueItem` struct fields remain plaintext in-process
- [ ] All existing `EmailQueueRepository` tests updated and passing
- [ ] All tests pass with timeout

---

## Technical Details

### Migration 000015 (SQLite) — email_queue block

Append to the shared `000015_encrypt_pii.up.sql`:

```sql
-- email_queue table
ALTER TABLE email_queue RENAME COLUMN to_email TO to_email_encrypted;
ALTER TABLE email_queue ADD COLUMN to_email_hash TEXT;
ALTER TABLE email_queue RENAME COLUMN to_name TO to_name_encrypted;

DROP INDEX IF EXISTS idx_email_queue_to_email;
CREATE INDEX idx_email_queue_to_email_hash ON email_queue(to_email_hash);
```

Down:

```sql
ALTER TABLE email_queue RENAME COLUMN to_email_encrypted TO to_email;
ALTER TABLE email_queue DROP COLUMN to_email_hash;
ALTER TABLE email_queue RENAME COLUMN to_name_encrypted TO to_name;

DROP INDEX IF EXISTS idx_email_queue_to_email_hash;
```

### Repository Constructor Change

```go
// Before:
func NewEmailQueueRepository(db Database) EmailQueueRepository

// After:
func NewEmailQueueRepository(db Database, enc crypto.Encryptor) EmailQueueRepository
```

### Normalization

Email normalization (lowercase + trim) applied before hashing, matching the pattern in Stories 02 and 03.

---

## Tasks

### Phase 1: Migration
- [ ] Append email_queue SQL blocks to `000015_encrypt_pii.up.sql`
- [ ] Append email_queue SQL blocks to `000015_encrypt_pii.down.sql`
- [ ] Verify migration applies and rolls back cleanly

### Phase 2: Update EmailQueueRepository (TDD)
- [ ] Write test: `TestEmailQueueRepository_Enqueue_EncryptsToEmail` — raw row has no plaintext recipient email
- [ ] Write test: `TestEmailQueueRepository_Enqueue_NullName` — nil name stored as SQL NULL
- [ ] Write test: `TestEmailQueueRepository_GetByEmail_UsesHash` — lookup by hash finds items
- [ ] Write test: `TestEmailQueueRepository_GetByEmail_Normalization` — mixed-case lookup works
- [ ] Write test: `TestEmailQueueRepository_List_DecryptsAll` — dequeued items have plaintext email
- [ ] Write test: `TestEmailQueueRepository_BodyNotEncrypted` — body_text is stored as-is
- [ ] Run tests (should fail)
- [ ] Update `EmailQueueRepository` constructor and all methods
- [ ] Run tests (should pass)

### Phase 3: Wire Encryptor
- [ ] Update `cmd/server/main.go` to pass `Encryptor` to `NewEmailQueueRepository`
- [ ] Run full test suite

---

## Testing Requirements

```go
func TestEmailQueueRepository_Enqueue_EncryptsToEmail(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewEmailQueueRepository(db, enc)

    item := &models.EmailQueueItem{
        ToEmail:  "recipient@example.com",
        Subject:  "Your invitation",
        BodyText: "Hello there",
    }
    err := repo.Enqueue(context.Background(), item)
    if err != nil {
        t.Fatalf("Enqueue: %v", err)
    }

    var rawEmail string
    db.QueryRow(context.Background(),
        "SELECT to_email_encrypted FROM email_queue WHERE id = ?", item.ID,
    ).Scan(&rawEmail)

    if rawEmail == "recipient@example.com" {
        t.Error("to_email stored as plaintext, expected ciphertext")
    }
}

func TestEmailQueueRepository_BodyNotEncrypted(t *testing.T) {
    db := testutil.SetupTestDBWithMigrations(t, "../../migrations/sqlite")
    enc, _ := crypto.NewEncryptor(make([]byte, 32))
    repo := repositories.NewEmailQueueRepository(db, enc)

    body := "Plain text body content"
    item := &models.EmailQueueItem{
        ToEmail:  "x@example.com",
        Subject:  "Test",
        BodyText: body,
    }
    _ = repo.Enqueue(context.Background(), item)

    var rawBody string
    db.QueryRow(context.Background(),
        "SELECT body_text FROM email_queue WHERE id = ?", item.ID,
    ).Scan(&rawBody)

    if rawBody != body {
        t.Errorf("body_text should not be encrypted: got %q, want %q", rawBody, body)
    }
}
```

---

## Dependencies

**Depends on:** Story 01 (crypto package), Story 02 (migration 000015 file created)  
**Blocks:** Nothing directly

**Note:** This completes migration 000015. After this story, all four existing tables (`users`, `invites`, `sessions`, `email_queue`) have their PII encrypted. Run the full test suite before starting Story 06.

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All existing and new tests pass: `go test -timeout 30s -race ./internal/db/...`
- [ ] No plaintext recipient email visible in raw DB rows (verified by test)
- [ ] `body_text` and `body_html` are confirmed NOT encrypted (verified by test)
- [ ] `go vet ./...` clean
