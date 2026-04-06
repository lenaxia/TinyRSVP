# User Story: Guest Account Models and Migration 000016

**Epic:** [Epic 13 - Guest Accounts & Encryption at Rest](README.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 4 hours  

---

## User Story

As a **developer**, I want **`GuestAccount`, `GuestSession`, and `GuestOTPCode` model structs and their corresponding database tables** so that **the guest auth system has a well-typed foundation to build on**.

---

## Acceptance Criteria

- [ ] `models.GuestAccount`, `models.GuestSession`, `models.GuestOTPCode` structs defined in `internal/models/guest_account.go`
- [ ] Migration `000016` creates `guest_accounts`, `guest_sessions`, `guest_otp_codes` tables
- [ ] Migration `000016` adds `guest_account_id` FK column to `invites`
- [ ] SQLite and PostgreSQL migration files both present
- [ ] `GuestAccount` enforces at least one of `EmailHash` or `PhoneHash` being non-nil (validated in `Validate()`)
- [ ] `GuestOTPCode.Purpose` is a typed string constant, not a raw string
- [ ] `GuestSession.IsExpired()` method implemented
- [ ] Migration down reverses all changes cleanly
- [ ] All model validation tests pass with timeout

---

## Technical Details

### Structs — `internal/models/guest_account.go`

```go
package models

import "time"

type GuestAccount struct {
    ID                   int64
    EmailEncrypted       *string
    EmailHash            *string
    PhoneEncrypted       *string
    PhoneHash            *string
    DisplayNameEncrypted *string
    CreatedAt            time.Time
    UpdatedAt            time.Time
}

// DisplayEmail and DisplayPhone are set by the repository after decryption.
// They are NOT persisted — repository populates them on read.
type GuestAccountView struct {
    GuestAccount
    Email       *string
    Phone       *string
    DisplayName *string
}

type GuestSession struct {
    ID                  string
    GuestAccountID      int64
    CreatedAt           time.Time
    ExpiresAt           time.Time
    LastAccessedAt      time.Time
    IPAddressEncrypted  *string
    UserAgentEncrypted  *string
}

func (s *GuestSession) IsExpired() bool {
    return time.Now().After(s.ExpiresAt)
}

type OTPPurpose string

const (
    OTPPurposeLogin  OTPPurpose = "login"
    OTPPurposeEnroll OTPPurpose = "enroll"
)

type GuestOTPCode struct {
    ID             int64
    GuestAccountID *int64
    IdentifierHash string
    CodeHash       string
    Purpose        OTPPurpose
    CreatedAt      time.Time
    ExpiresAt      time.Time
    UsedAt         *time.Time
}

func (o *GuestOTPCode) IsExpired() bool {
    return time.Now().After(o.ExpiresAt)
}

func (o *GuestOTPCode) IsUsed() bool {
    return o.UsedAt != nil
}
```

### Migration 000016 (SQLite)

```sql
-- migrations/sqlite/000016_guest_accounts.up.sql
CREATE TABLE guest_accounts (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    email_encrypted         TEXT,
    email_hash              TEXT UNIQUE,
    phone_encrypted         TEXT,
    phone_hash              TEXT UNIQUE,
    display_name_encrypted  TEXT,
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (email_hash IS NOT NULL OR phone_hash IS NOT NULL)
);

CREATE TABLE guest_sessions (
    id                      TEXT PRIMARY KEY,
    guest_account_id        INTEGER NOT NULL REFERENCES guest_accounts(id) ON DELETE CASCADE,
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at              TIMESTAMP NOT NULL,
    last_accessed_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address_encrypted    TEXT,
    user_agent_encrypted    TEXT
);

CREATE TABLE guest_otp_codes (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    guest_account_id    INTEGER REFERENCES guest_accounts(id) ON DELETE CASCADE,
    identifier_hash     TEXT NOT NULL,
    code_hash           TEXT NOT NULL,
    purpose             TEXT NOT NULL CHECK (purpose IN ('login', 'enroll')),
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at          TIMESTAMP NOT NULL,
    used_at             TIMESTAMP
);

CREATE INDEX idx_guest_sessions_account ON guest_sessions(guest_account_id);
CREATE INDEX idx_guest_sessions_expires ON guest_sessions(expires_at);
CREATE INDEX idx_guest_otp_identifier   ON guest_otp_codes(identifier_hash);
CREATE INDEX idx_guest_otp_expires      ON guest_otp_codes(expires_at);

ALTER TABLE invites ADD COLUMN guest_account_id INTEGER REFERENCES guest_accounts(id) ON DELETE SET NULL;
CREATE INDEX idx_invites_guest_account ON invites(guest_account_id);
```

```sql
-- migrations/sqlite/000016_guest_accounts.down.sql
DROP INDEX IF EXISTS idx_invites_guest_account;
-- SQLite does not support DROP COLUMN in older versions; use table rebuild if needed
-- For simplicity: drop and recreate invites without the column, or accept the column remains on down
ALTER TABLE invites DROP COLUMN guest_account_id;

DROP INDEX IF EXISTS idx_guest_otp_expires;
DROP INDEX IF EXISTS idx_guest_otp_identifier;
DROP INDEX IF EXISTS idx_guest_sessions_expires;
DROP INDEX IF EXISTS idx_guest_sessions_account;
DROP TABLE IF EXISTS guest_otp_codes;
DROP TABLE IF EXISTS guest_sessions;
DROP TABLE IF EXISTS guest_accounts;
```

### Validation

```go
func (a *GuestAccount) Validate() error {
    if a.EmailHash == nil && a.PhoneHash == nil {
        return &ValidationError{Field: "identifier", Message: "email or phone is required"}
    }
    return nil
}
```

---

## Tasks

### Phase 1: Models (TDD)
- [ ] Write test: `TestGuestAccount_Validate_NoIdentifier` — returns error
- [ ] Write test: `TestGuestAccount_Validate_EmailOnly` — passes
- [ ] Write test: `TestGuestAccount_Validate_PhoneOnly` — passes
- [ ] Write test: `TestGuestSession_IsExpired_Future` — returns false
- [ ] Write test: `TestGuestSession_IsExpired_Past` — returns true
- [ ] Write test: `TestGuestOTPCode_IsUsed_Nil` — returns false
- [ ] Write test: `TestGuestOTPCode_IsUsed_Set` — returns true
- [ ] Run tests (should fail)
- [ ] Implement `internal/models/guest_account.go`
- [ ] Run tests (should pass)

### Phase 2: Migration (TDD)
- [ ] Write migration test: up creates all three tables and `invites.guest_account_id`
- [ ] Write migration test: down removes all three tables and the FK column
- [ ] Create `migrations/sqlite/000016_guest_accounts.up.sql`
- [ ] Create `migrations/sqlite/000016_guest_accounts.down.sql`
- [ ] Create `migrations/postgres/000016_guest_accounts.up.sql` (PostgreSQL syntax equivalents)
- [ ] Create `migrations/postgres/000016_guest_accounts.down.sql`
- [ ] Run migration tests (should pass)

---

## Testing Requirements

```go
func TestGuestSession_IsExpired(t *testing.T) {
    tests := []struct {
        name    string
        session GuestSession
        want    bool
    }{
        {
            name:    "not expired",
            session: GuestSession{ExpiresAt: time.Now().Add(time.Hour)},
            want:    false,
        },
        {
            name:    "expired",
            session: GuestSession{ExpiresAt: time.Now().Add(-time.Second)},
            want:    true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.session.IsExpired(); got != tt.want {
                t.Errorf("IsExpired() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## Dependencies

**Depends on:** Story 01 (crypto package), Stories 02–05 (migration 000015 complete before 000016)  
**Blocks:** Story 07 (repositories), all subsequent guest auth stories

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] Model tests pass: `go test -timeout 30s ./internal/models/...`
- [ ] Migration applies and rolls back without error
- [ ] `go vet ./...` clean
