# User Story: Admin Routes

**Epic:** [08_EPIC_api.md](08_EPIC_api.md)  
**Priority:** Medium  
**Status:** Not Started  
**Estimated Effort:** 1.5 days

---

## User Story

As an **administrator**, I want **admin-only routes for user and system management** so that **I can manage users and configure system settings**.

---

## Acceptance Criteria

- [ ] GET /admin - Admin dashboard
- [ ] GET /admin/users - List users
- [ ] POST /admin/users - Create user
- [ ] PUT /admin/users/{id} - Update user
- [ ] DELETE /admin/users/{id} - Delete user
- [ ] GET /admin/settings - System settings
- [ ] PUT /admin/settings - Update settings
- [ ] Admin role required for all routes
- [ ] Audit logging for admin actions
- [ ] CSRF protection

---

## Technical Details

### Routes
```go
r.Route("/admin", func(r chi.Router) {
    r.Use(RequireAuth)
    r.Use(RequireAdmin)
    
    r.Get("/", handlers.AdminDashboard)
    
    r.Route("/users", func(r chi.Router) {
        r.Get("/", handlers.ListUsers)
        r.Post("/", handlers.CreateUser)
        r.Put("/{id}", handlers.UpdateUser)
        r.Delete("/{id}", handlers.DeleteUser)
    })
    
    r.Get("/settings", handlers.GetSettings)
    r.Put("/settings", handlers.UpdateSettings)
})
```

---

## Tasks

- [ ] Implement admin dashboard handler
- [ ] Implement user CRUD handlers
- [ ] Implement settings handlers
- [ ] Add admin role check
- [ ] Add audit logging
- [ ] Test admin routes
- [ ] Test permission enforcement

---

## Dependencies

**Depends on:** 
- 08_STORY_00_router_setup.md
- Epic 01 (Auth with RBAC)

**Blocks:** 08_STORY_16_admin_ui.md

---

## References

- **Epic:** [08_EPIC_api.md](08_EPIC_api.md)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All routes implemented
- [ ] Admin checks working
- [ ] Audit logging functional
- [ ] Tests passing
- [ ] Documentation complete
