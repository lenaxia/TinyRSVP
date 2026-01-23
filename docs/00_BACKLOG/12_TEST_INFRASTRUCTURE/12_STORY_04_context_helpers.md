# User Story: Auth Context Helpers

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Medium
**Status:** Complete ✅
**Estimated Effort:** 30 minutes
**Phase:** 1 - Foundation

---

## User Story

As a **developer**, I want **centralized auth context helpers** so that **I can easily create authenticated contexts for testing without duplicating setup code**.

---

## Acceptance Criteria

- [x] `internal/testutil/context.go` created ✅
- [x] `CreateTestContext()` function implemented ✅
- [x] `CreateAdminContext()` function implemented ✅
- [x] `CreateEventManagerContext()` function implemented ✅
- [x] `CreateAnonymousContext()` function implemented ✅
- [x] All functions have tests ✅
- [x] Documentation with examples ✅

---

## Implementation

```go
package testutil

import (
    "context"
    
    "github.com/lenaxia/tinyrsvp/internal/auth"
    "github.com/lenaxia/tinyrsvp/internal/models"
)

// CreateTestContext creates a context with the given user attached.
func CreateTestContext(user *models.User) context.Context {
    return context.WithValue(context.Background(), auth.UserContextKey, user)
}

// CreateAdminContext creates a context with a test admin user.
func CreateAdminContext() context.Context {
    return CreateTestContext(&models.User{
        ID:    1,
        Email: "admin@example.com",
        Name:  "Admin User",
        Role:  models.RoleAdmin,
    })
}

// CreateEventManagerContext creates a context with a test event manager user.
func CreateEventManagerContext() context.Context {
    return CreateTestContext(&models.User{
        ID:    2,
        Email: "manager@example.com",
        Name:  "Event Manager",
        Role:  models.RoleEventManager,
    })
}

// CreateAnonymousContext creates a context with no user attached.
func CreateAnonymousContext() context.Context {
    return context.Background()
}
```

---

## Dependencies

**Depends on:** Story 01

---

## Usage

```go
ctx := testutil.CreateAdminContext()
service.DoSomething(ctx, eventID)
```
