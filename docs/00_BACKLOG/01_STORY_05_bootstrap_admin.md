# User Story: Bootstrap Admin User

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 2 hours  
**Actual Effort:** TBD  
**Completed:** TBD

---

## User Story

As a **system administrator**, I want **the first user to automatically become an admin** so that **initial system setup is streamlined without manual database intervention**.

---

## Acceptance Criteria

- [ ] First authenticated user automatically assigned admin role
- [ ] Subsequent users assigned event manager role by default
- [ ] Bootstrap logic verified on empty database
- [ ] Admin can promote other users to admin
- [ ] Clear documentation on bootstrap process
- [ ] All tests pass with timeout

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
- [ ] Write test for empty database (first user)
- [ ] Write test for database with existing users
- [ ] Write test for concurrent first user creation
- [ ] Verify IsFirstUser implementation
- [ ] Run tests (should pass)

### Phase 2: Integration Testing (TDD)
- [ ] Write test for OIDC first user flow
- [ ] Write test for forward auth first user flow
- [ ] Write test for second user getting event manager role
- [ ] Write test for third user getting event manager role
- [ ] Run tests (should pass)

### Phase 3: Admin Promotion (TDD)
- [ ] Write test for admin promoting user
- [ ] Write test for non-admin attempting promotion
- [ ] Implement admin promotion logic
- [ ] Run tests (should pass)

### Phase 4: Documentation
- [ ] Document bootstrap process in README
- [ ] Add troubleshooting guide
- [ ] Document how to promote users to admin

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

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/auth/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] Concurrent creation tested
- [ ] Bootstrap process documented
- [ ] Admin promotion documented
- [ ] Changes committed to git

---

## Implementation Notes

### Race Condition Prevention

The bootstrap logic relies on database constraints (unique email) to prevent race conditions. If multiple users try to authenticate simultaneously on an empty database, only one will successfully become admin due to transaction isolation.

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
