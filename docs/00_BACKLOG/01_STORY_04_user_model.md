# User Story: User Model and Service

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)  
**Priority:** Critical  
**Status:** Partially Complete (Model exists, Service needed)  
**Estimated Effort:** 4 hours  
**Actual Effort:** TBD  
**Completed:** TBD

---

## User Story

As a **developer**, I want **a user service that manages user creation and role assignment** so that **users can be authenticated and authorized throughout the application**.

---

## Acceptance Criteria

- [ ] User model validated and complete
- [ ] UserService interface implemented
- [ ] User creation with email and name
- [ ] Get or create user functionality
- [ ] OIDC subject linking
- [ ] User retrieval by ID and email
- [ ] User role management
- [ ] Last login timestamp tracking
- [ ] User repository fully functional
- [ ] All tests pass with timeout

---

## Technical Details

### User Model (Already Exists)

```go
type User struct {
    ID          int64
    Email       string
    Name        string
    Role        UserRole
    OIDCSubject *string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    LastLoginAt *time.Time
}

type UserRole string

const (
    RoleAdmin        UserRole = "admin"
    RoleEventManager UserRole = "event_manager"
)
```

### UserService Interface

```go
type UserService interface {
    CreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
    GetOrCreateUser(ctx context.Context, email, name string, oidcSubject *string) (*models.User, error)
    GetUserByID(ctx context.Context, id int64) (*models.User, error)
    GetUserByEmail(ctx context.Context, email string) (*models.User, error)
    UpdateUser(ctx context.Context, user *models.User) error
    UpdateUserRole(ctx context.Context, userID int64, role models.UserRole) error
    DeleteUser(ctx context.Context, id int64) error
    ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
}
```

### User Repository Interface (Already Exists)

```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id int64) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    GetByOIDCSubject(ctx context.Context, subject string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    UpdateLastLogin(ctx context.Context, userID int64) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, limit, offset int) ([]*models.User, error)
    IsFirstUser(ctx context.Context) (bool, error)
}
```

### Bootstrap Admin Logic

The first user to authenticate automatically becomes an admin. This is handled by checking if the database has any users before assigning a role.

---

## Tasks

### Phase 1: Validate User Model (TDD)
- [ ] Review existing user model
- [ ] Write test for user validation
- [ ] Write test for role constants
- [ ] Ensure model matches LLD specification
- [ ] Run tests (should pass)

### Phase 2: UserService Implementation (TDD)
- [ ] Write test for CreateUser
- [ ] Write test for first user becoming admin
- [ ] Write test for subsequent users becoming event managers
- [ ] Write test for GetOrCreateUser with existing user
- [ ] Write test for GetOrCreateUser with new user
- [ ] Write test for OIDC subject linking
- [ ] Implement userService struct
- [ ] Run tests (should pass)

### Phase 3: User Retrieval (TDD)
- [ ] Write test for GetUserByID
- [ ] Write test for GetUserByEmail
- [ ] Write test for non-existent user
- [ ] Write test for ListUsers with pagination
- [ ] Implement retrieval methods
- [ ] Run tests (should pass)

### Phase 4: User Updates (TDD)
- [ ] Write test for UpdateUser
- [ ] Write test for UpdateUserRole
- [ ] Write test for last login tracking
- [ ] Write test for preventing role downgrade of last admin
- [ ] Implement update methods
- [ ] Run tests (should pass)

### Phase 5: User Deletion (TDD)
- [ ] Write test for DeleteUser
- [ ] Write test for preventing deletion of last admin
- [ ] Write test for cascade deletion of sessions
- [ ] Implement deletion method
- [ ] Run tests (should pass)

### Phase 6: Integration
- [ ] Wire UserService into authentication flows
- [ ] Test with OIDC authenticator
- [ ] Test with forward authenticator
- [ ] Verify first user becomes admin
- [ ] Document user service usage

---

## Testing Requirements

### Unit Tests

```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name        string
        email       string
        userName    string
        oidcSubject *string
        setupDB     func(*testing.T, repositories.UserRepository)
        wantRole    models.UserRole
        wantErr     bool
    }{
        {
            name:     "first user becomes admin",
            email:    "first@example.com",
            userName: "First User",
            setupDB:  func(t *testing.T, repo repositories.UserRepository) {},
            wantRole: models.RoleAdmin,
            wantErr:  false,
        },
        {
            name:     "second user becomes event manager",
            email:    "second@example.com",
            userName: "Second User",
            setupDB: func(t *testing.T, repo repositories.UserRepository) {
                repo.Create(context.Background(), &models.User{
                    Email: "first@example.com",
                    Name:  "First User",
                    Role:  models.RoleAdmin,
                })
            },
            wantRole: models.RoleEventManager,
            wantErr:  false,
        },
        {
            name:     "invalid email",
            email:    "not-an-email",
            userName: "Test User",
            wantErr:  true,
        },
        {
            name:     "empty email",
            email:    "",
            userName: "Test User",
            wantErr:  true,
        },
        {
            name:     "empty name",
            email:    "test@example.com",
            userName: "",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db := setupTestDB(t)
            defer db.Close()
            
            repo := repositories.NewUserRepository(db)
            if tt.setupDB != nil {
                tt.setupDB(t, repo)
            }
            
            service := auth.NewUserService(repo)
            user, err := service.CreateUser(context.Background(), tt.email, tt.userName, tt.oidcSubject)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr {
                if user.Email != tt.email {
                    t.Errorf("Expected email %q, got %q", tt.email, user.Email)
                }
                
                if user.Name != tt.userName {
                    t.Errorf("Expected name %q, got %q", tt.userName, user.Name)
                }
                
                if user.Role != tt.wantRole {
                    t.Errorf("Expected role %v, got %v", tt.wantRole, user.Role)
                }
                
                if user.ID == 0 {
                    t.Error("Expected non-zero user ID")
                }
            }
        })
    }
}

func TestUserService_GetOrCreateUser(t *testing.T) {
    tests := []struct {
        name          string
        email         string
        userName      string
        oidcSubject   *string
        setupDB       func(*testing.T, repositories.UserRepository)
        wantNew       bool
        wantRole      models.UserRole
    }{
        {
            name:        "creates new user",
            email:       "new@example.com",
            userName:    "New User",
            oidcSubject: stringPtr("oidc123"),
            setupDB:     func(t *testing.T, repo repositories.UserRepository) {},
            wantNew:     true,
            wantRole:    models.RoleAdmin,
        },
        {
            name:        "returns existing user by email",
            email:       "existing@example.com",
            userName:    "Existing User",
            oidcSubject: nil,
            setupDB: func(t *testing.T, repo repositories.UserRepository) {
                repo.Create(context.Background(), &models.User{
                    Email: "existing@example.com",
                    Name:  "Existing User",
                    Role:  models.RoleAdmin,
                })
            },
            wantNew:  false,
            wantRole: models.RoleAdmin,
        },
        {
            name:        "returns existing user by OIDC subject",
            email:       "user@example.com",
            userName:    "User",
            oidcSubject: stringPtr("oidc456"),
            setupDB: func(t *testing.T, repo repositories.UserRepository) {
                subject := "oidc456"
                repo.Create(context.Background(), &models.User{
                    Email:       "user@example.com",
                    Name:        "User",
                    Role:        models.RoleAdmin,
                    OIDCSubject: &subject,
                })
            },
            wantNew:  false,
            wantRole: models.RoleAdmin,
        },
        {
            name:        "links OIDC subject to existing user",
            email:       "user@example.com",
            userName:    "User",
            oidcSubject: stringPtr("oidc789"),
            setupDB: func(t *testing.T, repo repositories.UserRepository) {
                repo.Create(context.Background(), &models.User{
                    Email: "user@example.com",
                    Name:  "User",
                    Role:  models.RoleAdmin,
                })
            },
            wantNew:  false,
            wantRole: models.RoleAdmin,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db := setupTestDB(t)
            defer db.Close()
            
            repo := repositories.NewUserRepository(db)
            if tt.setupDB != nil {
                tt.setupDB(t, repo)
            }
            
            service := auth.NewUserService(repo)
            user, err := service.GetOrCreateUser(context.Background(), tt.email, tt.userName, tt.oidcSubject)
            
            if err != nil {
                t.Fatalf("GetOrCreateUser() error = %v", err)
            }
            
            if user.Email != tt.email {
                t.Errorf("Expected email %q, got %q", tt.email, user.Email)
            }
            
            if user.Role != tt.wantRole {
                t.Errorf("Expected role %v, got %v", tt.wantRole, user.Role)
            }
            
            if tt.oidcSubject != nil && user.OIDCSubject == nil {
                t.Error("Expected OIDC subject to be set")
            }
        })
    }
}

func TestUserService_UpdateUserRole(t *testing.T) {
    tests := []struct {
        name     string
        setupDB  func(*testing.T, repositories.UserRepository) int64
        newRole  models.UserRole
        wantErr  bool
    }{
        {
            name: "promote to admin",
            setupDB: func(t *testing.T, repo repositories.UserRepository) int64 {
                user := &models.User{
                    Email: "user@example.com",
                    Name:  "User",
                    Role:  models.RoleEventManager,
                }
                repo.Create(context.Background(), user)
                return user.ID
            },
            newRole: models.RoleAdmin,
            wantErr: false,
        },
        {
            name: "demote to event manager",
            setupDB: func(t *testing.T, repo repositories.UserRepository) int64 {
                repo.Create(context.Background(), &models.User{
                    Email: "admin1@example.com",
                    Name:  "Admin 1",
                    Role:  models.RoleAdmin,
                })
                user := &models.User{
                    Email: "admin2@example.com",
                    Name:  "Admin 2",
                    Role:  models.RoleAdmin,
                }
                repo.Create(context.Background(), user)
                return user.ID
            },
            newRole: models.RoleEventManager,
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db := setupTestDB(t)
            defer db.Close()
            
            repo := repositories.NewUserRepository(db)
            userID := tt.setupDB(t, repo)
            
            service := auth.NewUserService(repo)
            err := service.UpdateUserRole(context.Background(), userID, tt.newRole)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("UpdateUserRole() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr {
                user, err := repo.GetByID(context.Background(), userID)
                if err != nil {
                    t.Fatalf("Failed to retrieve updated user: %v", err)
                }
                
                if user.Role != tt.newRole {
                    t.Errorf("Expected role %v, got %v", tt.newRole, user.Role)
                }
            }
        })
    }
}
```

---

## Dependencies

**Depends on:** 
- User model (internal/models/user.go - already exists)
- User repository (internal/db/repositories/user_repository.go - already exists)
- Database connection (00_STORY_03_database_connection.md)

**Blocks:** 
- OIDC authentication (01_STORY_01_oidc_integration.md)
- Forward auth (01_STORY_02_forward_auth.md)
- Bootstrap admin (01_STORY_05_bootstrap_admin.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/auth/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] User service fully functional
- [ ] Bootstrap admin logic working
- [ ] OIDC subject linking tested
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Implementation Notes

### First User Bootstrap

The `IsFirstUser()` repository method checks if any users exist in the database. If not, the new user is assigned the admin role.

### OIDC Subject Linking

When a user authenticates via OIDC:
1. Try to find user by OIDC subject
2. If not found, try to find by email
3. If found by email but no OIDC subject, link the OIDC subject
4. If not found at all, create new user with OIDC subject

### Last Login Tracking

The `UpdateLastLogin()` method is called by the authenticator after successful authentication to track when users last accessed the system.

### Role Management

- Admin: Full system access, can manage users and system configuration
- Event Manager: Can create and manage their own events

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/01_AUTH_LLD.md](../lld/01_AUTH_LLD.md) - Section 5.4
- **User Model:** [internal/models/user.go](../../internal/models/user.go)
- **User Repository:** [internal/db/repositories/user_repository.go](../../internal/db/repositories/user_repository.go)
