# User Story: User Management CRUD

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)
**Priority:** Medium
**Status:** Complete
**Estimated Effort:** 6 hours
**Actual Effort:** 3 hours
**Completed:** 2026-01-07

---

## User Story

As an **admin**, I want **to manage users through CRUD operations** so that **I can add, view, update, and remove users from the system**.

---

## Acceptance Criteria

- [x] Admin can list all users
- [x] Admin can view user details
- [x] Admin can promote users to admin
- [x] Admin can demote admins to event manager
- [x] Admin cannot demote last admin
- [x] Admin can delete users
- [x] Admin cannot delete last admin
- [x] Deleting user cascades to sessions
- [x] Pagination working for user list
- [x] All tests pass with timeout

---

## Technical Details

### HTTP Endpoints

```
GET    /api/users           - List users (paginated)
GET    /api/users/:id       - Get user details
PATCH  /api/users/:id/role  - Update user role
DELETE /api/users/:id       - Delete user
```

### Request/Response Models

```go
type ListUsersResponse struct {
    Users  []*UserDTO `json:"users"`
    Total  int        `json:"total"`
    Limit  int        `json:"limit"`
    Offset int        `json:"offset"`
}

type UserDTO struct {
    ID          int64      `json:"id"`
    Email       string     `json:"email"`
    Name        string     `json:"name"`
    Role        string     `json:"role"`
    CreatedAt   time.Time  `json:"created_at"`
    LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

type UpdateRoleRequest struct {
    Role string `json:"role"`
}
```

---

## Tasks

### Phase 1: List Users (TDD)
- [x] Write test for list users endpoint
- [x] Write test for pagination
- [x] Write test for unauthorized access
- [x] Write test for admin-only access
- [x] Implement list users handler
- [x] Run tests (should pass)

### Phase 2: Get User (TDD)
- [x] Write test for get user by ID
- [x] Write test for non-existent user
- [x] Write test for unauthorized access
- [x] Implement get user handler
- [x] Run tests (should pass)

### Phase 3: Update Role (TDD)
- [x] Write test for promoting user to admin
- [x] Write test for demoting admin to event manager
- [x] Write test for preventing last admin demotion
- [x] Write test for invalid role
- [x] Write test for non-admin attempting update
- [x] Implement update role handler
- [x] Run tests (should pass)

### Phase 4: Delete User (TDD)
- [x] Write test for user deletion
- [x] Write test for cascade deletion of sessions
- [x] Write test for preventing last admin deletion
- [x] Write test for non-existent user
- [x] Write test for non-admin attempting deletion
- [x] Implement delete user handler
- [x] Run tests (should pass)

### Phase 5: Integration
- [ ] Wire handlers into HTTP router (deferred to Story 07)
- [ ] Add RBAC middleware (Story 07)
- [ ] Test full user management flow (Story 07)
- [x] Document API endpoints

---

## Testing Requirements

### Unit Tests

```go
func TestListUsers(t *testing.T) {
    tests := []struct {
        name       string
        limit      int
        offset     int
        user       *models.User
        wantStatus int
        wantCount  int
    }{
        {
            name:       "admin can list users",
            limit:      10,
            offset:     0,
            user:       &models.User{Role: models.RoleAdmin},
            wantStatus: http.StatusOK,
            wantCount:  3,
        },
        {
            name:       "event manager cannot list users",
            user:       &models.User{Role: models.RoleEventManager},
            wantStatus: http.StatusForbidden,
        },
        {
            name:       "pagination works",
            limit:      2,
            offset:     1,
            user:       &models.User{Role: models.RoleAdmin},
            wantStatus: http.StatusOK,
            wantCount:  2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

func TestUpdateUserRole(t *testing.T) {
    tests := []struct {
        name       string
        targetUser *models.User
        newRole    string
        actor      *models.User
        wantStatus int
        wantErr    string
    }{
        {
            name:       "admin can promote user",
            targetUser: &models.User{Role: models.RoleEventManager},
            newRole:    "admin",
            actor:      &models.User{Role: models.RoleAdmin},
            wantStatus: http.StatusOK,
        },
        {
            name:       "cannot demote last admin",
            targetUser: &models.User{Role: models.RoleAdmin},
            newRole:    "event_manager",
            actor:      &models.User{Role: models.RoleAdmin},
            wantStatus: http.StatusConflict,
            wantErr:    "cannot demote last admin",
        },
        {
            name:       "non-admin cannot change roles",
            targetUser: &models.User{Role: models.RoleEventManager},
            newRole:    "admin",
            actor:      &models.User{Role: models.RoleEventManager},
            wantStatus: http.StatusForbidden,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

func TestDeleteUser(t *testing.T) {
    tests := []struct {
        name       string
        targetUser *models.User
        actor      *models.User
        wantStatus int
        wantErr    string
    }{
        {
            name:       "admin can delete user",
            targetUser: &models.User{Role: models.RoleEventManager},
            actor:      &models.User{Role: models.RoleAdmin},
            wantStatus: http.StatusNoContent,
        },
        {
            name:       "cannot delete last admin",
            targetUser: &models.User{Role: models.RoleAdmin},
            actor:      &models.User{Role: models.RoleAdmin},
            wantStatus: http.StatusConflict,
            wantErr:    "cannot delete last admin",
        },
        {
            name:       "non-admin cannot delete users",
            targetUser: &models.User{Role: models.RoleEventManager},
            actor:      &models.User{Role: models.RoleEventManager},
            wantStatus: http.StatusForbidden,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

## Dependencies

**Depends on:** 
- User service (01_STORY_04_user_model.md)
- RBAC middleware (01_STORY_07_rbac_middleware.md)
- Permission checks (01_STORY_08_permission_checks.md)

**Blocks:** 
- Admin dashboard functionality

---

## Definition of Done

- [x] All acceptance criteria met
- [x] All tasks completed
- [x] All tests pass with timeout (`go test -timeout 30s ./internal/handlers/...`)
- [x] Test coverage >= 85%
- [x] Code formatted with `go fmt`
- [x] No errors from `go vet`
- [x] API documented
- [x] Last admin protection verified
- [x] Cascade deletion tested
- [ ] Changes committed to git

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/01_AUTH_LLD.md](../lld/01_AUTH_LLD.md)
- **Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)
