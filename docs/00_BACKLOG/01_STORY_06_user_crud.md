# User Story: User Management CRUD

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)  
**Priority:** Medium  
**Status:** Not Started  
**Estimated Effort:** 6 hours  
**Actual Effort:** TBD  
**Completed:** TBD

---

## User Story

As an **admin**, I want **to manage users through CRUD operations** so that **I can add, view, update, and remove users from the system**.

---

## Acceptance Criteria

- [ ] Admin can list all users
- [ ] Admin can view user details
- [ ] Admin can promote users to admin
- [ ] Admin can demote admins to event manager
- [ ] Admin cannot demote last admin
- [ ] Admin can delete users
- [ ] Admin cannot delete last admin
- [ ] Deleting user cascades to sessions
- [ ] Pagination working for user list
- [ ] All tests pass with timeout

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
- [ ] Write test for list users endpoint
- [ ] Write test for pagination
- [ ] Write test for unauthorized access
- [ ] Write test for admin-only access
- [ ] Implement list users handler
- [ ] Run tests (should pass)

### Phase 2: Get User (TDD)
- [ ] Write test for get user by ID
- [ ] Write test for non-existent user
- [ ] Write test for unauthorized access
- [ ] Implement get user handler
- [ ] Run tests (should pass)

### Phase 3: Update Role (TDD)
- [ ] Write test for promoting user to admin
- [ ] Write test for demoting admin to event manager
- [ ] Write test for preventing last admin demotion
- [ ] Write test for invalid role
- [ ] Write test for non-admin attempting update
- [ ] Implement update role handler
- [ ] Run tests (should pass)

### Phase 4: Delete User (TDD)
- [ ] Write test for user deletion
- [ ] Write test for cascade deletion of sessions
- [ ] Write test for preventing last admin deletion
- [ ] Write test for non-existent user
- [ ] Write test for non-admin attempting deletion
- [ ] Implement delete user handler
- [ ] Run tests (should pass)

### Phase 5: Integration
- [ ] Wire handlers into HTTP router
- [ ] Add RBAC middleware
- [ ] Test full user management flow
- [ ] Document API endpoints

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

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/handlers/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] API documented
- [ ] Last admin protection verified
- [ ] Cascade deletion tested
- [ ] Changes committed to git

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/01_AUTH_LLD.md](../lld/01_AUTH_LLD.md)
- **Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)
