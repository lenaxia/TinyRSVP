# User Story: Permission Checking Service

**Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)  
**Priority:** Critical  
**Status:** Not Started  
**Estimated Effort:** 5 hours  
**Actual Effort:** TBD  
**Completed:** TBD

---

## User Story

As a **developer**, I want **a centralized permission checking service** so that **authorization logic is consistent and testable throughout the application**.

---

## Acceptance Criteria

- [ ] Authorization checker interface implemented
- [ ] Admin role check functional
- [ ] Event manager role check functional
- [ ] Event ownership check functional
- [ ] Event permission methods working
- [ ] User management permission methods working
- [ ] System configuration permission methods working
- [ ] All tests pass with timeout

---

## Technical Details

### AuthorizationChecker Interface

```go
type AuthorizationChecker interface {
    CanCreateEvent(ctx context.Context, user *models.User) bool
    CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool
    CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool
    CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool
    CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool
    CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool
    CanManageUsers(ctx context.Context, user *models.User) bool
    CanConfigureSystem(ctx context.Context, user *models.User) bool
    IsAdmin(user *models.User) bool
    IsEventManager(user *models.User) bool
}
```

### Permission Rules

**Admin:**
- Can do everything
- Full system access
- Can manage all events
- Can manage all users

**Event Manager:**
- Can create events
- Can edit own events
- Can delete own draft/published events
- Can manage invites for own events
- Can view RSVPs for own events
- Cannot manage users
- Cannot configure system

---

## Tasks

### Phase 1: Role Checks (TDD)
- [ ] Write test for IsAdmin with admin user
- [ ] Write test for IsAdmin with non-admin user
- [ ] Write test for IsEventManager with admin
- [ ] Write test for IsEventManager with event manager
- [ ] Write test for IsEventManager with nil user
- [ ] Implement role check methods
- [ ] Run tests (should pass)

### Phase 2: Event Permissions (TDD)
- [ ] Write test for CanCreateEvent
- [ ] Write test for CanEditEvent (owner)
- [ ] Write test for CanEditEvent (non-owner)
- [ ] Write test for CanEditEvent (admin)
- [ ] Write test for CanDeleteEvent with different statuses
- [ ] Write test for CanViewEvent
- [ ] Implement event permission methods
- [ ] Run tests (should pass)

### Phase 3: Invite Permissions (TDD)
- [ ] Write test for CanManageInvites (owner)
- [ ] Write test for CanManageInvites (non-owner)
- [ ] Write test for CanManageInvites (admin)
- [ ] Implement invite permission methods
- [ ] Run tests (should pass)

### Phase 4: RSVP Permissions (TDD)
- [ ] Write test for CanViewRSVPs (owner)
- [ ] Write test for CanViewRSVPs (non-owner)
- [ ] Write test for CanViewRSVPs (admin)
- [ ] Implement RSVP permission methods
- [ ] Run tests (should pass)

### Phase 5: System Permissions (TDD)
- [ ] Write test for CanManageUsers
- [ ] Write test for CanConfigureSystem
- [ ] Implement system permission methods
- [ ] Run tests (should pass)

### Phase 6: Integration
- [ ] Integrate into handlers
- [ ] Test permission enforcement
- [ ] Document permission matrix
- [ ] Create permission reference guide

---

## Testing Requirements

### Unit Tests

```go
func TestAuthorizationChecker_IsAdmin(t *testing.T) {
    tests := []struct {
        name string
        user *models.User
        want bool
    }{
        {
            name: "admin user",
            user: &models.User{Role: models.RoleAdmin},
            want: true,
        },
        {
            name: "event manager user",
            user: &models.User{Role: models.RoleEventManager},
            want: false,
        },
        {
            name: "nil user",
            user: nil,
            want: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            checker := auth.NewAuthorizationChecker()
            
            got := checker.IsAdmin(tt.user)
            if got != tt.want {
                t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestAuthorizationChecker_IsEventManager(t *testing.T) {
    tests := []struct {
        name string
        user *models.User
        want bool
    }{
        {
            name: "admin user is event manager",
            user: &models.User{Role: models.RoleAdmin},
            want: true,
        },
        {
            name: "event manager user",
            user: &models.User{Role: models.RoleEventManager},
            want: true,
        },
        {
            name: "nil user",
            user: nil,
            want: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            checker := auth.NewAuthorizationChecker()
            
            got := checker.IsEventManager(tt.user)
            if got != tt.want {
                t.Errorf("IsEventManager() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestAuthorizationChecker_CanEditEvent(t *testing.T) {
    tests := []struct {
        name  string
        user  *models.User
        event *models.Event
        want  bool
    }{
        {
            name:  "admin can edit any event",
            user:  &models.User{ID: 1, Role: models.RoleAdmin},
            event: &models.Event{ID: 100, CreatedBy: 2},
            want:  true,
        },
        {
            name:  "owner can edit own event",
            user:  &models.User{ID: 1, Role: models.RoleEventManager},
            event: &models.Event{ID: 100, CreatedBy: 1},
            want:  true,
        },
        {
            name:  "non-owner cannot edit event",
            user:  &models.User{ID: 1, Role: models.RoleEventManager},
            event: &models.Event{ID: 100, CreatedBy: 2},
            want:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            checker := auth.NewAuthorizationChecker()
            
            got := checker.CanEditEvent(context.Background(), tt.user, tt.event)
            if got != tt.want {
                t.Errorf("CanEditEvent() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestAuthorizationChecker_CanDeleteEvent(t *testing.T) {
    tests := []struct {
        name  string
        user  *models.User
        event *models.Event
        want  bool
    }{
        {
            name:  "admin can delete any event",
            user:  &models.User{ID: 1, Role: models.RoleAdmin},
            event: &models.Event{ID: 100, CreatedBy: 2, Status: models.EventStatusPublished},
            want:  true,
        },
        {
            name:  "owner can delete draft event",
            user:  &models.User{ID: 1, Role: models.RoleEventManager},
            event: &models.Event{ID: 100, CreatedBy: 1, Status: models.EventStatusDraft},
            want:  true,
        },
        {
            name:  "owner can delete published event",
            user:  &models.User{ID: 1, Role: models.RoleEventManager},
            event: &models.Event{ID: 100, CreatedBy: 1, Status: models.EventStatusPublished},
            want:  true,
        },
        {
            name:  "owner cannot delete completed event",
            user:  &models.User{ID: 1, Role: models.RoleEventManager},
            event: &models.Event{ID: 100, CreatedBy: 1, Status: models.EventStatusCompleted},
            want:  false,
        },
        {
            name:  "non-owner cannot delete event",
            user:  &models.User{ID: 1, Role: models.RoleEventManager},
            event: &models.Event{ID: 100, CreatedBy: 2, Status: models.EventStatusDraft},
            want:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            checker := auth.NewAuthorizationChecker()
            
            got := checker.CanDeleteEvent(context.Background(), tt.user, tt.event)
            if got != tt.want {
                t.Errorf("CanDeleteEvent() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestAuthorizationChecker_CanManageUsers(t *testing.T) {
    tests := []struct {
        name string
        user *models.User
        want bool
    }{
        {
            name: "admin can manage users",
            user: &models.User{Role: models.RoleAdmin},
            want: true,
        },
        {
            name: "event manager cannot manage users",
            user: &models.User{Role: models.RoleEventManager},
            want: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            checker := auth.NewAuthorizationChecker()
            
            got := checker.CanManageUsers(context.Background(), tt.user)
            if got != tt.want {
                t.Errorf("CanManageUsers() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## Dependencies

**Depends on:** 
- User model (01_STORY_04_user_model.md)
- Event model (02_EPIC_events.md)

**Blocks:** 
- All authorization decisions throughout application
- RBAC middleware (01_STORY_07_rbac_middleware.md)
- Event handlers
- User management handlers

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] All tests pass with timeout (`go test -timeout 30s ./internal/auth/...`)
- [ ] Test coverage >= 85%
- [ ] Code formatted with `go fmt`
- [ ] No errors from `go vet`
- [ ] Permission matrix documented
- [ ] All scenarios tested
- [ ] Edge cases covered
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Implementation Notes

### Permission Matrix

| Action | Admin | Event Manager (Owner) | Event Manager (Non-Owner) |
|--------|-------|----------------------|---------------------------|
| Create Event | ✅ | ✅ | ✅ |
| Edit Event | ✅ | ✅ | ❌ |
| Delete Event (Draft/Published) | ✅ | ✅ | ❌ |
| Delete Event (Completed/Cancelled) | ✅ | ❌ | ❌ |
| View Event | ✅ | ✅ | ✅ |
| Manage Invites | ✅ | ✅ | ❌ |
| View RSVPs | ✅ | ✅ | ❌ |
| Manage Users | ✅ | ❌ | ❌ |
| Configure System | ✅ | ❌ | ❌ |

### Event Status Considerations

Event deletion rules depend on status:
- **Draft:** Owner or admin can delete
- **Published:** Owner or admin can delete
- **Completed:** Only admin can delete
- **Cancelled:** Only admin can delete

### Context-Aware Permissions

Some permission checks require context (e.g., event ownership). Always pass the necessary domain objects to permission methods.

### Fail Closed

All permission methods should default to denying access. Only explicitly allowed actions should return true.

---

## Usage Examples

### In Handlers

```go
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    event := getEventFromRequest(r)
    
    if !h.authChecker.CanEditEvent(r.Context(), user, event) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    // Proceed with update
}
```

### In Business Logic

```go
func (s *eventService) DeleteEvent(ctx context.Context, eventID int64) error {
    user, _ := auth.UserFromContext(ctx)
    event, err := s.repo.GetByID(ctx, eventID)
    if err != nil {
        return err
    }
    
    if !s.authChecker.CanDeleteEvent(ctx, user, event) {
        return models.ErrForbidden
    }
    
    return s.repo.Delete(ctx, eventID)
}
```

---

## References

- **README-LLM.md:** Type Safety Guidelines, TDD Requirements
- **LLD:** [lld/01_AUTH_LLD.md](../lld/01_AUTH_LLD.md) - Section 5.5
- **Epic:** [01_EPIC_auth.md](01_EPIC_auth.md)
