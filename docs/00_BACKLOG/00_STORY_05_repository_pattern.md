# User Story: Repository Pattern Implementation

**Epic:** [00_EPIC_foundation.md](00_EPIC_foundation.md)
**Priority:** High
**Status:** Complete
**Estimated Effort:** 6 hours
**Actual Effort:** 2 hours
**Completed:** 2026-01-06

---

## User Story

As a **developer**, I want **repository pattern implementation for data access** so that **database operations are abstracted and testable**.

---

## Acceptance Criteria

- [x] Domain models package created with all model structs
- [x] Base repository interface defined
- [x] User repository implemented with all methods
- [x] Session repository implemented with all methods
- [x] Config repository implemented with all methods
- [x] Transaction support in repositories
- [x] Error mapping from database to domain errors
- [x] All repositories fully tested
- [x] All tests pass with timeout

---

## Technical Details

### Repository Pattern Benefits

- **Abstraction:** Hide database implementation details
- **Testability:** Easy to mock for unit tests
- **Consistency:** Standardized data access patterns
- **Maintainability:** Centralized database logic
- **Flexibility:** Easy to swap database implementations

### Base Repository Interface

```go
package repositories

import "context"

type Repository interface {
    Create(ctx context.Context, entity interface{}) error
    GetByID(ctx context.Context, id int64) (interface{}, error)
    Update(ctx context.Context, entity interface{}) error
    Delete(ctx context.Context, id int64) error
}
```

### User Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id int64) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    GetByOIDCSubject(ctx context.Context, subject string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, limit, offset int) ([]*models.User, error)
    Count(ctx context.Context) (int, error)
    IsFirstUser(ctx context.Context) (bool, error)
    UpdateLastLogin(ctx context.Context, userID int64) error
}
```

### Session Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type SessionRepository interface {
    Create(ctx context.Context, session *models.Session) error
    GetByID(ctx context.Context, id string) (*models.Session, error)
    GetByUserID(ctx context.Context, userID int64) ([]*models.Session, error)
    Update(ctx context.Context, session *models.Session) error
    Delete(ctx context.Context, id string) error
    DeleteByUserID(ctx context.Context, userID int64) error
    DeleteExpired(ctx context.Context) (int64, error)
    UpdateLastAccessed(ctx context.Context, id string) error
}
```

### Config Repository Interface

```go
package repositories

import (
    "context"
    "github.com/yourusername/tinyrsvp/internal/models"
)

type ConfigRepository interface {
    Get(ctx context.Context, key string) (*models.Config, error)
    Set(ctx context.Context, key, value string) error
    Delete(ctx context.Context, key string) error
    GetAll(ctx context.Context) ([]*models.Config, error)
    GetHMACSecret(ctx context.Context) ([]byte, error)
    SetHMACSecret(ctx context.Context, secret []byte) error
}
```

---

## Tasks

### Phase 0: Domain Models Creation (TDD)
- [x] Create internal/models/errors.go with domain error types
- [x] Write tests for NotFoundError
- [x] Write tests for ConflictError
- [x] Write tests for ValidationError
- [x] Write tests for OptimisticLockError
- [x] Create internal/models/user.go with User struct and methods
- [x] Write tests for User.IsAdmin() and User.IsEventManager()
- [x] Create internal/models/session.go with Session struct
- [x] Write tests for Session.IsExpired()
- [x] Create internal/models/config.go with Config struct
- [x] Run tests (should pass)

### Phase 1: User Repository (TDD)
- [x] Write test for Create user
- [x] Write test for GetByID
- [x] Write test for GetByEmail
- [x] Write test for GetByOIDCSubject
- [x] Write test for Update user
- [x] Write test for Delete user
- [x] Write test for List users with pagination
- [x] Write test for Count users
- [x] Write test for IsFirstUser
- [x] Write test for UpdateLastLogin
- [x] Write test for duplicate email error
- [x] Write test for not found error
- [x] Implement UserRepository
- [x] Run tests (should pass)

### Phase 2: Session Repository (TDD)
- [x] Write test for Create session
- [x] Write test for GetByID
- [x] Write test for GetByUserID
- [x] Write test for Update session
- [x] Write test for Delete session
- [x] Write test for DeleteByUserID
- [x] Write test for DeleteExpired
- [x] Write test for UpdateLastAccessed
- [x] Write test for expired session handling
- [x] Implement SessionRepository
- [x] Run tests (should pass)

### Phase 3: Config Repository (TDD)
- [x] Write test for Get config
- [x] Write test for Set config
- [x] Write test for Delete config
- [x] Write test for GetAll configs
- [x] Write test for GetHMACSecret
- [x] Write test for SetHMACSecret
- [x] Write test for auto-generate HMAC secret
- [x] Implement ConfigRepository
- [x] Run tests (should pass)

### Phase 4: Error Mapping (TDD)
- [x] Write test for NotFoundError mapping
- [x] Write test for ConflictError mapping
- [x] Write test for ValidationError mapping
- [x] Implement error mapping functions
- [x] Run tests (should pass)

### Phase 5: Transaction Support (TDD)
- [x] Transaction support available via db.WithTransaction()
- [x] Repositories can be used within transactions
- [x] Transaction tests exist in db_test.go

---

## Testing Requirements

### User Repository Tests

```go
func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    tests := []struct {
        name    string
        user    *models.User
        wantErr bool
        errType error
    }{
        {
            name: "valid user",
            user: &models.User{
                Email: "test@example.com",
                Name:  "Test User",
                Role:  models.RoleEventManager,
            },
            wantErr: false,
        },
        {
            name: "duplicate email",
            user: &models.User{
                Email: "test@example.com",
                Name:  "Another User",
                Role:  models.RoleAdmin,
            },
            wantErr: true,
            errType: &models.ConflictError{},
        },
        {
            name: "empty email",
            user: &models.User{
                Email: "",
                Name:  "Test User",
                Role:  models.RoleEventManager,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.Create(context.Background(), tt.user)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && tt.errType != nil {
                if !errors.As(err, &tt.errType) {
                    t.Errorf("Expected error type %T, got %T", tt.errType, err)
                }
            }
            
            if !tt.wantErr {
                if tt.user.ID == 0 {
                    t.Error("Expected ID to be set")
                }
                if tt.user.CreatedAt.IsZero() {
                    t.Error("Expected CreatedAt to be set")
                }
            }
        })
    }
}

func TestUserRepository_GetByEmail(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    user := &models.User{
        Email: "test@example.com",
        Name:  "Test User",
        Role:  models.RoleEventManager,
    }
    
    if err := repo.Create(context.Background(), user); err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    tests := []struct {
        name    string
        email   string
        wantErr bool
        errType error
    }{
        {
            name:    "existing email",
            email:   "test@example.com",
            wantErr: false,
        },
        {
            name:    "non-existing email",
            email:   "notfound@example.com",
            wantErr: true,
            errType: &models.NotFoundError{},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            found, err := repo.GetByEmail(context.Background(), tt.email)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr && tt.errType != nil {
                if !errors.As(err, &tt.errType) {
                    t.Errorf("Expected error type %T, got %T", tt.errType, err)
                }
            }
            
            if !tt.wantErr {
                if found.Email != tt.email {
                    t.Errorf("Expected email %s, got %s", tt.email, found.Email)
                }
            }
        })
    }
}

func TestUserRepository_IsFirstUser(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    isFirst, err := repo.IsFirstUser(context.Background())
    if err != nil {
        t.Fatalf("IsFirstUser() error = %v", err)
    }
    
    if !isFirst {
        t.Error("Expected true for first user check")
    }
    
    user := &models.User{
        Email: "test@example.com",
        Name:  "Test User",
        Role:  models.RoleAdmin,
    }
    
    if err := repo.Create(context.Background(), user); err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    isFirst, err = repo.IsFirstUser(context.Background())
    if err != nil {
        t.Fatalf("IsFirstUser() error = %v", err)
    }
    
    if isFirst {
        t.Error("Expected false after creating first user")
    }
}
```

### Session Repository Tests

```go
func TestSessionRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    userRepo := NewUserRepository(db)
    sessionRepo := NewSessionRepository(db)
    
    user := createTestUser(t, userRepo)
    
    session := &models.Session{
        ID:        "test-session-id",
        UserID:    user.ID,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    err := sessionRepo.Create(context.Background(), session)
    if err != nil {
        t.Fatalf("Create() error = %v", err)
    }
    
    if session.CreatedAt.IsZero() {
        t.Error("Expected CreatedAt to be set")
    }
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    userRepo := NewUserRepository(db)
    sessionRepo := NewSessionRepository(db)
    
    user := createTestUser(t, userRepo)
    
    expiredSession := &models.Session{
        ID:        "expired-session",
        UserID:    user.ID,
        ExpiresAt: time.Now().Add(-1 * time.Hour),
    }
    
    validSession := &models.Session{
        ID:        "valid-session",
        UserID:    user.ID,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    if err := sessionRepo.Create(context.Background(), expiredSession); err != nil {
        t.Fatalf("Failed to create expired session: %v", err)
    }
    
    if err := sessionRepo.Create(context.Background(), validSession); err != nil {
        t.Fatalf("Failed to create valid session: %v", err)
    }
    
    deleted, err := sessionRepo.DeleteExpired(context.Background())
    if err != nil {
        t.Fatalf("DeleteExpired() error = %v", err)
    }
    
    if deleted != 1 {
        t.Errorf("Expected 1 deleted session, got %d", deleted)
    }
    
    _, err = sessionRepo.GetByID(context.Background(), "expired-session")
    if !errors.As(err, &models.NotFoundError{}) {
        t.Error("Expected NotFoundError for expired session")
    }
    
    _, err = sessionRepo.GetByID(context.Background(), "valid-session")
    if err != nil {
        t.Error("Valid session should still exist")
    }
}
```

### Config Repository Tests

```go
func TestConfigRepository_GetHMACSecret(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewConfigRepository(db)
    
    secret, err := repo.GetHMACSecret(context.Background())
    if err != nil {
        t.Fatalf("GetHMACSecret() error = %v", err)
    }
    
    if len(secret) < 32 {
        t.Errorf("Expected secret >= 32 bytes, got %d", len(secret))
    }
    
    secret2, err := repo.GetHMACSecret(context.Background())
    if err != nil {
        t.Fatalf("GetHMACSecret() error = %v", err)
    }
    
    if !bytes.Equal(secret, secret2) {
        t.Error("Expected same secret on subsequent calls")
    }
}

func TestConfigRepository_SetAndGet(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewConfigRepository(db)
    
    err := repo.Set(context.Background(), "test_key", "test_value")
    if err != nil {
        t.Fatalf("Set() error = %v", err)
    }
    
    config, err := repo.Get(context.Background(), "test_key")
    if err != nil {
        t.Fatalf("Get() error = %v", err)
    }
    
    if config.Value != "test_value" {
        t.Errorf("Expected value 'test_value', got %q", config.Value)
    }
}
```

### Test Helpers

```go
func createTestUser(t *testing.T, repo UserRepository) *models.User {
    t.Helper()
    
    user := &models.User{
        Email: fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
        Name:  "Test User",
        Role:  models.RoleEventManager,
    }
    
    if err := repo.Create(context.Background(), user); err != nil {
        t.Fatalf("Failed to create test user: %v", err)
    }
    
    return user
}
```

---

## Dependencies

**Depends on:** 
- [00_STORY_go_module_setup.md](00_STORY_go_module_setup.md)
- [00_STORY_database_connection.md](00_STORY_database_connection.md)
- [00_STORY_database_migrations.md](00_STORY_database_migrations.md)

**Blocks:** 
- All domain-specific epics (Auth, Events, Invites, etc.)

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout (`go test -timeout 30s ./internal/db/repositories/...`)
- [x] Test coverage >= 85%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] All repositories implemented
- [x] Error mapping verified
- [x] Transaction support verified
- [x] Documentation complete
- [x] Changes committed to git

---

## Notes

### Repository Pattern Best Practices
- Keep repositories focused on data access only
- No business logic in repositories
- Return domain errors, not database errors
- Use context for cancellation and timeouts
- Make repositories interface-based for testability

### Error Handling
- Map `sql.ErrNoRows` to `NotFoundError`
- Map unique constraint violations to `ConflictError`
- Preserve error context with wrapping
- Never expose database-specific errors to callers

### Testing Strategy
- Test happy paths and error paths
- Test edge cases (empty results, duplicates, etc.)
- Use table-driven tests for multiple scenarios
- Test with real database (in-memory SQLite)
- Verify error types, not just error presence

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/07_DATABASE_LLD.md](../lld/07_DATABASE_LLD.md) - Section 4 (Interfaces)
- **Repository Pattern:** https://martinfowler.com/eaaCatalog/repository.html
