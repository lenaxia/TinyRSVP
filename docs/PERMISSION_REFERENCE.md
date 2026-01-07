# Permission Reference Guide

This guide provides comprehensive documentation for the TinyRSVP authorization system, including when to use each permission method, code examples, and best practices.

## Table of Contents

1. [Overview](#overview)
2. [AuthorizationChecker Interface](#authorizationchecker-interface)
3. [Permission Methods](#permission-methods)
4. [Usage Examples](#usage-examples)
5. [Permission Matrix](#permission-matrix)
6. [Best Practices](#best-practices)

---

## Overview

The `AuthorizationChecker` provides centralized permission checking throughout the TinyRSVP application. It implements a fail-closed security model where access is denied by default unless explicitly allowed.

**Location:** [`internal/auth/permissions.go`](../internal/auth/permissions.go)

**Key Principles:**
- Fail closed: Deny access by default
- Context-aware: Some checks require domain objects (e.g., event ownership)
- Role-based: Admin and Event Manager roles with different capabilities
- Centralized: Single source of truth for all authorization logic

---

## AuthorizationChecker Interface

```go
type AuthorizationChecker interface {
    // Role checks
    IsAdmin(user *models.User) bool
    IsEventManager(user *models.User) bool
    
    // Event permissions
    CanCreateEvent(ctx context.Context, user *models.User) bool
    CanEditEvent(ctx context.Context, user *models.User, event *models.Event) bool
    CanDeleteEvent(ctx context.Context, user *models.User, event *models.Event) bool
    CanViewEvent(ctx context.Context, user *models.User, event *models.Event) bool
    
    // Invite permissions
    CanManageInvites(ctx context.Context, user *models.User, event *models.Event) bool
    
    // RSVP permissions
    CanViewRSVPs(ctx context.Context, user *models.User, event *models.Event) bool
    
    // System permissions
    CanManageUsers(ctx context.Context, user *models.User) bool
    CanConfigureSystem(ctx context.Context, user *models.User) bool
}
```

---

## Permission Methods

### Role Checks

#### IsAdmin

**Purpose:** Check if user has admin role

**Parameters:**
- `user *models.User` - User to check (can be nil)

**Returns:** `bool` - true if user is admin

**When to use:**
- Quick role checks without context
- Middleware role validation
- UI conditional rendering

**Example:**
```go
if authChecker.IsAdmin(user) {
    // Show admin-only UI elements
}
```

#### IsEventManager

**Purpose:** Check if user has event manager role (includes admins)

**Parameters:**
- `user *models.User` - User to check (can be nil)

**Returns:** `bool` - true if user is event manager or admin

**When to use:**
- Checking if user can manage events
- Middleware for event-related endpoints
- UI conditional rendering

**Example:**
```go
if authChecker.IsEventManager(user) {
    // Show event management UI
}
```

---

### Event Permissions

#### CanCreateEvent

**Purpose:** Check if user can create new events

**Parameters:**
- `ctx context.Context` - Request context
- `user *models.User` - User to check

**Returns:** `bool` - true if user can create events

**When to use:**
- Before creating a new event
- In event creation handlers
- UI conditional rendering for "Create Event" button

**Example:**
```go
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    if !h.authChecker.CanCreateEvent(r.Context(), user) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // Proceed with event creation
}
```

#### CanEditEvent

**Purpose:** Check if user can edit a specific event

**Parameters:**
- `ctx context.Context` - Request context
- `user *models.User` - User to check
- `event *models.Event` - Event to edit

**Returns:** `bool` - true if user can edit the event

**When to use:**
- Before updating event details
- In event update handlers
- UI conditional rendering for "Edit" button

**Rules:**
- Admins can edit any event
- Event owners can edit their own events
- Non-owners cannot edit events

**Example:**
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

#### CanDeleteEvent

**Purpose:** Check if user can delete a specific event

**Parameters:**
- `ctx context.Context` - Request context
- `user *models.User` - User to check
- `event *models.Event` - Event to delete

**Returns:** `bool` - true if user can delete the event

**When to use:**
- Before deleting an event
- In event deletion handlers
- UI conditional rendering for "Delete" button

**Rules:**
- Admins can delete any event in any status
- Event owners can delete draft or published events
- Event owners cannot delete completed or cancelled events
- Non-owners cannot delete events

**Example:**
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

#### CanViewEvent

**Purpose:** Check if user can view a specific event

**Parameters:**
- `ctx context.Context` - Request context
- `user *models.User` - User to check
- `event *models.Event` - Event to view

**Returns:** `bool` - true if user can view the event

**When to use:**
- Before displaying event details
- In event retrieval handlers
- List filtering

**Rules:**
- All event managers (including admins) can view all events

**Example:**
```go
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    event := getEventFromRequest(r)
    
    if !h.authChecker.CanViewEvent(r.Context(), user, event) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    respondJSON(w, http.StatusOK, event)
}
```

---

### Invite Permissions

#### CanManageInvites

**Purpose:** Check if user can manage invites for a specific event

**Parameters:**
- `ctx context.Context` - Request context
- `user *models.User` - User to check
- `event *models.Event` - Event to manage invites for

**Returns:** `bool` - true if user can manage invites

**When to use:**
- Before creating/updating/deleting invites
- In invite management handlers
- UI conditional rendering for invite management

**Rules:**
- Same as `CanEditEvent` (admins or event owners)

**Example:**
```go
func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    event := getEventFromRequest(r)
    
    if !h.authChecker.CanManageInvites(r.Context(), user, event) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // Proceed with invite creation
}
```

---

### RSVP Permissions

#### CanViewRSVPs

**Purpose:** Check if user can view RSVPs for a specific event

**Parameters:**
- `ctx context.Context` - Request context
- `user *models.User` - User to check
- `event *models.Event` - Event to view RSVPs for

**Returns:** `bool` - true if user can view RSVPs

**When to use:**
- Before displaying RSVP list
- In RSVP retrieval handlers
- UI conditional rendering for RSVP viewer

**Rules:**
- Same as `CanEditEvent` (admins or event owners)

**Example:**
```go
func (h *RSVPHandler) ListRSVPs(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    event := getEventFromRequest(r)
    
    if !h.authChecker.CanViewRSVPs(r.Context(), user, event) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // Proceed with RSVP listing
}
```

---

### System Permissions

#### CanManageUsers

**Purpose:** Check if user can manage other users

**Parameters:**
- `ctx context.Context` - Request context
- `user *models.User` - User to check

**Returns:** `bool` - true if user can manage users

**When to use:**
- Before listing/viewing/updating/deleting users
- In user management handlers
- UI conditional rendering for user management

**Rules:**
- Only admins can manage users

**Example:**
```go
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    if !h.authChecker.CanManageUsers(r.Context(), user) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // Proceed with user listing
}
```

#### CanConfigureSystem

**Purpose:** Check if user can configure system settings

**Parameters:**
- `ctx context.Context` - Request context
- `user *models.User` - User to check

**Returns:** `bool` - true if user can configure system

**When to use:**
- Before updating system configuration
- In configuration handlers
- UI conditional rendering for system settings

**Rules:**
- Only admins can configure system

**Example:**
```go
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    if !h.authChecker.CanConfigureSystem(r.Context(), user) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // Proceed with configuration update
}
```

---

## Usage Examples

### Handler Integration

```go
type UserHandler struct {
    userService UserService
    authChecker auth.AuthorizationChecker
}

func NewUserHandler(userService UserService, authChecker auth.AuthorizationChecker) *UserHandler {
    return &UserHandler{
        userService: userService,
        authChecker: authChecker,
    }
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, userIDStr string) {
    currentUser, _ := auth.UserFromContext(r.Context())
    if !h.authChecker.CanManageUsers(r.Context(), currentUser) {
        respondError(w, http.StatusForbidden, "insufficient permissions")
        return
    }
    
    // Proceed with deletion
}
```

### Service Layer Integration

```go
type eventService struct {
    repo        EventRepository
    authChecker auth.AuthorizationChecker
}

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

### Main Application Wiring

```go
func main() {
    // Initialize dependencies
    userRepo := repositories.NewUserRepository(database)
    userService := auth.NewUserService(userRepo)
    authChecker := auth.NewAuthorizationChecker()
    
    // Wire handlers with authChecker
    userHandler := handlers.NewUserHandler(userService, authChecker)
    
    // Register routes
    mux.Handle("/api/users", requireAuth(requireAdmin(http.HandlerFunc(userHandler.ListUsers))))
}
```

---

## Permission Matrix

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

---

## Best Practices

### 1. Always Check Permissions in Handlers

Every handler that performs privileged operations MUST check permissions before proceeding.

```go
// ✅ Good
func (h *Handler) PrivilegedAction(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    if !h.authChecker.CanDoAction(r.Context(), user) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // Proceed
}

// ❌ Bad - No permission check
func (h *Handler) PrivilegedAction(w http.ResponseWriter, r *http.Request) {
    // Directly proceeds without checking permissions
}
```

### 2. Use Appropriate HTTP Status Codes

- `403 Forbidden` - User is authenticated but lacks permission
- `401 Unauthorized` - User is not authenticated

```go
if user == nil {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

if !h.authChecker.CanDoAction(r.Context(), user) {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}
```

### 3. Check Permissions Early

Perform permission checks before expensive operations like database queries.

```go
// ✅ Good - Check permission first
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    event := getEventFromRequest(r)
    
    if !h.authChecker.CanEditEvent(r.Context(), user, event) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    // Expensive operation
    if err := h.service.Update(r.Context(), event); err != nil {
        // Handle error
    }
}

// ❌ Bad - Expensive operation before permission check
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
    event := getEventFromRequest(r)
    
    // Expensive operation
    if err := h.service.Update(r.Context(), event); err != nil {
        // Handle error
    }
    
    user, _ := auth.UserFromContext(r.Context())
    if !h.authChecker.CanEditEvent(r.Context(), user, event) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
}
```

### 4. Handle Nil Users Gracefully

The AuthorizationChecker handles nil users safely (returns false), but be explicit in your code.

```go
user, ok := auth.UserFromContext(r.Context())
if !ok || user == nil {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

if !h.authChecker.CanDoAction(r.Context(), user) {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}
```

### 5. Don't Bypass Authorization

Never bypass the AuthorizationChecker with direct role checks in handlers.

```go
// ❌ Bad - Direct role check bypasses centralized logic
func (h *Handler) Action(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    if user.Role != models.RoleAdmin {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
}

// ✅ Good - Use AuthorizationChecker
func (h *Handler) Action(w http.ResponseWriter, r *http.Request) {
    user, _ := auth.UserFromContext(r.Context())
    if !h.authChecker.CanDoAction(r.Context(), user) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
}
```

### 6. Test Permission Checks

Always write tests that verify permission checks work correctly.

```go
func TestHandler_WithAuthorizationCheck(t *testing.T) {
    tests := []struct {
        name           string
        currentUser    *models.User
        canDoAction    bool
        wantStatus     int
    }{
        {
            name: "admin can perform action",
            currentUser: &models.User{
                ID:   1,
                Role: models.RoleAdmin,
            },
            canDoAction: true,
            wantStatus:  http.StatusOK,
        },
        {
            name: "non-admin cannot perform action",
            currentUser: &models.User{
                ID:   2,
                Role: models.RoleEventManager,
            },
            canDoAction: false,
            wantStatus:  http.StatusForbidden,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockAuthChecker := &mockAuthorizationChecker{
                CanDoActionFunc: func(ctx context.Context, user *models.User) bool {
                    return tt.canDoAction
                },
            }
            
            handler := NewHandler(mockService, mockAuthChecker)
            
            req := httptest.NewRequest(http.MethodPost, "/api/action", nil)
            ctx := auth.WithUser(req.Context(), tt.currentUser)
            req = req.WithContext(ctx)
            w := httptest.NewRecorder()
            
            handler.Action(w, req)
            
            if w.Code != tt.wantStatus {
                t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
            }
        })
    }
}
```

---

## Related Documentation

- [Authentication LLD](lld/01_AUTH_LLD.md) - Low-level design for authentication and authorization
- [RBAC Middleware](../internal/middleware/README.md) - Role-based access control middleware
- [User Story 08](00_BACKLOG/01_STORY_08_permission_checks.md) - Permission checking service story

---

**Last Updated:** 2026-01-07
