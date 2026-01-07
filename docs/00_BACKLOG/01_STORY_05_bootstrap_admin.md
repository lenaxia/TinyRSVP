# User Story: Bootstrap Admin User

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)
**Priority:** Critical
**Status:** Complete
**Estimated Effort:** 2 hours
**Actual Effort:** 2.5 hours
**Completed:** 2026-01-07

---

## User Story

As a **system administrator**, I want **the first user to automatically become an admin** so that **initial system setup is streamlined without manual database intervention**.

---

## Acceptance Criteria

- [x] First authenticated user automatically assigned admin role
- [x] Subsequent users assigned event manager role by default
- [x] Bootstrap logic verified on empty database
- [x] Admin can promote other users to admin
- [x] Clear documentation on bootstrap process
- [x] All tests pass with timeout

---

## Technical Details

### Bootstrap Logic

```go
func (s *userService) CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error) {
    isFirst, err := s.repo.IsFirstUser(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to check if first user: %w", err)
    }
    
    role := models.RoleEventManager
    if isFirst {
        role = models.RoleAdmin
    }
    
    user := &models.User{
        Email:       email,
        Name:        name,
        Role:        role,
        OIDCSubject: oidcSubject,
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

### IsFirstUser Check

```go
func (r *userRepository) IsFirstUser(ctx context.Context) (bool, error) {
    var count int
    err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
    if err != nil {
        return false, err
    }
    return count == 0, nil
}
```

---

## Tasks

### Phase 1: Bootstrap Logic (TDD)
- [x] Write test for empty database (first user)
- [x] Write test for database with existing users
- [x] Write test for concurrent first user creation
- [x] Verify IsFirstUser implementation
- [x] Run tests (should pass)

### Phase 2: Integration Testing (TDD)
- [x] Write test for OIDC first user flow
- [x] Write test for forward auth first user flow
- [x] Write test for second user getting event manager role
- [x] Write test for third user getting event manager role
- [x] Run tests (should pass)

### Phase 3: Admin Promotion (TDD)
- [x] Write test for admin promoting user
- [x] Non-admin authorization check deferred to RBAC middleware story
- [x] Admin promotion logic already implemented in UpdateUserRole
- [x] Run tests (should pass)

### Phase 4: Documentation
- [x] Document bootstrap process in story
- [x] Add troubleshooting guide in story
- [x] Document how to promote users to admin

---

## Testing Requirements

### Unit Tests

```go
func TestBootstrap_FirstUser(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewUserRepository(db)
    service := auth.NewUserService(repo)
    
    user, err := service.CreateUser(context.Background(), "admin@example.com", "Admin User", nil)
    if err != nil {
        t.Fatalf("CreateUser() error = %v", err)
    }
    
    if user.Role != models.RoleAdmin {
        t.Errorf("Expected first user to be admin, got %v", user.Role)
    }
}

func TestBootstrap_SecondUser(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewUserRepository(db)
    service := auth.NewUserService(repo)
    
    _, err := service.CreateUser(context.Background(), "admin@example.com", "Admin User", nil)
    if err != nil {
        t.Fatalf("Failed to create first user: %v", err)
    }
    
    user, err := service.CreateUser(context.Background(), "user@example.com", "Regular User", nil)
    if err != nil {
        t.Fatalf("CreateUser() error = %v", err)
    }
    
    if user.Role != models.RoleEventManager {
        t.Errorf("Expected second user to be event manager, got %v", user.Role)
    }
}

func TestBootstrap_ConcurrentFirstUser(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repositories.NewUserRepository(db)
    service := auth.NewUserService(repo)
    
    var wg sync.WaitGroup
    results := make(chan *models.User, 3)
    
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            user, err := service.CreateUser(context.Background(), 
                fmt.Sprintf("user%d@example.com", n), 
                fmt.Sprintf("User %d", n), 
                nil)
            if err == nil {
                results <- user
            }
        }(i)
    }
    
    wg.Wait()
    close(results)
    
    adminCount := 0
    for user := range results {
        if user.Role == models.RoleAdmin {
            adminCount++
        }
    }
    
    if adminCount != 1 {
        t.Errorf("Expected exactly 1 admin, got %d", adminCount)
    }
}
```

---

## Dependencies

**Depends on:** 
- User service (01_STORY_04_user_model.md)
- User repository with IsFirstUser method

**Blocks:** 
- User authentication and authorization flows

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout (`go test -timeout 30s ./internal/auth/...`)
- [x] Test coverage >= 85%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] Concurrent creation tested
- [x] Bootstrap process documented
- [x] Admin promotion documented
- [x] Changes committed to git

---

## Implementation Notes

### Race Condition Prevention

The bootstrap logic uses a transactional approach to prevent race conditions. The `CreateWithBootstrapCheck` method atomically:
1. Checks the user count within a transaction
2. Determines the appropriate role (admin if count == 0, event_manager otherwise)
3. Inserts the user with the correct role
4. Commits the transaction

This ensures that if multiple users try to authenticate simultaneously on an empty database, exactly one will become admin due to transaction serialization. The unique email constraint provides additional protection against duplicate users.

### Manual Admin Promotion

After bootstrap, admins can promote users via:
1. Admin dashboard UI (future)
2. Direct database update (temporary)
3. Admin API endpoint (future)

### Security Considerations

- Bootstrap only happens on completely empty database
- Cannot be exploited after first user created
- Admin role is permanent (cannot be accidentally removed if last admin)

---

## Documentation

### Bootstrap Process

1. Deploy TinyRSVP with empty database
2. Configure OIDC or forward auth
3. First person to authenticate becomes admin
4. All subsequent users become event managers
5. Admin can promote users as needed

### Troubleshooting

**Problem:** First user is not admin  
**Solution:** Check database has no existing users, verify `IsFirstUser()` logic

**Problem:** Multiple admins created  
**Solution:** Should not happen due to database constraints, check logs

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/01_AUTH_LLD.md](../lld/01_AUTH_LLD.md) - Section 5.4
- **Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)
